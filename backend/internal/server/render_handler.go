package server

import (
	"context"
	"encoding/json"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"go.uber.org/zap"

	"github.com/user/jsbug/internal/chrome"
	"github.com/user/jsbug/internal/config"
	"github.com/user/jsbug/internal/fetcher"
	"github.com/user/jsbug/internal/parser"
	"github.com/user/jsbug/internal/screenshot"
	"github.com/user/jsbug/internal/session"
	"github.com/user/jsbug/internal/types"
)

// Fetcher interface for fetching pages
type Fetcher interface {
	Fetch(ctx context.Context, opts fetcher.FetchOptions) (*fetcher.FetchResult, error)
}

// RenderHandler handles render API requests
type RenderHandler struct {
	pool            *chrome.ChromePool
	fetcher         Fetcher
	parser          *parser.Parser
	config          *config.Config
	logger          *zap.Logger
	sseManager      *SSEManager
	tokenManager    *session.TokenManager
	screenshotStore *screenshot.ScreenshotStore
}

// NewRenderHandler creates a new RenderHandler
// tokenManager is optional - pass nil when captcha/session tokens are disabled
// screenshotStore is optional - pass nil to disable screenshot storage
func NewRenderHandler(pool *chrome.ChromePool, fetcher Fetcher, parser *parser.Parser, cfg *config.Config, logger *zap.Logger, tokenManager *session.TokenManager, screenshotStore *screenshot.ScreenshotStore) *RenderHandler {
	h := &RenderHandler{
		pool:            pool,
		fetcher:         fetcher,
		parser:          parser,
		config:          cfg,
		logger:          logger,
		tokenManager:    tokenManager,
		screenshotStore: screenshotStore,
	}

	if tokenManager != nil {
		logger.Info("Session token verification enabled")
	}

	return h
}

// SetSSEManager sets the SSE manager for publishing progress events
func (h *RenderHandler) SetSSEManager(manager *SSEManager) {
	h.sseManager = manager
}

// ServeHTTP handles POST /api/render requests
func (h *RenderHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()

	if r.Method != http.MethodPost {
		h.writeError(w, http.StatusMethodNotAllowed, types.ErrInvalidURL, "Method not allowed")
		return
	}

	// Parse request body
	var req types.RenderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, types.ErrInvalidURL, "Invalid JSON request body")
		return
	}

	// Apply defaults
	req.ApplyDefaults()

	// Validate session token if enabled
	if h.tokenManager != nil {
		if req.SessionToken == "" {
			h.writeError(w, http.StatusForbidden, types.ErrSessionTokenRequired, "Session token required")
			return
		}

		// Compute fingerprint from request headers
		fingerprint := session.HashFingerprint(
			r.Header.Get("User-Agent"),
			r.Header.Get("Accept-Language"),
			r.Header.Get("Accept-Encoding"),
		)

		if err := h.tokenManager.ValidateToken(req.SessionToken, fingerprint); err != nil {
			switch err {
			case session.ErrTokenExpired:
				h.writeError(w, http.StatusForbidden, types.ErrSessionTokenExpired, "Session token expired")
			case session.ErrFingerprintMismatch:
				h.writeError(w, http.StatusForbidden, types.ErrSessionTokenInvalid, "Session token invalid (fingerprint mismatch)")
			default:
				h.writeError(w, http.StatusForbidden, types.ErrSessionTokenInvalid, "Session token invalid")
			}
			return
		}
	}

	// Validate request
	if err := h.validateRequest(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, err.Code, err.Message)
		return
	}

	// Process request based on JS enabled flag
	var response *types.RenderResponse
	if req.JSEnabled {
		response = h.handleJSRender(r.Context(), &req)
	} else {
		response = h.handleFetch(r.Context(), &req)
	}

	h.writeJSON(w, response)

	logFields := []zap.Field{
		zap.String("url", req.URL),
		zap.Bool("js_enabled", req.JSEnabled),
		zap.Bool("success", response.Success),
		zap.Float64("total_time", time.Since(startTime).Seconds()),
	}
	if response.Data != nil {
		logFields = append(logFields, zap.Int("status_code", response.Data.StatusCode))
	}
	h.logger.Info("Render request", logFields...)
}

// validateRequest validates the render request
func (h *RenderHandler) validateRequest(req *types.RenderRequest) *types.RenderError {
	// Validate URL
	if req.URL == "" {
		return &types.RenderError{Code: types.ErrInvalidURL, Message: "URL is required"}
	}

	u, err := url.Parse(req.URL)
	if err != nil {
		return &types.RenderError{Code: types.ErrInvalidURL, Message: "Invalid URL format"}
	}

	if u.Scheme != "http" && u.Scheme != "https" {
		return &types.RenderError{Code: types.ErrInvalidURL, Message: "URL must use http or https scheme"}
	}

	if u.Host == "" {
		return &types.RenderError{Code: types.ErrInvalidURL, Message: "URL must have a host"}
	}

	// Validate timeout
	if !req.ValidateTimeout() {
		return &types.RenderError{
			Code:    types.ErrInvalidTimeout,
			Message: "Timeout must be between 1 and 60 seconds",
		}
	}

	// Validate wait event
	if !types.IsValidWaitEvent(req.WaitEvent) {
		return &types.RenderError{
			Code:    types.ErrInvalidWaitEvent,
			Message: "Invalid wait event: " + req.WaitEvent,
		}
	}

	return nil
}

// handleJSRender processes a request with JavaScript rendering
func (h *RenderHandler) handleJSRender(ctx context.Context, req *types.RenderRequest) *types.RenderResponse {
	requestID := req.RequestID

	// Check if pool is available
	if h.pool == nil {
		h.publishError(requestID, types.ErrChromeUnavailable, "Chrome pool is not available")
		return &types.RenderResponse{
			Success: false,
			Error: &types.RenderError{
				Code:    types.ErrChromeUnavailable,
				Message: "Chrome pool is not available",
			},
		}
	}

	// Acquire instance from pool
	instance, err := h.pool.Acquire()
	if err != nil {
		if err == chrome.ErrNoInstanceAvailable {
			h.publishError(requestID, types.ErrPoolExhausted, "Service unavailable, try again")
			return &types.RenderResponse{
				Success: false,
				Error: &types.RenderError{
					Code:    types.ErrPoolExhausted,
					Message: "Service unavailable, try again",
				},
			}
		}
		if err == chrome.ErrPoolShuttingDown {
			h.publishError(requestID, types.ErrPoolShuttingDown, "Service shutting down")
			return &types.RenderResponse{
				Success: false,
				Error: &types.RenderError{
					Code:    types.ErrPoolShuttingDown,
					Message: "Service shutting down",
				},
			}
		}
		// Other error (e.g., restart failed)
		h.publishError(requestID, types.ErrChromeUnavailable, err.Error())
		return &types.RenderResponse{
			Success: false,
			Error: &types.RenderError{
				Code:    types.ErrChromeUnavailable,
				Message: err.Error(),
			},
		}
	}
	defer h.pool.Release(instance)

	// Create renderer for this request
	renderer := chrome.NewRendererV2(instance, h.logger)

	// Publish started event
	h.publishStarted(requestID, req.URL)

	// Create blocklist
	blocklist := chrome.NewBlocklist(req.BlockAnalytics, req.BlockAds, req.BlockSocial, req.BlockedTypes)

	// Build render options
	userAgent := types.ResolveUserAgent(req.UserAgent)
	opts := chrome.RenderOptions{
		URL:       req.URL,
		UserAgent: userAgent,
		Timeout:   time.Duration(req.Timeout) * time.Second,
		WaitEvent: req.WaitEvent,
		Blocklist: blocklist,
		IsMobile:  isMobileUserAgent(userAgent),
	}

	// Publish navigating event
	h.publishNavigating(requestID, req.URL)

	// Render the page
	result, err := renderer.Render(ctx, opts)
	if err != nil {
		resp := h.handleRenderError(err)
		if resp.Error != nil {
			h.publishError(requestID, resp.Error.Code, resp.Error.Message)
		}
		return resp
	}

	// Publish capturing event
	h.publishCapturing(requestID, len(result.Network))

	// Publish parsing event
	h.publishParsing(requestID)

	// Parse HTML content (JS mode doesn't capture HTTP headers)
	parseResult, _ := h.parser.ParseWithOptions(result.HTML, parser.ParseOptions{
		PageURL: result.FinalURL,
	})

	// Publish complete event
	h.publishComplete(requestID, result.RenderTime)

	return h.buildJSResponse(result, parseResult)
}

// handleFetch processes a request without JavaScript rendering
func (h *RenderHandler) handleFetch(ctx context.Context, req *types.RenderRequest) *types.RenderResponse {
	requestID := req.RequestID

	if h.fetcher == nil {
		h.publishError(requestID, types.ErrFetchFailed, "HTTP fetcher is not available")
		return &types.RenderResponse{
			Success: false,
			Error: &types.RenderError{
				Code:    types.ErrFetchFailed,
				Message: "HTTP fetcher is not available",
			},
		}
	}

	// Publish started event
	h.publishStarted(requestID, req.URL)

	// Build fetch options
	opts := fetcher.FetchOptions{
		URL:             req.URL,
		UserAgent:       types.ResolveUserAgent(req.UserAgent),
		Timeout:         time.Duration(req.Timeout) * time.Second,
		FollowRedirects: req.ShouldFollowRedirects(),
	}

	// Publish navigating event
	h.publishNavigating(requestID, req.URL)

	// Fetch the page
	result, err := h.fetcher.Fetch(ctx, opts)
	if err != nil {
		resp := h.handleFetchError(err)
		if resp.Error != nil {
			h.publishError(requestID, resp.Error.Code, resp.Error.Message)
		}
		return resp
	}

	// Publish parsing event
	h.publishParsing(requestID)

	// Parse HTML content with headers
	parseResult, _ := h.parser.ParseWithOptions(result.HTML, parser.ParseOptions{
		PageURL:    result.FinalURL,
		XRobotsTag: result.GetXRobotsTag(),
		LinkHeader: result.GetLinkHeader(),
	})

	// Publish complete event
	h.publishComplete(requestID, result.FetchTime)

	return h.buildFetchResponse(result, parseResult)
}

// buildJSResponse builds response from JS render result
func (h *RenderHandler) buildJSResponse(result *chrome.RenderResult, parseResult *parser.ParseResult) *types.RenderResponse {
	data := &types.RenderData{
		StatusCode:    result.StatusCode,
		FinalURL:      result.FinalURL,
		RedirectURL:   result.RedirectURL,
		PageSizeBytes: result.PageSizeBytes,
		RenderTime:    result.RenderTime,
		HTML:          result.HTML,
		Requests:      result.Network,
		Console:       result.Console,
		JSErrors:      result.JSErrors,
		Lifecycle:     result.Lifecycle,
	}

	// Store screenshot and set ID if available
	if h.screenshotStore != nil && len(result.Screenshot) > 0 {
		data.ScreenshotID = h.screenshotStore.Store(result.Screenshot)
	}

	// Add parsed content
	if parseResult != nil {
		h.applyParseResult(data, parseResult)
	}

	// Enrich images with sizes from network requests (JS mode only)
	enrichImagesWithSizes(data.Images, data.Requests)

	return &types.RenderResponse{
		Success: true,
		Data:    data,
	}
}

// buildFetchResponse builds response from HTTP fetch result
func (h *RenderHandler) buildFetchResponse(result *fetcher.FetchResult, parseResult *parser.ParseResult) *types.RenderResponse {
	data := &types.RenderData{
		StatusCode:    result.StatusCode,
		FinalURL:      result.FinalURL,
		RedirectURL:   result.RedirectURL,
		PageSizeBytes: result.PageSizeBytes,
		RenderTime:    result.FetchTime,
		HTML:          result.HTML,
		XRobotsTag:    result.GetXRobotsTag(),
	}

	// Check for canonical in Link header
	if canonical := result.GetCanonicalFromHeader(); canonical != "" {
		data.CanonicalURL = canonical
	}

	// Add parsed content
	if parseResult != nil {
		h.applyParseResult(data, parseResult)
	}

	return &types.RenderResponse{
		Success: true,
		Data:    data,
	}
}

// applyParseResult applies parsed content to render data
func (h *RenderHandler) applyParseResult(data *types.RenderData, parseResult *parser.ParseResult) {
	data.Title = parseResult.Title
	data.MetaDescription = parseResult.MetaDescription
	data.MetaRobots = parseResult.MetaRobots
	data.H1 = parseResult.H1
	data.H2 = parseResult.H2
	data.H3 = parseResult.H3
	data.WordCount = parseResult.WordCount
	data.OpenGraph = parseResult.OpenGraph
	data.StructuredData = parseResult.StructuredData

	// New fields from extended extraction
	data.BodyText = parseResult.BodyText
	data.BodyTextTokensCount = parser.CountBodyTextTokens(parseResult.BodyText)
	data.BodyMarkdown = parseResult.BodyMarkdown
	data.TextHtmlRatio = parseResult.TextHtmlRatio
	data.HrefLangs = parseResult.HrefLangs
	data.Links = parseResult.Links
	data.Images = parseResult.Images
	data.MetaIndexable = parseResult.MetaIndexable
	data.MetaFollow = parseResult.MetaFollow

	// Use canonical from HTML if not set from header
	if data.CanonicalURL == "" {
		data.CanonicalURL = parseResult.CanonicalURL
	}
}

// handleRenderError converts render errors to response
func (h *RenderHandler) handleRenderError(err error) *types.RenderResponse {
	errMsg := err.Error()

	if strings.Contains(errMsg, "context deadline exceeded") {
		return &types.RenderResponse{
			Success: false,
			Error: &types.RenderError{
				Code:    types.ErrRenderTimeout,
				Message: "Render timeout exceeded",
			},
		}
	}

	if strings.Contains(errMsg, "net::ERR_NAME_NOT_RESOLVED") {
		return &types.RenderResponse{
			Success: false,
			Error: &types.RenderError{
				Code:    types.ErrDomainNotFound,
				Message: "Domain not found - check URL for typos",
			},
		}
	}

	if strings.Contains(errMsg, "failed to capture status code") {
		return &types.RenderResponse{
			Success: false,
			Error: &types.RenderError{
				Code:    types.ErrDomainNotFound,
				Message: "Domain not found - check URL for typos",
			},
		}
	}

	return &types.RenderResponse{
		Success: false,
		Error: &types.RenderError{
			Code:    types.ErrRenderFailed,
			Message: "Render failed: " + errMsg,
		},
	}
}

// handleFetchError converts fetch errors to response
func (h *RenderHandler) handleFetchError(err error) *types.RenderResponse {
	errMsg := err.Error()

	if strings.Contains(errMsg, "context deadline exceeded") ||
		strings.Contains(errMsg, "Client.Timeout") {
		return &types.RenderResponse{
			Success: false,
			Error: &types.RenderError{
				Code:    types.ErrRenderTimeout,
				Message: "Fetch timeout exceeded",
			},
		}
	}

	if strings.Contains(errMsg, "no such host") {
		return &types.RenderResponse{
			Success: false,
			Error: &types.RenderError{
				Code:    types.ErrDomainNotFound,
				Message: "Domain not found - check URL for typos",
			},
		}
	}

	return &types.RenderResponse{
		Success: false,
		Error: &types.RenderError{
			Code:    types.ErrFetchFailed,
			Message: "Fetch failed: " + errMsg,
		},
	}
}

// writeJSON writes a JSON response
func (h *RenderHandler) writeJSON(w http.ResponseWriter, response *types.RenderResponse) {
	w.Header().Set("Content-Type", "application/json")

	if !response.Success {
		// Set appropriate status code based on error
		statusCode := http.StatusInternalServerError
		if response.Error != nil {
			switch response.Error.Code {
			case types.ErrInvalidURL, types.ErrInvalidTimeout, types.ErrInvalidWaitEvent, types.ErrDomainNotFound:
				statusCode = http.StatusBadRequest
			case types.ErrRenderTimeout:
				statusCode = http.StatusRequestTimeout
			case types.ErrChromeUnavailable, types.ErrPoolExhausted, types.ErrPoolShuttingDown:
				statusCode = http.StatusServiceUnavailable
			case types.ErrSessionTokenRequired, types.ErrSessionTokenInvalid, types.ErrSessionTokenExpired:
				statusCode = http.StatusForbidden
			}
		}
		w.WriteHeader(statusCode)
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("Failed to write response", zap.Error(err))
	}
}

// writeError writes an error response
func (h *RenderHandler) writeError(w http.ResponseWriter, statusCode int, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	response := &types.RenderResponse{
		Success: false,
		Error: &types.RenderError{
			Code:    code,
			Message: message,
		},
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("Failed to write error response", zap.Error(err))
	}
}

// SSE publishing helpers - only publish if manager is set and request has ID
func (h *RenderHandler) publishStarted(requestID, url string) {
	if h.sseManager != nil && requestID != "" {
		h.sseManager.PublishStarted(requestID, url)
	}
}

func (h *RenderHandler) publishNavigating(requestID, url string) {
	if h.sseManager != nil && requestID != "" {
		h.sseManager.PublishNavigating(requestID, url)
	}
}

func (h *RenderHandler) publishCapturing(requestID string, requestCount int) {
	if h.sseManager != nil && requestID != "" {
		h.sseManager.PublishCapturing(requestID, requestCount)
	}
}

func (h *RenderHandler) publishParsing(requestID string) {
	if h.sseManager != nil && requestID != "" {
		h.sseManager.PublishParsing(requestID)
	}
}

func (h *RenderHandler) publishComplete(requestID string, renderTime float64) {
	if h.sseManager != nil && requestID != "" {
		h.sseManager.PublishComplete(requestID, renderTime)
	}
}

func (h *RenderHandler) publishError(requestID, code, message string) {
	if h.sseManager != nil && requestID != "" {
		h.sseManager.PublishError(requestID, code, message)
	}
}

// isMobileUserAgent checks if the User-Agent indicates a mobile device
func isMobileUserAgent(ua string) bool {
	ua = strings.ToLower(ua)
	mobileKeywords := []string{"mobile", "android", "iphone", "ipad", "ipod"}
	for _, keyword := range mobileKeywords {
		if strings.Contains(ua, keyword) {
			return true
		}
	}
	return false
}

// getClientIP extracts the client IP address from the request
// Checks common proxy headers before falling back to RemoteAddr
func getClientIP(r *http.Request) string {
	// Check CF-Connecting-IP first (if behind Cloudflare)
	if ip := r.Header.Get("CF-Connecting-IP"); ip != "" {
		return ip
	}
	// Check X-Forwarded-For
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		// First IP in the list is the original client
		if idx := strings.Index(xff, ","); idx > 0 {
			return strings.TrimSpace(xff[:idx])
		}
		return strings.TrimSpace(xff)
	}
	// Check X-Real-IP
	if ip := r.Header.Get("X-Real-IP"); ip != "" {
		return ip
	}
	// Fall back to RemoteAddr
	host, _, _ := net.SplitHostPort(r.RemoteAddr)
	return host
}

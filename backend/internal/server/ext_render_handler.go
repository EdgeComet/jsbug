package server

import (
	"crypto/subtle"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"go.uber.org/zap"

	"github.com/PuerkitoBio/goquery"
	"github.com/user/jsbug/internal/config"
	"github.com/user/jsbug/internal/parser"
	"github.com/user/jsbug/internal/types"
)

// ExtRenderHandler handles external API render requests
type ExtRenderHandler struct {
	renderHandler *RenderHandler
	apiKeys       map[string]bool
	logger        *zap.Logger
}

// NewExtRenderHandler creates a new ExtRenderHandler
func NewExtRenderHandler(renderHandler *RenderHandler, cfg *config.Config, logger *zap.Logger) *ExtRenderHandler {
	keys := make(map[string]bool)
	for _, k := range cfg.API.Keys {
		keys[k] = true
	}
	return &ExtRenderHandler{
		renderHandler: renderHandler,
		apiKeys:       keys,
		logger:        logger,
	}
}

func (h *ExtRenderHandler) validateAPIKey(provided string) bool {
	for key := range h.apiKeys {
		if subtle.ConstantTimeCompare([]byte(provided), []byte(key)) == 1 {
			return true
		}
	}
	return false
}

// ServeHTTP handles POST /api/ext/render requests
func (h *ExtRenderHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()

	if r.Method != http.MethodPost {
		h.writeError(w, http.StatusMethodNotAllowed, types.ErrMethodNotAllowed, "Method not allowed")
		return
	}

	apiKey := r.Header.Get("X-API-Key")
	if apiKey == "" {
		h.writeError(w, http.StatusUnauthorized, types.ErrAPIKeyRequired, "API key required")
		return
	}
	if !h.validateAPIKey(apiKey) {
		h.writeError(w, http.StatusForbidden, types.ErrAPIKeyInvalid, "Invalid API key")
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

	var extReq types.ExtRenderRequest
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&extReq); err != nil {
		h.logger.Warn("Invalid request body", zap.Error(err))
		h.writeError(w, http.StatusBadRequest, types.ErrInvalidRequestBody, "Invalid request body")
		return
	}

	req := extReq.ToRenderRequest()
	req.ApplyDefaults()

	if renderErr := h.renderHandler.validateRequest(req); renderErr != nil {
		h.writeError(w, types.ErrorCodeToHTTPStatus(renderErr.Code), renderErr.Code, renderErr.Message)
		return
	}

	var response *types.RenderResponse
	if req.JSEnabled {
		response = h.renderHandler.handleJSRender(r.Context(), req)
	} else {
		response = h.renderHandler.handleFetch(r.Context(), req)
	}

	if !response.Success {
		extResp := &types.ExtRenderResponse{
			Success: false,
			Error:   response.Error,
		}
		statusCode := h.errorToStatus(response.Error)
		h.writeJSON(w, statusCode, extResp)
		h.logRequest(extReq, apiKey, time.Since(startTime).Seconds(), statusCode)
		return
	}

	extData := buildExtResponse(response.Data, &extReq)

	if extReq.MaxContentLength > 0 {
		truncateContent(extData, extReq.MaxContentLength)
	}

	if extData.BodyText != nil {
		tokens := parser.CountBodyTextTokens(*extData.BodyText)
		extData.BodyTextTokensCount = &tokens
	}

	extResp := &types.ExtRenderResponse{
		Success: true,
		Data:    extData,
	}

	h.writeJSON(w, http.StatusOK, extResp)
	h.logRequest(extReq, apiKey, time.Since(startTime).Seconds(), http.StatusOK)
}

func buildExtResponse(data *types.RenderData, extReq *types.ExtRenderRequest) *types.ExtRenderData {
	ext := &types.ExtRenderData{
		StatusCode:      data.StatusCode,
		FinalURL:        data.FinalURL,
		RedirectURL:     data.RedirectURL,
		CanonicalURL:    data.CanonicalURL,
		PageSizeBytes:   data.PageSizeBytes,
		RenderTime:      data.RenderTime,
		MetaRobots:      data.MetaRobots,
		XRobotsTag:      data.XRobotsTag,
		MetaIndexable:   data.MetaIndexable,
		MetaFollow:      data.MetaFollow,
		Title:           data.Title,
		MetaDescription: data.MetaDescription,
		WordCount:       data.WordCount,
		TextHtmlRatio:   data.TextHtmlRatio,
		OpenGraph:       data.OpenGraph,
		HrefLangs:       data.HrefLangs,
	}

	if data.H1 != nil {
		ext.H1 = data.H1
	} else {
		ext.H1 = []string{}
	}
	if data.H2 != nil {
		ext.H2 = data.H2
	} else {
		ext.H2 = []string{}
	}
	if data.H3 != nil {
		ext.H3 = data.H3
	} else {
		ext.H3 = []string{}
	}
	if ext.OpenGraph == nil {
		ext.OpenGraph = map[string]string{}
	}
	if ext.HrefLangs == nil {
		ext.HrefLangs = []types.HrefLang{}
	}

	if extReq.IncludeHTML {
		html := data.HTML
		ext.HTML = &html
	}
	if extReq.IncludeText {
		text := data.BodyText
		ext.BodyText = &text
	}
	if extReq.IncludeMarkdown {
		md := data.BodyMarkdown
		ext.BodyMarkdown = &md
	}
	if extReq.IncludeLinks {
		if data.Links != nil {
			ext.Links = data.Links
		} else {
			ext.Links = []types.Link{}
		}
	}
	if extReq.IncludeImages {
		if data.Images != nil {
			ext.Images = data.Images
		} else {
			ext.Images = []types.Image{}
		}
	}
	if extReq.IncludeStructuredData {
		if data.StructuredData != nil {
			ext.StructuredData = data.StructuredData
		} else {
			ext.StructuredData = []json.RawMessage{}
		}
	}
	if extReq.IncludeSections {
		doc, err := goquery.NewDocumentFromReader(strings.NewReader(data.HTML))
		if err == nil {
			ext.Sections = parser.ExtractSections(doc)
		}
		if ext.Sections == nil {
			ext.Sections = []types.Section{}
		}
	}
	if extReq.IncludeScreenshot && extReq.JSEnabled && len(data.ScreenshotData) > 0 {
		encoded := base64.StdEncoding.EncodeToString(data.ScreenshotData)
		ext.Screenshot = &encoded
	}

	return ext
}

func truncateContent(data *types.ExtRenderData, maxLen int) {
	if data.HTML != nil {
		*data.HTML = truncateAtWordBoundary(*data.HTML, maxLen)
	}
	if data.BodyText != nil {
		*data.BodyText = truncateAtWordBoundary(*data.BodyText, maxLen)
	}
	if data.BodyMarkdown != nil {
		*data.BodyMarkdown = truncateAtWordBoundary(*data.BodyMarkdown, maxLen)
	}
	if len(data.Sections) > 0 {
		budget := maxLen
		truncated := make([]types.Section, 0, len(data.Sections))
		for _, s := range data.Sections {
			if budget <= 0 {
				break
			}
			bodyLen := len([]rune(s.BodyMarkdown))
			if bodyLen <= budget {
				truncated = append(truncated, s)
				budget -= bodyLen
			} else {
				s.BodyMarkdown = truncateAtWordBoundary(s.BodyMarkdown, budget)
				truncated = append(truncated, s)
				budget = 0
			}
		}
		data.Sections = truncated
	}
}

func truncateAtWordBoundary(s string, maxLen int) string {
	runes := []rune(s)
	if len(runes) <= maxLen {
		return s
	}
	truncated := string(runes[:maxLen])
	lastSpace := strings.LastIndex(truncated, " ")
	if lastSpace > 0 {
		return truncated[:lastSpace]
	}
	return truncated
}

func (h *ExtRenderHandler) errorToStatus(err *types.RenderError) int {
	if err == nil {
		return http.StatusInternalServerError
	}
	return types.ErrorCodeToHTTPStatus(err.Code)
}

func (h *ExtRenderHandler) writeJSON(w http.ResponseWriter, statusCode int, resp *types.ExtRenderResponse) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		h.logger.Error("Failed to write response", zap.Error(err))
	}
}

func (h *ExtRenderHandler) writeError(w http.ResponseWriter, statusCode int, code, message string) {
	resp := &types.ExtRenderResponse{
		Success: false,
		Error:   &types.RenderError{Code: code, Message: message},
	}
	h.writeJSON(w, statusCode, resp)
}

func (h *ExtRenderHandler) logRequest(req types.ExtRenderRequest, apiKey string, totalTime float64, status int) {
	maskedKey := apiKey
	if len(apiKey) > 4 {
		maskedKey = "***" + apiKey[len(apiKey)-4:]
	}
	h.logger.Info("Ext render request",
		zap.String("url", req.URL),
		zap.String("api_key", maskedKey),
		zap.Bool("js_enabled", req.JSEnabled),
		zap.Bool("include_html", req.IncludeHTML),
		zap.Bool("include_text", req.IncludeText),
		zap.Bool("include_markdown", req.IncludeMarkdown),
		zap.Bool("include_sections", req.IncludeSections),
		zap.Bool("include_links", req.IncludeLinks),
		zap.Bool("include_images", req.IncludeImages),
		zap.Bool("include_structured_data", req.IncludeStructuredData),
		zap.Bool("include_screenshot", req.IncludeScreenshot),
		zap.Float64("total_time", totalTime),
		zap.Int("status", status),
	)
}

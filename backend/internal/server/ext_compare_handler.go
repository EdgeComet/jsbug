package server

import (
	"crypto/subtle"
	"encoding/json"
	"net/http"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"

	"github.com/PuerkitoBio/goquery"
	"github.com/user/jsbug/internal/compare"
	"github.com/user/jsbug/internal/config"
	"github.com/user/jsbug/internal/parser"
	"github.com/user/jsbug/internal/types"
)

// ExtCompareHandler handles external API compare requests
type ExtCompareHandler struct {
	renderHandler *RenderHandler
	apiKeys       map[string]bool
	logger        *zap.Logger
}

// NewExtCompareHandler creates a new ExtCompareHandler
func NewExtCompareHandler(renderHandler *RenderHandler, cfg *config.Config, logger *zap.Logger) *ExtCompareHandler {
	keys := make(map[string]bool)
	for _, k := range cfg.API.Keys {
		keys[k] = true
	}
	return &ExtCompareHandler{
		renderHandler: renderHandler,
		apiKeys:       keys,
		logger:        logger,
	}
}

func (h *ExtCompareHandler) validateAPIKey(provided string) bool {
	for key := range h.apiKeys {
		if subtle.ConstantTimeCompare([]byte(provided), []byte(key)) == 1 {
			return true
		}
	}
	return false
}

// ServeHTTP handles POST /api/ext/compare requests
func (h *ExtCompareHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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

	var extReq types.ExtCompareRequest
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&extReq); err != nil {
		h.logger.Warn("Invalid request body", zap.Error(err))
		h.writeError(w, http.StatusBadRequest, types.ErrInvalidRequestBody, "Invalid request body")
		return
	}

	jsReq := extReq.ToJSRenderRequest()
	jsReq.ApplyDefaults()

	httpReq := extReq.ToHTTPRenderRequest()
	httpReq.ApplyDefaults()

	if renderErr := h.renderHandler.validateRequest(jsReq); renderErr != nil {
		h.writeError(w, types.ErrorCodeToHTTPStatus(renderErr.Code), renderErr.Code, renderErr.Message)
		return
	}

	// Run both fetches in parallel
	var jsResponse, httpResponse *types.RenderResponse
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		jsResponse = h.renderHandler.handleJSRender(r.Context(), jsReq)
	}()
	go func() {
		defer wg.Done()
		httpResponse = h.renderHandler.handleFetch(r.Context(), httpReq)
	}()
	wg.Wait()

	// Build FetchStatus for JS
	jsStatus := &types.FetchStatus{}
	if jsResponse.Success {
		jsStatus.Success = true
		jsStatus.StatusCode = jsResponse.Data.StatusCode
		jsStatus.RenderTime = jsResponse.Data.RenderTime
	} else {
		jsStatus.Success = false
		jsStatus.Error = jsResponse.Error
	}

	// Build FetchStatus for HTTP
	httpStatus := &types.FetchStatus{}
	if httpResponse.Success {
		httpStatus.Success = true
		httpStatus.StatusCode = httpResponse.Data.StatusCode
		httpStatus.RenderTime = httpResponse.Data.RenderTime
	} else {
		httpStatus.Success = false
		httpStatus.Error = httpResponse.Error
	}

	// Build JS primary content (only if JS fetch succeeded)
	var extData *types.ExtRenderData
	if jsResponse.Success {
		tmpExtReq := &types.ExtRenderRequest{
			URL:                   extReq.URL,
			FollowRedirects:       extReq.FollowRedirects,
			UserAgent:             extReq.UserAgent,
			Timeout:               extReq.Timeout,
			WaitEvent:             extReq.WaitEvent,
			JSEnabled:             true,
			BlockAnalytics:        extReq.BlockAnalytics,
			BlockAds:              extReq.BlockAds,
			BlockSocial:           extReq.BlockSocial,
			BlockedTypes:          extReq.BlockedTypes,
			MaxContentLength:      extReq.MaxContentLength,
			IncludeHTML:           extReq.IncludeHTML,
			IncludeText:           extReq.IncludeText,
			IncludeMarkdown:       extReq.IncludeMarkdown,
			IncludeSections:       extReq.IncludeSections,
			IncludeLinks:          extReq.IncludeLinks,
			IncludeImages:         extReq.IncludeImages,
			IncludeStructuredData: extReq.IncludeStructuredData,
		}
		extData = buildExtResponse(jsResponse.Data, tmpExtReq)

		if extReq.MaxContentLength > 0 {
			truncateContent(extData, extReq.MaxContentLength)
		}

		if extData.BodyText != nil {
			tokens := parser.CountBodyTextTokens(*extData.BodyText)
			extData.BodyTextTokensCount = &tokens
		}
	}

	// Build diff and rendering impact (only when BOTH fetches succeed)
	var diff *types.CompareDiff
	var impact *types.RenderingImpact
	if jsResponse.Success && httpResponse.Success {
		diff = &types.CompareDiff{
			Title:           compare.DiffString(jsResponse.Data.Title, httpResponse.Data.Title),
			MetaDescription: compare.DiffString(jsResponse.Data.MetaDescription, httpResponse.Data.MetaDescription),
			CanonicalURL:    compare.DiffString(jsResponse.Data.CanonicalURL, httpResponse.Data.CanonicalURL),
			MetaRobots:      compare.DiffString(jsResponse.Data.MetaRobots, httpResponse.Data.MetaRobots),
			H1:              compare.DiffStringSlice(jsResponse.Data.H1, httpResponse.Data.H1),
			H2:              compare.DiffStringSlice(jsResponse.Data.H2, httpResponse.Data.H2),
			H3:              compare.DiffStringSlice(jsResponse.Data.H3, httpResponse.Data.H3),
			WordCountJS:     jsResponse.Data.WordCount,
			WordCountNonJS:  httpResponse.Data.WordCount,
		}

		if extReq.IncludeSections {
			var jsSections, nonJSSections []types.Section
			jsDoc, err := goquery.NewDocumentFromReader(strings.NewReader(jsResponse.Data.HTML))
			if err == nil {
				jsSections = parser.ExtractSections(jsDoc)
			}
			nonJSDoc, err := goquery.NewDocumentFromReader(strings.NewReader(httpResponse.Data.HTML))
			if err == nil {
				nonJSSections = parser.ExtractSections(nonJSDoc)
			}
			diff.Sections = compare.DiffSections(jsSections, nonJSSections)
		}

		if extReq.IncludeLinks {
			diff.Links = compare.DiffLinks(jsResponse.Data.Links, httpResponse.Data.Links)
		}

		if extReq.IncludeImages {
			diff.Images = compare.DiffImages(jsResponse.Data.Images, httpResponse.Data.Images)
		}

		if extReq.IncludeStructuredData {
			diff.StructuredData = compare.DiffStructuredData(jsResponse.Data.StructuredData, httpResponse.Data.StructuredData)
		}

		impact = compare.ComputeRenderingImpact(jsResponse.Data, httpResponse.Data, diff)

		// Apply max_diff_length truncation to diff SectionDiff.NonJSBodyMarkdown fields
		if extReq.MaxDiffLength > 0 && len(diff.Sections) > 0 {
			budget := extReq.MaxDiffLength
			for i := range diff.Sections {
				if diff.Sections[i].Status != "changed" {
					continue
				}
				if budget <= 0 {
					diff.Sections[i].NonJSBodyMarkdown = ""
					continue
				}
				runeLen := len([]rune(diff.Sections[i].NonJSBodyMarkdown))
				if runeLen <= budget {
					budget -= runeLen
				} else {
					diff.Sections[i].NonJSBodyMarkdown = truncateAtWordBoundary(diff.Sections[i].NonJSBodyMarkdown, budget)
					budget = 0
				}
			}
		}
	}

	extResp := &types.ExtCompareResponse{
		Success: true,
		Data: &types.ExtCompareData{
			JSStatus:        jsStatus,
			HTTPStatus:      httpStatus,
			JS:              extData,
			Diff:            diff,
			RenderingImpact: impact,
		},
	}

	h.writeJSON(w, http.StatusOK, extResp)
	h.logRequest(extReq, apiKey, time.Since(startTime).Seconds(), http.StatusOK)
}

func (h *ExtCompareHandler) writeJSON(w http.ResponseWriter, statusCode int, resp *types.ExtCompareResponse) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		h.logger.Error("Failed to write response", zap.Error(err))
	}
}

func (h *ExtCompareHandler) writeError(w http.ResponseWriter, statusCode int, code, message string) {
	resp := &types.ExtCompareResponse{
		Success: false,
		Error:   &types.RenderError{Code: code, Message: message},
	}
	h.writeJSON(w, statusCode, resp)
}

func (h *ExtCompareHandler) logRequest(req types.ExtCompareRequest, apiKey string, totalTime float64, status int) {
	maskedKey := apiKey
	if len(apiKey) > 4 {
		maskedKey = "***" + apiKey[len(apiKey)-4:]
	}
	h.logger.Info("Ext compare request",
		zap.String("url", req.URL),
		zap.String("api_key", maskedKey),
		zap.Bool("include_html", req.IncludeHTML),
		zap.Bool("include_text", req.IncludeText),
		zap.Bool("include_markdown", req.IncludeMarkdown),
		zap.Bool("include_sections", req.IncludeSections),
		zap.Bool("include_links", req.IncludeLinks),
		zap.Bool("include_images", req.IncludeImages),
		zap.Bool("include_structured_data", req.IncludeStructuredData),
		zap.Float64("total_time", totalTime),
		zap.Int("status", status),
	)
}

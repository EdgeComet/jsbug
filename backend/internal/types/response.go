package types

import (
	"encoding/json"
	"net/http"
)

// Error codes
const (
	ErrInvalidURL           = "INVALID_URL"
	ErrInvalidTimeout       = "INVALID_TIMEOUT"
	ErrInvalidWaitEvent     = "INVALID_WAIT_EVENT"
	ErrRenderTimeout        = "RENDER_TIMEOUT"
	ErrRenderFailed         = "RENDER_FAILED"
	ErrFetchFailed          = "FETCH_FAILED"
	ErrChromeUnavailable    = "CHROME_UNAVAILABLE"
	ErrDomainNotFound       = "DOMAIN_NOT_FOUND"
	ErrPoolExhausted        = "POOL_EXHAUSTED"
	ErrPoolShuttingDown     = "POOL_SHUTTING_DOWN"
	ErrSessionTokenRequired = "SESSION_TOKEN_REQUIRED"
	ErrSessionTokenInvalid  = "SESSION_TOKEN_INVALID"
	ErrSessionTokenExpired  = "SESSION_TOKEN_EXPIRED"
	ErrAPIKeyRequired       = "API_KEY_REQUIRED"
	ErrAPIKeyInvalid        = "API_KEY_INVALID"
	ErrMethodNotAllowed     = "METHOD_NOT_ALLOWED"
	ErrInvalidRequestBody   = "INVALID_REQUEST_BODY"
	ErrSSRFBlocked          = "SSRF_BLOCKED"
)

// ErrorCodeToHTTPStatus maps an error code to the appropriate HTTP status code.
func ErrorCodeToHTTPStatus(code string) int {
	switch code {
	case ErrInvalidURL, ErrInvalidTimeout, ErrInvalidWaitEvent, ErrDomainNotFound, ErrInvalidRequestBody:
		return http.StatusBadRequest
	case ErrAPIKeyRequired:
		return http.StatusUnauthorized
	case ErrAPIKeyInvalid, ErrSessionTokenRequired, ErrSessionTokenInvalid, ErrSessionTokenExpired, ErrSSRFBlocked:
		return http.StatusForbidden
	case ErrMethodNotAllowed:
		return http.StatusMethodNotAllowed
	case ErrRenderTimeout:
		return http.StatusRequestTimeout
	case ErrChromeUnavailable, ErrPoolExhausted, ErrPoolShuttingDown:
		return http.StatusServiceUnavailable
	default:
		return http.StatusInternalServerError
	}
}

// MaxBodyTextBytes is the maximum size for extracted body text (3MB)
const MaxBodyTextBytes = 3 * 1024 * 1024

// MaxBodyMarkdownBytes is the maximum size for extracted body markdown (3MB)
const MaxBodyMarkdownBytes = 3 * 1024 * 1024

// HrefLang represents an hreflang alternate link
type HrefLang struct {
	Lang   string `json:"lang"`
	URL    string `json:"url"`
	Source string `json:"source"` // "link" or "header"
}

// Link represents a hyperlink extracted from the page
type Link struct {
	Href        string `json:"href"`
	Text        string `json:"text"`
	IsExternal  bool   `json:"is_external"`
	IsDofollow  bool   `json:"is_dofollow"`
	IsImageLink bool   `json:"is_image_link"`
	IsAbsolute  bool   `json:"is_absolute"`
	IsSocial    bool   `json:"is_social"`
	IsUgc       bool   `json:"is_ugc"`
	IsSponsored bool   `json:"is_sponsored"`
}

// Image represents an image extracted from the page
type Image struct {
	Src        string `json:"src"` // Always resolved to absolute URL
	Alt        string `json:"alt"`
	IsExternal bool   `json:"is_external"`
	IsAbsolute bool   `json:"is_absolute"` // Was original src absolute?
	IsInLink   bool   `json:"is_in_link"`
	LinkHref   string `json:"link_href,omitempty"`
	Size       int    `json:"size"` // bytes from network request, 0 if not found
}

// RenderResponse represents the API response
type RenderResponse struct {
	Success bool         `json:"success"`
	Data    *RenderData  `json:"data,omitempty"`
	Error   *RenderError `json:"error,omitempty"`
}

// RenderData contains all rendered page data
type RenderData struct {
	// Technical information
	StatusCode    int     `json:"status_code"`
	FinalURL      string  `json:"final_url"`
	RedirectURL   string  `json:"redirect_url,omitempty"` // Set when redirect detected but not followed
	CanonicalURL  string  `json:"canonical_url,omitempty"`
	PageSizeBytes int     `json:"page_size_bytes"`
	RenderTime    float64 `json:"render_time"`
	ScreenshotID  string  `json:"screenshot_id,omitempty"`
	MetaRobots    string  `json:"meta_robots,omitempty"`
	XRobotsTag    string  `json:"x_robots_tag,omitempty"`

	// Robots directives (parsed)
	MetaIndexable bool `json:"meta_indexable"`
	MetaFollow    bool `json:"meta_follow"`

	// Content information
	Title               string            `json:"title"`
	MetaDescription     string            `json:"meta_description,omitempty"`
	H1                  []string          `json:"h1,omitempty"`
	H2                  []string          `json:"h2,omitempty"`
	H3                  []string          `json:"h3,omitempty"`
	WordCount           int               `json:"word_count"`
	BodyTextTokensCount int               `json:"body_text_tokens_count"`
	OpenGraph           map[string]string `json:"open_graph,omitempty"`
	StructuredData      []json.RawMessage `json:"structured_data,omitempty"`

	// Body text and ratio
	BodyText      string  `json:"body_text,omitempty"`
	BodyMarkdown  string  `json:"body_markdown,omitempty"`
	TextHtmlRatio float64 `json:"text_html_ratio"`

	// HrefLang alternates
	HrefLangs []HrefLang `json:"hreflang,omitempty"`

	// Links with metadata
	Links []Link `json:"links,omitempty"`

	// Images with metadata
	Images []Image `json:"images,omitempty"`

	// Network information
	Requests []NetworkRequest `json:"requests,omitempty"`

	// Lifecycle timing
	Lifecycle []LifecycleEvent `json:"lifecycle,omitempty"`

	// Console and errors
	Console  []ConsoleMessage `json:"console,omitempty"`
	JSErrors []JSError        `json:"js_errors,omitempty"`

	// Raw HTML
	HTML string `json:"html,omitempty"`

	ScreenshotData []byte `json:"-"` // Raw screenshot bytes, not serialized to JSON
}

// NetworkRequest represents a single network request
type NetworkRequest struct {
	ID         string  `json:"id"`
	URL        string  `json:"url"`
	Method     string  `json:"method"`
	Status     int     `json:"status"`
	Type       string  `json:"type"`
	Size       int     `json:"size"`
	Time       float64 `json:"time"` // seconds
	IsInternal bool    `json:"is_internal"`
	Blocked    bool    `json:"blocked,omitempty"`
	Failed     bool    `json:"failed,omitempty"`
}

// LifecycleEvent represents a single lifecycle timing event
type LifecycleEvent struct {
	Event string  `json:"event"`
	Time  float64 `json:"time"` // seconds
}

// ConsoleMessage represents a console log entry
type ConsoleMessage struct {
	ID      string  `json:"id"`
	Level   string  `json:"level"`
	Message string  `json:"message"`
	Time    float64 `json:"time"` // seconds since render start
}

// JSError represents a JavaScript error
type JSError struct {
	Message    string  `json:"message"`
	Source     string  `json:"source,omitempty"`
	Line       int     `json:"line,omitempty"`
	Column     int     `json:"column,omitempty"`
	StackTrace string  `json:"stack_trace,omitempty"`
	Timestamp  float64 `json:"timestamp"`
}

// RenderError represents an API error
type RenderError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// Section represents a content section extracted from the page
type Section struct {
	SectionID       string `json:"section_id"`
	HeadingLevel    int    `json:"heading_level"`
	HeadingText     string `json:"heading_text"`
	BodyMarkdown    string `json:"body_markdown"`
	DetectionMethod string `json:"detection_method,omitempty"`
}

// ExtRenderData contains rendered page data for the external API.
// Pointer types are used for optional content fields so nil = omitted from JSON,
// while empty string = present but empty.
type ExtRenderData struct {
	// Always-included metadata fields
	StatusCode      int               `json:"status_code"`
	FinalURL        string            `json:"final_url"`
	RedirectURL     string            `json:"redirect_url"`
	CanonicalURL    string            `json:"canonical_url"`
	PageSizeBytes   int               `json:"page_size_bytes"`
	RenderTime      float64           `json:"render_time"`
	MetaRobots      string            `json:"meta_robots"`
	XRobotsTag      string            `json:"x_robots_tag"`
	MetaIndexable   bool              `json:"meta_indexable"`
	MetaFollow      bool              `json:"meta_follow"`
	Title           string            `json:"title"`
	MetaDescription string            `json:"meta_description"`
	H1              []string          `json:"h1"`
	H2              []string          `json:"h2"`
	H3              []string          `json:"h3"`
	WordCount       int               `json:"word_count"`
	TextHtmlRatio   float64           `json:"text_html_ratio"`
	OpenGraph       map[string]string `json:"open_graph"`
	HrefLangs       []HrefLang        `json:"hreflang"`

	// Opt-in content fields (pointer types: nil = omitted, non-nil = present)
	HTML                *string           `json:"html,omitempty"`
	BodyText            *string           `json:"body_text,omitempty"`
	BodyTextTokensCount *int              `json:"body_text_tokens_count,omitempty"`
	BodyMarkdown        *string           `json:"body_markdown,omitempty"`
	Sections            []Section         `json:"sections,omitempty"`
	Links               []Link            `json:"links,omitempty"`
	Images              []Image           `json:"images,omitempty"`
	StructuredData      []json.RawMessage `json:"structured_data,omitempty"`
	Screenshot          *string           `json:"screenshot,omitempty"`
}

// ExtRenderResponse represents the external API response
type ExtRenderResponse struct {
	Success bool           `json:"success"`
	Data    *ExtRenderData `json:"data,omitempty"`
	Error   *RenderError   `json:"error,omitempty"`
}

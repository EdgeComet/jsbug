package types

// WaitEvent constants
const (
	WaitDOMContentLoaded  = "DOMContentLoaded"
	WaitLoad              = "load"
	WaitNetworkIdle       = "networkIdle"
	WaitNetworkAlmostIdle = "networkAlmostIdle"
)

// UserAgent preset constants
const (
	UserAgentChrome          = "chrome"
	UserAgentFirefox         = "firefox"
	UserAgentSafari          = "safari"
	UserAgentMobile          = "mobile"
	UserAgentBot             = "bot"
	UserAgentGooglebot       = "googlebot"
	UserAgentGooglebotMobile = "googlebot-mobile"
	UserAgentBingbot         = "bingbot"
	UserAgentClaudeBot       = "claudebot"
	UserAgentClaudeUser      = "claude-user"
	UserAgentChatGPTUser     = "chatgpt-user"
	UserAgentGPTBot          = "gptbot"
)

// UserAgentPresets maps preset names to full user agent strings
var UserAgentPresets = map[string]string{
	UserAgentChrome:          "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
	UserAgentFirefox:         "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:121.0) Gecko/20100101 Firefox/121.0",
	UserAgentSafari:          "Mozilla/5.0 (Macintosh; Intel Mac OS X 14_2) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.2 Safari/605.1.15",
	UserAgentMobile:          "Mozilla/5.0 (iPhone; CPU iPhone OS 17_2 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.2 Mobile/15E148 Safari/604.1",
	UserAgentBot:             "Mozilla/5.0 (compatible; Googlebot/2.1; +http://www.google.com/bot.html)",
	UserAgentGooglebot:       "Mozilla/5.0 (compatible; Googlebot/2.1; +http://www.google.com/bot.html)",
	UserAgentGooglebotMobile: "Mozilla/5.0 (Linux; Android 6.0.1; Nexus 5X Build/MMB29P) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Mobile Safari/537.36 (compatible; Googlebot/2.1; +http://www.google.com/bot.html)",
	UserAgentBingbot:         "Mozilla/5.0 (compatible; bingbot/2.0; +http://www.bing.com/bingbot.htm)",
	UserAgentClaudeBot:       "Mozilla/5.0 AppleWebKit/537.36 (KHTML, like Gecko; compatible; ClaudeBot/1.0; +claudebot@anthropic.com)",
	UserAgentClaudeUser:      "Mozilla/5.0 AppleWebKit/537.36 (KHTML, like Gecko; compatible; Claude-User/1.0; +Claude-User@anthropic.com)",
	UserAgentChatGPTUser:     "Mozilla/5.0 AppleWebKit/537.36 (KHTML, like Gecko); compatible; ChatGPT-User/1.0; +https://openai.com/bot",
	UserAgentGPTBot:          "Mozilla/5.0 AppleWebKit/537.36 (KHTML, like Gecko); compatible; GPTBot/1.3; +https://openai.com/gptbot",
}

// ValidWaitEvents contains all valid wait event values
var ValidWaitEvents = map[string]bool{
	WaitDOMContentLoaded:  true,
	WaitLoad:              true,
	WaitNetworkIdle:       true,
	WaitNetworkAlmostIdle: true,
}

// Default values
const (
	DefaultUserAgent = UserAgentChrome
	DefaultTimeout   = 15
	DefaultWaitEvent = WaitLoad
	MinTimeout       = 1
	MaxTimeout       = 60
)

// RenderRequest represents an API request to render a page
type RenderRequest struct {
	RequestID       string   `json:"request_id"`
	URL             string   `json:"url"`
	JSEnabled       bool     `json:"js_enabled"`
	FollowRedirects *bool    `json:"follow_redirects,omitempty"` // default true
	UserAgent       string   `json:"user_agent,omitempty"`
	Timeout         int      `json:"timeout,omitempty"`
	WaitEvent       string   `json:"wait_event,omitempty"`
	BlockAnalytics  bool     `json:"block_analytics,omitempty"`
	BlockAds        bool     `json:"block_ads,omitempty"`
	BlockSocial     bool     `json:"block_social,omitempty"`
	BlockedTypes      []string `json:"blocked_types,omitempty"`
	CaptureScreenshot bool     `json:"-"`                       // Internal only, not JSON-exposed
	SessionToken      string   `json:"session_token,omitempty"`
}

// ExtRenderRequest represents an external API request with content inclusion options
type ExtRenderRequest struct {
	URL             string   `json:"url"`
	JSEnabled       bool     `json:"js_enabled"`
	FollowRedirects *bool    `json:"follow_redirects,omitempty"`
	UserAgent       string   `json:"user_agent"`
	Timeout         int      `json:"timeout"`
	WaitEvent       string   `json:"wait_event"`
	BlockAnalytics  bool     `json:"block_analytics"`
	BlockAds        bool     `json:"block_ads"`
	BlockSocial     bool     `json:"block_social"`
	BlockedTypes    []string `json:"blocked_types"`

	IncludeHTML           bool `json:"include_html"`
	IncludeText           bool `json:"include_text"`
	IncludeMarkdown       bool `json:"include_markdown"`
	IncludeSections       bool `json:"include_sections"`
	IncludeLinks          bool `json:"include_links"`
	IncludeImages         bool `json:"include_images"`
	IncludeStructuredData bool `json:"include_structured_data"`
	IncludeScreenshot     bool `json:"include_screenshot"`

	MaxContentLength int `json:"max_content_length"`
}

// ToRenderRequest converts an ExtRenderRequest to a RenderRequest
func (e *ExtRenderRequest) ToRenderRequest() *RenderRequest {
	followRedirects := true
	if e.FollowRedirects != nil {
		followRedirects = *e.FollowRedirects
	}
	req := &RenderRequest{
		URL:               e.URL,
		JSEnabled:         e.JSEnabled,
		FollowRedirects:   &followRedirects,
		UserAgent:         e.UserAgent,
		Timeout:           e.Timeout,
		WaitEvent:         e.WaitEvent,
		BlockAnalytics:    e.BlockAnalytics,
		BlockAds:          e.BlockAds,
		BlockSocial:       e.BlockSocial,
		BlockedTypes:      e.BlockedTypes,
		CaptureScreenshot: e.IncludeScreenshot,
	}
	return req
}

// ShouldFollowRedirects returns whether to follow redirects (default true)
func (r *RenderRequest) ShouldFollowRedirects() bool {
	if r.FollowRedirects == nil {
		return true
	}
	return *r.FollowRedirects
}

// ResolveUserAgent returns the full user agent string for a preset or the custom value
func ResolveUserAgent(preset string) string {
	if preset == "" {
		return UserAgentPresets[DefaultUserAgent]
	}
	if ua, ok := UserAgentPresets[preset]; ok {
		return ua
	}
	return preset
}

// IsValidWaitEvent checks if the given wait event is valid
func IsValidWaitEvent(event string) bool {
	if event == "" {
		return true
	}
	return ValidWaitEvents[event]
}

// ApplyDefaults applies default values to a RenderRequest
func (r *RenderRequest) ApplyDefaults() {
	if r.UserAgent == "" {
		r.UserAgent = DefaultUserAgent
	}
	if r.Timeout == 0 {
		r.Timeout = DefaultTimeout
	}
	if r.WaitEvent == "" {
		r.WaitEvent = DefaultWaitEvent
	}
}

// ValidateTimeout checks if the timeout is within valid range
func (r *RenderRequest) ValidateTimeout() bool {
	return r.Timeout >= MinTimeout && r.Timeout <= MaxTimeout
}

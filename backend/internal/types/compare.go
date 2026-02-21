package types

// ExtCompareRequest represents an external API request to compare JS-rendered vs non-JS versions of a page.
type ExtCompareRequest struct {
	URL             string   `json:"url"`
	FollowRedirects *bool    `json:"follow_redirects,omitempty"`
	UserAgent       string   `json:"user_agent"`
	Timeout         int      `json:"timeout"`
	WaitEvent       string   `json:"wait_event"`
	BlockAnalytics  bool     `json:"block_analytics"`
	BlockAds        bool     `json:"block_ads"`
	BlockSocial     bool     `json:"block_social"`
	BlockedTypes    []string `json:"blocked_types"`

	MaxContentLength int  `json:"max_content_length"`
	MaxDiffLength    int  `json:"max_diff_length"`
	IncludeHTML      bool `json:"include_html"`
	IncludeText      bool `json:"include_text"`
	IncludeMarkdown  bool `json:"include_markdown"`
	IncludeSections  bool `json:"include_sections"`
	IncludeLinks     bool `json:"include_links"`
	IncludeImages    bool `json:"include_images"`

	IncludeStructuredData bool `json:"include_structured_data"`
}

// ToJSRenderRequest converts an ExtCompareRequest to a RenderRequest for JS-enabled rendering.
func (e *ExtCompareRequest) ToJSRenderRequest() *RenderRequest {
	followRedirects := true
	if e.FollowRedirects != nil {
		followRedirects = *e.FollowRedirects
	}
	return &RenderRequest{
		URL:               e.URL,
		JSEnabled:         true,
		FollowRedirects:   &followRedirects,
		UserAgent:         e.UserAgent,
		Timeout:           e.Timeout,
		WaitEvent:         e.WaitEvent,
		BlockAnalytics:    e.BlockAnalytics,
		BlockAds:          e.BlockAds,
		BlockSocial:       e.BlockSocial,
		BlockedTypes:      e.BlockedTypes,
		CaptureScreenshot: false,
	}
}

// ToHTTPRenderRequest converts an ExtCompareRequest to a RenderRequest for non-JS HTTP fetching.
// Block* fields are not copied because they are JS-only features.
func (e *ExtCompareRequest) ToHTTPRenderRequest() *RenderRequest {
	followRedirects := true
	if e.FollowRedirects != nil {
		followRedirects = *e.FollowRedirects
	}
	return &RenderRequest{
		URL:               e.URL,
		JSEnabled:         false,
		FollowRedirects:   &followRedirects,
		UserAgent:         e.UserAgent,
		Timeout:           e.Timeout,
		WaitEvent:         e.WaitEvent,
		CaptureScreenshot: false,
	}
}

// ExtCompareResponse represents the external API response for a compare request.
type ExtCompareResponse struct {
	Success bool             `json:"success"`
	Data    *ExtCompareData  `json:"data,omitempty"`
	Error   *RenderError     `json:"error,omitempty"`
}

// ExtCompareData contains the comparison results between JS-rendered and non-JS versions.
type ExtCompareData struct {
	JSStatus        *FetchStatus     `json:"js_status"`
	HTTPStatus      *FetchStatus     `json:"http_status"`
	JS              *ExtRenderData   `json:"js"`
	Diff            *CompareDiff     `json:"diff"`
	RenderingImpact *RenderingImpact `json:"rendering_impact"`
}

// FetchStatus represents the outcome of a single fetch (JS or HTTP).
type FetchStatus struct {
	Success    bool         `json:"success"`
	StatusCode int          `json:"status_code,omitempty"`
	RenderTime float64      `json:"render_time,omitempty"`
	Error      *RenderError `json:"error,omitempty"`
}

// CompareDiff contains all detected differences between JS-rendered and non-JS versions.
type CompareDiff struct {
	// Metadata diffs (always computed when both succeed, omitted when identical)
	Title           *StringDiff      `json:"title,omitempty"`
	MetaDescription *StringDiff      `json:"meta_description,omitempty"`
	CanonicalURL    *StringDiff      `json:"canonical_url,omitempty"`
	MetaRobots      *StringDiff      `json:"meta_robots,omitempty"`
	H1              *StringSliceDiff `json:"h1,omitempty"`
	H2              *StringSliceDiff `json:"h2,omitempty"`
	H3              *StringSliceDiff `json:"h3,omitempty"`
	WordCountJS     int              `json:"word_count_js"`
	WordCountNonJS  int              `json:"word_count_non_js"`

	// Content diffs (only present when corresponding include flag is set)
	Sections       []SectionDiff       `json:"sections,omitempty"`
	Links          *LinksDiff          `json:"links,omitempty"`
	Images         *ImagesDiff         `json:"images,omitempty"`
	StructuredData *StructuredDataDiff `json:"structured_data,omitempty"`
}

// StringDiff represents a difference in a single string field between JS and non-JS versions.
type StringDiff struct {
	JSValue    string `json:"js_value"`
	NonJSValue string `json:"non_js_value"`
}

// StringSliceDiff represents differences in a string slice field between JS and non-JS versions.
type StringSliceDiff struct {
	Added   []string `json:"added"`
	Removed []string `json:"removed"`
}

// SectionDiff represents a difference in a content section between JS and non-JS versions.
type SectionDiff struct {
	SectionID        string `json:"section_id"`
	HeadingLevel     int    `json:"heading_level"`
	HeadingText      string `json:"heading_text"`
	Status           string `json:"status"` // "unchanged", "changed", "added_by_js", "removed_by_js"
	NonJSBodyMarkdown string `json:"non_js_body_markdown,omitempty"`
}

// LinksDiff represents differences in extracted links between JS and non-JS versions.
type LinksDiff struct {
	JSCount    int    `json:"js_count"`
	NonJSCount int    `json:"non_js_count"`
	Added      []Link `json:"added"`
	Removed    []Link `json:"removed"`
}

// ImagesDiff represents differences in extracted images between JS and non-JS versions.
type ImagesDiff struct {
	JSCount    int     `json:"js_count"`
	NonJSCount int     `json:"non_js_count"`
	Added      []Image `json:"added"`
	Removed    []Image `json:"removed"`
}

// StructuredDataDiff represents differences in structured data between JS and non-JS versions.
type StructuredDataDiff struct {
	Added   []string `json:"added"`
	Removed []string `json:"removed"`
	Changed []string `json:"changed"`
}

// RenderingImpact summarizes the overall impact of JavaScript rendering on the page.
type RenderingImpact struct {
	OverallChange        string  `json:"overall_change"` // "none", "minor", "major"
	TitleChanged         bool    `json:"title_changed"`
	MetaDescChanged      bool    `json:"meta_desc_changed"`
	CanonicalChanged     bool    `json:"canonical_changed"`
	H1Changed            bool    `json:"h1_changed"`
	ContentChangePercent float64 `json:"content_change_percent"`
	WordCountJS          int     `json:"word_count_js"`
	WordCountNonJS       int     `json:"word_count_non_js"`
	LinksAdded           int     `json:"links_added"`
	LinksRemoved         int     `json:"links_removed"`
	ImagesAdded          int     `json:"images_added"`
	ImagesRemoved        int     `json:"images_removed"`
	StructuredDataAdded  int     `json:"structured_data_added"`
	StructuredDataRemoved int    `json:"structured_data_removed"`
}

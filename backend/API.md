# jsbug Public API

Programmatic access to jsbug's page rendering and content extraction. Returns synchronous JSON responses authenticated via API key.

## Endpoints

```
POST /api/ext/render    - Render a page and extract content
POST /api/ext/compare   - Compare JS-rendered vs non-JS versions of a page
```

Both endpoints use the same authentication, user agent presets, wait events, and validation rules.

## Authentication

All endpoints require an `X-API-Key` header. API keys are configured in `config.yaml` or via environment variable.

**Config file:**
```yaml
api:
  enabled: true
  keys:
    - "your-key-here"
```

**Environment variable:**
```
JSBUG_API_KEYS=key1,key2,key3
```

Setting `JSBUG_API_KEYS` automatically enables the API. Empty segments are filtered out.

Keys are validated with constant-time comparison to prevent timing attacks.

## User Agent Presets

| Preset | Description |
|--------|-------------|
| `chrome` | Chrome 120 on Windows (default) |
| `firefox` | Firefox 121 on Windows |
| `safari` | Safari 17.2 on macOS |
| `mobile` | Safari on iPhone |
| `bot` | Generic bot (Googlebot UA) |
| `googlebot` | Googlebot desktop |
| `googlebot-mobile` | Googlebot mobile |
| `bingbot` | Bingbot |
| `claudebot` | ClaudeBot |
| `claude-user` | Claude-User |
| `chatgpt-user` | ChatGPT-User |
| `gptbot` | GPTBot |

Any string not matching a preset is used as-is (custom user agent).

## Wait Events (JS mode only)

| Event | Behavior |
|-------|----------|
| `DOMContentLoaded` | DOM ready, subresources may still load |
| `load` | Window load event fired (default) |
| `networkIdle` | No network activity for 500ms |
| `networkAlmostIdle` | Fewer than 2 network requests for 500ms |

## Shared Object Types

### Section Object

```json
{
  "section_id": "s1",
  "heading_level": 2,
  "heading_text": "Getting Started",
  "body_markdown": "Follow these steps:\n\n- Install the package\n- Run the setup command"
}
```

| Field | Type | Description |
|-------|------|-------------|
| `section_id` | string | Sequential ID: `s1`, `s2`, `s3`, ... |
| `heading_level` | int | 1-6 for h1-h6, 0 for intro text before first heading |
| `heading_text` | string | Heading text (empty for intro section) |
| `body_markdown` | string | Markdown content between this heading and the next |

Sections are split at every heading boundary (flat list, not hierarchical). Empty sections (no heading text and no body) are omitted.

### Link Object

| Field | Type | Description |
|-------|------|-------------|
| `href` | string | Link URL |
| `text` | string | Anchor text |
| `is_external` | bool | Points to a different domain |
| `is_dofollow` | bool | No `nofollow` rel attribute |
| `is_image_link` | bool | Contains an image element |
| `is_absolute` | bool | URL is absolute |
| `is_social` | bool | Points to a social media platform |
| `is_ugc` | bool | Has `rel="ugc"` |
| `is_sponsored` | bool | Has `rel="sponsored"` |

### Image Object

| Field | Type | Description |
|-------|------|-------------|
| `src` | string | Absolute URL of the image |
| `alt` | string | Alt text |
| `is_external` | bool | Hosted on a different domain |
| `is_absolute` | bool | Original src was absolute |
| `is_in_link` | bool | Image is wrapped in a link |
| `link_href` | string | Parent link URL (if `is_in_link`) |
| `size` | int | Size in bytes from network request (0 if unknown) |

## Error Codes

Both endpoints share the same error response format and validation error codes.

### Error Response Format

```json
{
  "success": false,
  "error": {
    "code": "INVALID_URL",
    "message": "URL is required"
  }
}
```

### Validation Errors

| Error Code | HTTP Status | Condition |
|------------|-------------|-----------|
| `API_KEY_REQUIRED` | 401 | Missing `X-API-Key` header |
| `API_KEY_INVALID` | 403 | Key not in allowed list |
| `METHOD_NOT_ALLOWED` | 405 | Non-POST request |
| `INVALID_REQUEST_BODY` | 400 | Malformed JSON, empty body, or unknown fields |
| `INVALID_URL` | 400 | Missing, malformed, or non-http(s) URL |
| `INVALID_TIMEOUT` | 400 | Timeout outside 1-60 range |
| `INVALID_WAIT_EVENT` | 400 | Unknown wait event value |
| `SSRF_BLOCKED` | 403 | Private/internal IP address blocked |
| `DOMAIN_NOT_FOUND` | 400 | DNS resolution failure |

### Runtime Errors (render endpoint only)

| Error Code | HTTP Status | Condition |
|------------|-------------|-----------|
| `RENDER_TIMEOUT` | 408 | Render or fetch exceeded timeout |
| `RENDER_FAILED` | 500 | Chrome rendering error |
| `FETCH_FAILED` | 500 | HTTP fetch error |
| `CHROME_UNAVAILABLE` | 503 | Chrome pool not initialized |
| `POOL_EXHAUSTED` | 503 | All Chrome instances busy |
| `POOL_SHUTTING_DOWN` | 503 | Server shutting down |

**Validation order:** HTTP method -> API key -> body size (1MB max) -> JSON parsing (unknown fields rejected) -> URL -> timeout -> wait event.

---

## POST /api/ext/render

Renders a page via Chrome (JS mode) or HTTP fetch (non-JS mode) and extracts structured content.

```
POST /api/ext/render
Content-Type: application/json
X-API-Key: <your-api-key>
```

### Request

#### Render Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `url` | string | *required* | Target URL (http or https) |
| `js_enabled` | bool | `false` | `true` = Chrome rendering, `false` = HTTP fetch |
| `follow_redirects` | bool | `true` | Follow HTTP redirects (up to 10 hops) |
| `user_agent` | string | `"chrome"` | Preset name or custom UA string |
| `timeout` | int | `15` | Render/fetch timeout in seconds (1-60) |
| `wait_event` | string | `"load"` | JS mode wait condition |
| `block_analytics` | bool | `false` | Block analytics scripts (Google Analytics, etc.) |
| `block_ads` | bool | `false` | Block ad network scripts |
| `block_social` | bool | `false` | Block social media scripts |
| `blocked_types` | string[] | `[]` | Resource types to block: `image`, `font`, `stylesheet`, `script`, `xhr`, `fetch` |
| `max_content_length` | int | `0` | Max characters per content field. `0` = no limit. Truncates at word boundary. |

#### Content Include Flags

All default to `false`. When `false`, the corresponding field is absent from the response.

| Flag | Adds to Response |
|------|-----------------|
| `include_html` | `html` - full HTML document |
| `include_text` | `body_text` - extracted plain text; `body_text_tokens_count` - token count |
| `include_markdown` | `body_markdown` - markdown conversion |
| `include_sections` | `sections` - heading-delimited content sections |
| `include_links` | `links` - extracted links with metadata |
| `include_images` | `images` - extracted images with src, alt, size |
| `include_structured_data` | `structured_data` - JSON-LD blocks |
| `include_screenshot` | `screenshot` - base64-encoded PNG (JS mode only, ignored in HTTP mode) |

### Response

#### Success Response

```json
{
  "success": true,
  "data": {
    "status_code": 200,
    "final_url": "https://example.com/",
    "redirect_url": "",
    "canonical_url": "https://example.com/",
    "page_size_bytes": 1256,
    "render_time": 0.45,
    "meta_robots": "",
    "x_robots_tag": "",
    "meta_indexable": true,
    "meta_follow": true,
    "title": "Example Domain",
    "meta_description": "",
    "h1": ["Example Domain"],
    "h2": [],
    "h3": [],
    "word_count": 28,
    "text_html_ratio": 0.0422,
    "open_graph": {},
    "hreflang": []
  }
}
```

#### Always-Included Fields

These fields are present in every successful response regardless of include flags:

| Field | Type | Description |
|-------|------|-------------|
| `status_code` | int | HTTP status code |
| `final_url` | string | URL after redirects |
| `redirect_url` | string | Redirect target (if not followed) |
| `canonical_url` | string | Canonical URL from `<link>` or Link header |
| `page_size_bytes` | int | Response body size in bytes |
| `render_time` | float | Time to render/fetch in seconds |
| `meta_robots` | string | Raw meta robots content |
| `x_robots_tag` | string | X-Robots-Tag header value |
| `meta_indexable` | bool | Parsed from robots directives |
| `meta_follow` | bool | Parsed from robots directives |
| `title` | string | Page `<title>` |
| `meta_description` | string | Meta description |
| `h1` | string[] | H1 heading texts |
| `h2` | string[] | H2 heading texts |
| `h3` | string[] | H3 heading texts |
| `word_count` | int | Word count of body text |
| `text_html_ratio` | float | Text-to-HTML size ratio |
| `open_graph` | object | OpenGraph meta tags |
| `hreflang` | HrefLang[] | hreflang alternates (`lang`, `url`, `source`) |

#### Opt-In Content Fields

Present only when the corresponding `include_*` flag is `true`:

| Field | Type | Flag |
|-------|------|------|
| `html` | string | `include_html` |
| `body_text` | string | `include_text` |
| `body_text_tokens_count` | int | `include_text` |
| `body_markdown` | string | `include_markdown` |
| `sections` | Section[] | `include_sections` |
| `links` | Link[] | `include_links` |
| `images` | Image[] | `include_images` |
| `structured_data` | json[] | `include_structured_data` |
| `screenshot` | string | `include_screenshot` |

When a field is requested but the page has no content for it, the field is present with an empty value (empty string, empty array). When not requested, the field is absent from the JSON.

### Content Truncation

When `max_content_length` > 0, each content field is independently truncated to that many characters at the nearest word boundary:

- `html`, `body_text`, `body_markdown` - truncated independently
- `sections` - sections are filled in order; once the character budget is exhausted, remaining sections are dropped. A section that exceeds the remaining budget is truncated at a word boundary.
- Token counts (`body_text_tokens_count`) are computed after truncation.

### Examples

**Metadata only (minimal request):**
```bash
curl -s -X POST http://localhost:9301/api/ext/render \
  -H "Content-Type: application/json" \
  -H "X-API-Key: YOUR_KEY" \
  -d '{"url": "https://example.com"}'
```

**Content extraction (HTTP mode):**
```bash
curl -s -X POST http://localhost:9301/api/ext/render \
  -H "Content-Type: application/json" \
  -H "X-API-Key: YOUR_KEY" \
  -d '{
    "url": "https://example.com",
    "include_text": true,
    "include_markdown": true,
    "include_sections": true,
    "include_links": true
  }'
```

**JS rendering with screenshot:**
```bash
curl -s -X POST http://localhost:9301/api/ext/render \
  -H "Content-Type: application/json" \
  -H "X-API-Key: YOUR_KEY" \
  -d '{
    "url": "https://example.com",
    "js_enabled": true,
    "timeout": 20,
    "wait_event": "networkIdle",
    "include_html": true,
    "include_screenshot": true
  }'
```

**SEO audit (bot UA, resource blocking, content extraction):**
```bash
curl -s -X POST http://localhost:9301/api/ext/render \
  -H "Content-Type: application/json" \
  -H "X-API-Key: YOUR_KEY" \
  -d '{
    "url": "https://example.com",
    "js_enabled": true,
    "user_agent": "googlebot",
    "block_analytics": true,
    "block_ads": true,
    "include_text": true,
    "include_links": true,
    "include_images": true,
    "include_structured_data": true,
    "max_content_length": 50000
  }'
```

### Implementation Notes

- Separate handler from the internal `/api/render` (no SSE streaming, no session tokens, no request_id tracking).
- Shares the same Chrome pool as internal requests.
- Screenshots are JS mode only. In HTTP mode, `include_screenshot` is ignored and the field is absent.
- Screenshot is returned inline as base64-encoded PNG (not stored in the screenshot store).
- Strict JSON parsing: unknown fields in the request body return 400 `INVALID_REQUEST_BODY`.
- Request body is capped at 1MB.
- Sections are extracted by re-parsing the HTML with goquery (not computed during the main render pipeline).

---

## POST /api/ext/compare

Fetches a URL twice in parallel (JS-rendered via Chrome, plain HTTP without JS) and returns a structured diff showing what JavaScript changes on the page. The JS-rendered content is returned as primary data, with a diff overlay and a rendering impact summary.

```
POST /api/ext/compare
Content-Type: application/json
X-API-Key: <your-api-key>
```

### Request

#### Compare Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `url` | string | *required* | Target URL (http or https) |
| `follow_redirects` | bool | `true` | Follow HTTP redirects (up to 10 hops) |
| `user_agent` | string | `"chrome"` | Preset name or custom UA string. Applied to both fetches. |
| `timeout` | int | `15` | Render/fetch timeout in seconds (1-60). Applied to each fetch independently. |
| `wait_event` | string | `"load"` | JS mode wait condition |
| `block_analytics` | bool | `false` | Block analytics scripts (JS fetch only) |
| `block_ads` | bool | `false` | Block ad network scripts (JS fetch only) |
| `block_social` | bool | `false` | Block social media scripts (JS fetch only) |
| `blocked_types` | string[] | `[]` | Resource types to block: `image`, `font`, `stylesheet`, `script`, `xhr`, `fetch` (JS fetch only) |
| `max_content_length` | int | `0` | Max characters for primary JS content fields. `0` = no limit. Truncates at word boundary. |
| `max_diff_length` | int | `0` | Max characters for diff overlay text content. `0` = no limit. Truncates at word boundary. |

**Not available** (differs from `/api/ext/render`):
- `js_enabled` - always runs both modes, not configurable
- `include_screenshot` - not supported in compare mode

#### Content Include Flags

All default to `false`. These control what appears in the primary JS content AND which diff categories are computed.

| Flag | Primary Content | Diff Computed |
|------|----------------|---------------|
| `include_html` | `js.html` | No diff (too noisy) |
| `include_text` | `js.body_text` + `js.body_text_tokens_count` | No diff (use sections instead) |
| `include_markdown` | `js.body_markdown` | No diff (use sections instead) |
| `include_sections` | `js.sections` | `diff.sections` - section-level diff |
| `include_links` | `js.links` | `diff.links` - added/removed link lists |
| `include_images` | `js.images` | `diff.images` - added/removed image lists |
| `include_structured_data` | `js.structured_data` | `diff.structured_data` - presence-level by @type |

### Response

#### Top-Level Structure

```json
{
  "success": true,
  "data": {
    "js_status": { ... },
    "http_status": { ... },
    "js": { ... },
    "diff": { ... },
    "rendering_impact": { ... }
  }
}
```

HTTP status code is always 200 when the compare operation itself succeeds (even if individual fetches fail). Individual fetch failures are reported in `js_status` / `http_status`.

#### FetchStatus

Present for both `js_status` and `http_status`. Reports whether each individual fetch succeeded.

On success:
```json
{
  "success": true,
  "status_code": 200,
  "render_time": 1.23
}
```

On failure:
```json
{
  "success": false,
  "error": {
    "code": "RENDER_TIMEOUT",
    "message": "Chrome rendering timed out"
  }
}
```

| Field | Type | Description |
|-------|------|-------------|
| `success` | bool | Whether this individual fetch succeeded |
| `status_code` | int | HTTP status code of the fetched page (omitted on failure) |
| `render_time` | float | Time in seconds (omitted on failure) |
| `error` | object | Error details (omitted on success) |

#### JS Primary Content

The `js` field contains the full JS-rendered page data in the same format as `/api/ext/render` response `data`. Content fields are controlled by `include_*` flags. Truncated by `max_content_length`.

- When the JS fetch fails, `js` is `null`.
- When the JS fetch succeeds but HTTP fetch fails, `js` is populated but `diff` and `rendering_impact` are `null`.

#### Diff Overlay

Present only when both fetches succeed. `null` when either fetch fails.

**Metadata diffs** (always computed when both succeed, only emitted when values differ):

| Field | Type | When Present |
|-------|------|-------------|
| `title` | StringDiff | Title differs between versions |
| `meta_description` | StringDiff | Meta description differs |
| `canonical_url` | StringDiff | Canonical URL differs |
| `meta_robots` | StringDiff | Meta robots differs |
| `h1` | StringSliceDiff | H1 headings differ |
| `h2` | StringSliceDiff | H2 headings differ |
| `h3` | StringSliceDiff | H3 headings differ |
| `word_count_js` | int | Always present (context) |
| `word_count_non_js` | int | Always present (context) |

**StringDiff** format:
```json
{
  "js_value": "rendered title",
  "non_js_value": "loading..."
}
```

**StringSliceDiff** format (set-based comparison):
```json
{
  "added": ["heading in JS only"],
  "removed": ["heading in non-JS only"]
}
```

**Section diff** (requires `include_sections: true`):

Each section is assigned a status based on matching by heading text and level. Omitted from the diff when there are no section differences.

| Status | Meaning | `non_js_body_markdown` |
|--------|---------|----------------------|
| `unchanged` | Section exists in both with identical content | Omitted |
| `changed` | Section exists in both but content differs | Contains the non-JS version |
| `added_by_js` | Section exists only in JS-rendered version | Omitted |
| `removed_by_js` | Section exists only in non-JS version | Omitted |

```json
{
  "diff": {
    "sections": [
      {
        "section_id": "s1",
        "heading_level": 0,
        "heading_text": "",
        "status": "changed",
        "non_js_body_markdown": "Loading..."
      },
      {
        "section_id": "s2",
        "heading_level": 2,
        "heading_text": "Features",
        "status": "added_by_js"
      }
    ]
  }
}
```

**Links diff** (requires `include_links: true`):

Always present when `include_links` is set, even if both versions have identical links.

| Field | Type | Description |
|-------|------|-------------|
| `js_count` | int | Total link count in JS version |
| `non_js_count` | int | Total link count in non-JS version |
| `added` | Link[] | Links in JS but not non-JS (deduplicated by href, sorted alphabetically) |
| `removed` | Link[] | Links in non-JS but not JS (deduplicated by href, sorted alphabetically) |

**Images diff** (requires `include_images: true`):

Always present when `include_images` is set, even if both versions have identical images. Same structure as links diff, matching by `src`:

| Field | Type | Description |
|-------|------|-------------|
| `js_count` | int | Total image count in JS version |
| `non_js_count` | int | Total image count in non-JS version |
| `added` | Image[] | Images in JS but not non-JS (deduplicated by src, sorted alphabetically) |
| `removed` | Image[] | Images in non-JS but not JS (deduplicated by src, sorted alphabetically) |

**Structured data diff** (requires `include_structured_data: true`):

Presence-level comparison by JSON-LD `@type`. Omitted from the diff when there are no differences.

| Field | Type | Description |
|-------|------|-------------|
| `added` | string[] | `@type` values present in JS but not non-JS |
| `removed` | string[] | `@type` values present in non-JS but not JS |
| `changed` | string[] | `@type` values present in both but with different JSON content |

#### Rendering Impact

Present only when both fetches succeed. `null` when either fetch fails.

```json
{
  "rendering_impact": {
    "overall_change": "major",
    "title_changed": true,
    "meta_desc_changed": true,
    "canonical_changed": false,
    "h1_changed": true,
    "content_change_percent": 87.5,
    "word_count_js": 450,
    "word_count_non_js": 12,
    "links_added": 37,
    "links_removed": 2,
    "images_added": 15,
    "images_removed": 0,
    "structured_data_added": 2,
    "structured_data_removed": 0
  }
}
```

| Field | Type | Description |
|-------|------|-------------|
| `overall_change` | string | `"none"`, `"minor"`, or `"major"` |
| `title_changed` | bool | Title differs between versions |
| `meta_desc_changed` | bool | Meta description differs |
| `canonical_changed` | bool | Canonical URL differs |
| `h1_changed` | bool | H1 headings differ |
| `content_change_percent` | float | 0-100, based on word count difference |
| `word_count_js` | int | Word count from JS version |
| `word_count_non_js` | int | Word count from non-JS version |
| `links_added` | int | Links added by JS (0 if `include_links` not set) |
| `links_removed` | int | Links removed by JS (0 if `include_links` not set) |
| `images_added` | int | Images added by JS (0 if `include_images` not set) |
| `images_removed` | int | Images removed by JS (0 if `include_images` not set) |
| `structured_data_added` | int | JSON-LD types added by JS (0 if `include_structured_data` not set) |
| `structured_data_removed` | int | JSON-LD types removed by JS (0 if `include_structured_data` not set) |

**`overall_change` classification:**

- `"none"` - no metadata changed AND `content_change_percent < 5` AND no links/images/structured_data added or removed
- `"major"` - any of: `title_changed` OR `content_change_percent > 30` OR `links_added > 10` OR `images_added > 5`
- `"minor"` - everything else

### Error Handling

Validation errors use the same codes and HTTP statuses listed in [Error Codes](#error-codes). The validation order is identical to `/api/ext/render`.

Both fetches can fail independently. The overall response is still `success: true` with HTTP 200. Individual failures are reported in the `js_status` and `http_status` objects. When either fetch fails, `diff` and `rendering_impact` are `null`. When the JS fetch fails, `js` is also `null`.

### Content Truncation

Two independent truncation budgets:

**`max_content_length`** (primary JS content):
- `html`, `body_text`, `body_markdown` - each truncated independently at word boundary
- `sections` - filled in order; once budget exhausted, remaining sections dropped; last section truncated at word boundary
- `body_text_tokens_count` computed after truncation

**`max_diff_length`** (diff overlay):
- `SectionDiff.non_js_body_markdown` fields - shared budget, sections processed in order, same word-boundary truncation logic
- Metadata StringDiff values (title, meta_description, etc.) - not truncated (short metadata strings)
- Link/image added/removed arrays - not truncated (structured objects, not text)

### Examples

**Minimal request (metadata diff and rendering impact only):**
```bash
curl -s -X POST http://localhost:9301/api/ext/compare \
  -H "Content-Type: application/json" \
  -H "X-API-Key: YOUR_KEY" \
  -d '{"url": "https://example.com"}'
```

**Full comparison with content limits:**
```bash
curl -s -X POST http://localhost:9301/api/ext/compare \
  -H "Content-Type: application/json" \
  -H "X-API-Key: YOUR_KEY" \
  -d '{
    "url": "https://example.com",
    "include_sections": true,
    "include_links": true,
    "include_images": true,
    "include_structured_data": true,
    "max_content_length": 50000,
    "max_diff_length": 10000
  }'
```

### Implementation Notes

- Both fetches (JS via Chrome, HTTP via fetcher) run in parallel. Total time = max(jsTime, httpTime).
- The `timeout` field applies to each fetch independently (not cumulative).
- Shares the same Chrome pool as `/api/ext/render` and internal requests.
- Block fields (`block_analytics`, `block_ads`, `block_social`, `blocked_types`) apply to the JS fetch only. The HTTP fetch uses a plain HTTP client.
- Strict JSON parsing: unknown fields in the request body return 400 `INVALID_REQUEST_BODY`.
- Request body is capped at 1MB.
- Sections are extracted by re-parsing the HTML with goquery (same approach as `/api/ext/render`).

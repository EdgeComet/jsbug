# jsbug Public API

Programmatic access to jsbug's page rendering and content extraction. Returns synchronous JSON responses authenticated via API key.

## Endpoint

```
POST /api/ext/render
Content-Type: application/json
X-API-Key: <your-api-key>
```

## Authentication

API keys are configured in `config.yaml` or via environment variable.

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

## Request

### Render Fields

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

### Content Include Flags

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

### User Agent Presets

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

### Wait Events (JS mode only)

| Event | Behavior |
|-------|----------|
| `DOMContentLoaded` | DOM ready, subresources may still load |
| `load` | Window load event fired (default) |
| `networkIdle` | No network activity for 500ms |
| `networkAlmostIdle` | Fewer than 2 network requests for 500ms |

## Response

### Success Response

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

### Always-Included Fields

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

### Opt-In Content Fields

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

### Error Response

```json
{
  "success": false,
  "error": {
    "code": "INVALID_URL",
    "message": "URL is required"
  }
}
```

## Error Codes

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
| `RENDER_TIMEOUT` | 408 | Render or fetch exceeded timeout |
| `RENDER_FAILED` | 500 | Chrome rendering error |
| `FETCH_FAILED` | 500 | HTTP fetch error |
| `CHROME_UNAVAILABLE` | 503 | Chrome pool not initialized |
| `POOL_EXHAUSTED` | 503 | All Chrome instances busy |
| `POOL_SHUTTING_DOWN` | 503 | Server shutting down |

**Validation order:** HTTP method -> API key -> body size (1MB max) -> JSON parsing (unknown fields rejected) -> URL -> timeout -> wait event.

## Content Truncation

When `max_content_length` > 0, each content field is independently truncated to that many characters at the nearest word boundary:

- `html`, `body_text`, `body_markdown` - truncated independently
- `sections` - sections are filled in order; once the character budget is exhausted, remaining sections are dropped. A section that exceeds the remaining budget is truncated at a word boundary.
- Token counts (`body_text_tokens_count`) are computed after truncation.

## Examples

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

## Implementation Notes

- Separate handler from the internal `/api/render` (no SSE streaming, no session tokens, no request_id tracking).
- Shares the same Chrome pool as internal requests.
- Screenshots are JS mode only. In HTTP mode, `include_screenshot` is ignored and the field is absent.
- Screenshot is returned inline as base64-encoded PNG (not stored in the screenshot store).
- Strict JSON parsing: unknown fields in the request body return 400 `INVALID_REQUEST_BODY`.
- Request body is capped at 1MB.
- Sections are extracted by re-parsing the HTML with goquery (not computed during the main render pipeline).

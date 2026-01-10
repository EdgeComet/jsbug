# jsbug - JavaScript Page Renderer

A Go-based service for rendering JavaScript-heavy web pages using headless Chrome. Extracts SEO-relevant content including titles, meta tags, headings, structured data, and more.

## Features

- **JavaScript Rendering**: Renders pages with full JS execution using headless Chrome
- **HTTP Fetching**: Non-JS mode for simple HTML fetching
- **Content Extraction**: Extracts title, meta tags, headings, links, Open Graph, JSON-LD
- **Network Capture**: Records all network requests with timing and blocking info
- **Console Capture**: Logs console messages and JavaScript errors
- **SSE Progress**: Real-time progress updates via Server-Sent Events
- **Resource Blocking**: Block analytics, ads, social scripts, or specific resource types

## Prerequisites

- Go 1.22 or later
- Chrome/Chromium (for JS rendering mode)

## Installation

```bash
# Clone the repository
git clone <repository-url>
cd backend

# Install dependencies
go mod download

# Build the binary
go build -o bin/jsbug ./cmd/jsbug
```

## Configuration

Create a `config.yaml` file:

```yaml
server:
  host: "0.0.0.0"
  port: 9301
  timeout: 30              # Request timeout in seconds
  cors_origins:
    - "*"

chrome:
  pool_size: 4             # Number of Chrome instances (1-16)
  warmup_url: "https://example.com/"  # URL to load on instance start
  restart_after_count: 50  # Restart instance after N renders (0 = disabled)
  restart_after_time: 30m  # Restart instance after duration (0 = disabled)

logging:
  level: "info"            # debug, info, warn, error
  format: "console"        # json or console
  file_path: ""            # Optional: path to log file (empty = console only)

captcha:
  enabled: false           # Enable Cloudflare Turnstile verification
  secret_key: ""           # Turnstile secret key (required if enabled)
```

### Environment Variables

Configuration can be overridden with environment variables:

- `JSBUG_PORT` - Server port
- `JSBUG_POOL_SIZE` - Chrome pool size
- `JSBUG_LOG_LEVEL` - Log level (debug, info, warn, error)
- `JSBUG_CORS_ORIGINS` - CORS origins (comma-separated)
- `JSBUG_CAPTCHA_ENABLED` - Enable captcha (true/false)
- `JSBUG_CAPTCHA_SECRET_KEY` - Captcha secret key

## Usage

```bash
# Run with default config
./bin/jsbug

# Run with custom config
./bin/jsbug -c /path/to/config.yaml
```

## API

### Health Check

```
GET /health
```

Returns service health status.

### Render Page

```
POST /api/render
Content-Type: application/json

{
  "url": "https://example.com",
  "js_enabled": true,
  "user_agent": "chrome",
  "timeout": 15,
  "wait_event": "load",
  "block_analytics": true,
  "block_ads": true,
  "block_social": false,
  "blocked_types": ["font", "image"]
}
```

**Parameters:**

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| url | string | required | URL to render |
| js_enabled | bool | false | Enable JavaScript rendering |
| user_agent | string | "chrome" | User agent preset or custom string |
| timeout | int | 15 | Timeout in seconds (1-60) |
| wait_event | string | "load" | Event to wait for |
| block_analytics | bool | false | Block analytics scripts |
| block_ads | bool | false | Block ad scripts |
| block_social | bool | false | Block social media scripts |
| blocked_types | array | [] | Resource types to block |

**User Agent Presets:** chrome, firefox, safari, mobile, bot

**Wait Events:** DOMContentLoaded, load, networkIdle, networkAlmostIdle

**Response:**

```json
{
  "success": true,
  "data": {
    "status_code": 200,
    "final_url": "https://example.com/",
    "title": "Example Page",
    "meta_description": "Page description",
    "meta_robots": "index, follow",
    "canonical_url": "https://example.com/",
    "h1": ["Main Heading"],
    "h2": ["Section 1", "Section 2"],
    "h3": [],
    "word_count": 150,
    "internal_links": 5,
    "external_links": 2,
    "open_graph": {
      "title": "OG Title",
      "image": "https://example.com/image.jpg"
    },
    "structured_data": [...],
    "requests": [...],
    "network_summary": {
      "total_requests": 25,
      "blocked_requests": 5,
      "failed_requests": 0,
      "total_bytes": 150000
    },
    "lifecycle": {
      "dom_content_loaded_ms": 500,
      "load_ms": 1200
    },
    "console": [...],
    "js_errors": [...],
    "render_time": 1.5,
    "page_size_bytes": 50000,
    "html": "<!DOCTYPE html>..."
  }
}
```

### SSE Progress Stream

```
GET /api/render/stream?request_id=<id>
```

Streams progress events for a render request.

**Events:**

- `started` - Render began
- `navigating` - Navigating to URL
- `waiting` - Waiting for lifecycle event
- `capturing` - Capturing network events
- `parsing` - Parsing HTML content
- `complete` - Render complete
- `error` - Error occurred

## Development

### Run Tests

```bash
# Unit tests (no Chrome required)
go test ./internal/...

# All tests including integration (Chrome required)
go test -tags chrome ./...
```

### Build

```bash
make build
```

### Lint

```bash
go vet ./...
go fmt ./...
```

## Project Structure

```
backend/
├── cmd/
│   └── jsbug/
│       └── main.go           # Application entry point
├── internal/
│   ├── chrome/               # Chrome instance and rendering
│   ├── config/               # Configuration loading
│   ├── errors/               # Custom error types
│   ├── fetcher/              # HTTP fetching
│   ├── logger/               # Logging setup
│   ├── parser/               # HTML parsing
│   ├── server/               # HTTP server and handlers
│   └── types/                # Request/response types
├── tests/
│   └── integration/          # Integration tests
├── config.yaml               # Default configuration
├── go.mod
└── README.md
```

## License

MIT License

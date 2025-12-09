# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build & Test Commands

```bash
# Build
make build                    # Build binary to bin/jsbug

# Run tests
make test                     # Unit tests only (no Chrome required)
make test-integration         # All tests including integration (Chrome required)
make test-verbose             # Verbose test output
make test-coverage            # Generate coverage report

# Run single test
go test -v -run TestName ./internal/...

# Lint
make lint                     # Runs go vet and go fmt

# Run the server
make run                      # Build and run with config.yaml
./bin/jsbug -c config.yaml    # Run directly
```

## Architecture Overview

jsbug is a JavaScript page renderer that extracts SEO content from web pages. It operates in two modes:

**Request Flow:**
1. `POST /api/render` → `RenderHandler` (internal/server/render_handler.go)
2. Based on `js_enabled` flag:
   - **JS Mode**: `Renderer` → Chrome via chromedp → `Parser` → Response
   - **HTTP Mode**: `Fetcher` → HTTP client → `Parser` → Response
3. Real-time progress updates via SSE at `/api/render/stream`

**Core Components:**

| Package | Purpose |
|---------|---------|
| `internal/chrome/` | Chrome instance management and page rendering via chromedp |
| `internal/server/` | HTTP server, handlers, SSE manager |
| `internal/fetcher/` | Simple HTTP fetching (non-JS mode) |
| `internal/parser/` | HTML content extraction (title, meta, headings, JSON-LD) |
| `internal/types/` | Request/response types and validation |
| `internal/config/` | YAML config loading with env var overrides (JSBUG_*) |

**Chrome Package Details:**
- `Instance` - Manages Chrome process lifecycle and browser context
- `Renderer` - Handles page navigation, wait events, and HTML capture
- `EventCollector` - Captures network requests, console logs, JS errors via CDP
- `Blocklist` - Filters analytics, ads, social scripts, and resource types

**Wait Events:** DOMContentLoaded, load, networkIdle, networkAlmostIdle

## Key Files

- `cmd/jsbug/main.go` - Entry point, wires up all components
- `internal/server/render_handler.go` - Main API logic, orchestrates render/fetch + parse
- `internal/chrome/renderer.go` - Chrome rendering with configurable wait strategies
- `internal/chrome/events.go` - CDP event collection (network, console, lifecycle)

# jsbug

A JavaScript Page Renderer and SEO Content Analyzer that compares how websites render with and without JavaScript enabled.

## Features

- **Side-by-side comparison** - Render the same URL with JS enabled vs disabled
- **SEO content extraction** - Title, meta tags, headings, structured data, canonical URLs
- **Network analysis** - Track all HTTP requests with status, size, and timing
- **Console capture** - JavaScript console logs and errors
- **Resource blocking** - Filter out analytics, ads, and social scripts
- **Real-time progress** - SSE streaming for render status updates

## Tech Stack

| Component | Technology |
|-----------|------------|
| Backend | Go 1.24, chromedp, goquery |
| Frontend | React 19, TypeScript, Vite |
| Browser | Headless Chrome via DevTools Protocol |

## Quick Start

### Backend

```bash
cd backend
go mod download
make build
./bin/jsbug -c config.yaml
```

Server runs on `http://localhost:9301`

### Frontend

```bash
cd frontend
npm install
npm run dev
```

Dev server runs on `http://localhost:5173`

## API

**POST /api/render**

```json
{
  "url": "https://example.com",
  "js_enabled": true,
  "timeout": 10,
  "wait_event": "networkIdle",
  "blocking": {
    "analytics": true,
    "ads": true
  }
}
```

Returns page content, SEO data, network requests, console logs, and rendered HTML.

## License

MIT

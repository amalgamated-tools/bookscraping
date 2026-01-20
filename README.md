# BookScraping

A full-stack application for managing and synchronizing book libraries from Booklore (e-book management system) with Goodreads integration. Features a Go backend, TypeScript/Svelte frontend, and SQLite database.

## Features

- 📚 Sync book library from Booklore server to local database
- 📖 Browse and manage book collections with series organization
- 🔍 Complete series with missing books from Goodreads
- 🎨 Modern web interface for browsing and management
- 🔄 Real-time sync status with Server-Sent Events
- 🚀 Single executable with embedded frontend
- 🔒 Local-first architecture (self-hosted)

## Architecture

### High-Level Overview

```
┌─────────────────────────────────────────────────────────────┐
│                  Frontend (SvelteKit + TypeScript)          │
│  ├─ Pages: Books, Series, Configuration                    │
│  ├─ Real-time updates via Server-Sent Events               │
│  └─ Stores: Config (credentials), WebSocket (SSE)          │
└─────────────────────────────────────────────────────────────┘
                       ↓ (HTTP REST API)
┌─────────────────────────────────────────────────────────────┐
│                   Backend (Go HTTP Server)                  │
│  ├─ /api/config        → Credential management             │
│  ├─ /api/books         → List/search books with authors    │
│  ├─ /api/series        → List/manage series                │
│  ├─ /api/sync          → Sync from Booklore to SQLite      │
│  ├─ /api/events        → Server-Sent Events stream         │
│  └─ Integrations: Booklore API (JWT), Goodreads (scrape)   │
└─────────────────────────────────────────────────────────────┘
         ↓                    ↓                    ↓
    ┌─────────┐        ┌──────────────┐    ┌──────────────┐
    │ SQLite  │        │  Booklore    │    │  Goodreads   │
    │   DB    │        │    API       │    │   .com       │
    └─────────┘        └──────────────┘    └──────────────┘
```

### Key Components

#### Backend (Go)
- **Server** (`pkg/server`): HTTP handlers for books, series, config, and sync operations
- **Database** (`pkg/db`): Type-safe SQL queries via sqlc
- **Booklore Client** (`pkg/booklore`): API authentication and book fetching
- **Goodreads Client** (`pkg/goodreads`): Web scraper for series and books

#### Frontend (SvelteKit)
- **Pages**: Books list, book details, series list, series details, configuration
- **Stores**: Configuration (Booklore credentials), WebSocket (SSE events)
- **API Client**: Centralized TypeScript API layer

#### Database (SQLite)
- **Tables**: `books`, `authors`, `series`, `book_authors` (junction), `series_authors` (junction), `configuration`
- **Key Fields**: Books track `is_missing` (from Goodreads but not owned), series link to Goodreads IDs

### Data Flow

**Sync from Booklore:**
1. User provides Booklore credentials in UI
2. Backend authenticates with Booklore API (JWT)
3. Fetches all books with authors and series information
4. Upserts to SQLite with proper relationships
5. Real-time progress via SSE

**Complete Series from Goodreads:**
1. User clicks "Fetch from Goodreads" on series detail
2. Backend scrapes Goodreads series page (HTML parsing)
3. Identifies missing books (not in local DB)
4. Creates "missing" book records (`is_missing=1`)
5. Returns sync statistics to frontend

### Tech Stack

- **Backend**: Go 1.25+, net/http, sqlc, goquery
- **Frontend**: SvelteKit, Svelte 5, TypeScript, Vite
- **Database**: SQLite (modernc.org/sqlite)
- **Build**: Make, pnpm, Vite

## Setup & Development

### Prerequisites

- Go 1.25+
- Node.js 18+ and pnpm
- Make

### Build

```bash
# Build everything
make build

# Build just backend
go build -o bin/bookscraping-server ./cmd/server

# Build just frontend
cd frontend && pnpm install && pnpm run build
```

### Development

```bash
# Run dev environment (requires overmind or foreman)
make dev

# Or manually:
# Terminal 1 - Backend with auto-reload
cd /path/to/repo && air

# Terminal 2 - Frontend dev server
cd frontend && pnpm dev
```

Frontend dev server (port 5173) proxies API calls to backend (port 8080).

### Database

```bash
# Run migrations
go run ./cmd/cli migrate

# Generate type-safe queries from SQL
make sqlc
```

## Running

```bash
./bin/bookscraping-server
```

Server starts on `http://localhost:8080`

Configure Booklore credentials in the `/config` page, then sync your library.

## Goodreads Integration

This project includes Goodreads web scraping capabilities (via `pkg/goodreads`) to:
- Fetch series information by series ID
- Search for books and authors
- Parse book details from Goodreads pages

Uses [goquery](https://github.com/PuerkitoBio/goquery) for HTML parsing. Since Goodreads deprecated their public API, web scraping is used to access series data. Respect Goodreads' terms of service and rate limits.

## Environment Variables

```bash
# Server
PORT=8080                          # HTTP server port
LOG_LEVEL=info                     # Log level (debug, info, warn, error)

# Database
DB_PATH=./bookscraping.db          # SQLite database file path
```


## Disclaimer

This application is not affiliated with Goodreads, Amazon, or Booklore. It scrapes publicly available data from Goodreads. Please use responsibly and respect their terms of service.

## 📊 Telemetry

This project includes very minimal, privacy-respecting telemetry to help understand how many unique installations exist.

What is collected (once per install):

* Randomly generated install ID (UUID)
* Application version
* Operating system & architecture
* Timestamp of first start

What is NOT collected:

* IP addresses
* Hostnames
* Usernames
* Any application data

Telemetry is sent once, on the first container start only.

## Opt-out

You can disable telemetry entirely by setting:

```bash
TELEMETRY_ENABLED=false
```

Or in Docker Compose:

```yaml
environment:
  - TELEMETRY_ENABLED=false
```

Telemetry is used only for aggregate usage counts and project planning.
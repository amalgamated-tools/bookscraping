# GitHub Copilot Instructions for BookScraping

## Project Overview

BookScraping is a full-stack application for managing and synchronizing book libraries from Booklore (e-book management system) with Goodreads integration. It features a Go backend, TypeScript/Svelte frontend, and SQLite database.

**Key Features:**
- Sync book library from Booklore server to local database
- Browse and manage book collections with series organization
- Complete series with missing books from Goodreads
- Modern SvelteKit web interface
- Real-time sync status via Server-Sent Events
- Single executable with embedded frontend
- Local-first architecture (self-hosted)

## Tech Stack

### Backend
- **Language**: Go 1.25+
- **HTTP**: Standard library `net/http` with `*http.ServeMux` for routing
- **Database**: SQLite (modernc.org/sqlite) with sqlc for type-safe queries
- **Logging**: `log/slog` for structured logging
- **External APIs**: 
  - Booklore API (JWT authentication)
  - Goodreads web scraping (goquery)

### Frontend
- **Framework**: SvelteKit with Svelte 5
- **Language**: TypeScript (strict mode enabled)
- **Build Tool**: Vite
- **Package Manager**: pnpm
- **State Management**: Svelte stores in `/frontend/src/lib/stores/`

### Database
- **Engine**: SQLite
- **Schema Management**: SQL migrations in `db/migrations/`
- **Type Safety**: sqlc generates Go code from SQL queries
- **Key Tables**: `books`, `authors`, `series`, `book_authors`, `series_authors`, `configuration`

## Architecture

### High-Level Flow
```
Frontend (SvelteKit) → HTTP REST API → Backend (Go) → SQLite
                                    ↓
                        External: Booklore API, Goodreads
```

### Key Components
- **pkg/server**: HTTP handlers for books, series, config, and sync operations
- **pkg/db**: Type-safe SQL queries via sqlc
- **pkg/booklore**: API client for Booklore authentication and book fetching
- **pkg/goodreads**: Web scraper for series and book data
- **frontend/src**: SvelteKit pages, components, and API client

### Data Flow
1. **Booklore Sync**: User credentials → JWT auth → Fetch books → Upsert to SQLite → Real-time SSE
2. **Goodreads Series**: Series ID → HTML scrape → Parse books → Mark missing → Update DB

## Coding Conventions

### Go Code Style

#### Imports
```go
// Standard library first
import (
    "context"
    "fmt"
    "log/slog"
    
    // Third-party packages after blank line
    "github.com/google/uuid"
    "modernc.org/sqlite"
)
```

#### Naming
- **Functions/Variables**: `camelCase` (unexported), `PascalCase` (exported)
- **Constants**: `ALLCAPS` or `PascalCase` for exported
- **Packages**: Single word, lowercase

#### Error Handling
- Always return errors explicitly
- Never panic in library code
- Use `fmt.Errorf` with `%w` for error wrapping
```go
if err != nil {
    return fmt.Errorf("failed to fetch books: %w", err)
}
```

#### Comments
- Exported functions must have comments starting with function name
```go
// FetchBooks retrieves all books from the Booklore API.
func FetchBooks(ctx context.Context) ([]Book, error) {
```

#### Structs
- Use struct tags for JSON serialization and validation
```go
type Book struct {
    Title  string `json:"title" validate:"required"`
    Author string `json:"author"`
}
```

#### HTTP Handlers
- Use `*http.ServeMux` for routing
- Return errors from handler functions
- Use proper HTTP status codes
```go
func (s *Server) handleBooks(w http.ResponseWriter, r *http.Request) error {
    // Handler logic
    return nil // or error
}
```

### TypeScript/Svelte Style

#### Imports
```typescript
// Group: standard library, third-party, then local
import { onMount } from 'svelte';
import type { PageData } from './$types';
import { configStore } from '$lib/stores/configStore';
```

#### Naming
- **Variables/Functions**: `camelCase`
- **Components/Classes**: `PascalCase`
- **Files**: `kebab-case` or `+page.svelte` (SvelteKit convention)

#### Types
- Always use explicit types (avoid `any`)
- Use generics where appropriate
```typescript
interface Book {
    id: number;
    title: string;
    authors: Author[];
}

async function fetchBooks(): Promise<Book[]> {
    // Implementation
}
```

#### Stores
- Place in `/frontend/src/lib/stores/` with `.ts` extension
- Use writable stores for mutable state
```typescript
import { writable } from 'svelte/store';

export const configStore = writable<Config | null>(null);
```

#### Components
- Use `+page.svelte` for routes
- Use `+layout.svelte` for shared layouts
- Use `+page.ts` for data loading

### General Practices

#### Database
- Use sqlc for type-safe queries (config in `sqlc.yaml`)
- Write raw SQL in `db/query/` files
- Run `make sqlc` to regenerate Go code
- Place migrations in `db/migrations/` with timestamp prefixes

#### Testing
- Go: Use `testing` package and `testify` for assertions
- Place tests in `_test.go` files next to code
- Run: `go test ./...` or `make test-go`
- Frontend: TypeScript type checking via `pnpm run check`

#### Formatting
- Run `make fmt` before committing
- Go: Uses `go fmt`
- Frontend: Uses Prettier

#### Logging
- Use `log/slog` for structured logging
- Log levels: Debug, Info, Warn, Error
```go
slog.Info("syncing books", "count", len(books))
slog.Error("failed to sync", "error", err)
```

## Build & Development

### Commands
```bash
# Build everything
make build

# Build server only
make build-server
# or
go build -o bin/bookscraping-server ./cmd/server

# Build frontend only
make build-frontend

# Development (requires overmind/foreman)
make dev

# Run tests
make test           # All tests
make test-go        # Go tests only
make test-frontend  # Frontend type check and build

# Format code
make fmt

# Database
go run ./cmd/cli migrate  # Run migrations
make sqlc                 # Generate type-safe queries

# Frontend specific
cd frontend && pnpm run dev      # Dev server on :5173
cd frontend && pnpm run build    # Production build
cd frontend && pnpm run check    # Type check
cd frontend && pnpm run format   # Format with Prettier
```

### Development Workflow
1. **Backend**: Use `air` for hot-reload (configured in `.air.toml`)
2. **Frontend**: Dev server on port 5173 proxies API to backend on port 8080
3. **Database**: Migrations run via `go run ./cmd/cli migrate`

## Common Tasks

### Adding a New API Endpoint
1. Add handler method to `pkg/server/server.go`
2. Register route in `NewServer()` or `setupRoutes()`
3. Add corresponding frontend API method in `frontend/src/lib/api.ts`
4. Use structured logging and proper error handling

### Adding a Database Table
1. Create migration in `db/migrations/YYYYMMDDHHMMSS_description.sql`
2. Write queries in `db/query/query.sql`
3. Run `make sqlc` to generate Go code
4. Use generated methods from `pkg/db/`

### Adding a Frontend Page
1. Create `frontend/src/routes/path/+page.svelte`
2. Add data loading in `+page.ts` if needed
3. Add API methods in `frontend/src/lib/api.ts`
4. Use TypeScript types for all data

### Web Scraping with Goodreads
- Located in `pkg/goodreads/`
- Uses goquery for HTML parsing
- Respect rate limits and terms of service
- Handle errors gracefully (pages may change)

## Important Notes

### Module Path
- Go module: `github.com/amalgamated-tools/bookscraping`

### Environment Variables
```bash
PORT=8080                    # HTTP server port
LOG_LEVEL=info              # Log level (debug, info, warn, error)
DB_PATH=./bookscraping.db   # SQLite database file path
```

### Embedded Frontend
- Frontend builds to `frontend/build/`
- Copied to `pkg/server/dist/` for embedding
- Served from `/` route by Go server in production

### Real-time Updates
- Server-Sent Events (SSE) for sync progress
- Endpoint: `/api/events`
- Frontend store: `websocketStore.ts`

## Security Considerations

- Never commit secrets or credentials
- Booklore credentials stored in SQLite `configuration` table
- Use proper input validation for user data
- Sanitize HTML when scraping Goodreads
- Rate limit external API calls

## License

GNU AGPLv3 - This is a copyleft license. Any modifications must also be open source.

## Additional Resources

- Main README: `/README.md`
- Agent Guidelines: `/AGENTS.md`
- Database Schema: `/db/migrations/`
- API Documentation: Check handler functions in `/pkg/server/`

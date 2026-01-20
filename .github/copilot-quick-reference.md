# GitHub Copilot Quick Reference

Quick reference for common patterns and commands in the BookScraping project.

## Quick Start

```bash
# First time setup
make build              # Build everything
go run ./cmd/cli migrate  # Run database migrations

# Development
make dev               # Start both frontend and backend

# Run server
./bin/bookscraping-server
```

## File Locations

| What | Where |
|------|-------|
| HTTP Handlers | `pkg/server/*.go` |
| Database Queries | `db/query/query.sql` |
| Generated DB Code | `pkg/db/query.sql.go` |
| Migrations | `db/migrations/*.sql` |
| Frontend Pages | `frontend/src/routes/` |
| API Client | `frontend/src/lib/api.ts` |
| Svelte Stores | `frontend/src/lib/stores/` |
| Server Entry Point | `cmd/server/main.go` |
| CLI Tools | `cmd/cli/main.go` |

## Code Snippets

### Go HTTP Handler
```go
func (s *Server) handleEndpoint(w http.ResponseWriter, r *http.Request) error {
    // Get ID from URL path
    id := r.PathValue("id")
    
    // Query database
    result, err := s.queries.GetSomething(r.Context(), id)
    if err != nil {
        return fmt.Errorf("get something: %w", err)
    }
    
    // Return JSON
    w.Header().Set("Content-Type", "application/json")
    return json.NewEncoder(w).Encode(result)
}
```

### Frontend API Call
```typescript
interface Book {
    id: number;
    title: string;
    authors: Author[];
}

export async function fetchBooks(): Promise<Book[]> {
    const response = await fetch('/api/books');
    if (!response.ok) {
        throw new Error(`HTTP ${response.status}: ${response.statusText}`);
    }
    return response.json();
}
```

### Svelte Store
```typescript
import { writable } from 'svelte/store';

interface Config {
    bookloreUrl: string;
    username: string;
}

export const configStore = writable<Config | null>(null);

// Usage in component
import { configStore } from '$lib/stores/configStore';
let config = $state($configStore);
```

### sqlc Query
```sql
-- name: GetBook :one
SELECT * FROM books WHERE id = ? LIMIT 1;

-- name: ListBooks :many
SELECT * FROM books ORDER BY title;

-- name: CreateBook :one
INSERT INTO books (title, isbn) 
VALUES (?, ?) 
RETURNING *;

-- name: UpdateBook :exec
UPDATE books SET title = ? WHERE id = ?;

-- name: DeleteBook :exec
DELETE FROM books WHERE id = ?;
```

### Database Migration
```sql
-- db/migrations/20260120000000_add_column.sql
-- Add a new column to books table
ALTER TABLE books ADD COLUMN publisher TEXT;

-- Create an index
CREATE INDEX idx_books_publisher ON books(publisher);
```

### Server-Sent Events (SSE)
```go
// Backend - Send SSE event
func (s *Server) sendEvent(eventType, data string) {
    s.mu.Lock()
    defer s.mu.Unlock()
    for client := range s.clients {
        fmt.Fprintf(client, "event: %s\ndata: %s\n\n", eventType, data)
        client.(http.Flusher).Flush()
    }
}

// Frontend - Subscribe to events
const eventSource = new EventSource('/api/events');
eventSource.addEventListener('sync:progress', (e) => {
    const data = JSON.parse(e.data);
    console.log('Sync progress:', data);
});
```

### Error Handling
```go
// Go - Wrap errors with context
if err != nil {
    return fmt.Errorf("failed to fetch books: %w", err)
}

// TypeScript - Type-safe error handling
try {
    const books = await fetchBooks();
    return books;
} catch (error) {
    console.error('Failed to fetch books:', error);
    throw error;
}
```

### Logging
```go
import "log/slog"

// Structured logging
slog.Info("sync started", "source", "booklore", "count", bookCount)
slog.Error("sync failed", "error", err, "duration", elapsed)
slog.Debug("processing book", "id", book.ID, "title", book.Title)
```

## Common Commands

### Development
```bash
air                        # Hot-reload Go server
cd frontend && pnpm dev    # Frontend dev server with proxy
make dev                   # Both (requires overmind/foreman)
```

### Build
```bash
make build                 # Everything
make build-server          # Server with embedded frontend
make build-frontend        # Frontend only
go build -o bin/server ./cmd/server  # Manual server build
```

### Test
```bash
make test                  # All tests
make test-go               # Go tests
make test-frontend         # TypeScript check + build
go test ./pkg/server/      # Specific package
go test -run TestName      # Specific test
```

### Database
```bash
go run ./cmd/cli migrate              # Run migrations
make sqlc                              # Generate from query.sql
sqlite3 bookscraping.db ".schema"      # View schema
sqlite3 bookscraping.db "SELECT * FROM books LIMIT 5"
```

### Format & Lint
```bash
make fmt                   # Format all code
go fmt ./...               # Format Go only
cd frontend && pnpm format # Format frontend only
gofmt -d .                 # Show Go formatting diff
```

### Git
```bash
git status
git add .
git commit -m "message"
git --no-pager diff        # View changes (no pager)
git --no-pager log -5      # Recent commits
```

## API Endpoints

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/books` | List all books with authors |
| GET | `/api/books/:id` | Get book details |
| GET | `/api/series` | List all series |
| GET | `/api/series/:id` | Get series with books |
| POST | `/api/series/:id/fetch` | Fetch series from Goodreads |
| GET | `/api/config` | Get configuration |
| PUT | `/api/config` | Update configuration |
| POST | `/api/sync` | Sync from Booklore |
| GET | `/api/events` | SSE event stream |

## Database Schema

### Core Tables
```sql
books (
    id, title, isbn, cover_url, 
    booklore_id, series_id, series_position, 
    is_missing, created_at
)

authors (id, name, booklore_id, goodreads_id)

series (id, title, booklore_id, goodreads_id)

book_authors (book_id, author_id)

series_authors (series_id, author_id)

configuration (key, value)
```

## Environment Variables

```bash
export PORT=8080
export LOG_LEVEL=debug
export DB_PATH=./bookscraping.db
```

## Troubleshooting

### Frontend build fails
```bash
cd frontend
rm -rf node_modules pnpm-lock.yaml
pnpm install
pnpm run build
```

### Go build fails
```bash
go mod tidy           # Clean up dependencies
go mod download       # Re-download modules
make clean            # Clean artifacts
make build            # Rebuild
```

### Database issues
```bash
# Regenerate database code
make sqlc

# Reset database (DELETES DATA!)
rm bookscraping.db
go run ./cmd/cli migrate
```

### sqlc errors
```bash
# Check query syntax in db/query/query.sql
# Verify sqlc.yaml configuration
sqlc generate -f sqlc.yaml
```

## Import Paths

### Go
```go
import (
    "context"
    "fmt"
    "log/slog"
    
    "github.com/amalgamated-tools/bookscraping/pkg/db"
    "github.com/amalgamated-tools/bookscraping/pkg/server"
)
```

### TypeScript
```typescript
import { onMount } from 'svelte';
import type { PageData } from './$types';
import { configStore } from '$lib/stores/configStore';
import { fetchBooks } from '$lib/api';
```

## Git Ignore Patterns

Never commit:
- `bin/` - Compiled binaries
- `frontend/build/` - Frontend build output
- `frontend/node_modules/` - Dependencies
- `pkg/server/dist/` - Embedded frontend
- `*.db` - Database files
- `.env` - Environment files
- `.DS_Store` - macOS files

## Key Dependencies

### Go
- `modernc.org/sqlite` - SQLite driver
- `github.com/PuerkitoBio/goquery` - HTML parsing
- `github.com/google/uuid` - UUID generation
- `github.com/stretchr/testify` - Testing

### Frontend
- `@sveltejs/kit` - SvelteKit framework
- `svelte` - Svelte 5
- `typescript` - Type checking
- `vite` - Build tool

## Performance Tips

- Use indexes on foreign keys
- Batch database operations
- Use context with timeouts
- Enable SQLite WAL mode
- Minimize SSE connections
- Lazy load routes in frontend

## Security Checklist

- ✓ Use prepared statements (sqlc)
- ✓ Validate all inputs
- ✓ Sanitize HTML from scraping
- ✓ Rate limit external APIs
- ✓ No hardcoded secrets
- ✓ Use HTTPS for external calls
- ✓ Proper error messages (no data leaks)

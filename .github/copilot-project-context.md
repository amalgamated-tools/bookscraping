# BookScraping Project Context

## Detailed Architecture

### System Components

#### 1. Frontend (SvelteKit + TypeScript)

**Structure:**
```
frontend/src/
├── routes/              # SvelteKit pages
│   ├── +page.svelte    # Home/Books list
│   ├── +layout.svelte  # Root layout
│   ├── series/         # Series pages
│   │   ├── +page.svelte              # Series list
│   │   └── [id]/+page.svelte         # Series detail
│   └── config/         # Configuration page
│       └── +page.svelte
├── lib/
│   ├── api.ts          # Centralized API client
│   └── stores/         # Svelte stores
│       ├── configStore.ts      # Booklore credentials
│       └── websocketStore.ts   # SSE connection
└── app.d.ts            # TypeScript definitions
```

**Key Patterns:**
- SvelteKit file-based routing (uses `+page.svelte` convention)
- Server-Sent Events for real-time sync status
- Centralized API client with typed responses
- Reactive stores for global state

**API Client Pattern:**
```typescript
// frontend/src/lib/api.ts
export async function fetchBooks(): Promise<Book[]> {
    const response = await fetch('/api/books');
    if (!response.ok) throw new Error('Failed to fetch books');
    return response.json();
}
```

#### 2. Backend (Go HTTP Server)

**Structure:**
```
pkg/
├── server/            # HTTP handlers and server setup
│   ├── server.go      # Main server with embedded frontend
│   ├── config.go      # Configuration endpoints
│   ├── series.go      # Series management
│   ├── sync.go        # Booklore sync logic
│   └── events.go      # Server-Sent Events
├── db/                # Database layer (sqlc generated)
│   ├── query.sql.go   # Generated queries
│   ├── querier.go     # Generated interface
│   └── models.go      # Generated models
├── booklore/          # Booklore API client
│   └── client.go
└── goodreads/         # Goodreads web scraper
    ├── series_parser.go
    └── book.go

cmd/
├── server/            # Main server entry point
│   └── main.go
└── cli/               # CLI tools (migrations)
    └── main.go
```

**HTTP Routing Pattern:**
```go
// pkg/server/server.go
func NewServer(db *sql.DB) *Server {
    mux := http.NewServeMux()
    s := &Server{db: db, queries: dbpkg.New(db), mux: mux}
    
    // API routes
    mux.HandleFunc("GET /api/books", s.handleBooks)
    mux.HandleFunc("GET /api/series", s.handleSeries)
    mux.HandleFunc("POST /api/sync", s.handleSync)
    
    return s
}
```

**Error Handling Pattern:**
```go
func (s *Server) handleBooks(w http.ResponseWriter, r *http.Request) error {
    books, err := s.queries.ListBooks(r.Context())
    if err != nil {
        return fmt.Errorf("list books: %w", err)
    }
    return json.NewEncoder(w).Encode(books)
}
```

#### 3. Database Layer (SQLite + sqlc)

**Schema Overview:**

```sql
-- Core tables
books (id, title, isbn, cover_url, booklore_id, series_id, series_position, is_missing)
authors (id, name, booklore_id, goodreads_id)
series (id, title, booklore_id, goodreads_id)

-- Junction tables
book_authors (book_id, author_id)
series_authors (series_id, author_id)

-- Configuration
configuration (key, value)
```

**Key Relationships:**
- Books can belong to one series (series_id FK)
- Books can have multiple authors (via book_authors junction)
- Series can have multiple authors (via series_authors junction)
- `is_missing` flag indicates books fetched from Goodreads but not owned

**sqlc Workflow:**
1. Write SQL queries in `db/query/query.sql`
2. Run `make sqlc` to generate type-safe Go code
3. Use generated methods: `queries.ListBooks(ctx)`, `queries.InsertBook(ctx, params)`

**Example Query:**
```sql
-- name: ListBooks :many
SELECT b.*, 
       json_group_array(json_object('id', a.id, 'name', a.name)) as authors
FROM books b
LEFT JOIN book_authors ba ON b.id = ba.book_id
LEFT JOIN authors a ON ba.author_id = a.id
GROUP BY b.id;
```

### Data Flow Patterns

#### Booklore Sync Flow

```
User Input (UI) → POST /api/sync
    ↓
Server validates credentials
    ↓
Authenticate with Booklore API (JWT)
    ↓
Fetch books from Booklore (paginated)
    ↓
For each book:
    - Upsert authors
    - Upsert series
    - Upsert book with relationships
    - Emit SSE progress event
    ↓
Return sync statistics
```

**Implementation Details:**
- Uses goroutines for concurrent processing
- Implements exponential backoff for API retries
- Tracks progress via channels and SSE
- Transactional database updates

#### Goodreads Series Completion Flow

```
User clicks "Fetch from Goodreads" → POST /api/series/{id}/fetch
    ↓
Get series Goodreads ID from database
    ↓
Scrape Goodreads series page (HTML parsing)
    ↓
Extract all books in series
    ↓
Compare with owned books in database
    ↓
For missing books:
    - Create book record with is_missing=true
    - Link to series
    - Link to authors
    ↓
Return sync statistics
```

**Implementation Details:**
- Uses goquery for HTML parsing
- Handles pagination on Goodreads series pages
- Rate limits requests (1 second delay)
- Gracefully handles HTML structure changes

#### Real-time Updates (SSE)

```
Frontend subscribes → GET /api/events
    ↓
Server creates SSE connection
    ↓
Backend processes emit events
    ↓
Server broadcasts to all connections
    ↓
Frontend receives and updates UI
```

**Event Types:**
- `sync:progress` - Book sync progress
- `sync:complete` - Sync finished
- `sync:error` - Sync error occurred

### External Integrations

#### Booklore API
- **Authentication**: JWT tokens via username/password
- **Endpoints**: `/api/v1/auth/login`, `/api/v1/books`
- **Data Format**: JSON with nested authors and series
- **Rate Limits**: Not specified (use reasonable delays)

#### Goodreads Web Scraping
- **Target Pages**: Series pages (`/series/show/{id}`)
- **Parsing**: goquery (jQuery-like selectors)
- **Challenges**: 
  - No official API (deprecated in 2020)
  - HTML structure may change
  - Must respect robots.txt and terms of service
- **Rate Limiting**: 1-2 second delays between requests

### Build and Deployment

#### Build Process

**Frontend Build:**
```bash
cd frontend
pnpm install         # Install dependencies
pnpm run build       # Vite builds to frontend/build/
```

**Backend Build:**
```bash
# Copy frontend build to pkg/server/dist/
cp -r frontend/build pkg/server/dist

# Compile Go server with embedded frontend
go build -o bin/bookscraping-server ./cmd/server
```

**Result:** Single executable `bin/bookscraping-server` with embedded UI

#### Development Mode

**With Process Manager (overmind/foreman):**
```bash
make dev  # Runs both frontend and backend
```

**Procfile.dev:**
```
web: cd frontend && pnpm run dev
api: air  # Go hot-reload server
```

**Manual:**
```bash
# Terminal 1 - Backend with hot-reload
air

# Terminal 2 - Frontend dev server
cd frontend && pnpm run dev
```

Frontend dev server (port 5173) proxies API calls to backend (port 8080).

#### Database Migrations

**Creating a migration:**
```bash
# Migrations are timestamped SQL files
db/migrations/20260115002404_create_authors.sql
```

**Running migrations:**
```bash
go run ./cmd/cli migrate
```

**Migration pattern:**
```sql
-- Migration: create_books
CREATE TABLE IF NOT EXISTS books (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    title TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
```

### Testing Strategy

#### Go Tests
- Unit tests for business logic
- Integration tests for database operations
- Mock database interface (mockery)
- Table-driven tests

**Example:**
```go
func TestFetchBooks(t *testing.T) {
    tests := []struct {
        name    string
        setup   func(*mock.Querier)
        wantErr bool
    }{
        // Test cases
    }
    // Run tests
}
```

#### Frontend Tests
- TypeScript strict mode catches type errors
- Build validates component correctness
- No unit tests currently (TODO)

### Performance Considerations

#### Database
- SQLite with WAL mode for concurrent reads
- Indexes on foreign keys
- Batch inserts for sync operations
- JSON aggregation for complex queries

#### Frontend
- Static site generation (SSG) where possible
- Code splitting via Vite
- Lazy loading of routes

#### API
- Concurrent book processing with goroutines
- Connection pooling for database
- HTTP keep-alive for external APIs

### Security Model

#### Authentication
- No user authentication (single-user/self-hosted)
- Booklore credentials stored encrypted in SQLite
- No session management required

#### Input Validation
- Validate all API inputs
- Sanitize HTML from Goodreads scraping
- Use prepared statements (sqlc) to prevent SQL injection

#### External Requests
- Validate URLs before scraping
- Timeout on HTTP requests
- Handle malicious HTML gracefully

### Common Gotchas

1. **Frontend Build**: Must run `make build-frontend` before `make build-server`
2. **Database**: Schema changes require running migrations before `make sqlc`
3. **SSE Connections**: Browser limits concurrent SSE connections (6 per domain)
4. **Goodreads**: HTML structure can change, breaking scraper
5. **Go Embed**: Changes to frontend require rebuilding server binary

### Extension Points

#### Adding New Data Sources
1. Create client in `pkg/newsource/`
2. Implement sync logic in `pkg/server/sync_newsource.go`
3. Add API endpoint for triggering sync
4. Add UI configuration in frontend

#### Adding New Book Metadata
1. Add column to books table (migration)
2. Update queries in `db/query/query.sql`
3. Run `make sqlc` to regenerate Go code
4. Update API responses and frontend types

#### Adding Search Functionality
1. Create full-text search query in SQL
2. Add handler in `pkg/server/`
3. Add API method in `frontend/src/lib/api.ts`
4. Create search UI component

### Useful SQL Queries

**Books with authors:**
```sql
SELECT b.*, GROUP_CONCAT(a.name, ', ') as authors
FROM books b
LEFT JOIN book_authors ba ON b.id = ba.book_id
LEFT JOIN authors a ON ba.author_id = a.id
GROUP BY b.id;
```

**Incomplete series:**
```sql
SELECT s.*, COUNT(b.id) as book_count
FROM series s
LEFT JOIN books b ON s.id = b.series_id
GROUP BY s.id
HAVING book_count < s.expected_count;
```

**Missing books (from Goodreads):**
```sql
SELECT * FROM books WHERE is_missing = 1;
```

# Agent Guidelines for BookScraping

## Build & Test Commands

### Go
- **Build server**: `make build-server` or `go build -o bin/bookscraping-server ./cmd/server`
- **Build everything**: `make build` (builds frontend + server)
- **Run tests**: `go test ./...`
- **Run single test**: `go test -run TestName ./path/to/package`
- **Format code**: `go fmt ./...`
- **Development**: `make dev` (runs frontend + backend via overmind/foreman)

### Frontend (TypeScript + Svelte)
- **Dev server**: `cd frontend && pnpm run dev`
- **Build**: `cd frontend && pnpm run build`
- **Type check**: `cd frontend && pnpm run check`
- **Format**: `cd frontend && pnpm run format`

### Database
- **Run migrations**: `go run ./cmd/cli migrate`
- **Generate sqlc**: `make sqlc` (generates type-safe DB code from SQL)

## Code Style Guidelines

### Go
- **Imports**: Standard library first, then blank line, then third-party packages
- **Naming**: `camelCase` for functions/variables, `PascalCase` for exported, `ALLCAPS` for constants
- **Error handling**: Return errors explicitly; never panic in library code
- **Comments**: Exported functions must have comment starting with function name
- **Structs**: Use struct tags for `json` and validation (e.g., `json:"title" validate:"required"`)
- **Logging**: Use `log/slog` for structured logging
- **HTTP**: Use `*http.ServeMux` for routing, return errors from handlers

### TypeScript + Svelte
- **Strict mode**: Enabled (`"strict": true` in tsconfig.json)
- **Imports**: Use ESM syntax; group standard library, third-party, then local imports
- **Naming**: `camelCase` for functions/variables, `PascalCase` for components/classes
- **Types**: Always use explicit types (not `any`); use generics where appropriate
- **Stores**: Use Svelte stores in `/frontend/src/lib/stores/` with `.ts` extension
- **Components**: Use `+page.svelte`/`+layout.svelte` for routes (SvelteKit convention)

### General
- **Formatting**: Run `make fmt` before committing
- **Database**: Use sqlc for type-safe queries (config in `sqlc.yaml`); migrations in `db/migrations/`
- **Modules**: Go module is `github.com/amalgamated-tools/bookscraping`
- **No Cursor/Copilot rules**: No `.cursorrules` or copilot instructions file exists

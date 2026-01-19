.PHONY: all build build-frontend build-server dev clean

# Default target
all: build

# Build everything
build: build-frontend build-server

# Build the Svelte frontend
build-frontend:
	cd frontend && pnpm install && pnpm run build
	# Copy built assets to server package for embedding
	rm -rf pkg/server/dist
	cp -r frontend/build pkg/server/dist

# Build the Go server (with embedded frontend)
build-server: build-frontend
	go build -o bin/bookscraping-server ./cmd/server

# Run both frontend and backend in development (requires foreman/overmind/goreman)
dev:
	overmind start -f Procfile.dev || goreman -f Procfile.dev start || foreman start -f Procfile.dev

# Development mode - run frontend dev server with proxy to Go backend
dev-frontend:
	cd frontend && pnpm run dev

# Run the Go server
dev-server:
	go run ./cmd/server

# Clean build artifacts
clean:
	rm -rf bin/
	rm -rf frontend/build
	rm -rf frontend/.svelte-kit
	rm -rf pkg/server/dist
	mkdir -p pkg/server/dist

# Install frontend dependencies
install-frontend:
	cd frontend && pnpm install

# Format code
fmt:
	go fmt ./...
	cd frontend && pnpm run format

# Run database migrations
migrate:
	go run ./cmd/cli migrate

# Generate sqlc
sqlc:
	sqlc generate

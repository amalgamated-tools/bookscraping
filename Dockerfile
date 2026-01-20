# Build stage for frontend
FROM node:22-alpine AS frontend-builder

WORKDIR /app/frontend

# Copy frontend source
COPY frontend/package.json frontend/pnpm-lock.yaml ./

# Install pnpm
RUN npm install -g pnpm

# Install dependencies
RUN pnpm install --frozen-lockfile

# Copy frontend code
COPY frontend/ .

# Build frontend
RUN pnpm run build


# Build stage for backend
FROM golang:1.25.6-alpine AS backend-builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache make

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY cmd/ ./cmd/
COPY pkg/ ./pkg/
COPY db/ ./db/
COPY sqlc.yaml ./

# Copy built frontend from frontend-builder
COPY --from=frontend-builder /app/frontend/build ./pkg/server/dist

ARG VERSION=dev

# Build the server
RUN go build -ldflags="-X main.Version=${VERSION}" -o bin/bookscraping-server ./cmd/server


# Final runtime stage
FROM alpine:latest

# Install runtime dependencies
RUN apk add --no-cache ca-certificates tzdata wget

WORKDIR /app

# Copy the built binary from backend-builder
COPY --from=backend-builder /app/bin/bookscraping-server ./

# Copy migrations for dbmate
COPY --from=backend-builder /app/db/migrations ./db/migrations

# Create data directory for mounted volume
RUN mkdir -p /data

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8080/ || exit 1

# Run the server
ENTRYPOINT ["./bookscraping-server"]

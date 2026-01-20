# Telemetry Worker

A Cloudflare Worker that collects anonymous telemetry data about BookScraping installations using [Hono](https://github.com/honojs/hono).

## Purpose

This worker receives telemetry pings from BookScraping clients and stores installation metadata in Cloudflare KV storage. It tracks:
- Installation IDs (unique identifiers for each installation)
- Version information
- OS and architecture
- First seen timestamp

Only the first ping from each unique installation is stored to count new installations.

## Endpoints

- `POST /` - Submit telemetry data
  - Request body: `{ install_id: string, version?: string, os?: string, arch?: string, timestamp?: string }`
  - Response: `{ status: string, install_id: string, new_install: boolean }`
- `GET /` - Health check (returns "ok")

## Setup

1. Sign up for [Cloudflare Workers](https://workers.dev)
2. Install dependencies: `pnpm install`
3. Login to Cloudflare: `wrangler login`
4. Deploy: `wrangler deploy`

## Development

1. Run `wrangler dev` to start a local worker at `http://localhost:8787`
2. Make requests to `http://localhost:8787/` to test the endpoints
3. Changes in `src/` will automatically reload in the worker

# Telemetry Worker

A Cloudflare Worker that collects anonymous telemetry data about BookScraping installations using [Hono](https://github.com/honojs/hono).

## Purpose

This worker receives telemetry pings from BookScraping clients and stores installation metadata in Cloudflare KV storage. It tracks:
- Installation IDs (unique identifiers for each installation)
- Version information
- OS and architecture
- First seen timestamp

1. Sign up for [Cloudflare Workers](https://workers.dev). The free tier is more than enough for most use cases.
2. Clone this project and install dependencies with `npm install`
3. Run `wrangler login` to login to your Cloudflare account in wrangler
4. Create a KV namespace for telemetry data:
   ```bash
   wrangler kv:namespace create "TELEMETRY_KV"
   ```
   This will output a namespace ID that you'll need in the next step.
5. Configure your KV namespace:
   - Copy `wrangler.jsonc.example` to `wrangler.jsonc` (if not already present)
   - Uncomment the `kv_namespaces` section in `wrangler.jsonc`
   - Replace `YOUR_KV_NAMESPACE_ID_HERE` with the ID from step 4
6. Run `wrangler deploy` to publish the API to Cloudflare Workers

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

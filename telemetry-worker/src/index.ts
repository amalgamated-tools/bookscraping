import { Hono } from 'hono'

type Bindings = {
	TELEMETRY_KV: KVNamespace
}

type TelemetryPayload = {
	install_id: string
	version?: string
	os?: string
	arch?: string
	timestamp?: string
}

const app = new Hono<{ Bindings: Bindings }>()

app.post('/', async (c) => {
	console.log('Received telemetry data')
	let payload: TelemetryPayload

	try {
		console.log('Parsing JSON payload')
		payload = await c.req.json()
	} catch {
		console.error('Invalid JSON payload')
		return c.text('Invalid JSON', 400)
	}

	console.log('Payload parsed successfully:', payload)

	const { install_id, version, os, arch, timestamp } = payload

	if (!install_id) {
		console.error('Missing install_id in payload')
		return c.text('Missing install_id', 400)
	}

	const key = `install:${install_id}`

	console.log(`Storing telemetry data for key: ${key}`, { version, os, arch, timestamp })

	// Only count first time we see this install
	const existing = await c.env.TELEMETRY_KV.get(key)
	console.log(`Existing entry for ${key}:`, existing)

	if (!existing) {
		console.log(`Storing new telemetry entry for ${key}`)
		await c.env.TELEMETRY_KV.put(
			key,
			JSON.stringify({
				version,
				os,
				arch,
				first_seen: timestamp ?? new Date().toISOString()
			})
		)
	}

	return c.json({ status: 'success', install_id: install_id, new_install: !existing }, 200)
})

/**
 * Optional health check
 */
app.get('/', (c) => c.text('ok'))

export default app

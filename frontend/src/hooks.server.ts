import type { Handle } from "@sveltejs/kit";

const API_PROXY_TARGET = "http://localhost:8080";

export const handle: Handle = async ({ event, resolve }) => {
  // Proxy /api requests to the Go backend in development
  if (event.url.pathname.startsWith("/api")) {
    const targetUrl = `${API_PROXY_TARGET}${event.url.pathname}${event.url.search}`;

    const response = await fetch(targetUrl, {
      method: event.request.method,
      headers: event.request.headers,
      body:
        event.request.method !== "GET" && event.request.method !== "HEAD"
          ? await event.request.text()
          : undefined,
    });

    return new Response(response.body, {
      status: response.status,
      statusText: response.statusText,
      headers: response.headers,
    });
  }

  return resolve(event);
};

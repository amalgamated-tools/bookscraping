import { browser } from "$app/environment";
import { writable } from "svelte/store";

export interface SSEState {
	status: "connecting" | "open" | "closed" | "error";
	eventSource: EventSource | null;
	lastMessage: string | null;
}

const createSSEStore = () => {
	const { subscribe, set, update } = writable<SSEState>({
		status: "closed",
		eventSource: null,
		lastMessage: null
	});

	let eventSource: EventSource | null = null;
	let reconnectTimeout: number | undefined;
	let reconnectAttempts = 0;

	const connect = () => {
		if (!browser || eventSource) {
			return;
		}

		set({ status: "connecting", eventSource: null, lastMessage: null });

		const url = `/api/events`;
		console.log("Attempting to connect to SSE:", url);
		eventSource = new EventSource(url);

		eventSource.addEventListener("open", () => {
			console.log("SSE connection opened");
			reconnectAttempts = 0;
			set({ status: "open", eventSource, lastMessage: null });
		});

		eventSource.addEventListener("message", event => {
			console.log("SSE message received:", event.data);
			update(state => ({
				...state,
				lastMessage: typeof event.data === "string" ? event.data : null
			}));
		});

		eventSource.addEventListener("error", () => {
			console.error("SSE error occurred");
			eventSource?.close();
			eventSource = null;
			set({ status: "error", eventSource: null, lastMessage: null });

			if (browser) {
				const delay = Math.min(1000 * 2 ** reconnectAttempts, 10000);
				reconnectAttempts += 1;
				console.log("Will reconnect in", delay, "ms (attempt", reconnectAttempts, ")");
				reconnectTimeout = window.setTimeout(connect, delay);
			}
		});
	};

	const disconnect = () => {
		if (reconnectTimeout) {
			window.clearTimeout(reconnectTimeout);
			reconnectTimeout = undefined;
		}

		if (eventSource) {
			eventSource.close();
			eventSource = null;
		}

		set({ status: "closed", eventSource: null, lastMessage: null });
	};

	return {
		subscribe,
		connect,
		disconnect
	};
};

export const websocketStore = createSSEStore();

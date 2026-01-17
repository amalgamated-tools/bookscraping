import { browser } from "$app/environment";
import { writable } from "svelte/store";

export interface WebSocketState {
	status: "connecting" | "open" | "closed" | "error";
	socket: WebSocket | null;
	lastMessage: string | null;
}

const createWebSocketStore = () => {
	const { subscribe, set, update } = writable<WebSocketState>({
		status: "closed",
		socket: null,
		lastMessage: null
	});

	let socket: WebSocket | null = null;
	let reconnectTimeout: number | undefined;
	let reconnectAttempts = 0;

	const connect = () => {
		if (!browser || socket) {
			return;
		}

		set({ status: "connecting", socket: null, lastMessage: null });

		const protocol = window.location.protocol === "https:" ? "wss" : "ws";
		const url = `${protocol}://${window.location.host}/ws`;
		console.log("Attempting to connect to WebSocket:", url);
		socket = new WebSocket(url);

		socket.addEventListener("open", () => {
			console.log("WebSocket connection opened");
			reconnectAttempts = 0;
			set({ status: "open", socket, lastMessage: null });
		});

		socket.addEventListener("message", event => {
			console.log("WebSocket message received:", event.data);
			update(state => ({
				...state,
				lastMessage: typeof event.data === "string" ? event.data : null
			}));
		});

		socket.addEventListener("error", () => {
			console.error("WebSocket error occurred");
			set({ status: "error", socket, lastMessage: null });
		});

		socket.addEventListener("close", () => {
			console.log("WebSocket connection closed");
			socket = null;
			set({ status: "closed", socket: null, lastMessage: null });

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

		if (socket) {
			socket.close();
			socket = null;
		}

		set({ status: "closed", socket: null, lastMessage: null });
	};

	const send = (message: string) => {
		if (!socket) {
			console.warn("WebSocket not initialized - connection hasn't started yet");
			return;
		}
		const states: Record<number, string> = {
			0: "CONNECTING",
			1: "OPEN",
			2: "CLOSING",
			3: "CLOSED"
		};
		const stateName = states[socket.readyState] || "UNKNOWN";
		if (socket.readyState !== WebSocket.OPEN) {
			console.warn(`WebSocket not open. Current state: ${socket.readyState} (${stateName})`);
			return;
		}
		console.log("WebSocket message sent:", message);
		socket.send(message);
	};

	return {
		subscribe,
		connect,
		disconnect,
		send
	};
};

export const websocketStore = createWebSocketStore();

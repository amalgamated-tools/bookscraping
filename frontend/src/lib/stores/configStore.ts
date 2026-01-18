import { writable } from 'svelte/store';
import { browser } from '$app/environment';
import { api } from '$lib/api';

export interface Config {
	serverUrl: string;
	username: string;
	password: string;
}

// Create a writable store with initial empty config
export const configStore = writable<Config>({
	serverUrl: '',
	username: '',
	password: ''
});

// Track whether config has been loaded
let configLoaded = false;

// Load config from the API - only once
export async function loadConfig() {
	if (!browser || configLoaded) {
		console.log("Config already loaded or not in browser.");
		return;
	}

	try {
		console.log("Loading config from API...");
		const config = await api.getConfig();
		console.log("Config loaded:", config);
		configStore.set(config);
		configLoaded = true;
	} catch (err) {
		console.error("Failed to load config:", err);
		configLoaded = true; // Mark as loaded even on error to avoid retry loops
	}
}

// Reset the loaded flag when needed (e.g., after saving config)
export function resetConfigCache() {
	configLoaded = false;
}

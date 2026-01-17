<script lang="ts">
	import { browser } from "$app/environment";
	import { configStore, loadConfig } from "$lib/stores/configStore";
	import { websocketStore } from "$lib/stores/websocketStore";

	let { children } = $props();
	let isConfigured = $state(false);

	const sendTestMessage = () => {
		const testMessage = `Test message at ${new Date().toISOString()}`;
		websocketStore.send(testMessage);
	};

	$effect(() => {
		if (browser) {
			// Load config once when layout mounts
			loadConfig();
			websocketStore.connect();
			console.log("WebSocket connected");
			// Subscribe to config store to check if configured
			const unsubscribe = configStore.subscribe((config) => {
				isConfigured = !!(
					config.serverUrl &&
					config.username &&
					config.password
				);
			});

			return () => {
				websocketStore.disconnect();
				unsubscribe();
			};
		}
	});
</script>

<div class="app">
	<header>
		<nav>
			{#if isConfigured}
				<a href="/">Home</a>
				<a href="/books">Books</a>
				<a href="/series">Series</a>
			{/if}
			<a href="/config">Config</a>
			<button onclick={sendTestMessage} class="test-button">Send Test WS Message</button>
		</nav>
	</header>

	<main>
		{@render children()}
	</main>

	<footer>
		<p>BookScraping &copy; {new Date().getFullYear()}</p>
	</footer>
</div>

<style>
	:global(body) {
		margin: 0;
		font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto,
			Oxygen, Ubuntu, Cantarell, "Open Sans", "Helvetica Neue", sans-serif;
		background-color: #f5f5f5;
		color: #333;
	}

	.app {
		display: flex;
		flex-direction: column;
		min-height: 100vh;
	}

	header {
		background-color: #2c3e50;
		padding: 1rem;
	}

	nav {
		max-width: 1200px;
		margin: 0 auto;
		display: flex;
		gap: 1.5rem;
	}

	nav a {
		color: white;
		text-decoration: none;
		font-weight: 500;
		transition: opacity 0.2s;
	}

	nav a:hover {
		opacity: 0.8;
	}

	.test-button {
		background-color: #27ae60;
		color: white;
		border: none;
		padding: 0.5rem 1rem;
		border-radius: 4px;
		cursor: pointer;
		font-weight: 500;
		transition: opacity 0.2s;
		margin-left: auto;
	}

	.test-button:hover {
		opacity: 0.8;
	}

	main {
		flex: 1;
		max-width: 1200px;
		margin: 0 auto;
		padding: 2rem;
		width: 100%;
		box-sizing: border-box;
	}

	footer {
		background-color: #2c3e50;
		color: white;
		text-align: center;
		padding: 1rem;
	}
</style>

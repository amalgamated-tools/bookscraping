<script lang="ts">
	import { browser } from "$app/environment";

	let { children } = $props();
	let isConfigured = $state(false);

	function checkConfiguration() {
		if (browser) {
			const serverUrl = localStorage.getItem("serverUrl");
			const username = localStorage.getItem("username");
			const password = localStorage.getItem("password");
			isConfigured = !!(serverUrl && username && password);
		}
	}

	$effect(() => {
		checkConfiguration();
		
		if (browser) {
			window.addEventListener('storage', checkConfiguration);
			return () => window.removeEventListener('storage', checkConfiguration);
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

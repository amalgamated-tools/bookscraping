<script lang="ts">
	import { browser } from "$app/environment";
	import { api } from "$lib/api";
	import {
		configStore,
		loadConfig,
		resetConfigCache,
	} from "$lib/stores/configStore";

	let serverUrl = $state("");
	let username = $state("");
	let password = $state("");
	let success = $state(false);
	let testing = $state(false);
	let testMessage = $state("");
	let testSuccess = $state(false);

	$effect(() => {
		if (browser) {
			// Load config once when page mounts
			console.debug("Config page mounted, loading config...");
			loadConfig();

			// Subscribe to config store to populate form
			const unsubscribe = configStore.subscribe((config) => {
				serverUrl = config.serverUrl || "";
				username = config.username || "";
				password = config.password || "";
			});

			return unsubscribe;
		}
	});

	async function handleTest() {
		if (!serverUrl || !username || !password) {
			testMessage = "Please fill in all fields first";
			testSuccess = false;
			return;
		}

		testing = true;
		testMessage = "";

		try {
			const response = await api.testConnection(
				serverUrl,
				username,
				password,
			);
			testSuccess = true;
			testMessage = response.message || "Connection successful!";
		} catch (e) {
			console.error("Test connection failed:", e);
			testSuccess = false;
			testMessage = e instanceof Error ? e.message : "Connection failed";
		} finally {
			testing = false;
		}
	}

	async function handleSave(e: Event) {
		e.preventDefault();

		if (browser) {
			try {
				// Save to backend only - the backend will handle authentication
				await api.saveConfig(serverUrl, username, password);

				// Update the store with new config
				resetConfigCache();
				await loadConfig();

				success = true;
				setTimeout(() => {
					success = false;
				}, 3000);
			} catch (e) {
				console.error("Failed to save config:", e);
			}
		}
	}
</script>

<svelte:head>
	<title>Config - BookScraping</title>
</svelte:head>

<div class="config-page">
	<h1>⚙️ Configuration</h1>

	<form onsubmit={handleSave} class="config-form">
		<div class="form-group">
			<label for="serverUrl">Booklore Server URL</label>
			<input
				id="serverUrl"
				type="url"
				bind:value={serverUrl}
				placeholder="https://booklore.example.com"
				required
			/>
		</div>

		<div class="form-group">
			<label for="username">Username</label>
			<input
				id="username"
				type="text"
				bind:value={username}
				placeholder="username"
				required
			/>
		</div>

		<div class="form-group">
			<label for="password">Password</label>
			<input
				id="password"
				type="password"
				bind:value={password}
				placeholder="password"
				required
			/>
		</div>

		<div class="button-group">
			<button
				type="button"
				class="test-btn"
				onclick={handleTest}
				disabled={testing}
			>
				{testing ? "Testing..." : "Test Connection"}
			</button>
			<button type="submit">Save Configuration</button>
		</div>

		{#if testMessage}
			<div class="message {testSuccess ? 'success' : 'error'}">
				{testMessage}
			</div>
		{/if}

		{#if success}
			<div class="message success">Configuration saved successfully!</div>
		{/if}
	</form>
</div>

<style>
	.config-page h1 {
		margin-bottom: 1.5rem;
	}

	.config-form {
		background: white;
		padding: 2rem;
		border-radius: 8px;
		box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
		max-width: 500px;
	}

	.form-group {
		margin-bottom: 1.5rem;
	}

	.form-group label {
		display: block;
		margin-bottom: 0.5rem;
		font-weight: 500;
		color: #2c3e50;
	}

	.form-group input {
		width: 100%;
		padding: 0.75rem;
		border: 2px solid #ddd;
		border-radius: 4px;
		font-size: 1rem;
		box-sizing: border-box;
	}

	.form-group input:focus {
		outline: none;
		border-color: #2c3e50;
	}

	.button-group {
		display: flex;
		gap: 1rem;
		margin-bottom: 1rem;
	}

	button {
		flex: 1;
		padding: 0.75rem 1.5rem;
		background: #2c3e50;
		color: white;
		border: none;
		border-radius: 4px;
		font-size: 1rem;
		cursor: pointer;
		font-weight: 500;
	}

	button:hover {
		background: #34495e;
	}

	button:disabled {
		background: #95a5a6;
		cursor: not-allowed;
	}

	.test-btn {
		background: #fff;
		color: #2c3e50;
		border: 2px solid #2c3e50;
	}

	.test-btn:hover {
		background: #f8f9fa;
	}

	.message {
		margin-top: 1rem;
		padding: 0.75rem;
		border-radius: 4px;
		text-align: center;
	}

	.success {
		background-color: #d4edda;
		border: 1px solid #c3e6cb;
		color: #155724;
	}

	.error {
		background-color: #f8d7da;
		border: 1px solid #f5c6cb;
		color: #721c24;
	}
</style>

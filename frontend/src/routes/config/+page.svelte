<script lang="ts">
	import { browser } from "$app/environment";
	import { api } from "$lib/api";

	let serverUrl = $state("");
	let username = $state("");
	let password = $state("");
	let success = $state(false);
	let testing = $state(false);
	let testMessage = $state("");
	let testSuccess = $state(false);

	$effect(() => {
		if (browser) {
			serverUrl = localStorage.getItem("serverUrl") || "";
			username = localStorage.getItem("username") || "";
			password = localStorage.getItem("password") || "";
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
			const response = await api.bookloreLogin(serverUrl, username, password);
			if (browser) {
				localStorage.setItem("accessToken", response.accessToken);
				localStorage.setItem("refreshToken", response.refreshToken);
			}
			testSuccess = true;
			testMessage = "Connection successful!";
		} catch (e) {
			testSuccess = false;
			testMessage = e instanceof Error ? e.message : "Connection failed";
		} finally {
			testing = false;
		}
	}

	async function handleSave(e: Event) {
		e.preventDefault();
		
		if (browser) {
			localStorage.setItem("serverUrl", serverUrl);
			localStorage.setItem("username", username);
			localStorage.setItem("password", password);

			// Try to login and get tokens
			try {
				const response = await api.bookloreLogin(serverUrl, username, password);
				localStorage.setItem("accessToken", response.accessToken);
				localStorage.setItem("refreshToken", response.refreshToken);
			} catch (e) {
				console.error("Failed to login during save:", e);
				// We still saved the credentials, so we consider it a partial success
				// potentially we could show a warning here
			}
			
			success = true;
			setTimeout(() => {
				success = false;
			}, 3000);

			window.dispatchEvent(new Event('storage'));
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
			<label for="serverUrl">Server URL</label>
			<input
				id="serverUrl"
				type="url"
				bind:value={serverUrl}
				placeholder="https://example.com"
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
			<button type="button" class="test-btn" onclick={handleTest} disabled={testing}>
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

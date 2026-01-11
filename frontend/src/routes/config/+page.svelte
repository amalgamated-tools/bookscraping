<script lang="ts">
	import { browser } from "$app/environment";

	let serverUrl = $state("");
	let username = $state("");
	let password = $state("");
	let success = $state(false);

	$effect(() => {
		if (browser) {
			serverUrl = localStorage.getItem("serverUrl") || "";
			username = localStorage.getItem("username") || "";
			password = localStorage.getItem("password") || "";
		}
	});

	function handleSave(e: Event) {
		e.preventDefault();
		
		if (browser) {
			localStorage.setItem("serverUrl", serverUrl);
			localStorage.setItem("username", username);
			localStorage.setItem("password", password);
			
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

		<button type="submit">Save Configuration</button>

		{#if success}
			<div class="success">Configuration saved successfully!</div>
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

	button {
		width: 100%;
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

	.success {
		margin-top: 1rem;
		padding: 0.75rem;
		background-color: #d4edda;
		border: 1px solid #c3e6cb;
		border-radius: 4px;
		color: #155724;
		text-align: center;
	}
</style>

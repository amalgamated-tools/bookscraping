<script lang="ts">
	import { page } from '$app/state';
	import { api, type Series } from '$lib/api';
	import { onMount } from 'svelte';

	let series = $state<Series | null>(null);
	let loading = $state(true);
	let error = $state<string | null>(null);

	onMount(async () => {
		const id = parseInt(page.params.id);
		try {
			series = await api.getSeriesById(id);
		} catch (e) {
			error = e instanceof Error ? e.message : 'Failed to load series';
		} finally {
			loading = false;
		}
	});
</script>

<svelte:head>
	<title>{series?.name ?? 'Series'} - BookScraping</title>
</svelte:head>

<div class="series-detail">
	<a href="/series" class="back-link">← Back to Series</a>

	{#if loading}
		<div class="loading">Loading series...</div>
	{:else if error}
		<div class="error">{error}</div>
	{:else if series}
		<div class="series-header">
			<h1>{series.name}</h1>
		</div>

		{#if series.description}
			<section class="description">
				<h2>Description</h2>
				<p>{series.description}</p>
			</section>
		{/if}

		{#if series.url}
			<section class="links">
				<h2>Links</h2>
				<a href={series.url} target="_blank" rel="noopener">
					View on Goodreads →
				</a>
			</section>
		{/if}
	{/if}
</div>

<style>
	.back-link {
		display: inline-block;
		margin-bottom: 1rem;
		color: #2c3e50;
		text-decoration: none;
	}

	.back-link:hover {
		text-decoration: underline;
	}

	.loading {
		text-align: center;
		padding: 2rem;
		color: #666;
	}

	.error {
		background-color: #fee;
		border: 1px solid #fcc;
		border-radius: 8px;
		padding: 1rem;
		color: #c00;
	}

	.series-header {
		margin-bottom: 2rem;
	}

	.series-header h1 {
		margin-bottom: 0.5rem;
	}

	section {
		background: white;
		border-radius: 8px;
		padding: 1.5rem;
		margin-bottom: 1rem;
		box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
	}

	section h2 {
		margin-top: 0;
		margin-bottom: 1rem;
		font-size: 1.2rem;
		color: #2c3e50;
	}

	.description p {
		line-height: 1.6;
		margin: 0;
	}

	.links a {
		color: #2c3e50;
		font-weight: 500;
	}
</style>

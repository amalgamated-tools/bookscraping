<script lang="ts">
	import { api, type Series } from '$lib/api';
	import { onMount } from 'svelte';

	let series = $state<Series[]>([]);
	let loading = $state(true);
	let error = $state<string | null>(null);
	let page = $state(1);
	let total = $state(0);
	let perPage = 20;

	async function loadSeries() {
		loading = true;
		try {
			const response = await api.getSeries(page, perPage);
			series = response.data;
			total = response.total;
		} catch (e) {
			error = e instanceof Error ? e.message : 'Failed to load series';
		} finally {
			loading = false;
		}
	}

	onMount(loadSeries);

	function nextPage() {
		if (page * perPage < total) {
			page++;
			loadSeries();
		}
	}

	function prevPage() {
		if (page > 1) {
			page--;
			loadSeries();
		}
	}
</script>

<svelte:head>
	<title>Series - BookScraping</title>
</svelte:head>

<div class="series-page">
	<h1>📚 Series</h1>

	{#if loading}
		<div class="loading">Loading series...</div>
	{:else if error}
		<div class="error">{error}</div>
	{:else}
		<p class="count">Showing {series.length} of {total} series</p>

		<div class="series-grid">
			{#each series as s}
				<a href="/series/{s.id}" class="series-card">
					<h3>{s.name}</h3>
					{#if s.description}
						<p class="description">{s.description.slice(0, 150)}{s.description.length > 150 ? '...' : ''}</p>
					{/if}
					{#if s.url}
						<span class="has-link">🔗 Goodreads</span>
					{/if}
				</a>
			{/each}
		</div>

		<div class="pagination">
			<button onclick={prevPage} disabled={page <= 1}>← Previous</button>
			<span>Page {page} of {Math.ceil(total / perPage)}</span>
			<button onclick={nextPage} disabled={page * perPage >= total}>Next →</button>
		</div>
	{/if}
</div>

<style>
	.series-page h1 {
		margin-bottom: 1rem;
	}

	.count {
		color: #666;
		margin-bottom: 1rem;
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

	.series-grid {
		display: grid;
		grid-template-columns: repeat(auto-fill, minmax(300px, 1fr));
		gap: 1rem;
		margin-bottom: 2rem;
	}

	.series-card {
		background: white;
		border-radius: 8px;
		padding: 1.25rem;
		box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
		text-decoration: none;
		color: inherit;
		transition: transform 0.2s, box-shadow 0.2s;
	}

	.series-card:hover {
		transform: translateY(-2px);
		box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
	}

	.series-card h3 {
		margin: 0 0 0.5rem;
		font-size: 1.1rem;
		color: #2c3e50;
	}

	.series-card .description {
		margin: 0 0 0.5rem;
		font-size: 0.9rem;
		color: #666;
		line-height: 1.4;
	}

	.has-link {
		font-size: 0.8rem;
		color: #553b08;
	}

	.pagination {
		display: flex;
		justify-content: center;
		align-items: center;
		gap: 1rem;
	}

	.pagination button {
		padding: 0.5rem 1rem;
		border: none;
		background: #2c3e50;
		color: white;
		border-radius: 4px;
		cursor: pointer;
	}

	.pagination button:disabled {
		background: #ccc;
		cursor: not-allowed;
	}

	.pagination button:not(:disabled):hover {
		background: #34495e;
	}
</style>

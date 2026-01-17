<script lang="ts">
	import { page } from '$app/state';
	import { api, type Series, type Book } from '$lib/api';
	import { onMount } from 'svelte';

	let series = $state<Series | null>(null);
	let books = $state<Book[]>([]);
	let loading = $state(true);
	let error = $state<string | null>(null);
	let pageNumber = $state(1);

	onMount(async () => {
		const id = parseInt(page.params.id ?? '0');
		const urlParams = new URLSearchParams(window.location.search);
		pageNumber = parseInt(urlParams.get('page') ?? '1');

		try {
			series = await api.getSeriesById(id);
			books = await api.getSeriesBooks(id);
			// Sort books by series number
			books.sort((a, b) => (a.series_number ?? 0) - (b.series_number ?? 0));
		} catch (e) {
			error = e instanceof Error ? e.message : 'Failed to load series';
		} finally {
			loading = false;
		}
	});

	function getBackLink(): string {
		if (pageNumber > 1) {
			return `/series?page=${pageNumber}`;
		}
		return '/series';
	}
</script>

<svelte:head>
	<title>{series?.name ?? 'Series'} - BookScraping</title>
</svelte:head>

<div class="series-detail">
	<a href={getBackLink()} class="back-link">← Back to Series</a>

	{#if loading}
		<div class="loading">Loading series...</div>
	{:else if error}
		<div class="error">{error}</div>
	{:else if series}
		<div class="series-header">
			<h1>{series.name}</h1>
			{#if series.authors && series.authors.length > 0}
				<p class="series-authors">By {series.authors.join(', ')}</p>
			{/if}
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

		{#if books.length > 0}
			<section class="books">
				<h2>Books in Series ({books.length})</h2>
				<div class="books-list">
					{#each books as book (book.id)}
						<div class="book-item">
							{#if book.series_number}
								<span class="book-number">#{book.series_number}</span>
							{/if}
							<div class="book-info">
								<h3>{book.title}</h3>
								{#if book.authors && book.authors.length > 0}
									<p class="book-authors">{book.authors.join(', ')}</p>
								{/if}
							</div>
						</div>
					{/each}
				</div>
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
		font-weight: 500;
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
		margin: 0 0 0.5rem;
		font-size: 1.8rem;
		color: #2c3e50;
	}

	.series-authors {
		margin: 0;
		font-size: 1.05rem;
		color: #5a6c7d;
		font-weight: 500;
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

	.books-list {
		display: flex;
		flex-direction: column;
		gap: 0.75rem;
	}

	.book-item {
		display: flex;
		align-items: flex-start;
		gap: 1rem;
		padding: 1rem;
		background: #f9f9f9;
		border-radius: 6px;
		border-left: 4px solid #2c3e50;
	}

	.book-number {
		display: inline-block;
		min-width: 3rem;
		padding: 0.25rem 0.5rem;
		background: #2c3e50;
		color: white;
		border-radius: 4px;
		font-size: 0.85rem;
		font-weight: 600;
		text-align: center;
		margin-top: 0.25rem;
	}

	.book-info {
		flex: 1;
		min-width: 0;
	}

	.book-info h3 {
		margin: 0 0 0.25rem;
		font-size: 1.05rem;
		color: #2c3e50;
	}

	.book-authors {
		margin: 0;
		font-size: 0.9rem;
		color: #666;
	}

	.book-item.missing {
		opacity: 0.6;
		border-left-color: #999;
	}

	.book-item.missing .book-number {
		background: #999;
	}

	.missing-badge {
		display: inline-block;
		padding: 0.25rem 0.5rem;
		background: #ddd;
		color: #666;
		border-radius: 4px;
		font-size: 0.75rem;
		font-weight: 600;
		margin-left: 0.5rem;
	}
</style>

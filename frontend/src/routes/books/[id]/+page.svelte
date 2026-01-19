<script lang="ts">
	import { page } from '$app/state';
	import { api, type Book } from '$lib/api';
	import { onMount } from 'svelte';

	let book = $state<Book | null>(null);
	let loading = $state(true);
	let error = $state<string | null>(null);

	onMount(async () => {
		const id = parseInt(page.params.id ?? '0');

		if (!id) {
			error = 'Invalid book ID';
			loading = false;
			return;
		}

		try {
			book = await api.getBook(id);
		} catch (e) {
			error = e instanceof Error ? e.message : 'Failed to load book';
		} finally {
			loading = false;
		}
	});
</script>

<svelte:head>
	<title>{book?.title ?? 'Book'} - BookScraping</title>
</svelte:head>

<div class="book-detail">
	<a href="/" class="back-link">← Back to Home</a>

	{#if loading}
		<div class="loading">Loading book...</div>
	{:else if error}
		<div class="error">{error}</div>
	{:else if book}
		<div class="book-header">
			<h1>{book.title}</h1>
			{#if book.authors && book.authors.length > 0}
				<p class="authors">By {book.authors.join(', ')}</p>
			{/if}
		</div>

		{#if book.series_name}
			<section class="series-info">
				<h2>Series Information</h2>
				<p>
					<strong>{book.series_name}</strong>
					{#if book.series_number}
						<span class="series-number">Book #{book.series_number}</span>
					{/if}
				</p>
				{#if book.series_id}
					<a href="/series/{book.series_id}">View full series →</a>
				{/if}
			</section>
		{/if}

		{#if book.description}
			<section class="description">
				<h2>Description</h2>
				<p>{book.description}</p>
			</section>
		{/if}

		<section class="details">
			<h2>Details</h2>
			<dl>
				{#if book.isbn13}
					<div class="detail-item">
						<dt>ISBN-13</dt>
						<dd>{book.isbn13}</dd>
					</div>
				{/if}
				{#if book.isbn10}
					<div class="detail-item">
						<dt>ISBN-10</dt>
						<dd>{book.isbn10}</dd>
					</div>
				{/if}
				{#if book.language}
					<div class="detail-item">
						<dt>Language</dt>
						<dd>{book.language}</dd>
					</div>
				{/if}
				{#if book.goodreads_id}
					<div class="detail-item">
						<dt>Goodreads ID</dt>
						<dd>{book.goodreads_id}</dd>
					</div>
				{/if}
				{#if book.is_missing}
					<div class="detail-item">
						<dt>Status</dt>
						<dd><span class="missing-badge">Missing (from Goodreads)</span></dd>
					</div>
				{/if}
			</dl>
		</section>
	{:else}
		<div class="not-found">Book not found</div>
	{/if}
</div>

<style>
	.book-detail {
		max-width: 800px;
		margin: 0 auto;
	}

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

	.not-found {
		text-align: center;
		padding: 2rem;
		color: #999;
		font-size: 1.1rem;
	}

	.book-header {
		margin-bottom: 2rem;
	}

	.book-header h1 {
		margin: 0 0 0.5rem;
		font-size: 2rem;
		color: #2c3e50;
	}

	.authors {
		margin: 0;
		font-size: 1.1rem;
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

	.series-info p {
		margin: 0.5rem 0;
	}

	.series-number {
		background: #f0f0f0;
		padding: 0.25rem 0.5rem;
		border-radius: 4px;
		font-size: 0.9rem;
		margin-left: 0.5rem;
	}

	.series-info a {
		color: #2c3e50;
		font-weight: 500;
		text-decoration: none;
	}

	.series-info a:hover {
		text-decoration: underline;
	}

	.description p {
		line-height: 1.6;
		margin: 0;
		color: #555;
	}

	.details dl {
		margin: 0;
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
		gap: 1rem;
	}

	.detail-item {
		border-bottom: 1px solid #f0f0f0;
		padding-bottom: 0.5rem;
	}

	.detail-item dt {
		font-weight: 600;
		color: #2c3e50;
		font-size: 0.9rem;
		text-transform: uppercase;
		letter-spacing: 0.5px;
	}

	.detail-item dd {
		margin: 0.25rem 0 0;
		color: #666;
		font-family: monospace;
		font-size: 0.9rem;
	}

	.missing-badge {
		display: inline-block;
		padding: 0.25rem 0.75rem;
		background: #fff3cd;
		border: 1px solid #ffc107;
		border-radius: 4px;
		color: #856404;
		font-size: 0.85rem;
		font-weight: 600;
	}
</style>

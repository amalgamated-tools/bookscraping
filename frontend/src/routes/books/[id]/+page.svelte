<script lang="ts">
	import { page } from '$app/state';
	import { api, type Book } from '$lib/api';
	import { onMount } from 'svelte';

	let book = $state<Book | null>(null);
	let loading = $state(true);
	let error = $state<string | null>(null);

	onMount(async () => {
		const id = parseInt(page.params.id ?? '0');
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
	<a href="/books" class="back-link">← Back to Books</a>

	{#if loading}
		<div class="loading">Loading book...</div>
	{:else if error}
		<div class="error">{error}</div>
	{:else if book}
		<div class="book-header">
			<h1>{book.title}</h1>
			{#if book.series_name}
				<p class="series">
					<a href="/series">{book.series_name}</a> #{book.series_number}
				</p>
			{/if}
		</div>

		{#if book.description}
			<section class="description">
				<h2>Description</h2>
				<p>{book.description}</p>
			</section>
		{/if}

		<section class="identifiers">
			<h2>Identifiers</h2>
			<dl>
				{#if book.isbn13}
					<dt>ISBN-13</dt>
					<dd>{book.isbn13}</dd>
				{/if}
				{#if book.isbn10}
					<dt>ISBN-10</dt>
					<dd>{book.isbn10}</dd>
				{/if}
				{#if book.asin}
					<dt>ASIN</dt>
					<dd>
						<a href="https://www.amazon.com/dp/{book.asin}" target="_blank" rel="noopener">
							{book.asin}
						</a>
					</dd>
				{/if}
				{#if book.goodreads_id}
					<dt>Goodreads</dt>
					<dd>
						<a href="https://www.goodreads.com/book/show/{book.goodreads_id}" target="_blank" rel="noopener">
							{book.goodreads_id}
						</a>
					</dd>
				{/if}
				{#if book.google_id}
					<dt>Google Books</dt>
					<dd>
						<a href="https://books.google.com/books?id={book.google_id}" target="_blank" rel="noopener">
							{book.google_id}
						</a>
					</dd>
				{/if}
				{#if book.language}
					<dt>Language</dt>
					<dd>{book.language}</dd>
				{/if}
			</dl>
		</section>
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

	.book-header {
		margin-bottom: 2rem;
	}

	.book-header h1 {
		margin-bottom: 0.5rem;
	}

	.series {
		font-size: 1.1rem;
		color: #666;
	}

	.series a {
		color: #2c3e50;
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

	dl {
		display: grid;
		grid-template-columns: auto 1fr;
		gap: 0.5rem 1rem;
		margin: 0;
	}

	dt {
		font-weight: 600;
		color: #666;
	}

	dd {
		margin: 0;
	}

	dd a {
		color: #2c3e50;
	}
</style>

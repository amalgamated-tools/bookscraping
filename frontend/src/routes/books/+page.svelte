<script lang="ts">
	import { api, type Book } from "$lib/api";
	import { onMount } from "svelte";

	let books = $state<Book[]>([]);
	let loading = $state(true);
	let error = $state<string | null>(null);
	let page = $state(1);
	let total = $state(0);
	let perPage = 20;

	async function loadBooks() {
		loading = true;
		try {
			const response = await api.getBooks(page, perPage);
			books = response.data ?? [];
			total = response.total;
		} catch (e) {
			error = e instanceof Error ? e.message : "Failed to load books";
		} finally {
			loading = false;
		}
	}

	onMount(loadBooks);

	function nextPage() {
		if (page * perPage < total) {
			page++;
			loadBooks();
		}
	}

	function prevPage() {
		if (page > 1) {
			page--;
			loadBooks();
		}
	}
</script>

<svelte:head>
	<title>Books - BookScraping</title>
</svelte:head>

<div class="books-page">
	<h1>📚 Books</h1>

	{#if loading}
		<div class="loading">Loading books...</div>
	{:else if error}
		<div class="error">{error}</div>
	{:else}
		<p class="count">Showing {books.length} of {total} books</p>

		<div class="book-grid">
			{#each books as book}
				<a href="/books/{book.id}" class="book-card">
					<h3>{book.title}</h3>
					{#if book.series_name}
						<p class="series">
							{book.series_name} #{book.series_number}
						</p>
					{/if}
					<div class="meta">
						{#if book.isbn13}
							<span class="badge">ISBN</span>
						{/if}
						{#if book.goodreads_id}
							<span class="badge goodreads">GR</span>
						{/if}
						{#if book.asin}
							<span class="badge amazon">ASIN</span>
						{/if}
					</div>
				</a>
			{/each}
		</div>

		<div class="pagination">
			<button onclick={prevPage} disabled={page <= 1}>← Previous</button>
			<span>Page {page} of {Math.ceil(total / perPage)}</span>
			<button onclick={nextPage} disabled={page * perPage >= total}
				>Next →</button
			>
		</div>
	{/if}
</div>

<style>
	.books-page h1 {
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

	.book-grid {
		display: grid;
		grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
		gap: 1rem;
		margin-bottom: 2rem;
	}

	.book-card {
		background: white;
		border-radius: 8px;
		padding: 1rem;
		box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
		text-decoration: none;
		color: inherit;
		transition:
			transform 0.2s,
			box-shadow 0.2s;
	}

	.book-card:hover {
		transform: translateY(-2px);
		box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
	}

	.book-card h3 {
		margin: 0 0 0.5rem;
		font-size: 1.1rem;
	}

	.book-card .series {
		margin: 0 0 0.5rem;
		font-size: 0.9rem;
		color: #666;
	}

	.meta {
		display: flex;
		gap: 0.5rem;
		flex-wrap: wrap;
	}

	.badge {
		font-size: 0.7rem;
		padding: 0.2rem 0.5rem;
		border-radius: 4px;
		background: #eee;
		color: #666;
	}

	.badge.goodreads {
		background: #553b08;
		color: white;
	}

	.badge.amazon {
		background: #ff9900;
		color: black;
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

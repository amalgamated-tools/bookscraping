<script lang="ts">
	import { api, type Book, type BookloreBook } from "$lib/api";
	import { onMount } from "svelte";

	let books = $state<Book[]>([]);
	let allBooks = $state<BookloreBook[]>([]); // Store all fetched books
	let loading = $state(true);
	let error = $state<string | null>(null);
	let page = $state(1);
	let total = $state(0);
	let perPage = 20;

	// New state for configuration
	let isConfigured = $state(false);

	function mapBookloreToBook(b: BookloreBook): Book {
		return {
			id: b.id,
			book_id: b.metadata.bookId,
			title: b.metadata.title,
			description: b.metadata.description || "",
			series_name: b.metadata.seriesName,
			series_number: b.metadata.seriesNumber,
			// We don't have the series ID in the book metadata from Booklore yet
			// series_id: b.metadata.seriesId, 
			asin: b.metadata.asin,
			isbn10: b.metadata.isbn10,
			isbn13: b.metadata.isbn13,
			language: b.metadata.language,
			hardcover_id: b.metadata.hardcoverId,
			hardcover_book_id: b.metadata.hardcoverBookId,
			goodreads_id: b.metadata.goodreadsId,
			google_id: b.metadata.googleId,
		};
	}

	async function loadBooks() {
		loading = true;
		error = null;

		try {
			// First check if we have configuration
			const serverUrl = localStorage.getItem("serverUrl");
			const accessToken = localStorage.getItem("accessToken");

			if (serverUrl && accessToken) {
				isConfigured = true;
				// Only fetch if we haven't already fetched all books or if we want to refresh
				if (allBooks.length === 0) {
					console.log("Fetching all books from Booklore...");
					allBooks = await api.getBookloreBooks(
						serverUrl,
						accessToken,
					);
					total = allBooks.length;
				}

				// Calculate pagination from local data
				const start = (page - 1) * perPage;
				const end = start + perPage;
				books = allBooks.slice(start, end).map(mapBookloreToBook);
			} else {
				// Fallback to local API if not configured
				isConfigured = false;
				const response = await api.getBooks(page, perPage);
				books = response.data ?? [];
				total = response.total;
			}
		} catch (e) {
			error = e instanceof Error ? e.message : "Failed to load books";
			// If Booklore fails, maybe fallback to local? For now just show error
		} finally {
			loading = false;
		}
	}

	onMount(() => {
		loadBooks();
	});

	function nextPage() {
		if (page * perPage < total) {
			page++;
			// If configured, we just slice the array, otherwise we need to fetch
			if (isConfigured) {
				const start = (page - 1) * perPage;
				const end = start + perPage;
				books = allBooks.slice(start, end).map(mapBookloreToBook);
			} else {
				loadBooks();
			}
		}
	}

	function prevPage() {
		if (page > 1) {
			page--;
			// If configured, we just slice the array, otherwise we need to fetch
			if (isConfigured) {
				const start = (page - 1) * perPage;
				const end = start + perPage;
				books = allBooks.slice(start, end).map(mapBookloreToBook);
			} else {
				loadBooks();
			}
		}
	}

	// Function to force refresh from API
	function handleRefresh() {
		allBooks = []; // Clear cache to force refetch
		page = 1;
		loadBooks();
	}
</script>

<svelte:head>
	<title>Books - BookScraping</title>
</svelte:head>

<div class="books-page">
	<div class="header">
		<h1>📚 Books</h1>
		<div class="header-actions">
			{#if isConfigured}
				<span class="source-badge">Using Booklore API</span>
			{/if}
			<button
				onclick={handleRefresh}
				disabled={loading}
				class="refresh-btn"
			>
				{loading ? "Refreshing..." : "🔄 Refresh"}
			</button>
		</div>
	</div>

	{#if loading}
		<div class="loading">Loading books...</div>
	{:else if error}
		<div class="error">{error}</div>
	{:else}
		<p class="count">Showing {books.length} of {total} books</p>

		<div class="book-grid">
			{#each books as book}
				<div class="book-card-wrapper">
					<div class="book-card-content">
						<a href="/books/{book.id}" class="title-link">
							<h3>{book.title}</h3>
						</a>
						{#if book.series_name}
							<p class="series">
								{#if book.series_id}
									<a href="/series/{book.series_id}" class="series-link">
										{book.series_name}
									</a>
								{:else}
									{book.series_name}
								{/if}
								#{book.series_number}
							</p>
						{/if}
					</div>
					<div class="meta">
						{#if book.isbn13}
							<span class="badge" title={book.isbn13}>ISBN</span>
						{/if}
						{#if book.goodreads_id}
							<a
								href="https://www.goodreads.com/book/show/{book.goodreads_id}"
								target="_blank"
								rel="noopener noreferrer"
								class="badge goodreads"
								title={book.goodreads_id}
							>
								GR
							</a>
						{/if}
						{#if book.asin}
							<span class="badge amazon" title={book.asin}>ASIN</span>
						{/if}
						{#if book.hardcover_id}
							<a
								href="https://hardcover.app/books/{book.hardcover_id}"
								target="_blank"
								rel="noopener noreferrer"
								class="badge hardcover"
								title={book.hardcover_id}
							>
								HC
							</a>
						{/if}
					</div>
				</div>
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
	.header {
		display: flex;
		justify-content: space-between;
		align-items: center;
		margin-bottom: 1rem;
	}

	.header-actions {
		display: flex;
		align-items: center;
		gap: 1rem;
	}

	.source-badge {
		font-size: 0.8rem;
		padding: 0.25rem 0.5rem;
		background: #e2e8f0;
		color: #4a5568;
		border-radius: 4px;
		font-weight: 500;
	}

	.books-page h1 {
		margin: 0;
	}

	.refresh-btn {
		padding: 0.5rem 1rem;
		border: 2px solid #2c3e50;
		background: white;
		color: #2c3e50;
		border-radius: 4px;
		cursor: pointer;
		font-size: 0.9rem;
		font-weight: 500;
	}

	.refresh-btn:hover:not(:disabled) {
		background: #2c3e50;
		color: white;
	}

	.refresh-btn:disabled {
		border-color: #ccc;
		color: #ccc;
		cursor: not-allowed;
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

	.book-card-wrapper {
		background: white;
		border-radius: 8px;
		padding: 1rem;
		box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
		transition:
			transform 0.2s,
			box-shadow 0.2s;
		display: flex;
		flex-direction: column;
		justify-content: space-between;
	}

	.book-card-wrapper:hover {
		transform: translateY(-2px);
		box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
	}

	.book-card-content {
		margin-bottom: 0.5rem;
	}

	.title-link {
		text-decoration: none;
		color: inherit;
		display: block;
	}

	.title-link h3 {
		margin: 0 0 0.5rem;
		font-size: 1.1rem;
	}

	.series {
		margin: 0 0 0.5rem;
		font-size: 0.9rem;
		color: #666;
	}

	.series-link {
		color: #2c3e50;
		text-decoration: none;
		font-weight: 500;
	}

	.series-link:hover {
		text-decoration: underline;
	}

	.meta {
		display: flex;
		gap: 0.5rem;
		flex-wrap: wrap;
		margin-top: auto;
	}

	.badge {
		font-size: 0.7rem;
		padding: 0.2rem 0.5rem;
		border-radius: 4px;
		background: #eee;
		color: #666;
		text-decoration: none;
		display: inline-block;
	}

	.badge.goodreads {
		background: #553b08;
		color: white;
	}

	.badge.amazon {
		background: #ff9900;
		color: black;
	}

	.badge.hardcover {
		background: #1e3a8a;
		color: white;
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

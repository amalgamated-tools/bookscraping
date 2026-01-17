<script lang="ts">
	import { api, type Book, type BookloreBook } from "$lib/api";
	import { websocketStore } from "$lib/stores/websocketStore";
	import { onMount } from "svelte";

	let books = $state<Book[]>([]);
	let allBooks = $state<BookloreBook[]>([]); // Store all fetched books
	let loading = $state(true);
	let error = $state<string | null>(null);
	let page = $state(1);
	let total = $state(0);
	let perPage = 21;

	// New state for configuration
	let isConfigured = $state(false);

	let syncing = $state(false);
	let syncProgress: string[] = $state([]);

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
			// We always just load from our own API now
			isConfigured = true;
			
			const response = await api.getBooks(page, perPage);
			books = response.data ?? [];
			total = response.total;
		} catch (e) {
			error = e instanceof Error ? e.message : "Failed to load books";
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
			loadBooks();
		}
	}

	function prevPage() {
		if (page > 1) {
			page--;
			loadBooks();
		}
	}

	// Function to force refresh from API
	async function handleRefresh() {
		loading = true;
		error = null;
		
		try {
			// Reload the local book list
			page = 1;
			await loadBooks();
		} catch (e) {
			error = e instanceof Error ? e.message : "Failed to refresh books";
			loading = false;
		}
	}

	async function handleSync() {
		syncing = true;
		syncProgress = [];
		error = null;

		try {
			// Subscribe to SSE events
			const unsubscribe = websocketStore.subscribe((state) => {
				if (state.lastMessage && state.lastMessage !== "connected") {
					syncProgress = [...syncProgress, state.lastMessage];
					// Auto-scroll to bottom
					setTimeout(() => {
						const progressDiv =
							document.querySelector(".sync-progress");
						if (progressDiv) {
							progressDiv.scrollTop = progressDiv.scrollHeight;
						}
					}, 0);
				}
			});

			// Trigger sync
			await api.syncBooks();

			// Wait a moment for final events to arrive
			await new Promise((resolve) => setTimeout(resolve, 500));

			unsubscribe();

			// Reload books after sync completes
			page = 1;
			await loadBooks();
		} catch (e) {
			error = e instanceof Error ? e.message : "Failed to sync";
		} finally {
			syncing = false;
		}
	}
</script>

<svelte:head>
	<title>Books - BookScraping</title>
</svelte:head>

<div class="books-page">
	<div class="header">
		<h1>📚 Books</h1>
		<div class="header-actions">
			<button
				onclick={handleRefresh}
				disabled={loading || syncing}
				class="refresh-btn"
			>
				{loading ? "Refreshing..." : "🔄 Refresh"}
			</button>
			<button onclick={handleSync} disabled={syncing} class="sync-btn">
				{syncing ? "Syncing..." : "⬇️ Sync"}
			</button>
		</div>
	</div>

	{#if syncing}
		<div class="sync-container">
			<h2>Sync Progress</h2>
			<div class="sync-progress">
				{#each syncProgress as message}
					<div class="progress-line">{message}</div>
				{/each}
			</div>
		</div>
	{/if}

	{#if loading}
		<div class="loading">Loading books...</div>
	{:else if error}
		<div class="error">{error}</div>
	{:else}
		<p class="count">
			Showing {(page - 1) * perPage + 1}-{Math.min(
				page * perPage,
				total,
			)} of {total} books
		</p>

		<div class="book-grid">
			{#each books as book}
				<div class="book-card-wrapper">
					<div class="book-card-header">
						{#if book.authors && book.authors.length > 0}
							<div class="author-section">
								<span class="author-label">By</span>
								<p class="author">{book.authors.join(", ")}</p>
							</div>
						{/if}
						<a href="/books/{book.id}" class="title-link">
							<h3>{book.title}</h3>
						</a>
					</div>

					<div class="book-card-content">
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
		gap: 0.5rem;
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

	.refresh-btn,
	.sync-btn {
		padding: 0.5rem 1rem;
		border: 2px solid #2c3e50;
		background: white;
		color: #2c3e50;
		border-radius: 4px;
		cursor: pointer;
		font-size: 0.9rem;
		font-weight: 500;
	}

	.sync-btn {
		border-color: #27ae60;
		color: #27ae60;
	}

	.refresh-btn:hover:not(:disabled),
	.sync-btn:hover:not(:disabled) {
		background: #2c3e50;
		color: white;
	}

	.sync-btn:hover:not(:disabled) {
		background: #27ae60;
		border-color: #27ae60;
	}

	.refresh-btn:disabled,
	.sync-btn:disabled {
		border-color: #ccc;
		color: #ccc;
		cursor: not-allowed;
	}

	.sync-container {
		background: #f5f5f5;
		border-radius: 8px;
		padding: 1rem;
		margin-bottom: 2rem;
	}

	.sync-container h2 {
		margin: 0 0 1rem 0;
		font-size: 1rem;
		color: #2c3e50;
	}

	.sync-progress {
		background: white;
		border: 1px solid #ddd;
		border-radius: 4px;
		padding: 1rem;
		height: 300px;
		overflow-y: auto;
		font-family: monospace;
		font-size: 0.85rem;
	}

	.progress-line {
		padding: 0.25rem 0;
		color: #333;
	}

	.progress-line:last-child {
		color: #27ae60;
		font-weight: bold;
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

	.book-card-header {
		margin-bottom: 0.75rem;
		border-bottom: 2px solid #e8f0f7;
		padding-bottom: 0.75rem;
	}

	.author-section {
		background: linear-gradient(135deg, #e8f0f7 0%, #f5f9fc 100%);
		padding: 0.5rem 0.75rem;
		border-radius: 6px;
		border-left: 3px solid #2c3e50;
		margin-bottom: 0.5rem;
	}

	.author-label {
		font-size: 0.7rem;
		color: #7f8c8d;
		text-transform: uppercase;
		font-weight: 600;
		letter-spacing: 0.5px;
	}

	.author {
		margin: 0.25rem 0 0 0;
		font-size: 1rem;
		color: #2c3e50;
		font-weight: 600;
		line-height: 1.3;
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
		margin: 0;
		font-size: 1.05rem;
		font-weight: 600;
		color: #1a1a1a;
		line-height: 1.4;
	}

	.series {
		margin: 0.5rem 0 0 0;
		font-size: 0.85rem;
		color: #7f8c8d;
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

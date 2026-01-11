<script lang="ts">
	import { api, type Book } from '$lib/api';

	let query = $state('');
	let results = $state<Book[]>([]);
	let loading = $state(false);
	let error = $state<string | null>(null);
	let searched = $state(false);
	let searchType = $state<'local' | 'goodreads'>('local');

	async function handleSearch(e: Event) {
		e.preventDefault();
		if (!query.trim()) return;

		loading = true;
		error = null;
		searched = true;

		try {
			if (searchType === 'local') {
				results = await api.searchBooks(query);
			} else {
				results = await api.searchGoodreads(query);
			}
		} catch (e) {
			error = e instanceof Error ? e.message : 'Search failed';
			results = [];
		} finally {
			loading = false;
		}
	}
</script>

<svelte:head>
	<title>Search - BookScraping</title>
</svelte:head>

<div class="search-page">
	<h1>🔍 Search</h1>

	<form onsubmit={handleSearch} class="search-form">
		<div class="search-type">
			<label>
				<input type="radio" bind:group={searchType} value="local" />
				Local Library
			</label>
			<label>
				<input type="radio" bind:group={searchType} value="goodreads" />
				Goodreads
			</label>
		</div>
		<div class="search-input">
			<input
				type="text"
				bind:value={query}
				placeholder="Search for books..."
				class="search-box"
			/>
			<button type="submit" disabled={loading}>
				{loading ? 'Searching...' : 'Search'}
			</button>
		</div>
	</form>

	{#if error}
		<div class="error">{error}</div>
	{/if}

	{#if searched && !loading}
		<div class="results">
			<h2>Results ({results.length})</h2>
			{#if results.length === 0}
				<p class="no-results">No books found for "{query}"</p>
			{:else}
				<div class="book-grid">
					{#each results as book}
						<a href="/books/{book.id}" class="book-card">
							<h3>{book.title}</h3>
							{#if book.series_name}
								<p class="series">{book.series_name} #{book.series_number}</p>
							{/if}
							{#if book.description}
								<p class="description">{book.description.slice(0, 100)}...</p>
							{/if}
						</a>
					{/each}
				</div>
			{/if}
		</div>
	{/if}
</div>

<style>
	.search-page h1 {
		margin-bottom: 1.5rem;
	}

	.search-form {
		background: white;
		padding: 1.5rem;
		border-radius: 8px;
		box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
		margin-bottom: 2rem;
	}

	.search-type {
		display: flex;
		gap: 1.5rem;
		margin-bottom: 1rem;
	}

	.search-type label {
		display: flex;
		align-items: center;
		gap: 0.5rem;
		cursor: pointer;
	}

	.search-input {
		display: flex;
		gap: 0.5rem;
	}

	.search-box {
		flex: 1;
		padding: 0.75rem 1rem;
		border: 2px solid #ddd;
		border-radius: 4px;
		font-size: 1rem;
	}

	.search-box:focus {
		outline: none;
		border-color: #2c3e50;
	}

	.search-form button {
		padding: 0.75rem 1.5rem;
		background: #2c3e50;
		color: white;
		border: none;
		border-radius: 4px;
		font-size: 1rem;
		cursor: pointer;
	}

	.search-form button:hover:not(:disabled) {
		background: #34495e;
	}

	.search-form button:disabled {
		background: #ccc;
	}

	.error {
		background-color: #fee;
		border: 1px solid #fcc;
		border-radius: 8px;
		padding: 1rem;
		color: #c00;
		margin-bottom: 1rem;
	}

	.results h2 {
		margin-bottom: 1rem;
	}

	.no-results {
		color: #666;
		text-align: center;
		padding: 2rem;
	}

	.book-grid {
		display: grid;
		grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
		gap: 1rem;
	}

	.book-card {
		background: white;
		border-radius: 8px;
		padding: 1rem;
		box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
		text-decoration: none;
		color: inherit;
		transition: transform 0.2s, box-shadow 0.2s;
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

	.book-card .description {
		margin: 0;
		font-size: 0.85rem;
		color: #888;
		line-height: 1.4;
	}
</style>

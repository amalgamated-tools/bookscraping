<script lang="ts">
	import { api, type Book } from '$lib/api';
	import { onMount } from 'svelte';

	let bookCount = $state(0);
	let seriesCount = $state(0);
	let recentBooks = $state<Book[]>([]);
	let loading = $state(true);
	let error = $state<string | null>(null);

	onMount(async () => {
		try {
			const [booksRes, seriesRes] = await Promise.all([
				api.getBooks(1, 5),
				api.getSeries(1, 5)
			]);
			bookCount = booksRes.total;
			seriesCount = seriesRes.total;
			recentBooks = booksRes.data;
		} catch (e) {
			error = e instanceof Error ? e.message : 'Failed to load data';
		} finally {
			loading = false;
		}
	});
</script>

<svelte:head>
	<title>BookScraping - Home</title>
</svelte:head>

<div class="home">
	<h1>📚 BookScraping</h1>
	<p class="subtitle">A Goodreads scraping and book management application</p>

	{#if loading}
		<div class="loading">Loading...</div>
	{:else if error}
		<div class="error">
			<p>{error}</p>
			<p class="hint">Make sure the API server is running.</p>
		</div>
	{:else}
		<div class="stats">
			<div class="stat-card">
				<span class="stat-number">{bookCount}</span>
				<span class="stat-label">Books</span>
			</div>
			<div class="stat-card">
				<span class="stat-number">{seriesCount}</span>
				<span class="stat-label">Series</span>
			</div>
		</div>

		{#if recentBooks.length > 0}
			<section class="recent-books">
				<h2>Recent Books</h2>
				<div class="book-grid">
					{#each recentBooks as book}
						<a href="/books/{book.id}" class="book-card">
							<h3>{book.title}</h3>
							{#if book.series_name}
								<p class="series">{book.series_name} #{book.series_number}</p>
							{/if}
						</a>
					{/each}
				</div>
			</section>
		{/if}
	{/if}

	<section class="features">
		<h2>Features</h2>
		<ul>
			<li>📖 Browse and search your book collection</li>
			<li>🔍 Search Goodreads for book information</li>
			<li>📚 Track book series and reading order</li>
			<li>🔗 Link books with Goodreads, ISBN, and more</li>
		</ul>
	</section>
</div>

<style>
	.home {
		text-align: center;
	}

	h1 {
		font-size: 2.5rem;
		margin-bottom: 0.5rem;
	}

	.subtitle {
		color: #666;
		font-size: 1.2rem;
		margin-bottom: 2rem;
	}

	.loading {
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

	.hint {
		font-size: 0.9rem;
		color: #666;
	}

	.stats {
		display: flex;
		gap: 2rem;
		justify-content: center;
		margin-bottom: 2rem;
	}

	.stat-card {
		background: white;
		border-radius: 12px;
		padding: 1.5rem 3rem;
		box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
		display: flex;
		flex-direction: column;
	}

	.stat-number {
		font-size: 2.5rem;
		font-weight: bold;
		color: #2c3e50;
	}

	.stat-label {
		color: #666;
		text-transform: uppercase;
		font-size: 0.8rem;
		letter-spacing: 1px;
	}

	.recent-books {
		margin: 2rem 0;
		text-align: left;
	}

	.book-grid {
		display: grid;
		grid-template-columns: repeat(auto-fill, minmax(200px, 1fr));
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
		font-size: 1rem;
	}

	.book-card .series {
		margin: 0;
		font-size: 0.85rem;
		color: #666;
	}

	.features {
		text-align: left;
		background: white;
		border-radius: 12px;
		padding: 1.5rem 2rem;
		box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
	}

	.features h2 {
		margin-top: 0;
	}

	.features ul {
		list-style: none;
		padding: 0;
	}

	.features li {
		padding: 0.5rem 0;
		font-size: 1.1rem;
	}
</style>

<script lang="ts">
    import { api, type SeriesWithStats } from "$lib/api";
    import { browser } from "$app/environment";
    import { onMount } from "svelte";
    import { configStore } from "$lib/stores/configStore";

    let allSeries = $state<SeriesWithStats[]>([]);
    let incompleteSeries = $state<SeriesWithStats[]>([]);
    let loading = $state(true);
    let error = $state<string | null>(null);
    let isConfigured = $state(false);

    onMount(() => {
        if (!browser) {
            return;
        }

        // Subscribe to config changes - config is already loaded by layout
        const unsubscribe = configStore.subscribe((config) => {
            isConfigured = !!(
                config.serverUrl &&
                config.username &&
                config.password
            );
        });

        if (!isConfigured) {
            loading = false;
            return unsubscribe;
        }

        // Load data if configured
        (async () => {
            try {
                console.log("Loading series data...");
                const seriesRes = await api.getSeriesWithStats(1, 100);
                allSeries = seriesRes.data ?? [];

                // Find incomplete series (those with missing books > 0)
                incompleteSeries = allSeries
                    .filter((series) => series.missing_books > 0)
                    .sort((a, b) => b.missing_books - a.missing_books);
            } catch (e) {
                console.error("Failed to load series data", e);
                error = e instanceof Error ? e.message : "Failed to load data";
            } finally {
                loading = false;
            }
        })();

        return unsubscribe;
    });
</script>

<svelte:head>
    <title>BookScraping - Home</title>
</svelte:head>

<div class="home">
    <h1>📚 BookScraping</h1>

    {#if !isConfigured}
        <div class="config-notice">
            <h2>⚙️ Configuration Required</h2>
            <p>
                Please <a href="/config">configure your server settings</a> to get
                started.
            </p>
        </div>
    {:else if loading}
        <div class="loading">Loading series data...</div>
    {:else if error}
        <div class="error">
            <p>{error}</p>
            <p class="hint">Make sure the API server is running.</p>
        </div>
    {:else}
        <div class="stats">
            <div class="stat-card">
                <span class="stat-number">{allSeries.length}</span>
                <span class="stat-label">Total Series</span>
            </div>
            <div class="stat-card">
                <span class="stat-number">{incompleteSeries.length}</span>
                <span class="stat-label">Incomplete</span>
            </div>
        </div>

        {#if incompleteSeries.length > 0}
            <section class="series-section incomplete">
                <h2>📖 Incomplete Series</h2>
                <p class="section-subtitle">Series with missing books</p>
                <div class="series-grid">
                    {#each incompleteSeries.slice(0, 6) as series}
                        <div class="series-card">
                            <div class="series-header">
                                <h3>{series.name}</h3>
                                <span class="missing-badge">{series.missing_books} missing</span>
                            </div>
                            {#if series.description}
                                <p class="description">{series.description.substring(0, 100)}...</p>
                            {/if}
                            <div class="series-stats">
                                <span>{series.total_books} books total</span>
                                <span>{series.total_books - series.missing_books} owned</span>
                            </div>
                            <a href="/series/{series.id}" class="action-btn">View Series</a>
                        </div>
                    {/each}
                </div>
                {#if incompleteSeries.length > 6}
                    <p class="view-all"><a href="/series">View all {incompleteSeries.length} incomplete series</a></p>
                {/if}
            </section>
        {/if}

        {#if allSeries.length > 0}
            <section class="series-section all">
                <h2>📚 All Series</h2>
                <p class="section-subtitle">Browse your complete series collection</p>
                <div class="series-grid">
                    {#each allSeries.slice(0, 6) as series}
                        <div class="series-card">
                            <div class="series-header">
                                <h3>{series.name}</h3>
                            </div>
                            {#if series.description}
                                <p class="description">{series.description.substring(0, 100)}...</p>
                            {/if}
                            <div class="series-stats">
                                <span>{series.total_books} books</span>
                                {#if series.authors && series.authors.length > 0}
                                    <span>{series.authors.join(", ")}</span>
                                {/if}
                            </div>
                            <a href="/series/{series.id}" class="action-btn">View Series</a>
                        </div>
                    {/each}
                </div>
                <p class="view-all"><a href="/series">Browse all {allSeries.length} series</a></p>
            </section>
        {/if}
    {/if}
</div>

<style>
    .home {
        text-align: center;
    }

    h1 {
        font-size: 2.5rem;
        margin-bottom: 0.5rem;
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

    .config-notice {
        background-color: #fff3cd;
        border: 1px solid #ffc107;
        border-radius: 8px;
        padding: 2rem;
        margin: 2rem auto;
        max-width: 500px;
    }

    .config-notice h2 {
        margin-top: 0;
        color: #856404;
    }

    .config-notice p {
        margin-bottom: 0;
        color: #856404;
    }

    .config-notice a {
        color: #856404;
        font-weight: 600;
        text-decoration: underline;
    }

    .config-notice a:hover {
        color: #533f03;
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

    .series-section {
        margin: 2rem 0;
        text-align: left;
    }

    .series-section h2 {
        margin-bottom: 0.25rem;
        font-size: 1.5rem;
    }

    .section-subtitle {
        color: #666;
        font-size: 0.9rem;
        margin: 0 0 1.5rem 0;
    }

    .series-grid {
        display: grid;
        grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
        gap: 1.5rem;
        margin-bottom: 1rem;
    }

    .series-card {
        background: white;
        border-radius: 8px;
        padding: 1.5rem;
        box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
        transition: box-shadow 0.2s;
        display: flex;
        flex-direction: column;
        height: 100%;
    }

    .series-card:hover {
        box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
    }

    .series-header {
        display: flex;
        justify-content: space-between;
        align-items: flex-start;
        gap: 1rem;
        margin-bottom: 0.75rem;
    }

    .series-card h3 {
        margin: 0;
        font-size: 1.1rem;
        flex: 1;
        text-align: left;
    }

    .missing-badge {
        background-color: #ff6b6b;
        color: white;
        padding: 0.25rem 0.75rem;
        border-radius: 12px;
        font-size: 0.75rem;
        font-weight: 600;
        white-space: nowrap;
    }

    .description {
        margin: 0.5rem 0;
        font-size: 0.9rem;
        color: #666;
        line-height: 1.4;
    }

    .series-stats {
        display: flex;
        flex-direction: column;
        gap: 0.5rem;
        margin: 1rem 0;
        font-size: 0.85rem;
        color: #888;
        flex: 1;
    }

    .action-btn {
        display: inline-block;
        margin-top: auto;
        padding: 0.5rem 1rem;
        background: #2c3e50;
        color: white;
        border-radius: 4px;
        text-decoration: none;
        font-size: 0.9rem;
        font-weight: 500;
        transition: background 0.2s;
    }

    .action-btn:hover {
        background: #34495e;
    }

    .view-all {
        text-align: center;
        margin-top: 1rem;
    }

    .view-all a {
        color: #2c3e50;
        text-decoration: none;
        font-weight: 500;
    }

    .view-all a:hover {
        text-decoration: underline;
    }

    .incomplete {
        background: white;
        border-radius: 12px;
        padding: 2rem;
        box-shadow: 0 2px 8px rgba(0, 0, 0, 0.05);
    }

    .all {
        background: white;
        border-radius: 12px;
        padding: 2rem;
        box-shadow: 0 2px 8px rgba(0, 0, 0, 0.05);
    }
</style>

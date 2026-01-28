<script lang="ts">
    import { api, type Series } from "$lib/api";
    import { onMount } from "svelte";

    let series = $state<Series[]>([]);
    let loading = $state(true);
    let error = $state<string | null>(null);
    let page = $state(1);
    let total = $state(0);
    let perPage = 20;
    let syncing = $state(false);
    let syncProgress = $state<{ status: string; progress?: number } | null>(null);
    let syncMessage = $state<string | null>(null);
    let eventSource: EventSource | null = null;

    async function loadSeries() {
        loading = true;
        try {
            const response = await api.getSeries(page, perPage);
            series = response.data ?? [];
            total = response.total;
        } catch (e) {
            error = e instanceof Error ? e.message : "Failed to load series";
        } finally {
            loading = false;
        }
    }

    async function syncSeries() {
        syncing = true;
        syncProgress = null;
        syncMessage = null;
        error = null;
        try {
            await api.syncBooks();
        } catch (e) {
            error = e instanceof Error ? e.message : "Failed to sync series";
            syncing = false;
        }
    }

    function connectToSSE() {
        try {
            eventSource = new EventSource("/api/events");

            eventSource.onmessage = (event) => {
                try {
                    const data = JSON.parse(event.data);
                    handleSyncEvent(data);
                } catch (e) {
                    console.error("Failed to parse SSE event:", e);
                }
            };

            eventSource.onerror = (error) => {
                console.error("SSE connection error:", error);
                disconnectFromSSE();
            };
        } catch (e) {
            console.error("Failed to connect to SSE:", e);
        }
    }

    function disconnectFromSSE() {
        if (eventSource) {
            eventSource.close();
            eventSource = null;
        }
    }

    function handleSyncEvent(event: any) {
        console.log("Received sync event:", event);

        if (event.type === "sync_started") {
            syncProgress = { status: "Starting sync..." };
        } else if (event.type === "sync_progress") {
            const progress = event.progress || 0;
            syncProgress = {
                status: event.message,
                progress: progress,
            };
        } else if (event.type === "sync_complete") {
            syncProgress = {
                status: "Sync completed!",
                progress: 100,
            };
            // Reload series after a brief delay
            setTimeout(async () => {
                try {
                    await loadSeries();
                    syncMessage = `Synced successfully! (${event.synced_books || 0} books)`;
                    syncProgress = null;
                    syncing = false;
                    setTimeout(() => {
                        syncMessage = null;
                    }, 5000);
                } catch (e) {
                    console.error("Failed to reload series after sync:", e);
                    syncing = false;
                }
            }, 500);
        } else if (event.type === "sync_error") {
            error = event.message;
            syncProgress = null;
            syncing = false;
        }
    }

    onMount(() => {
        loadSeries();
        connectToSSE();

        return () => {
            disconnectFromSSE();
        };
    });

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
    <div class="header">
        <h1>📚 Series</h1>
        <button onclick={syncSeries} disabled={loading || syncing} class="refresh-btn">
            {syncing ? "Syncing..." : "🔄 Sync"}
        </button>
    </div>

    {#if error && !syncProgress}
        <div class="error">{error}</div>
    {/if}

    {#if syncMessage}
        <div class="success">{syncMessage}</div>
    {/if}

    {#if syncProgress}
        <div class="sync-progress">
            <div class="progress-message">{syncProgress.status}</div>
            {#if syncProgress.progress !== undefined}
                <div class="progress-bar">
                    <div class="progress-fill" style="width: {syncProgress.progress}%"></div>
                </div>
                <div class="progress-text">{Math.round(syncProgress.progress)}%</div>
            {/if}
        </div>
    {/if}

    {#if loading && !syncing}
        <div class="loading">Loading series...</div>
    {:else if !error && !syncProgress}
        <p class="count">Showing {series.length} of {total} series</p>

        <div class="series-grid">
            {#each series as s}
                <a href="/series/{s.id}" class="series-card">
                    <h3>{s.name}</h3>
                    {#if s.authors && s.authors.length > 0}
                        <p class="author">By {s.authors.join(", ")}</p>
                    {/if}
                    {#if s.description}
                        <p class="description">
                            {s.description.slice(0, 150)}{s.description.length >
                            150
                                ? "..."
                                : ""}
                        </p>
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

    .series-page h1 {
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
        margin-bottom: 1rem;
    }

    .success {
        background-color: #efe;
        border: 1px solid #cfc;
        border-radius: 8px;
        padding: 1rem;
        color: #060;
        margin-bottom: 1rem;
    }

    .sync-progress {
        background-color: #e3f2fd;
        border: 1px solid #90caf9;
        border-radius: 8px;
        padding: 1.5rem;
        color: #1565c0;
        margin-bottom: 1rem;
    }

    .progress-message {
        margin-bottom: 1rem;
        font-weight: 500;
    }

    .progress-bar {
        width: 100%;
        height: 24px;
        background-color: #bbdefb;
        border-radius: 4px;
        overflow: hidden;
        margin-bottom: 0.5rem;
    }

    .progress-fill {
        height: 100%;
        background-color: #2196f3;
        transition: width 0.3s ease;
        display: flex;
        align-items: center;
        justify-content: center;
        color: white;
        font-size: 0.75rem;
        font-weight: 600;
    }

    .progress-text {
        text-align: center;
        font-size: 0.9rem;
        font-weight: 500;
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
        transition:
            transform 0.2s,
            box-shadow 0.2s;
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

    .series-card .author {
        margin: 0.25rem 0 0.5rem;
        font-size: 0.95rem;
        color: #5a6c7d;
        font-weight: 500;
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

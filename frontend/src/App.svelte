<script lang="ts">
    import { onMount } from 'svelte';
    
    // --- Type Definitions ---
    
    interface AppSummary {
        flatpakAppId: string;
        name: string;
        summary: string;
        iconUrl: string;
        version: string;
        developer: string;
    }

    type ProgressStatus = 'starting' | 'downloading' | 'installing' | 'completed' | 'removing' | 'error';

    interface AppProgress {
        status: ProgressStatus;
        percentage: number;
    }

    interface ProgressPayload {
        appId: string;
        status: ProgressStatus;
        percentage: number;
    }

    interface Category {
        id: string;
        label: string;
    }

    // --- Wails Bindings ---
    // Note: In a strict Wails TS setup, you can also import these directly from '../wailsjs/go/main/App'
    const wailsApp = (window as any).go.main.App;
    const runtime = (window as any).runtime;

    const categories: Category[] = [
        { id: 'AudioVideo', label: 'Audio & Video' },
        { id: 'Development', label: 'Development' },
        { id: 'Education', label: 'Education' },
        { id: 'Game', label: 'Games' },
        { id: 'Graphics', label: 'Graphics' },
        { id: 'Network', label: 'Internet & Network' },
        { id: 'Office', label: 'Office' },
        { id: 'Science', label: 'Science' },
        { id: 'System', label: 'System Utilities' }
    ];

    // --- Reactive State ---
    
    let apps: AppSummary[] = [];
    let viewTitle: string = 'Discover';
    let activeCategory: string | null = 'Discover';
    let searchQuery: string = '';
    let isLoading: boolean = true;
    let errorMessage: string = '';
    
    let catalogStatus: string = 'Catalog Syncing...';
    let isCatalogError: boolean = false;

    // Record maps string App IDs to their specific installation progress
    let appProgress: Record<string, AppProgress> = {};

    let searchTimeout: ReturnType<typeof setTimeout>;

    onMount(() => {
        // 1. Listen for background catalog sync
        runtime.EventsOn("catalog:ready", () => {
            catalogStatus = 'Catalog Synced ✓';
            isCatalogError = false;
        });

        runtime.EventsOn("catalog:error", (err: string) => {
            catalogStatus = 'Sync Error';
            isCatalogError = true;
            console.error("Catalog sync failed:", err);
        });

        // 2. Listen for real-time Flatpak execution progress
        runtime.EventsOn("flatpak:progress", (payload: ProgressPayload) => {
            appProgress = {
                ...appProgress,
                [payload.appId]: {
                    status: payload.status,
                    percentage: payload.percentage
                }
            };
        });

        loadDiscover();
    });

    // --- Data Fetching Methods ---

    async function loadDiscover(): Promise<void> {
        activeCategory = 'Discover';
        viewTitle = 'Discover';
        isLoading = true;
        errorMessage = '';
        try {
            apps = await wailsApp.GetDiscoverApps() || [];
        } catch (err) {
            errorMessage = String(err);
        } finally {
            isLoading = false;
        }
    }

    async function loadCategory(catId: string, catLabel: string): Promise<void> {
        activeCategory = catId;
        viewTitle = `Category: ${catLabel}`;
        searchQuery = ''; 
        isLoading = true;
        errorMessage = '';
        try {
            apps = await wailsApp.GetAppsByCategory(catId) || [];
        } catch (err) {
            errorMessage = String(err);
        } finally {
            isLoading = false;
        }
    }

    function handleSearch(): void {
        clearTimeout(searchTimeout);
        activeCategory = null; 

        searchTimeout = setTimeout(async () => {
            const query = searchQuery.trim();
            if (query.length > 2) {
                viewTitle = `Search: "${query}"`;
                isLoading = true;
                try {
                    apps = await wailsApp.SearchApps(query) || [];
                } catch (err) {
                    errorMessage = String(err);
                } finally {
                    isLoading = false;
                }
            } else if (query.length === 0) {
                loadDiscover();
            }
        }, 400);
    }

    // --- System Actions ---

    function handleInstall(appId: string): void {
        appProgress = { ...appProgress, [appId]: { status: 'starting', percentage: 0 } };
        wailsApp.InstallApp(appId, true).catch((err: Error) => console.error(err));
    }

    function handleUninstall(appId: string): void {
        appProgress = { ...appProgress, [appId]: { status: 'removing', percentage: 0 } };
        wailsApp.UninstallApp(appId, true).catch((err: Error) => console.error(err));
    }

    // Helper interface for the UI button state
    interface ButtonProps {
        text: string;
        disabled: boolean;
        class: string;
        action?: () => void;
    }

    function getButtonProps(appId: string): ButtonProps {
        const prog = appProgress[appId];
        if (!prog) return { text: 'Install', disabled: false, class: 'btn-install', action: () => handleInstall(appId) };

        switch (prog.status) {
            case 'starting': return { text: 'Starting...', disabled: true, class: 'btn-install' };
            case 'downloading': return { text: `Downloading ${prog.percentage}%`, disabled: true, class: 'btn-install' };
            case 'installing': return { text: 'Installing to System...', disabled: true, class: 'btn-install' };
            case 'completed': return { text: 'Uninstall', disabled: false, class: 'btn-uninstall', action: () => handleUninstall(appId) };
            case 'removing': return { text: 'Removing...', disabled: true, class: 'btn-uninstall' };
            case 'error': return { text: 'Failed - Retry', disabled: false, class: 'btn-install', action: () => handleInstall(appId) };
            default: return { text: 'Install', disabled: false, class: 'btn-install', action: () => handleInstall(appId) };
        }
    }

    function handleImageError(e: Event) {
      const target = e.currentTarget as HTMLImageElement;
      target.src = 'https://dl.flathub.org/assets/default/settings.svg';
  }
</script>

<main id="app">
    <aside class="sidebar">
        <input 
            type="text" 
            class="search-box" 
            placeholder="Search apps..." 
            spellcheck="false" 
            bind:value={searchQuery}
            on:input={handleSearch}
        />
        <h3 class="category-header">Categories</h3>
        <ul class="category-list">
            <li 
                class="category-item {activeCategory === 'Discover' ? 'active' : ''}" 
                on:click={loadDiscover}
            >
                Discover (Latest)
            </li>
            {#each categories as cat}
                <li 
                    class="category-item {activeCategory === cat.id ? 'active' : ''}" 
                    on:click={() => loadCategory(cat.id, cat.label)}
                >
                    {cat.label}
                </li>
            {/each}
        </ul>
    </aside>

    <section class="main-content">
        <header class="header">
            <h1>{viewTitle}</h1>
            <span class="catalog-status {isCatalogError ? 'error' : 'success'}">
                {catalogStatus}
            </span>
        </header>

        <div class="app-grid">
            {#if isLoading}
                <div class="message">Loading applications...</div>
            {:else if errorMessage}
                <div class="message error-msg">Failed to load data: {errorMessage}</div>
            {:else if apps.length === 0}
                <div class="message">No applications found.</div>
            {:else}
                {#each apps as app}
                    {@const btnProps = getButtonProps(app.flatpakAppId)}
                    {@const prog = appProgress[app.flatpakAppId]}
                    
                    <article class="app-card">
                        <div class="app-header">
                            <img 
                                class="app-icon" 
                                src={app.iconUrl || 'https://dl.flathub.org/assets/default/settings.svg'} 
                                alt={app.name}
                                on:error={handleImageError}
                            />
                            <div>
                                <h3 class="app-title">{app.name}</h3>
                                <p class="app-dev">{app.developer || 'Flathub'}</p>
                            </div>
                        </div>
                        <p class="app-summary">{app.summary}</p>
                        
                        <div class="action-area">
                            <button 
                                class="btn {btnProps.class}" 
                                disabled={btnProps.disabled}
                                on:click={btnProps.action}
                            >
                                {btnProps.text}
                            </button>
                            
                            {#if prog && (prog.status === 'downloading' || prog.status === 'installing')}
                                <div class="progress-container">
                                    <div class="progress-bar" style="width: {prog.percentage}%"></div>
                                </div>
                            {/if}
                        </div>
                    </article>
                {/each}
            {/if}
        </div>
    </section>
</main>

<style>
    /* Scoped Svelte CSS */
    :global(body) {
        margin: 0;
        font-family: system-ui, -apple-system, sans-serif;
        background-color: #f8f9fa;
        color: #212529;
        user-select: none;
    }

    #app {
        display: flex;
        height: 100vh;
        width: 100vw;
        overflow: hidden;
    }

    /* Sidebar Styles */
    .sidebar {
        width: 250px;
        min-width: 250px;
        background-color: #ffffff;
        border-right: 1px solid #e3e6f0;
        display: flex;
        flex-direction: column;
        padding: 20px;
        box-sizing: border-box;
        overflow-y: auto;
    }

    .search-box {
        width: 100%;
        padding: 10px;
        border-radius: 8px;
        border: 1px solid #e3e6f0;
        margin-bottom: 24px;
        background-color: #f8f9fa;
        box-sizing: border-box;
    }

    .category-header {
        font-size: 0.8rem;
        text-transform: uppercase;
        color: #6c757d;
        margin-bottom: 10px;
    }

    .category-list {
        list-style: none;
        padding: 0;
        margin: 0;
    }

    .category-item {
        padding: 12px;
        cursor: pointer;
        border-radius: 8px;
        margin-bottom: 4px;
        font-weight: 500;
        color: #6c757d;
        transition: all 0.2s;
    }

    .category-item:hover {
        background-color: #f8f9fa;
        color: #212529;
    }

    .category-item.active {
        background-color: #e0edff;
        color: #0061f2;
    }

    /* Main Content Styles */
    .main-content {
        flex: 1;
        overflow-y: auto;
        padding: 32px;
    }

    .header {
        display: flex;
        justify-content: space-between;
        align-items: center;
        margin-bottom: 24px;
    }

    .header h1 {
        margin: 0;
    }

    .catalog-status {
        font-size: 0.9rem;
        color: #6c757d;
    }
    .catalog-status.success { color: #198754; }
    .catalog-status.error { color: #dc3545; }

    .app-grid {
        display: grid;
        grid-template-columns: repeat(auto-fill, minmax(240px, 1fr));
        gap: 24px;
    }

    .message {
        grid-column: 1 / -1;
        text-align: center;
        color: #6c757d;
        padding: 40px;
    }
    .error-msg { color: #dc3545; }

    /* Card Styles */
    .app-card {
        background: white;
        padding: 20px;
        border-radius: 12px;
        border: 1px solid #e3e6f0;
        display: flex;
        flex-direction: column;
        transition: transform 0.2s, box-shadow 0.2s;
    }

    .app-card:hover {
        transform: translateY(-2px);
        box-shadow: 0 4px 12px rgba(0,0,0,0.05);
    }

    .app-header {
        display: flex;
        align-items: center;
        gap: 16px;
        margin-bottom: 12px;
    }

    .app-icon {
        width: 56px;
        height: 56px;
        object-fit: contain;
        border-radius: 12px;
    }

    .app-title {
        margin: 0;
        font-size: 1.1rem;
        font-weight: 600;
    }

    .app-dev {
        margin: 0;
        font-size: 0.8rem;
        color: #6c757d;
    }

    .app-summary {
        font-size: 0.9rem;
        color: #6c757d;
        flex-grow: 1;
        display: -webkit-box;
        -webkit-line-clamp: 3;
        -webkit-box-orient: vertical;
        overflow: hidden;
        margin-bottom: 16px;
    }

    /* Buttons & Progress */
    .action-area {
        margin-top: auto;
    }

    .btn {
        width: 100%;
        padding: 10px;
        border-radius: 8px;
        border: none;
        font-weight: 600;
        cursor: pointer;
        transition: opacity 0.2s;
    }

    .btn-install {
        background-color: #0061f2;
        color: white;
    }

    .btn-uninstall {
        background-color: #dc3545;
        color: white;
    }

    .btn:disabled {
        opacity: 0.6;
        cursor: not-allowed;
    }

    .progress-container {
        height: 6px;
        background-color: #e3e6f0;
        border-radius: 3px;
        margin-top: 12px;
        overflow: hidden;
    }

    .progress-bar {
        height: 100%;
        background-color: #0061f2;
        transition: width 0.1s linear;
    }
</style>
<script lang="ts">
	import './app.css';
	import { onMount } from 'svelte';
	// lucide-svelte icons
	import {
		Sparkles,
		MonitorPlay,
		Code,
		GraduationCap,
		Gamepad2,
		Palette,
		Globe,
		Briefcase,
		FlaskConical,
		Wrench,
		Sun,
		Moon,
		Laptop,
		User,
		HardDriveDownload,
		Download,
		Trash2,
		RefreshCw,
		Loader2
	} from '@lucide/svelte';

	// shadcn-svelte imports
    import * as Popover from "$lib/components/ui/popover";
    import { Progress } from "$lib/components/ui/progress";

	// --- Type Definitions ---
	// (Keeping your existing interfaces: AppSummary, ProgressStatus, AppProgress, ProgressPayload)
	interface AppSummary {
		flatpakAppId: string;
		name: string;
		summary: string;
		iconUrl: string;
		version: string;
		developer: string
	}

	type ProgressStatus = 
		'starting' |
		'downloading' |
		'installing' |
		'completed' |
		'removing' |
		'error';

	interface InstalledApp { appId: string; name: string; version: string; updateAvailable: boolean; }
	interface AppProgress { status: ProgressStatus; percentage: number; name?: string; }
	interface ProgressPayload { appId: string; status: ProgressStatus; percentage: number; }

	interface Category {
		id: string;
		label: string;
		icon: any // Svelte component type
	}

	const wailsApp = (window as any).go.main.App;
	const runtime = (window as any).runtime;

	// Added icons to the category map
	const categories: Category[] = [
		{ id: 'AudioVideo', label: 'Audio & Video', icon: MonitorPlay },
		{ id: 'Development', label: 'Development', icon: Code },
		{ id: 'Education', label: 'Education', icon: GraduationCap },
		{ id: 'Game', label: 'Games', icon: Gamepad2 },
		{ id: 'Graphics', label: 'Graphics', icon: Palette },
		{ id: 'Network', label: 'Internet', icon: Globe },
		{ id: 'Office', label: 'Office', icon: Briefcase },
		{ id: 'Science', label: 'Science', icon: FlaskConical },
		{ id: 'System', label: 'System', icon: Wrench }
	];

	// --- State ---
	let apps: AppSummary[] = [];
	let installedApps: InstalledApp[] = [];
	let viewTitle: string = 'Discover';
	let activeCategory: string | null = 'Discover';
	let searchQuery: string = '';
	let isLoading: boolean = true;
	let errorMessage: string = '';
	let appProgress: Record<string, AppProgress> = {};
	let searchTimeout: ReturnType<typeof setTimeout>;


	// Computed property: count active background tasks
    $: activeTasksCount = Object.values(appProgress).filter(p => 
        p.status === 'downloading' || p.status === 'installing' || p.status === 'removing'
    ).length;

	// --- Theme Management ---
	type Theme = 'light' | 'dark' | 'system';

	let currentTheme: Theme = 'system';

	function applyTheme(theme: Theme) {
		currentTheme = theme;

		const root = document.documentElement;

		if (theme === 'system') {
			const systemPrefersDark = window.matchMedia('(prefers-color-scheme: dark)').matches;

			root.classList.toggle('dark', systemPrefersDark);
		} else {
			root.classList.toggle('dark', theme === 'dark');
		}

		// Optional: Save to localStorage so it persists across reloads
		localStorage.setItem('theme', theme);
	}

	// Initialize after DOM is ready (replaces svelte's onMount to avoid importing onMount)
	if (typeof window !== 'undefined') {
		const init = () => {
			// Load saved theme or default to system
			const savedTheme = localStorage.getItem('theme') as Theme || 'system';

			applyTheme(savedTheme);

			// Listen for OS-level theme changes if set to system
			window.matchMedia('(prefers-color-scheme: dark)').addEventListener('change', () => {
				if (currentTheme === 'system') applyTheme('system');
			});

			// Wails Events
			runtime.EventsOn("flatpak:progress", (payload: ProgressPayload) => {
				appProgress = {
					...appProgress,
					[payload.appId]: { status: payload.status, percentage: payload.percentage }
				};
			});

			loadDiscover();
		};

		if (document.readyState === 'loading') {
			window.addEventListener('DOMContentLoaded', init);
		} else {
			init();
		}
	}

	// --- Data Fetching Methods (unchanged) ---
	async function loadDiscover(): Promise<void> {
		activeCategory = 'Discover';
		viewTitle = 'Discover';
		isLoading = true;
		errorMessage = '';

		try {
			apps = await wailsApp.GetDiscoverApps() || [];
		} catch(err) {
			errorMessage = String(err);
		} finally {
			isLoading = false;
		}
	}

	async function loadCategory(catId: string, catLabel: string): Promise<void> {
		activeCategory = catId;
		viewTitle = catLabel;
		searchQuery = '';
		isLoading = true;
		errorMessage = '';

		try {
			apps = await wailsApp.GetAppsByCategory(catId) || [];
		} catch(err) {
			errorMessage = String(err);
		} finally {
			isLoading = false;
		}
	}

	async function loadInstalled(): Promise<void> {
        activeCategory = 'Installed'; 
        viewTitle = 'Installed Applications'; 
        searchQuery = ''; 
        isLoading = true;
        try { 
            installedApps = await wailsApp.GetInstalledApps() || []; 
        } catch (err) { 
            console.error("Failed to load installed apps:", err);
        } finally { 
            isLoading = false; 
        }
    }

	function handleSearch(): void {
		clearTimeout(searchTimeout);
		activeCategory = null;

		searchTimeout = setTimeout(
			async () => {
				const query = searchQuery.trim();

				if (query.length > 2) {
					viewTitle = `Search: "${query}"`;
					isLoading = true;

					try {
						apps = await wailsApp.SearchApps(query) || [];
					} catch(err) {
						errorMessage = String(err);
					} finally {
						isLoading = false;
					}
				} else if (query.length === 0) {
					loadDiscover();
				}
			},
			400
		);
	}

	// --- System Actions ---
    function handleInstall(appId: string): void {
        appProgress = { ...appProgress, [appId]: { status: 'starting', percentage: 0 } };
        wailsApp.InstallApp(appId, true).catch(console.error);
    }

    function handleUninstall(appId: string): void {
        appProgress = { ...appProgress, [appId]: { status: 'removing', percentage: 0 } };
        wailsApp.UninstallApp(appId, true).catch(console.error);
    }

	// The Update function uses the same endpoint as Install under the hood in flatpak, 
    // but we can bind specifically to the Update command we made in Go.
    function handleUpdate(appId: string): void {
        appProgress = { ...appProgress, [appId]: { status: 'starting', percentage: 0 } };
        // NOTE: Ensure wailsApp.UpdateApp is bound in your app.go just like InstallApp!
        wailsApp.UpdateApp(appId, true).catch(console.error); 
    }

    // Helper to color the shadcn Progress bar based on state
    function getProgressColorClass(status: ProgressStatus): string {
        if (status === 'removing') return '[&>div]:bg-red-500';
        if (status === 'installing') return '[&>div]:bg-green-500';
        return '[&>div]:bg-blue-500'; // downloading or starting
    }
	
	// --- UI Helpers ---
	function handleImageError(e: Event) {
		(e.currentTarget as HTMLImageElement).src = 'https://dl.flathub.org/assets/default/settings.svg';
	}

	onMount(() => {
        runtime.EventsOn("flatpak:progress", (payload: ProgressPayload) => {
            // Keep track of the app name for the popover UI if possible, fallback to ID
            const appName = apps.find(a => a.flatpakAppId === payload.appId)?.name || payload.appId;
            
            appProgress = { 
                ...appProgress, 
                [payload.appId]: { 
                    status: payload.status, 
                    percentage: payload.percentage,
                    name: appName
                } 
            };

            // Clean up completed/error items from the popover after 5 seconds
            if (payload.status === 'completed' || payload.status === 'error') {
                setTimeout(() => {
                    const newProgress = { ...appProgress };
                    delete newProgress[payload.appId];
                    appProgress = newProgress;
                    
                    // If we are looking at the Installed tab, refresh it automatically!
                    if (activeCategory === 'Installed') loadInstalled();
                }, 5000);
            }
        });

        loadDiscover();
    });
</script>

<main class="flex h-screen w-screen overflow-hidden bg-background text-foreground select-none"
>
	<aside
		class="flex flex-col w-64 min-w-[16rem] bg-card border-r border-border p-5 overflow-y-auto"
	>
		<input
			type="text"
			class="w-full px-3 py-2 mb-6 text-sm bg-muted border border-border rounded-lg focus:outline-none focus:ring-2 focus:ring-primary/50"
			placeholder="Search apps..."
			spellcheck="false"
			bind:value={searchQuery}
			on:input={handleSearch}
		/>

		<h3
			class="text-xs font-semibold tracking-wider text-muted-foreground uppercase mb-3 px-2"
		>Categories</h3>

		<ul class="flex-1 space-y-1">
			<li>
				<button
					class="w-full flex items-center gap-3 px-3 py-2.5 text-sm font-medium rounded-xl transition-colors
                           {activeCategory === 'Discover'
						? 'bg-primary/10 text-primary'
						: 'text-muted-foreground hover:bg-muted hover:text-foreground'}"
					on:click={loadDiscover}
				><Sparkles class="w-4 h-4" />Discover</button>
			</li>

			{#each categories as cat}
				<li>
					<button
						class="w-full flex items-center gap-3 px-3 py-2.5 text-sm font-medium rounded-xl transition-colors
                               {activeCategory === cat.id
							? 'bg-primary/10 text-primary'
							: 'text-muted-foreground hover:bg-muted hover:text-foreground'}"
						on:click={() => loadCategory(cat.id, cat.label)}
					>
						<svelte:component this={cat.icon} class="w-4 h-4" />
						{cat.label}
					</button>
				</li>
			{/each}
			<li>
				<button 
					class="w-full flex items-center gap-3 px-3 py-2.5 text-sm font-medium rounded-xl transition-colors
						{activeCategory === 'Installed' ? 'bg-primary/10 text-primary' : 'text-muted-foreground hover:bg-muted hover:text-foreground'}"
					on:click={loadInstalled}
				>
					<HardDriveDownload class="w-4 h-4" />
					Installed Apps
				</button>
			</li>
		</ul>

		<div
			class="pt-4 mt-4 border-t border-border flex justify-between items-center px-1 text-muted-foreground"
		>
			<button
				class="p-2 rounded-xl hover:bg-muted hover:text-foreground {currentTheme === 'light' ? 'bg-muted text-foreground' : ''}"
				on:click={() => applyTheme('light')}
				title="Light Mode"
			><Sun class="w-4 h-4" /></button>

			<button
				class="p-2 rounded-xl hover:bg-muted hover:text-foreground {currentTheme === 'system' ? 'bg-muted text-foreground' : ''}"
				on:click={() => applyTheme('system')}
				title="System Theme"
			><Laptop class="w-4 h-4" /></button>

			<button
				class="p-2 rounded-xl hover:bg-muted hover:text-foreground {currentTheme === 'dark' ? 'bg-muted text-foreground' : ''}"
				on:click={() => applyTheme('dark')}
				title="Dark Mode"
			><Moon class="w-4 h-4" /></button>
		</div>

		        <div class="pt-4 mt-4 border-t border-border flex justify-between items-center">
            
            <Popover.Root>
                <Popover.Trigger
                    class="relative p-2 rounded-xl hover:bg-muted text-muted-foreground hover:text-foreground transition-colors"
                    title="Activity Center"
                >
                    <User class="w-5 h-5" />
                    {#if activeTasksCount > 0}
                        <span class="absolute top-1 right-1 flex h-3 w-3">
                            <span class="animate-ping absolute inline-flex h-full w-full rounded-full bg-blue-400 opacity-75"></span>
                            <span class="relative inline-flex rounded-full h-3 w-3 bg-blue-500"></span>
                        </span>
                    {/if}
                </Popover.Trigger>
                
                <Popover.Content side="right" align="end" class="w-80 p-4">
                    <h4 class="font-medium text-sm mb-4">Background Tasks</h4>
                    {#if Object.keys(appProgress).length === 0}
                        <p class="text-sm text-muted-foreground text-center py-4">No active tasks.</p>
                    {:else}
                        <div class="space-y-4">
                            {#each Object.entries(appProgress) as [id, prog]}
                                <div class="space-y-2">
                                    <div class="flex justify-between text-xs font-medium">
                                        <span class="truncate pr-2">{prog.name}</span>
                                        <span class="capitalize text-muted-foreground">{prog.status} {prog.percentage}%</span>
                                    </div>
                                    <Progress value={prog.percentage} class="h-2 {getProgressColorClass(prog.status)}" />
                                </div>
                            {/each}
                        </div>
                    {/if}
                </Popover.Content>
            </Popover.Root>

        </div>

	</aside>

	{#if activeCategory !== 'Installed'}
	<section class="flex-1 p-8 overflow-y-auto">
		<header class="mb-8"><h1 class="text-3xl font-bold tracking-tight">{viewTitle}</h1></header>

		<div
			class="grid grid-cols-[repeat(auto-fill,minmax(260px,1fr))] gap-6"
		>
			{#if isLoading}
				<div
					class="col-span-full text-center text-muted-foreground py-10"
				>Loading applications...</div>
			{:else if errorMessage}
				<div
					class="col-span-full text-center text-destructive py-10"
				>Failed to load: {errorMessage}</div>
			{:else if apps.length === 0}
				<div
					class="col-span-full text-center text-muted-foreground py-10"
				>No applications found.</div>
			{:else}
				{#each apps as app}
					<article
						class="flex flex-col bg-card border border-border p-5 rounded-2xl shadow-sm hover:shadow-md hover:-translate-y-0.5 transition-all"
					>
						<div class="flex items-center gap-4 mb-4">
							<img
								class="w-14 h-14 object-contain rounded-xl"
								src={app.iconUrl}
								alt={app.name}
								on:error={handleImageError}
							/>

							<div>
								<h3 class="font-semibold text-base leading-tight">{app.name}</h3>
								<p class="text-xs text-muted-foreground mt-1">{app.developer || 'Flathub'}</p>
							</div>
						</div>

						<p
							class="text-sm text-muted-foreground line-clamp-3 mb-6 flex-1"
						>{app.summary}</p>
					</article>
				{/each}
			{/if}
		</div>
	</section>
	{/if}
	{#if activeCategory === 'Installed'}
	<section class="flex-1 p-8 overflow-y-auto">
        <header class="mb-8">
            <h1 class="text-3xl font-bold tracking-tight">{viewTitle}</h1>
        </header>

        <div class="grid grid-cols-[repeat(auto-fill,minmax(300px,1fr))] gap-4">
                {#if isLoading}
                    <p class="col-span-full text-muted-foreground">Scanning system...</p>
                {:else if installedApps.length === 0}
                    <p class="col-span-full text-muted-foreground">No Flatpak applications installed.</p>
                {:else}
                    {#each installedApps as app}
                        {@const isBusy = !!appProgress[app.appId]}
                        <article class="flex items-center justify-between bg-card border border-border p-4 rounded-xl shadow-sm">
                            <div class="overflow-hidden pr-4">
                                <h3 class="font-semibold text-sm truncate">{app.name}</h3>
                                <p class="text-xs text-muted-foreground">Version: {app.version}</p>
                            </div>
                            <div class="flex gap-2 shrink-0">
                                {#if app.updateAvailable}
                                    <button 
                                        class="p-2 bg-blue-100 text-blue-700 hover:bg-blue-200 dark:bg-blue-900/30 dark:text-blue-400 rounded-lg transition-colors disabled:opacity-40 disabled:cursor-not-allowed"
                                        title={isBusy ? 'Updating...' : 'Update Available'}
                                        disabled={isBusy}
                                        on:click={() => handleUpdate(app.appId)}
                                    >
                                        {#if isBusy}
                                            <Loader2 class="w-4 h-4 animate-spin" />
                                        {:else}
                                            <RefreshCw class="w-4 h-4" />
                                        {/if}
                                    </button>
                                {/if}
                                <button 
                                    class="p-2 bg-red-100 text-red-700 hover:bg-red-200 dark:bg-red-900/30 dark:text-red-400 rounded-lg transition-colors disabled:opacity-40 disabled:cursor-not-allowed"
                                    title={isBusy ? 'Removing...' : 'Uninstall'}
                                    disabled={isBusy}
                                    on:click={() => handleUninstall(app.appId)}
                                >
                                    {#if isBusy && appProgress[app.appId]?.status === 'removing'}
                                        <Loader2 class="w-4 h-4 animate-spin" />
                                    {:else}
                                        <Trash2 class="w-4 h-4" />
                                    {/if}
                                </button>
                            </div>
                        </article>
                    {/each}
                {/if}
            </div>
        </section>
	{/if}

</main>

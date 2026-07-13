<script lang="ts">
	import './app.css';
	import { onMount, onDestroy } from 'svelte';
	import { fly, fade } from 'svelte/transition';
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
		PackageCheck,
		CloudBackup,
		SunMoon,
		Loader2,
		Settings,
		Check,
        User2,
		TrendingUp,
		X,
		ExternalLink

	} from '@lucide/svelte';

	// shadcn-svelte imports
    import * as Popover from "$lib/components/ui/popover";
    import { Progress } from "$lib/components/ui/progress";

	// --- Type Definitions ---
	// (Keeping your existing interfaces: AppSummary, ProgressStatus, AppProgress, ProgressPayload)
	interface AppDetails {
		flatpakAppId: string;
		name: string;
		summary: string;
		description: string;
		homepageUrl: string;
		bugtrackerUrl: string;
		iconUrl: string;
		version: string;
		developer: string;
		screenshots: string[];
		releaseDate: string;
	}

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
	let discoverShowAll: boolean = false;

	// Reactive sets for O(1) install-status lookups on the Discover page
	$: installedAppIds = new Set<string>(installedApps.map(a => a.appId));
	$: updateableAppIds = new Set<string>(installedApps.filter(a => a.updateAvailable).map(a => a.appId));

	// --- Popular view state ---
	type PopularTab = 'discover' | 'games' | 'create';
	let popularTab: PopularTab = 'discover';
	let heroApps: AppSummary[] = [];
	let restApps: AppSummary[] = [];
	let heroIndex: number = 0;
	let heroSlideInterval: ReturnType<typeof setInterval> | null = null;

	// Gradient definitions for hero cards (as inline CSS to avoid Tailwind purging).
	const heroGradients: Array<{ from: string; via: string; to: string }> = [
		{ from: '#7c3aed', via: '#9333ea', to: '#3b82f6' },  // violet → blue
		{ from: '#f97316', via: '#ef4444', to: '#ec4899' },  // orange → pink
		{ from: '#10b981', via: '#14b8a6', to: '#06b6d4' },  // emerald → cyan
		{ from: '#3b82f6', via: '#6366f1', to: '#8b5cf6' },  // blue → purple
		{ from: '#f43f5e', via: '#f97316', to: '#f59e0b' },  // rose → amber
	];

	// Placeholder featured promo cards — populated by external API in a future iteration.
	interface FeaturedPromo { badge: string; title: string; subtitle: string; gradient: { from: string; to: string }; }
	const featuredPromos: FeaturedPromo[] = [
		{ badge: 'Featured', title: "Editor's Choice",   subtitle: 'Hand-picked highlights from our team',      gradient: { from: '#5b21b6', to: '#7c3aed' } },
		{ badge: 'New',      title: 'Fresh Arrivals',      subtitle: 'Brand new apps just added to Flathub',     gradient: { from: '#9d174d', to: '#e11d48' } },
		{ badge: 'Staff Pick', title: 'Must-Have Tools',   subtitle: 'Essential apps for your Linux desktop',    gradient: { from: '#075985', to: '#0ea5e9' } },
	];


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

	let updateQueue: string[] = [];
	let isUpdatingAll: boolean = false;
	let currentQueueAppId: string | null = null;
	let mediaQuery: MediaQueryList;
	let themeListener: (e: MediaQueryListEvent) => void;

	$: updateableAppsCount = installedApps.filter(a => a.updateAvailable).length;

	function handleUpdateAll(): void {
		if (isUpdatingAll) return;
		const toUpdate = installedApps
			.filter(a => a.updateAvailable && !appProgress[a.appId])
			.map(a => a.appId);
		if (toUpdate.length === 0) return;
		updateQueue = toUpdate;
		isUpdatingAll = true;
		processNextUpdate();
	}

	function processNextUpdate(): void {
		if (updateQueue.length === 0) {
			isUpdatingAll = false;
			currentQueueAppId = null;
			return;
		}
		currentQueueAppId = updateQueue[0];
		handleUpdate(currentQueueAppId);
	}

	let selectedAppDetails: AppDetails | null = null;
	let isDetailsOpen: boolean = false;
	let isDetailsLoading: boolean = false;
	let zoomedScreenshot: string | null = null;

	async function openDetails(appId: string): Promise<void> {
		isDetailsOpen = true;
		isDetailsLoading = true;
		selectedAppDetails = null;
		zoomedScreenshot = null;
		try {
			selectedAppDetails = await wailsApp.GetAppDetails(appId);
		} catch (err) {
			console.error("Failed to load app details:", err);
			const basicApp = apps.find(a => a.flatpakAppId === appId) || 
			                  restApps.find(a => a.flatpakAppId === appId) ||
			                  heroApps.find(a => a.flatpakAppId === appId);
			
			if (basicApp) {
				selectedAppDetails = {
					flatpakAppId: basicApp.flatpakAppId,
					name: basicApp.name,
					summary: basicApp.summary,
					description: `<p>${basicApp.summary}</p><p class="text-xs text-muted-foreground mt-4">Full description is currently unavailable because the AppStream catalog is still syncing. You can still install or run this app.</p>`,
					homepageUrl: "",
					bugtrackerUrl: "",
					iconUrl: basicApp.iconUrl,
					version: basicApp.version || "Unknown",
					developer: basicApp.developer || "Flathub",
					screenshots: [],
					releaseDate: ""
				};
			}
		} finally {
			isDetailsLoading = false;
		}
	}

	function closeDetails(): void {
		isDetailsOpen = false;
		selectedAppDetails = null;
		zoomedScreenshot = null;
	}


	// --- Data Fetching Methods (unchanged) ---
	async function loadDiscover(): Promise<void> {
		activeCategory = 'Discover';
		viewTitle = 'Discover';
		discoverShowAll = false;
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

	// --- Popular view helpers ---
	function stopHeroSlide(): void {
		if (heroSlideInterval !== null) {
			clearInterval(heroSlideInterval);
			heroSlideInterval = null;
		}
	}

	function startHeroSlide(): void {
		stopHeroSlide();
		if (heroApps.length > 1) {
			heroSlideInterval = setInterval(() => {
				heroIndex = (heroIndex + 1) % heroApps.length;
			}, 4000);
		}
	}

	async function loadPopular(tab: PopularTab = 'discover'): Promise<void> {
		activeCategory = 'Popular';
		viewTitle = 'Popular';
		popularTab = tab;
		heroIndex = 0;
		searchQuery = '';
		isLoading = true;
		errorMessage = '';
		stopHeroSlide();

		try {
			let allApps: AppSummary[] = [];
			if (tab === 'discover') {
				allApps = await wailsApp.GetPopularApps() || [];
			} else if (tab === 'games') {
				allApps = await wailsApp.GetPopularGames() || [];
			} else {
				allApps = await wailsApp.GetPopularCreate() || [];
			}
			heroApps = allApps.slice(0, 5);
			restApps = allApps.slice(5);
		} catch (err) {
			errorMessage = String(err);
		} finally {
			isLoading = false;
		}

		startHeroSlide();
	}

	// Stop the carousel whenever the user navigates away from Popular.
	$: if (activeCategory !== 'Popular') stopHeroSlide();

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
		// Load saved theme or default to system
		const savedTheme = localStorage.getItem('theme') as Theme || 'system';
		applyTheme(savedTheme);

		// Listen for OS-level theme changes if set to system
		mediaQuery = window.matchMedia('(prefers-color-scheme: dark)');
		themeListener = (e: MediaQueryListEvent) => {
			if (currentTheme === 'system') {
				applyTheme('system');
			}
		};
		mediaQuery.addEventListener('change', themeListener);

		// Listen for Wails events
		runtime.EventsOn("flatpak:progress", (payload: ProgressPayload) => {
			// Find name in apps or installedApps
			const appName = apps.find(a => a.flatpakAppId === payload.appId)?.name || 
			                installedApps.find(a => a.appId === payload.appId)?.name || 
			                payload.appId;

			appProgress = { 
				...appProgress, 
				[payload.appId]: { 
					status: payload.status, 
					percentage: payload.percentage,
					name: appName
				} 
			};

			// Handle queue advancement
			if (isUpdatingAll && payload.appId === currentQueueAppId) {
				if (payload.status === 'completed' || payload.status === 'error') {
					updateQueue = updateQueue.filter(id => id !== currentQueueAppId);
					processNextUpdate();
				}
			}

			// Clean up completed/error items from the popover after 5 seconds
			if (payload.status === 'completed' || payload.status === 'error') {
				setTimeout(() => {
					const newProgress = { ...appProgress };
					delete newProgress[payload.appId];
					appProgress = newProgress;

					// Refresh installed apps cache so Discover page buttons update
					wailsApp.GetInstalledApps().then((res: InstalledApp[]) => {
						installedApps = res || [];
					}).catch(console.error);
					
					// If we are looking at the Installed tab, refresh it automatically!
					if (activeCategory === 'Installed') loadInstalled();
				}, 5000);
			}
		});

		// Pre-load installed apps so Discover page can show Get/Update buttons immediately
		wailsApp.GetInstalledApps().then((res: InstalledApp[]) => {
			installedApps = res || [];
		}).catch(console.error);

		loadDiscover();
	});

	onDestroy(() => {
		stopHeroSlide();
		if (mediaQuery && themeListener) {
			mediaQuery.removeEventListener('change', themeListener);
		}
	});
</script>

<main class="flex h-screen w-screen overflow-hidden bg-background text-foreground select-none"
>
	<aside
		class="flex flex-col w-52 min-w-[13rem] bg-card border-r border-border p-3 overflow-y-auto"
	>
		<input
			type="text"
			class="w-full px-3 py-1.5 mb-4 text-sm bg-muted border border-border rounded-lg focus:outline-none focus:ring-2 focus:ring-primary/50"
			placeholder="Search apps..."
			spellcheck="false"
			bind:value={searchQuery}
			on:input={handleSearch}
		/>

		<h3
			class="text-xs font-semibold tracking-wider text-muted-foreground uppercase mb-2 px-2"
		>Categories</h3>

		<ul class="flex-1 space-y-0.5">
			<li>
				<button
					class="w-full flex items-center gap-2 px-2 py-2 text-sm font-medium rounded-lg transition-colors
                           {activeCategory === 'Discover'
						? 'bg-primary/10 text-primary'
						: 'text-muted-foreground hover:bg-muted hover:text-foreground'}"
					on:click={loadDiscover}
				><Sparkles class="w-4 h-4 shrink-0" />Discover</button>
			</li>
			<li>
				<button 
					class="w-full flex items-center gap-2 px-2 py-2 text-sm font-medium rounded-lg transition-colors
						{activeCategory === 'Popular' ? 'bg-primary/10 text-primary' : 'text-muted-foreground hover:bg-muted hover:text-foreground'}"
					on:click={() => loadPopular('discover')}
				>
					<TrendingUp class="w-4 h-4 shrink-0" />
					Popular
				</button>
			</li>
			{#each categories as cat}
				<li>
					<button
						class="w-full flex items-center gap-2 px-2 py-2 text-sm font-medium rounded-lg transition-colors
                               {activeCategory === cat.id
							? 'bg-primary/10 text-primary'
							: 'text-muted-foreground hover:bg-muted hover:text-foreground'}"
						on:click={() => loadCategory(cat.id, cat.label)}
					>
						<svelte:component this={cat.icon} class="w-4 h-4 shrink-0" />
						{cat.label}
					</button>
				</li>
			{/each}
			<li>
				<button 
					class="w-full flex items-center gap-2 px-2 py-2 text-sm font-medium rounded-lg transition-colors
						{activeCategory === 'Installed' ? 'bg-primary/10 text-primary' : 'text-muted-foreground hover:bg-muted hover:text-foreground'}"
					on:click={loadInstalled}
				>
					<HardDriveDownload class="w-4 h-4 shrink-0" />
					Installed Apps
				</button>
			</li>
		</ul>



		<div class="pt-3 mt-3 border-t border-border flex items-center">
            
			<Popover.Root>
                <Popover.Trigger
                    class="relative p-2 rounded-xl hover:bg-muted text-muted-foreground hover:text-foreground transition-colors"
                    title="Theme"
                >
                    <SunMoon class="w-5 h-5" />
                </Popover.Trigger>
                
                <Popover.Content side="right" align="end" class="w-64 p-3">
                    <!-- Theme -->
                    <p class="text-xs font-semibold tracking-wider text-muted-foreground uppercase px-2 mb-2">Theme</p>
                    <div class="flex flex-col gap-0.5">
                        <button
                            class="flex items-center gap-3 px-2 py-2 rounded-lg text-sm w-full transition-colors
                                {currentTheme === 'light' ? 'bg-primary/10 text-primary font-medium' : 'text-muted-foreground hover:bg-muted hover:text-foreground'}"
                            on:click={() => applyTheme('light')}
                        >
                            <Sun class="w-4 h-4 shrink-0" />
                            <span>Light</span>
                            {#if currentTheme === 'light'}<Check class="w-4 h-4 ml-auto" />{/if}
                        </button>
                        <button
                            class="flex items-center gap-3 px-2 py-2 rounded-lg text-sm w-full transition-colors
                                {currentTheme === 'dark' ? 'bg-primary/10 text-primary font-medium' : 'text-muted-foreground hover:bg-muted hover:text-foreground'}"
                            on:click={() => applyTheme('dark')}
                        >
                            <Moon class="w-4 h-4 shrink-0" />
                            <span>Dark</span>
                            {#if currentTheme === 'dark'}<Check class="w-4 h-4 ml-auto" />{/if}
                        </button>
                        <button
                            class="flex items-center gap-3 px-2 py-2 rounded-lg text-sm w-full transition-colors
                                {currentTheme === 'system' ? 'bg-primary/10 text-primary font-medium' : 'text-muted-foreground hover:bg-muted hover:text-foreground'}"
                            on:click={() => applyTheme('system')}
                        >
                            <Laptop class="w-4 h-4 shrink-0" />
                            <span>System</span>
                            {#if currentTheme === 'system'}<Check class="w-4 h-4 ml-auto" />{/if}
                        </button>
                    </div>
                </Popover.Content>
            </Popover.Root>
			<Popover.Root>
                <Popover.Trigger
                    class="relative p-2 rounded-xl hover:bg-muted text-muted-foreground hover:text-foreground transition-colors"
                    title="Backup/ Restore"
                >
                    <CloudBackup class="w-5 h-5" />
                </Popover.Trigger>
                
                <Popover.Content side="right" align="end" class="w-64 p-3">

                    <p class="text-xs font-semibold tracking-wider text-muted-foreground uppercase px-2 mb-2">Backup/ Restore</p>
                    <div class="flex flex-col gap-0.5">
                    </div>

                </Popover.Content>
            </Popover.Root>
            <Popover.Root>
                <Popover.Trigger
                    class="relative p-2 rounded-xl hover:bg-muted text-muted-foreground hover:text-foreground transition-colors"
                    title="Activity Center"
                >
                    <PackageCheck class="w-5 h-5" />
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

	{#if activeCategory !== 'Installed' && activeCategory !== 'Popular'}
	<section class="flex-1 overflow-y-auto">

		{#if activeCategory === 'Discover' && !discoverShowAll}
		<!-- ─────────────────────────────────────────────────── -->
		<!-- DISCOVER DASHBOARD                                   -->
		<!-- ─────────────────────────────────────────────────── -->
		<div class="px-8 pt-8 pb-10">
			<h1 class="text-3xl font-bold tracking-tight mb-6">Discover</h1>

			{#if isLoading}
				<div class="flex items-center justify-center h-48 text-muted-foreground">
					<Loader2 class="w-5 h-5 animate-spin mr-2" />Loading...
				</div>
			{:else}
				<!-- ── Featured promo cards (placeholder; external API in future) ── -->
				<div class="grid grid-cols-3 gap-4 mb-8">
					{#each featuredPromos as promo}
					<div
						class="relative rounded-2xl overflow-hidden h-44 cursor-pointer group"
						style="background: linear-gradient(135deg, {promo.gradient.from}, {promo.gradient.to})"
					>
						<div class="absolute inset-0 p-5 flex flex-col justify-between text-white">
							<span class="text-xs font-bold uppercase tracking-widest" style="opacity:0.7">{promo.badge}</span>
							<div>
								<h3 class="text-xl font-bold leading-tight">{promo.title}</h3>
								<p class="text-sm mt-1" style="opacity:0.75">{promo.subtitle}</p>
							</div>
						</div>
						<div class="absolute inset-0 bg-black/0 group-hover:bg-black/10 transition-colors rounded-2xl"></div>
					</div>
					{/each}
				</div>

				<!-- ── Section header ── -->
				<hr class="border-border mb-5" />
				<div class="flex items-center justify-between mb-2">
					<h2 class="text-lg font-bold tracking-tight">New Apps and Updates</h2>
					<button
						class="text-sm font-medium text-primary hover:underline"
						on:click={() => discoverShowAll = true}
					>See All</button>
				</div>

				<!-- ── 3 × 2 app columns (uniform fixed-height cells) ── -->
				{#if apps.length === 0}
					<p class="text-sm text-muted-foreground py-4">No applications found.</p>
				{:else}
					<div class="overflow-x-auto">
					<div class="grid grid-cols-3 gap-3 min-w-[600px]">
						{#each [0, 1, 2] as colIdx}
						<div class="bg-card border border-border rounded-2xl overflow-hidden">
							{#each apps.slice(colIdx * 2, colIdx * 2 + 2) as app, j}
							{@const isBusy    = !!appProgress[app.flatpakAppId]}
							{@const isInstd   = installedAppIds.has(app.flatpakAppId)}
							{@const hasUpdate = updateableAppIds.has(app.flatpakAppId)}
							<div class="flex flex-col justify-between px-3 py-2.5 h-24 hover:bg-muted/50 transition-colors {j > 0 ? 'border-t border-border' : ''}">
								<div class="flex items-start gap-2.5 cursor-pointer" on:click={() => openDetails(app.flatpakAppId)}>
									<img
										class="w-10 h-10 rounded-xl object-contain shrink-0"
										src={app.iconUrl}
										alt={app.name}
										on:error={handleImageError}
									/>
									<div class="flex-1 overflow-hidden">
										<p class="text-xs font-semibold leading-tight line-clamp-1 text-zinc-800 dark:text-zinc-200">{app.name}</p>
										<p class="text-[11px] leading-snug mt-0.5 line-clamp-2 text-zinc-500 dark:text-zinc-400">{app.summary}</p>
									</div>
								</div>
								<div class="flex justify-end">
									{#if isBusy}
										<div class="flex items-center gap-1">
											<Loader2 class="w-3 h-3 animate-spin text-primary" />
											<span class="text-[10px] text-muted-foreground">{appProgress[app.flatpakAppId]?.percentage}%</span>
										</div>
									{:else if hasUpdate}
										<button
											class="px-2.5 py-1 rounded-full text-[11px] font-semibold bg-blue-100 text-blue-700 hover:bg-blue-200 dark:bg-blue-900/30 dark:text-blue-400 transition-colors"
											on:click={() => handleUpdate(app.flatpakAppId)}
										>Update</button>
									{:else if isInstd}
										<Check class="w-4 h-4 text-green-500" />
									{:else}
										<button
											class="px-2.5 py-1 rounded-full text-[11px] font-semibold bg-primary/10 text-primary hover:bg-primary/20 transition-colors"
											on:click={() => handleInstall(app.flatpakAppId)}
										>Get</button>
									{/if}
								</div>
							</div>
							{/each}
						</div>
						{/each}
					</div>
					</div>
				{/if}
			{/if}
		</div>

		{:else}
		<!-- ─────────────────────────────────────────────────── -->
		<!-- STANDARD GRID  (categories / search / see-all)      -->
		<!-- ─────────────────────────────────────────────────── -->
		<div class="p-8">
			<header class="mb-8 flex items-center gap-3">
				{#if activeCategory === 'Discover' && discoverShowAll}
				<button
					class="text-sm font-medium text-primary hover:underline shrink-0"
					on:click={() => discoverShowAll = false}
				>← Discover</button>
				<span class="text-muted-foreground">/</span>
				{/if}
				<h1 class="text-3xl font-bold tracking-tight">
					{discoverShowAll ? 'New Apps and Updates' : viewTitle}
				</h1>
			</header>

			<div class="grid grid-cols-[repeat(auto-fill,minmax(260px,1fr))] gap-6">
				{#if isLoading}
					<div class="col-span-full text-center text-muted-foreground py-10">Loading applications...</div>
				{:else if errorMessage}
					<div class="col-span-full text-center text-destructive py-10">Failed to load: {errorMessage}</div>
				{:else if apps.length === 0}
					<div class="col-span-full text-center text-muted-foreground py-10">No applications found.</div>
				{:else}
					{#each apps as app}
					{@const isBusy    = !!appProgress[app.flatpakAppId]}
					{@const isInstd   = installedAppIds.has(app.flatpakAppId)}
					{@const hasUpdate = updateableAppIds.has(app.flatpakAppId)}
					{@const isQueued  = isUpdatingAll && updateQueue.includes(app.flatpakAppId) && app.flatpakAppId !== currentQueueAppId}
					<article class="flex flex-col h-52 bg-card border border-border p-5 rounded-2xl shadow-sm hover:shadow-md hover:-translate-y-0.5 transition-all">
						<div class="flex items-center gap-4 mb-3 shrink-0 cursor-pointer" on:click={() => openDetails(app.flatpakAppId)}>
							<img
								class="w-14 h-14 object-contain rounded-xl shrink-0"
								src={app.iconUrl}
								alt={app.name}
								on:error={handleImageError}
							/>
							<div class="min-w-0">
								<h3 class="font-semibold text-base leading-tight truncate">{app.name}</h3>
								<p class="text-xs text-muted-foreground mt-1 truncate">{app.developer || 'Flathub'}</p>
							</div>
						</div>
						<p class="text-sm text-muted-foreground line-clamp-3 flex-1 leading-snug mb-3 cursor-pointer" on:click={() => openDetails(app.flatpakAppId)}>{app.summary}</p>
						<div class="flex justify-end mt-auto pt-2 shrink-0">
							{#if isQueued}
								<div class="flex items-center gap-1.5">
									<Loader2 class="w-3.5 h-3.5 animate-spin text-muted-foreground" />
									<span class="text-xs text-muted-foreground">Queued</span>
								</div>
							{:else if isBusy}
								<div class="flex items-center gap-1.5">
									<Loader2 class="w-3.5 h-3.5 animate-spin text-primary" />
									<span class="text-xs text-muted-foreground">{appProgress[app.flatpakAppId]?.percentage}%</span>
								</div>
							{:else if hasUpdate}
								<button
									class="px-3 py-1.5 rounded-full text-xs font-semibold bg-blue-100 text-blue-700 hover:bg-blue-200 dark:bg-blue-900/30 dark:text-blue-400 transition-colors"
									on:click={() => handleUpdate(app.flatpakAppId)}
								>Update</button>
							{:else if isInstd}
								<Check class="w-4 h-4 text-green-500" />
							{:else}
								<button
									class="px-3 py-1.5 rounded-full text-xs font-semibold bg-primary/10 text-primary hover:bg-primary/20 transition-colors"
									on:click={() => handleInstall(app.flatpakAppId)}
								>Get</button>
							{/if}
						</div>
					</article>
					{/each}
				{/if}
			</div>
		</div>
		{/if}

	</section>
	{/if}
	{#if activeCategory === 'Popular'}
	<section class="flex-1 overflow-y-auto">
		<!-- Header + Tab Button Group -->
		<div class="px-8 pt-8 pb-6">
			<div class="inline-flex rounded-xl border border-border bg-muted p-1 gap-1">
				<button
					class="flex items-center gap-1.5 px-4 py-1.5 text-sm font-medium rounded-lg transition-colors
						{popularTab === 'discover' ? 'bg-background shadow-sm text-foreground' : 'text-muted-foreground hover:text-foreground'}"
					on:click={() => loadPopular('discover')}
				>
					<Sparkles class="w-3.5 h-3.5" />Discover
				</button>
				<button
					class="flex items-center gap-1.5 px-4 py-1.5 text-sm font-medium rounded-lg transition-colors
						{popularTab === 'games' ? 'bg-background shadow-sm text-foreground' : 'text-muted-foreground hover:text-foreground'}"
					on:click={() => loadPopular('games')}
				>
					<Gamepad2 class="w-3.5 h-3.5" />Games
				</button>
				<button
					class="flex items-center gap-1.5 px-4 py-1.5 text-sm font-medium rounded-lg transition-colors
						{popularTab === 'create' ? 'bg-background shadow-sm text-foreground' : 'text-muted-foreground hover:text-foreground'}"
					on:click={() => loadPopular('create')}
				>
					<Palette class="w-3.5 h-3.5" />Create
				</button>
			</div>
		</div>

		{#if isLoading}
			<div class="flex items-center justify-center py-20 text-muted-foreground">
				<Loader2 class="w-5 h-5 animate-spin mr-2" />Loading...
			</div>
		{:else if errorMessage}
			<div class="text-center text-destructive py-20">Failed to load: {errorMessage}</div>
		{:else}
			<!-- Hero Carousel -->
			{#if heroApps.length > 0}
			<div class="relative mx-8 mb-8 rounded-3xl overflow-hidden" style="height: 280px;">
				{#each heroApps as app, i}
				{@const g = heroGradients[i % heroGradients.length]}
				<div
					class="absolute inset-0 flex items-center gap-8 p-8 transition-opacity duration-500
						{i === heroIndex ? 'opacity-100 z-10' : 'opacity-0 z-0'}"
					style="background: linear-gradient(135deg, {g.from}, {g.via}, {g.to})"
				>
					<img
						class="w-28 h-28 rounded-2xl object-contain shrink-0 shadow-2xl cursor-pointer"
						src={app.iconUrl}
						alt={app.name}
						on:error={handleImageError}
						on:click={() => openDetails(app.flatpakAppId)}
					/>
					<div class="flex-1 text-white min-w-0 cursor-pointer" on:click={() => openDetails(app.flatpakAppId)}>
						<p class="text-xs font-bold uppercase tracking-widest mb-2" style="opacity:0.6">
							#{i + 1} · Most Popular
						</p>
						<h2 class="text-3xl font-bold leading-tight truncate mb-1">{app.name}</h2>
						<p class="text-sm mb-3" style="opacity:0.75">{app.developer || 'Flathub'}</p>
						<p class="text-sm line-clamp-2 max-w-xl" style="opacity:0.65">{app.summary}</p>
					</div>
				</div>
				{/each}

				<!-- Pill dot navigation -->
				<div class="absolute bottom-5 left-0 right-0 flex justify-center gap-2 z-20">
					{#each heroApps as _, i}
					<button
						class="rounded-full bg-white transition-all duration-200
							{i === heroIndex ? 'w-6 h-2.5 opacity-100' : 'w-2.5 h-2.5 opacity-40 hover:opacity-70'}"
						on:click={() => { heroIndex = i; startHeroSlide(); }}
						aria-label="Slide {i + 1}"
					></button>
					{/each}
				</div>
			</div>
			{/if}

			<!-- Rest of apps grid -->
			{#if restApps.length > 0}
			<div class="px-8 pb-8">
				<h2 class="text-xs font-semibold uppercase tracking-wider text-muted-foreground mb-4">More Apps</h2>
				<div class="grid grid-cols-[repeat(auto-fill,minmax(260px,1fr))] gap-6">
					{#each restApps as app}
					{@const isBusy    = !!appProgress[app.flatpakAppId]}
					{@const isInstd   = installedAppIds.has(app.flatpakAppId)}
					{@const hasUpdate = updateableAppIds.has(app.flatpakAppId)}
					{@const isQueued  = isUpdatingAll && updateQueue.includes(app.flatpakAppId) && app.flatpakAppId !== currentQueueAppId}
					<article class="flex flex-col h-52 bg-card border border-border p-5 rounded-2xl shadow-sm hover:shadow-md hover:-translate-y-0.5 transition-all">
						<div class="flex items-center gap-4 mb-3 shrink-0 cursor-pointer" on:click={() => openDetails(app.flatpakAppId)}>
							<img
								class="w-14 h-14 object-contain rounded-xl shrink-0"
								src={app.iconUrl}
								alt={app.name}
								on:error={handleImageError}
							/>
							<div class="min-w-0">
								<h3 class="font-semibold text-base leading-tight truncate">{app.name}</h3>
								<p class="text-xs text-muted-foreground mt-1 truncate">{app.developer || 'Flathub'}</p>
							</div>
						</div>
						<p class="text-sm text-muted-foreground line-clamp-3 flex-1 leading-snug mb-3 cursor-pointer" on:click={() => openDetails(app.flatpakAppId)}>{app.summary}</p>
						<div class="flex justify-end mt-auto pt-2 shrink-0">
							{#if isQueued}
								<div class="flex items-center gap-1.5">
									<Loader2 class="w-3.5 h-3.5 animate-spin text-muted-foreground" />
									<span class="text-xs text-muted-foreground">Queued</span>
								</div>
							{:else if isBusy}
								<div class="flex items-center gap-1.5">
									<Loader2 class="w-3.5 h-3.5 animate-spin text-primary" />
									<span class="text-xs text-muted-foreground">{appProgress[app.flatpakAppId]?.percentage}%</span>
								</div>
							{:else if hasUpdate}
								<button
									class="px-3 py-1.5 rounded-full text-xs font-semibold bg-blue-100 text-blue-700 hover:bg-blue-200 dark:bg-blue-900/30 dark:text-blue-400 transition-colors"
									on:click={() => handleUpdate(app.flatpakAppId)}
								>Update</button>
							{:else if isInstd}
								<Check class="w-4 h-4 text-green-500" />
							{:else}
								<button
									class="px-3 py-1.5 rounded-full text-xs font-semibold bg-primary/10 text-primary hover:bg-primary/20 transition-colors"
									on:click={() => handleInstall(app.flatpakAppId)}
								>Get</button>
							{/if}
						</div>
					</article>
					{/each}
				</div>
			</div>
			{/if}
		{/if}
	</section>
	{/if}
	{#if activeCategory === 'Installed'}
		<section class="flex-1 p-8 overflow-y-auto">
        <header class="mb-8 flex items-center justify-between">
            <h1 class="text-3xl font-bold tracking-tight">{viewTitle}</h1>
			{#if updateableAppsCount > 0}
				<button 
					class="flex items-center gap-2 px-4 py-2 text-sm font-semibold bg-blue-600 text-white hover:bg-blue-700 dark:bg-blue-600 dark:hover:bg-blue-700 rounded-xl shadow-sm transition-colors disabled:opacity-50 disabled:cursor-not-allowed animate-fade-in"
					disabled={isUpdatingAll || activeTasksCount > 0}
					on:click={handleUpdateAll}
				>
					{#if isUpdatingAll}
						<Loader2 class="w-4 h-4 animate-spin" />
						<span>Updating All ({updateQueue.length} remaining)</span>
					{:else}
						<RefreshCw class="w-4 h-4" />
						<span>Update All ({updateableAppsCount})</span>
					{/if}
				</button>
			{/if}
        </header>

        <div class="grid grid-cols-[repeat(auto-fill,minmax(300px,1fr))] gap-4">
                {#if isLoading}
                    <p class="col-span-full text-muted-foreground">Scanning system...</p>
                {:else if installedApps.length === 0}
                    <p class="col-span-full text-muted-foreground">No Flatpak applications installed.</p>
                {:else}
                    {#each installedApps as app}
                        {@const isBusy = !!appProgress[app.appId]}
                        {@const isQueued = isUpdatingAll && updateQueue.includes(app.appId) && app.appId !== currentQueueAppId}
                        <article class="flex items-center justify-between bg-card border border-border p-4 rounded-xl shadow-sm">
                            <div class="overflow-hidden pr-4 flex-1 cursor-pointer" on:click={() => openDetails(app.appId)}>
                                <h3 class="font-semibold text-sm truncate">{app.name}</h3>
                                <p class="text-xs text-muted-foreground">
									{#if isQueued}
										<span class="text-blue-500 font-medium animate-pulse">Queued for update...</span>
									{:else if isBusy}
										<span class="text-primary font-medium">{appProgress[app.appId]?.status}... {appProgress[app.appId]?.percentage}%</span>
									{:else}
										Version: {app.version}
									{/if}
								</p>
                            </div>
                            <div class="flex gap-2 shrink-0">
                                {#if app.updateAvailable}
                                    <button 
                                        class="p-2 bg-blue-100 text-blue-700 hover:bg-blue-200 dark:bg-blue-900/30 dark:text-blue-400 rounded-lg transition-colors disabled:opacity-40 disabled:cursor-not-allowed"
                                        title={isQueued ? 'Queued...' : isBusy ? 'Updating...' : 'Update Available'}
                                        disabled={isBusy || isQueued}
                                        on:click={() => handleUpdate(app.appId)}
                                    >
                                        {#if isBusy || isQueued}
                                            <Loader2 class="w-4 h-4 animate-spin" />
                                        {:else}
                                            <RefreshCw class="w-4 h-4" />
                                        {/if}
                                    </button>
                                {/if}
                                <button 
                                    class="p-2 bg-red-100 text-red-700 hover:bg-red-200 dark:bg-red-900/30 dark:text-red-400 rounded-lg transition-colors disabled:opacity-40 disabled:cursor-not-allowed"
                                    title={isBusy ? 'Removing...' : 'Uninstall'}
                                    disabled={isBusy || isQueued}
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

	{#if isDetailsOpen}
		<!-- Backdrop overlay -->
		<div 
			class="fixed inset-0 z-50 bg-black/40 dark:bg-black/60 backdrop-blur-sm"
			on:click={closeDetails}
			transition:fade={{ duration: 200 }}
		></div>

		<!-- Side-sheet container -->
		<div 
			class="fixed top-0 right-0 z-50 h-screen w-[460px] max-w-full bg-background border-l border-border shadow-2xl flex flex-col"
			transition:fly={{ x: 460, duration: 300 }}
		>
			<!-- Header -->
			<div class="flex items-center justify-between p-4 border-b border-border shrink-0">
				<button 
					class="p-1.5 hover:bg-muted rounded-lg transition-colors text-muted-foreground hover:text-foreground"
					on:click={closeDetails}
					title="Close"
				>
					<X class="w-5 h-5" />
				</button>
				<span class="text-xs font-semibold text-muted-foreground uppercase tracking-widest">App Details</span>
				<div class="w-8 h-8"></div> <!-- balanced spacer -->
			</div>

			<!-- Scrollable Details -->
			<div class="flex-1 overflow-y-auto p-6 space-y-6">
				{#if isDetailsLoading}
					<div class="flex flex-col items-center justify-center h-64 text-muted-foreground gap-3">
						<Loader2 class="w-8 h-8 animate-spin text-primary" />
						<p class="text-sm">Loading details...</p>
					</div>
				{:else if selectedAppDetails}
					{@const app = selectedAppDetails}
					{@const isBusy = !!appProgress[app.flatpakAppId]}
					{@const isInstd = installedAppIds.has(app.flatpakAppId)}
					{@const hasUpdate = updateableAppIds.has(app.flatpakAppId)}
					{@const isQueued = isUpdatingAll && updateQueue.includes(app.flatpakAppId) && app.flatpakAppId !== currentQueueAppId}

					<!-- Header Block -->
					<div class="flex items-start gap-4">
						<img 
							class="w-20 h-20 rounded-2xl object-contain bg-card border border-border p-2 shrink-0"
							src={app.iconUrl} 
							alt={app.name} 
							on:error={handleImageError}
						/>
						<div class="min-w-0 flex-1">
							<h2 class="text-xl font-bold tracking-tight leading-tight truncate text-foreground">{app.name}</h2>
							<p class="text-xs text-primary font-semibold mt-0.5 truncate">{app.developer || 'Flathub'}</p>
							<p class="text-xs text-muted-foreground mt-1 line-clamp-2 leading-relaxed">{app.summary}</p>
						</div>
					</div>

					<!-- Actions Bar -->
					<div class="flex gap-2 shrink-0">
						{#if isQueued}
							<button 
								class="flex-1 flex items-center justify-center gap-2 py-2.5 px-4 rounded-xl text-sm font-semibold bg-muted text-muted-foreground cursor-not-allowed border border-border"
								disabled
							>
								<Loader2 class="w-4 h-4 animate-spin" />
								<span>Queued for Update</span>
							</button>
						{:else if isBusy}
							<div class="flex-1 flex flex-col gap-1.5 bg-muted/50 p-2.5 rounded-xl border border-border">
								<div class="flex justify-between text-xs font-semibold px-1">
									<span class="capitalize text-primary">{appProgress[app.flatpakAppId]?.status}</span>
									<span>{appProgress[app.flatpakAppId]?.percentage}%</span>
								</div>
								<Progress value={appProgress[app.flatpakAppId]?.percentage} class="h-2 {getProgressColorClass(appProgress[app.flatpakAppId]?.status)}" />
							</div>
						{:else if hasUpdate}
							<button 
								class="flex-1 py-2.5 px-4 rounded-xl text-sm font-semibold bg-blue-600 hover:bg-blue-700 text-white shadow-sm transition-colors"
								on:click={() => handleUpdate(app.flatpakAppId)}
							>
								Update to {app.version}
							</button>
						{:else if isInstd}
							<div class="flex gap-2 w-full">
								<div 
									class="flex-1 py-2.5 px-4 rounded-xl text-sm font-semibold bg-green-500/10 text-green-600 dark:text-green-400 border border-green-500/20 flex items-center justify-center gap-1.5"
								>
									<Check class="w-4 h-4" />
									<span>Installed</span>
								</div>
								<button 
									class="px-4 py-2.5 bg-red-100 hover:bg-red-200 text-red-700 dark:bg-red-900/30 dark:hover:bg-red-900/50 dark:text-red-400 rounded-xl transition-colors text-sm font-semibold"
									on:click={() => handleUninstall(app.flatpakAppId)}
								>
									Uninstall
								</button>
							</div>
						{:else}
							<button 
								class="flex-1 py-2.5 px-4 rounded-xl text-sm font-semibold bg-primary hover:bg-primary/90 text-primary-foreground shadow-sm transition-colors"
								on:click={() => handleInstall(app.flatpakAppId)}
							>
								Install
							</button>
						{/if}
					</div>

					<!-- Screenshot Carousel -->
					{#if app.screenshots && app.screenshots.length > 0}
						<div class="space-y-2">
							<h3 class="text-xs font-bold uppercase tracking-wider text-muted-foreground">Screenshots</h3>
							<div class="flex gap-3 overflow-x-auto pb-2 scrollbar-thin scrollbar-thumb-rounded">
								{#each app.screenshots as src}
									<img 
										class="h-44 rounded-xl border border-border object-cover cursor-zoom-in hover:brightness-95 transition-all shadow-sm shrink-0"
										src={src} 
										alt="Screenshot of {app.name}"
										on:click={() => zoomedScreenshot = src}
									/>
								{/each}
							</div>
						</div>
					{/if}

					<!-- HTML Description -->
					<div class="space-y-2">
						<h3 class="text-xs font-bold uppercase tracking-wider text-muted-foreground">About</h3>
						<div class="text-sm leading-relaxed text-muted-foreground space-y-3 prose dark:prose-invert max-w-none">
							{@html app.description}
						</div>
					</div>

					<!-- Metadata -->
					<div class="border-t border-border pt-5 space-y-4">
						<h3 class="text-xs font-bold uppercase tracking-wider text-muted-foreground">Information</h3>
						<div class="grid grid-cols-2 gap-y-4 gap-x-2 text-xs">
							<div>
								<p class="text-muted-foreground font-medium mb-0.5">Developer</p>
								<p class="font-semibold truncate text-foreground">{app.developer || 'Flathub'}</p>
							</div>
							<div>
								<p class="text-muted-foreground font-medium mb-0.5">Version</p>
								<p class="font-semibold truncate text-foreground">{app.version || 'Unknown'}</p>
							</div>
							{#if app.releaseDate}
								<div>
									<p class="text-muted-foreground font-medium mb-0.5">Released On</p>
									<p class="font-semibold truncate text-foreground">{app.releaseDate}</p>
								</div>
							{/if}
							<div>
								<p class="text-muted-foreground font-medium mb-0.5">Flatpak ID</p>
								<p class="font-semibold truncate text-foreground select-all" title="Application ID">{app.flatpakAppId}</p>
							</div>
							{#if app.homepageUrl}
								<div class="col-span-2">
									<p class="text-muted-foreground font-medium mb-0.5">Project Links</p>
									<div class="flex flex-wrap gap-4 mt-1">
										<a 
											href={app.homepageUrl} 
											target="_blank" 
											class="flex items-center gap-1 text-primary hover:underline font-semibold"
										>
											<span>Homepage</span>
											<ExternalLink class="w-3.5 h-3.5" />
										</a>
										{#if app.bugtrackerUrl}
											<a 
												href={app.bugtrackerUrl} 
												target="_blank" 
												class="flex items-center gap-1 text-primary hover:underline font-semibold"
											>
												<span>Bug Tracker</span>
												<ExternalLink class="w-3.5 h-3.5" />
											</a>
										{/if}
									</div>
								</div>
							{/if}
						</div>
					</div>
				{/if}
			</div>
		</div>
	{/if}

	{#if zoomedScreenshot}
		<div 
			class="fixed inset-0 z-50 bg-black/90 flex items-center justify-center p-8 cursor-zoom-out"
			on:click={() => zoomedScreenshot = null}
			transition:fade={{ duration: 150 }}
		>
			<img 
				class="max-w-full max-h-full rounded-xl object-contain shadow-2xl border border-white/10"
				src={zoomedScreenshot} 
				alt="Zoomed Screenshot"
			/>
		</div>
	{/if}
</main>

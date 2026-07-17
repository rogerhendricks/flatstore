import { writable, derived, get } from 'svelte/store';
import type {
	AppDetails,
	AppSummary,
	InstalledApp,
	AppProgress,
	ProgressPayload,
	PopularTab,
	Theme,
	CatalogStatus
} from '$lib/types';

const wailsApp = (window as any).go.main.App;
const runtime = (window as any).runtime;

// --- Core view state ---
export const apps = writable<AppSummary[]>([]);
export const installedApps = writable<InstalledApp[]>([]);
export const viewTitle = writable<string>('Discover');
export const activeCategory = writable<string | null>('Discover');
export const searchQuery = writable<string>('');
export const isLoading = writable<boolean>(true);
export const errorMessage = writable<string>('');
export const appProgress = writable<Record<string, AppProgress>>({});
export const discoverShowAll = writable<boolean>(false);

// --- AppStream catalog sync state ---
// 'syncing' until the Go backend finishes downloading/parsing appstream.xml.gz
// (or we discover it already finished, via IsCatalogReady()). Details fetched
// via GetAppDetails() rely on this catalog, so the UI can use this status to
// show a subtle "still syncing" hint.
export const catalogStatus = writable<CatalogStatus>('syncing');
export const catalogError = writable<string>('');

let searchTimeout: ReturnType<typeof setTimeout>;

// Reactive sets for O(1) install-status lookups
export const installedAppIds = derived(installedApps, ($installedApps) =>
	new Set<string>($installedApps.map((a) => a.appId))
);
export const updateableAppIds = derived(installedApps, ($installedApps) =>
	new Set<string>($installedApps.filter((a) => a.updateAvailable).map((a) => a.appId))
);
export const updateableAppsCount = derived(
	installedApps,
	($installedApps) => $installedApps.filter((a) => a.updateAvailable).length
);
export const activeTasksCount = derived(appProgress, ($appProgress) =>
	Object.values($appProgress).filter(
		(p) => p.status === 'downloading' || p.status === 'installing' || p.status === 'removing'
	).length
);

// --- Popular view state ---
export const popularTab = writable<PopularTab>('discover');
export const heroApps = writable<AppSummary[]>([]);
export const restApps = writable<AppSummary[]>([]);
export const heroIndex = writable<number>(0);
let heroSlideInterval: ReturnType<typeof setInterval> | null = null;

export function stopHeroSlide(): void {
	if (heroSlideInterval !== null) {
		clearInterval(heroSlideInterval);
		heroSlideInterval = null;
	}
}

export function startHeroSlide(): void {
	stopHeroSlide();
	const currentHeroApps = get(heroApps);
	if (currentHeroApps.length > 1) {
		heroSlideInterval = setInterval(() => {
			heroIndex.update((i) => (i + 1) % currentHeroApps.length);
		}, 4000);
	}
}

// Stop the carousel whenever the user navigates away from Popular.
activeCategory.subscribe((cat) => {
	if (cat !== 'Popular') stopHeroSlide();
});

// --- Update queue state ---
export const updateQueue = writable<string[]>([]);
export const isUpdatingAll = writable<boolean>(false);
export const currentQueueAppId = writable<string | null>(null);

export function handleUpdateAll(): void {
	if (get(isUpdatingAll)) return;
	const toUpdate = get(installedApps)
		.filter((a) => a.updateAvailable && !get(appProgress)[a.appId])
		.map((a) => a.appId);
	if (toUpdate.length === 0) return;
	updateQueue.set(toUpdate);
	isUpdatingAll.set(true);
	processNextUpdate();
}

function processNextUpdate(): void {
	const queue = get(updateQueue);
	if (queue.length === 0) {
		isUpdatingAll.set(false);
		currentQueueAppId.set(null);
		return;
	}
	const nextId = queue[0];
	currentQueueAppId.set(nextId);
	handleUpdate(nextId);
}

// --- Details page state ---
export const selectedAppDetails = writable<AppDetails | null>(null);
export const selectedAppIdForPage = writable<string | null>(null);
export const isDetailsLoading = writable<boolean>(false);
export const zoomedScreenshot = writable<string | null>(null);

export async function openDetails(appId: string): Promise<void> {
	selectedAppIdForPage.set(appId);
	isDetailsLoading.set(true);
	selectedAppDetails.set(null);
	zoomedScreenshot.set(null);
	try {
		selectedAppDetails.set(await wailsApp.GetAppDetails(appId));
	} catch (err) {
		console.error('Failed to load app details:', err);
		const basicApp =
			get(apps).find((a) => a.flatpakAppId === appId) ||
			get(restApps).find((a) => a.flatpakAppId === appId) ||
			get(heroApps).find((a) => a.flatpakAppId === appId);

		if (basicApp) {
			selectedAppDetails.set({
				flatpakAppId: basicApp.flatpakAppId,
				name: basicApp.name,
				summary: basicApp.summary,
				description: `<p>${basicApp.summary}</p><p class="text-xs text-muted-foreground mt-4">Full description is currently unavailable because the AppStream catalog is still syncing. You can still install or run this app.</p>`,
				homepageUrl: '',
				bugtrackerUrl: '',
				helpUrl: '',
				vcsBrowserUrl: '',
				iconUrl: basicApp.iconUrl,
				version: basicApp.version || 'Unknown',
				developer: basicApp.developer || 'Flathub',
				verified: false,
				screenshots: [],
				releaseDate: '',
				ageRating: 'Everyone',
				license: 'Unknown',
				releases: []
			});
		}
	} finally {
		isDetailsLoading.set(false);
	}
}

export function closeDetails(): void {
	selectedAppIdForPage.set(null);
	selectedAppDetails.set(null);
	zoomedScreenshot.set(null);
}

export function handleOpenApp(appId: string): void {
	wailsApp.OpenApp(appId).catch(console.error);
}

// --- Data Fetching Methods ---
export async function loadDiscover(): Promise<void> {
	selectedAppIdForPage.set(null);
	activeCategory.set('Discover');
	viewTitle.set('Discover');
	discoverShowAll.set(false);
	isLoading.set(true);
	errorMessage.set('');

	try {
		apps.set((await wailsApp.GetDiscoverApps()) || []);
	} catch (err) {
		errorMessage.set(String(err));
	} finally {
		isLoading.set(false);
	}
}

export async function loadCategory(catId: string, catLabel: string): Promise<void> {
	selectedAppIdForPage.set(null);
	activeCategory.set(catId);
	viewTitle.set(catLabel);
	searchQuery.set('');
	isLoading.set(true);
	errorMessage.set('');

	try {
		apps.set((await wailsApp.GetAppsByCategory(catId)) || []);
	} catch (err) {
		errorMessage.set(String(err));
	} finally {
		isLoading.set(false);
	}
}

// Shows all apps published by a given developer, using the standard grid
// view (same layout as search results). Triggered by clicking the developer
// name link on the app details page.
export async function loadByDeveloper(developer: string): Promise<void> {
	selectedAppIdForPage.set(null);
	activeCategory.set(null);
	viewTitle.set(`Apps by ${developer}`);
	searchQuery.set('');
	isLoading.set(true);
	errorMessage.set('');

	try {
		apps.set((await wailsApp.GetAppsByDeveloper(developer)) || []);
	} catch (err) {
		errorMessage.set(String(err));
	} finally {
		isLoading.set(false);
	}
}

export async function loadInstalled(): Promise<void> {
	selectedAppIdForPage.set(null);
	activeCategory.set('Installed');
	viewTitle.set('Installed Applications');
	searchQuery.set('');
	isLoading.set(true);
	try {
		installedApps.set((await wailsApp.GetInstalledApps()) || []);
	} catch (err) {
		console.error('Failed to load installed apps:', err);
	} finally {
		isLoading.set(false);
	}
}

export async function loadPopular(tab: PopularTab = 'discover'): Promise<void> {
	selectedAppIdForPage.set(null);
	activeCategory.set('Popular');
	viewTitle.set('Popular');
	popularTab.set(tab);
	heroIndex.set(0);
	searchQuery.set('');
	isLoading.set(true);
	errorMessage.set('');
	stopHeroSlide();

	try {
		let allApps: AppSummary[] = [];
		if (tab === 'discover') {
			allApps = (await wailsApp.GetPopularApps()) || [];
		} else if (tab === 'games') {
			allApps = (await wailsApp.GetPopularGames()) || [];
		} else {
			allApps = (await wailsApp.GetPopularCreate()) || [];
		}
		heroApps.set(allApps.slice(0, 5));
		restApps.set(allApps.slice(5));
	} catch (err) {
		errorMessage.set(String(err));
	} finally {
		isLoading.set(false);
	}

	startHeroSlide();
}

export function handleSearch(): void {
	clearTimeout(searchTimeout);
	activeCategory.set(null);
	selectedAppIdForPage.set(null);

	searchTimeout = setTimeout(async () => {
		const query = get(searchQuery).trim();

		if (query.length > 2) {
			viewTitle.set(`Search: "${query}"`);
			isLoading.set(true);

			try {
				apps.set((await wailsApp.SearchApps(query)) || []);
			} catch (err) {
				errorMessage.set(String(err));
			} finally {
				isLoading.set(false);
			}
		} else if (query.length === 0) {
			loadDiscover();
		}
	}, 400);
}

// --- System Actions ---
export function handleInstall(appId: string): void {
	appProgress.update((p) => ({ ...p, [appId]: { status: 'starting', percentage: 0 } }));
	wailsApp.InstallApp(appId, true).catch(console.error);
}

export function handleUninstall(appId: string): void {
	appProgress.update((p) => ({ ...p, [appId]: { status: 'removing', percentage: 0 } }));
	wailsApp.UninstallApp(appId, true).catch(console.error);
}

export function handleUpdate(appId: string): void {
	appProgress.update((p) => ({ ...p, [appId]: { status: 'starting', percentage: 0 } }));
	wailsApp.UpdateApp(appId, true).catch(console.error);
}

// Helper to color the shadcn Progress bar based on state
export function getProgressColorClass(status: string): string {
	if (status === 'removing') return '[&>div]:bg-red-500';
	if (status === 'installing') return '[&>div]:bg-green-500';
	return '[&>div]:bg-blue-500'; // downloading or starting
}

// --- UI Helpers ---
export function handleImageError(e: Event) {
	(e.currentTarget as HTMLImageElement).src = 'https://dl.flathub.org/assets/default/settings.svg';
}

// --- Theme Management ---
export const currentTheme = writable<Theme>('system');
let mediaQuery: MediaQueryList;
let themeListener: (e: MediaQueryListEvent) => void;

export function applyTheme(theme: Theme) {
	currentTheme.set(theme);

	const root = document.documentElement;

	if (theme === 'system') {
		const systemPrefersDark = window.matchMedia('(prefers-color-scheme: dark)').matches;
		root.classList.toggle('dark', systemPrefersDark);
	} else {
		root.classList.toggle('dark', theme === 'dark');
	}

	localStorage.setItem('theme', theme);
}

// --- Lifecycle wiring (called from App.svelte onMount / onDestroy) ---
export function initApp(): void {
	const savedTheme = (localStorage.getItem('theme') as Theme) || 'system';
	applyTheme(savedTheme);

	mediaQuery = window.matchMedia('(prefers-color-scheme: dark)');
	themeListener = () => {
		if (get(currentTheme) === 'system') {
			applyTheme('system');
		}
	};
	mediaQuery.addEventListener('change', themeListener);

	runtime.EventsOn('flatpak:progress', (payload: ProgressPayload) => {
		const appName =
			get(apps).find((a) => a.flatpakAppId === payload.appId)?.name ||
			get(installedApps).find((a) => a.appId === payload.appId)?.name ||
			payload.appId;

		appProgress.update((p) => ({
			...p,
			[payload.appId]: {
				status: payload.status,
				percentage: payload.percentage,
				name: appName
			}
		}));

		// Handle queue advancement
		if (get(isUpdatingAll) && payload.appId === get(currentQueueAppId)) {
			if (payload.status === 'completed' || payload.status === 'error') {
				updateQueue.update((q) => q.filter((id) => id !== get(currentQueueAppId)));
				processNextUpdate();
			}
		}

		// Clean up completed/error items from the popover after 5 seconds
		if (payload.status === 'completed' || payload.status === 'error') {
			setTimeout(() => {
				appProgress.update((p) => {
					const newProgress = { ...p };
					delete newProgress[payload.appId];
					return newProgress;
				});

				wailsApp
					.GetInstalledApps()
					.then((res: InstalledApp[]) => {
						installedApps.set(res || []);
					})
					.catch(console.error);

				if (get(activeCategory) === 'Installed') loadInstalled();
			}, 5000);
		}
	});

	wailsApp
		.GetInstalledApps()
		.then((res: InstalledApp[]) => {
			installedApps.set(res || []);
		})
		.catch(console.error);

	// AppStream catalog sync — check current status immediately in case the
	// backend already finished syncing before the frontend mounted, and also
	// listen for the events emitted once startup() completes the sync.
	wailsApp
		.IsCatalogReady()
		.then((ready: boolean) => {
			if (ready) catalogStatus.set('ready');
		})
		.catch(console.error);

	runtime.EventsOn('catalog:ready', () => {
		catalogStatus.set('ready');
		catalogError.set('');
	});

	runtime.EventsOn('catalog:error', (message: string) => {
		catalogStatus.set('error');
		catalogError.set(message || 'Failed to sync AppStream catalog.');
	});

	loadDiscover();
}

export function destroyApp(): void {
	stopHeroSlide();
	if (mediaQuery && themeListener) {
		mediaQuery.removeEventListener('change', themeListener);
	}
}

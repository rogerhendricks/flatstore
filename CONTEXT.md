# FlatStore — LLM Continuation Context

This document is a comprehensive handoff note so an LLM can resume development without re-exploring the codebase from scratch.

---

## What This Project Is

**FlatStore** is a native Linux desktop application that acts as a GUI front-end for Flathub (the Flatpak app store). It is built with:

- **Wails v2** — embeds a Chromium webview into a native Go binary; Go methods are called directly from the frontend via auto-generated (or here, manually maintained) JS/TS bindings.
- **Svelte 4 + TypeScript** — modular component architecture (as of July 2026 refactor); no routing library, just reactive Svelte stores. See "Frontend Architecture" below — `App.svelte` is now a thin composition shell (~40 lines) that renders feature components based on store state.
- **Tailwind CSS v3** — utility-first styling. Dynamic colour values (gradients, etc.) MUST use inline `style=` attributes, not dynamic Tailwind class names, because Tailwind purges classes it cannot statically detect at build time.
- **shadcn-svelte** — provides `Popover` and `Progress` components from `$lib/components/ui/`.
- **lucide-svelte** — icon library; icons imported individually.

---

## Wails Bindings — CRITICAL

The files `frontend/wailsjs/go/main/App.d.ts` and `frontend/wailsjs/go/main/App.js` are **manually maintained**. Wails normally auto-generates these, but in this workflow they are edited by hand.

**Every new public method added to `App` in `app.go` must also be added to both binding files.**

Pattern in `App.js`:
```js
export function MyNewMethod(arg1) {
  return window['go']['main']['App']['MyNewMethod'](arg1);
}
```

Pattern in `App.d.ts`:
```typescript
export function MyNewMethod(arg1: string): Promise<ReturnType>;
```

The Go type `flathub.AppSummary` maps to `flathub.AppSummary` in TypeScript (imported from `../models`).

---

## File Map

| File | Purpose |
|---|---|
| `main.go` | Wails entry point; sets window Title/Width/Height/MinWidth/MinHeight and binds `app` |
| `app.go` | All Wails-bound Go methods (`GetDiscoverApps`, `GetAppsByCategory`, `SearchApps`, `InstallApp`, `UninstallApp`, `UpdateApp`, `GetInstalledApps`, `GetPopularApps`, `GetPopularGames`, `GetPopularCreate`, `GetAppDetails`, `OpenApp`, `IsCatalogReady`) |
| `internal/flathub/client.go` | HTTP client for Flathub REST API v2; defines `AppSummary`, `apiHit`, `apiResponse`; fetch methods |
| `internal/flathub/system.go` | `SystemManager` — wraps `flatpak` CLI; runs operations in goroutines; parses stdout for progress; emits `flatpak:progress` Wails events |
| `internal/appstream/manager.go` | Downloads and parses `appstream.xml.gz` from Flathub at startup (`Sync()`); `GetApp(id)` backs `App.GetAppDetails`; `IsReady()` backs the bound `App.IsCatalogReady` method. `GetAppsByCategory` exists on the manager but is not yet exposed to the frontend (see Known Gaps). |
| `frontend/src/App.svelte` | Thin composition shell — mounts `initApp()`/`destroyApp()` lifecycle and renders `Sidebar`, `DiscoverView`, `PopularView`, `InstalledView`, `AppDetailsPage`, `ScreenshotZoomModal` based on store state |
| `frontend/src/lib/stores/appStore.ts` | **Single source of truth** — all Svelte stores (apps, installedApps, appProgress, theme, popular/hero state, details/zoom state, `catalogStatus`/`catalogError`), derived stores (`installedAppIds`, `updateableAppIds`, `activeTasksCount`, `updateableAppsCount`), and all business-logic functions (`loadDiscover`, `loadCategory`, `loadInstalled`, `loadPopular`, `handleSearch`, `handleInstall`/`handleUninstall`/`handleUpdate`, `handleUpdateAll`, `openDetails`/`closeDetails`, `applyTheme`, `initApp`/`destroyApp` which wires the Wails `flatpak:progress`, `catalog:ready`, and `catalog:error` event listeners) |
| `frontend/src/lib/types.ts` | All shared TypeScript interfaces/types (`AppDetails`, `AppSummary`, `InstalledApp`, `AppProgress`, `ProgressPayload`, `Category`, `Theme`, `PopularTab`, `FeaturedPromo`, `HeroGradient`) |
| `frontend/src/lib/constants.ts` | Static data: `categories` (sidebar list w/ icons), `heroGradients`, `featuredPromos` |
| `frontend/src/lib/components/Sidebar.svelte` | Left nav — search input, category list, theme/backup/activity popovers |
| `frontend/src/lib/components/AppCard.svelte` | Full app card (used in standard grids) |
| `frontend/src/lib/components/AppCardCompact.svelte` | Compact row card (used in Discover dashboard and Popular "More Apps" 3×N layouts) |
| `frontend/src/lib/components/AppGrid.svelte` | Reusable responsive grid of `AppCard`s with loading/error/empty states |
| `frontend/src/lib/components/CompactAppGrid.svelte` | Reusable 3-column uniform fixed-height grid of `AppCardCompact`s; evenly chunks any `apps` array across 3 columns. Shared by `DiscoverDashboard.svelte` and `PopularView.svelte` |
| `frontend/src/lib/components/DiscoverDashboard.svelte` | Discover landing dashboard — promo cards + "New Apps and Updates" 3×2 grid |
| `frontend/src/lib/components/DiscoverView.svelte` | Switches between `DiscoverDashboard` and standard `AppGrid` based on `discoverShowAll`/`activeCategory` |
| `frontend/src/lib/components/PopularView.svelte` | Popular page — tab buttons, hero carousel, "More Apps" grid |
| `frontend/src/lib/components/InstalledView.svelte` | Installed apps list + Update All button |
| `frontend/src/lib/components/AppDetailsPage.svelte` | Full app detail page (hero header, actions, metadata, screenshots, description) |
| `frontend/src/lib/components/ScreenshotZoomModal.svelte` | Fullscreen screenshot zoom overlay |
| `frontend/wailsjs/go/main/App.d.ts` | TypeScript type declarations for Go methods |
| `frontend/wailsjs/go/main/App.js` | JS bridge calling `window.go.main.App.*` |
| `frontend/wailsjs/go/models.ts` | TypeScript interface for `flathub.AppSummary` (auto-generated by Wails, leave it alone) |

---

## Go Types

### `flathub.AppSummary` (the standard DTO everywhere)
```go
type AppSummary struct {
    FlatpakAppId string `json:"flatpakAppId"`
    Name         string `json:"name"`
    Summary      string `json:"summary"`
    IconUrl      string `json:"iconUrl"`
    Version      string `json:"version"`
    Developer    string `json:"developer"`
}
```

### `flathub.InstalledApp`
```go
type InstalledApp struct {
    AppID           string `json:"appId"`
    Name            string `json:"name"`
    Version         string `json:"version"`
    UpdateAvailable bool   `json:"updateAvailable"`
}
```

### `flathub.ProgressPayload` (emitted as Wails event `"flatpak:progress"`)
```go
type ProgressPayload struct {
    AppID      string `json:"appId"`
    Status     string `json:"status"`     // "starting"|"downloading"|"installing"|"removing"|"completed"|"error"
    Percentage int    `json:"percentage"` // 0–100
}
```

### AppStream catalog events (emitted once from `startup()` in `app.go`)
```go
runtime.EventsEmit(ctx, "catalog:ready", true)      // sync succeeded
runtime.EventsEmit(ctx, "catalog:error", err.Error()) // sync failed, payload is the error string
```
The bound method `App.IsCatalogReady() bool` (wraps `appstream.Manager.IsReady()`) lets the frontend check current status on mount in case it missed the event (e.g. hot-reload during dev).

---

## Flathub REST API v2

Base URL: `https://flathub.org/api/v2`  
No authentication required.

| Method | Endpoint | Returns |
|---|---|---|
| GET | `/collection/recently-updated` | `{"hits":[...]}` — recently updated apps |
| GET | `/collection/popular` | `{"hits":[...]}` — most-installed apps |
| GET | `/collection/category/{cat}` | `{"hits":[...]}` — apps by category (lowercase: `game`, `audiovideo`, `graphics`, etc.) |
| POST | `/search` | `{"hits":[...]}` — body: `{"query":"search term"}` |

**Known bad endpoint (returns 404):** `/api/v2/popular/last-month` — do not use.

Icon URLs come directly from the `icon` field in API responses — they are absolute `https://` URLs.  
Fallback icon on error: `https://dl.flathub.org/assets/default/settings.svg`

---

## Frontend Architecture

**As of the July 2026 refactor, `App.svelte` was split from a single ~1240-line file into a store + component architecture.** All state and logic now lives in [frontend/src/lib/stores/appStore.ts](frontend/src/lib/stores/appStore.ts) as Svelte stores; components import and subscribe to (`$storeName`) or call exported functions from this module. **When adding new state or logic, add it to `appStore.ts`, not to individual components**, unless it is truly local UI-only state (e.g. a component-scoped `let` for a transient toggle).

### State (Svelte stores in `lib/stores/appStore.ts`)

```typescript
export const apps = writable<AppSummary[]>([]);           // current grid/list content
export const installedApps = writable<InstalledApp[]>([]);
export const activeCategory = writable<string | null>('Discover'); // 'Discover' | 'Popular' | 'Installed' | category id | null (search)
export const viewTitle = writable<string>('Discover');
export const searchQuery = writable<string>('');
export const isLoading = writable<boolean>(true);
export const errorMessage = writable<string>('');
export const appProgress = writable<Record<string, AppProgress>>({});  // keyed by flatpakAppId
export const discoverShowAll = writable<boolean>(false);   // false = dashboard, true = full grid

// Popular page
export const popularTab = writable<PopularTab>('discover');
export const heroApps = writable<AppSummary[]>([]);        // first 5 of popular result
export const restApps = writable<AppSummary[]>([]);        // remainder
export const heroIndex = writable<number>(0);               // current hero slide
// heroSlideInterval is a module-private variable, controlled via startHeroSlide()/stopHeroSlide()

// Derived stores
export const installedAppIds = derived(installedApps, ($i) => new Set($i.map(a => a.appId)));
export const updateableAppIds = derived(installedApps, ($i) => new Set($i.filter(a => a.updateAvailable).map(a => a.appId)));
export const updateableAppsCount = derived(installedApps, ($i) => $i.filter(a => a.updateAvailable).length);
export const activeTasksCount = derived(appProgress, ($p) => Object.values($p).filter(p => ['downloading','installing','removing'].includes(p.status)).length);
```

Details-page state (`selectedAppDetails`, `selectedAppIdForPage`, `isDetailsLoading`, `zoomedScreenshot`), update-queue state (`updateQueue`, `isUpdatingAll`, `currentQueueAppId`), and theme state (`currentTheme`) are also stores in this same file.

### Routing (no router library)
Navigation is controlled by the `activeCategory` store. Composition happens in `App.svelte`:
```svelte
{#if $selectedAppIdForPage}
  <AppDetailsPage />
{:else}
  {#if $activeCategory !== 'Installed' && $activeCategory !== 'Popular'}
    <DiscoverView />       <!-- internally switches DiscoverDashboard vs standard AppGrid via discoverShowAll -->
  {/if}
  {#if $activeCategory === 'Popular'}
    <PopularView />
  {/if}
  {#if $activeCategory === 'Installed'}
    <InstalledView />
  {/if}
{/if}
```

### Go → Frontend events
```typescript
// Wired inside initApp() in lib/stores/appStore.ts, called from App.svelte's onMount
runtime.EventsOn("flatpak:progress", (payload: ProgressPayload) => { ... });
runtime.EventsOn("catalog:ready", () => { ... });   // sets catalogStatus store to 'ready'
runtime.EventsOn("catalog:error", (message: string) => { ... }); // sets catalogStatus to 'error', catalogError to message
```
After `completed` or `error`, a 5-second `setTimeout` cleans up the `appProgress` store and refreshes `installedApps`. `destroyApp()` (called from `onDestroy`) tears down the theme media-query listener and stops the hero slide interval.

`initApp()` also calls `wailsApp.IsCatalogReady()` once on mount to pick up the case where the catalog finished syncing before the events were subscribed to (e.g. very fast sync, or dev hot-reload).

### App call pattern (using direct `window.go` bridge, not imports)
Defined once at the top of `lib/stores/appStore.ts` and used by all the functions in that module:
```typescript
const wailsApp = (window as any).go.main.App;
const runtime  = (window as any).runtime;
```
Components never call `wailsApp` directly — they import functions from `appStore.ts` (e.g. `handleInstall`, `openDetails`) instead.

---

## UI Pages

### Discover Dashboard (`lib/components/DiscoverDashboard.svelte`, shown when `activeCategory === 'Discover' && !discoverShowAll`)
- Three featured promo cards — currently **static placeholder data** (see `featuredPromos` in `lib/constants.ts`)
- "New Apps and Updates" section header with "See All" button
- **3×N grid** via `CompactAppGrid.svelte`: apps are evenly chunked into 3 columns (`Math.ceil(length/3)` per column), each column is a `bg-card` rounded card containing `AppCardCompact.svelte` rows
  - Each cell: `h-24` fixed height, `flex-col justify-between`, icon+text on top, action button bottom-right
  - App name: `text-xs font-semibold text-zinc-800 dark:text-zinc-200 line-clamp-1`
  - Summary: `text-[11px] text-zinc-500 dark:text-zinc-400 line-clamp-2`
  - Buttons: Get (primary), Update (blue), ✓ check (green), or spinner+% if busy
  - Grid wrapped in `overflow-x-auto` with `min-w-[600px]` on inner grid

### Standard Grid (`lib/components/AppGrid.svelte` + `AppCard.svelte`, shown via `DiscoverView.svelte` when `discoverShowAll` or any category/search)
- `grid-cols-[repeat(auto-fill,minmax(260px,1fr))]` responsive grid
- Each app: `article` card with icon (56×56), name, developer, summary
- Has "← Discover" breadcrumb (in `DiscoverView.svelte`) when `discoverShowAll` is true

### Popular Page (`lib/components/PopularView.svelte`, shown when `activeCategory === 'Popular'`)
- Tab buttons: Discover / Games / Create (calls `loadPopular(tab)` from the store)
- **Hero carousel**: top 5 apps, 280px tall rounded card, gradient background per slide, auto-advances every 4s
  - Pill dot navigation at bottom
  - Gradient colours: 5 entries in `heroGradients` array in `lib/constants.ts` (inline style to avoid Tailwind purge)
- **"More Apps" grid** below carousel: uses `CompactAppGrid.svelte` (same uniform 3×N fixed-height cell layout as the Discover dashboard) — has Get/Update/✓ buttons via `AppCardCompact.svelte`

### Installed Apps (`lib/components/InstalledView.svelte`, shown when `activeCategory === 'Installed'`)
- Flat list of installed Flatpaks
- Update button (blue, RefreshCw icon) if `updateAvailable`
- Uninstall button (red, Trash2 icon)
- Spinner shown if operation is in progress for that app
- "Update All" button in header when any installed app has an update available (calls `handleUpdateAll()` from the store)

---

## Known Gaps / Remaining Work

These were identified during development and not yet implemented:

1. ~~**App detail view**~~ — **DONE.** Implemented in `lib/components/AppDetailsPage.svelte`; clicking an app card calls `openDetails(appId)` (from `appStore.ts`) which fetches full details via `wailsApp.GetAppDetails` and shows a full page with screenshots, description, metadata, and install/update/uninstall actions.

2. **Featured promo cards** — Static placeholder data. Should connect to a real editorial API or curated list. The `featuredPromos` const now lives in `frontend/src/lib/constants.ts`.

3. ~~**Popular page "More Apps" buttons**~~ — **DONE.** `PopularView.svelte` now reuses `AppGrid.svelte`/`AppCard.svelte`, so the rest-apps grid has the same Get/Update/✓ button logic as everywhere else.

4. **Backup / Restore popover** — The sidebar CloudBackup popover (now in `lib/components/Sidebar.svelte`) exists but the inner `<div>` is empty. Intended feature: export a list of installed Flatpak app IDs to a file, and import/reinstall from that file.

5. ~~**AppStream catalogue**~~ — **PARTIALLY DONE.** The frontend now listens for `catalog:ready`/`catalog:error` via `catalogStatus`/`catalogError` stores in `appStore.ts` (wired in `initApp()`), and also checks `wailsApp.IsCatalogReady()` on mount in case the sync finished before the listener attached. `Sidebar.svelte` shows a subtle "Syncing app catalog..." indicator while `catalogStatus === 'syncing'` and an amber "Catalog sync failed" indicator (with the error in a `title` tooltip) on `'error'`. `App.GetAppDetails` already used `appstream.Manager.GetApp` for rich descriptions/screenshots/ratings before this change. **Still remaining:** `appstream.Manager.GetAppsByCategory` is not exposed to the frontend at all — there is no bound Go method or store wiring for **offline category browsing** using the local catalog as a fallback when the Flathub API is unreachable.

6. ~~**Update All button**~~ — **DONE.** Implemented in `lib/components/InstalledView.svelte`, backed by `handleUpdateAll()`/`updateQueue`/`isUpdatingAll`/`currentQueueAppId` in `appStore.ts`.

7. ~~**Popular page "More Apps" grid lacks uniform cell sizing**~~ — **DONE.** Extracted the Discover dashboard's 3-column fixed-height layout into a shared `lib/components/CompactAppGrid.svelte`, which evenly chunks any `apps` array into 3 columns (`Math.ceil(length/3)` per column) of `AppCardCompact.svelte` rows. `PopularView.svelte`'s "More Apps" section and `DiscoverDashboard.svelte`'s "New Apps and Updates" section both now use this same component, so cell sizing is uniform across both pages.

8. ~~**Add Developer Link**~~ In the AppDetailsPage.svelte the developer needs to be a link to a page showing all apps from that developer
---

## Build & Dev Commands

```bash
# Live development (hot reload)
wails dev

# Production build
wails build

# Go tests
go test ./...

# Svelte type check
cd frontend && npx svelte-check --tsconfig ./tsconfig.json

# Go build check (no output = clean)
go build ./...
```

---

## Common Pitfalls

| Problem | Solution |
|---|---|
| Dynamic Tailwind gradient classes missing in build | Use inline `style="background: linear-gradient(...)"` not `bg-gradient-to-r from-[#xxx]` |
| Self-closing Svelte button tag | Always `<button ...></button>` not `<button ... />` |
| Wails binding not found at runtime | Add the method to both `App.d.ts` AND `App.js` in `frontend/wailsjs/go/main/` |
| Variable cell heights in grid | Use fixed `h-{n}` height + `line-clamp-{n}` + `overflow-hidden` on the cell div |
| Flathub `/popular/last-month` returns 404 | Use `/collection/popular` instead |
| `appProgress` badge shows on PackageCheck icon | The badge `<span>` inside `<Popover.Trigger>` needs `absolute` positioning; the trigger itself needs `relative` |
| New state/logic accidentally added to a component instead of the store | All shared state and business logic belongs in `frontend/src/lib/stores/appStore.ts` as a Svelte store or exported function; components should only hold local, non-shared UI state |
| Component can't see shared state | Import the store from `$lib/stores/appStore` and prefix with `$` (e.g. `$isLoading`) to auto-subscribe in the template |
| Catalog-dependent features (details page) show fallback/limited data right after app launch | `appstream.Manager.Sync()` runs in a goroutine from `startup()` and can take a few seconds; check `$catalogStatus` (`'syncing'`\|`'ready'`\|`'error'`) from `appStore.ts` before assuming catalog-backed data (e.g. `GetAppDetails`) is fully available |


2. What details are in AppStream for each app?
AppStream (under the Freedesktop specification) stores rich application catalog metadata in XML format. For each component (application), it contains:

Identification: Unique App ID (reverse-DNS style e.g., org.mozilla.firefox), Application name (<name>), Developer/Publisher (<developer_name>).
Descriptions: Short summary (<summary>) and detailed rich-text description (<description>).
System Category Mapping: List of application category tags (<categories>).
Visuals: Cache/remote URLs for app icons (<icon>) and screenshot image galleries (<screenshots>).
Technical Details: Project licenses (<project_license>), metadata licenses (<metadata_license>).
Links: Clickable homepage, bug tracker, donation, or support links (<url>).
Version/Release History: Versions and dates of releases (<releases>).
Content Ratings (Age Ratings): An Open Age Rating Service (OARS) matrix (<content_rating>) describing levels of cartoon violence, realistic violence, mature language, etc.
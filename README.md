# FlatStore

A native Linux desktop app store for [Flathub](https://flathub.org), built with [Wails v2](https://wails.io) (Go backend) and Svelte + TypeScript (frontend). FlatStore lets you browse, search, install, update, and uninstall Flatpak applications without touching the terminal.

---

## Features

- **Discover** — browse recently-updated apps in a compact 3×2 card grid with featured promo banners
- **Popular** — hero carousel of the top 5 apps with sub-tabs for All / Games / Creative apps
- **Categories** — Audio & Video, Development, Education, Games, Graphics, Internet, Office, Science, System
- **Search** — live debounced search via the Flathub v2 API
- **Installed Apps** — list all installed Flatpaks with one-click Update / Uninstall
- **Activity Center** — animated badge + popover showing live download / install / remove progress
- **Theme switching** — Light, Dark, and System (OS preference) modes persisted to `localStorage`
- **Backup / Restore** — sidebar popover (placeholder, in progress)

---

## Tech Stack

| Layer | Technology |
|---|---|
| App framework | [Wails v2](https://wails.io) |
| Backend language | Go 1.23 |
| Frontend | Svelte 4 + TypeScript |
| Styling | Tailwind CSS v3 |
| UI components | shadcn-svelte (`popover`, `progress`) |
| Icons | lucide-svelte |
| Build tool | Vite |
| System integration | `flatpak` CLI (via `exec.CommandContext`) |
| App catalogue | Flathub AppStream XML (`appstream.xml.gz`) |

---

## Project Structure

```
flatstore/
├── main.go                     # Wails entry point; window config (min 960×600)
├── app.go                      # Wails-bound Go methods exposed to the frontend
├── app_test.go
├── go.mod
├── internal/
│   ├── flathub/
│   │   ├── client.go           # Flathub REST API v2 client
│   │   ├── client_test.go
│   │   ├── system.go           # flatpak CLI wrapper + real-time progress events
│   │   ├── system_test.go
│   │   └── models.go
│   └── appstream/
│       ├── manager.go          # AppStream XML catalogue sync & lookup
│       ├── manager_test.go
│       └── models.go
├── frontend/
│   ├── src/
│   │   ├── App.svelte          # Main UI (single component)
│   │   ├── app.css / style.css
│   │   └── lib/components/ui/ # shadcn-svelte components
│   └── wailsjs/go/main/
│       ├── App.d.ts            # TypeScript bindings (manually maintained)
│       └── App.js              # JS bindings (manually maintained)
└── build/                      # Platform-specific build assets
```

---

## Getting Started

### Prerequisites

- Go 1.23+
- Node.js 18+ and npm
- [Wails CLI](https://wails.io/docs/gettingstarted/installation): `go install github.com/wailsapp/wails/v2/cmd/wails@latest`
- `flatpak` installed and `flathub` remote configured

### Development

```bash
wails dev
```

Hot-reload is provided by Vite. The Go dev server is also available at `http://localhost:34115` for browser-based testing of Go methods.

### Build

```bash
wails build
```

Output binary is placed in `build/bin/`.

---

## API Reference

FlatStore uses the **Flathub REST API v2** (`https://flathub.org/api/v2`). No authentication is required.

| Endpoint | Used for |
|---|---|
| `GET /collection/recently-updated` | Discover page grid |
| `GET /collection/popular` | Popular → Discover tab |
| `GET /collection/category/{cat}` | Category pages + Popular Games/Create |
| `POST /search` | App search (`{"query": "..."}`) |

AppStream metadata (`https://dl.flathub.org/repo/appstream/x86_64/appstream.xml.gz`) is synced at startup for richer offline catalogue data.

---

## Architecture Notes

- **Progress events**: `flatpak` CLI output is parsed line-by-line in a goroutine and forwarded to the frontend via Wails `runtime.EventsEmit("flatpak:progress", ...)`. The frontend listens with `runtime.EventsOn`.
- **Wails bindings**: The files in `frontend/wailsjs/go/main/` (`App.d.ts` and `App.js`) are **manually maintained** — they are not auto-generated in this workflow. Every new Go method on `App` must be added to both files.
- **Dynamic Tailwind gradients**: Gradient colours are applied via inline `style=` attributes rather than Tailwind utility classes to prevent Tailwind's build-time purge from removing them.
- **Min window size**: Set to 960×600 in `main.go` via `MinWidth`/`MinHeight` options.

---

## Roadmap / Known Gaps

- [ ] Backup / Restore popover — UI shell exists, functionality not yet implemented
- [ ] App detail view — clicking an app card does nothing yet
- [ ] Featured promo cards on Discover — currently static placeholder data; intended to connect to a real editorial API
- [ ] Popular page "More Apps" grid — no Install/Update buttons yet
- [ ] Dark mode icon contrast — some Flathub icons have transparent backgrounds that are hard to see on dark cards
- [ ] AppStream catalogue integration — `appstream.Manager` syncs on startup but is not yet wired into category/search results


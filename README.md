# Wails3 Template

A production-ready **Wails3** desktop application template — frameless window, custom title bar, shadcn/ui components, 15 built-in themes with live switching, system tray, global shortcuts, and a complete cross-platform CI pipeline.

Built by abstracting the infrastructure from [CPM OCR Studio](https://github.com/HycJack/models_run_with_go_wails3) and the theme/settings system from [terax-clone](https://github.com/HycJack/terax-clone).

## Features

- **Frameless window** with a custom drag-region header + platform-native window controls (macOS traffic lights / Windows min-max-close)
- **15 built-in themes** (Claude, Dracula, Tokyo Night, Catppuccin, Nord, Gruvbox, Rose Pine, Everforest, Kanagawa, Solarized, etc.) with click-to-switch + live preview
- **Light / Dark / System** appearance mode toggle
- **shadcn/ui** component library (Button, Switch, Slider, Tabs, Label, Separator) + **lucide-react** icons
- **Typed settings persistence** — Go struct → Wails3 generates typed TS bindings → zustand store with debounced disk writes
- **System tray** + global shortcuts (`CmdOrCtrl+Alt+M` to show window)
- **Atomic config writes** (tmp + rename) with `sync.Mutex` concurrency guard
- **Cross-platform CI** (Windows, macOS, Linux) with artifact uploads + GitHub Releases on tag push

## Tech Stack

| Layer | Technology |
|-------|-----------|
| Backend | Go 1.27+ · Wails3 v3.0.0-beta |
| Frontend | React 18 · TypeScript · Vite · Tailwind CSS v4 |
| UI | shadcn/ui · Radix Primitives · lucide-react |
| State | zustand (debounced persist to Go backend) |
| Build | Wails3 Taskfile · NSIS (Windows) · DMG (macOS) |

## Project Structure

```
main.go                      Wails3 entry (config → services → window → tray)
internal/
  config/                    Generic JSON config (Load/Save/EnsureDirs + HTTPClient)
  app/
    state.go                 Shared State (App/Window refs, Emit, OpenFolder)
    greeter_service.go       Example service (Greet + Ping, demonstrates bindings)
    settings_service.go      Typed Preferences persistence (atomic write + mutex)
    tray.go                  System tray + global shortcuts + hide-to-tray
frontend/
  src/
    App.tsx                  Frameless layout: header + sidebar + content
    pages/
      HomePage.tsx           Example: calling bound Go methods
      SettingsPage.tsx       General (mode/zoom/toggles) + Themes picker
    components/
      ui/                    shadcn components (button, switch, slider, ...)
      WindowControls.tsx     Platform window controls (min/max/close)
    modules/
      settings/store.ts      zustand store (debounced persist via SettingsService)
      theme/                 Theme engine (types, applyTheme, ThemeProvider, 15 themes)
    lib/                     cn() util, toast()
  vite.config.js             Vite + React + Tailwind + Wails plugin
build/                       Cross-platform build assets (Taskfiles, icons, packaging)
Taskfile.yml                 wails3 task entry (build / dev / package / run)
```

## Quick Start

### Prerequisites

- **Go** 1.27+
- **Node.js** 20+
- **wails3 CLI**: `go install github.com/wailsapp/wails/v3/cmd/wails3@latest`
- **C compiler** (for CGO): GCC (Windows: mingw-w64, macOS: Xcode CLT, Linux: gcc)

### Build & Run

```bash
# 1. Install frontend dependencies
cd frontend && npm install && cd ..

# 2. Generate Wails3 bindings (TS models from Go structs)
wails3 generate bindings

# 3. Build the production binary
wails3 build          # → bin/skeleton.exe (Windows) / bin/skeleton (Linux) / bin/skeleton.app (macOS)

# 4. Development mode (hot reload)
wails3 task dev       # Vite dev server + Go backend, auto-reload on change
```

## How to Extend

### Add a new Service

```go
// internal/app/my_service.go
package app

type MyService struct{ state *State }

func NewMyService(s *State) *MyService { return &MyService{state: s} }

// Every exported method becomes callable from the frontend after
// `wails3 generate bindings`.
func (s *MyService) DoSomething(input string) (string, error) {
    return "result: " + input, nil
}
```

Register in `main.go`:
```go
Services: []application.Service{
    application.NewService(app.NewGreeterService(state)),
    application.NewService(app.NewSettingsService(state)),
    application.NewService(app.NewMyService(state)),  // ← add here
},
```

Then:
```bash
wails3 generate bindings   # regenerate TS bindings
```

Frontend usage:
```ts
import { DoSomething } from "@bindings/skeleton/internal/app/myservice";
const result = await DoSomething("hello");
```

### Add a new setting

1. Add field to `Preferences` struct in `internal/app/settings_service.go` + `DefaultPreferences`
2. Add to `Preferences` type + `DEFAULT_PREFERENCES` in `frontend/src/modules/settings/store.ts`
3. Add UI control in `SettingsPage.tsx`
4. `wails3 generate bindings` — the typed model updates automatically

### Add a new theme

Create `frontend/src/modules/theme/themes/my-theme.ts`:
```ts
import type { Theme } from "../types";

export const myTheme: Theme = {
  id: "my-theme",
  name: "My Theme",
  variants: {
    dark: {
      colors: {
        background: "#1a1b26",
        foreground: "#c0caf5",
        primary: "#7aa2f7",
        // ... see ThemeColors for all available tokens
      },
    },
  },
};
```

Register in `themes/index.ts`:
```ts
import { myTheme } from "./my-theme";
const BUILTIN: Theme[] = [..., myTheme];
```

### Add a new page

1. Create `frontend/src/pages/MyPage.tsx`
2. Add to `NAV` array in `App.tsx`:
```ts
const NAV = [
  { key: "home", label: "首页", icon: Home },
  { key: "settings", label: "设置", icon: Settings },
  { key: "mine", label: "我的", icon: User },  // ← add here
] as const;
```

## CI/CD

GitHub Actions workflows in `.github/workflows/`:

| Workflow | Platform | Artifact |
|----------|----------|----------|
| `build-windows.yml` | Windows | `skeleton.exe` + NSIS installer |
| `build-macos.yml` | macOS (universal) | `skeleton.app` + DMG |
| `build-linux.yml` | Linux | `skeleton` AppImage |

All workflows:
- Push to `main` → build + upload artifact
- Push a `v*` tag → build + create draft GitHub Release with binaries
- Manual dispatch supported

## License

MIT

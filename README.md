# Music Assist 

A Windows desktop application for browsing numbered musical notation images and generating print-ready Docx documents.

Built with **Go + Wails v3** on the backend and **Vue 3 + TypeScript + Element Plus + Tailwind CSS** on the frontend.

## Features

- **Image Viewer**
  - Navigate with arrow keys, mouse wheel, or double-click
  - Fullscreen browsing with auto-fading title tips
  - Base64-cached preview for smooth switching

- **Source Management**
  - Add folders (recursive / non-recursive) or individual image files
  - Remove sources with a single click
  - Automatic change detection when the window regains focus

- **Search & Sort**
  - Search by name, pinyin, or pinyin initials (powered by `tiny-pinyin`)
  - Multiple sort strategies, persisted across sessions

- **Docx Generation**
  - One-click export of the current library to a print-ready `.docx`
  - Auto-generated catalog page with date
  - A4 layout, image auto-sizing, page numbers, optional per-directory splitting

- **Remote Sync**
  - Pull notation packs from a remote HTTP server
  - Real-time sync progress via Wails events
  - Remote server repository: [render-test](https://github.com/wtl568594290/render-test)

- **Auto Update**
  - Check / download / verify / install / restart in one flow
  - Custom Gitee Releases provider with progress and speed display
  - Manual check from the About tab, automatic check every 2 hours

## Tech Stack

| Layer    | Technology                                                            |
| -------- | --------------------------------------------------------------------- |
| Platform | Windows 10 / 11 (x64)                                                 |
| Backend  | Go 1.26, Wails v3, wordZero (Docx), imaging, yaml, ini                |
| Frontend | Vue 3, TypeScript, Vite 7, Element Plus, Tailwind CSS 4, VueUse       |
| Build    | Task (Taskfile), wails3 CLI, NSIS / MSIX packaging                    |
| Updater  | Custom Gitee Releases provider (wails updater)                        |

## Project Layout

```
music-assist/
├── main.go              # Wails application entry
├── build.go             # go:embed static config / README
├── services/            # Wails-bound services
│   ├── appService.go
│   ├── soureService.go  # Image source & change detection
│   ├── documentService.go # Docx generation
│   ├── syncService.go   # Remote sync
│   └── updateService.go # Auto update
├── updater/             # Custom Gitee update provider
├── tools/               # Helpers (sort, notifier, static config, ...)
├── config/              # Runtime config persistence
├── build/               # Windows build / packaging assets (NSIS, MSIX)
└── frontend/            # Vue 3 + Vite SPA
    └── src/views/home/  # Main UI (viewer, sources, doc, sync, update, about)
```

## Getting Started

### Prerequisites

- Go 1.26+
- Node.js 22+ and npm
- [wails3 CLI](https://v3.wails.io/getting-started/installation/)

### Develop

```bash
wails3 dev
```

Starts Vite dev server and the Wails app with hot-reload.

### Build & Package

```bash
# Build executable
wails3 build

# Package installer (NSIS / MSIX)
wails3 package
```

Outputs are placed under `bin/`.

## Remote Sync Server

The companion remote sync server lives in a separate repository:

**[https://github.com/wtl568594290/render-test](https://github.com/wtl568594290/render-test)**

Use it together with the **Remote URL** field inside the app to download shared notation packs.

## License

This project is licensed under the [MIT License](LICENSE).

Copyright © 2026 wtl1989.

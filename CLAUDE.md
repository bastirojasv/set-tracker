# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build Commands

**Full build with icon embedding (PowerShell):**
```powershell
./do_build.ps1
```
Generates a multi-size ICO from `bear.png` via `windres`, then compiles the Windows GUI executable.

**Simple build (Batch):**
```
build.bat
```
Requires Go 1.22+ and MSYS2 MinGW-w64 GCC on PATH.

**Manual build:**
```bash
go build -ldflags="-H windowsgui" -o BearTracker.exe .
```

**Dependencies:**
```bash
go mod download
```

There are no tests or linting configurations in this project.

## Architecture Overview

**SetTracker** (branded "Bear Tracker") is a Windows desktop GUI app for tracking tennis/racquet sport match sets. It is written in Go using the Fyne framework with SQLite for persistence.

All Go source files are in a single `main` package. The architecture has three clear layers:

### UI Layer
- `app.go` — The central `App` struct that owns the Fyne window, database handle, and all navigation. Methods on `*App` implement screen transitions (`show*`) and actions (`recordGame`, `buildMain`).
- `screen_*.go` — One file per view: `onboarding`, `newset`, `activeset`, `victory`, `history`, `h2h`, `profile`.
- `widgets.go` — Custom Fyne widgets (`ActionButton`, `ColorSwatch`, `SetTypeToggle`, `Stepper`, `SmallButton`) that embed `widget.BaseWidget` and implement `CreateRenderer()`.
- `theme.go` / `colors.go` — Custom Fyne theme implementing `Color()`, `Font()`, `Icon()`, `Size()`. Design tokens (`ColorBG`, `ColorSurface`, `ColorText`, `ColorUser`, `ColorOpponent`, etc.) defined in `colors.go`.

### Data Layer
- `db.go` — All SQLite access via methods on a `*DB` receiver. Uses raw SQL (no ORM), WAL mode, and foreign keys. Three tables: `user_profile`, `sets`, `games`.
- `models.go` — Plain Go structs: `UserProfile`, `SetRecord`, `GameRecord`, `ActiveSetState`.

### Entry Point
- `main.go` — Creates the Fyne app with the embedded `bear.png` icon, opens the DB, and calls `app.start()`.

### Key Data Flows
1. **App start:** `app.start()` → check DB for profile → Onboarding if missing, else load `ActiveSetState`.
2. **Record game:** `recordGame(winner)` → `db.RecordGame()` → updates `games` + `sets` tables → refreshes score display in-place.
3. **Complete set:** Victory screen → `db.CompleteSet()` → sets `status='completed'` and `user_won` flag → navigates to History.
4. **History/H2H:** `db.LoadHistory()` / `db.LoadH2H(opponent)` — same query with optional `WHERE opponent_name = ?` filter.

## Set Type Values
- `"FT"` — First-to-N (win after reaching N games)
- `"BO"` — Best-of-N (win after ⌈N/2⌉ games)
- `"LS"` — Libre/Free (manual termination)

## Conventions
- UI labels are in Spanish (JUGADOR, ADVERSARIO, GANÓ, etc.).
- Screen methods follow `show*` naming (`showNewSet`, `showActiveSet`); builder helpers use `build*`.
- Colors are stored as `#RRGGBB` hex strings in the DB.
- `ActiveSetState` is hydrated from the DB on startup and updated incrementally in memory as games are recorded.
- The DB file lives at `%APPDATA%\SetTracker\settracker.db` at runtime.
- Windows resource compilation artifacts (`icon.rc`, `rsrc.syso`) are committed and used during the build to embed the icon.

# jenkins_tui (jui)

A terminal (TUI) client for browsing Jenkins from your terminal. jui provides interactive views for the build queue, executors, projects, builds and build logs.

Status: functional — the app supports browsing jobs, viewing build history and logs, multi-branch projects, and live refresh while builds are running. Starting new builds is not implemented yet.

## Features
- Home view with: build queue, build executors and project list
- Project pages: build list, build details and live-tailed build log
- Multi-branch job support (branch listing and navigation)
- Configuration screen to set the Jenkins API URL, User, Token and custom HTTP headers
- Keyboard-driven navigation and an in-app help popup (press `?`)

## Keyboard shortcuts (quick)
- `?` — open help
- `q` / `Ctrl-C` — quit
- `Tab` — cycle focus between panels
- `1` / `2` / `3` — focus Queue / Executors / Projects
- `Enter` — open selected item (project or build log)
- `C` — open Config
- Config page: `Enter` edit API URL, `a` add header, `e` edit header, `d` delete header
- Project page: `B` focus builds, `L` focus log, `Esc` go back

## Requirements
- Go 1.20+ (or compatible)
- Network access to a Jenkins server (set API URL in config)

## Configuration
On first run a config file is created at `config/config.ini` and a default header `Content-Type = application/json` is added.
- Set the Jenkins API base URL, User and Token from the in-app Config screen (press `C`) or by editing `config/config.ini`.
- Additional HTTP headers can be added/edited from the Config screen.

## Project layout
- `main.go` — application entrypoint
- `config/` — config loader and persistence (uses INI)
- `view/` — TUI views (home, project, config, help)
- `data/` — API client helpers

## Installing
```bash
go install github.com/glgaspar/jui@latest
# or build locally
go build -o jui
```

## Running
1. Run the binary (`jui`)
2. If API URL is empty the app opens the Config screen — set the Jenkins host (e.g. `jenkins.example.com` or full URL)
3. Use `?` for the help popup with full controls

## Notes for contributors
- Uses `rivo/tview` and `gdamore/tcell` for the TUI
- Config persistence uses `gopkg.in/ini.v1`
- Live refresh for the Home page is implemented (periodic polling); build logs are polled while a build is running
- Opening issues or PRs welcome; please run `go build` and test changes locally
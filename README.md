# 📡 Antenna

[![Release](https://img.shields.io/github/v/release/Caryyon/antenna?style=flat-square)](https://github.com/Caryyon/antenna/releases)
[![Downloads](https://img.shields.io/github/downloads/Caryyon/antenna/total?style=flat-square)](https://github.com/Caryyon/antenna/releases)
[![License](https://img.shields.io/github/license/Caryyon/antenna?style=flat-square)](LICENSE)
[![Go](https://img.shields.io/badge/Go-1.22+-00ADD8?style=flat-square&logo=go)](https://golang.org)
[![Wails](https://img.shields.io/badge/Wails-v2-red?style=flat-square)](https://wails.io)

**The missing dashboard for [OpenClaw](https://github.com/openclaw/openclaw) AI agents.**

Monitor your agent sessions, track costs, and watch sub-agents work — all from a native desktop app.

![Antenna Screenshot](screenshot.png)

## Why Antenna?

OpenClaw is powerful, but when you're running multiple agents, sub-agents, and cron jobs, it's hard to see what's happening. Antenna gives you:

- **Real-time visibility** — See all sessions at a glance
- **Cost tracking** — Know exactly what you're spending today vs. total
- **Sub-agent monitoring** — Track spawned tasks and their progress  
- **Cron job status** — See your scheduled jobs and their history
- **Native performance** — No browser tab, just a fast desktop app

## Download

**[⬇️ Download Latest Release](https://github.com/Caryyon/antenna/releases/latest)**

| Platform | File | Install |
|----------|------|---------|
| **macOS** | `Antenna-macOS.dmg` | Open DMG → Drag to Applications |
| **Windows** | `Antenna-Windows.zip` | Extract → Run `Antenna.exe` |
| **Linux** | `Antenna-Linux.tar.gz` | Extract → Run `./Antenna` |

> **First launch on macOS:** Right-click → Open (to bypass Gatekeeper for unsigned apps)

## Features

| Feature | Description |
|---------|-------------|
| 🖥️ **Native App** | Runs in its own window, no browser needed |
| 🔄 **Live Updates** | Auto-refreshes every 5 seconds |
| 📊 **Session Tracking** | Main sessions, sub-agents, and cron jobs |
| 💰 **Cost Monitoring** | Today's spend vs. total spend |
| 🏷️ **Smart Labels** | Shows cron job names from your config |
| 🌙 **Dark Theme** | Easy on the eyes, matches your terminal |
| ⌨️ **Keyboard Shortcuts** | Cmd+R to refresh, Cmd+W to close |

## Two Interfaces

Antenna ships in two flavors:

| | **GUI** (Desktop) | **TUI** (Terminal) |
|---|---|---|
| **Tech** | Wails + Web frontend | Bubble Tea + Lip Gloss |
| **Install** | Download from Releases | Download from Releases or `go install` |
| **Entry point** | `cmd/antenna/` (root `main.go`) | `cmd/antenna-tui/` |
| **Platforms** | macOS, Windows, Linux | Anywhere with a terminal |

### TUI Screenshot

```
📡 Antenna — OpenClaw Monitor  Today: $0.1234  Total: $1.5678  Sessions: 12
────────────────────────────────────────────────────────────────
  24h Activity  ▁▁▃▅▇█▆▃▁▁▁▁▁▂▃▅▇▆▄▂▁▁▁▁

  ● My Session           main    42     $0.0500   $0.1200   5m ago
  ○ cron-heartbeat       cron    10     $0.0100   $0.0300   2h ago
  ● antenna-tui          sub     8      $0.0080   $0.0080   just now

j/k: navigate  enter: details  r: refresh  tab: toggle  q: quit
```

## Requirements

- [OpenClaw](https://github.com/openclaw/openclaw) installed with session data in `~/.openclaw`

## Build from Source

### GUI (Desktop App)

```bash
# Install Wails CLI
go install github.com/wailsapp/wails/v2/cmd/wails@v2.9.0

# Clone
git clone https://github.com/Caryyon/antenna.git
cd antenna

# Run in dev mode
wails dev

# Build release
wails build
```

### TUI (Terminal)

```bash
git clone https://github.com/Caryyon/antenna.git
cd antenna

# Build
go build -o antenna-tui ./cmd/antenna-tui

# Or install directly
go install github.com/Caryyon/antenna/cmd/antenna-tui@latest

# Run
./antenna-tui
```

### TUI Configuration

| Env Variable | Default | Description |
|---|---|---|
| `OPENCLAW_DIR` | `~/.openclaw` | Path to OpenClaw data directory |
| `ANTENNA_INTERVAL` | `5s` | Auto-refresh polling interval |

### TUI Keybindings

| Key | Action |
|---|---|
| `j` / `k` / `↑` / `↓` | Navigate sessions |
| `Enter` | View session details |
| `Esc` / `q` | Back / Quit |
| `Tab` | Toggle list ↔ detail |
| `r` | Force refresh |

## Roadmap

- [ ] Remote host support (SSH to monitor remote OpenClaw instances)
- [ ] Menu bar mode (always visible in system tray)
- [ ] Notifications for long-running tasks
- [ ] Session filtering and search
- [ ] Cost alerts and budgets

## Contributing

Contributions welcome! See [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

## Related Projects

- [OpenClaw](https://github.com/openclaw/openclaw) — The AI agent framework Antenna monitors
- [Wails](https://wails.io) — The Go framework powering Antenna's native UI

## License

[MIT](LICENSE) © 2026 Cary Wolff

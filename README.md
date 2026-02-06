# 📡 Antenna

[![Release](https://img.shields.io/github/v/release/Caryyon/antenna?style=flat-square)](https://github.com/Caryyon/antenna/releases)
[![License](https://img.shields.io/github/license/Caryyon/antenna?style=flat-square)](LICENSE)
[![Go](https://img.shields.io/badge/Go-1.22+-00ADD8?style=flat-square&logo=go)](https://golang.org)
[![Wails](https://img.shields.io/badge/Wails-v2-red?style=flat-square)](https://wails.io)

A native desktop app for monitoring [OpenClaw](https://github.com/openclaw/openclaw) agent sessions.

![Antenna Screenshot](screenshot.png)

## Download

**[⬇️ Download Latest Release](https://github.com/Caryyon/antenna/releases/latest)**

| Platform | File | Install |
|----------|------|---------|
| **macOS** | `Antenna-macOS.dmg` | Open DMG → Drag to Applications |
| **Windows** | `Antenna-Windows.zip` | Extract → Run `Antenna.exe` |
| **Linux** | `Antenna-Linux.tar.gz` | Extract → Run `./Antenna` |

> **Note:** On macOS, first launch requires right-click → Open (unsigned app).

## Features

- 🖥️ **Native app** — Runs in its own window, no browser needed
- 🔄 **Live updates** — Auto-refreshes every 5 seconds
- 📊 **Session tracking** — Main sessions, sub-agents, cron jobs
- 💰 **Cost monitoring** — Today's spend vs total spend
- 🏷️ **Smart labels** — Shows cron job names from your config
- 🌙 **Dark theme** — Easy on the eyes

## Requirements

- [OpenClaw](https://github.com/openclaw/openclaw) installed with session data in `~/.openclaw`

## Build from Source

### Prerequisites

- [Go 1.22+](https://golang.org/dl/)
- [Node.js 20+](https://nodejs.org/)
- [Wails CLI](https://wails.io/docs/gettingstarted/installation)

### Build

```bash
# Install Wails
go install github.com/wailsapp/wails/v2/cmd/wails@v2.9.0

# Clone
git clone https://github.com/Caryyon/antenna.git
cd antenna

# Run in dev mode
wails dev

# Or build release
wails build
```

## Contributing

Contributions welcome! See [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

## License

[MIT](LICENSE) © 2026 Cary Wolff

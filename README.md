# 📡 Antenna

A native desktop app for monitoring [OpenClaw](https://github.com/openclaw/openclaw) sessions.

![Antenna](screenshot.png)

## Download

**[→ Download Latest Release](https://github.com/Caryyon/antenna/releases/latest)**

| Platform | File |
|----------|------|
| **macOS** | `Antenna-macOS.zip` |
| **Windows** | `Antenna-Windows.zip` |
| **Linux** | `Antenna-Linux.tar.gz` |

### Installation

- **macOS**: Unzip → drag `Antenna.app` to Applications
- **Windows**: Unzip → run `Antenna.exe`
- **Linux**: Extract → run `./Antenna`

## Features

- **Native app** — Runs in its own window, no browser
- **Live updates** — Auto-refreshes every 5 seconds
- **Session tracking** — Main sessions, sub-agents, cron jobs
- **Cost monitoring** — Today's spend vs total

## Requirements

OpenClaw must be installed with session data in `~/.openclaw`

## Build from Source

```bash
# Install Wails
go install github.com/wailsapp/wails/v2/cmd/wails@v2.9.0

# Clone & build
git clone https://github.com/Caryyon/antenna.git
cd antenna
wails build
```

## License

MIT

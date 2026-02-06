# 📡 Antenna

A real-time session monitor for [OpenClaw](https://github.com/openclaw/openclaw).

![Antenna Dashboard](https://raw.githubusercontent.com/Caryyon/antenna/main/screenshot.png)

## Features

- **Live dashboard** — Auto-refreshes every 5 seconds
- **Session tracking** — Main sessions, sub-agents, and cron jobs
- **Cost monitoring** — Today's spend vs total spend
- **Cron job names** — Shows actual job names from config
- **Transcript viewer** — Click any session to view full conversation

## Installation

### Using Go

```bash
go install github.com/Caryyon/antenna@latest
```

### Download Binary

Grab the latest release from [Releases](https://github.com/Caryyon/antenna/releases) for your platform:
- macOS (Intel & Apple Silicon)
- Linux (amd64 & arm64)
- Windows (amd64 & arm64)

### From Source

```bash
git clone https://github.com/Caryyon/antenna.git
cd antenna
go build -o antenna .
```

## Usage

```bash
# Start the dashboard
antenna

# Custom port
PORT=8080 antenna

# Custom OpenClaw directory
OPENCLAW_DIR=/path/to/.openclaw antenna
```

Then open http://localhost:3600

## Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | `3600` | HTTP server port |
| `OPENCLAW_DIR` | `~/.openclaw` | OpenClaw data directory |

## How It Works

Antenna reads directly from OpenClaw's local data:

- `~/.openclaw/agents/main/sessions/` — Session transcripts
- `~/.openclaw/agents/main/sessions/sessions.json` — Session metadata
- `~/.openclaw/cron/jobs.json` — Cron job definitions

No API keys needed — read-only access to local files.

## License

MIT

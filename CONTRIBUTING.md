# Contributing to Antenna

Thanks for your interest in contributing! Antenna is a simple project and we welcome contributions of all kinds.

## Getting Started

### Prerequisites

- [Go 1.22+](https://golang.org/dl/)
- [Node.js 20+](https://nodejs.org/)
- [Wails v2](https://wails.io/docs/gettingstarted/installation)

### Setup

```bash
# Clone the repo
git clone https://github.com/Caryyon/antenna.git
cd antenna

# Install Wails CLI
go install github.com/wailsapp/wails/v2/cmd/wails@v2.9.0

# Run in dev mode (hot reload)
wails dev
```

### Building

```bash
# Build for your current platform
wails build

# Build for specific platform
wails build -platform darwin/universal  # macOS
wails build -platform windows/amd64     # Windows
wails build -platform linux/amd64       # Linux
```

## Project Structure

```
antenna/
├── app.go              # Go backend - session data loading
├── main.go             # Wails app entry point
├── frontend/
│   ├── src/
│   │   ├── main.js     # Frontend logic
│   │   └── style.css   # Styles
│   └── index.html
├── build/              # Platform-specific build configs
└── wails.json          # Wails configuration
```

## How to Contribute

### Reporting Bugs

Open an issue with:
- What you expected to happen
- What actually happened
- Steps to reproduce
- Your OS and Antenna version

### Suggesting Features

Open an issue describing:
- The problem you're trying to solve
- Your proposed solution
- Any alternatives you've considered

### Pull Requests

1. Fork the repo
2. Create a branch (`git checkout -b feature/cool-thing`)
3. Make your changes
4. Test locally with `wails dev`
5. Commit with a clear message
6. Push and open a PR

### Code Style

- Go: Follow standard Go conventions (`gofmt`)
- JavaScript: Keep it simple, no framework needed
- CSS: Use CSS variables for theming

## Areas We'd Love Help With

- [ ] Menu bar icon/tray support
- [ ] Keyboard shortcuts
- [ ] Session filtering/search
- [ ] Dark/light theme toggle
- [ ] Notifications for new sessions
- [ ] Performance improvements for large session counts

## Questions?

Open an issue or find us on [Discord](https://discord.com/invite/clawd).

## License

By contributing, you agree that your contributions will be licensed under the MIT License.

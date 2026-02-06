# Changelog

All notable changes to Antenna will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.0.2] - 2026-02-06

### Added
- Custom antenna app icon (Gmork green theme)
- About dialog with version and GitHub links
- Help menu with links to GitHub, issues, and OpenClaw docs
- Keyboard shortcut Cmd+R to refresh
- Better DMG installer with drag-to-Applications

### Fixed
- Fixed infinite "Loading..." state - now shows error messages
- Added empty state when no sessions found
- Better null/undefined handling in frontend

## [1.0.0] - 2026-02-06

### Added
- Initial release
- Native desktop app using Wails v2
- Session monitoring dashboard
- Live auto-refresh (5 second intervals)
- Cost tracking (today and total)
- Support for main sessions, sub-agents, and cron jobs
- Cron job name resolution from config
- Dark theme with JetBrains Mono font
- macOS universal binary (Intel + Apple Silicon)
- Windows x64 executable
- Linux x64 binary

[1.0.2]: https://github.com/Caryyon/antenna/releases/tag/v1.0.2
[1.0.1]: https://github.com/Caryyon/antenna/releases/tag/v1.0.1
[1.0.0]: https://github.com/Caryyon/antenna/releases/tag/v1.0.0

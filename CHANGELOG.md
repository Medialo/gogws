# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.0.0] - 2026-01-25

### Added

#### Core Commands
- `status` — Show status of all repositories (alias: `st`)
- `fetch` — Fetch updates from all remotes
- `ff` — Fast-forward pull all repositories
- `update` — Clone missing repositories and workspaces
- `clone` — Clone specific repositories by path
- `check` — Check workspace consistency

#### Initialization
- `init` — Parent command for workspace initialization
- `init projects` — Discover git repositories and create projects.gws
- `init workspaces` — Interactively configure nested workspaces
- `init gitignore` — Generate or update .gitignore with GWS section

#### Configuration
- `config list` — List available configuration keys
- `config get` — Get a configuration value
- `config set` — Set a configuration value

#### Execution Engine
- Parallel execution with configurable workers (`--parallel`)
- Serial mode for debugging (`--parallel=1`)
- Stop on first error (`--stop-on-error`)
- Automatic serial fallback when SSH agent unavailable

#### Hooks System
- Git-style file-based hook discovery
- Global hooks in `~/.gws/hooks/`
- Local hooks in `<workspace>/.gws/hooks/`
- Trust system with three modes: `ask`, `all`, `skip`
- Trusted workspaces configuration with wildcard patterns
- Environment variables passed to hooks

#### Configuration
- `.gws/` directory support for workspace configuration
- Legacy root-level files still supported (`.projects.gws`, `.workspaces.gws`)
- User configuration in `~/.gws/config.yaml`
- Environment variable overrides (`GOGWS_*`)

#### Output
- Colored terminal output with Lipgloss
- JSON export (`--format=json`)
- YAML export (`--format=yaml`)
- Filter dirty repos only (`--only-changes`)
- Custom themes support

#### Other
- Nested workspaces via `.workspaces.gws`
- Custom gitignore templates
- Shell completion (bash, zsh, fish, powershell)
- Cross-platform support (Linux, macOS, Windows)

### Compatibility

- Compatible with gws `.projects.gws` file format
- Supports multiple remotes per repository
- Automatic migration path from legacy config file locations

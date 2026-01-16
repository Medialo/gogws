# gogws

Fast, parallel Git workspace manager written in Go.

Manage dozens of Git repositories as a single workspace. Clone, fetch, pull, and check status across all your projects with one command.

```
┌─────────────────────────────────────────────────────────────────────────┐
│  GOGWS - Workspace Status - my-workspace                                │
├─────────────────────────────────────────────────────────────────────────┤
│  ● api                    main     ✓                                    │
│  ● frontend               main     ↑2                                   │
│  ● shared-lib             develop  3 uncommitted                        │
│  ○ new-service            —        (not cloned)                         │
├─────────────────────────────────────────────────────────────────────────┤
│  Projects: 4  │  Clean: 1  │  Changed: 2  │  Missing: 1                 │
└─────────────────────────────────────────────────────────────────────────┘
```

## Features

- **Parallel execution** — Configurable workers for fast operations across many repos
- **Nested workspaces** — Organize projects hierarchically with sub-workspaces
- **Git-style hooks** — Global and local hooks with trust system
- **Multiple output formats** — Text, JSON, YAML for scripting and CI/CD
- **Smart defaults** — Works out of the box, highly configurable when needed
- **Cross-platform** — Linux, macOS, Windows
- **gws compatible** — Drop-in replacement for [gws](https://github.com/StreakyCobra/gws)

## Installation

### From source

```bash
go install github.com/medialo/gogws/cmd/gogws@latest
```

### Build locally

```bash
git clone https://github.com/medialo/gogws.git
cd gogws
make build
```

## Quick Start

```bash
# Initialize workspace from existing repositories
cd ~/projects
gogws init

# Check status of all repositories
gogws status

# Fetch updates from all remotes
gogws fetch

# Fast-forward pull all repositories
gogws ff

# Clone missing repositories
gogws update
```

## Workspace Structure

```
my-workspace/
├── .gws/
│   ├── projects.gws       # List of repositories
│   ├── workspaces.gws     # Nested workspaces (optional)
│   └── hooks/             # Local hooks (optional)
├── project-a/
├── project-b/
├── team-x/                # Nested workspace
│   ├── .gws/
│   │   └── projects.gws
│   ├── service-1/
│   └── service-2/
└── .gitignore             # Auto-generated
```

## Projects File

The `.gws/projects.gws` file defines your repositories:

```bash
# path | remote-url [remote-name] | remote-url2 [remote-name2]
api              | git@github.com:company/api.git
frontend         | git@github.com:company/frontend.git
shared/lib       | git@github.com:company/lib.git origin | git@github.com:upstream/lib.git upstream
```

## Documentation

- [Getting Started](docs/getting-started.md) — Installation and first workspace
- [CLI Reference](docs/cli.md) — All commands and flags
- [Configuration](docs/configuration.md) — Config files, env vars, themes
- [Hooks](docs/hooks.md) — Git-style hooks with trust system
- [Nested Workspaces](docs/workspaces.md) — Hierarchical workspace organization
- [Examples](docs/examples.md) — Workflows, scripts, CI/CD integration
- [Development](docs/development.md) — Contributing and architecture

## Comparison

| Feature | gogws | gws | mu-repo |
|---------|:-----:|:---:|:-------:|
| Language | Go | Bash | Python |
| Parallel execution | ✅ | ❌ | ✅ |
| Nested workspaces | ✅ | ❌ | ❌ |
| Git-style hooks | ✅ | ✅ | ❌ |
| Hook trust system | ✅ | ❌ | ❌ |
| JSON/YAML export | ✅ | ❌ | ❌ |
| Stop on error | ✅ | ❌ | ✅ |
| gws compatible | ✅ | — | ❌ |
| Single binary | ✅ | ❌ | ❌ |

## License

[AGPL-3.0](LICENSE)

## Credits

Inspired by [gws](https://github.com/StreakyCobra/gws) by StreakyCobra and [mu-repo](https://github.com/fabioz/mu-repo) by Fabio Zadrozny.

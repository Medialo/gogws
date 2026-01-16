# Getting Started

This guide will help you install gogws and set up your first workspace in under 5 minutes.

## Prerequisites

- **Go 1.21+** (for building from source)
- **Git 2.x+**

## Installation

### From Source

```bash
go install github.com/medialo/gogws/cmd/gogws@latest
```

### Build Locally

```bash
git clone https://github.com/medialo/gogws.git
cd gogws
go build -o gogws ./cmd/gogws

# Or with Make
make build
```

### Verify Installation

```bash
gogws version
```

## Your First Workspace

### Option A: From Existing Repositories

If you already have repositories cloned in a directory:

```
~/projects/
├── api/
├── frontend/
├── mobile-app/
└── shared-lib/
```

Initialize gogws:

```bash
cd ~/projects
gogws init
```

This will:
1. Scan for all git repositories
2. Create `.gws/projects.gws` with discovered repos
3. Generate a `.gitignore` for the workspace

Output:

```
Scanning for git repositories...
Found 4 repositories

Created .gws/projects.gws:
  api              | git@github.com:company/api.git
  frontend         | git@github.com:company/frontend.git
  mobile-app       | git@github.com:company/mobile-app.git
  shared-lib       | git@github.com:company/shared-lib.git

Created .gitignore with GWS rules
```

### Option B: From a Projects File

Create a new workspace from scratch:

```bash
mkdir ~/work
cd ~/work
mkdir -p .gws
```

Create `.gws/projects.gws`:

```bash
api              | git@github.com:company/api.git
frontend         | git@github.com:company/frontend.git
shared-lib       | git@github.com:company/shared-lib.git
```

Clone all repositories:

```bash
gogws update
```

Output:

```
Cloning 3 missing projects...

✓ Cloned: api, frontend, shared-lib
```

## Basic Workflow

```
  ┌──────────────────────────────────────────────────────────────────┐
  │                      Daily Workflow                               │
  ├──────────────────────────────────────────────────────────────────┤
  │                                                                   │
  │    ┌──────────┐                                                   │
  │    │  status  │  ←── Check current state                          │
  │    └────┬─────┘                                                   │
  │         │                                                         │
  │         ▼                                                         │
  │    ┌──────────┐                                                   │
  │    │  fetch   │  ←── Download new commits from remotes            │
  │    └────┬─────┘                                                   │
  │         │                                                         │
  │         ▼                                                         │
  │    ┌──────────┐                                                   │
  │    │    ff    │  ←── Fast-forward pull (safe, no merge commits)   │
  │    └────┬─────┘                                                   │
  │         │                                                         │
  │         ▼                                                         │
  │    ┌──────────┐                                                   │
  │    │  update  │  ←── Clone any new repositories                   │
  │    └──────────┘                                                   │
  │                                                                   │
  └──────────────────────────────────────────────────────────────────┘
```

### 1. Check Status

```bash
gogws status
```

Shows all repositories with:
- Current branch
- Uncommitted changes
- Untracked files
- Ahead/behind remote

### 2. Fetch Updates

```bash
gogws fetch
```

Downloads new commits from all remotes without modifying your working directories.

### 3. Fast-Forward Pull

```bash
gogws ff
```

Pulls changes only when fast-forward is possible (no merge commits created).

### 4. Clone Missing Repositories

```bash
gogws update
```

Clones any repositories defined in `projects.gws` that don't exist locally.

## Workspace Structure

After initialization, your workspace looks like this:

```
my-workspace/
│
├── .gws/                      # Configuration directory
│   ├── projects.gws           # Repository definitions
│   ├── workspaces.gws         # Nested workspaces (optional)
│   └── hooks/                 # Local hooks (optional)
│       ├── pre-fetch
│       └── post-update
│
├── .gitignore                 # Auto-generated, ignores sub-repos
├── .ignore.gws                # Patterns to ignore (optional)
│
├── project-a/                 # Your repositories
├── project-b/
└── project-c/
```

## Common Options

### Parallel Execution

By default, gogws runs 5 operations in parallel:

```bash
# Use 10 workers
gogws fetch --parallel=10

# Run sequentially (useful for debugging)
gogws fetch --parallel=1
```

### Output Formats

```bash
# Default colored text
gogws status

# JSON (for scripting)
gogws status --format=json

# YAML
gogws status --format=yaml
```

### Filter Results

```bash
# Show only repositories with changes
gogws status --only-changes
```

## Next Steps

- [CLI Reference](cli.md) — All commands and flags
- [Configuration](configuration.md) — Config files and environment variables
- [Hooks](hooks.md) — Automate tasks with hooks
- [Nested Workspaces](workspaces.md) — Organize large projects
- [Examples](examples.md) — Real-world workflows

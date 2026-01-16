# CLI Reference

Complete reference for all gogws commands and flags.

## Synopsis

```
gogws [command] [flags]
gogws [command] [subcommand] [flags]
```

## Global Flags

These flags are available for all commands:

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--parallel` | int | 5 | Number of parallel workers (0=auto, 1=serial) |
| `--stop-on-error` | bool | false | Stop execution on first error |
| `--format` | string | text | Output format: `text`, `json`, `yaml` |
| `--no-color` | bool | false | Disable colored output |
| `--only-changes` | bool | false | Show only repositories with changes |
| `--trust-hooks` | string | ask | Hook trust mode: `ask`, `all`, `skip` |
| `--verbose`, `-v` | bool | false | Enable verbose output |
| `--config` | string | | Custom config file path |
| `--theme` | string | | Custom theme file path |
| `--help`, `-h` | bool | | Show help for command |

## Commands

### Workspace Management

#### `gogws init`

Initialize workspace configuration files.

```bash
gogws init [subcommand] [flags]
```

Running `gogws init` without a subcommand is equivalent to `gogws init projects`.

##### `gogws init projects`

Discover git repositories and create `.gws/projects.gws`.

```bash
gogws init projects [flags]
```

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--reset` | bool | false | Overwrite existing projects.gws file |
| `--gitignore` | bool | true | Also generate .gitignore file |

**Example:**

```bash
# Discover repos and create config
gogws init projects

# Overwrite existing config
gogws init projects --reset

# Skip gitignore generation
gogws init projects --gitignore=false
```

##### `gogws init workspaces`

Interactively configure nested workspaces.

```bash
gogws init workspaces
```

##### `gogws init gitignore`

Generate or update `.gitignore` with GWS-specific rules.

```bash
gogws init gitignore [flags]
```

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--force` | bool | false | Update even if GWS section already exists |
| `--remove` | bool | false | Remove GWS section from .gitignore |

**Example:**

```bash
# Add/update GWS section
gogws init gitignore

# Force update existing section
gogws init gitignore --force

# Remove GWS section
gogws init gitignore --remove
```

---

#### `gogws check`

Check workspace consistency. Reports repositories in four categories:

- **Known** — Defined in projects.gws and exists
- **Missing** — Defined in projects.gws but not cloned
- **Unknown** — Exists but not in projects.gws
- **Ignored** — Matches patterns in .ignore.gws

```bash
gogws check
```

**Example output:**

```
Known repositories (3):
  ✓ api
  ✓ frontend
  ✓ shared-lib

Missing repositories (1):
  ✗ new-service

Unknown repositories (2):
  ? temp-project
  ? experiments/test
```

---

### Repository Operations

#### `gogws status`

Show status of all repositories in the workspace.

```bash
gogws status [flags]
```

**Aliases:** `st`

**Output columns:**
- Repository path
- Current branch
- Sync status (ahead ↑ / behind ↓)
- Working tree status (uncommitted, untracked)

**Example:**

```bash
# Default text output
gogws status

# JSON output for scripting
gogws status --format=json

# Only show repos with changes
gogws status --only-changes
```

**JSON output structure:**

```json
[
  {
    "path": "api",
    "branch": "main",
    "exists": true,
    "clean": true,
    "ahead": 0,
    "behind": 2,
    "uncommitted": 0,
    "untracked": 0,
    "has_remote": true
  }
]
```

---

#### `gogws fetch`

Fetch updates from all remotes for all repositories.

```bash
gogws fetch [flags]
```

**Example:**

```bash
# Fetch with 10 parallel workers
gogws fetch --parallel=10

# Stop on first error
gogws fetch --stop-on-error
```

**Hooks:** `pre-fetch`, `post-fetch`

---

#### `gogws ff`

Fast-forward pull all repositories.

Only pulls when fast-forward is possible (no merge commits created). Repositories with local commits ahead of remote are skipped.

```bash
gogws ff [flags]
```

**Example:**

```bash
gogws ff
```

**Hooks:** `pre-ff`, `post-ff`

---

#### `gogws update`

Clone all missing repositories and workspaces.

```bash
gogws update [flags]
```

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--skip-projects` | bool | false | Only clone workspaces, skip projects |
| `--skip-workspaces` | bool | false | Only clone projects, skip workspaces |

**Example:**

```bash
# Clone everything
gogws update

# Only clone nested workspaces
gogws update --skip-projects

# Only clone direct projects
gogws update --skip-workspaces
```

**Hooks:** `pre-update`, `post-update`

---

#### `gogws clone`

Clone one or more specific repositories by their path.

```bash
gogws clone <path>... [flags]
```

**Example:**

```bash
# Clone single repository
gogws clone api

# Clone multiple repositories
gogws clone api frontend shared-lib
```

**Hooks:** `pre-clone`, `post-clone`

---

### Configuration

#### `gogws config`

Manage user configuration stored in `~/.gws/config.yaml`.

```bash
gogws config [subcommand]
```

##### `gogws config list`

List all available configuration keys.

```bash
gogws config list
```

**Output:**

```
Available configuration keys:
  trusted-workspaces
```

##### `gogws config get`

Get a configuration value.

```bash
gogws config get <key>
```

**Example:**

```bash
gogws config get trusted-workspaces
# Output:
#   - /home/user/work/*
#   - /home/user/personal/**
```

##### `gogws config set`

Set a configuration value.

```bash
gogws config set <key> <value>
```

**Example:**

```bash
gogws config set trusted-workspaces /home/user/work/*
```

---

### Utilities

#### `gogws version`

Print version information.

```bash
gogws version
```

---

#### `gogws completion`

Generate shell completion script.

```bash
gogws completion <shell>
```

**Supported shells:** `bash`, `zsh`, `fish`, `powershell`

**Example:**

```bash
# Bash
gogws completion bash > /etc/bash_completion.d/gogws

# Zsh
gogws completion zsh > "${fpath[1]}/_gogws"

# Fish
gogws completion fish > ~/.config/fish/completions/gogws.fish

# PowerShell
gogws completion powershell > gogws.ps1
```

---

## Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Success |
| 1 | General error or command failed |
| 2 | No workspace found (missing .projects.gws) |

## Environment Variables

| Variable | Description |
|----------|-------------|
| `GOGWS_PARALLEL` | Default parallel workers |
| `GOGWS_FORMAT` | Default output format |
| `NO_COLOR` | Disable colored output (any value) |

## Execution Modes

### Parallel Mode (default)

Operations run concurrently with configurable number of workers:

```bash
gogws fetch --parallel=10
```

### Serial Mode

Operations run one at a time. Useful for debugging:

```bash
gogws fetch --parallel=1
```

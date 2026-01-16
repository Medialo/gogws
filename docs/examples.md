# Examples

Practical examples and workflows for common gogws use cases.

## Daily Workflows

### Morning Sync

Start your day by syncing all repositories:

```bash
cd ~/work

# Check current state
gogws status --only-changes

# Fetch all updates
gogws fetch

# Pull where possible (fast-forward only)
gogws ff

# Check what's new
gogws status
```

## Team Onboarding

### New Team Member Setup

```bash
# Clone the workspace repository
git clone git@github.com:company/workspace.git ~/work
cd ~/work

# Clone all 50+ repositories (parallel by default)
gogws update

# Verify everything is cloned
gogws check
```

### Selective Clone

Clone only specific repositories:

```bash
# Clone just what you need
gogws clone api frontend shared-lib

# Or clone workspaces without projects, then select
gogws update --skip-projects
gogws clone team-backend/api team-backend/worker
```

## Multi-Remote Workflows

### Fork Workflow

Working with a fork and upstream:

**.gws/projects.gws:**
```bash
my-fork | git@github.com:me/project.git origin | git@github.com:upstream/project.git upstream
```

**Daily workflow:**
```bash
# Fetch from both remotes
gogws fetch

# Status shows ahead/behind for origin
gogws status

# To sync with upstream (manual per-repo):
cd my-fork
git fetch upstream
git merge upstream/main
```

### Multiple Remotes for Backup

**.gws/projects.gws:**
```bash
important-project | git@github.com:company/project.git origin | git@backup.company.com:project.git backup
```

## Scripting with JSON Output

### List Dirty Repositories

```bash
gogws status --format=json | jq -r '.[] | select(.clean == false) | .path'
```

### Count Repos Ahead of Remote

```bash
gogws status --format=json | jq '[.[] | select(.ahead > 0)] | length'
```

### Export to CSV

```bash
gogws status --format=json | jq -r '
  ["path","branch","ahead","behind","uncommitted","untracked"],
  (.[] | [.path, .branch, .ahead, .behind, .uncommitted, .untracked])
  | @csv
' > status.csv
```

### Find Repos on Specific Branch

```bash
gogws status --format=json | jq -r '.[] | select(.branch == "develop") | .path'
```

## Performance Optimization

### Large Workspaces

For workspaces with 100+ repositories:

```bash
# Increase parallelism
gogws fetch --parallel=20

# Use stop-on-error for faster failure detection
gogws update --parallel=20 --stop-on-error
```

### Slow Networks

For slow or unreliable connections:

```bash
# Reduce parallelism to avoid timeouts
gogws fetch --parallel=3

# Or run serially for debugging
gogws fetch --parallel=1 --verbose
```

## Troubleshooting Workflows

### Debug Mode

```bash
# Maximum verbosity
gogws fetch --parallel=1 --verbose
```
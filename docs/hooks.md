# Hooks

Hooks allow you to run custom scripts at various points during gogws command execution. They follow a git-like convention with executable files discovered automatically.

## Hook Locations

Hooks can be defined at two levels:

### Global Hooks (User-level)

Located in `~/.gws/hooks/`. These hooks are **always trusted** and run automatically.

```
~/.gws/
└── hooks/
    ├── pre-init
    ├── post-init
    ├── pre-update
    └── ...
```

### Local Hooks (Workspace-level)

Located in `<workspace>/.gws/hooks/`. These hooks require trust verification before execution.

```
<workspace>/
└── .gws/
    └── hooks/
        ├── pre-init
        ├── post-init
        └── ...
```

## Hook Priority

Local hooks **override** global hooks. If a hook exists in both locations, only the local hook will be executed.

## Available Hooks

| Hook | Trigger | Description |
|------|---------|-------------|
| `pre-init` | Before init command | Before workspace initialization |
| `post-init` | After init command | After workspace initialization |
| `pre-update` | Before update command | Before cloning missing repos |
| `post-update` | After update command | After cloning missing repos |
| `pre-clone` | Before cloning a repository | Before each individual clone |
| `post-clone` | After cloning a repository | After each individual clone |
| `pre-fetch` | Before fetch command | Before fetching all repos |
| `post-fetch` | After fetch command | After fetching all repos |
| `pre-ff` | Before fast-forward pull | Before pulling all repos |
| `post-ff` | After fast-forward pull | After pulling all repos |
| `pre-check` | Before check command | Before workspace check |
| `post-check` | After check command | After workspace check |

## Trust System

Local hooks from external/cloned workspaces are not automatically trusted for security reasons.

### Trust Mode Flag

Use the `--trust-hooks` flag to control behavior:

| Mode | Description |
|------|-------------|
| `ask` (default) | Prompt for each untrusted hook with options to run, skip, or trust |
| `all` | Run all hooks without prompting |
| `skip` | Skip all untrusted local hooks |

Example:
```bash
gogws update --trust-hooks=skip
gogws fetch --trust-hooks=all
```

### Trusted Workspaces Configuration

Add workspace paths to your trusted list in `~/.gws/config.yaml`:

```yaml
use-agent: true
trusted-workspaces:
  - "/home/user/work/*"           # All direct children of /home/user/work
  - "/home/user/personal/**"      # All descendants of /home/user/personal
  - "/home/user/myproject"        # Specific workspace path
```

Wildcard patterns:
- `*` - matches any single directory level
- `**` - matches any number of directory levels (recursive)

### Interactive Trust Prompt

When `--trust-hooks=ask` (default) and a local hook is found in an untrusted workspace:

```
[hook:local] Hook 'pre-update' found at: /path/to/workspace/.gws/hooks/pre-update
Workspace: /path/to/workspace
This workspace is not in your trusted list.

Options:
  [r] Run this hook
  [s] Skip this hook
  [t] Run and add workspace to trusted list
Choose [r/s/t]:
```

## Environment Variables

All hooks receive these environment variables:

| Variable | Description |
|----------|-------------|
| `GOGWS_COMMAND` | The command being executed (`init`, `update`, `clone`, `fetch`, `ff`, `check`) |
| `GOGWS_WORKSPACE` | The absolute path to the workspace root directory |
| `GOGWS_HOOK_NAME` | The name of the hook being executed |
| `GOGWS_HOOK_ORIGIN` | The origin of the hook (`global` or `local`) |

## Execution Behavior

- Hooks run **synchronously** - the command waits for the hook to complete
- Hooks run in the **workspace root directory**
- Hooks inherit the **current environment** plus gogws-specific variables
- Hook **stdout/stderr** are passed through to the terminal
- If a hook **exits with non-zero**, the parent command fails with an error
- If a hook file **doesn't exist**, it is silently skipped
- Hook origin is displayed: `[hook:global]` or `[hook:local]` or `[hook:local:trusted]`

## Creating Hooks

1. Create the hooks directory:
```bash
# Global hooks
mkdir -p ~/.gws/hooks

# Local hooks (workspace-level)
mkdir -p .gws/hooks
```

2. Create your hook script (no extension needed, git-style):
```bash
#!/bin/bash

echo "Running $GOGWS_COMMAND in $GOGWS_WORKSPACE"
echo "Hook: $GOGWS_HOOK_NAME (origin: $GOGWS_HOOK_ORIGIN)"

# Example: Send notification after update
if [ "$GOGWS_COMMAND" = "update" ]; then
    echo "Workspace updated!"
fi
```

3. Make it executable:
```bash
chmod +x ~/.gws/hooks/post-update
# or
chmod +x .gws/hooks/post-update
```

## Configuration Files Location

Configuration files can now be placed in the `.gws/` directory:

```
<workspace>/
├── .gws/
│   ├── hooks/           # Hook scripts
│   ├── projects.gws     # Projects definition (preferred)
│   └── workspaces.gws   # Workspaces definition (preferred)
├── .projects.gws        # Legacy location (warning if both exist)
└── .workspaces.gws      # Legacy location (warning if both exist)
```

If both `.gws/projects.gws` and `.projects.gws` exist, the `.gws/` version takes priority and a warning is displayed asking to remove the legacy file.

## Examples

### Post-Update: Slack Notification

Send a Slack notification when repositories are cloned:

```bash
#!/bin/bash
# ~/.gws/hooks/post-update

if [ -n "$SLACK_WEBHOOK" ]; then
  curl -s -X POST "$SLACK_WEBHOOK" \
    -H 'Content-type: application/json' \
    -d "{
      \"text\": \"Workspace updated\",
      \"blocks\": [
        {
          \"type\": \"section\",
          \"text\": {
            \"type\": \"mrkdwn\",
            \"text\": \"*Workspace updated*\n$GOGWS_WORKSPACE\"
          }
        }
      ]
    }"
fi
```

### Pre-Fetch: VPN Check

Ensure VPN is connected before fetching from private repositories:

```bash
#!/bin/bash
# .gws/hooks/pre-fetch

INTERNAL_HOST="git.internal.company.com"

if ! ping -c1 -W2 "$INTERNAL_HOST" &>/dev/null; then
  echo "ERROR: Cannot reach $INTERNAL_HOST"
  echo "Please connect to VPN before fetching."
  exit 1
fi

echo "VPN connection verified"
```

### Post-Clone: Auto-Install Dependencies

Automatically install dependencies for cloned repositories:

```bash
#!/bin/bash
# .gws/hooks/post-clone

echo "Installing dependencies for cloned repositories..."

for dir in */; do
  [ -d "$dir/.git" ] || continue
  
  if [ -f "$dir/package.json" ]; then
    echo "→ npm install in $dir"
    (cd "$dir" && npm ci --silent)
  elif [ -f "$dir/go.mod" ]; then
    echo "→ go mod download in $dir"
    (cd "$dir" && go mod download)
  elif [ -f "$dir/requirements.txt" ]; then
    echo "→ pip install in $dir"
    (cd "$dir" && pip install -q -r requirements.txt)
  elif [ -f "$dir/Gemfile" ]; then
    echo "→ bundle install in $dir"
    (cd "$dir" && bundle install --quiet)
  fi
done

echo "Dependencies installed"
```

### Post-Update: Run Setup Scripts

Run project-specific setup scripts after cloning:

```bash
#!/bin/bash
# .gws/hooks/post-update

for dir in */; do
  [ -d "$dir/.git" ] || continue
  
  if [ -x "$dir/scripts/setup.sh" ]; then
    echo "Running setup for $dir"
    (cd "$dir" && ./scripts/setup.sh)
  fi
done
```

### Pre-FF: Check for Uncommitted Changes

Warn before pulling if there are uncommitted changes:

```bash
#!/bin/bash
# ~/.gws/hooks/pre-ff

DIRTY_REPOS=""

for dir in */; do
  [ -d "$dir/.git" ] || continue
  
  if ! git -C "$dir" diff --quiet 2>/dev/null; then
    DIRTY_REPOS="$DIRTY_REPOS  - $dir\n"
  fi
done

if [ -n "$DIRTY_REPOS" ]; then
  echo "WARNING: The following repositories have uncommitted changes:"
  echo -e "$DIRTY_REPOS"
  echo "These will NOT be pulled to avoid conflicts."
fi
```

## Troubleshooting

### Hook Not Executing

1. **Check file is executable:**
   ```bash
   ls -la .gws/hooks/
   chmod +x .gws/hooks/your-hook
   ```

2. **Check hook name matches exactly:**
   Hook names must match exactly (no extension):
   - ✅ `pre-fetch`
   - ❌ `pre-fetch.sh`
   - ❌ `pre_fetch`

3. **Run with verbose mode:**
   ```bash
   gogws fetch --verbose
   ```
   This shows hook discovery and execution.

4. **Check shebang line:**
   Ensure the first line is a valid interpreter:
   ```bash
   #!/bin/bash
   # or
   #!/usr/bin/env bash
   ```

### Permission Denied

```
Error: permission denied: .gws/hooks/pre-fetch
```

**Solution:**
```bash
chmod +x .gws/hooks/pre-fetch
```

On Windows, ensure the file has execute permissions or use a `.bat`/`.ps1` extension.

### Trust Issues

**"Workspace is not in your trusted list"**

Options:
1. **One-time run:** Choose `[r]` at the prompt
2. **Permanently trust:** Choose `[t]` to add to trusted list
3. **Skip hooks:** Use `--trust-hooks=skip`
4. **Add to config:** Edit `~/.gws/config.yaml`:
   ```yaml
   trusted-workspaces:
     - "/path/to/workspace"
   ```

**Bypass for CI/CD:**
```bash
gogws update --trust-hooks=all
```

### Hook Fails with Exit Code

If a hook exits with non-zero, the parent command fails:

```
[hook:local] Running pre-fetch...
ERROR: VPN not connected
Error: pre-fetch hook failed: exit status 1
```

To continue despite hook failures, modify your hook to return 0:

```bash
#!/bin/bash
# Non-blocking hook
do_something || echo "Warning: do_something failed"
exit 0  # Always succeed
```

### Debugging Hooks

Add debug output to your hook:

```bash
#!/bin/bash
set -x  # Print commands as they execute

echo "GOGWS_COMMAND: $GOGWS_COMMAND"
echo "GOGWS_WORKSPACE: $GOGWS_WORKSPACE"
echo "GOGWS_HOOK_NAME: $GOGWS_HOOK_NAME"
echo "GOGWS_HOOK_ORIGIN: $GOGWS_HOOK_ORIGIN"
echo "PWD: $(pwd)"

# Your hook logic here
```

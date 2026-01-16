# gogws - Git Workspace Manager

A modern Git workspace management tool written in Go, inspired by [gws](https://github.com/StreakyCobra/gws).

## Features

- 🎨 Beautiful CLI output with colors and formatting (Lipgloss)
- 🖥️ Interactive TUI mode (Bubble Tea) 
- ⚡ Parallel operations for speed
- 🔧 Configurable via files, environment variables, and flags (Viper)
- 🎯 Compatible with gws `.projects.gws` format
- 🪝 Hooks support for custom workflows
- 📊 Multiple output formats (text, JSON, YAML)
- 🎨 Customizable themes
- 🌍 Cross-platform (Windows, Linux, macOS)
- 🔐 Flexible authentication (SSH keys, tokens, no agent required)

## Installation

### From source

```bash
go install github.com/yourusername/gogws/cmd/gogws@latest
```

### Build locally

```bash
make build
```

### Build for all platforms

```bash
make build-all
```

## Quick Start

1. Initialize a workspace by discovering existing repositories:

```bash
cd ~/projects
gogws init
```

2. Or create a `.projects.gws` file manually:

```
work/project1 | git@github.com:user/project1.git
work/project2 | git@github.com:user/project2.git origin | git@github.com:upstream/project2.git upstream
personal/myapp | https://github.com/user/myapp.git
```

3. Clone all missing repositories:

```bash
gogws update
```

4. Check status of all repositories:

```bash
gogws status
# or simply
gogws
```

## Commands

- `gogws init` - Discover existing repositories and create `.projects.gws`
- `gogws status` - Show status of all repositories (default command)
- `gogws update` - Clone all missing repositories
- `gogws clone <repo>` - Clone specific repository
- `gogws fetch` - Fetch updates from origin for all repos
- `gogws ff` - Fast-forward pull all repositories
- `gogws check` - Check workspace consistency
- `gogws version` - Show version

## Global Flags

- `--config <file>` - Custom config file
- `--theme <file>` - Custom theme file
- `--parallel <n>` - Number of parallel operations (default: 5)
- `--format <text|json|yaml>` - Output format
- `--no-color` - Disable colored output
- `--only-changes` - Show only repositories with changes
- `--tui` - Use interactive TUI mode

## Authentication

gogws supports multiple authentication methods and works without SSH agents:

### HTTPS with Token (Recommended)

```bash
export GIT_TOKEN="ghp_your_github_token"
gogws update
```

### HTTPS with Username/Password

```bash
export GIT_USERNAME="your_username"
export GIT_PASSWORD="your_password"
gogws update
```

### SSH with Key File

```bash
export SSH_KEY_PATH="~/.ssh/id_rsa"
gogws update
```

If not specified, gogws will automatically look for `~/.ssh/id_rsa`.

### No Authentication

For public repositories, no authentication is needed:

```bash
gogws update  # Works directly for public repos
```

**Note**: gogws does NOT require SSH agents (Pageant, ssh-agent) to be running. It can work with SSH keys directly or use HTTPS authentication.

## Configuration

### .projects.gws

Compatible with gws format:

```
path/to/repo | url1 [remote1] | url2 [remote2]
```

### .ignore.gws

Regular expressions to ignore specific projects:

```
^work/
-test$
```

### ~/.config/gogws/config.yaml

```yaml
theme: ~/.config/gogws/theme.yaml
parallel: 10
hooks:
  pre_update: ./scripts/pre-update.sh
  post_update: ./scripts/post-update.sh
```

### Environment Variables

All config options can be set via `GOGWS_*` environment variables:

```bash
export GOGWS_PARALLEL=10
export GOGWS_FORMAT=json
```

## Export Formats

Export status to JSON or YAML for scripting:

```bash
# JSON export
gogws status --format json > status.json

# YAML export  
gogws status --format yaml > status.yaml

# Use with jq
gogws status --format json | jq '.repositories[] | select(.clean == false)'
```

## Hooks

Configure hooks in `~/.config/gogws/config.yaml`:

```yaml
hooks:
  pre_init: ./scripts/pre-init.sh
  post_init: ./scripts/post-init.sh
  pre_update: ./scripts/pre-update.sh
  post_update: ./scripts/post-update.sh
  pre_clone: ./scripts/pre-clone.sh
  post_clone: ./scripts/post-clone.sh
  pre_fetch: ./scripts/pre-fetch.sh
  post_fetch: ./scripts/post-fetch.sh
  pre_ff: ./scripts/pre-ff.sh
  post_ff: ./scripts/post-ff.sh
  pre_check: ./scripts/pre-check.sh
  post_check: ./scripts/post-check.sh
```

Hooks receive environment variables:
- `GOGWS_COMMAND` - Command being executed
- `GOGWS_WORKSPACE` - Workspace root path

## Custom Themes

Create a custom theme file (YAML):

```yaml
colors:
  title: "63"
  success: "46"
  warning: "214"
  error: "196"
  info: "39"
  subtle: "241"
  path: "33"
  branch: "141"
  remote: "220"
  status: "46"
  stats: "250"
styles:
  title_bold: true
  path_bold: true
```

Use with `--theme` flag:

```bash
gogws status --theme ~/.config/gogws/mytheme.yaml
```

## Troubleshooting

### "SSH agent requested, but could not detect Pageant or Windows native SSH agent"

**Solution**: Use one of these authentication methods:

1. **HTTPS with token** (recommended):
   ```bash
   export GIT_TOKEN="your_github_token"
   gogws update
   ```

2. **Specify SSH key directly**:
   ```bash
   export SSH_KEY_PATH="~/.ssh/id_ed25519"
   gogws update
   ```

3. **Use HTTPS URLs** instead of SSH URLs in `.projects.gws`:
   ```
   # Instead of:
   repo | git@github.com:user/repo.git
   
   # Use:
   repo | https://github.com/user/repo.git
   ```

### "authentication required"

Make sure you have valid credentials set via environment variables or switch to public repository URLs.

### "no workspace found"

Make sure you have a `.projects.gws` file. Run `gogws init` to create one.

## Examples

See [EXAMPLES.md](EXAMPLES.md) for more usage examples.

## Documentation

- [README.md](README.md) - This file
- [QUICKSTART.md](QUICKSTART.md) - 5-minute tutorial
- [EXAMPLES.md](EXAMPLES.md) - Practical usage examples
- [DEVELOPMENT.md](DEVELOPMENT.md) - Developer documentation
- [CHANGELOG.md](CHANGELOG.md) - Version history

## License

MIT

## Credits

Inspired by [gws](https://github.com/StreakyCobra/gws) by StreakyCobra.

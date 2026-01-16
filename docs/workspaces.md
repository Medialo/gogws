# Nested Workspaces

gogws supports hierarchical workspace organization through nested workspaces. A workspace can contain other workspaces, each with their own projects and configuration.

## Concept

```
┌─────────────────────────────────────────────────────────────────────┐
│  company-workspace/                                                  │
│  ├── .gws/                                                          │
│  │   ├── projects.gws      ← Direct projects                        │
│  │   └── workspaces.gws    ← Nested workspace definitions           │
│  │                                                                   │
│  ├── shared-libs/          ← Direct project                         │
│  ├── docs/                 ← Direct project                         │
│  │                                                                   │
│  ├── team-backend/         ← Nested workspace                       │
│  │   ├── .gws/                                                      │
│  │   │   └── projects.gws                                           │
│  │   ├── api/                                                       │
│  │   ├── worker/                                                    │
│  │   └── migrations/                                                │
│  │                                                                   │
│  └── team-frontend/        ← Nested workspace                       │
│      ├── .gws/                                                      │
│      │   └── projects.gws                                           │
│      ├── web-app/                                                   │
│      ├── mobile-app/                                                │
│      └── design-system/                                             │
└─────────────────────────────────────────────────────────────────────┘
```

## Use Cases

### Multi-Team Organization

Each team manages their own workspace with their projects:

```
company/
├── team-platform/     # Platform team's repos
├── team-mobile/       # Mobile team's repos
├── team-data/         # Data team's repos
└── shared/            # Shared libraries (direct project)
```

### Domain-Driven Organization

Organize by business domain:

```
ecommerce/
├── domain-catalog/    # Product catalog services
├── domain-orders/     # Order management
├── domain-payments/   # Payment processing
└── domain-shipping/   # Shipping & logistics
```

### Monorepo with Satellites

Main monorepo with satellite repositories:

```
main-workspace/
├── monorepo/          # Main monorepo (direct project)
├── tools/             # Workspace: development tools
├── infrastructure/    # Workspace: IaC repos
└── docs/              # Workspace: documentation repos
```

## Configuration

### Workspaces File

Create `.gws/workspaces.gws` to define nested workspaces:

```bash
# path | remote-url [remote-name]
team-backend  | git@github.com:company/backend-workspace.git
team-frontend | git@github.com:company/frontend-workspace.git origin
team-mobile   | git@github.com:company/mobile-workspace.git
```

### Complete Example

**Parent workspace:** `company/.gws/projects.gws`
```bash
# Shared projects at company level
shared-libs   | git@github.com:company/shared-libs.git
docs          | git@github.com:company/docs.git
```

**Parent workspace:** `company/.gws/workspaces.gws`
```bash
# Nested workspaces
team-backend  | git@github.com:company/backend-workspace.git
team-frontend | git@github.com:company/frontend-workspace.git
```

**Nested workspace:** `company/team-backend/.gws/projects.gws`
```bash
# Backend team projects
api           | git@github.com:company/api.git
worker        | git@github.com:company/worker.git
migrations    | git@github.com:company/migrations.git
```

## Commands

### Clone Everything

Clone all workspaces and all their projects recursively:

```bash
gogws update
```

```
Cloning 2 missing workspaces...
✓ Cloned: team-backend, team-frontend

Cloning 3 missing projects in team-backend...
✓ Cloned: api, worker, migrations

Cloning 3 missing projects in team-frontend...
✓ Cloned: web-app, mobile-app, design-system

Cloning 2 missing projects...
✓ Cloned: shared-libs, docs
```

### Clone Only Workspaces

Clone workspace repositories without their projects:

```bash
gogws update --skip-projects
```

Useful for:
- Quick workspace structure setup
- When you only need specific projects later

### Clone Only Direct Projects

Clone direct projects, skip nested workspaces:

```bash
gogws update --skip-workspaces
```

Useful for:
- Working only on shared/root-level projects
- Avoiding large nested workspaces

### Status with Workspaces

Status shows workspace summary:

```bash
gogws status
```

```
╭──────────────────────────────────────────────╮
│  GOGWS - Workspace Status - company          │
╰──────────────────────────────────────────────╯

  Workspaces
  ◈ team-backend (3 projects)
  ◈ team-frontend (3 projects)

  Projects
  ● shared-libs    main     ✓
  ● docs           main     ↑2

┌──────────────────────────────────────────────────┐
│ Projects: 2  │  Workspaces: 2 (6 sub-projects)   │
└──────────────────────────────────────────────────┘
```

## Gitignore Management

### Auto-Generated Rules

When you initialize a workspace with nested workspaces, gogws generates appropriate `.gitignore` rules:

```gitignore
# === GWS START ===
*
!.gitignore
!README.md
!.gws/
!.gws/**
!*.gws
# === GWS END ===
```

This ensures:
- Sub-projects and workspaces are ignored (not tracked)
- Configuration files are tracked
- README is tracked

### Custom Template

Create `~/.gws/templates/gitignore.tmpl` for custom rules:

```
*
!.gitignore
!README.md
!{{.ConfigDir}}/
!{{.ConfigDir}}/**

# Also track CI config
!.github/
!.github/**

# Track documentation
!docs/
!docs/*.md
```

## Workspace Discovery

### How gogws Finds Workspaces

1. Read `.gws/workspaces.gws` (or `.workspaces.gws`)
2. For each entry, check if directory exists
3. If exists, load that workspace's configuration
4. Recursively discover nested workspaces

### Workspace Status

```bash
gogws check
```

Shows workspaces in categories:
- **Known** — Defined and cloned
- **Missing** — Defined but not cloned
- **Unknown** — Directory exists but not defined

## Best Practices

### 1. Keep Workspaces Focused

Each workspace should represent a logical grouping:
- One team
- One domain
- One product

❌ Don't create workspaces just for organization:
```
workspace/
├── a-projects/
├── b-projects/
└── c-projects/
```

✅ Use meaningful boundaries:
```
workspace/
├── team-auth/
├── team-billing/
└── shared-libs/
```

### 2. Consistent Structure

Use the same `.gws/` layout in all workspaces:

```
any-workspace/
├── .gws/
│   ├── projects.gws
│   └── hooks/         # Optional
└── ...
```

### 3. Version Workspace Configuration

Track `.gws/` directory in git:

```bash
cd my-workspace
git init
git add .gws/
git commit -m "Initialize workspace configuration"
```

### 4. Trust Carefully

Only trust workspaces you control. For shared/external workspaces:

```bash
# Review hooks before trusting
cat other-workspace/.gws/hooks/*

# Run without trust for one-time use
gogws update --trust-hooks=skip
```

### 5. Use Ignore Patterns

Exclude workspaces from commands with `.ignore.gws`:

```
# Don't include archived workspace in operations
^archived-projects/
```

## Troubleshooting

### Workspace Not Detected

**Symptoms:** Nested workspace not showing in status

**Check:**
1. Entry exists in `.gws/workspaces.gws`
2. Remote URL is correct
3. Directory name matches path in config

### Recursive Clone Fails

**Symptoms:** Parent clones but children don't

**Solution:**
```bash
# Clone workspaces first
gogws update --skip-projects

# Then clone projects
gogws update --skip-workspaces
```

### Circular Dependencies

**Symptoms:** gogws hangs or errors on circular workspace references

**Solution:** Avoid circular workspace definitions. Workspace A should not contain B if B contains A.

```
# ❌ Bad: circular
workspace-a/
└── workspace-b/
    └── workspace-a/  # Circular!

# ✅ Good: flat or tree
parent/
├── workspace-a/
└── workspace-b/
```

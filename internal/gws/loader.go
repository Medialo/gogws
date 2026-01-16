package gws

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
)

type Loader struct {
	root      string
	recursive bool
	maxDepth  int
	visited   map[string]bool
}

func New(path string) *Loader {
	return &Loader{
		root:      path,
		recursive: true,
		maxDepth:  DefaultMaxDepth,
		visited:   make(map[string]bool),
	}
}

func (l *Loader) Recursive(enabled bool) *Loader {
	l.recursive = enabled
	return l
}

func (l *Loader) MaxDepth(depth int) *Loader {
	l.maxDepth = depth
	return l
}

func (l *Loader) Load() (*Workspace, error) {
	return l.loadRecursive(l.root, 0)
}

func (l *Loader) loadRecursive(root string, depth int) (*Workspace, error) {
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return nil, err
	}

	if l.visited[absRoot] {
		slog.Debug("Skipping already visited workspace", "path", absRoot)
		return nil, nil
	}
	l.visited[absRoot] = true

	if depth > l.maxDepth {
		slog.Warn("Maximum workspace depth reached", "path", root)
		return nil, nil
	}

	slog.Debug("Loading workspace", "depth", depth, "path", root)

	ws := &Workspace{
		Root:     absRoot,
		Path:     root,
		Name:     filepath.Base(absRoot),
		Exists:   true,
		Projects: []Project{},
		Children: []*Workspace{},
	}

	projectsLocation := getProjectsFileLocation(absRoot)
	if projectsLocation != nil {
		if projectsLocation.HasDuplicate {
			legacyPath := filepath.Join(absRoot, ProjectsFileName)
			slog.Warn("Duplicate projects file found - using .gws/projects.gws, please remove the legacy file",
				"legacy", legacyPath,
				"used", projectsLocation.Path)
		}

		projects, err := parseProjectsFile(absRoot)
		if err != nil {
			slog.Warn("Failed to read projects", "path", root, "err", err)
		} else {
			for _, p := range projects {
				project := Project{
					Path:    p.Path,
					Remotes: p.Remotes,
					Exists:  false,
				}

				projectPath := filepath.Join(absRoot, p.Path)
				if _, err := os.Stat(projectPath); err == nil {
					project.Exists = true
				}

				ws.Projects = append(ws.Projects, project)
			}
		}
	}

	workspacesLocation := getWorkspacesFileLocation(absRoot)
	if workspacesLocation != nil {
		if workspacesLocation.HasDuplicate {
			legacyPath := filepath.Join(absRoot, WorkspacesFileName)
			slog.Warn("Duplicate workspaces file found - using .gws/workspaces.gws, please remove the legacy file",
				"legacy", legacyPath,
				"used", workspacesLocation.Path)
		}

		childRefs, err := parseWorkspacesFile(absRoot)
		if err != nil {
			slog.Warn("Failed to read workspaces", "path", root, "err", err)
		} else {
			for _, childRef := range childRefs {
				child := &Workspace{
					Path:     childRef.Path,
					Name:     childRef.Name,
					Remote:   childRef.Remote,
					Exists:   false,
					Projects: []Project{},
					Children: []*Workspace{},
				}

				wsPath := filepath.Join(absRoot, childRef.Path)
				if _, err := os.Stat(wsPath); err == nil {
					child.Exists = true

					if l.recursive {
						resolved, err := l.loadRecursive(wsPath, depth+1)
						if err != nil {
							child.Error = err
						} else if resolved != nil {
							child.Root = resolved.Root
							child.Projects = resolved.Projects
							child.Children = resolved.Children
						}
					}
				}

				ws.Children = append(ws.Children, child)
			}
		}
	}

	slog.Debug(" > Loaded workspace", "path", root, "projects", len(ws.Projects), "children", len(ws.Children))
	return ws, nil
}

func FindRoot() (*Workspace, error) {
	dir, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	for {
		hasProjects := hasProjectsFile(dir) || hasProjectsFileInConfigDir(dir)
		hasWorkspaces := hasWorkspacesFile(dir) || hasWorkspacesFileInConfigDir(dir)

		if hasProjects || hasWorkspaces {
			return &Workspace{
				Root:   dir,
				Path:   dir,
				Name:   filepath.Base(dir),
				Exists: true,
			}, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return nil, fmt.Errorf("no workspace found (no %s or %s/%s file found in current or parent directories)",
				ProjectsFileName, ConfigDirName, "projects.gws")
		}
		dir = parent
	}
}

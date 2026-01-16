package gws

import (
	"fmt"
	"os"
	"path/filepath"

	"gogws/internal/log"
)

func FindWorkspaceRoot() (*Workspace, error) {
	dir, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	for {
		ws := &Workspace{
			Root:   dir,
			Exists: true,
			Name:   filepath.Base(dir),
		}

		projectsPath := filepath.Join(dir, ProjectsFileName)
		if _, err := os.Stat(projectsPath); err == nil {
			ws.HasProjectsFile = true
		}

		workspacesPath := filepath.Join(dir, WorkspacesFileName)
		if _, err := os.Stat(workspacesPath); err == nil {
			ws.HasWorkspacesFile = true
		}

		if ws.HasProjectsFile || ws.HasWorkspacesFile {
			return ws, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return nil, fmt.Errorf("no workspace found (no %s or %s file found in current or parent directories)",
				ProjectsFileName, WorkspacesFileName)
		}
		dir = parent
	}
}

type Resolver struct {
	visited  map[string]bool
	maxDepth int
}

func NewResolver() *Resolver {
	return &Resolver{
		visited:  make(map[string]bool),
		maxDepth: 100,
	}
}

func (r *Resolver) SetMaxDepth(depth int) {
	r.maxDepth = depth
}

func (r *Resolver) Resolve(root string) (*Workspace, error) {
	return r.resolveRecursive(root, 0)
}

func (r *Resolver) resolveRecursive(root string, depth int) (*Workspace, error) {
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return nil, err
	}

	if r.visited[absRoot] {
		log.Debug("Skipping already visited workspace", "path", absRoot)
		return nil, nil
	}
	r.visited[absRoot] = true

	if depth > r.maxDepth {
		log.Warn("Maximum workspace depth reached", "path", root)
		return nil, nil
	}

	log.Debug("Resolving workspace", "depth", depth, "path", root)

	parser := NewParser(root)

	ws := &Workspace{
		Name:       filepath.Base(root),
		Root:       absRoot,
		Exists:     true,
		Projects:   []Project{},
		Workspaces: []WorkspaceRef{},
		Children:   []*Workspace{},
	}

	ws.HasProjectsFile = parser.HasProjectsFile()
	ws.HasWorkspacesFile = parser.HasWorkspacesFile()

	if ws.HasProjectsFile {
		projects, err := parser.ParseProjects()
		if err != nil {
			log.Warn("Failed to read projects", "path", root, "err", err)
		} else {
			for _, p := range projects {
				project := Project{
					Path:    p.Path,
					Remotes: p.Remotes,
					Exists:  false,
				}

				projectPath := filepath.Join(root, p.Path)
				if _, err := os.Stat(projectPath); err == nil {
					project.Exists = true
				}

				ws.Projects = append(ws.Projects, project)
			}
		}
	}

	if ws.HasWorkspacesFile {
		workspaces, err := parser.ParseWorkspaces()
		if err != nil {
			log.Warn("Failed to read workspaces", "path", root, "err", err)
		} else {
			for _, wsRef := range workspaces {
				entry := WorkspaceRef{
					Path:   wsRef.Path,
					Remote: wsRef.Remote,
					Exists: false,
				}

				wsPath := filepath.Join(root, wsRef.Path)
				if _, err := os.Stat(wsPath); err == nil {
					entry.Exists = true

					childWs, err := r.resolveRecursive(wsPath, depth+1)
					if err != nil {
						entry.Error = err
					} else if childWs != nil {
						ws.Children = append(ws.Children, childWs)
						entry.ProjectCount = len(childWs.Projects)
						entry.HasChildren = len(childWs.Children) > 0
					}
				}

				ws.Workspaces = append(ws.Workspaces, entry)
			}
		}
	}

	return ws, nil
}

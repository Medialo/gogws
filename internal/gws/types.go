package gws

const (
	ProjectsFileName   = ".projects.gws"
	WorkspacesFileName = ".workspaces.gws"
	IgnoreFileName     = ".ignore.gws"
	DefaultParallel    = 5
)

type Remote struct {
	Name string
	URL  string
}

type Project struct {
	Path    string
	Remotes []Remote
	Exists  bool
}

type WorkspaceRef struct {
	Path         string
	Remote       Remote
	Exists       bool
	ProjectCount int
	HasChildren  bool
	Error        error
}

type Workspace struct {
	Name              string
	Root              string
	HasProjectsFile   bool
	HasWorkspacesFile bool
	Exists            bool

	Projects   []Project
	Workspaces []WorkspaceRef

	Children []*Workspace
}

func (w *Workspace) AllProjects() []Project {
	all := make([]Project, 0, len(w.Projects))
	all = append(all, w.Projects...)
	for _, child := range w.Children {
		all = append(all, child.AllProjects()...)
	}
	return all
}

func (w *Workspace) AllWorkspaces() []WorkspaceRef {
	all := make([]WorkspaceRef, 0, len(w.Workspaces))
	all = append(all, w.Workspaces...)
	for _, child := range w.Children {
		all = append(all, child.AllWorkspaces()...)
	}
	return all
}

func (w *Workspace) TotalProjectCount() int {
	count := len(w.Projects)
	for _, child := range w.Children {
		count += child.TotalProjectCount()
	}
	return count
}

func (w *Workspace) TotalWorkspaceCount() int {
	count := len(w.Workspaces)
	for _, child := range w.Children {
		count += child.TotalWorkspaceCount()
	}
	return count
}

func (w *Workspace) MissingProjects() []Project {
	var missing []Project
	for _, p := range w.Projects {
		if !p.Exists {
			missing = append(missing, p)
		}
	}
	return missing
}

func (w *Workspace) MissingWorkspaces() []WorkspaceRef {
	var missing []WorkspaceRef
	for _, ws := range w.Workspaces {
		if !ws.Exists {
			missing = append(missing, ws)
		}
	}
	return missing
}

func (w *Workspace) AllMissingProjects() []Project {
	missing := w.MissingProjects()
	for _, child := range w.Children {
		missing = append(missing, child.AllMissingProjects()...)
	}
	return missing
}

func (w *Workspace) AllMissingWorkspaces() []WorkspaceRef {
	missing := w.MissingWorkspaces()
	for _, child := range w.Children {
		missing = append(missing, child.AllMissingWorkspaces()...)
	}
	return missing
}

func (ws *WorkspaceRef) ToProject() Project {
	return Project{
		Path:    ws.Path,
		Remotes: []Remote{ws.Remote},
		Exists:  ws.Exists,
	}
}

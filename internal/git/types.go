package git

type RepositoryStatus struct {
	Path        string
	Exists      bool
	Clean       bool
	Branch      string
	Ahead       int
	Behind      int
	Uncommitted int
	Untracked   int
	HasRemote   bool
	Error       error
}

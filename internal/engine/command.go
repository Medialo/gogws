package engine

type CommandType int

const (
	CommandTypeGit CommandType = iota
	CommandTypeShell
	CommandTypeCustom
)

type RepoCommand struct {
	RepoPath string
	RepoName string
	Type     CommandType
	Args     []string
	Action   func() (string, error)
	Context  map[string]any
	order    int
}

func NewGitCommand(repoPath, repoName string, args ...string) RepoCommand {
	return RepoCommand{
		RepoPath: repoPath,
		RepoName: repoName,
		Type:     CommandTypeGit,
		Args:     args,
		Context:  make(map[string]any),
	}
}

func NewShellCommand(repoPath, repoName, command string) RepoCommand {
	return RepoCommand{
		RepoPath: repoPath,
		RepoName: repoName,
		Type:     CommandTypeShell,
		Args:     []string{command},
		Context:  make(map[string]any),
	}
}

func NewCustomCommand(repoPath, repoName string, action func() (string, error)) RepoCommand {
	return RepoCommand{
		RepoPath: repoPath,
		RepoName: repoName,
		Type:     CommandTypeCustom,
		Action:   action,
		Context:  make(map[string]any),
	}
}

func (c *RepoCommand) WithContext(key string, value any) RepoCommand {
	if c.Context == nil {
		c.Context = make(map[string]any)
	}
	c.Context[key] = value
	return *c
}

func (c *RepoCommand) GetContext(key string) (any, bool) {
	if c.Context == nil {
		return nil, false
	}
	val, ok := c.Context[key]
	return val, ok
}

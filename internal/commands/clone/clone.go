package clone

import (
	"fmt"
	"path/filepath"

	"gogws/internal/config"
	"gogws/internal/git"
	"gogws/internal/gws"
	"gogws/internal/log"
	"gogws/internal/ui/cli"

	"github.com/spf13/cobra"
)

func NewCommand(getConfig func() *config.Config) *cobra.Command {
	return &cobra.Command{
		Use:   "clone [repository...]",
		Short: "Clone specific repositories",
		Long: `Clone one or more specific repositories by their path.

Use --skip-passphrase to skip projects requiring SSH passphrase.`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runClone(getConfig, args)
		},
	}
}

func runClone(getConfig func() *config.Config, args []string) error {
	cfg := getConfig()
	if cfg == nil {
		return fmt.Errorf("no workspace found (no .projects.gws file)")
	}

	log.Debugf("Running clone command in workspace: %s", cfg.WorkspaceRoot)

	parser := gws.NewParser(cfg.WorkspaceRoot)
	projects, err := parser.ParseProjects()
	if err != nil {
		return fmt.Errorf("failed to load projects: %w", err)
	}

	projectMap := make(map[string]gws.Project)
	for _, project := range projects {
		projectMap[project.Path] = project
	}

	renderer := cli.NewRenderer()

	for _, repoPath := range args {
		project, exists := projectMap[repoPath]
		if !exists {
			fmt.Println(renderer.RenderError(fmt.Sprintf("%s: not found in .projects.gws", repoPath)))
			continue
		}

		fullPath := filepath.Join(cfg.WorkspaceRoot, project.Path)
		status := git.GetStatus(fullPath)
		if status.Exists {
			fmt.Println(renderer.RenderWarning(fmt.Sprintf("%s: already exists", repoPath)))
			continue
		}

		log.Debugf("Cloning %s...", repoPath)
		fmt.Println(renderer.RenderInfo(fmt.Sprintf("Cloning %s...", repoPath)))

		remotes := toGitRemotes(project.Remotes)
		err := git.CloneWorkspace(cfg.WorkspaceRoot, project.Path, remotes)
		if err != nil {
			if git.IsPassphraseRequiredError(err) {
				fmt.Println(renderer.RenderWarning(fmt.Sprintf("%s: skipped (passphrase required)", repoPath)))
			} else {
				fmt.Println(renderer.RenderError(fmt.Sprintf("%s: %v", repoPath, err)))
			}
		} else {
			fmt.Println(renderer.RenderSuccess(repoPath))
		}
	}

	return nil
}

func toGitRemotes(remotes []gws.Remote) []git.Remote {
	result := make([]git.Remote, len(remotes))
	for i, r := range remotes {
		result[i] = git.Remote{Name: r.Name, URL: r.URL}
	}
	return result
}

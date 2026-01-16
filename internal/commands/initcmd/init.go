package initcmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gogws/internal/config"
	"gogws/internal/git"
	"gogws/internal/gws"
	"gogws/internal/hooks"
	"gogws/internal/log"
	"gogws/internal/ui/cli"

	"github.com/spf13/cobra"
)

func NewCommand(getConfig func() *config.Config) *cobra.Command {
	return &cobra.Command{
		Use:   "init",
		Short: "Initialize workspace by discovering existing repositories",
		Long: `Scan the workspace directory for existing git repositories and 
create a .projects.gws file with all discovered repositories.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runInit(getConfig)
		},
	}
}

func runInit(getConfig func() *config.Config) error {
	cfg := getConfig()
	if cfg != nil {
		if err := hooks.PreInit(cfg); err != nil {
			return fmt.Errorf("pre-init hook failed: %w", err)
		}
	}

	workspaceRoot, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	log.Debug("Initializing workspace", "path", workspaceRoot)

	projectsFile := filepath.Join(workspaceRoot, gws.ProjectsFileName)
	if _, err := os.Stat(projectsFile); err == nil {
		return fmt.Errorf("%s already exists. Remove it first if you want to reinitialize", gws.ProjectsFileName)
	}

	renderer := cli.NewRenderer()
	fmt.Println(renderer.RenderInfo("Scanning workspace for git repositories..."))

	discovered, err := git.DiscoverRepositories(workspaceRoot, 10)
	if err != nil {
		return fmt.Errorf("failed to discover repositories: %w", err)
	}

	if len(discovered) == 0 {
		return fmt.Errorf("no git repositories found in workspace")
	}

	log.Debug("Found repositories", "count", len(discovered))

	projects := make([]gws.Project, len(discovered))
	for i, d := range discovered {
		projects[i] = gws.Project{
			Path:    d.Path,
			Remotes: toGwsRemotes(d.Remotes),
		}
	}
	fmt.Println(renderer.RenderProjectsList(projects))

	file, err := os.Create(projectsFile)
	if err != nil {
		return fmt.Errorf("failed to create %s: %w", gws.ProjectsFileName, err)
	}
	defer file.Close()

	var projectPaths []string
	for _, project := range projects {
		projectPaths = append(projectPaths, project.Path)
		var remoteParts []string
		for _, remote := range project.Remotes {
			remoteParts = append(remoteParts, fmt.Sprintf("%s %s", remote.URL, remote.Name))
		}
		line := fmt.Sprintf("%s | %s\n", project.Path, strings.Join(remoteParts, " | "))
		if _, err := file.WriteString(line); err != nil {
			return fmt.Errorf("failed to write to %s: %w", gws.ProjectsFileName, err)
		}
	}

	fmt.Println(renderer.RenderSuccess(fmt.Sprintf("Created %s with %d repositories", gws.ProjectsFileName, len(projects))))

	if cfg != nil {
		if err := hooks.PostInit(cfg, projectPaths); err != nil {
			return fmt.Errorf("post-init hook failed: %w", err)
		}
	}

	return nil
}

func toGwsRemotes(remotes []git.Remote) []gws.Remote {
	result := make([]gws.Remote, len(remotes))
	for i, r := range remotes {
		result[i] = gws.Remote{Name: r.Name, URL: r.URL}
	}
	return result
}

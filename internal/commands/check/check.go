package check

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
		Use:   "check",
		Short: "Check workspace consistency",
		Long: `Check the workspace for all repositories (known, unknown, ignored, missing).
This can be slow for large workspaces.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCheck(getConfig)
		},
	}
}

func runCheck(getConfig func() *config.Config) error {
	cfg := getConfig()
	if cfg == nil {
		return fmt.Errorf("no workspace found (no .projects.gws file)")
	}

	log.Debug("Running check command", "workspace", cfg.WorkspaceRoot)

	parser := gws.NewParser(cfg.WorkspaceRoot)
	projects, err := parser.ParseProjects()
	if err != nil {
		return fmt.Errorf("failed to load projects: %w", err)
	}
	log.Debug("Loaded projects", "projects", projects)

	renderer := cli.NewRenderer()

	fmt.Println(renderer.RenderInfo("Checking known repositories..."))

	missing := 0
	for _, project := range projects {
		repoPath := filepath.Join(cfg.WorkspaceRoot, project.Path)
		status := git.GetStatus(repoPath)
		if !status.Exists {
			fmt.Println(renderer.RenderError(fmt.Sprintf("Missing: %s", project.Path)))
			missing++
		}
	}

	if missing == 0 {
		fmt.Println(renderer.RenderSuccess("All known repositories are present"))
	} else {
		fmt.Println(renderer.RenderWarning(fmt.Sprintf("%d repositories are missing", missing)))
	}

	fmt.Println()
	fmt.Println(renderer.RenderInfo("Scanning for unknown repositories..."))

	knownPaths := make([]string, len(projects))
	for i, p := range projects {
		knownPaths[i] = p.Path
	}

	unknown, err := git.FindUnknownRepositories(cfg.WorkspaceRoot, knownPaths)
	if err != nil {
		return fmt.Errorf("failed to check unknown repositories: %w", err)
	}

	if len(unknown) == 0 {
		fmt.Println(renderer.RenderSuccess("No unknown repositories found"))
	} else {
		fmt.Println(renderer.RenderWarning(fmt.Sprintf("Found %d unknown repositories:", len(unknown))))
		for _, path := range unknown {
			fmt.Println(renderer.RenderInfo(fmt.Sprintf("  %s", path)))
		}
	}

	return nil
}

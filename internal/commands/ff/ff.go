package ff

import (
	"fmt"
	"log/slog"
	"path/filepath"

	"gogws/internal/config"
	"gogws/internal/engine"
	"gogws/internal/git"
	"gogws/internal/gws"
	"gogws/internal/hooks"
	"gogws/internal/ui/cli"

	"github.com/spf13/cobra"
)

func NewCommand(getConfig func() *config.Config) *cobra.Command {
	return &cobra.Command{
		Use:   "ff",
		Short: "Fast-forward pull all repositories",
		Long:  `Fast-forward pull from origin for all repositories (only if fast-forward is possible).`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runFF(getConfig)
		},
	}
}

func runFF(getConfig func() *config.Config) error {
	cfg := getConfig()
	if cfg == nil {
		return fmt.Errorf("no workspace found (no .projects.gws file)")
	}

	if err := hooks.PreFF(cfg.WorkspaceRoot); err != nil {
		return fmt.Errorf("pre-ff hook failed: %w", err)
	}

	slog.Debug("Running ff command", "workspace", cfg.WorkspaceRoot)

	ws, err := gws.New(cfg.WorkspaceRoot).Recursive(false).Load()
	if err != nil {
		return fmt.Errorf("failed to load projects: %w", err)
	}

	renderer := cli.NewRenderer()
	output := engine.NewOutputHandler(renderer, false)

	commands := make([]engine.RepoCommand, 0, len(ws.Projects))
	var skippedResults []engine.Result

	for _, p := range ws.Projects {
		repoPath := filepath.Join(cfg.WorkspaceRoot, p.Path)
		status := git.GetStatus(repoPath)

		if !status.Exists {
			cmd := engine.NewGitCommand(repoPath, p.Path, "pull", "--ff-only")
			skippedResults = append(skippedResults, engine.Skip(cmd, "not cloned yet"))
			continue
		}

		commands = append(commands, engine.NewGitCommand(repoPath, p.Path, "pull", "--ff-only"))
	}

	result := engine.Execute(commands, engine.ExecuteOptions{
		Parallel:    cfg.Parallel,
		StopOnError: cfg.StopOnError,
	})

	for _, r := range skippedResults {
		result.AddResult(r)
	}

	output.RenderSummary(result, "Pulled")

	if err := hooks.PostFF(cfg.WorkspaceRoot, result.SuccessCount()); err != nil {
		return fmt.Errorf("post-ff hook failed: %w", err)
	}

	return nil
}

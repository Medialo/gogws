package status

import (
	"fmt"
	"path/filepath"

	"gogws/internal/config"
	"gogws/internal/export"
	"gogws/internal/git"
	"gogws/internal/gws"
	"gogws/internal/log"
	"gogws/internal/ui/cli"

	"github.com/spf13/cobra"
)

func NewCommand(getConfig func() *config.Config) *cobra.Command {
	return &cobra.Command{
		Use:     "status",
		Aliases: []string{"st"},
		Short:   "Show the status of all repositories in the workspace",
		Long: `Display the status of all repositories defined in .projects.gws file.
Shows uncommitted changes, untracked files, and sync status with remotes.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runStatus(getConfig)
		},
	}
}

func runStatus(getConfig func() *config.Config) error {
	cfg := getConfig()
	if cfg == nil {
		return fmt.Errorf("no workspace found (no .projects.gws file)")
	}

	log.Debug("Running status command", "workspace", cfg.WorkspaceRoot)

	resolver := gws.NewResolver()
	resolved, err := resolver.Resolve(cfg.WorkspaceRoot)
	if err != nil {
		return fmt.Errorf("failed to resolve workspace: %w", err)
	}

	if len(resolved.Projects) == 0 && len(resolved.Workspaces) == 0 {
		return fmt.Errorf("no projects or workspaces found")
	}

	log.Debug("Found projects and workspaces", "projects", len(resolved.Projects), "workspaces", len(resolved.Workspaces))

	if cfg.TUI {
		return runStatusTUI(cfg, resolved)
	}

	var statuses []git.RepositoryStatus
	for _, project := range resolved.Projects {
		log.Debug("Getting status", "project", project.Path)
		repoPath := filepath.Join(cfg.WorkspaceRoot, project.Path)
		status := git.GetStatus(repoPath)
		status.Path = project.Path
		statuses = append(statuses, status)
	}

	if cfg.Format == "json" || cfg.Format == "yaml" {
		output, err := export.Format(statuses, cfg.Format)
		if err != nil {
			return fmt.Errorf("failed to export status: %w", err)
		}
		fmt.Println(output)
		return nil
	}

	renderer := cli.NewRenderer()
	output := renderer.RenderStatus(statuses, resolved, resolved.Workspaces, cfg.OnlyChanges)
	fmt.Println(output)

	return nil
}

func runStatusTUI(cfg *config.Config, resolved *gws.Workspace) error {
	return fmt.Errorf("TUI mode not yet implemented - coming soon!")
}

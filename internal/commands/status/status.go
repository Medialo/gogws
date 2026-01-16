package status

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"path/filepath"
	"sync"

	"gogws/internal/config"
	"gogws/internal/engine"
	"gogws/internal/export"
	"gogws/internal/git"
	"gogws/internal/gws"
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

	slog.Debug("Running status command", "workspace", cfg.WorkspaceRoot)

	ws, err := gws.New(cfg.WorkspaceRoot).Load()
	if err != nil {
		return fmt.Errorf("failed to resolve workspace: %w", err)
	}

	if len(ws.Projects) == 0 && len(ws.Children) == 0 {
		return fmt.Errorf("no projects or workspaces found")
	}

	slog.Debug("Found projects and workspaces", "projects", len(ws.Projects), "workspaces", len(ws.Children))

	statuses := getStatuses(cfg.WorkspaceRoot, ws.Projects, cfg.Parallel)

	if cfg.Format == "json" || cfg.Format == "yaml" {
		output, err := export.Format(statuses, cfg.Format)
		if err != nil {
			return fmt.Errorf("failed to export status: %w", err)
		}
		fmt.Println(output)
		return nil
	}

	renderer := cli.NewRenderer()
	output := renderer.RenderStatus(statuses, ws, ws.Children, cfg.OnlyChanges)
	fmt.Println(output)

	return nil
}

func getStatuses(workspaceRoot string, projects []gws.Project, parallel int) []git.RepositoryStatus {
	if len(projects) == 0 {
		return nil
	}

	var mu sync.Mutex
	statusMap := make(map[string]git.RepositoryStatus)

	commands := make([]engine.RepoCommand, 0, len(projects))

	for _, p := range projects {
		repoPath := filepath.Join(workspaceRoot, p.Path)
		projectPath := p.Path

		cmd := engine.NewCustomCommand(
			repoPath,
			projectPath,
			func() (string, error) {
				status := git.GetStatus(repoPath)
				status.Path = projectPath

				data, err := json.Marshal(status)
				if err != nil {
					return "", err
				}
				return string(data), nil
			},
		)
		commands = append(commands, cmd)
	}

	result := engine.Execute(commands, engine.ExecuteOptions{
		Parallel: parallel,
		OnComplete: func(r engine.Result) {
			if r.Success && r.Stdout != "" {
				var status git.RepositoryStatus
				if err := json.Unmarshal([]byte(r.Stdout), &status); err == nil {
					mu.Lock()
					statusMap[r.Command.RepoName] = status
					mu.Unlock()
				}
			}
		},
	})

	statuses := make([]git.RepositoryStatus, 0, len(projects))
	for _, r := range result.Results {
		if status, ok := statusMap[r.Command.RepoName]; ok {
			statuses = append(statuses, status)
		} else {
			statuses = append(statuses, git.RepositoryStatus{
				Path:   r.Command.RepoName,
				Exists: false,
				Error:  r.Error,
			})
		}
	}

	return statuses
}

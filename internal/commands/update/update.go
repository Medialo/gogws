package update

import (
	"fmt"
	"path/filepath"
	"sync"

	"gogws/internal/config"
	"gogws/internal/git"
	"gogws/internal/gws"
	"gogws/internal/log"
	"gogws/internal/ui/cli"

	"github.com/spf13/cobra"
)

var (
	skipProjects   bool
	skipWorkspaces bool
)

func NewCommand(getConfig func() *config.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update",
		Short: "Clone all missing repositories and workspaces",
		Long: `Clone all repositories defined in .projects.gws and workspaces defined 
in .workspaces.gws that are not yet present in the workspace.

Use --skip-projects to only clone workspaces (recursive).
Use --skip-workspaces to only clone projects.
Use --skip-passphrase to skip projects requiring SSH passphrase.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runUpdate(getConfig)
		},
	}

	cmd.Flags().BoolVar(&skipProjects, "skip-projects", false, "skip cloning projects, only clone workspaces")
	cmd.Flags().BoolVar(&skipWorkspaces, "skip-workspaces", false, "skip cloning workspaces, only clone projects")

	return cmd
}

func runUpdate(getConfig func() *config.Config) error {
	cfg := getConfig()
	if cfg == nil {
		return fmt.Errorf("no workspace found (no .projects.gws file)")
	}

	log.Debugf("Running update command in workspace: %s", cfg.WorkspaceRoot)

	resolver := gws.NewResolver()
	resolved, err := resolver.Resolve(cfg.WorkspaceRoot)
	if err != nil {
		return fmt.Errorf("failed to resolve workspace: %w", err)
	}

	renderer := cli.NewRenderer()

	if !skipWorkspaces && len(resolved.Workspaces) > 0 {
		if err := cloneWorkspaces(cfg.WorkspaceRoot, resolved, renderer, cfg.Parallel); err != nil {
			log.Warn("Some workspaces failed to clone", "err", err)
		}
	}

	if !skipProjects {
		missingProjects := resolved.MissingProjects()
		if err := cloneProjects(cfg.WorkspaceRoot, missingProjects, renderer, cfg.Parallel); err != nil {
			return err
		}
	}

	return nil
}

func cloneWorkspaces(workspaceRoot string, resolved *gws.Workspace, renderer *cli.Renderer, parallel int) error {
	toClone := resolved.MissingWorkspaces()
	if len(toClone) == 0 {
		fmt.Println(renderer.RenderSuccess("All workspaces are already cloned"))
		return nil
	}

	fmt.Println(renderer.RenderInfo(fmt.Sprintf("Cloning %d missing workspaces...", len(toClone))))
	fmt.Println()

	cloned := 0
	failed := 0
	skippedPassphrase := 0

	for _, ws := range toClone {
		wsPath := filepath.Join(workspaceRoot, ws.Path)
		log.Debugf("Cloning workspace: %s from %s", ws.Path, ws.Remote.URL)

		remotes := []git.Remote{{Name: ws.Remote.Name, URL: ws.Remote.URL}}
		err := git.CloneWorkspace(workspaceRoot, ws.Path, remotes)

		if err != nil {
			if git.IsPassphraseRequiredError(err) {
				fmt.Println(renderer.RenderWarning(fmt.Sprintf("%s: skipped (passphrase required)", ws.Path)))
				skippedPassphrase++
			} else {
				fmt.Println(renderer.RenderError(fmt.Sprintf("%s: %v", ws.Path, err)))
				failed++
			}
		} else {
			fmt.Println(renderer.RenderSuccess(fmt.Sprintf("Workspace: %s", ws.Path)))
			cloned++

			childParser := gws.NewParser(wsPath)
			if childParser.HasWorkspacesFile() {
				log.Debugf("Workspace %s has sub-workspaces, will be processed on next update", ws.Path)
			}
		}
	}

	fmt.Println()
	if failed > 0 || skippedPassphrase > 0 {
		msg := fmt.Sprintf("Cloned %d workspaces", cloned)
		if failed > 0 {
			msg += fmt.Sprintf(", %d failed", failed)
		}
		if skippedPassphrase > 0 {
			msg += fmt.Sprintf(", %d skipped (passphrase)", skippedPassphrase)
		}
		fmt.Println(renderer.RenderWarning(msg))
	} else {
		fmt.Println(renderer.RenderSuccess(fmt.Sprintf("Successfully cloned %d workspaces", cloned)))
	}

	return nil
}

func cloneProjects(workspaceRoot string, toClone []gws.Project, renderer *cli.Renderer, parallel int) error {
	if len(toClone) == 0 {
		fmt.Println(renderer.RenderSuccess("All projects are already cloned"))
		return nil
	}

	fmt.Println(renderer.RenderInfo(fmt.Sprintf("Cloning %d missing projects...", len(toClone))))
	fmt.Println()

	if git.GetSkipPassphrase() || !git.GetUseAgent() {
		return cloneProjectsSequential(workspaceRoot, toClone, renderer)
	}

	return cloneProjectsParallel(workspaceRoot, toClone, renderer, parallel)
}

func cloneProjectsSequential(workspaceRoot string, toClone []gws.Project, renderer *cli.Renderer) error {
	cloned := 0
	failed := 0
	skippedPassphrase := 0

	for i, p := range toClone {
		log.Debugf("Cloning project: %s", p.Path)
		fmt.Printf("[%d/%d] %s\n", i+1, len(toClone), renderer.RenderInfo(p.Path))

		remotes := toGitRemotes(p.Remotes)
		err := git.CloneWorkspace(workspaceRoot, p.Path, remotes)

		if err != nil {
			if git.IsPassphraseRequiredError(err) {
				fmt.Println(renderer.RenderWarning("  skipped (passphrase required)"))
				skippedPassphrase++
			} else {
				fmt.Println(renderer.RenderError(fmt.Sprintf("  %v", err)))
				failed++
			}
		} else {
			fmt.Println(renderer.RenderSuccess("  done"))
			cloned++
		}
	}

	fmt.Println()
	if failed > 0 || skippedPassphrase > 0 {
		msg := fmt.Sprintf("Cloned %d projects", cloned)
		if failed > 0 {
			msg += fmt.Sprintf(", %d failed", failed)
		}
		if skippedPassphrase > 0 {
			msg += fmt.Sprintf(", %d skipped (passphrase)", skippedPassphrase)
		}
		fmt.Println(renderer.RenderWarning(msg))
	} else {
		fmt.Println(renderer.RenderSuccess(fmt.Sprintf("Successfully cloned %d projects", cloned)))
	}

	return nil
}

func cloneProjectsParallel(workspaceRoot string, toClone []gws.Project, renderer *cli.Renderer, parallel int) error {
	var wg sync.WaitGroup
	sem := make(chan struct{}, parallel)
	var mu sync.Mutex
	cloned := 0
	failed := 0

	for _, project := range toClone {
		wg.Add(1)
		go func(p gws.Project) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			log.Debugf("Cloning project: %s", p.Path)

			mu.Lock()
			current := cloned + failed + 1
			fmt.Printf("\r%s", renderer.RenderProgress(current, len(toClone), p.Path))
			mu.Unlock()

			remotes := toGitRemotes(p.Remotes)
			err := git.CloneWorkspace(workspaceRoot, p.Path, remotes)

			mu.Lock()
			if err != nil {
				fmt.Printf("\r%s\n", renderer.RenderError(fmt.Sprintf("%s: %v", p.Path, err)))
				failed++
			} else {
				fmt.Printf("\r%s\n", renderer.RenderSuccess(p.Path))
				cloned++
			}
			mu.Unlock()
		}(project)
	}

	wg.Wait()
	fmt.Println()

	if failed > 0 {
		fmt.Println(renderer.RenderWarning(fmt.Sprintf("Cloned %d projects, %d failed", cloned, failed)))
	} else {
		fmt.Println(renderer.RenderSuccess(fmt.Sprintf("Successfully cloned %d projects", cloned)))
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

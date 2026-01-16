package fetch

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

func NewCommand(getConfig func() *config.Config) *cobra.Command {
	return &cobra.Command{
		Use:   "fetch",
		Short: "Fetch updates from origin for all repositories",
		Long: `Fetch updates from origin remote for all repositories in the workspace.

Use --skip-passphrase to skip projects requiring SSH passphrase.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runFetch(getConfig)
		},
	}
}

func runFetch(getConfig func() *config.Config) error {
	cfg := getConfig()
	if cfg == nil {
		return fmt.Errorf("no workspace found (no .projects.gws file)")
	}

	log.Debugf("Running fetch command in workspace: %s", cfg.WorkspaceRoot)

	parser := gws.NewParser(cfg.WorkspaceRoot)
	projects, err := parser.ParseProjects()
	if err != nil {
		return fmt.Errorf("failed to load projects: %w", err)
	}

	renderer := cli.NewRenderer()

	if git.GetSkipPassphrase() || !git.GetUseAgent() {
		return fetchProjectsSequential(cfg.WorkspaceRoot, projects, renderer, cfg.OnlyChanges)
	}

	return fetchProjectsParallel(cfg.WorkspaceRoot, projects, renderer, cfg.Parallel, cfg.OnlyChanges)
}

func fetchProjectsSequential(workspaceRoot string, projects []gws.Project, renderer *cli.Renderer, onlyChanges bool) error {
	fetched := 0
	failed := 0
	skipped := 0
	skippedPassphrase := 0

	for _, p := range projects {
		repoPath := filepath.Join(workspaceRoot, p.Path)
		status := git.GetStatus(repoPath)
		if !status.Exists {
			skipped++
			if !onlyChanges {
				fmt.Println(renderer.RenderWarning(fmt.Sprintf("%s: not cloned yet", p.Path)))
			}
			continue
		}

		log.Debugf("Fetching %s...", p.Path)
		err := git.Fetch(repoPath)

		if err != nil {
			if git.IsPassphraseRequiredError(err) {
				fmt.Println(renderer.RenderWarning(fmt.Sprintf("%s: skipped (passphrase required)", p.Path)))
				skippedPassphrase++
			} else {
				fmt.Println(renderer.RenderError(fmt.Sprintf("%s: %v", p.Path, err)))
				failed++
			}
		} else {
			if !onlyChanges {
				fmt.Println(renderer.RenderSuccess(p.Path))
			}
			fetched++
		}
	}

	fmt.Println()
	if failed > 0 || skippedPassphrase > 0 {
		msg := fmt.Sprintf("Fetched %d repositories", fetched)
		if failed > 0 {
			msg += fmt.Sprintf(", %d failed", failed)
		}
		if skippedPassphrase > 0 {
			msg += fmt.Sprintf(", %d skipped (passphrase)", skippedPassphrase)
		}
		if skipped > 0 {
			msg += fmt.Sprintf(", %d not cloned", skipped)
		}
		fmt.Println(renderer.RenderWarning(msg))
	} else {
		fmt.Println(renderer.RenderSuccess(fmt.Sprintf("Successfully fetched %d repositories", fetched)))
	}

	return nil
}

func fetchProjectsParallel(workspaceRoot string, projects []gws.Project, renderer *cli.Renderer, parallel int, onlyChanges bool) error {
	var wg sync.WaitGroup
	sem := make(chan struct{}, parallel)
	var mu sync.Mutex
	fetched := 0
	failed := 0
	skipped := 0

	for _, project := range projects {
		wg.Add(1)
		go func(p gws.Project) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			repoPath := filepath.Join(workspaceRoot, p.Path)
			status := git.GetStatus(repoPath)
			if !status.Exists {
				mu.Lock()
				skipped++
				if !onlyChanges {
					fmt.Println(renderer.RenderWarning(fmt.Sprintf("%s: not cloned yet", p.Path)))
				}
				mu.Unlock()
				return
			}

			log.Debugf("Fetching %s...", p.Path)
			err := git.Fetch(repoPath)

			mu.Lock()
			if err != nil {
				fmt.Println(renderer.RenderError(fmt.Sprintf("%s: %v", p.Path, err)))
				failed++
			} else {
				if !onlyChanges {
					fmt.Println(renderer.RenderSuccess(p.Path))
				}
				fetched++
			}
			mu.Unlock()
		}(project)
	}

	wg.Wait()

	fmt.Println()
	if failed > 0 {
		fmt.Println(renderer.RenderWarning(fmt.Sprintf("Fetched %d repositories, %d failed, %d skipped", fetched, failed, skipped)))
	} else {
		fmt.Println(renderer.RenderSuccess(fmt.Sprintf("Successfully fetched %d repositories", fetched)))
	}

	return nil
}

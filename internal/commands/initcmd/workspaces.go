package initcmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gogws/internal/config"
	"gogws/internal/gitignore"
	"gogws/internal/gws"
	"gogws/internal/ui/cli"

	"github.com/spf13/cobra"
)

var (
	workspacesGitignore bool
)

func newWorkspacesCommand(getConfig func() *config.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "workspaces",
		Short: "Interactively configure sub-workspaces",
		Long: `Scan subdirectories and interactively select which ones should be
configured as sub-workspaces. For each selected directory, you can provide
a git remote URL.

Creates a .gws/workspaces.gws file with the configured workspaces.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runInitWorkspaces()
		},
	}

	cmd.Flags().BoolVar(&workspacesGitignore, "gitignore", true, "generate .gitignore file")

	return cmd
}

func runInitWorkspaces() error {
	workspaceRoot, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	renderer := cli.NewRenderer()
	reader := bufio.NewReader(os.Stdin)

	entries, err := os.ReadDir(workspaceRoot)
	if err != nil {
		return fmt.Errorf("failed to read directory: %w", err)
	}

	var subdirs []string
	for _, entry := range entries {
		if entry.IsDir() && !strings.HasPrefix(entry.Name(), ".") {
			subdirs = append(subdirs, entry.Name())
		}
	}

	if len(subdirs) == 0 {
		fmt.Println(renderer.RenderWarning("No subdirectories found"))
		return nil
	}

	fmt.Println(renderer.RenderInfo("Found subdirectories:"))
	for i, dir := range subdirs {
		fmt.Printf("  [%d] %s\n", i+1, dir)
	}
	fmt.Println()

	var selectedWorkspaces []workspaceEntry

	for _, dir := range subdirs {
		fmt.Printf("Configure '%s' as a workspace? [y/N/q]: ", dir)
		input, err := reader.ReadString('\n')
		if err != nil {
			return err
		}
		input = strings.TrimSpace(strings.ToLower(input))

		if input == "q" || input == "quit" {
			break
		}

		if input != "y" && input != "yes" {
			continue
		}

		fmt.Printf("  Git remote URL for '%s' (or press Enter to skip): ", dir)
		urlInput, err := reader.ReadString('\n')
		if err != nil {
			return err
		}
		urlInput = strings.TrimSpace(urlInput)

		entry := workspaceEntry{
			Path: dir,
			Name: dir,
		}

		if urlInput != "" {
			entry.RemoteURL = urlInput
			entry.RemoteName = "origin"
		}

		selectedWorkspaces = append(selectedWorkspaces, entry)
		fmt.Println(renderer.RenderSuccess(fmt.Sprintf("  Added '%s'", dir)))
	}

	if len(selectedWorkspaces) == 0 {
		fmt.Println(renderer.RenderWarning("No workspaces configured"))
		return nil
	}

	gwsDir := filepath.Join(workspaceRoot, gws.ConfigDirName)
	if err := os.MkdirAll(gwsDir, 0755); err != nil {
		return fmt.Errorf("failed to create %s directory: %w", gws.ConfigDirName, err)
	}

	workspacesFile := filepath.Join(gwsDir, "workspaces."+gws.FileExtension)

	file, err := os.Create(workspacesFile)
	if err != nil {
		return fmt.Errorf("failed to create %s: %w", workspacesFile, err)
	}
	defer file.Close()

	for _, ws := range selectedWorkspaces {
		var line string
		if ws.RemoteURL != "" {
			line = fmt.Sprintf("%s | %s %s\n", ws.Path, ws.RemoteURL, ws.RemoteName)
		} else {
			line = fmt.Sprintf("# %s (no remote configured)\n", ws.Path)
		}
		if _, err := file.WriteString(line); err != nil {
			return fmt.Errorf("failed to write to %s: %w", workspacesFile, err)
		}
	}

	fmt.Println()
	fmt.Println(renderer.RenderSuccess(fmt.Sprintf("Created %s with %d workspaces", workspacesFile, len(selectedWorkspaces))))

	if workspacesGitignore {
		if err := gitignore.EnsureGWSSection(workspaceRoot); err != nil {
			fmt.Println(renderer.RenderWarning(fmt.Sprintf("Failed to generate .gitignore: %v", err)))
		} else {
			fmt.Println(renderer.RenderSuccess("Generated .gitignore"))
		}
	}

	return nil
}

type workspaceEntry struct {
	Path       string
	Name       string
	RemoteURL  string
	RemoteName string
}

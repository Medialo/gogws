package initcmd

import (
	"gogws/internal/config"

	"github.com/spf13/cobra"
)

func NewCommand(getConfig func() *config.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize workspace configuration",
		Long: `Initialize workspace configuration files.

Available subcommands:
  projects    - Discover git repositories and create projects.gws
  workspaces  - Interactively configure sub-workspaces
  gitignore   - Generate or update .gitignore for GWS

Running 'gogws init' without subcommand is equivalent to 'gogws init projects'.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			projectsCmd := newProjectsCommand(getConfig)
			return projectsCmd.RunE(projectsCmd, args)
		},
	}

	cmd.AddCommand(newProjectsCommand(getConfig))
	cmd.AddCommand(newWorkspacesCommand(getConfig))
	cmd.AddCommand(newGitignoreCommand())

	return cmd
}

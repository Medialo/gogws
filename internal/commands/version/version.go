package version

import (
	"fmt"

	"github.com/spf13/cobra"
)

var Version = "1.0.0"

func NewCommand() *cobra.Command {
	return &cobra.Command{
		Use:     "version",
		Short:   "Print the version number of gogws",
		Aliases: []string{"v"},
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("gogws version %s\n", Version)
		},
	}
}

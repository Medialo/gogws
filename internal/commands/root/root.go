package root

import (
	"fmt"
	"gogws/internal/config"
	"gogws/internal/git"
	"gogws/internal/log"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile        string
	themeFile      string
	parallel       int
	format         string
	noColor        bool
	onlyChanges    bool
	useTUI         bool
	verbose        bool
	skipPassphrase bool
)

var rootCmd = &cobra.Command{
	Use:   "gogws",
	Short: "Git Workspace Manager - Manage multiple git repositories with ease",
	Long: `gogws is a modern Git workspace management tool written in Go.
It helps you manage multiple git repositories in a workspace with
interactive TUI and beautiful CLI output.

Compatible with gws project files (.projects.gws)`,
	PersistentPreRunE: persistentPreRun,
}

func persistentPreRun(cmd *cobra.Command, args []string) error {
	log.Debug("persistentPreRun::")
	log.SetVerbose(verbose)
	git.SetSkipPassphrase(skipPassphrase)

	// intercept config command, useless to load config
	if cmd.Name() == "config" {
		return nil
	}

	if err := config.Initialize(); err != nil {
		log.Debugf("Config initialization skipped: %v", err)
	}

	if config.IsInitialized() {
		config.ApplyFlags(themeFile, parallel, format, noColor, onlyChanges, useTUI)
	}

	return nil
}

func NewCommand() *cobra.Command {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default: $HOME/.config/gogws/config.yaml)")
	rootCmd.PersistentFlags().StringVar(&themeFile, "theme", "", "theme file")
	rootCmd.PersistentFlags().IntVar(&parallel, "parallel", 0, "number of parallel operations (default: 5)")
	rootCmd.PersistentFlags().StringVar(&format, "format", "text", "output format (text, json, yaml)")
	rootCmd.PersistentFlags().BoolVar(&noColor, "no-color", false, "disable colored output")
	rootCmd.PersistentFlags().BoolVar(&onlyChanges, "only-changes", false, "show only repositories with changes")
	rootCmd.PersistentFlags().BoolVar(&useTUI, "tui", false, "use interactive TUI mode")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "enable verbose output")
	rootCmd.PersistentFlags().BoolVar(&skipPassphrase, "skip-passphrase", false, "skip projects that require SSH passphrase")

	viper.BindPFlag("theme", rootCmd.PersistentFlags().Lookup("theme"))
	viper.BindPFlag("parallel", rootCmd.PersistentFlags().Lookup("parallel"))
	viper.BindPFlag("format", rootCmd.PersistentFlags().Lookup("format"))
	viper.BindPFlag("no_color", rootCmd.PersistentFlags().Lookup("no-color"))

	// help format command override
	//rootCmd.SetHelpFunc(func(command *cobra.Command, args []string) {
	//	help.HelpFunc(command, args)
	//})
	//	originalHelpFunc := rootCmd.HelpFunc()
	//	rootCmd.SetHelpFunc(func(c *cobra.Command, args []string) {
	//		// Créer un buffer pour capturer
	//		buf := new(bytes.Buffer)
	//
	//		// Sauvegarder la sortie originale
	//		originalOut := c.OutOrStdout()
	//
	//		// Rediriger temporairement vers le buffer
	//		c.SetOut(buf)
	//
	//		// Exécuter la fonction help originale (qui exécute le template)
	//		originalHelpFunc(c, args)
	//
	//		// Récupérer le markdown généré
	//		markdownOutput := buf.String()
	//
	//		// *** TRAITER LE RÉSULTAT ICI ***
	//		renderer, _ := glamour.NewTermRenderer(glamour.WithAutoStyle())
	//		styledOutput, _ := renderer.Render(markdownOutput)
	//
	//		// Restaurer la sortie originale et afficher
	//		c.SetOut(originalOut)
	//		c.Print(styledOutput)
	//	})
	//
	//	rootCmd.SetHelpTemplate(strings.ReplaceAll(`{{if .Runnable}}
	//# Usage:
	//
	//@@@bash
	//{{.UseLine}}
	//@@@
	//
	//{{end}}{{if .HasAvailableSubCommands}}
	//
	//  {{.CommandPath}} [command]{{end}}{{if gt (len .Aliases) 0}}
	//
	//# Aliases:
	//  {{.NameAndAliases}}{{end}}{{if .HasExample}}
	//
	//Examples:
	//{{.Example}}{{end}}{{if .HasAvailableSubCommands}}{{$cmds := .Commands}}{{if eq (len .Groups) 0}}
	//
	//# Available Commands:{{range $cmds}}{{if (or .IsAvailableCommand (eq .Name "help"))}}
	//  - {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}{{else}}{{range $group := .Groups}}
	//
	//{{.Title}}{{range $cmds}}{{if (and (eq .GroupID $group.ID) (or .IsAvailableCommand (eq .Name "help")))}}
	//  {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}{{end}}{{if not .AllChildCommandsHaveGroup}}
	//
	//# Additional Commands:{{range $cmds}}{{if (and (eq .GroupID "") (or .IsAvailableCommand (eq .Name "help")))}}
	//  {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}{{end}}{{end}}{{end}}{{if .HasAvailableLocalFlags}}
	//
	//# Flags:
	// - {{.LocalFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}
	//
	//{{if .HasAvailableInheritedFlags}}
	//# Global Flags:
	//{{.InheritedFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}
	//{{if .HasHelpSubCommands}}
	//
	//## Additional help topics:{{range .Commands}}{{if .IsAdditionalHelpTopicCommand}}
	//  {{rpad .CommandPath .CommandPathPadding}} {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableSubCommands}}
	//
	//> Use "{{.CommandPath}} [command] --help" for more information about a command.{{end}}
	//`, "@@@", "```"))

	return rootCmd
}

func GetConfig() *config.Config {
	return config.GetConfig()
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := os.UserHomeDir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		viper.AddConfigPath(home + "/.config/gogws")
		viper.SetConfigName("config")
		viper.SetConfigType("yaml")
	}

	viper.SetEnvPrefix("GOGWS")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err == nil {
		if viper.GetBool("verbose") {
			fmt.Println("Using config file:", viper.ConfigFileUsed())
		}
	}
}

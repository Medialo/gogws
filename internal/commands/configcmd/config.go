package configcmd

import (
	"fmt"
	"strconv"
	"strings"

	"gogws/internal/config"
	"gogws/internal/log"
	"gogws/internal/ui/cli"

	"github.com/spf13/cobra"
)

func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Manage gogws configuration",
		Long:  `View and manage gogws user configuration stored in ~/.gws/config.yaml`,
		RunE:  runConfigShow,
	}

	cmd.AddCommand(newGetCommand())
	cmd.AddCommand(newSetCommand())
	cmd.AddCommand(newListCommand())

	return cmd
}

func newGetCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "get <key>",
		Short: "Get a configuration value",
		Long:  `Get a specific configuration value by key.`,
		Args:  cobra.ExactArgs(1),
		RunE:  runConfigGet,
	}
}

func newSetCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "set <key> <value>",
		Short: "Set a configuration value",
		Long: `Set a configuration value.

Available keys:
  use-agent    Enable/disable SSH agent mode (true/false)
               When true: uses go-git library with SSH agent
               When false: uses system git commands`,
		Args: cobra.ExactArgs(2),
		RunE: runConfigSet,
	}
}

func newListCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all available configuration keys",
		RunE:  runConfigList,
	}
}

func runConfigShow(cmd *cobra.Command, args []string) error {
	log.Debug("Loading user configuration...")

	resolved, err := config.LoadUserConfigResolved()
	if err != nil {
		return fmt.Errorf("failed to load user config: %w", err)
	}

	configPath, _ := config.GetUserConfigPath()

	renderer := cli.NewRenderer()

	fmt.Println(renderer.RenderHeader("GOGWS Configuration"))
	fmt.Println()
	fmt.Printf("  File: %s\n\n", configPath)

	fmt.Println(renderer.RenderConfigValue("use-agent", resolved.UseAgent.Value, string(resolved.UseAgent.Source)))

	fmt.Println()
	fmt.Println("  Environment variables:")
	for _, key := range config.GetAvailableConfigKeys() {
		envVar := config.GetEnvVarName(key)
		fmt.Printf("    %s -> %s\n", key, envVar)
	}

	return nil
}

func runConfigGet(cmd *cobra.Command, args []string) error {
	key := args[0]
	log.Debug("Getting config value", "key", key)

	resolved, err := config.LoadUserConfigResolved()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	switch key {
	case "use-agent":
		fmt.Printf("%v (source: %s)\n", resolved.UseAgent.Value, resolved.UseAgent.Source)
	default:
		return fmt.Errorf("unknown configuration key: %s\n\nAvailable keys:\n  %s",
			key, strings.Join(config.GetAvailableConfigKeys(), "\n  "))
	}

	return nil
}

func runConfigSet(cmd *cobra.Command, args []string) error {
	key := args[0]
	valueStr := args[1]

	log.Debug("Setting config value", "key", key, "value", valueStr)

	var value interface{}

	switch key {
	case "use-agent":
		boolVal, err := strconv.ParseBool(strings.ToLower(valueStr))
		if err != nil {
			return fmt.Errorf("invalid boolean value for %s: %s (use true/false)", key, valueStr)
		}
		value = boolVal
	default:
		return fmt.Errorf("unknown configuration key: %s\n\nAvailable keys:\n  %s",
			key, strings.Join(config.GetAvailableConfigKeys(), "\n  "))
	}

	if err := config.SetUserConfigValue(key, value); err != nil {
		return fmt.Errorf("failed to set config value: %w", err)
	}

	renderer := cli.NewRenderer()
	fmt.Println(renderer.RenderSuccess(fmt.Sprintf("Set %s = %v", key, value)))
	return nil
}

func runConfigList(cmd *cobra.Command, args []string) error {
	renderer := cli.NewRenderer()

	fmt.Println(renderer.RenderHeader("Available Configuration Keys"))
	fmt.Println()

	keys := config.GetAvailableConfigKeys()
	for _, key := range keys {
		envVar := config.GetEnvVarName(key)
		fmt.Printf("  %s\n", key)
		fmt.Printf("    env: %s\n", envVar)

		switch key {
		case "use-agent":
			fmt.Printf("    type: boolean (true/false)\n")
			fmt.Printf("    desc: Enable SSH agent mode for go-git\n")
		}
		fmt.Println()
	}

	return nil
}

package config

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

const (
	UserConfigDir  = ".gws"
	UserConfigFile = "config.yaml"
)

type ConfigSource string

const (
	SourceDefault ConfigSource = "default"
	SourceFile    ConfigSource = "file"
	SourceEnv     ConfigSource = "env"
)

type ConfigValue[T any] struct {
	Value  T
	Source ConfigSource
}

type UserConfig struct {
	TrustedWorkspaces []string `yaml:"trusted-workspaces,omitempty"`
}

type UserConfigResolved struct {
	TrustedWorkspaces ConfigValue[[]string]
}

func GetUserConfigPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(homeDir, UserConfigDir, UserConfigFile), nil
}

func GetUserConfigDir() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(homeDir, UserConfigDir), nil
}

func LoadUserConfigResolved() (*UserConfigResolved, error) {
	resolved := &UserConfigResolved{
		TrustedWorkspaces: ConfigValue[[]string]{Value: []string{}, Source: SourceDefault},
	}

	configPath, err := GetUserConfigPath()
	if err != nil {
		return resolved, nil
	}

	data, err := os.ReadFile(configPath)
	if err == nil {
		var fileCfg UserConfig
		if yaml.Unmarshal(data, &fileCfg) == nil {
			if fileCfg.TrustedWorkspaces != nil {
				resolved.TrustedWorkspaces = ConfigValue[[]string]{Value: fileCfg.TrustedWorkspaces, Source: SourceFile}
			}
		}
	}

	return resolved, nil
}

func LoadUserConfig() (*UserConfig, error) {
	resolved, err := LoadUserConfigResolved()
	if err != nil {
		return nil, err
	}
	return &UserConfig{
		TrustedWorkspaces: resolved.TrustedWorkspaces.Value,
	}, nil
}

func SaveUserConfig(cfg *UserConfig) error {
	configDir, err := GetUserConfigDir()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(configDir, 0755); err != nil {
		return err
	}

	configPath, err := GetUserConfigPath()
	if err != nil {
		return err
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}

	return os.WriteFile(configPath, data, 0644)
}

func SetUserConfigValue(key string, value interface{}) error {
	cfg, err := LoadUserConfig()
	if err != nil {
		return err
	}

	switch key {
	case "trusted-workspaces":
		if v, ok := value.([]string); ok {
			cfg.TrustedWorkspaces = v
		}
	}

	return SaveUserConfig(cfg)
}

func AddTrustedWorkspace(pattern string) error {
	cfg, err := LoadUserConfig()
	if err != nil {
		return err
	}

	for _, existing := range cfg.TrustedWorkspaces {
		if existing == pattern {
			return nil
		}
	}

	cfg.TrustedWorkspaces = append(cfg.TrustedWorkspaces, pattern)
	return SaveUserConfig(cfg)
}

func GetUserConfigValue(key string) (interface{}, error) {
	cfg, err := LoadUserConfig()
	if err != nil {
		return nil, err
	}

	switch key {
	case "trusted-workspaces":
		return cfg.TrustedWorkspaces, nil
	default:
		return nil, nil
	}
}

func GetAvailableConfigKeys() []string {
	return []string{"trusted-workspaces"}
}

func GetEnvVarName(key string) string {
	switch key {
	case "trusted-workspaces":
		return ""
	default:
		return ""
	}
}

func GetUserHooksDir() (string, error) {
	configDir, err := GetUserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(configDir, "hooks"), nil
}

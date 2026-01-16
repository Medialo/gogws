package config

import (
	"os"
	"path/filepath"
	"strconv"

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
	UseAgent bool `yaml:"use-agent"`
}

type UserConfigResolved struct {
	UseAgent ConfigValue[bool]
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
		UseAgent: ConfigValue[bool]{Value: true, Source: SourceDefault},
	}

	configPath, err := GetUserConfigPath()
	if err != nil {
		return resolved, nil
	}

	data, err := os.ReadFile(configPath)
	if err == nil {
		var fileCfg UserConfig
		if yaml.Unmarshal(data, &fileCfg) == nil {
			resolved.UseAgent = ConfigValue[bool]{Value: fileCfg.UseAgent, Source: SourceFile}
		}
	}

	if envVal := os.Getenv("GOGWS_CONFIG_USE_AGENT"); envVal != "" {
		if boolVal, err := strconv.ParseBool(envVal); err == nil {
			resolved.UseAgent = ConfigValue[bool]{Value: boolVal, Source: SourceEnv}
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
		UseAgent: resolved.UseAgent.Value,
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
	case "use-agent":
		if v, ok := value.(bool); ok {
			cfg.UseAgent = v
		}
	}

	return SaveUserConfig(cfg)
}

func GetUserConfigValue(key string) (interface{}, error) {
	cfg, err := LoadUserConfig()
	if err != nil {
		return nil, err
	}

	switch key {
	case "use-agent":
		return cfg.UseAgent, nil
	default:
		return nil, nil
	}
}

func GetAvailableConfigKeys() []string {
	return []string{"use-agent"}
}

func GetEnvVarName(key string) string {
	switch key {
	case "use-agent":
		return "GOGWS_CONFIG_USE_AGENT"
	default:
		return ""
	}
}

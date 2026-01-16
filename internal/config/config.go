package config

import (
	"sync"

	"gogws/internal/git"
	"gogws/internal/gws"
	"gogws/internal/log"
)

type Config struct {
	WorkspaceRoot string
	HasProjects   bool
	HasWorkspaces bool
	ProjectsFile  string
	IgnoreFile    string
	ThemeFile     string
	Parallel      int
	Format        string
	NoColor       bool
	OnlyChanges   bool
	TUI           bool
	UseAgent      bool
	Hooks         HooksConfig
}

type HooksConfig struct {
	PreInit    string
	PostInit   string
	PreUpdate  string
	PostUpdate string
	PreClone   string
	PostClone  string
	PreFetch   string
	PostFetch  string
	PreFF      string
	PostFF     string
	PreCheck   string
	PostCheck  string
}

var (
	globalConfig *Config
	configMu     sync.RWMutex
	initialized  bool
)

func Initialize() error {
	configMu.Lock()
	defer configMu.Unlock()

	if initialized {
		return nil
	}

	log.Debug("Initializing configuration manager...")

	cfg, err := load()
	if err != nil {
		return err
	}

	globalConfig = cfg
	git.SetUseAgent(cfg.UseAgent)
	initialized = true

	log.Debug("Configuration loaded", "workspaceRoot", cfg.WorkspaceRoot, "hasProjects", cfg.HasProjects, "hasWorkspaces", cfg.HasWorkspaces, "useAgent", cfg.UseAgent, "parallel", cfg.Parallel)

	return nil
}

func GetConfig() *Config {
	configMu.RLock()
	defer configMu.RUnlock()
	return globalConfig
}

func MustGetConfig() *Config {
	cfg := GetConfig()
	if cfg == nil {
		panic("config not initialized - call Initialize() first")
	}
	return cfg
}

func IsInitialized() bool {
	configMu.RLock()
	defer configMu.RUnlock()
	return initialized
}

func Reload() error {
	configMu.Lock()
	initialized = false
	configMu.Unlock()
	return Initialize()
}

func ApplyFlags(themeFile string, parallel int, format string, noColor, onlyChanges, tui bool) {
	configMu.Lock()
	defer configMu.Unlock()

	if globalConfig == nil {
		return
	}

	if themeFile != "" {
		globalConfig.ThemeFile = themeFile
	}
	if parallel > 0 {
		globalConfig.Parallel = parallel
	}
	if format != "" {
		globalConfig.Format = format
	}
	globalConfig.NoColor = noColor
	globalConfig.OnlyChanges = onlyChanges
	globalConfig.TUI = tui

	git.SetUseAgent(globalConfig.UseAgent)
}

func load() (*Config, error) {
	cfg := &Config{
		ProjectsFile: gws.ProjectsFileName,
		IgnoreFile:   gws.IgnoreFileName,
		Parallel:     gws.DefaultParallel,
		Format:       "text",
		UseAgent:     true,
	}

	resolved, err := LoadUserConfigResolved()
	if err == nil && resolved != nil {
		cfg.UseAgent = resolved.UseAgent.Value
		log.Debug("User config loaded", "use-agent", resolved.UseAgent.Value, "source", resolved.UseAgent.Source)
	}

	wsInfo, err := gws.FindWorkspaceRoot()
	if err != nil {
		return nil, err
	}
	cfg.WorkspaceRoot = wsInfo.Root
	cfg.HasProjects = wsInfo.HasProjectsFile
	cfg.HasWorkspaces = wsInfo.HasWorkspacesFile

	return cfg, nil
}

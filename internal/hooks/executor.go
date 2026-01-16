package hooks

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"gogws/internal/config"
)

type Context struct {
	Command       string
	WorkspaceRoot string
	Projects      []string
	Data          map[string]interface{}
}

func Execute(cfg *config.Config, hookPath string, ctx Context) error {
	if hookPath == "" {
		return nil
	}

	absPath := hookPath
	if !filepath.IsAbs(hookPath) {
		absPath = filepath.Join(cfg.WorkspaceRoot, hookPath)
	}

	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		return nil
	}

	cmd := exec.Command(absPath)
	cmd.Dir = cfg.WorkspaceRoot
	cmd.Env = append(os.Environ(),
		fmt.Sprintf("GOGWS_COMMAND=%s", ctx.Command),
		fmt.Sprintf("GOGWS_WORKSPACE=%s", ctx.WorkspaceRoot),
	)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

func PreInit(cfg *config.Config) error {
	return Execute(cfg, cfg.Hooks.PreInit, Context{
		Command:       "init",
		WorkspaceRoot: cfg.WorkspaceRoot,
	})
}

func PostInit(cfg *config.Config, projects []string) error {
	return Execute(cfg, cfg.Hooks.PostInit, Context{
		Command:       "init",
		WorkspaceRoot: cfg.WorkspaceRoot,
		Projects:      projects,
	})
}

func PreUpdate(cfg *config.Config) error {
	return Execute(cfg, cfg.Hooks.PreUpdate, Context{
		Command:       "update",
		WorkspaceRoot: cfg.WorkspaceRoot,
	})
}

func PostUpdate(cfg *config.Config, cloned []string) error {
	return Execute(cfg, cfg.Hooks.PostUpdate, Context{
		Command:       "update",
		WorkspaceRoot: cfg.WorkspaceRoot,
		Projects:      cloned,
	})
}

func PreClone(cfg *config.Config, repoPath string) error {
	return Execute(cfg, cfg.Hooks.PreClone, Context{
		Command:       "clone",
		WorkspaceRoot: cfg.WorkspaceRoot,
		Projects:      []string{repoPath},
	})
}

func PostClone(cfg *config.Config, repoPath string, success bool) error {
	return Execute(cfg, cfg.Hooks.PostClone, Context{
		Command:       "clone",
		WorkspaceRoot: cfg.WorkspaceRoot,
		Projects:      []string{repoPath},
		Data: map[string]interface{}{
			"success": success,
		},
	})
}

func PreFetch(cfg *config.Config) error {
	return Execute(cfg, cfg.Hooks.PreFetch, Context{
		Command:       "fetch",
		WorkspaceRoot: cfg.WorkspaceRoot,
	})
}

func PostFetch(cfg *config.Config, fetched int) error {
	return Execute(cfg, cfg.Hooks.PostFetch, Context{
		Command:       "fetch",
		WorkspaceRoot: cfg.WorkspaceRoot,
		Data: map[string]interface{}{
			"fetched": fetched,
		},
	})
}

func PreFF(cfg *config.Config) error {
	return Execute(cfg, cfg.Hooks.PreFF, Context{
		Command:       "ff",
		WorkspaceRoot: cfg.WorkspaceRoot,
	})
}

func PostFF(cfg *config.Config, pulled int) error {
	return Execute(cfg, cfg.Hooks.PostFF, Context{
		Command:       "ff",
		WorkspaceRoot: cfg.WorkspaceRoot,
		Data: map[string]interface{}{
			"pulled": pulled,
		},
	})
}

func PreCheck(cfg *config.Config) error {
	return Execute(cfg, cfg.Hooks.PreCheck, Context{
		Command:       "check",
		WorkspaceRoot: cfg.WorkspaceRoot,
	})
}

func PostCheck(cfg *config.Config, unknown []string) error {
	return Execute(cfg, cfg.Hooks.PostCheck, Context{
		Command:       "check",
		WorkspaceRoot: cfg.WorkspaceRoot,
		Projects:      unknown,
	})
}

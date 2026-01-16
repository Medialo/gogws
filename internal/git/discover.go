package git

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-git/go-git/v5"
)

type DiscoveredRepo struct {
	Path    string
	Remotes []Remote
}

func DiscoverRepositories(rootPath string, maxDepth int) ([]DiscoveredRepo, error) {
	var repos []DiscoveredRepo

	err := filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() {
			return nil
		}

		relPath, err := filepath.Rel(rootPath, path)
		if err != nil {
			return err
		}

		depth := len(strings.Split(relPath, string(os.PathSeparator)))
		if relPath != "." && depth > maxDepth {
			return filepath.SkipDir
		}

		gitDir := filepath.Join(path, ".git")
		if _, err := os.Stat(gitDir); err == nil {
			if relPath == "." {
				return nil
			}

			repo, err := git.PlainOpen(path)
			if err != nil {
				return filepath.SkipDir
			}

			remotes, err := repo.Remotes()
			if err != nil || len(remotes) == 0 {
				return filepath.SkipDir
			}

			discovered := DiscoveredRepo{
				Path:    relPath,
				Remotes: make([]Remote, 0),
			}

			for _, remote := range remotes {
				cfg := remote.Config()
				if len(cfg.URLs) > 0 {
					discovered.Remotes = append(discovered.Remotes, Remote{
						Name: cfg.Name,
						URL:  cfg.URLs[0],
					})
				}
			}

			if len(discovered.Remotes) > 0 {
				repos = append(repos, discovered)
			}

			return filepath.SkipDir
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to discover repositories: %w", err)
	}

	return repos, nil
}

func FindUnknownRepositories(rootPath string, knownPaths []string) ([]string, error) {
	allRepos, err := DiscoverRepositories(rootPath, 10)
	if err != nil {
		return nil, err
	}

	known := make(map[string]bool)
	for _, path := range knownPaths {
		known[path] = true
	}

	var unknown []string
	for _, repo := range allRepos {
		if !known[repo.Path] {
			unknown = append(unknown, repo.Path)
		}
	}

	return unknown, nil
}

package git

import (
	"fmt"
	"os/exec"
	"path/filepath"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
)

type Remote struct {
	Name string
	URL  string
}

func Clone(targetPath string, remotes []Remote) error {
	if !useAgent {
		return cloneExec(targetPath, remotes)
	}
	return cloneLib(targetPath, remotes)
}

func cloneLib(targetPath string, remotes []Remote) error {
	if len(remotes) == 0 {
		return fmt.Errorf("no remotes defined")
	}

	primaryRemote := remotes[0]

	auth, err := getAuthMethod(primaryRemote.URL)
	if err != nil {
		return fmt.Errorf("failed to setup authentication: %w", err)
	}

	repo, err := git.PlainClone(targetPath, false, &git.CloneOptions{
		URL:          primaryRemote.URL,
		RemoteName:   primaryRemote.Name,
		Progress:     nil,
		SingleBranch: false,
		Auth:         auth,
	})
	if err != nil {
		return fmt.Errorf("failed to clone repository: %w", err)
	}

	for i := 1; i < len(remotes); i++ {
		remote := remotes[i]
		_, err := repo.CreateRemote(&config.RemoteConfig{
			Name: remote.Name,
			URLs: []string{remote.URL},
		})
		if err != nil {
			return fmt.Errorf("failed to add remote %s: %w", remote.Name, err)
		}
	}

	return nil
}

func cloneExec(targetPath string, remotes []Remote) error {
	if len(remotes) == 0 {
		return fmt.Errorf("no remotes defined")
	}

	primaryRemote := remotes[0]

	cmd := exec.Command("git", "clone", primaryRemote.URL, targetPath)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to clone repository: %s", string(output))
	}

	for i := 1; i < len(remotes); i++ {
		remote := remotes[i]
		cmd := exec.Command("git", "remote", "add", remote.Name, remote.URL)
		cmd.Dir = targetPath
		if output, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("failed to add remote %s: %s", remote.Name, string(output))
		}
	}

	return nil
}

func CloneWorkspace(workspaceRoot string, path string, remotes []Remote) error {
	targetPath := filepath.Join(workspaceRoot, path)
	return Clone(targetPath, remotes)
}

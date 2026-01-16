package git

import (
	"fmt"
	"os/exec"
	"path/filepath"
)

type Remote struct {
	Name string
	URL  string
}

func Clone(targetPath string, remotes []Remote) error {
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

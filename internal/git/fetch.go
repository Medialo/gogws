package git

import (
	"fmt"
	"log/slog"
	"os/exec"
)

func Fetch(repoPath string) error {
	slog.Debug("Fetching repository", "path", repoPath)

	cmd := exec.Command("git", "fetch", "--all")
	cmd.Dir = repoPath
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to fetch: %s", string(output))
	}

	return nil
}

func Pull(repoPath string) error {
	cmd := exec.Command("git", "pull", "--ff-only")
	cmd.Dir = repoPath
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to pull: %s", string(output))
	}

	return nil
}

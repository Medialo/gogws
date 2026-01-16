package git

import (
	"fmt"
	"os/exec"

	"github.com/go-git/go-git/v5"
)

func Fetch(repoPath string) error {
	if !useAgent {
		return fetchExec(repoPath)
	}
	return fetchLib(repoPath)
}

func fetchLib(repoPath string) error {
	repo, err := git.PlainOpen(repoPath)
	if err != nil {
		return fmt.Errorf("failed to open repository: %w", err)
	}

	remote, err := repo.Remote("origin")
	if err != nil {
		return fmt.Errorf("failed to get remote: %w", err)
	}

	refSpecs := remote.Config().Fetch
	if len(refSpecs) == 0 {
		return fmt.Errorf("no fetch refspecs configured")
	}

	auth, _ := getAuthMethod(remote.Config().URLs[0])

	err = repo.Fetch(&git.FetchOptions{
		RemoteName: "origin",
		Progress:   nil,
		Auth:       auth,
	})
	if err != nil && err != git.NoErrAlreadyUpToDate {
		return fmt.Errorf("failed to fetch: %w", err)
	}

	return nil
}

func fetchExec(repoPath string) error {
	cmd := exec.Command("git", "fetch", "--all")
	cmd.Dir = repoPath
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to fetch: %s", string(output))
	}

	return nil
}

func Pull(repoPath string) error {
	if !useAgent {
		return pullExec(repoPath)
	}
	return pullLib(repoPath)
}

func pullLib(repoPath string) error {
	repo, err := git.PlainOpen(repoPath)
	if err != nil {
		return fmt.Errorf("failed to open repository: %w", err)
	}

	wt, err := repo.Worktree()
	if err != nil {
		return fmt.Errorf("failed to get worktree: %w", err)
	}

	remote, err := repo.Remote("origin")
	if err != nil {
		return fmt.Errorf("failed to get remote: %w", err)
	}

	auth, _ := getAuthMethod(remote.Config().URLs[0])

	err = wt.Pull(&git.PullOptions{
		RemoteName: "origin",
		Progress:   nil,
		Force:      false,
		Auth:       auth,
	})

	if err != nil && err != git.NoErrAlreadyUpToDate {
		return fmt.Errorf("failed to pull: %w", err)
	}

	return nil
}

func pullExec(repoPath string) error {
	cmd := exec.Command("git", "pull", "--ff-only")
	cmd.Dir = repoPath
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to pull: %s", string(output))
	}

	return nil
}

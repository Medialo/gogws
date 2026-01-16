package git

import (
	"bufio"
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
)

func GetStatus(repoPath string) RepositoryStatus {
	if !useAgent {
		return getStatusExec(repoPath)
	}
	return getStatusLib(repoPath)
}

func GetStatusDetailed(repoPath string) RepositoryStatus {
	return getStatusExecDetailed(repoPath)
}

func getStatusLib(repoPath string) RepositoryStatus {
	status := RepositoryStatus{
		Path:   repoPath,
		Exists: false,
	}

	repo, err := git.PlainOpen(repoPath)
	if err != nil {
		if err == git.ErrRepositoryNotExists {
			return status
		}
		status.Error = err
		return status
	}

	status.Exists = true

	remotes, err := repo.Remotes()
	if err == nil {
		status.HasRemote = len(remotes) > 0
	}

	head, err := repo.Head()
	if err != nil {
		status.Error = err
		return status
	}

	status.Branch = head.Name().Short()

	wt, err := repo.Worktree()
	if err != nil {
		status.Error = err
		return status
	}

	wtStatus, err := wt.Status()
	if err != nil {
		status.Error = err
		return status
	}

	status.Clean = wtStatus.IsClean()

	uncommitted := 0
	untracked := 0
	for _, fileStatus := range wtStatus {
		if fileStatus.Worktree == git.Untracked {
			untracked++
		} else if fileStatus.Worktree != git.Unmodified || fileStatus.Staging != git.Unmodified {
			uncommitted++
		}
	}
	status.Uncommitted = uncommitted
	status.Untracked = untracked

	if status.HasRemote {
		ahead, behind, err := getAheadBehind(repo, head)
		if err == nil {
			status.Ahead = ahead
			status.Behind = behind
		}
	}

	return status
}

func getStatusExec(repoPath string) RepositoryStatus {
	status := RepositoryStatus{
		Path:   repoPath,
		Exists: false,
	}

	cmd := exec.Command("git", "rev-parse", "--git-dir")
	cmd.Dir = repoPath
	if err := cmd.Run(); err != nil {
		return status
	}
	status.Exists = true

	cmd = exec.Command("git", "remote")
	cmd.Dir = repoPath
	if output, err := cmd.Output(); err == nil {
		status.HasRemote = len(strings.TrimSpace(string(output))) > 0
	}

	cmd = exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	cmd.Dir = repoPath
	if output, err := cmd.Output(); err == nil {
		status.Branch = strings.TrimSpace(string(output))
	} else {
		status.Error = err
		return status
	}

	cmd = exec.Command("git", "status", "--porcelain")
	cmd.Dir = repoPath
	if output, err := cmd.Output(); err == nil {
		lines := strings.Split(string(output), "\n")
		uncommitted := 0
		untracked := 0
		for _, line := range lines {
			if len(line) < 2 {
				continue
			}
			if strings.HasPrefix(line, "??") {
				untracked++
			} else {
				uncommitted++
			}
		}
		status.Uncommitted = uncommitted
		status.Untracked = untracked
		status.Clean = uncommitted == 0 && untracked == 0
	}

	if status.HasRemote {
		cmd = exec.Command("git", "rev-list", "--left-right", "--count", fmt.Sprintf("%s...origin/%s", status.Branch, status.Branch))
		cmd.Dir = repoPath
		if output, err := cmd.Output(); err == nil {
			parts := strings.Fields(strings.TrimSpace(string(output)))
			if len(parts) == 2 {
				status.Ahead, _ = strconv.Atoi(parts[0])
				status.Behind, _ = strconv.Atoi(parts[1])
			}
		}
	}

	return status
}

func getStatusExecDetailed(repoPath string) RepositoryStatus {
	status := RepositoryStatus{
		Path:   repoPath,
		Exists: false,
	}

	cmd := exec.Command("git", "status", "--porcelain=v2", "--branch")
	cmd.Dir = repoPath
	output, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			if exitErr.ExitCode() == 128 {
				return status
			}
		}
		status.Error = err
		return status
	}

	status.Exists = true
	status.Clean = true

	scanner := bufio.NewScanner(strings.NewReader(string(output)))
	aheadBehindRegex := regexp.MustCompile(`\+(\d+) -(\d+)`)

	for scanner.Scan() {
		line := scanner.Text()

		if strings.HasPrefix(line, "# branch.head ") {
			status.Branch = strings.TrimPrefix(line, "# branch.head ")
		} else if strings.HasPrefix(line, "# branch.upstream ") {
			status.HasRemote = true
		} else if strings.HasPrefix(line, "# branch.ab ") {
			matches := aheadBehindRegex.FindStringSubmatch(line)
			if len(matches) == 3 {
				status.Ahead, _ = strconv.Atoi(matches[1])
				status.Behind, _ = strconv.Atoi(matches[2])
			}
		} else if strings.HasPrefix(line, "? ") {
			status.Untracked++
			status.Clean = false
		} else if strings.HasPrefix(line, "1 ") || strings.HasPrefix(line, "2 ") {
			status.Uncommitted++
			status.Clean = false
		}
	}

	return status
}

func getAheadBehind(repo *git.Repository, head *plumbing.Reference) (ahead, behind int, err error) {
	branch := head.Name().Short()

	remote, err := repo.Remote("origin")
	if err != nil {
		return 0, 0, err
	}

	refs, err := remote.List(&git.ListOptions{})
	if err != nil {
		return 0, 0, err
	}

	var remoteRef *plumbing.Reference
	remoteBranchName := plumbing.NewRemoteReferenceName("origin", branch)
	for _, ref := range refs {
		if ref.Name() == remoteBranchName {
			remoteRef = ref
			break
		}
	}

	if remoteRef == nil {
		return 0, 0, nil
	}

	localCommit, err := repo.CommitObject(head.Hash())
	if err != nil {
		return 0, 0, err
	}

	remoteCommit, err := repo.CommitObject(remoteRef.Hash())
	if err != nil {
		return 0, 0, err
	}

	isAncestor, err := localCommit.IsAncestor(remoteCommit)
	if err != nil {
		return 0, 0, err
	}

	if isAncestor {
		behind, err = countCommitsBetween(repo, localCommit.Hash, remoteCommit.Hash)
		if err != nil {
			return 0, 0, err
		}
		return 0, behind, nil
	}

	isAncestor, err = remoteCommit.IsAncestor(localCommit)
	if err != nil {
		return 0, 0, err
	}

	if isAncestor {
		ahead, err = countCommitsBetween(repo, remoteCommit.Hash, localCommit.Hash)
		if err != nil {
			return 0, 0, err
		}
		return ahead, 0, nil
	}

	return 0, 0, nil
}

func countCommitsBetween(repo *git.Repository, from, to plumbing.Hash) (int, error) {
	commits, err := repo.Log(&git.LogOptions{
		From: to,
	})
	if err != nil {
		return 0, err
	}

	count := 0
	err = commits.ForEach(func(c *object.Commit) error {
		if c.Hash == from {
			return fmt.Errorf("reached base")
		}
		count++
		return nil
	})

	if err != nil && err.Error() != "reached base" {
		return 0, err
	}

	return count, nil
}

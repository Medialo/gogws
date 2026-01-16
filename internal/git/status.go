package git

import (
	"bufio"
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

func GetStatus(repoPath string) RepositoryStatus {
	return getStatusExec(repoPath)
}

func GetStatusDetailed(repoPath string) RepositoryStatus {
	return getStatusExecDetailed(repoPath)
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

	branches, err := getBranches(repoPath)
	if err == nil {
		status.Branches = branches
		for _, b := range branches {
			if b.IsCurrent {
				status.Ahead = b.Ahead
				status.Behind = b.Behind
				break
			}
		}
	}

	return status
}

func getBranches(repoPath string) ([]BranchStatus, error) {
	cmd := exec.Command("git", "for-each-ref",
		"--format=%(refname:short)|%(upstream:short)|%(HEAD)",
		"refs/heads/")
	cmd.Dir = repoPath
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var branches []BranchStatus
	scanner := bufio.NewScanner(strings.NewReader(string(output)))

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		parts := strings.Split(line, "|")
		if len(parts) < 3 {
			continue
		}

		branchName := parts[0]
		upstream := parts[1]
		isCurrent := parts[2] == "*"

		branch := BranchStatus{
			Name:      branchName,
			IsCurrent: isCurrent,
			Upstream:  upstream,
		}

		if upstream != "" {
			ahead, behind := getAheadBehind(repoPath, branchName, upstream)
			branch.Ahead = ahead
			branch.Behind = behind
		}

		branches = append(branches, branch)
	}

	return branches, nil
}

func getAheadBehind(repoPath, branch, upstream string) (ahead, behind int) {
	cmd := exec.Command("git", "rev-list", "--left-right", "--count",
		fmt.Sprintf("%s...%s", branch, upstream))
	cmd.Dir = repoPath
	output, err := cmd.Output()
	if err != nil {
		return 0, 0
	}

	parts := strings.Fields(strings.TrimSpace(string(output)))
	if len(parts) == 2 {
		ahead, _ = strconv.Atoi(parts[0])
		behind, _ = strconv.Atoi(parts[1])
	}

	return ahead, behind
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

	branches, err := getBranches(repoPath)
	if err == nil {
		status.Branches = branches
	}

	return status
}

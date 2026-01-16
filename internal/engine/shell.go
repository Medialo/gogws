package engine

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"os/exec"
	"runtime"
	"time"
)

func ExecuteGit(repoPath string, args ...string) (stdout, stderr string, err error) {
	return ExecuteGitWithTimeout(repoPath, 0, args...)
}

func ExecuteGitWithTimeout(repoPath string, timeout time.Duration, args ...string) (stdout, stderr string, err error) {
	slog.Debug("Executing git command", "repoPath", repoPath, "args", args, "timeout", timeout)
	var ctx context.Context
	var cancel context.CancelFunc

	if timeout > 0 {
		ctx, cancel = context.WithTimeout(context.Background(), timeout)
		defer cancel()
	} else {
		ctx = context.Background()
	}

	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Dir = repoPath

	var stdoutBuf, stderrBuf bytes.Buffer
	cmd.Stdout = &stdoutBuf
	cmd.Stderr = &stderrBuf

	err = cmd.Run()

	stdout = stdoutBuf.String()
	stderr = stderrBuf.String()

	if ctx.Err() == context.DeadlineExceeded {
		err = fmt.Errorf("command timed out after %v", timeout)
	}

	return
}

func ExecuteShell(repoPath, command string) (stdout, stderr string, err error) {
	return ExecuteShellWithTimeout(repoPath, command, 0)
}

func ExecuteShellWithTimeout(repoPath, command string, timeout time.Duration) (stdout, stderr string, err error) {
	var ctx context.Context
	var cancel context.CancelFunc

	if timeout > 0 {
		ctx, cancel = context.WithTimeout(context.Background(), timeout)
		defer cancel()
	} else {
		ctx = context.Background()
	}

	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.CommandContext(ctx, "cmd", "/C", command)
	} else {
		cmd = exec.CommandContext(ctx, "sh", "-c", command)
	}
	cmd.Dir = repoPath

	var stdoutBuf, stderrBuf bytes.Buffer
	cmd.Stdout = &stdoutBuf
	cmd.Stderr = &stderrBuf

	err = cmd.Run()

	stdout = stdoutBuf.String()
	stderr = stderrBuf.String()

	if ctx.Err() == context.DeadlineExceeded {
		err = fmt.Errorf("command timed out after %v", timeout)
	}

	return
}

func executeCommand(cmd RepoCommand, timeout time.Duration) (stdout, stderr string, err error) {
	switch cmd.Type {
	case CommandTypeGit:
		return ExecuteGitWithTimeout(cmd.RepoPath, timeout, cmd.Args...)
	case CommandTypeShell:
		if len(cmd.Args) > 0 {
			return ExecuteShellWithTimeout(cmd.RepoPath, cmd.Args[0], timeout)
		}
		return "", "", fmt.Errorf("shell command requires at least one argument")
	case CommandTypeCustom:
		if cmd.Action != nil {
			output, err := cmd.Action()
			return output, "", err
		}
		return "", "", fmt.Errorf("custom command requires an action function")
	default:
		return "", "", fmt.Errorf("unknown command type: %d", cmd.Type)
	}
}

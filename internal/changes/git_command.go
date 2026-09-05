package changes

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

const (
	gitCommandTimeout = 30 * time.Second
	gitOutputLimit    = 16 * 1024 * 1024
)

var (
	errGitCommandTimeout = errors.New("git command timed out")
	errGitOutputLimit    = errors.New("git command output exceeded limit")
)

type limitedOutput struct {
	buffer    bytes.Buffer
	limit     int
	truncated bool
}

func (w *limitedOutput) Write(p []byte) (int, error) {
	written := len(p)
	remaining := w.limit - w.buffer.Len()
	if remaining <= 0 {
		w.truncated = written > 0
		return written, nil
	}
	if len(p) > remaining {
		p = p[:remaining]
		w.truncated = true
	}
	_, _ = w.buffer.Write(p)
	return written, nil
}

func (w *limitedOutput) String() string {
	return w.buffer.String()
}

type boundedCommandResult struct {
	stdout limitedOutput
	stderr limitedOutput
}

func runBoundedCommand(ctx context.Context, outputLimit int, name string, args ...string) (boundedCommandResult, error) {
	result := boundedCommandResult{
		stdout: limitedOutput{limit: outputLimit},
		stderr: limitedOutput{limit: outputLimit},
	}
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Stdout = &result.stdout
	cmd.Stderr = &result.stderr
	err := cmd.Run()
	return result, err
}

func executeGit(cwd string, args ...string) (boundedCommandResult, error) {
	ctx, cancel := context.WithTimeout(context.Background(), gitCommandTimeout)
	defer cancel()

	gitArgs := append([]string{"-C", cwd}, args...)
	result, err := runBoundedCommand(ctx, gitOutputLimit, "git", gitArgs...)
	if errors.Is(ctx.Err(), context.DeadlineExceeded) {
		return result, fmt.Errorf("%w after %s", errGitCommandTimeout, gitCommandTimeout)
	}
	if result.stdout.truncated || result.stderr.truncated {
		return result, fmt.Errorf("%w (%d bytes per stream)", errGitOutputLimit, gitOutputLimit)
	}
	return result, err
}

// IsGitRepository reports whether cwd is inside a Git worktree. Provider
// failures such as a missing executable, timeout, or excessive output are
// returned separately from the normal non-repository result.
func IsGitRepository(cwd string) (bool, error) {
	result, err := executeGit(cwd, "rev-parse", "--is-inside-work-tree")
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			return false, nil
		}
		return false, fmt.Errorf("%w: inspect repository: %v", ErrGitFailure, err)
	}
	return strings.TrimSpace(result.stdout.String()) == "true", nil
}

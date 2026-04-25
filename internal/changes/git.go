package changes

import (
	"errors"
	"fmt"
	"os/exec"
	"strings"

	"kyn/internal/matcher"
)

func fromGitDiff(cwd string, base string, head string) ([]string, error) {
	rangeSpec := fmt.Sprintf("%s...%s", base, head)
	cmd := exec.Command("git", "-C", cwd, "diff", "--name-status", "-M", rangeSpec)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("%w: %v: %s", ErrGitFailure, err, strings.TrimSpace(string(out)))
	}

	lines := strings.Split(string(out), "\n")
	paths := make([]string, 0, len(lines))
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		fields := strings.Split(line, "\t")
		if len(fields) < 2 {
			continue
		}

		status := fields[0]
		switch {
		case strings.HasPrefix(status, "A"), strings.HasPrefix(status, "M"):
			paths = append(paths, matcher.NormalizePath(fields[1]))
		case strings.HasPrefix(status, "R"):
			if len(fields) < 3 {
				continue
			}
			paths = append(paths, matcher.NormalizePath(fields[2]))
		case strings.HasPrefix(status, "D"):
			// Deleted files are intentionally excluded for MVP evaluation.
		default:
			// Keep behavior resilient for unexpected statuses by treating final path as changed.
			paths = append(paths, matcher.NormalizePath(fields[len(fields)-1]))
		}
	}

	if len(paths) == 0 {
		return nil, errors.New("git diff produced no changed files")
	}

	return paths, nil
}

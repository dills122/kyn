package changes

import (
	"errors"
	"fmt"
	"os/exec"
	"strings"

	"kyn/internal/matcher"
)

func fromGitDiff(cwd string, base string, head string) ([]Change, error) {
	rangeSpec := fmt.Sprintf("%s...%s", base, head)
	cmd := exec.Command("git", "-C", cwd, "diff", "--name-status", "-M", rangeSpec)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("%w: %v: %s", ErrGitFailure, err, strings.TrimSpace(string(out)))
	}

	lines := strings.Split(string(out), "\n")
	outChanges := make([]Change, 0, len(lines))
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
			normalized := matcher.NormalizePath(fields[1])
			if normalized == "" {
				continue
			}
			outChanges = append(outChanges, Change{
				Path:   normalized,
				Status: mapStatus(status),
			})
		case strings.HasPrefix(status, "R"):
			if len(fields) < 3 {
				continue
			}
			normalized := matcher.NormalizePath(fields[2])
			if normalized == "" {
				continue
			}
			outChanges = append(outChanges, Change{
				Path:   normalized,
				Status: StatusRenamed,
			})
		case strings.HasPrefix(status, "D"):
			// Deleted files are intentionally excluded for MVP evaluation.
		default:
			// Keep behavior resilient for unexpected statuses by treating final path as changed.
			normalized := matcher.NormalizePath(fields[len(fields)-1])
			if normalized == "" {
				continue
			}
			outChanges = append(outChanges, Change{
				Path:   normalized,
				Status: StatusModified,
			})
		}
	}

	if len(outChanges) == 0 {
		return nil, errors.New("git diff produced no changed files")
	}

	return outChanges, nil
}

func mapStatus(raw string) Status {
	switch {
	case strings.HasPrefix(raw, "A"):
		return StatusAdded
	case strings.HasPrefix(raw, "M"):
		return StatusModified
	case strings.HasPrefix(raw, "R"):
		return StatusRenamed
	case strings.HasPrefix(raw, "D"):
		return StatusDeleted
	default:
		return StatusModified
	}
}

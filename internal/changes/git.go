package changes

import (
	"fmt"
	"strings"

	"github.com/dills122/kyn/internal/matcher"
)

func fromGitDiff(cwd string, base string, head string) ([]Change, error) {
	rangeSpec := fmt.Sprintf("%s...%s", base, head)
	// -z prints every path literally, NUL-terminated. Without it git wraps a
	// path with non-ASCII or special characters in quotes and C escapes
	// ("caf\303\251.go"), which never matches a glob or exists on disk.
	result, err := executeGit(cwd, "diff", "--name-status", "-z", "-M", rangeSpec)
	if err != nil {
		detail := strings.TrimSpace(result.stderr.String())
		if detail == "" {
			return nil, fmt.Errorf("%w: %v", ErrGitFailure, err)
		}
		return nil, fmt.Errorf("%w: %v: %s", ErrGitFailure, err, detail)
	}

	tokens := strings.Split(result.stdout.String(), "\x00")
	outChanges := make([]Change, 0, len(tokens)/2)
	for i := 0; i < len(tokens); {
		status := tokens[i]
		i++
		if status == "" {
			continue
		}

		// Renames and copies carry two paths (old, then new); everything else one.
		paths := 1
		if strings.HasPrefix(status, "R") || strings.HasPrefix(status, "C") {
			paths = 2
		}
		if i+paths > len(tokens) {
			return nil, fmt.Errorf("%w: malformed git diff output near status %q", ErrGitFailure, status)
		}
		fields := tokens[i : i+paths]
		i += paths

		switch {
		case strings.HasPrefix(status, "D"):
			// Deleted files are intentionally excluded for MVP evaluation.
		case strings.HasPrefix(status, "R"):
			normalized, err := matcher.NormalizeRelativePath(fields[1])
			if err != nil {
				return nil, fmt.Errorf("%w: invalid renamed path: %v", ErrGitFailure, err)
			}
			if normalized == "" {
				continue
			}
			outChanges = append(outChanges, Change{
				Path:   normalized,
				Status: StatusRenamed,
			})
		default:
			// A, M, and any unexpected status: the final path is the changed one.
			normalized, err := matcher.NormalizeRelativePath(fields[len(fields)-1])
			if err != nil {
				return nil, fmt.Errorf("%w: invalid changed path: %v", ErrGitFailure, err)
			}
			if normalized == "" {
				continue
			}
			outChanges = append(outChanges, Change{
				Path:   normalized,
				Status: mapStatus(status),
			})
		}
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

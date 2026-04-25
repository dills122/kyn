package changes

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"kyn/internal/matcher"
)

func fromCSV(csv string) ([]string, error) {
	parts := strings.Split(csv, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		normalized := matcher.NormalizePath(p)
		if normalized == "" {
			continue
		}
		out = append(out, normalized)
	}
	return out, nil
}

func fromFile(cwd string, filePath string) ([]string, error) {
	p := filePath
	if !filepath.IsAbs(p) {
		p = filepath.Join(cwd, p)
	}

	f, err := os.Open(p)
	if err != nil {
		return nil, fmt.Errorf("open --files-from: %w", err)
	}
	defer f.Close()

	sc := bufio.NewScanner(f)
	out := make([]string, 0, 64)
	for sc.Scan() {
		line := sc.Text()
		normalized := matcher.NormalizePath(line)
		if normalized == "" {
			continue
		}
		out = append(out, normalized)
	}
	if err := sc.Err(); err != nil {
		return nil, fmt.Errorf("read --files-from: %w", err)
	}

	return out, nil
}

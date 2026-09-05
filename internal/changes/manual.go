package changes

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/dills122/kyn/internal/matcher"
)

func fromCSV(csv string) ([]Change, error) {
	parts := strings.Split(csv, ",")
	out := make([]Change, 0, len(parts))
	for _, p := range parts {
		normalized, err := matcher.NormalizeRelativePath(p)
		if err != nil {
			return nil, fmt.Errorf("invalid --files path: %w", err)
		}
		if normalized == "" {
			continue
		}
		out = append(out, Change{
			Path:   normalized,
			Status: StatusModified,
		})
	}
	return out, nil
}

func fromFile(cwd string, filePath string) ([]Change, error) {
	if filePath == "-" {
		return readList(os.Stdin, "stdin")
	}

	p := filePath
	if !filepath.IsAbs(p) {
		p = filepath.Join(cwd, p)
	}

	f, err := os.Open(p)
	if err != nil {
		return nil, fmt.Errorf("open --files-from: %w", err)
	}
	defer func() {
		_ = f.Close()
	}()

	return readList(f, "--files-from")
}

func readList(r io.Reader, source string) ([]Change, error) {
	sc := bufio.NewScanner(r)
	out := make([]Change, 0, 64)
	lineNumber := 0
	for sc.Scan() {
		lineNumber++
		line := sc.Text()
		normalized, err := matcher.NormalizeRelativePath(line)
		if err != nil {
			return nil, fmt.Errorf("read %s line %d: %w", source, lineNumber, err)
		}
		if normalized == "" {
			continue
		}
		out = append(out, Change{
			Path:   normalized,
			Status: StatusModified,
		})
	}
	if err := sc.Err(); err != nil {
		return nil, fmt.Errorf("read %s: %w", source, err)
	}

	return out, nil
}

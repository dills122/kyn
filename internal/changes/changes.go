package changes

import (
	"errors"
	"fmt"
	"sort"
)

var ErrGitFailure = errors.New("git provider failure")

type Input struct {
	Cwd       string
	FilesCSV  string
	FilesFrom string
	Base      string
	Head      string
}

func Collect(in Input) ([]string, error) {
	var files []string
	var err error

	switch {
	case in.FilesCSV != "":
		files, err = fromCSV(in.FilesCSV)
	case in.FilesFrom != "":
		files, err = fromFile(in.Cwd, in.FilesFrom)
	default:
		files, err = fromGitDiff(in.Cwd, in.Base, in.Head)
	}
	if err != nil {
		return nil, err
	}

	uniq := make(map[string]struct{}, len(files))
	out := make([]string, 0, len(files))
	for _, f := range files {
		if f == "" {
			continue
		}
		if _, exists := uniq[f]; exists {
			continue
		}
		uniq[f] = struct{}{}
		out = append(out, f)
	}

	sort.Strings(out)
	if len(out) == 0 {
		return nil, fmt.Errorf("no changed files were collected")
	}

	return out, nil
}

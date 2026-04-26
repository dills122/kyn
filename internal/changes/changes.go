package changes

import (
	"errors"
	"fmt"
	"sort"
)

var ErrGitFailure = errors.New("git provider failure")

type Status string

const (
	StatusAdded    Status = "added"
	StatusModified Status = "modified"
	StatusDeleted  Status = "deleted"
	StatusRenamed  Status = "renamed"
)

type Change struct {
	Path   string
	Status Status
}

type Result struct {
	Files        []string
	StatusByFile map[string]Status
}

type Input struct {
	Cwd       string
	FilesCSV  string
	FilesFrom string
	Base      string
	Head      string
}

func Collect(in Input) ([]string, error) {
	result, err := CollectDetailed(in)
	if err != nil {
		return nil, err
	}
	return result.Files, nil
}

func CollectDetailed(in Input) (Result, error) {
	var changes []Change
	var err error

	switch {
	case in.FilesCSV != "":
		changes, err = fromCSV(in.FilesCSV)
	case in.FilesFrom != "":
		changes, err = fromFile(in.Cwd, in.FilesFrom)
	default:
		changes, err = fromGitDiff(in.Cwd, in.Base, in.Head)
	}
	if err != nil {
		return Result{}, err
	}

	statusByFile := make(map[string]Status, len(changes))
	out := make([]string, 0, len(changes))
	for _, c := range changes {
		if c.Path == "" {
			continue
		}
		if _, exists := statusByFile[c.Path]; exists {
			continue
		}
		statusByFile[c.Path] = c.Status
		out = append(out, c.Path)
	}

	sort.Strings(out)
	if len(out) == 0 {
		return Result{}, fmt.Errorf("no changed files were collected")
	}

	return Result{
		Files:        out,
		StatusByFile: statusByFile,
	}, nil
}

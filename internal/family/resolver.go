package family

import (
	"fmt"
	"path"
	"sort"
	"strings"

	"kyn/internal/config"
	"kyn/internal/matcher"
)

func Resolve(cfg config.Config, changedFiles []string) ([]Instance, error) {
	type acc struct {
		familyID string
		name     string
		sources  map[string]struct{}
		kin      map[string]string
	}

	instances := make(map[string]*acc)

	for _, fam := range cfg.Families {
		for _, file := range changedFiles {
			file = matcher.NormalizePath(file)
			if file == "" {
				continue
			}

			include, err := matcher.MatchAny(fam.Include, file)
			if err != nil {
				return nil, fmt.Errorf("family %q include match failed for %q: %w", fam.ID, file, err)
			}
			if !include {
				continue
			}

			if len(fam.Exclude) > 0 {
				excluded, err := matcher.MatchAny(fam.Exclude, file)
				if err != nil {
					return nil, fmt.Errorf("family %q exclude match failed for %q: %w", fam.ID, file, err)
				}
				if excluded {
					continue
				}
			}

			ctx := buildTemplateContext(file, fam.BaseName.StripSuffixes)
			instanceName := ctx.Base
			if ctx.Dir != "" {
				instanceName = ctx.Dir + "/" + ctx.Base
			}
			instanceName = matcher.NormalizePath(instanceName)

			key := fam.ID + "|" + instanceName
			a, ok := instances[key]
			if !ok {
				a = &acc{
					familyID: fam.ID,
					name:     instanceName,
					sources:  map[string]struct{}{},
					kin:      map[string]string{},
				}
				for kinName, kinTemplate := range fam.Kin {
					a.kin[kinName] = resolveTemplate(kinTemplate, ctx)
				}
				instances[key] = a
			}
			a.sources[file] = struct{}{}
		}
	}

	out := make([]Instance, 0, len(instances))
	for _, a := range instances {
		sourceFiles := make([]string, 0, len(a.sources))
		for f := range a.sources {
			sourceFiles = append(sourceFiles, f)
		}
		sort.Strings(sourceFiles)

		out = append(out, Instance{
			FamilyID:    a.familyID,
			Name:        a.name,
			SourceFiles: sourceFiles,
			Kin:         a.kin,
		})
	}

	sort.Slice(out, func(i, j int) bool {
		if out[i].FamilyID == out[j].FamilyID {
			return out[i].Name < out[j].Name
		}
		return out[i].FamilyID < out[j].FamilyID
	})

	return out, nil
}

func buildTemplateContext(file string, stripSuffixes []string) templateContext {
	file = matcher.NormalizePath(file)
	dir := path.Dir(file)
	if dir == "." {
		dir = ""
	}
	filename := path.Base(file)
	ext := path.Ext(filename)
	name := strings.TrimSuffix(filename, ext)
	base := name
	for _, suffix := range stripSuffixes {
		base = strings.TrimSuffix(base, suffix)
	}

	return templateContext{
		Dir:  dir,
		File: file,
		Name: name,
		Base: base,
		Ext:  ext,
	}
}

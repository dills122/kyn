package family

import (
	"fmt"
	"path"
	"sort"
	"strings"

	"github.com/dills122/kyn/internal/config"
	"github.com/dills122/kyn/internal/matcher"
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
		includePatterns := fam.SourceInclude()
		excludePatterns := fam.SourceExclude()
		for _, file := range changedFiles {
			file, err := matcher.NormalizeRelativePath(file)
			if err != nil {
				return nil, fmt.Errorf("family %q received invalid changed path: %w", fam.ID, err)
			}
			if file == "" {
				continue
			}

			include, err := matcher.MatchAny(includePatterns, file)
			if err != nil {
				return nil, fmt.Errorf("family %q include match failed for %q: %w", fam.ID, file, err)
			}
			if !include {
				continue
			}

			if len(excludePatterns) > 0 {
				excluded, err := matcher.MatchAny(excludePatterns, file)
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
				kinNames := make([]string, 0, len(fam.Kin))
				for kinName := range fam.Kin {
					kinNames = append(kinNames, kinName)
				}
				sort.Strings(kinNames)
				for _, kinName := range kinNames {
					kinTemplate := fam.Kin[kinName]
					resolved := resolveTemplate(kinTemplate, ctx)
					normalized, err := matcher.NormalizeRelativePath(resolved)
					if err != nil {
						return nil, fmt.Errorf("family %q kin %q resolved unsafe path: %w", fam.ID, kinName, err)
					}
					a.kin[kinName] = normalized
				}
				instances[key] = a
			} else if err := checkKinAgreement(fam, ctx, file, instanceName, a.kin); err != nil {
				return nil, err
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

// checkKinAgreement guards against kin templates whose resolved path would
// depend on which source file happened to be the first one to create this
// family instance. An instance's dedup key is family+{dir}/{base}, so {dir}
// and {base} are already guaranteed identical across every source file in
// the instance — only a template using {ext} (or {file}/{name}, which embed
// the extension too) can disagree once a second source file with a
// different extension joins the same instance. Rather than silently keeping
// whichever file happened to resolve the instance first (order-dependent on
// the sorted changed-file list), fail with an actionable error.
func checkKinAgreement(fam config.Family, ctx templateContext, file string, instanceName string, existing map[string]string) error {
	for kinName, kinTemplate := range fam.Kin {
		resolved := resolveTemplate(kinTemplate, ctx)
		normalized, err := matcher.NormalizeRelativePath(resolved)
		if err != nil {
			return fmt.Errorf("family %q kin %q resolved unsafe path: %w", fam.ID, kinName, err)
		}
		if prior, ok := existing[kinName]; ok && normalized != prior {
			return fmt.Errorf(
				"family %q instance %q: kin %q template %q resolves to different paths for different source files in this instance (%q from %q vs %q); "+
					"this usually means the template uses {ext}, {file}, or {name} and the instance's source files have different extensions",
				fam.ID, instanceName, kinName, kinTemplate, normalized, file, prior,
			)
		}
	}
	return nil
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

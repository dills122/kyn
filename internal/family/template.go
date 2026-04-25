package family

import (
	"strings"

	"kyn/internal/matcher"
)

type templateContext struct {
	Dir  string
	File string
	Name string
	Base string
	Ext  string
}

func resolveTemplate(tpl string, ctx templateContext) string {
	dir := ctx.Dir
	if dir == "" {
		dir = "."
	}
	replacer := strings.NewReplacer(
		"{dir}", dir,
		"{file}", ctx.File,
		"{name}", ctx.Name,
		"{base}", ctx.Base,
		"{ext}", ctx.Ext,
	)
	return matcher.NormalizePath(replacer.Replace(tpl))
}

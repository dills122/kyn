package config

import (
	"slices"
	"testing"
)

func TestFamilySourceHelpers(t *testing.T) {
	fam := Family{
		Include: []string{"legacy/**/*.go"},
		Exclude: []string{"legacy/**/*.gen.go"},
		Groups: GroupMap{
			"source": {
				Include: []string{"src/**/*.go"},
				Exclude: []string{"src/**/*.gen.go"},
			},
		},
	}

	if !slices.Equal(fam.SourceInclude(), []string{"src/**/*.go"}) {
		t.Fatalf("unexpected source include: %v", fam.SourceInclude())
	}
	if !slices.Equal(fam.SourceExclude(), []string{"src/**/*.gen.go"}) {
		t.Fatalf("unexpected source exclude: %v", fam.SourceExclude())
	}

	fam.Groups["source"] = GroupDef{}
	if !slices.Equal(fam.SourceInclude(), []string{"legacy/**/*.go"}) {
		t.Fatalf("expected fallback include, got %v", fam.SourceInclude())
	}
	if !slices.Equal(fam.SourceExclude(), []string{"legacy/**/*.gen.go"}) {
		t.Fatalf("expected fallback exclude, got %v", fam.SourceExclude())
	}
}

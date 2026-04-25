package matcher

import "testing"

func TestMatchAny(t *testing.T) {
	tests := []struct {
		name     string
		patterns []string
		path     string
		want     bool
	}{
		{
			name:     "matches doublestar pattern",
			patterns: []string{"libs/**/*.component.ts"},
			path:     "libs/ui/button/button.component.ts",
			want:     true,
		},
		{
			name:     "does not match",
			patterns: []string{"libs/**/*.stories.ts"},
			path:     "libs/ui/button/button.component.ts",
			want:     false,
		},
		{
			name:     "matches one of many",
			patterns: []string{"**/*.md", "libs/**/*.component.ts"},
			path:     "libs/ui/button/button.component.ts",
			want:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := MatchAny(tt.patterns, tt.path)
			if err != nil {
				t.Fatalf("MatchAny returned error: %v", err)
			}
			if got != tt.want {
				t.Fatalf("MatchAny=%v, want %v", got, tt.want)
			}
		})
	}
}

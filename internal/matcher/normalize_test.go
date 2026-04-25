package matcher

import "testing"

func TestNormalizePath(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{in: " ./libs\\ui\\button.ts ", want: "libs/ui/button.ts"},
		{in: "a/../b/file.ts", want: "b/file.ts"},
		{in: ".", want: ""},
		{in: "", want: ""},
	}

	for _, tt := range tests {
		got := NormalizePath(tt.in)
		if got != tt.want {
			t.Fatalf("NormalizePath(%q)=%q, want %q", tt.in, got, tt.want)
		}
	}
}

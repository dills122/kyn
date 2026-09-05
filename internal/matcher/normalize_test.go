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

func TestNormalizeRelativePath(t *testing.T) {
	tests := []struct {
		name    string
		in      string
		want    string
		wantErr bool
	}{
		{name: "slash normalized", in: ` ./libs\ui\button.ts `, want: "libs/ui/button.ts"},
		{name: "internal clean", in: "a/../b/file.ts", want: "b/file.ts"},
		{name: "empty", in: ".", want: ""},
		{name: "parent", in: "../outside.ts", wantErr: true},
		{name: "cleaned parent", in: "a/../../outside.ts", wantErr: true},
		{name: "absolute", in: "/tmp/outside.ts", wantErr: true},
		{name: "windows drive", in: `C:\tmp\outside.ts`, wantErr: true},
		{name: "nul byte", in: "src/a\x00.go", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NormalizeRelativePath(tt.in)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("NormalizeRelativePath(%q) returned nil error", tt.in)
				}
				return
			}
			if err != nil {
				t.Fatalf("NormalizeRelativePath(%q) returned error: %v", tt.in, err)
			}
			if got != tt.want {
				t.Fatalf("NormalizeRelativePath(%q)=%q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

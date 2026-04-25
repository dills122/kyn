package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestValidateCheckOptions(t *testing.T) {
	valid := checkOptions{
		Format: "text",
		FailOn: "error",
	}

	tests := []struct {
		name    string
		modify  func(*checkOptions)
		wantErr bool
	}{
		{
			name: "valid files mode",
			modify: func(o *checkOptions) {
				o.FilesCSV = "a.ts"
			},
		},
		{
			name: "valid files-from mode",
			modify: func(o *checkOptions) {
				o.FilesFrom = "changed.txt"
			},
		},
		{
			name: "valid git mode",
			modify: func(o *checkOptions) {
				o.Base = "origin/main"
				o.Head = "HEAD"
			},
		},
		{
			name: "valid stdin mode",
			modify: func(o *checkOptions) {
				o.Stdin = true
			},
		},
		{
			name:    "invalid no mode",
			wantErr: true,
		},
		{
			name: "invalid multiple modes",
			modify: func(o *checkOptions) {
				o.FilesCSV = "a.ts"
				o.FilesFrom = "changed.txt"
			},
			wantErr: true,
		},
		{
			name: "invalid stdin plus files-from",
			modify: func(o *checkOptions) {
				o.Stdin = true
				o.FilesFrom = "-"
			},
			wantErr: true,
		},
		{
			name: "invalid mixed files and git includes selected modes",
			modify: func(o *checkOptions) {
				o.FilesCSV = "a.ts"
				o.Base = "origin/main"
				o.Head = "HEAD"
			},
			wantErr: true,
		},
		{
			name: "invalid partial git mode",
			modify: func(o *checkOptions) {
				o.Base = "origin/main"
			},
			wantErr: true,
		},
		{
			name: "invalid format",
			modify: func(o *checkOptions) {
				o.FilesCSV = "a.ts"
				o.Format = "xml"
			},
			wantErr: true,
		},
		{
			name: "invalid fail-on",
			modify: func(o *checkOptions) {
				o.FilesCSV = "a.ts"
				o.FailOn = "info"
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := valid
			if tt.modify != nil {
				tt.modify(&o)
			}

			err := validateCheckOptions(o)
			if tt.wantErr && err == nil {
				t.Fatal("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("expected no error, got %v", err)
			}
			if tt.name == "invalid mixed files and git includes selected modes" && err != nil {
				if !strings.Contains(err.Error(), "selected: files, git") {
					t.Fatalf("expected selected mode details, got %v", err)
				}
			}
		})
	}
}

func TestResolveCWD(t *testing.T) {
	t.Run("valid dir", func(t *testing.T) {
		dir := t.TempDir()
		got, err := resolveCWD(dir)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != dir {
			t.Fatalf("expected %s, got %s", dir, got)
		}
	})

	t.Run("not directory", func(t *testing.T) {
		dir := t.TempDir()
		f := filepath.Join(dir, "file.txt")
		if err := os.WriteFile(f, []byte("x"), 0o600); err != nil {
			t.Fatalf("write file: %v", err)
		}
		_, err := resolveCWD(f)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}

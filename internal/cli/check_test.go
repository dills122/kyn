package cli

import (
	"os"
	"os/exec"
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
			name: "valid sarif format for check",
			modify: func(o *checkOptions) {
				o.FilesCSV = "a.ts"
				o.Format = "sarif"
			},
		},
		{
			name: "valid rdjson format for check",
			modify: func(o *checkOptions) {
				o.FilesCSV = "a.ts"
				o.Format = "rdjson"
			},
		},
		{
			name: "valid checkstyle format for check",
			modify: func(o *checkOptions) {
				o.FilesCSV = "a.ts"
				o.Format = "checkstyle"
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
			wantErr: false,
		},
		{
			name: "invalid no mode when strict",
			modify: func(o *checkOptions) {
				o.StrictInput = true
			},
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
			name: "invalid sarif format for explain",
			modify: func(o *checkOptions) {
				o.FilesCSV = "a.ts"
				o.Format = "sarif"
			},
			wantErr: true,
		},
		{
			name: "invalid rdjson format for explain",
			modify: func(o *checkOptions) {
				o.FilesCSV = "a.ts"
				o.Format = "rdjson"
			},
			wantErr: true,
		},
		{
			name: "invalid checkstyle format for explain",
			modify: func(o *checkOptions) {
				o.FilesCSV = "a.ts"
				o.Format = "checkstyle"
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

			command := "check"
			allowMachineFormats := true
			if tt.name == "invalid sarif format for explain" || tt.name == "invalid rdjson format for explain" || tt.name == "invalid checkstyle format for explain" {
				command = "explain"
				allowMachineFormats = false
			}

			err := validateCheckOptions(o, command, allowMachineFormats)
			if tt.wantErr && err == nil {
				t.Fatal("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("expected no error, got %v", err)
			}
			if tt.name == "invalid mixed files and git includes selected modes" && err != nil {
				if !strings.Contains(err.Error(), "files + git") {
					t.Fatalf("expected selected mode details, got %v", err)
				}
			}
			if tt.name == "invalid partial git mode" && err != nil {
				if !strings.Contains(err.Error(), "expected both --base and --head") {
					t.Fatalf("expected expected-vs-observed message, got %v", err)
				}
			}
			if (tt.name == "invalid sarif format for explain" || tt.name == "invalid rdjson format for explain" || tt.name == "invalid checkstyle format for explain") && err != nil {
				if !strings.Contains(err.Error(), "explain supports text|json") {
					t.Fatalf("expected explain format restriction, got %v", err)
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

func TestApplyAutoInputMode(t *testing.T) {
	t.Run("non-git cwd with no mode errors", func(t *testing.T) {
		dir := t.TempDir()
		_, _, err := applyAutoInputMode(checkOptions{}, dir)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})

	t.Run("strict mode skips auto", func(t *testing.T) {
		dir := t.TempDir()
		got, auto, err := applyAutoInputMode(checkOptions{StrictInput: true}, dir)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if auto {
			t.Fatal("expected auto=false, got true")
		}
		if got.Base != "" || got.Head != "" {
			t.Fatalf("expected empty base/head, got %q/%q", got.Base, got.Head)
		}
	})

	t.Run("git cwd with no mode uses defaults", func(t *testing.T) {
		dir := t.TempDir()
		runGitCheck(t, dir, "init")

		got, auto, err := applyAutoInputMode(checkOptions{}, dir)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !auto {
			t.Fatal("expected auto=true, got false")
		}
		if got.Base != "origin/main" || got.Head != "HEAD" {
			t.Fatalf("unexpected refs %q/%q", got.Base, got.Head)
		}
	})

	t.Run("env overrides defaults", func(t *testing.T) {
		dir := t.TempDir()
		runGitCheck(t, dir, "init")
		t.Setenv("KYN_BASE_REF", "upstream/trunk")
		t.Setenv("KYN_HEAD_REF", "feature-branch")

		got, auto, err := applyAutoInputMode(checkOptions{}, dir)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !auto {
			t.Fatal("expected auto=true, got false")
		}
		if got.Base != "upstream/trunk" || got.Head != "feature-branch" {
			t.Fatalf("unexpected refs %q/%q", got.Base, got.Head)
		}
	})
}

func runGitCheck(t *testing.T, cwd string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", append([]string{"-C", cwd}, args...)...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v failed: %v\n%s", args, err, string(out))
	}
	return string(out)
}

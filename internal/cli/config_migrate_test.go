package cli

import (
	"os"
	"path/filepath"
	"testing"
)

func TestResolveOutputPath(t *testing.T) {
	cwd := t.TempDir()
	rel := resolveOutputPath(cwd, "out/kyn.v2.yaml")
	wantRel := filepath.Join(cwd, "out", "kyn.v2.yaml")
	if rel != wantRel {
		t.Fatalf("relative path = %q, want %q", rel, wantRel)
	}

	abs := filepath.Join(cwd, "abs.yaml")
	gotAbs := resolveOutputPath(cwd, abs)
	if gotAbs != abs {
		t.Fatalf("absolute path = %q, want %q", gotAbs, abs)
	}
}

func TestEnsureWritableTarget(t *testing.T) {
	dir := t.TempDir()
	missing := filepath.Join(dir, "missing.yaml")
	if err := ensureWritableTarget(missing, false); err != nil {
		t.Fatalf("unexpected error for missing target: %v", err)
	}

	existing := filepath.Join(dir, "existing.yaml")
	if err := os.WriteFile(existing, []byte("x"), 0o600); err != nil {
		t.Fatalf("write existing: %v", err)
	}
	if err := ensureWritableTarget(existing, false); err == nil {
		t.Fatal("expected error for existing target without force")
	}
	if err := ensureWritableTarget(existing, true); err != nil {
		t.Fatalf("unexpected error with force: %v", err)
	}
}

func TestPathForHint(t *testing.T) {
	cwd := t.TempDir()
	inside := filepath.Join(cwd, "nested", "kyn.v2.yaml")
	if got := pathForHint(cwd, inside); got != filepath.Join("nested", "kyn.v2.yaml") {
		t.Fatalf("inside hint = %q", got)
	}

	outside := filepath.Join(filepath.Dir(cwd), "outside.yaml")
	if got := pathForHint(cwd, outside); got != outside {
		t.Fatalf("outside hint = %q, want absolute path", got)
	}
}

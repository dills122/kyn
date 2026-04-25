package changes

import (
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

func TestCollectFromCSV(t *testing.T) {
	files, err := Collect(Input{
		FilesCSV: "b.ts,a.ts, ./a.ts ,a.ts",
	})
	if err != nil {
		t.Fatalf("Collect returned error: %v", err)
	}
	want := []string{"a.ts", "b.ts"}
	if !slices.Equal(files, want) {
		t.Fatalf("got %v, want %v", files, want)
	}
}

func TestCollectFromFile(t *testing.T) {
	dir := t.TempDir()
	listPath := filepath.Join(dir, "changed.txt")
	content := "c.ts\n./b.ts\na.ts\n\n"
	if err := os.WriteFile(listPath, []byte(content), 0o600); err != nil {
		t.Fatalf("write file list: %v", err)
	}

	files, err := Collect(Input{
		Cwd:       dir,
		FilesFrom: "changed.txt",
	})
	if err != nil {
		t.Fatalf("Collect returned error: %v", err)
	}
	want := []string{"a.ts", "b.ts", "c.ts"}
	if !slices.Equal(files, want) {
		t.Fatalf("got %v, want %v", files, want)
	}
}

func TestCollectFromGitDiffNameStatus(t *testing.T) {
	dir := t.TempDir()

	runGit(t, dir, "init")
	runGit(t, dir, "config", "user.email", "test@example.com")
	runGit(t, dir, "config", "user.name", "Test")

	write(t, filepath.Join(dir, "a.txt"), "old")
	write(t, filepath.Join(dir, "old.txt"), "old")
	write(t, filepath.Join(dir, "del.txt"), "old")
	runGit(t, dir, "add", ".")
	runGit(t, dir, "commit", "-m", "base")
	base := strings.TrimSpace(runGit(t, dir, "rev-parse", "HEAD"))

	write(t, filepath.Join(dir, "a.txt"), "new")
	write(t, filepath.Join(dir, "b.txt"), "new")
	runGit(t, dir, "mv", "old.txt", "new.txt")
	runGit(t, dir, "rm", "del.txt")
	runGit(t, dir, "add", ".")
	runGit(t, dir, "commit", "-m", "head")
	head := strings.TrimSpace(runGit(t, dir, "rev-parse", "HEAD"))

	files, err := Collect(Input{
		Cwd:  dir,
		Base: base,
		Head: head,
	})
	if err != nil {
		t.Fatalf("Collect returned error: %v", err)
	}

	// D is excluded for MVP; rename destination is included.
	want := []string{"a.txt", "b.txt", "new.txt"}
	if !slices.Equal(files, want) {
		t.Fatalf("got %v, want %v", files, want)
	}
}

func write(t *testing.T, path string, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

func runGit(t *testing.T, cwd string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", append([]string{"-C", cwd}, args...)...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v failed: %v\n%s", args, err, string(out))
	}
	return string(out)
}

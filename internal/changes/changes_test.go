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

func TestCollectDetailedStatusesFromCSV(t *testing.T) {
	result, err := CollectDetailed(Input{
		FilesCSV: "a.ts,b.ts",
	})
	if err != nil {
		t.Fatalf("CollectDetailed returned error: %v", err)
	}
	if result.StatusByFile["a.ts"] != StatusModified || result.StatusByFile["b.ts"] != StatusModified {
		t.Fatalf("expected modified statuses, got %+v", result.StatusByFile)
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

func TestCollectFromStdin(t *testing.T) {
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}
	t.Cleanup(func() {
		_ = r.Close()
	})

	prev := os.Stdin
	os.Stdin = r
	t.Cleanup(func() { os.Stdin = prev })

	_, _ = w.WriteString("z.ts\n./y.ts\nx.ts\n")
	_ = w.Close()

	files, err := Collect(Input{
		FilesFrom: "-",
	})
	if err != nil {
		t.Fatalf("Collect returned error: %v", err)
	}
	want := []string{"x.ts", "y.ts", "z.ts"}
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

func TestCollectDetailedFromGitDiffStatuses(t *testing.T) {
	dir := t.TempDir()

	runGit(t, dir, "init")
	runGit(t, dir, "config", "user.email", "test@example.com")
	runGit(t, dir, "config", "user.name", "Test")

	write(t, filepath.Join(dir, "a.txt"), "old")
	write(t, filepath.Join(dir, "old.txt"), "old")
	runGit(t, dir, "add", ".")
	runGit(t, dir, "commit", "-m", "base")
	base := strings.TrimSpace(runGit(t, dir, "rev-parse", "HEAD"))

	write(t, filepath.Join(dir, "a.txt"), "new")
	write(t, filepath.Join(dir, "b.txt"), "new")
	runGit(t, dir, "mv", "old.txt", "new.txt")
	runGit(t, dir, "add", ".")
	runGit(t, dir, "commit", "-m", "head")
	head := strings.TrimSpace(runGit(t, dir, "rev-parse", "HEAD"))

	result, err := CollectDetailed(Input{
		Cwd:  dir,
		Base: base,
		Head: head,
	})
	if err != nil {
		t.Fatalf("CollectDetailed returned error: %v", err)
	}
	if result.StatusByFile["a.txt"] != StatusModified {
		t.Fatalf("a.txt expected modified, got %s", result.StatusByFile["a.txt"])
	}
	if result.StatusByFile["b.txt"] != StatusAdded {
		t.Fatalf("b.txt expected added, got %s", result.StatusByFile["b.txt"])
	}
	if result.StatusByFile["new.txt"] != StatusRenamed {
		t.Fatalf("new.txt expected renamed, got %s", result.StatusByFile["new.txt"])
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
	prefix := []string{"-C", cwd, "-c", "commit.gpgsign=false"}
	cmd := exec.Command("git", append(prefix, args...)...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v failed: %v\n%s", args, err, string(out))
	}
	return string(out)
}

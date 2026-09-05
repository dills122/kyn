package e2e

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestInitGuidanceRunsInDeclaredShell(t *testing.T) {
	repo := filepath.Join(t.TempDir(), "repo $with % and ' quote")
	if err := os.MkdirAll(filepath.Join(repo, "src"), 0o755); err != nil {
		t.Fatalf("create repo: %v", err)
	}
	if err := os.WriteFile(filepath.Join(repo, "src", "button.component.ts"), []byte("export const button = true;\n"), 0o600); err != nil {
		t.Fatalf("write component: %v", err)
	}
	for _, args := range [][]string{
		{"init"},
		{"config", "user.email", "kyn@example.invalid"},
		{"config", "user.name", "Kyn Test"},
		{"add", "."},
		{"commit", "-m", "fixture"},
		{"update-ref", "refs/remotes/origin/main", "HEAD"},
	} {
		gitArgs := append([]string{"-c", "commit.gpgsign=false"}, args...)
		cmd := exec.Command("git", gitArgs...)
		cmd.Dir = repo
		if output, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, output)
		}
	}
	if err := os.WriteFile(filepath.Join(repo, "src", "button.component.ts"), []byte("export const button = false;\n"), 0o600); err != nil {
		t.Fatalf("update component: %v", err)
	}
	for _, args := range [][]string{{"add", "."}, {"commit", "-m", "change component"}} {
		gitArgs := append([]string{"-c", "commit.gpgsign=false"}, args...)
		cmd := exec.Command("git", gitArgs...)
		cmd.Dir = repo
		if output, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, output)
		}
	}

	configPath := "config files/kyn's $config%.yaml"
	stdout, stderr, exitCode := runKyn(t, t.TempDir(), []string{
		"init",
		"--cwd", repo,
		"--config", configPath,
	})
	if exitCode != 0 {
		t.Fatalf("init exit code = %d\nstdout:\n%s\nstderr:\n%s", exitCode, stdout, stderr)
	}

	wantShell := "POSIX shell"
	if runtime.GOOS == "windows" {
		wantShell = "PowerShell"
	}
	if !strings.Contains(stdout, "Next steps ("+wantShell+"):") {
		t.Fatalf("init output does not declare %s:\n%s", wantShell, stdout)
	}

	for _, label := range []string{"2. Preview: ", "3. Enforce: ", "4. Diagnose: "} {
		command := guidanceCommand(t, stdout, label)
		runGuidanceCommand(t, t.TempDir(), command)
	}
}

func guidanceCommand(t *testing.T, output, label string) string {
	t.Helper()
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, label) {
			return strings.TrimPrefix(line, label)
		}
	}
	t.Fatalf("guidance output does not contain %q:\n%s", label, output)
	return ""
}

func runGuidanceCommand(t *testing.T, dir, command string) {
	t.Helper()
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("powershell.exe", "-NoProfile", "-NonInteractive", "-Command", command)
	} else {
		cmd = exec.Command("/bin/sh", "-c", command)
	}
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "PATH="+filepath.Dir(kynBinary)+string(os.PathListSeparator)+os.Getenv("PATH"))
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("run emitted command %q: %v\n%s", command, err, output)
	}
}

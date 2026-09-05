package e2e

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestE2EPresetBootstrap exercises the exact commands documented in
// docs/presets.md: `kyn init --preset <preset>` followed by
// `kyn check --dry-run-resolve`. It catches malformed preset YAML
// (hardcoded as Go strings in internal/cli/init.go) that unit tests
// checking string content alone would not.
func TestE2EPresetBootstrap(t *testing.T) {
	presets := []string{"web-ui", "api", "proto", "iac"}

	for _, preset := range presets {
		preset := preset
		t.Run(preset, func(t *testing.T) {
			dir := t.TempDir()

			_, stderr, exitCode := runKyn(t, dir, []string{"init", "--preset", preset})
			if exitCode != 0 {
				t.Fatalf("kyn init --preset %s: exit=%d stderr=%s", preset, exitCode, stderr)
			}

			configPath := filepath.Join(dir, "kyn.config.yaml")
			if _, err := os.Stat(configPath); err != nil {
				t.Fatalf("kyn init --preset %s: config not written: %v", preset, err)
			}

			// No git repo here; pass an explicit (unrelated) changed file
			// so input-mode selection succeeds. The point of this
			// assertion is that the preset config loads and resolves
			// cleanly, matching docs/presets.md step 3
			// (`kyn check --dry-run-resolve`), not any particular match.
			stdout, stderr, exitCode := runKyn(t, dir, []string{
				"check", "-c", "kyn.config.yaml", "--dry-run-resolve", "--format", "json", "--files", "kyn.config.yaml",
			})
			if exitCode != 0 {
				t.Fatalf(
					"kyn check --dry-run-resolve on preset %s config: exit=%d\n--- stdout ---\n%s\n--- stderr ---\n%s",
					preset, exitCode, stdout, stderr,
				)
			}
		})
	}
}

func TestE2EEmptyInputUsesFailOnEmptyPolicy(t *testing.T) {
	dir := t.TempDir()
	_, stderr, exitCode := runKyn(t, dir, []string{"init", "--preset", "web-ui"})
	if exitCode != 0 {
		t.Fatalf("kyn init: exit=%d stderr=%s", exitCode, stderr)
	}

	stdout, stderr, exitCode := runKynStdin(t, dir, []string{
		"check", "--files-from", "-", "--format", "json",
	}, "")
	if exitCode != 0 {
		t.Fatalf("empty input: exit=%d, want 0\nstdout:\n%s\nstderr:\n%s", exitCode, stdout, stderr)
	}
	if !strings.Contains(stdout, `"ok": true`) {
		t.Fatalf("empty input report should pass:\n%s", stdout)
	}

	stdout, stderr, exitCode = runKynStdin(t, dir, []string{
		"check", "--files-from", "-", "--format", "json", "--fail-on-empty",
	}, "")
	if exitCode != 1 {
		t.Fatalf("empty input with --fail-on-empty: exit=%d, want 1\nstdout:\n%s\nstderr:\n%s", exitCode, stdout, stderr)
	}
	if !strings.Contains(stdout, `"ruleId": "fail-on-empty"`) {
		t.Fatalf("fail-on-empty report missing synthetic result:\n%s", stdout)
	}
}

func TestE2ERejectsChangedPathOutsideCWD(t *testing.T) {
	dir := t.TempDir()
	_, stderr, exitCode := runKyn(t, dir, []string{"init", "--preset", "web-ui"})
	if exitCode != 0 {
		t.Fatalf("kyn init: exit=%d stderr=%s", exitCode, stderr)
	}

	_, stderr, exitCode = runKyn(t, dir, []string{"check", "--files", "../outside.ts"})
	if exitCode != 2 {
		t.Fatalf("unsafe changed path: exit=%d, want 2\nstderr:\n%s", exitCode, stderr)
	}
	if !strings.Contains(stderr, "repository-relative") {
		t.Fatalf("unsafe changed path error is not actionable:\n%s", stderr)
	}
}

// TestE2EMigrateV1ToV2PreservesBehavior copies the angular-storybook v1
// fixture, records `kyn check` behavior, migrates the config to v2
// in place, and asserts the exact same check run produces the same exit
// code and rendered output. This is the real-world contract of `kyn
// config migrate`: the migrated config must evaluate identically.
func TestE2EMigrateV1ToV2PreservesBehavior(t *testing.T) {
	dir := t.TempDir()
	copyProjectFixture(t, "angular-storybook", dir)

	checkArgs := []string{
		"check", "-c", "kyn.config.yaml",
		"--files", "libs/ui/button/button.component.ts,libs/ui/button/button.component.html",
		"--format", "json",
	}

	preStdout, preStderr, preExit := runKyn(t, dir, checkArgs)
	if preExit != 1 {
		t.Fatalf("pre-migration check: exit=%d, want 1\nstdout:\n%s\nstderr:\n%s", preExit, preStdout, preStderr)
	}

	_, migStderr, migExit := runKyn(t, dir, []string{
		"config", "migrate", "-c", "kyn.config.yaml", "--from", "v1", "--to", "v2", "--in-place", "--force",
	})
	if migExit != 0 {
		t.Fatalf("config migrate: exit=%d stderr=%s", migExit, migStderr)
	}

	postStdout, postStderr, postExit := runKyn(t, dir, checkArgs)
	if postExit != preExit {
		t.Fatalf("post-migration exit=%d, want %d (unchanged from pre-migration)\nstderr:\n%s", postExit, preExit, postStderr)
	}
	if postStdout != preStdout {
		t.Fatalf(
			"migrated config changed check behavior\n--- pre-migration ---\n%s\n--- post-migration ---\n%s",
			preStdout, postStdout,
		)
	}
}

// copyProjectFixture copies an e2e/projects/<name> fixture tree into dst,
// skipping the e2e-only scenarios.json and testdata/ golden directory so
// dst looks like a plain checkout of the example project.
func copyProjectFixture(t *testing.T, name string, dst string) {
	t.Helper()

	src := filepath.Join("projects", name)
	err := filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		if rel == "." {
			return nil
		}
		if rel == "scenarios.json" || rel == "testdata" {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		target := filepath.Join(dst, rel)
		if info.IsDir() {
			return os.MkdirAll(target, 0o755)
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		return os.WriteFile(target, data, 0o600)
	})
	if err != nil {
		t.Fatalf("copy fixture %s: %v", name, err)
	}
}

package changes

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"
)

func TestRunBoundedCommandLimitsOutput(t *testing.T) {
	result, err := runBoundedCommand(
		context.Background(),
		8,
		os.Args[0],
		"-test.run=^TestGitCommandHelperProcess$",
		"--",
		"output",
	)
	if err != nil {
		t.Fatalf("runBoundedCommand returned error: %v", err)
	}
	if !result.stdout.truncated {
		t.Fatalf("expected stdout to be marked truncated; captured %q", result.stdout.String())
	}
	if got := result.stdout.String(); got != "xxxxxxxx" {
		t.Fatalf("captured stdout = %q, want eight bytes", got)
	}
}

func TestRunBoundedCommandHonorsContext(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	_, err := runBoundedCommand(
		ctx,
		1024,
		os.Args[0],
		"-test.run=^TestGitCommandHelperProcess$",
		"--",
		"sleep",
	)
	if err == nil {
		t.Fatal("expected command cancellation error")
	}
	if ctx.Err() != context.DeadlineExceeded {
		t.Fatalf("context error = %v, want deadline exceeded", ctx.Err())
	}
}

func TestGitCommandHelperProcess(t *testing.T) {
	separator := -1
	for i, arg := range os.Args {
		if arg == "--" {
			separator = i
			break
		}
	}
	if separator == -1 || separator+1 >= len(os.Args) {
		return
	}

	switch os.Args[separator+1] {
	case "output":
		_, _ = fmt.Fprint(os.Stdout, strings.Repeat("x", 64))
	case "sleep":
		time.Sleep(time.Second)
	default:
		os.Exit(2)
	}
}

package main

import "testing"

func TestMainUsesExecuteAndExit(t *testing.T) {
	prevExecute := execute
	prevExit := exit
	t.Cleanup(func() {
		execute = prevExecute
		exit = prevExit
	})

	executed := false
	gotCode := -1
	execute = func() int {
		executed = true
		return 7
	}
	exit = func(code int) {
		gotCode = code
	}

	main()

	if !executed {
		t.Fatal("expected execute to be called")
	}
	if gotCode != 7 {
		t.Fatalf("exit code = %d, want 7", gotCode)
	}
}

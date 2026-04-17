package main

import (
	"bytes"
	"os"
	"os/exec"
)

// runGit runs `git <args...>` inheriting stdio. Returns the exit code.
func runGit(args ...string) int {
	cmd := exec.Command("git", args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return waitCode(cmd.Run())
}

// runGitWithStdin runs `git <args...>` feeding stdin from a byte slice.
// stdout and stderr are inherited. Returns the exit code.
func runGitWithStdin(stdin []byte, args ...string) int {
	cmd := exec.Command("git", args...)
	cmd.Stdin = bytes.NewReader(stdin)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return waitCode(cmd.Run())
}

func waitCode(err error) int {
	if err == nil {
		return 0
	}
	if ee, ok := err.(*exec.ExitError); ok {
		return ee.ExitCode()
	}
	// git not found or unable to spawn.
	return 127
}

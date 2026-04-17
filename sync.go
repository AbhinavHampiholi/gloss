package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// cmdFetch runs `git fetch <remote> +refs/notes/gloss:refs/notes/gloss`.
// If the remote doesn't have the notes ref yet, report that and exit 0
// rather than leaking git's scary "couldn't find remote ref" error.
func cmdFetch(args []string) int {
	remote, code := singleRemoteArg("fetch", args)
	if code != 0 {
		return code
	}

	refspec := fmt.Sprintf("+%s:%s", notesRef, notesRef)

	cmd := exec.Command("git", "fetch", remote, refspec)
	cmd.Stdout = os.Stdout
	var errBuf bytes.Buffer
	cmd.Stderr = &errBuf
	err := cmd.Run()
	stderr := errBuf.String()

	if err != nil {
		if strings.Contains(stderr, "couldn't find remote ref") {
			fmt.Printf("git-gloss: %s has no gloss notes yet (nothing to fetch)\n", remote)
			return 0
		}
		// Pass git's error through; preserve its exit code.
		fmt.Fprint(os.Stderr, stderr)
		return waitCode(err)
	}

	// Progress/status from git fetch goes to stderr even on success.
	fmt.Fprint(os.Stderr, stderr)
	return 0
}

// cmdPush pushes refs/notes/gloss to the given remote (default origin).
// If there are no local notes yet, report that and exit 0 instead of
// letting git complain about an unknown src refspec.
func cmdPush(args []string) int {
	remote, code := singleRemoteArg("push", args)
	if code != 0 {
		return code
	}

	if err := exec.Command("git", "rev-parse", "--verify", "--quiet", notesRef).Run(); err != nil {
		fmt.Println("git-gloss: no local gloss notes yet (nothing to push)")
		return 0
	}

	return runGit("push", remote, fmt.Sprintf("%s:%s", notesRef, notesRef))
}

// singleRemoteArg parses an optional single remote argument, defaulting
// to "origin". Returns (remote, exitCode). If exitCode != 0, an error
// was already printed and the caller should exit with that code.
func singleRemoteArg(sub string, args []string) (string, int) {
	switch len(args) {
	case 0:
		return "origin", 0
	case 1:
		return args[0], 0
	default:
		fmt.Fprintf(os.Stderr, "git-gloss %s: takes at most one argument (remote name)\n", sub)
		return "", 2
	}
}

package main

import (
	"fmt"
	"os"
)

// cmdCopy copies a note from one commit to another. Thin wrapper around
// `git notes copy`, constrained to the gloss namespace.
//
// Primary use case: manually attaching a note onto a squash-merged
// commit. In normal workflow this is unnecessary — `note` and `show`
// use the PR-number fallback automatically. Keep this as an escape
// hatch for situations where automatic resolution isn't possible
// (e.g. no gh auth, private PR, or the resolved SHA is different
// from what you want).
func cmdCopy(args []string) int {
	force := false
	positional := make([]string, 0, 2)
	for _, a := range args {
		switch a {
		case "-f", "--force":
			force = true
		case "-h", "--help":
			fmt.Println("Usage: git gloss copy [-f] <FROM> <TO>")
			return 0
		default:
			positional = append(positional, a)
		}
	}
	if len(positional) != 2 {
		fmt.Fprintln(os.Stderr, "Usage: git gloss copy [-f] <FROM> <TO>")
		return 2
	}
	from, to := positional[0], positional[1]

	cmdArgs := []string{"notes", "--ref=" + notesRef, "copy"}
	if force {
		cmdArgs = append(cmdArgs, "-f")
	}
	cmdArgs = append(cmdArgs, from, to)
	return runGit(cmdArgs...)
}

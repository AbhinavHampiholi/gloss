package main

import (
	"fmt"
	"os"
)

// cmdShow is `git show --show-notes=gloss`. All extra args pass through,
// so `git gloss show HEAD~3 --stat` works as expected.
func cmdShow(args []string) int {
	return runGit(append([]string{"show", "--show-notes=" + notesRef}, args...)...)
}

// cmdNote prints the raw gloss note for a commit to stdout, with no
// decoration. Exits 0 on success, 1 silently if no note exists, 2 on
// bad usage. Designed for scripts and agents:
//
//	gloss note HEAD          # current commit's context
//	gloss note abc123 | jq . # parse it if you stored JSON
func cmdNote(args []string) int {
	rev := "HEAD"
	switch len(args) {
	case 0:
	case 1:
		rev = args[0]
	default:
		fmt.Fprintln(os.Stderr, "git-gloss note: takes at most one argument (a commit-ish)")
		return 2
	}

	code := runGit("notes", "--ref="+notesRef, "show", rev)
	// git notes show exits 1 with "error: no note found for object ..."
	// on stderr. That's already informative; just propagate the code.
	return code
}

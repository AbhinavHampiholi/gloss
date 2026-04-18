package main

import (
	"fmt"
	"os"
	"strings"
)

// cmdShow is `git show` augmented with the gloss note for the given
// commit, if one exists.
//
// If the commit has a direct note, git's native `--show-notes` renders
// it between the message and the diff (standard layout).
//
// If there's no direct note but the commit subject contains a PR number
// ("(#N)") and gh resolves PR #N's head commit to a commit that does
// have a note, we render the commit without native notes and append the
// note at the end, annotated with its source.
func cmdShow(args []string) int {
	// Find the rev argument (last non-flag token). Defaults to HEAD.
	rev := "HEAD"
	for _, a := range args {
		if !strings.HasPrefix(a, "-") {
			rev = a
		}
	}

	// Direct-note path: let git render everything natively.
	if _, err := noteBody(rev); err == nil {
		return runGit(append([]string{"show", "--show-notes=" + notesRef}, args...)...)
	}

	// No direct note. Show the commit as-is, then try the PR fallback.
	if code := runGit(append([]string{"show"}, args...)...); code != 0 {
		return code
	}
	r, err := resolveNote(rev, true)
	if err != nil {
		return 0 // nothing to append; normal "no note" outcome
	}

	// Print a gloss notes section after the diff, with source attribution.
	fmt.Printf("\nNotes (gloss, via PR #%d @ %s):\n", r.viaPR, shortSHA(r.sourceSHA))
	for _, line := range strings.Split(strings.TrimRight(string(r.body), "\n"), "\n") {
		fmt.Println("    " + line)
	}
	return 0
}

// cmdNote prints the raw gloss note for rev to stdout.
//
// By default, if no note is attached directly, falls back to PR-number
// lookup via gh (see resolveNote). Pass --no-fallback to disable.
//
// When a fallback resolves, we emit a single-line notice on stderr so
// the user knows where the note came from. Stdout stays pristine so
// `git gloss note SHA | claude` pipes cleanly either way.
func cmdNote(args []string) int {
	allowFallback := true
	rev := "HEAD"
	extra := 0
	for _, a := range args {
		switch {
		case a == "--no-fallback":
			allowFallback = false
		case strings.HasPrefix(a, "-"):
			fmt.Fprintf(os.Stderr, "git-gloss note: unknown flag %q\n", a)
			return 2
		default:
			rev = a
			extra++
		}
	}
	if extra > 1 {
		fmt.Fprintln(os.Stderr, "git-gloss note: takes at most one commit-ish")
		return 2
	}

	r, err := resolveNote(rev, allowFallback)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: no note found for object %s.\n", rev)
		return 1
	}

	os.Stdout.Write(r.body)
	if !strings.HasSuffix(string(r.body), "\n") {
		fmt.Println()
	}
	if r.viaPR > 0 {
		fmt.Fprintf(os.Stderr, "gloss: resolved via PR #%d @ %s\n",
			r.viaPR, shortSHA(r.sourceSHA))
	}
	return 0
}

func shortSHA(sha string) string {
	if len(sha) > 12 {
		return sha[:12]
	}
	return sha
}

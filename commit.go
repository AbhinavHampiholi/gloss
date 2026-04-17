package main

import (
	"fmt"
	"io"
	"os"
	"strings"
)

// cmdCommit is `git commit` plus an optional -C / --context <file|->
// which, on successful commit, is attached to HEAD as a note in
// refs/notes/gloss.
//
// The -C flag is stripped before dispatching; every other flag is
// passed through unchanged, so `git commit`'s full surface area
// (-a, --amend, --author, -F, -S, ...) works as normal.
//
// If the context source is empty (empty file, empty stdin), no note
// is attached and the commit proceeds as a plain `git commit`.
func cmdCommit(args []string) int {
	contextSrc, rest, err := extractContextFlag(args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "git-gloss commit: %v\n", err)
		return 2
	}

	// Read the context *before* committing so we can surface file I/O
	// errors without leaving a commit without its intended note.
	var note []byte
	if contextSrc != "" {
		note, err = readContext(contextSrc)
		if err != nil {
			fmt.Fprintf(os.Stderr, "git-gloss commit: %v\n", err)
			return 1
		}
	}

	if code := runGit(append([]string{"commit"}, rest...)...); code != 0 {
		return code
	}

	if len(strings.TrimSpace(string(note))) == 0 {
		return 0
	}

	// Attach the note to HEAD. We use -F - so the note can contain
	// anything (including leading dashes) without shell-quoting gymnastics.
	return runGitWithStdin(note,
		"notes", "--ref="+notesRef, "add", "-F", "-", "HEAD")
}

// extractContextFlag pulls -C/--context (and its value) out of args,
// returning the value and the remaining args in original order.
//
// Supports:  -C value   --context value   -C=value   --context=value
func extractContextFlag(args []string) (value string, rest []string, err error) {
	rest = make([]string, 0, len(args))
	found := false

	for i := 0; i < len(args); i++ {
		a := args[i]

		switch {
		case a == "-C" || a == "--context":
			if found {
				return "", nil, fmt.Errorf("-C/--context specified more than once")
			}
			if i+1 >= len(args) {
				return "", nil, fmt.Errorf("-C/--context requires a value (file path or '-' for stdin)")
			}
			value = args[i+1]
			i++
			found = true
		case strings.HasPrefix(a, "-C="):
			if found {
				return "", nil, fmt.Errorf("-C/--context specified more than once")
			}
			value = strings.TrimPrefix(a, "-C=")
			found = true
		case strings.HasPrefix(a, "--context="):
			if found {
				return "", nil, fmt.Errorf("-C/--context specified more than once")
			}
			value = strings.TrimPrefix(a, "--context=")
			found = true
		default:
			rest = append(rest, a)
		}
	}
	return value, rest, nil
}

func readContext(src string) ([]byte, error) {
	if src == "-" {
		b, err := io.ReadAll(os.Stdin)
		if err != nil {
			return nil, fmt.Errorf("reading context from stdin: %w", err)
		}
		return b, nil
	}
	b, err := os.ReadFile(src)
	if err != nil {
		return nil, fmt.Errorf("reading context from %s: %w", src, err)
	}
	return b, nil
}

// git-gloss: attach verbose, commit-pinned context to git commits
// via a dedicated notes namespace (refs/notes/gloss).
//
// Invoked as either `git-gloss <cmd>` or `git gloss <cmd>` (git
// auto-discovers executables named git-* on $PATH).
package main

import (
	"fmt"
	"os"
)

const notesRef = "refs/notes/gloss"

const usage = `git-gloss: attach verbose context to git commits

Usage:
  git gloss init                             Configure this repo for gloss
  git gloss commit -m <msg> [-C <file|->]    Commit, optionally with context
  git gloss show <commit>                    Show a commit with its context
  git gloss note <commit>                    Print a commit's raw context

Commit flags:
  -C, --context <file>   Read context from file. Use "-" for stdin.
  All other flags are passed through to 'git commit'.
`

func main() {
	args := os.Args[1:]
	if len(args) == 0 {
		fmt.Fprint(os.Stderr, usage)
		os.Exit(2)
	}

	cmd, rest := args[0], args[1:]

	switch cmd {
	case "init":
		os.Exit(cmdInit(rest))
	case "commit":
		os.Exit(cmdCommit(rest))
	case "show":
		os.Exit(cmdShow(rest))
	case "note":
		os.Exit(cmdNote(rest))
	case "-h", "--help", "help":
		fmt.Print(usage)
		os.Exit(0)
	default:
		fmt.Fprintf(os.Stderr, "git-gloss: unknown command %q\n\n%s", cmd, usage)
		os.Exit(2)
	}
}

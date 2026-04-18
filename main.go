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
  git gloss init                               Configure this repo for gloss
  git gloss commit -m <msg> [-C <file|->]      Commit, optionally with context
  git gloss list  [-n N] [--glossed] [--all]   Recent commits, ● if glossed
  git gloss show  <commit>                     Show a commit with its context
  git gloss note  <commit> [--no-fallback]     Print a commit's raw context
  git gloss copy  [-f] <from> <to>             Copy a note from one commit to another
  git gloss fetch [remote]                     Fetch gloss notes from remote
  git gloss push  [remote]                     Push local gloss notes to remote

Commit flags:
  -C, --context <file>   Read context from file. Use "-" for stdin.
  All other flags are passed through to 'git commit'.

note/show fallback:
  If a commit has no note directly, note/show try to resolve via "(#N)"
  in the commit subject: 'gh pr view N --json headRefOid' → lookup on
  that SHA. Disable with --no-fallback (note only). Requires 'gh' on
  PATH and authenticated; silent no-op otherwise.
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
	case "list", "ls":
		os.Exit(cmdList(rest))
	case "copy", "cp":
		os.Exit(cmdCopy(rest))
	case "fetch":
		os.Exit(cmdFetch(rest))
	case "push":
		os.Exit(cmdPush(rest))
	case "-h", "--help", "help":
		fmt.Print(usage)
		os.Exit(0)
	default:
		fmt.Fprintf(os.Stderr, "git-gloss: unknown command %q\n\n%s", cmd, usage)
		os.Exit(2)
	}
}

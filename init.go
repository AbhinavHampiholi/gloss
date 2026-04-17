package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// cmdInit configures the current repo to make gloss notes work well:
//   - notes.rewriteRef so notes follow commits through rebase/amend
//   - fetch/push refspecs on every remote so notes travel over the wire
//
// Idempotent: re-running is safe. Existing matching refspecs are not
// duplicated; new remotes added later can be picked up by re-running.
func cmdInit(args []string) int {
	if len(args) > 0 {
		fmt.Fprintf(os.Stderr, "git-gloss init: unexpected argument(s): %v\n", args)
		return 2
	}

	if !insideGitRepo() {
		fmt.Fprintln(os.Stderr, "git-gloss init: not inside a git repository")
		return 128
	}

	// 1. Make notes survive rebase/amend.
	if code := runGit("config", "notes.rewriteRef", notesRef); code != 0 {
		return code
	}

	// 2. Add fetch/push refspecs to every configured remote.
	remotes, err := listRemotes()
	if err != nil {
		fmt.Fprintf(os.Stderr, "git-gloss init: %v\n", err)
		return 1
	}
	if len(remotes) == 0 {
		fmt.Println("git-gloss: configured notes.rewriteRef")
		fmt.Println("git-gloss: no remotes found; run 'git gloss init' again after adding one")
		return 0
	}

	refspec := fmt.Sprintf("+%s:%s", notesRef, notesRef)
	for _, r := range remotes {
		if err := ensureRefspec(r, "fetch", refspec); err != nil {
			fmt.Fprintf(os.Stderr, "git-gloss init: %v\n", err)
			return 1
		}
		if err := ensureRefspec(r, "push", refspec); err != nil {
			fmt.Fprintf(os.Stderr, "git-gloss init: %v\n", err)
			return 1
		}
	}

	fmt.Printf("git-gloss: configured %d remote(s): %s\n",
		len(remotes), strings.Join(remotes, ", "))
	return 0
}

func insideGitRepo() bool {
	cmd := exec.Command("git", "rev-parse", "--is-inside-work-tree")
	cmd.Stderr = nil
	return cmd.Run() == nil
}

func listRemotes() ([]string, error) {
	out, err := exec.Command("git", "remote").Output()
	if err != nil {
		return nil, fmt.Errorf("git remote: %w", err)
	}
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	var remotes []string
	for _, l := range lines {
		if l = strings.TrimSpace(l); l != "" {
			remotes = append(remotes, l)
		}
	}
	return remotes, nil
}

// ensureRefspec adds refspec to remote.<remote>.<key> only if not already present.
func ensureRefspec(remote, key, refspec string) error {
	cfgKey := fmt.Sprintf("remote.%s.%s", remote, key)

	out, err := exec.Command("git", "config", "--get-all", cfgKey).Output()
	// Exit 1 just means "key not set"; that's fine.
	if err != nil {
		if ee, ok := err.(*exec.ExitError); ok && ee.ExitCode() == 1 {
			// not set yet
		} else {
			return fmt.Errorf("git config --get-all %s: %w", cfgKey, err)
		}
	}
	for _, existing := range strings.Split(string(out), "\n") {
		if strings.TrimSpace(existing) == refspec {
			return nil // already there
		}
	}

	cmd := exec.Command("git", "config", "--add", cfgKey, refspec)
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

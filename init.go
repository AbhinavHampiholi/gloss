package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// cmdInit configures the current repo for gloss.
//
// Sets:
//   - notes.rewriteRef so notes follow commits through rebase/amend.
//
// Does NOT touch remote fetch/push refspecs. A concrete refspec like
// +refs/notes/gloss:refs/notes/gloss causes plain `git fetch`/`git pull`
// to fail until the remote actually has that ref, which it never does
// on a fresh repo. Use `git gloss fetch` / `git gloss push` instead,
// which transport notes explicitly and handle the empty case.
//
// As a courtesy, if this repo has legacy gloss refspecs from an earlier
// version of gloss init, we remove them so `git pull` stops erroring.
//
// Idempotent: safe to re-run.
func cmdInit(args []string) int {
	if len(args) > 0 {
		fmt.Fprintf(os.Stderr, "git-gloss init: unexpected argument(s): %v\n", args)
		return 2
	}
	if !insideGitRepo() {
		fmt.Fprintln(os.Stderr, "git-gloss init: not inside a git repository")
		return 128
	}

	if code := runGit("config", "notes.rewriteRef", notesRef); code != 0 {
		return code
	}

	removed, err := stripLegacyRefspecs()
	if err != nil {
		fmt.Fprintf(os.Stderr, "git-gloss init: %v\n", err)
		return 1
	}

	fmt.Println("git-gloss: configured notes.rewriteRef=" + notesRef)
	if removed > 0 {
		fmt.Printf("git-gloss: removed %d legacy gloss refspec(s) from your remotes\n", removed)
	}
	fmt.Println("git-gloss: use 'git gloss push' / 'git gloss fetch' to move notes between repos")
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
	var remotes []string
	for _, l := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		if l = strings.TrimSpace(l); l != "" {
			remotes = append(remotes, l)
		}
	}
	return remotes, nil
}

// stripLegacyRefspecs removes +refs/notes/gloss:refs/notes/gloss entries
// that older versions of `gloss init` added to remote fetch/push config.
// Returns how many entries were removed across all remotes.
func stripLegacyRefspecs() (int, error) {
	remotes, err := listRemotes()
	if err != nil {
		return 0, err
	}
	legacy := fmt.Sprintf("+%s:%s", notesRef, notesRef)
	removed := 0
	for _, r := range remotes {
		for _, key := range []string{"fetch", "push"} {
			cfgKey := fmt.Sprintf("remote.%s.%s", r, key)
			// --unset-all with an exact-match value pattern. Silently exits 5
			// ("no such section" / "no matching values"), which is fine.
			cmd := exec.Command("git", "config", "--unset-all", cfgKey, regexEscape(legacy))
			if err := cmd.Run(); err == nil {
				removed++
			}
		}
	}
	return removed, nil
}

// regexEscape escapes the characters git config treats as regex metacharacters
// in its value-pattern matching, so we match the literal refspec exactly.
func regexEscape(s string) string {
	const meta = `\.+*?()[]{}|^$`
	var b strings.Builder
	for _, r := range s {
		if strings.ContainsRune(meta, r) {
			b.WriteByte('\\')
		}
		b.WriteRune(r)
	}
	return b.String()
}

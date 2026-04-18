package main

import (
	"errors"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

// errNoNote is returned by resolveNote when no note was found by any path.
var errNoNote = errors.New("no note found")

// resolved describes a note lookup result.
type resolved struct {
	body      []byte // raw note bytes
	sourceSHA string // commit SHA the note is actually attached to
	viaPR     int    // PR number used to resolve, or 0 if direct lookup
}

// resolveNote returns the gloss note for rev.
//
// Lookup order:
//  1. Direct: a note on rev itself in refs/notes/gloss.
//  2. Fallback (if allowFallback): extract "(#N)" from rev's commit subject,
//     use `gh pr view N --json headRefOid` to find the original PR head SHA,
//     look up a note there.
//
// The fallback fails silently (returns errNoNote) if gh is not on PATH,
// not authenticated, the PR isn't found, or the head SHA has no note.
// Callers that want strictly-direct behavior pass allowFallback=false.
func resolveNote(rev string, allowFallback bool) (*resolved, error) {
	// 1. Direct lookup.
	if body, err := noteBody(rev); err == nil {
		sha := fullSHA(rev)
		return &resolved{body: body, sourceSHA: sha}, nil
	}
	if !allowFallback {
		return nil, errNoNote
	}

	// 2. Try PR-number fallback.
	pr, ok := prNumberFromSubject(rev)
	if !ok {
		return nil, errNoNote
	}
	if _, err := exec.LookPath("gh"); err != nil {
		return nil, errNoNote
	}
	headSHA, err := ghHeadSHA(pr)
	if err != nil || headSHA == "" {
		return nil, errNoNote
	}
	body, err := noteBody(headSHA)
	if err != nil {
		return nil, errNoNote
	}
	return &resolved{body: body, sourceSHA: headSHA, viaPR: pr}, nil
}

// noteBody returns the raw note body for rev, or an error if none exists.
func noteBody(rev string) ([]byte, error) {
	out, err := exec.Command("git", "notes", "--ref="+notesRef, "show", rev).Output()
	if err != nil {
		return nil, err
	}
	return out, nil
}

// noteExists reports whether a note is retrievable for rev.
//
// If allowFallback is true, follows the same PR-number resolution as
// resolveNote, but without fetching the note body — cheaper for callers
// that only need to decide a marker (e.g. `list`). Never errors; just
// returns a boolean. gh lookup failures of any kind count as "no note".
func noteExists(rev string, allowFallback bool) bool {
	if err := exec.Command("git", "notes", "--ref="+notesRef, "show", rev).Run(); err == nil {
		return true
	}
	if !allowFallback {
		return false
	}
	pr, ok := prNumberFromSubject(rev)
	if !ok {
		return false
	}
	if _, err := exec.LookPath("gh"); err != nil {
		return false
	}
	headSHA, err := ghHeadSHA(pr)
	if err != nil || headSHA == "" {
		return false
	}
	return exec.Command("git", "notes", "--ref="+notesRef, "show", headSHA).Run() == nil
}

// fullSHA resolves rev to a full 40-char SHA, or returns rev unchanged on failure.
func fullSHA(rev string) string {
	out, err := exec.Command("git", "rev-parse", rev).Output()
	if err != nil {
		return rev
	}
	return strings.TrimSpace(string(out))
}

// prNumberFromSubject pulls a trailing "(#N)" or "#N" out of rev's commit
// subject. Handles the formats GitHub uses for squash and merge commits.
// Returns (0, false) if no PR number is present.
func prNumberFromSubject(rev string) (int, bool) {
	out, err := exec.Command("git", "log", "-1", "--format=%s", rev).Output()
	if err != nil {
		return 0, false
	}
	subject := strings.TrimSpace(string(out))
	// Match "(#1234)" or "#1234" near the end of the subject line.
	re := regexp.MustCompile(`#(\d+)\)?\s*$`)
	m := re.FindStringSubmatch(subject)
	if m == nil {
		return 0, false
	}
	n, err := strconv.Atoi(m[1])
	if err != nil {
		return 0, false
	}
	return n, true
}

// ghHeadSHA asks gh for the head commit SHA of PR n.
func ghHeadSHA(n int) (string, error) {
	cmd := exec.Command("gh", "pr", "view", strconv.Itoa(n),
		"--json", "headRefOid", "-q", ".headRefOid")
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

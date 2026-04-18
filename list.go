package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

// cmdList prints recent commits with a marker indicating whether each
// has a gloss note attached.
//
//	● <short-sha> <subject>    has a gloss note
//	· <short-sha> <subject>    no note
//
// Flags:
//
//	-n N          limit to the N most recent commits (default 20)
//	--glossed     only show commits that have a note
//	<revrange>    any argument that doesn't start with '-' is passed to
//	              git log verbatim (e.g. 'main..HEAD', 'HEAD~50..')
func cmdList(args []string) int {
	n := 20
	glossedOnly := false
	var revrange []string

	for i := 0; i < len(args); i++ {
		a := args[i]
		switch {
		case a == "-n" || a == "--max-count":
			if i+1 >= len(args) {
				fmt.Fprintf(os.Stderr, "git-gloss list: %s requires a value\n", a)
				return 2
			}
			v, err := strconv.Atoi(args[i+1])
			if err != nil || v < 1 {
				fmt.Fprintf(os.Stderr, "git-gloss list: %s: invalid count %q\n", a, args[i+1])
				return 2
			}
			n = v
			i++
		case strings.HasPrefix(a, "-n"):
			v, err := strconv.Atoi(strings.TrimPrefix(a, "-n"))
			if err != nil || v < 1 {
				fmt.Fprintf(os.Stderr, "git-gloss list: invalid count %q\n", a)
				return 2
			}
			n = v
		case a == "--glossed":
			glossedOnly = true
		case a == "-h" || a == "--help":
			fmt.Println("Usage: git gloss list [-n N] [--glossed] [<revrange>]")
			return 0
		case strings.HasPrefix(a, "-"):
			fmt.Fprintf(os.Stderr, "git-gloss list: unknown flag %q\n", a)
			return 2
		default:
			revrange = append(revrange, a)
		}
	}

	// Build the git log invocation. Use tab as a safe separator between
	// the short sha and the subject since subjects can contain anything
	// printable except newlines.
	logArgs := []string{"log", "--format=%h\t%s"}
	if len(revrange) == 0 {
		logArgs = append(logArgs, "-n", strconv.Itoa(n))
	} else {
		// When a revrange is supplied, we still honor -n as an upper bound.
		logArgs = append(logArgs, "-n", strconv.Itoa(n))
		logArgs = append(logArgs, revrange...)
	}

	cmd := exec.Command("git", logArgs...)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		fmt.Fprintf(os.Stderr, "git-gloss list: %v\n", err)
		return 1
	}
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "git-gloss list: %v\n", err)
		return 1
	}

	w := bufio.NewWriter(os.Stdout)
	defer w.Flush()

	scanner := bufio.NewScanner(stdout)
	// Commit subjects can be long. Bump the buffer ceiling.
	scanner.Buffer(make([]byte, 64*1024), 1024*1024)

	for scanner.Scan() {
		line := scanner.Text()
		tab := strings.IndexByte(line, '\t')
		if tab < 0 {
			continue
		}
		sha := line[:tab]
		subj := line[tab+1:]

		has := hasGlossNote(sha)
		if glossedOnly && !has {
			continue
		}
		marker := "·"
		if has {
			marker = "●"
		}
		fmt.Fprintf(w, "%s %s %s\n", marker, sha, subj)
	}
	if err := scanner.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "git-gloss list: %v\n", err)
		return 1
	}
	if err := cmd.Wait(); err != nil {
		return waitCode(err)
	}
	return 0
}

// hasGlossNote reports whether the given commit has a note in refs/notes/gloss.
// Uses `git notes show` and discards output; exit 0 = present, nonzero = absent.
func hasGlossNote(rev string) bool {
	cmd := exec.Command("git", "notes", "--ref="+notesRef, "show", rev)
	cmd.Stdout = nil
	cmd.Stderr = nil
	return cmd.Run() == nil
}

package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"sort"
	"strconv"
	"strings"
	"sync"
)

// cmdList prints recent commits with a marker indicating whether each
// has a gloss note attached.
//
//	● <short-sha> <subject>    has a gloss note
//	· <short-sha> <subject>    no note
//
// Flags:
//
//	-n N            limit to the N most recent commits (default 20)
//	--glossed       only show commits that have a note
//	--all           list every commit in the repo with a note, regardless
//	                of reachability from HEAD. Sorted by commit time desc.
//	--no-fallback   skip the PR-number fallback when deciding markers;
//	                faster, but squash-merged commits won't be marked ●.
//	<revrange>      any argument that doesn't start with '-' is passed to
//	                git log verbatim (e.g. 'main..HEAD', 'HEAD~50..').
//
// By default, marker resolution mirrors `note` / `show`: commits without
// a direct note are checked via the PR-number fallback (gh). Checks run
// in parallel across commits so walltime stays bounded.
func cmdList(args []string) int {
	n := 20
	glossedOnly := false
	listAll := false
	noFallback := false
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
		case a == "--all":
			listAll = true
		case a == "--no-fallback":
			noFallback = true
		case a == "-h" || a == "--help":
			fmt.Println("Usage: git gloss list [-n N] [--glossed] [--all] [--no-fallback] [<revrange>]")
			return 0
		case strings.HasPrefix(a, "-"):
			fmt.Fprintf(os.Stderr, "git-gloss list: unknown flag %q\n", a)
			return 2
		default:
			revrange = append(revrange, a)
		}
	}

	if listAll {
		if len(revrange) > 0 {
			fmt.Fprintln(os.Stderr, "git-gloss list: --all cannot be combined with a revrange")
			return 2
		}
		return listAllGlossed()
	}

	return listReachable(n, glossedOnly, !noFallback, revrange)
}

// listReachable walks commits reachable from HEAD (or the given revrange)
// and marks each with ● / · depending on note presence (via direct lookup
// and — if allowFallback — the PR-number fallback). Resolution runs in
// parallel across commits, capped at maxListConcurrency.
const maxListConcurrency = 10

func listReachable(n int, glossedOnly, allowFallback bool, revrange []string) int {
	logArgs := []string{"log", "--format=%h\t%s", "-n", strconv.Itoa(n)}
	logArgs = append(logArgs, revrange...)

	out, err := exec.Command("git", logArgs...).Output()
	if err != nil {
		fmt.Fprintf(os.Stderr, "git-gloss list: %v\n", err)
		return waitCode(err)
	}

	type row struct {
		sha    string
		subj   string
		marker string
	}

	var rows []row
	scanner := bufio.NewScanner(strings.NewReader(string(out)))
	scanner.Buffer(make([]byte, 64*1024), 1024*1024)
	for scanner.Scan() {
		line := scanner.Text()
		tab := strings.IndexByte(line, '\t')
		if tab < 0 {
			continue
		}
		rows = append(rows, row{sha: line[:tab], subj: line[tab+1:]})
	}

	// Resolve markers in parallel. Each worker decides ● vs · on its row.
	sem := make(chan struct{}, maxListConcurrency)
	var wg sync.WaitGroup
	for i := range rows {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()
			if noteExists(rows[i].sha, allowFallback) {
				rows[i].marker = "●"
			} else {
				rows[i].marker = "·"
			}
		}(i)
	}
	wg.Wait()

	w := bufio.NewWriter(os.Stdout)
	defer w.Flush()
	for _, r := range rows {
		if glossedOnly && r.marker != "●" {
			continue
		}
		fmt.Fprintf(w, "%s %s %s\n", r.marker, r.sha, r.subj)
	}
	return 0
}

// listAllGlossed enumerates every entry in refs/notes/gloss and prints
// the commit it's attached to, regardless of whether that commit is
// reachable from any branch or tag.
func listAllGlossed() int {
	out, err := exec.Command("git", "notes", "--ref="+notesRef, "list").Output()
	if err != nil {
		// No notes ref yet = nothing to list.
		return 0
	}
	raw := strings.TrimSpace(string(out))
	if raw == "" {
		return 0
	}

	type row struct {
		sha   string
		short string
		subj  string
		t     int64 // commit time; 0 for unreachable
	}
	var rows []row

	for _, line := range strings.Split(raw, "\n") {
		parts := strings.Fields(line)
		if len(parts) < 2 {
			continue
		}
		commitSHA := parts[1]
		info, err := exec.Command("git", "log", "-1", "--no-walk",
			"--format=%h\t%ct\t%s", commitSHA).Output()
		if err != nil {
			rows = append(rows, row{sha: commitSHA, short: shortSHA(commitSHA), subj: "(commit object missing)"})
			continue
		}
		fields := strings.SplitN(strings.TrimSpace(string(info)), "\t", 3)
		if len(fields) != 3 {
			continue
		}
		ts, _ := strconv.ParseInt(fields[1], 10, 64)
		rows = append(rows, row{sha: commitSHA, short: fields[0], subj: fields[2], t: ts})
	}

	// Newest first; unreachable commits (t=0) sink to the bottom.
	sort.SliceStable(rows, func(i, j int) bool { return rows[i].t > rows[j].t })

	w := bufio.NewWriter(os.Stdout)
	defer w.Flush()
	for _, r := range rows {
		fmt.Fprintf(w, "● %s %s\n", r.short, r.subj)
	}
	return 0
}

// hasGlossNote reports whether the given commit has a direct note in
// refs/notes/gloss. Does NOT use the PR-number fallback.
func hasGlossNote(rev string) bool {
	cmd := exec.Command("git", "notes", "--ref="+notesRef, "show", rev)
	cmd.Stdout = nil
	cmd.Stderr = nil
	return cmd.Run() == nil
}

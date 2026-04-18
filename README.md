# gloss

A `git` subcommand that attaches arbitrary text to a commit as a git note
in the `refs/notes/gloss` namespace.

## Install

```
brew install --HEAD AbhinavHampiholi/tap/gloss
```

Build from source:

```
git clone https://github.com/AbhinavHampiholi/gloss && cd gloss && make install
```

## Setup

Run once per repository:

```
git gloss init
```

Sets `notes.rewriteRef=refs/notes/gloss`. Does not modify remote
refspecs. Idempotent. Removes any legacy `+refs/notes/gloss:refs/notes/gloss`
refspecs added by earlier versions.

## Commands

| Command                             | Description                                                 |
|-------------------------------------|-------------------------------------------------------------|
| `git gloss init`                        | Configure the current repository.                       |
| `git gloss commit -m MSG [-C SRC]`      | `git commit`, then attach `SRC` as a note. Skipped if `SRC` is empty. |
| `git gloss list [-n N] [--glossed] [--all]` | Recent commits, prefixed `●` if glossed else `·`. `--all`: every commit in the repo that has a note, regardless of reachability. |
| `git gloss show REV`                    | `git show REV` with the note rendered inline.           |
| `git gloss note REV [--no-fallback]`    | Print the raw note body. Exit 1 if none exists.         |
| `git gloss copy [-f] FROM TO`           | Copy a note from one commit to another.                 |
| `git gloss fetch [REMOTE]`              | `git fetch REMOTE +refs/notes/gloss:refs/notes/gloss`. No-op on missing remote ref. |
| `git gloss push  [REMOTE]`              | `git push REMOTE refs/notes/gloss:refs/notes/gloss`. No-op if local ref is absent. |

Defaults: `REV` = `HEAD`, `REMOTE` = `origin`.

`-C SRC` accepts a file path or `-` for stdin. All other flags are
forwarded to `git commit` unchanged.

## Storage

Notes are plain-text blobs in the git object store, keyed by commit SHA,
under the ref `refs/notes/gloss`. No schema. Transport is explicit via
`git gloss fetch` / `git gloss push`; `git push` / `git pull` are
unaffected.

## Squash-merge fallback

When a commit is squash-merged into another branch, the new commit has a
different SHA and no note. If the commit subject ends with `(#N)` (the
convention GitHub uses for squashed PRs), `note` and `show` fall back to:

1. `gh pr view N --json headRefOid` to obtain the original head SHA.
2. A second note lookup on that SHA.

Requires `gh` on `PATH`, authenticated for the PR's repository. Silent
no-op otherwise. `note` reports the resolved source on stderr; stdout
remains the raw note bytes for pipe consumers. Pass `--no-fallback` to
`note` to disable.

`copy` is a manual escape hatch for cases where automatic resolution
isn't possible (e.g. unusual merge strategy, private PR, or the source
SHA is known by some other means).

## Transport

`git push` does not move notes. Notes live on `refs/notes/gloss` and
must be pushed explicitly:

```
git push              # commits + branches, as usual
git gloss push        # notes
```

Similarly, `git fetch` / `git pull` do not retrieve notes:

```
git gloss fetch       # pulls down notes from origin
```

To push both in one step, alias as you see fit:

```
alias gpush='git push && git gloss push'
```

## Example: manual use with Claude Code

Capture the current session as the note (macOS; substitute `xclip -o`
or `wl-paste` on Linux):

```
/export                                      # in Claude Code → clipboard
pbpaste | git gloss commit -m MSG -C -       # in shell
git gloss push                               # share with collaborators
```

Resume from a prior commit in a fresh session:

```
git gloss fetch                              # if resuming from someone else's commit
git gloss note SHA | claude
```

No additional tooling required.

## License

MIT.

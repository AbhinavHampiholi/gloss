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
| `git gloss init`                    | Configure the current repository.                           |
| `git gloss commit -m MSG [-C SRC]`  | `git commit`, then attach `SRC` as a note. Skipped if `SRC` is empty. |
| `git gloss show REV`                | `git show REV --show-notes=refs/notes/gloss`.               |
| `git gloss note REV`                | Print the raw note body. Exit 1 if none exists.             |
| `git gloss fetch [REMOTE]`          | `git fetch REMOTE +refs/notes/gloss:refs/notes/gloss`. No-op on missing remote ref. |
| `git gloss push  [REMOTE]`          | `git push REMOTE refs/notes/gloss:refs/notes/gloss`. No-op if local ref is absent. |

Defaults: `REV` = `HEAD`, `REMOTE` = `origin`.

`-C SRC` accepts a file path or `-` for stdin. All other flags are
forwarded to `git commit` unchanged.

## Storage

Notes are plain-text blobs in the git object store, keyed by commit SHA,
under the ref `refs/notes/gloss`. No schema. Transport is explicit via
`git gloss fetch` / `git gloss push`; `git push` / `git pull` are
unaffected.

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

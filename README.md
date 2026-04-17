# gloss

Attach verbose, commit-pinned context to git commits.

`gloss` is a tiny git subcommand that captures rationale — the kind of
reasoning that's too long for a commit message but evaporates if left in
chat transcripts — and stores it as a git note attached to the exact
commit it describes.

```bash
git gloss commit -m "Bound retry loop" -C rationale.md
git gloss show HEAD           # commit + diff + context inline
git gloss note HEAD           # raw context, for scripts and agents
```

## Install

```bash
brew install --HEAD AbhinavHampiholi/tap/gloss
```

Or from source:

```bash
git clone https://github.com/AbhinavHampiholi/gloss
cd gloss && make install
```

## One-time per repo

```bash
git gloss init
```

Configures `notes.rewriteRef` (so notes survive rebase) and adds fetch/push
refspecs on every remote (so notes travel with `git push`/`git fetch`).
Idempotent.

## Commands

| Command                              | What it does                                              |
|--------------------------------------|-----------------------------------------------------------|
| `git gloss init`                     | Configure the current repo for gloss.                     |
| `git gloss commit -m <msg> -C <file>`| Commit + attach `<file>` as a note.                       |
| `git gloss commit -m <msg> -C -`     | Commit + attach stdin as a note.                          |
| `git gloss commit -m <msg>`          | Commit with no note (equivalent to `git commit`).         |
| `git gloss show <sha>`               | `git show` with the note inline.                          |
| `git gloss note <sha>`               | Print the raw note. Exit 1 (silent) if there isn't one.   |

Every other `git commit` flag (`-a`, `--amend`, `--author`, `-F`, `-S`, …)
passes straight through.

## Why

Commit messages are capped short by convention. PR descriptions aren't
commit-scoped and get lost on squash. Chat transcripts with LLMs evaporate
when the session closes. But a commit SHA is the only stable, code-state-
aware identifier in the stack — attaching rationale there solves the
pinning problem by construction.

Git already has the primitives for this (notes, refspecs, rewriteRef).
They're unused in practice because `git commit` has no `--note` flag,
notes don't push/fetch by default, and they don't survive rebase. `gloss`
is the ergonomic + convention layer that makes them actually usable.

## Design

- **No LLM in the tool.** `gloss` is plumbing. The user (or an agent)
  decides what to put in the note; `gloss` just captures bytes.
- **Four verbs only.** `init`, `commit`, `show`, `note`. Everything else
  stays as plain `git`.
- **Git-native.** Installed as `git-gloss` on `$PATH`, discovered by git's
  built-in subcommand dispatch. No wrapper, no shadowing of git itself.
- **Go stdlib only.** Single static binary, no runtime dependencies.

## Agent usage

Drop this in `CLAUDE.md` / `AGENTS.md` at your repo root:

```markdown
## Committing
This repository uses `git gloss commit` instead of `git commit`.

- Use `git gloss commit -m "<subject>" -C <file>` or `-C -` (stdin).
- The context should capture rationale from this session that wouldn't
  be obvious from reading the diff alone — intent, why this approach
  over alternatives, constraints, caveats.
- Scope the context to the files in the staged diff, not the whole session.
- All other git commands work as normal.
```

## License

MIT.

# Git workflow

This repository is versioned on the `main` branch. The first commit is tagged
`baseline` and represents the known-good state before future changes.

## Daily commands

```bash
git status                         # what changed
git diff                           # unstaged changes
git diff --cached                  # staged changes
git add -A                         # stage source changes
git commit -m "feat: ..."          # create a restore point
git log --oneline --decorate       # commit history
```

## Going back

```bash
git restore --source=baseline -- path/to/file   # restore one file
git diff baseline -- path/to/file               # compare with the baseline
git revert <commit>                             # undo a commit safely
git reset --hard baseline                       # return the whole tree to the baseline
```

`git reset --hard` discards uncommitted changes. Prefer `git revert` for commits
that are already shared, and create a branch before experimenting:

```bash
git switch -c feature/my-change
```

## What is versioned

Source code, documentation, tests, `.sui` examples, run scripts and configuration
are versioned. Build outputs are ignored because they can be regenerated and
would make the repository large:

- `*.exe`, `dist/`, `examples/*/dist/`, `examples/*/bin/`
- `dist_*/`, `/testcombo2/`, `/testev/`, `/testout2/`
- backup/local files such as `*.old`, `*~`, `*.broken`, `*.pid`, `server_out.txt`

Rebuild after checking out an older revision:

```bash
go build -o simpleui.exe .
go test -count=1 ./...
simpleui build examples/process_monitor/app.sui --out examples/process_monitor/dist
```

# Git workflow

This repository is versioned on the `main` branch. The first commit is tagged
`baseline` and represents the known-good state before future changes.

## Mandatory for every agent

Every agent must commit all tracked changes before finishing a task:

```bash
git add -A
git commit -m "type(scope): summary"
git status --short          # must be empty except ignored files
```

Do not leave tracked changes uncommitted. If the task is incomplete or tests are
failing and cannot be fixed immediately, commit with a `wip:` prefix (preferably
on a branch) so the work is recoverable. Never force-add generated outputs.

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

Source code, documentation, tests, `.sui` examples, run scripts, configuration
and the `Test/testdata/` fixtures are versioned. Build outputs are ignored because
they can be regenerated and would make the repository large:

- `*.exe`, `dist/`, `examples/*/dist/`, `examples/*/bin/`
- `dist_*/`, `Test/dist_*/`, `Test/testcombo2/`, `Test/testev/`, `Test/testout2/`
- backup/local files such as `*.old`, `*~`, `*.broken`, `*.pid`, `server_out.txt`

Rebuild after checking out an older revision:

```bash
go build -o simpleui.exe .
go test -count=1 ./...
simpleui build examples/process_monitor/app.sui --out examples/process_monitor/dist
```

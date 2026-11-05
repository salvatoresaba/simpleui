# Claude Code

Read [`AGENTS.md`](AGENTS.md) first: it is the canonical workflow, rule set and
repository map for coding agents.

Then use the machine-readable index:

- [`docs/manifest.json`](docs/manifest.json) — documents, tags, invariants,
  recipes and examples;
- [`docs/AGENT_RECIPES.md`](docs/AGENT_RECIPES.md) — local and remote interface
  patterns;
- [`docs/INDEX.md`](docs/INDEX.md) — human navigation.

For a compact context bundle:

```bash
simpleui docs --agent
```

After editing a `.sui` file, always run:

```bash
simpleui check app.sui --json
```

After any change to tracked files, create a git commit before finishing
(`git add -A && git commit -m "type(scope): summary"`). Do not leave tracked
changes uncommitted. See `GIT_WORKFLOW.md`.

Do not invent syntax or bypass the security rules in `AGENTS.md`.

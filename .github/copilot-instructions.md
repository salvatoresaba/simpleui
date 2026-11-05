# SimpleUI coding instructions

Read [`AGENTS.md`](../AGENTS.md) first. It is the canonical workflow and
security rule set for coding agents.

Use the machine-readable documentation index at
[`docs/manifest.json`](../docs/manifest.json) and the recipes at
[`docs/AGENT_RECIPES.md`](../docs/AGENT_RECIPES.md). For a compact context:

```bash
simpleui docs --agent
```

After editing a `.sui` file:

```bash
simpleui check app.sui --json
```

Do not invent syntax, bypass `cssSafe`/`*Safe`/`srcAllow`, or disable the
security defaults in `AGENTS.md`.

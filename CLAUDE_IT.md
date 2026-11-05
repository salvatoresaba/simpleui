# Claude Code

Leggi prima [`AGENTS_IT.md`](AGENTS_IT.md): è la guida canonica in italiano per
gli agenti (workflow, regole e mappa del repository). La versione inglese
canonica resta [`AGENTS.md`](AGENTS.md).

Usa poi l'indice machine-readable:

- [`docs/manifest.json`](docs/manifest.json) — documenti, tag, invarianti,
  recipe ed esempi;
- [`docs/AGENT_RECIPES_IT.md`](docs/AGENT_RECIPES_IT.md) — pattern per
  interfacce locali e remote;
- [`docs/INDEX_IT.md`](docs/INDEX_IT.md) — navigazione umana in italiano.

Per un contesto compatto:

```bash
simpleui docs --agent
```

Dopo ogni modifica a un file `.sui` esegui sempre:

```bash
simpleui check app.sui --json
```

Dopo qualsiasi modifica ai file tracciati crea un commit git prima di
concludere (`git add -A && git commit -m "type(scope): summary"`). Non lasciare
modifiche tracciate non committate. Vedi `GIT_WORKFLOW.md`.

Non inventare sintassi e non aggirare le regole di sicurezza descritte in
`AGENTS_IT.md`.

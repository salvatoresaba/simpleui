# SimpleUI VS Code extension (Hermes integration)

VS Code extension for SimpleUI (`.sui`) files. It drives the SimpleUI
compiler instead of re-implementing any compiler rule.

## Commands

| Command | Action |
|---|---|
| Hermes: Check .sui | Runs `simpleui check <file> --json` and shows the diagnostics in the Problems panel. |
| Hermes: Build .sui | Runs `simpleui build <file> --out <workspace>/dist` and offers to open the generated `index.html`. |

Both commands work on the active `.sui` editor; when no `.sui` file is
active, they ask which file to use.

## Settings

| Setting | Default | Meaning |
|---|---|---|
| `simpleui.compilerPath` | `""` | Absolute path to the compiler. When empty the extension looks on `PATH`, then in the workspace root. |
| `simpleui.buildOutputDirectory` | `dist` | Output directory of the web build, relative to the workspace folder. |

## Requirements

The SimpleUI compiler must be available. Build it from the repository root:

```bash
go build -o simpleui.exe ./cmd/simpleui
```

Then put it on `PATH`, leave it in the workspace root, or point
`simpleui.compilerPath` at it.

## Build from source

```bash
npm install
npm run compile
npx --yes @vscode/vsce package --allow-missing-repository
```

`install.bat` (Windows) and `install.sh` (macOS and Linux) run the same steps.
The generated `simpleui-vscode-0.1.0.vsix` is the packaged extension.

## Install the packaged extension

In VS Code: Extensions -> ... -> Install from VSIX... and pick
`simpleui-vscode-0.1.0.vsix`.

// SimpleUI VS Code extension - Hermes integration.
//
// The extension drives the SimpleUI compiler instead of re-implementing any
// compiler rule: it runs `simpleui check --json` on the active .sui file and
// `simpleui build` on demand, then surfaces the result in the editor and in
// an output channel.

import * as fs from 'fs';
import * as path from 'path';
import { spawnSync } from 'child_process';
import * as vscode from 'vscode';

interface CheckDiagnostic {
  line: number;
  col: number;
  kind: string;
  msg: string;
  hint: string;
}

interface CheckResult {
  ok: boolean;
  file: string;
  diagnostics: CheckDiagnostic[];
}

interface Compiler {
  command: string;
}

interface RunResult {
  code: number;
  stdout: string;
  stderr: string;
}

let channel: vscode.OutputChannel;
let problems: vscode.DiagnosticCollection;

export function activate(context: vscode.ExtensionContext): void {
  channel = vscode.window.createOutputChannel('SimpleUI');
  problems = vscode.languages.createDiagnosticCollection('sui');
  context.subscriptions.push(channel, problems);
  context.subscriptions.push(
    vscode.commands.registerCommand('hermes.checkSui', () => checkSui()),
    vscode.commands.registerCommand('hermes.buildSui', () => buildSui())
  );
}

export function deactivate(): void {
  problems?.clear();
  problems?.dispose();
  channel?.dispose();
}

// checkSui type-checks the active .sui file and reports the diagnostics in
// the Problems panel.
async function checkSui(): Promise<void> {
  const file = await pickSuiFile();
  if (!file) {
    return;
  }
  const compiler = resolveCompiler();
  if (!compiler) {
    reportMissingCompiler();
    return;
  }
  channel.show(true);
  channel.appendLine(`> ${compiler.command} check ${file} --json`);
  let run: RunResult;
  try {
    run = runCompiler(compiler.command, ['check', file, '--json'], path.dirname(file));
  } catch (err) {
    channel.appendLine(`cannot start the compiler: ${String(err)}`);
    vscode.window.showErrorMessage('SimpleUI: the compiler could not be started. See the SimpleUI output channel.');
    return;
  }
  channel.appendLine(run.stdout.trim() || run.stderr.trim());
  if (run.stderr.trim()) {
    channel.appendLine(run.stderr.trim());
  }
  let parsed: CheckResult | undefined;
  try {
    parsed = JSON.parse(run.stdout.trim()) as CheckResult;
  } catch (err) {
    channel.appendLine(`cannot parse the compiler output: ${String(err)}`);
    vscode.window.showErrorMessage('SimpleUI: the compiler did not return JSON. See the SimpleUI output channel.');
    return;
  }
  publishDiagnostics(file, parsed.diagnostics || []);
  if (parsed.ok) {
    vscode.window.showInformationMessage(`SimpleUI: ${path.basename(file)} is ok`);
    return;
  }
  const errors = (parsed.diagnostics || []).filter((d) => (d.kind || 'error') === 'error').length;
  vscode.window.showErrorMessage(`SimpleUI: ${errors} error(s) in ${path.basename(file)}`);
}

// buildSui compiles the active .sui file into a web bundle.
async function buildSui(): Promise<void> {
  const file = await pickSuiFile();
  if (!file) {
    return;
  }
  const compiler = resolveCompiler();
  if (!compiler) {
    reportMissingCompiler();
    return;
  }
  const outDir = path.join(workspaceRoot() || path.dirname(file), outputDirectory());
  channel.show(true);
  channel.appendLine(`> ${compiler.command} build ${file} --out ${outDir}`);
  let run: RunResult;
  try {
    run = runCompiler(compiler.command, ['build', file, '--out', outDir], path.dirname(file));
  } catch (err) {
    channel.appendLine(`cannot start the compiler: ${String(err)}`);
    vscode.window.showErrorMessage('SimpleUI: the compiler could not be started. See the SimpleUI output channel.');
    return;
  }
  channel.appendLine(run.stdout.trim());
  if (run.stderr.trim()) {
    channel.appendLine(run.stderr.trim());
  }
  if (run.code !== 0) {
    vscode.window.showErrorMessage('SimpleUI: the build failed. See the SimpleUI output channel.');
    return;
  }
  const index = path.join(outDir, 'index.html');
  const open = await vscode.window.showInformationMessage(
    `SimpleUI: built ${path.basename(file)} into ${outDir}`,
    'Open preview'
  );
  if (open === 'Open preview') {
    await vscode.env.openExternal(vscode.Uri.file(index));
  }
}

// pickSuiFile returns the active .sui file, or asks the user to choose one.
async function pickSuiFile(): Promise<string | undefined> {
  const editor = vscode.window.activeTextEditor;
  if (editor && editor.document.fileName.toLowerCase().endsWith('.sui')) {
    return editor.document.fileName;
  }
  const picked = await vscode.window.showOpenDialog({
    canSelectMany: false,
    title: 'Select a .sui file',
    filters: { 'SimpleUI files': ['sui'] },
  });
  if (!picked || picked.length === 0) {
    return undefined;
  }
  return picked[0].fsPath;
}

// resolveCompiler finds the compiler: the configured path first, then PATH,
// then the workspace root.
function resolveCompiler(): Compiler | undefined {
  const configured = vscode.workspace
    .getConfiguration('simpleui')
    .get<string>('compilerPath', '')
    .trim();
  if (configured) {
    const resolved = path.isAbsolute(configured)
      ? configured
      : path.join(workspaceRoot() || process.cwd(), configured);
    return fs.existsSync(resolved) ? { command: resolved } : undefined;
  }
  const onPath = findOnPath('simpleui');
  if (onPath) {
    return { command: onPath };
  }
  const root = workspaceRoot();
  if (root) {
    const names = process.platform === 'win32' ? ['simpleui.exe', 'simpleui'] : ['simpleui'];
    for (const name of names) {
      const candidate = path.join(root, name);
      if (fs.existsSync(candidate)) {
        return { command: candidate };
      }
    }
  }
  return undefined;
}

// findOnPath locates an executable using PATH and, on Windows, PATHEXT.
function findOnPath(binary: string): string | undefined {
  const dirs = (process.env.PATH || '').split(path.delimiter).filter(Boolean);
  const exts =
    process.platform === 'win32'
      ? (process.env.PATHEXT || '.EXE;.CMD;.BAT').split(';').filter(Boolean)
      : [''];
  for (const dir of dirs) {
    for (const ext of exts) {
      const candidate = path.join(dir, binary + ext);
      try {
        if (fs.existsSync(candidate) && fs.statSync(candidate).isFile()) {
          return candidate;
        }
      } catch {
        // Unreadable directory entry: skip it.
      }
    }
  }
  return undefined;
}

function runCompiler(command: string, args: string[], cwd: string): RunResult {
  const res = spawnSync(command, args, { cwd, encoding: 'utf8', shell: false });
  if (res.error) {
    throw res.error;
  }
  return {
    code: typeof res.status === 'number' ? res.status : 1,
    stdout: res.stdout || '',
    stderr: res.stderr || '',
  };
}

function publishDiagnostics(file: string, diagnostics: CheckDiagnostic[]): void {
  const items = diagnostics.map((d) => {
    const line = Math.max(0, (d.line || 1) - 1);
    const col = Math.max(0, (d.col || 1) - 1);
    const range = new vscode.Range(line, col, line, col + 1);
    const severity = severityFor(d.kind);
    const message = d.hint ? `${d.msg} (${d.hint})` : d.msg;
    return new vscode.Diagnostic(range, message, severity);
  });
  problems.set(vscode.Uri.file(file), items);
}

function severityFor(kind: string): vscode.DiagnosticSeverity {
  switch ((kind || 'error').toLowerCase()) {
    case 'warning':
      return vscode.DiagnosticSeverity.Warning;
    case 'info':
      return vscode.DiagnosticSeverity.Information;
    case 'hint':
      return vscode.DiagnosticSeverity.Hint;
    default:
      return vscode.DiagnosticSeverity.Error;
  }
}

function workspaceRoot(): string | undefined {
  const folders = vscode.workspace.workspaceFolders;
  return folders && folders.length ? folders[0].uri.fsPath : undefined;
}

function outputDirectory(): string {
  return vscode.workspace.getConfiguration('simpleui').get<string>('buildOutputDirectory', 'dist');
}

function reportMissingCompiler(): void {
  vscode.window
    .showErrorMessage(
      'SimpleUI: the compiler was not found. Build it with "go build -o simpleui.exe ./cmd/simpleui" or set "simpleui.compilerPath".'
    )
    .then((choice) => {
      if (choice) {
        channel.show(true);
        channel.appendLine('The compiler was not found on PATH and not in the workspace root.');
      }
    });
}

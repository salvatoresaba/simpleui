// Hermes VS Code Extension - entry point
import * as vscode from 'vscode';

export function activate(context: vscode.ExtensionContext) {
  context.subscriptions.push(
    vscode.commands.registerCommand('hermes.buildSui', () => {
      vscode.window.showInformationMessage('Hermes: build .sui');
    }),
    vscode.commands.registerCommand('hermes.checkSui', () => {
      vscode.window.showInformationMessage('Hermes: check .sui');
    })
  );
}
export function deactivate() {}

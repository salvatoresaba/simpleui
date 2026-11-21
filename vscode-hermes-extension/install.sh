#!/bin/sh
# Build and package the SimpleUI VS Code extension.
set -e
cd "$(dirname "$0")"
if [ ! -d node_modules ]; then
  echo "Installing dependencies..."
  npm install
fi
echo "Compiling..."
npm run compile
echo "Packaging..."
npx --yes @vscode/vsce package --allow-missing-repository
echo "Ready: install the .vsix in VS Code via Extensions > Install from VSIX"

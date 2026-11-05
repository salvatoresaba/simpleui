#!/bin/bash
set -e
cd "$(dirname "$0")"
if [ ! -d node_modules ]; then npm install; fi
echo "Compiling..."
npm run compile
echo "Packaging..."
npx vsce package
echo "Install the .vsix in VS Code: Extensions > Install from VSIX"

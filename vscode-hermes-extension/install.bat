@echo off
REM Auto-install Hermes VS Code extension
cd /d "%~dp0"
if not exist node_modules (
  echo Installing dependencies...
  npm install
)
echo Compiling...
npm run compile
echo Packaging...
npx vsce package
if exist *.vsix (
  echo Ready: install the .vsix in VS Code via Extensions > Install from VSIX
) else (
  echo Package failed.
)

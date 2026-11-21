@echo off
REM Build and package the SimpleUI VS Code extension.
cd /d "%~dp0"
if not exist node_modules (
  echo Installing dependencies...
  npm install
)
echo Compiling...
call npm run compile
if errorlevel 1 (
  echo Compile failed.
  exit /b 1
)
echo Packaging...
call npx --yes @vscode/vsce package --allow-missing-repository
if exist *.vsix (
  echo Ready: install the .vsix in VS Code via Extensions > Install from VSIX
) else (
  echo Package failed.
  exit /b 1
)

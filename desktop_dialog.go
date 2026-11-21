package sui

// desktop_dialog.go - source for the native file/folder picker included in
// every desktop build. The generated program uses OS dialogs:
//
//	Windows: PowerShell + System.Windows.Forms (OpenFileDialog/FolderBrowserDialog)
//	macOS:   osascript + choose file/choose folder
//	Linux:   zenity, with kdialog as fallback
//
// The picker is user-driven and protected by the same per-run bridge token as
// the other desktop endpoints. It only returns paths; fs/exec capabilities are
// still enforced separately.
const desktopDialogSource = `package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os/exec"
	"runtime"
	"strings"
)

func makePickFile(secret string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if postForm(r) != nil || !secretOk(secret, r) {
			writePickJSON(w, false, nil, "forbidden: missing or invalid token")
			return
		}
		multiple := r.PostFormValue("multiple") == "1"
		paths, err := pickFiles(r.Context(), multiple)
		if err != nil {
			writePickJSON(w, false, nil, "dialog.pickFile: "+err.Error())
			return
		}
		if !multiple && len(paths) > 1 {
			paths = paths[:1]
		}
		writePickJSON(w, true, paths, "")
	}
}

func makePickDirectory(secret string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if postForm(r) != nil || !secretOk(secret, r) {
			writePickJSON(w, false, nil, "forbidden: missing or invalid token")
			return
		}
		path, err := pickDirectory(r.Context())
		if err != nil {
			writePickJSON(w, false, nil, "dialog.pickDirectory: "+err.Error())
			return
		}
		var data []string
		if path != "" {
			data = []string{path}
		}
		writePickJSON(w, true, data, "")
	}
}

func writePickJSON(w http.ResponseWriter, ok bool, data []string, message string) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"ok":      ok,
		"data":    data,
		"message": message,
	})
}

func pickFiles(ctx context.Context, multiple bool) ([]string, error) {
	switch runtime.GOOS {
	case "windows":
		return pickWindowsFiles(ctx, multiple)
	case "darwin":
		return pickDarwinFiles(ctx, multiple)
	default:
		return pickLinuxFiles(ctx, multiple)
	}
}

func pickDirectory(ctx context.Context) (string, error) {
	switch runtime.GOOS {
	case "windows":
		return pickWindowsDirectory(ctx)
	case "darwin":
		return pickDarwinDirectory(ctx)
	default:
		return pickLinuxDirectory(ctx)
	}
}

// ---- Windows ---------------------------------------------------------------

func windowsFileScript(multiple bool) string {
	multi := "$false"
	if multiple {
		multi = "$true"
	}
	return "Add-Type -AssemblyName System.Windows.Forms\n" +
		"$d = New-Object System.Windows.Forms.OpenFileDialog\n" +
		"$d.Multiselect = " + multi + "\n" +
		"$d.CheckFileExists = $true\n" +
		"$d.Title = 'Select file'\n" +
		"if ($d.ShowDialog() -eq [System.Windows.Forms.DialogResult]::OK) {\n" +
		"  ConvertTo-Json -InputObject @($d.FileNames) -Compress\n" +
		"} else { '[]' }\n"
}

func windowsDirectoryScript() string {
	return "Add-Type -AssemblyName System.Windows.Forms\n" +
		"$d = New-Object System.Windows.Forms.FolderBrowserDialog\n" +
		"$d.Description = 'Select folder'\n" +
		"$d.ShowNewFolderButton = $true\n" +
		"if ($d.ShowDialog() -eq [System.Windows.Forms.DialogResult]::OK) {\n" +
		"  ConvertTo-Json -InputObject @($d.SelectedPath) -Compress\n" +
		"} else { '[]' }\n"
}

func runPowerShell(ctx context.Context, script string) ([]byte, error) {
	name := "powershell.exe"
	if _, err := exec.LookPath(name); err != nil {
		name = "pwsh"
		if _, err := exec.LookPath(name); err != nil {
			return nil, errors.New("PowerShell is not available")
		}
	}
	cmd := exec.CommandContext(ctx, name, "-NoProfile", "-NonInteractive", "-STA", "-ExecutionPolicy", "Bypass", "-Command", script)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return out, fmt.Errorf("powershell dialog: %v: %s", err, strings.TrimSpace(string(out)))
	}
	return out, nil
}

func pickWindowsFiles(ctx context.Context, multiple bool) ([]string, error) {
	out, err := runPowerShell(ctx, windowsFileScript(multiple))
	if err != nil {
		return nil, err
	}
	return parsePathJSON(out)
}

func pickWindowsDirectory(ctx context.Context) (string, error) {
	out, err := runPowerShell(ctx, windowsDirectoryScript())
	if err != nil {
		return "", err
	}
	paths, err := parsePathJSON(out)
	if err != nil || len(paths) == 0 {
		return "", err
	}
	return paths[0], nil
}

func parsePathJSON(out []byte) ([]string, error) {
	s := strings.TrimSpace(string(out))
	if s == "" {
		return nil, nil
	}
	var arr []string
	if err := json.Unmarshal([]byte(s), &arr); err == nil {
		return cleanPaths(arr), nil
	}
	var one string
	if err := json.Unmarshal([]byte(s), &one); err == nil {
		if strings.TrimSpace(one) == "" {
			return nil, nil
		}
		return []string{one}, nil
	}
	return nil, fmt.Errorf("invalid dialog output")
}

func cleanPaths(paths []string) []string {
	out := make([]string, 0, len(paths))
	for _, p := range paths {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

// ---- macOS ----------------------------------------------------------------

func darwinFileScript(multiple bool) string {
	if multiple {
		return strings.Join([]string{
			"try",
			"set fs to choose file with prompt \"Select files\" with multiple selections allowed",
			"set out to \"\"",
			"repeat with f in fs",
			"set out to out & POSIX path of f & linefeed",
			"end repeat",
			"return out",
			"on error number -128",
			"return \"\"",
			"end try",
		}, "\n")
	}
	return strings.Join([]string{
		"try",
		"set f to choose file with prompt \"Select file\"",
		"return POSIX path of f",
		"on error number -128",
		"return \"\"",
		"end try",
	}, "\n")
}

func darwinDirectoryScript() string {
	return strings.Join([]string{
		"try",
		"set d to choose folder with prompt \"Select folder\"",
		"return POSIX path of d",
		"on error number -128",
		"return \"\"",
		"end try",
	}, "\n")
}

func runOSAScript(ctx context.Context, script string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, "osascript", "-e", script)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return out, fmt.Errorf("osascript dialog: %v: %s", err, strings.TrimSpace(string(out)))
	}
	return out, nil
}

func pickDarwinFiles(ctx context.Context, multiple bool) ([]string, error) {
	out, err := runOSAScript(ctx, darwinFileScript(multiple))
	if err != nil {
		return nil, err
	}
	return splitNativePaths(out), nil
}

func pickDarwinDirectory(ctx context.Context) (string, error) {
	out, err := runOSAScript(ctx, darwinDirectoryScript())
	if err != nil {
		return "", err
	}
	paths := splitNativePaths(out)
	if len(paths) == 0 {
		return "", nil
	}
	return paths[0], nil
}

func splitNativePaths(out []byte) []string {
	s := strings.TrimSpace(string(out))
	if s == "" {
		return nil
	}
	lines := strings.Split(s, "\n")
	return cleanPaths(lines)
}

// ---- Linux ----------------------------------------------------------------

func pickLinuxFiles(ctx context.Context, multiple bool) ([]string, error) {
	if zenity, err := exec.LookPath("zenity"); err == nil {
		args := []string{"--file-selection", "--title=Select file"}
		if multiple {
			args = append(args, "--multiple", "--separator=\n")
		}
		out, err := exec.CommandContext(ctx, zenity, args...).CombinedOutput()
		if err != nil {
			if commandCancelled(err) {
				return nil, nil
			}
			return nil, fmt.Errorf("zenity dialog: %v: %s", err, strings.TrimSpace(string(out)))
		}
		return splitNativePaths(out), nil
	}
	if kdialog, err := exec.LookPath("kdialog"); err == nil {
		args := []string{"--getopenfilename", "--title", "Select file"}
		if multiple {
			args = append(args, "--multiple")
		}
		out, err := exec.CommandContext(ctx, kdialog, args...).CombinedOutput()
		if err != nil {
			if commandCancelled(err) {
				return nil, nil
			}
			return nil, fmt.Errorf("kdialog dialog: %v: %s", err, strings.TrimSpace(string(out)))
		}
		return splitNativePaths(out), nil
	}
	return nil, errors.New("no native file dialog found (install zenity or kdialog)")
}

func pickLinuxDirectory(ctx context.Context) (string, error) {
	if zenity, err := exec.LookPath("zenity"); err == nil {
		out, err := exec.CommandContext(ctx, zenity, "--file-selection", "--directory", "--title=Select folder").CombinedOutput()
		if err != nil {
			if commandCancelled(err) {
				return "", nil
			}
			return "", fmt.Errorf("zenity dialog: %v: %s", err, strings.TrimSpace(string(out)))
		}
		paths := splitNativePaths(out)
		if len(paths) == 0 {
			return "", nil
		}
		return paths[0], nil
	}
	if kdialog, err := exec.LookPath("kdialog"); err == nil {
		out, err := exec.CommandContext(ctx, kdialog, "--getexistingdirectory", "--title", "Select folder").CombinedOutput()
		if err != nil {
			if commandCancelled(err) {
				return "", nil
			}
			return "", fmt.Errorf("kdialog dialog: %v: %s", err, strings.TrimSpace(string(out)))
		}
		paths := splitNativePaths(out)
		if len(paths) == 0 {
			return "", nil
		}
		return paths[0], nil
	}
	return "", errors.New("no native folder dialog found (install zenity or kdialog)")
}

func commandCancelled(err error) bool {
	var exitErr *exec.ExitError
	return errors.As(err, &exitErr) && exitErr.ExitCode() == 1
}
`

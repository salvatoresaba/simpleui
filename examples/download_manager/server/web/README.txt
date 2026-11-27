The compiled SimpleUI app lives here.

build.ps1 (and build.bat) write index.html, style.css and app.js in this
directory and the Go server embeds them, so a single executable carries the
app and the download manager.

This placeholder keeps `go build` working before the first frontend build.

// userdata — a tiny "server executable": it answers the table data as JSON
// on standard output, so the SimpleUI desktop client can launch it through
// the bridge (`exec.run`) instead of calling an HTTP server.
//
//	go build -o userdata.exe .
//
// The SUI client (../app.sui) runs `exec.run("userdata.exe", [])`, parses the
// JSON with json.parse and renders the rows in a table. Keep this binary next
// to the compiled desktop app: the bridge resolves relative command names
// against the app's own directory.
package main

import "fmt"

func main() {
	rows := `[["1","Alice","alice@example.dev"],` +
		`["2","Bob","bob@example.dev"],` +
		`["3","Carol","carol@example.dev"],` +
		`["4","Dan","dan@example.dev"]]`
	fmt.Println(rows)
}

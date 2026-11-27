package core

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// The example must show the URL of the file to download and add a row even
// when the backend does not answer: the operator has a field to type the URL
// in any case, and the "+" button always adds a row.
func TestDownloadManagerTargetRowWithoutTheBackend(t *testing.T) {
	src, err := os.ReadFile(filepath.Join("examples", "download_manager", "app.sui"))
	if err != nil {
		t.Fatalf("read the example: %v", err)
	}
	// The backend of the example is not running during the test: the port
	// points to a closed one, so the app has to work without an answer.
	code := strings.ReplaceAll(string(src), "127.0.0.1:8090", "127.0.0.1:9")
	runHeadlessFeature(t, code, featureScriptHead+`
    await wait(700);
    var rows = document.querySelectorAll('.target-row');
    if (rows.length < 1) throw new Error('the row of the target is missing');
    var url = rows[0].querySelector('.target-url');
    if (!url || url.tagName !== 'INPUT') throw new Error('the field of the URL is missing');
    var before = rows.length;
    var add = document.querySelector('.btn-add');
    if (!add) throw new Error('the add button is missing');
    add.click();
    await wait(600);
    var after = document.querySelectorAll('.target-row').length;
    if (after !== before + 1) throw new Error('the add button did not add a row: ' + before + ' -> ' + after);
`+featureScriptTail)
}

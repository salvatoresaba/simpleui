const fs = require('fs');
const p = 'C:\\Users\\ingsa\\Downloads\\GO_PROJECT_LANG\\jsgen.go';
let s = fs.readFileSync(p, 'utf8');
const bad = `\tif e.tag == "row" {
\t\tsb.WriteString(" class=\\"row" + classSuffix(e) + "\\"")
\t} else if e.tag == "col" {
\t\tsb.WriteString(" class=\\"col" + classSuffix(e) + "\\"")
\t} else if e.tag == "spacer" {
\t\tsb.WriteString(" class=\\"spacer" + classSuffix(e) + "\\"")
\t} else {
\t\t// all other elements handled by writeUserAttrs (classes/types for new controls)
\t\twriteUserAttrs(&sb, e, ctx)
\t}
\t\tif e.tag == "input" && len(e.text) > 0 {
\t\t\tsb.WriteString(" placeholder=\\"")
\t\t\tpartsContentInto(&sb, e.text)
\t\t\tsb.WriteString("\\"")
\t\t}
\t\twriteUserAttrs(&sb, e, ctx)
\t}
\tif !genCompBody {`;
const good = `\tif e.tag == "row" {
\t\tsb.WriteString(" class=\\"row" + classSuffix(e) + "\\"")
\t} else if e.tag == "col" {
\t\tsb.WriteString(" class=\\"col" + classSuffix(e) + "\\"")
\t} else if e.tag == "spacer" {
\t\tsb.WriteString(" class=\\"spacer" + classSuffix(e) + "\\"")
\t} else {
\t\t// all other elements handled by writeUserAttrs (classes/types for new controls)
\t\twriteUserAttrs(&sb, e, ctx)
\t}
\tif e.tag == "input" && len(e.text) > 0 {
\t\tsb.WriteString(" placeholder=\\"")
\t\tpartsContentInto(&sb, e.text)
\t\tsb.WriteString("\\"")
\t}
\tif !genCompBody {`;
if (s.indexOf(bad) >= 0) {
  s = s.split(bad).join(good);
  fs.writeFileSync(p, s);
  console.log('fixed');
} else {
  console.log('notfound');
}

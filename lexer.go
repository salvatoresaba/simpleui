package main

// lexer.go — the SimpleUI tokenizer (SPEC §2.1). One flat package `main`;
// see SPEC.md §1 for layout and §7 for Go idioms. This file is owned by
// Agent A (front-end).

// Token is one lexical token. `text` is the literal text: for numbers the
// literal digits, for strings the RAW inner text (escapes are NOT decoded
// here; the parser decodes them), for `err` tokens the diagnostic message.
type Token struct {
	kind   string // see the token-kind list below
	text   string
	line   int // 1-based
	col    int // 1-based
	offset int // byte offset of the token start in the source
}

// keywordOf returns the token kind for a keyword, or "" if w is a plain
// identifier.
func keywordOf(w string) string {
	if w == "state" {
		return "state"
	}
	if w == "type" {
		return "type"
	}
	if w == "struct" {
		return "struct"
	}
	if w == "fn" {
		return "fn"
	}
	if w == "async" {
		return "async"
	}
	if w == "if" {
		return "if"
	}
	if w == "else" {
		return "else"
	}
	if w == "loop" {
		return "loop"
	}
	if w == "while" {
		return "while"
	}
	if w == "in" {
		return "in"
	}
	if w == "return" {
		return "return"
	}
	if w == "break" {
		return "break"
	}
	if w == "continue" {
		return "continue"
	}
	if w == "let" {
		return "let"
	}
	if w == "import" {
		return "import"
	}
	if w == "from" {
		return "from"
	}
	if w == "view" {
		return "view"
	}
	if w == "title" {
		return "title"
	}
	if w == "true" {
		return "true"
	}
	if w == "false" {
		return "false"
	}
	if w == "null" {
		return "null"
	}
	if w == "and" {
		return "and"
	}
	if w == "or" {
		return "or"
	}
	if w == "not" {
		return "not"
	}
	if w == "css" {
		return "css"
	}
	return ""
}

func isDigitByte(c byte) bool {
	return c >= '0' && c <= '9'
}

func isAlphaByte(c byte) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z')
}

func isIdentStart(c byte) bool {
	return isAlphaByte(c) || c == '_'
}

func isIdentPart(c byte) bool {
	return isAlphaByte(c) || isDigitByte(c) || c == '_'
}

// Lex turns a source string into a token stream. Newlines are significant
// (each one is emitted as a `newline` token), comments are dropped, and the
// final token is always `eof`. Lexical problems (unknown characters and
// unterminated strings) yield `err` tokens plus diagnostics and lexing
// continues, so the parser can recover.
func Lex(source string, file string) ([]Token, DiagList) {
	var toks []Token
	var dl DiagList
	dl.Kind("lexical")
	n := len(source)
	line := 1
	col := 1
	i := 0
	for i < n {
		c := source[i]
		if c == '\n' {
			toks = append(toks, Token{kind: "newline", text: "", line: line, col: col, offset: i})
			i++
			line++
			col = 1
		} else if c == ' ' || c == '\t' || c == '\r' {
			i++
			col++
		} else if c == '#' {
			// Comment: run to end of line (the newline itself will be emitted
			// on the next iteration).
			for i < n && source[i] != '\n' {
				i++
				col++
			}
		} else if isIdentStart(c) {
			start := i
			sl := line
			sc := col
			so := i
			i++
			col++
			for i < n && isIdentPart(source[i]) {
				i++
				col++
			}
			w := source[start:i]
			k := keywordOf(w)
			if k == "" {
				k = "ident"
			}
			toks = append(toks, Token{kind: k, text: w, line: sl, col: sc, offset: so})
		} else if isDigitByte(c) {
			start := i
			sl := line
			sc := col
			so := i
			for i < n && isDigitByte(source[i]) {
				i++
				col++
			}
			if i < n && source[i] == '.' && i+1 < n && isDigitByte(source[i+1]) {
				// digits.digits is a float
				i++
				col++
				for i < n && isDigitByte(source[i]) {
					i++
					col++
				}
				toks = append(toks, Token{kind: "float", text: source[start:i], line: sl, col: sc, offset: so})
			} else if i < n && source[i] == '.' && (i+1 >= n || source[i+1] != '.') {
				// "1." with no digits after the dot (a leading ".." would be
				// the range operator, so that case is left for the next
				// iteration).
				i++
				col++
				dl.Add(file, sl, sc, "malformed number: '.' must be followed by digits")
				toks = append(toks, Token{kind: "err", text: "malformed number literal", line: sl, col: sc, offset: so})
			} else {
				toks = append(toks, Token{kind: "int", text: source[start:i], line: sl, col: sc, offset: so})
			}
		} else if c == '"' {
			start := i
			sl := line
			sc := col
			so := i
			i++
			col++
			closed := false
			for i < n {
				ch := source[i]
				if ch == '"' {
					closed = true
					i++
					col++
					break
				}
				if ch == '\\' {
					// Skip the escaped character too (covers \" and anything
					// else); the raw text keeps both characters.
					i++
					col++
					if i < n {
						i++
						col++
					}
				} else if ch == '\n' {
					// Newlines inside strings are allowed lexically; the
					// parser reports them as errors.
					i++
					line++
					col = 1
				} else {
					i++
					col++
				}
			}
			if closed {
				toks = append(toks, Token{kind: "str", text: source[start+1 : i-1], line: sl, col: sc, offset: so})
			} else {
				dl.AddEx(file, sl, sc, "", "unterminated string literal", "add a closing \" at the end of the string")
				toks = append(toks, Token{kind: "err", text: "unterminated string literal", line: sl, col: sc, offset: so})
			}
		} else {
			sl := line
			sc := col
			so := i
			kind := ""
			width := 1
			if i+1 < n {
				two := source[i : i+2]
				if two == "->" {
					kind = "->"
					width = 2
				} else if two == ".." {
					kind = ".."
					width = 2
				} else if two == "==" {
					kind = "=="
					width = 2
				} else if two == "!=" {
					kind = "!="
					width = 2
				} else if two == "<=" {
					kind = "<="
					width = 2
				} else if two == ">=" {
					kind = ">="
					width = 2
				} else if two == "+=" {
					kind = "+="
					width = 2
				} else if two == "++" {
					kind = "++"
					width = 2
				} else if two == "-=" {
					kind = "-="
					width = 2
				} else if two == "--" {
					kind = "--"
					width = 2
				} else if two == "*=" {
					kind = "*="
					width = 2
				} else if two == "/=" {
					kind = "/="
					width = 2
				} else if two == "%=" {
					kind = "%="
					width = 2
				}
			}
			if kind == "" {
				if c == '(' {
					kind = "("
				} else if c == ')' {
					kind = ")"
				} else if c == '{' {
					kind = "{"
				} else if c == '}' {
					kind = "}"
				} else if c == '[' {
					kind = "["
				} else if c == ']' {
					kind = "]"
				} else if c == '=' {
					kind = "="
				} else if c == '<' {
					kind = "<"
				} else if c == '>' {
					kind = ">"
				} else if c == '+' {
					kind = "+"
				} else if c == '-' {
					kind = "-"
				} else if c == '*' {
					kind = "*"
				} else if c == '/' {
					kind = "/"
				} else if c == '%' {
					kind = "%"
				} else if c == '.' {
					kind = "."
				} else if c == ',' {
					kind = ","
				} else if c == ':' {
					kind = ":"
				} else if c == '?' {
					kind = "?"
				}
			}
			if kind == "" {
				msg := "unexpected character"
				hint := "simpleui does not understand this character; remove it or replace it with a valid operator"
				if c == '{' || c == '}' {
					hint = "braces inside strings must be written as \\{ and \\}"
				}
				dl.AddEx(file, sl, sc, "", msg, hint)
				toks = append(toks, Token{kind: "err", text: msg, line: sl, col: sc, offset: so})
				i++
				col++
			} else {
				toks = append(toks, Token{kind: kind, text: source[so : so+width], line: sl, col: sc, offset: so})
				i += width
				col += width
			}
		}
	}
	toks = append(toks, Token{kind: "eof", text: "", line: line, col: col, offset: i})
	return toks, dl
}

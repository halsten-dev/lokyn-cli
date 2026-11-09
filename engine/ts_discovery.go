package engine

import (
	"bytes"
	"strconv"
	"strings"
	"unicode/utf8"
)

// typeScriptKeyFunctions maps the frontend translation helpers, which all take
// the key first, to whether the key is plural.
var typeScriptKeyFunctions = map[string]bool{"t": false, "tKey": false, "tp": true}

// expressionKeywords are followed by an operand, so a '/' or '<' after them
// starts a regular expression or a JSX element.
var expressionKeywords = map[string]bool{
	"return": true, "case": true, "yield": true, "await": true, "typeof": true,
	"void": true, "delete": true, "throw": true, "in": true, "of": true, "else": true,
}

type scanMode int

const (
	code         scanMode = iota // Code outside any template or JSX.
	templateCode                 // Code in "${…}", returning to template text.
	jsxCode                      // Code in a JSX "{…}", returning to the tag or children.
	jsxTag                       // Between a JSX tag's '<' and its '>' or "/>".
	jsxChildren                  // Text between an element's tags, where quotes are plain text.
)

// scanFrame is an open construct; depth is the brace depth at which the "}"
// closing a code frame returns to its enclosing frame.
type scanFrame struct {
	mode  scanMode
	depth int
}

// typeScriptFile scans one file. It is a lexical scan rather than a parse: the
// helpers are plain functions, so telling code from comments, strings,
// templates, regular expressions and JSX text is enough to find their calls in
// a single pass.
type typeScriptFile struct {
	keys *DiscoveredKeys
	vars *DiscoveredVars
	file string
	src  []byte
	jsx  bool

	frames []scanFrame
	depth  int
	prev   byte   // Last significant code byte; tells an operand from an operator.
	word   []byte // The identifier just read, if prev ends it.

	// line is the line at lineOffset; calls arrive in source order, so lines
	// are counted incrementally and the scan stays linear.
	line       int
	lineOffset int
}

func discoverTypeScript(keys *DiscoveredKeys, vars *DiscoveredVars, file string, src []byte) {
	f := typeScriptFile{keys: keys, vars: vars, file: file, src: src, jsx: strings.HasSuffix(file, ".tsx"), line: 1}
	for i := 0; i < len(src); {
		switch f.mode() {
		case jsxTag:
			i = f.tag(i)
		case jsxChildren:
			i = f.children(i)
		default:
			i = f.codeToken(i)
		}
	}
}

func (f *typeScriptFile) mode() scanMode {
	if len(f.frames) == 0 {
		return code
	}
	return f.frames[len(f.frames)-1].mode
}

func (f *typeScriptFile) push(mode scanMode) {
	f.frames = append(f.frames, scanFrame{mode, f.depth})
	if mode == templateCode || mode == jsxCode {
		f.prev, f.word = '{', nil
	}
}

func (f *typeScriptFile) pop() {
	f.frames = f.frames[:len(f.frames)-1]
}

// endElement closes the innermost JSX element. Back in code, the whole element
// was an operand, so a following '/' or '<' is an operator.
func (f *typeScriptFile) endElement() {
	f.pop()
	if mode := f.mode(); mode != jsxTag && mode != jsxChildren {
		f.prev, f.word = ')', nil
	}
}

// operandExpected reports whether the code read so far ends with an operator,
// so that what follows starts an operand.
func (f *typeScriptFile) operandExpected() bool {
	if f.word != nil {
		return expressionKeywords[string(f.word)]
	}
	return f.prev == 0 || strings.IndexByte("(,=:[!&|?{};+-*%<>~^", f.prev) >= 0
}

// codeToken reads the code token at i and returns the index past it.
func (f *typeScriptFile) codeToken(i int) int {
	src := f.src
	if end, ok := commentEnd(src, i); ok {
		return end
	}
	c := src[i]
	if isIdentifierByte(c) {
		end := identifierEnd(src, i)
		next := end
		plural, isKeyFunction := typeScriptKeyFunctions[string(src[i:end])]
		if isKeyFunction && f.prev != '.' && string(f.word) != "function" {
			next = f.call(end, plural)
		}
		f.word, f.prev = src[i:end], src[end-1]
		return next
	}
	next := i + 1
	switch {
	case c == '"' || c == '\'':
		next = quotedEnd(src, i)
	case c == '`':
		return f.enterTemplateText(i + 1)
	case c == '/' && f.operandExpected():
		next = regexEnd(src, i)
	case c == '<' && f.jsx && f.operandExpected() && jsxTagStart(src, i+1):
		f.push(jsxTag)
		return i + 1
	case c == '{':
		f.depth++
	case c == '}':
		if top := len(f.frames) - 1; top >= 0 && f.frames[top].depth == f.depth {
			mode := f.frames[top].mode
			f.pop()
			if mode == templateCode {
				return f.enterTemplateText(i + 1)
			}
			return i + 1
		}
		f.depth--
	}
	if !isSpace(c) {
		f.prev, f.word = c, nil
	}
	return next
}

// tag reads inside a JSX tag, where attribute strings have no escapes.
func (f *typeScriptFile) tag(i int) int {
	switch c := f.src[i]; {
	case c == '{':
		f.push(jsxCode)
	case c == '"' || c == '\'':
		if end := bytes.IndexByte(f.src[i+1:], c); end >= 0 {
			return i + end + 2
		}
		return len(f.src)
	case c == '/' && i+1 < len(f.src) && f.src[i+1] == '>':
		f.endElement()
		return i + 2
	case c == '>':
		f.frames[len(f.frames)-1].mode = jsxChildren
	}
	return i + 1
}

// children reads JSX text up to the next expression or tag.
func (f *typeScriptFile) children(i int) int {
	switch f.src[i] {
	case '{':
		f.push(jsxCode)
	case '<':
		if i+1 < len(f.src) && f.src[i+1] == '/' {
			f.endElement()
			if end := bytes.IndexByte(f.src[i:], '>'); end >= 0 {
				return i + end + 1
			}
			return len(f.src)
		}
		f.push(jsxTag)
	default:
		if end := bytes.IndexAny(f.src[i:], "{<"); end >= 0 {
			return i + end
		}
		return len(f.src)
	}
	return i + 1
}

// jsxTagStart reports whether a '<' before i opens a JSX element rather than
// a generic arrow function's type parameters: <T,>, <T = U> or <T extends U>.
func jsxTagStart(src []byte, i int) bool {
	switch {
	case i >= len(src):
		return false
	case src[i] == '>':
		return true // A fragment.
	case !isIdentifierByte(src[i]) || '0' <= src[i] && src[i] <= '9':
		return false
	}
	next := skipSpace(src, identifierEnd(src, i))
	if next >= len(src) || src[next] == ',' || src[next] == '=' {
		return false
	}
	word := identifierEnd(src, next)
	return string(src[next:word]) != "extends" || word >= len(src) || !isSpace(src[word])
}

// enterTemplateText skips template text from i and, if it stops at a
// substitution, enters its code.
func (f *typeScriptFile) enterTemplateText(i int) int {
	end, open := templateText(f.src, i)
	if open {
		f.push(templateCode)
	} else {
		f.prev, f.word = '`', nil
	}
	return end
}

// call records the call of a key function whose name ends at i. It returns
// where scanning resumes: past a literal key, or at the arguments so that calls
// nested in them are found too.
func (f *typeScriptFile) call(i int, plural bool) int {
	open := skipSpace(f.src, i)
	if open >= len(f.src) || f.src[open] != '(' {
		return i
	}
	argument := skipSpace(f.src, open+1)
	if next, ok := f.literalKey(argument, plural); ok {
		return next
	}
	expression := string(bytes.TrimSpace(f.src[argument:expressionEnd(f.src, argument)]))
	if expression != "" {
		f.line += bytes.Count(f.src[f.lineOffset:argument], []byte{'\n'})
		f.lineOffset = argument
		appendDiscoveredVar(f.vars, DiscoveredVar{
			Var: Var(f.file + ":" + expression), Expression: expression,
			File: f.file, Line: f.line, IsPlural: plural,
		})
	}
	return i
}

// literalKey records a call whose first argument, starting at i, is a string or
// an interpolation-free template and nothing else.
func (f *typeScriptFile) literalKey(i int, plural bool) (next int, ok bool) {
	src := f.src
	if i >= len(src) {
		return i, false
	}
	var end int
	switch src[i] {
	case '"', '\'':
		end = quotedEnd(src, i)
	case '`':
		var open bool
		if end, open = templateText(src, i+1); open {
			return i, false
		}
	default:
		return i, false
	}
	if end-1 <= i || src[end-1] != src[i] {
		return i, false // Unterminated.
	}
	if after := skipSpace(src, end); after >= len(src) || (src[after] != ',' && src[after] != ')') {
		return i, false
	}
	appendDiscoveredKey(f.keys, unescapeTypeScript(src[i+1:end-1]), plural, nil)
	return end, true
}

// templateText scans template text from i to just past its closing backtick,
// or to just past a "${" with open set.
func templateText(src []byte, i int) (end int, open bool) {
	for ; i < len(src); i++ {
		switch src[i] {
		case '\\':
			i++
		case '`':
			return i + 1, false
		case '$':
			if i+1 < len(src) && src[i+1] == '{' {
				return i + 2, true
			}
		}
	}
	return len(src), false
}

// expressionEnd returns the index of the first ',' or closing bracket at or
// after i that is not nested inside the expression.
func expressionEnd(src []byte, i int) int {
	depth := 0
	for i < len(src) {
		if end, ok := commentEnd(src, i); ok {
			i = end
			continue
		}
		switch c := src[i]; c {
		case '"', '\'':
			i = quotedEnd(src, i)
			continue
		case '`':
			i = templateEnd(src, i+1)
			continue
		case '(', '[', '{':
			depth++
		case ')', ']', '}', ',':
			if depth == 0 {
				return i
			}
			if c != ',' {
				depth--
			}
		}
		i++
	}
	return i
}

// templateEnd returns the index just past the template starting at i,
// including any substitutions.
func templateEnd(src []byte, i int) int {
	for {
		end, open := templateText(src, i)
		if !open {
			return end
		}
		i = expressionEnd(src, end) + 1
	}
}

// commentEnd reports whether a comment starts at i and returns the index past it.
func commentEnd(src []byte, i int) (int, bool) {
	if src[i] != '/' || i+1 >= len(src) {
		return i, false
	}
	switch src[i+1] {
	case '/':
		if end := bytes.IndexByte(src[i:], '\n'); end >= 0 {
			return i + end, true
		}
		return len(src), true
	case '*':
		if end := bytes.Index(src[i+2:], []byte("*/")); end >= 0 {
			return i + 2 + end + 2, true
		}
		return len(src), true
	}
	return i, false
}

// quotedEnd returns the index just past the string starting at i. Strings
// cannot span lines, so an unterminated one ends at its line's end.
func quotedEnd(src []byte, i int) int {
	quote := src[i]
	for i++; i < len(src); i++ {
		switch src[i] {
		case '\\':
			i++
		case quote:
			return i + 1
		case '\n':
			return i
		}
	}
	return len(src)
}

// regexEnd returns the index just past the regular expression starting at i;
// a '/' inside a character class does not close it.
func regexEnd(src []byte, i int) int {
	inClass := false
	for i++; i < len(src); i++ {
		switch src[i] {
		case '\\':
			i++
		case '[':
			inClass = true
		case ']':
			inClass = false
		case '/':
			if !inClass {
				return i + 1
			}
		case '\n':
			return i
		}
	}
	return len(src)
}

func skipSpace(src []byte, i int) int {
	for i < len(src) && isSpace(src[i]) {
		i++
	}
	return i
}

func identifierEnd(src []byte, i int) int {
	for i < len(src) && isIdentifierByte(src[i]) {
		i++
	}
	return i
}

func isSpace(c byte) bool {
	return c == ' ' || c == '\t' || c == '\n' || c == '\r'
}

func isIdentifierByte(c byte) bool {
	return c == '_' || c == '$' || c >= utf8.RuneSelf ||
		'a' <= c && c <= 'z' || 'A' <= c && c <= 'Z' || '0' <= c && c <= '9'
}

func unescapeTypeScript(raw []byte) string {
	if bytes.IndexByte(raw, '\\') < 0 {
		return string(raw)
	}
	var text strings.Builder
	text.Grow(len(raw))
	for i := 0; i < len(raw); i++ {
		if raw[i] != '\\' || i+1 == len(raw) {
			text.WriteByte(raw[i])
			continue
		}
		i++
		escape := raw[i]
		if simple := strings.IndexByte("nrtbfv0", escape); simple >= 0 {
			text.WriteByte("\n\r\t\b\f\v\x00"[simple])
			continue
		}
		var digits []byte
		consumed := 0
		switch {
		case escape == 'x' && i+2 < len(raw):
			digits, consumed = raw[i+1:i+3], 2
		case escape == 'u' && i+1 < len(raw) && raw[i+1] == '{':
			if end := bytes.IndexByte(raw[i:], '}'); end >= 0 {
				digits, consumed = raw[i+2:i+end], end
			}
		case escape == 'u' && i+4 < len(raw):
			digits, consumed = raw[i+1:i+5], 4
		}
		code, err := strconv.ParseUint(string(digits), 16, 32)
		switch {
		case digits != nil && err == nil:
			text.WriteRune(rune(code))
			i += consumed
		case escape == '\r':
			// A line continuation vanishes, including a CRLF one.
			if i+1 < len(raw) && raw[i+1] == '\n' {
				i++
			}
		case escape != '\n':
			text.WriteByte(escape) // Any other escaped character stands for itself.
		}
	}
	return text.String()
}

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

// typeScriptFile carries the state that outlives one call of a key function.
type typeScriptFile struct {
	keys *DiscoveredKeys
	vars *DiscoveredVars
	file string
	src  []byte
	// line is the line at lineOffset; calls arrive in source order, so lines
	// are counted incrementally and the scan stays linear.
	line       int
	lineOffset int
}

// discoverTypeScript is a lexical scan rather than a parse: the helpers are
// plain functions, so skipping comments, strings, templates and regular
// expressions is enough to find their calls in a single pass.
func discoverTypeScript(keys *DiscoveredKeys, vars *DiscoveredVars, file string, src []byte) {
	f := typeScriptFile{keys: keys, vars: vars, file: file, src: src, line: 1}
	// substitutions holds the brace depth at which each open template
	// substitution ("${ … }") returns to template text.
	var substitutions []int
	depth := 0
	var prev byte // Last significant code byte; tells a regular expression from a division.
	var word []byte
	for i := 0; i < len(src); {
		c := src[i]
		if end, ok := commentEnd(src, i); ok {
			i = end
			continue
		}
		if isIdentifierByte(c) {
			end := i + 1
			for end < len(src) && isIdentifierByte(src[end]) {
				end++
			}
			next := end
			plural, isKeyFunction := typeScriptKeyFunctions[string(src[i:end])]
			if isKeyFunction && prev != '.' && string(word) != "function" {
				next = f.call(end, plural)
			}
			word, prev, i = src[i:end], src[end-1], next
			continue
		}
		switch {
		case c == '"' || c == '\'':
			i = quotedEnd(src, i)
		case c == '`':
			i = templateTextEnd(src, i+1, &substitutions, depth)
		case c == '/' && (prev == 0 || strings.IndexByte("(,=:[!&|?{};+-*%<>~^", prev) >= 0):
			i = regexEnd(src, i)
		case c == '}' && len(substitutions) > 0 && substitutions[len(substitutions)-1] == depth:
			substitutions = substitutions[:len(substitutions)-1]
			i = templateTextEnd(src, i+1, &substitutions, depth)
		default:
			switch c {
			case '{':
				depth++
			case '}':
				depth--
			}
			i++
		}
		if c != ' ' && c != '\t' && c != '\n' && c != '\r' {
			prev, word = c, nil
		}
	}
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
	if next, ok := literalKey(f.keys, f.src, argument, plural); ok {
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
func literalKey(keys *DiscoveredKeys, src []byte, i int, plural bool) (next int, ok bool) {
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
	appendDiscoveredKey(keys, unescapeTypeScript(src[i+1:end-1]), plural, nil)
	return end, true
}

// templateTextEnd skips template text from i and, if it stops at a
// substitution, records the brace depth that closes it.
func templateTextEnd(src []byte, i int, substitutions *[]int, depth int) int {
	end, open := templateText(src, i)
	if open {
		*substitutions = append(*substitutions, depth)
	}
	return end
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
// cannot span lines, so an apostrophe in JSX text only hides the rest of its line.
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
	for i < len(src) && (src[i] == ' ' || src[i] == '\t' || src[i] == '\n' || src[i] == '\r') {
		i++
	}
	return i
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

// Package markdown converts a constrained subset of Markdown into the
// ClickUp "rich text" comment block array documented at
// https://developer.clickup.com/docs/comment-formatting.md.
//
// ClickUp's comment API accepts a `comment` field: an array of blocks shaped
// like {"text": "...", "attributes": {...}}. Inline formatting (bold,
// italic, code, links) is carried directly on the run that has the text.
// Block-level formatting (lists, code blocks, blockquotes) is carried on a
// separate "\n" run's attributes, mirroring the Quill Delta format ClickUp
// builds on: each source line ends with a "\n" run whose attributes (if any)
// mark the whole line as part of a bullet/ordered list, a code block, or a
// blockquote.
//
// Blockquote attribute: undocumented at the time this package was written
// (2026-09-22). Confirmed empirically by reading the raw `comment` array of
// an existing ClickUp comment that renders as a blockquote (task 868kft8kg,
// "GPT-5 Codex for Jeremy", 2026-09-11): the trailing "\n" run carries
// {"blockquote": {}} — an empty object, the same shape family as
// {"list": {"list": "bullet"}} and {"code-block": {"code-block": "plain"}}.
// See LEARNINGS.md for the raw JSON excerpt.
package markdown

import (
	"regexp"
	"strings"
)

// Block is one element of a ClickUp rich-text `comment` array.
type Block struct {
	Text       string         `json:"text"`
	Attributes map[string]any `json:"attributes,omitempty"`
}

// PlainText reconstructs a readable plain-text approximation of a block
// array by concatenating each run's text (including the literal "\n" runs
// that carry block attributes). It drops all formatting, so it's only meant
// for display fallback (e.g. echoing a created comment back to the user)
// when the API response doesn't include a rendered comment_text.
func PlainText(blocks []Block) string {
	var sb strings.Builder
	for _, b := range blocks {
		sb.WriteString(b.Text)
	}

	return sb.String()
}

var (
	orderedListRe = regexp.MustCompile(`^(\d+)\.\s+(.*)$`)
	headingRe     = regexp.MustCompile(`^#{1,6}\s+(.*)$`)
	fenceRe       = regexp.MustCompile("^```")
	linkDetectRe  = regexp.MustCompile(`\[[^\]]+\]\([^)]+\)`)
	bulletLineRe  = regexp.MustCompile(`(?m)^(-|\*)\s`)
	orderedLineRe = regexp.MustCompile(`(?m)^\d+\.\s`)
	quoteLineRe   = regexp.MustCompile(`(?m)^>`)
)

// HasMarkdown reports whether text contains any of the markdown markers this
// package converts: **bold**, `code`, fenced code blocks, "> " blockquotes,
// "- "/"* " bullets, "1. " ordered items, or [text](url) links. It does not
// trigger on bare "*"/"_" (italics) or bare URLs alone, since those are
// common in ordinary prose and would cause false positives.
func HasMarkdown(text string) bool {
	if strings.Contains(text, "**") {
		return true
	}

	if strings.Contains(text, "`") {
		return true
	}

	if strings.Contains(text, "```") {
		return true
	}

	if quoteLineRe.MatchString(text) {
		return true
	}

	if bulletLineRe.MatchString(text) {
		return true
	}

	if orderedLineRe.MatchString(text) {
		return true
	}

	if linkDetectRe.MatchString(text) {
		return true
	}

	return false
}

// ToBlocks converts markdown source into a ClickUp rich-text block array.
// Unrecognized syntax is passed through as literal text.
func ToBlocks(src string) []Block {
	lines := strings.Split(strings.ReplaceAll(src, "\r\n", "\n"), "\n")

	// Drop a single trailing blank line produced by a trailing "\n" in the
	// source; it doesn't represent a meaningful extra paragraph break.
	if len(lines) > 1 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}

	var blocks []Block

	inFence := false

	for i, line := range lines {
		last := i == len(lines)-1

		if fenceRe.MatchString(strings.TrimSpace(line)) {
			inFence = !inFence
			continue // fence delimiter lines are markers only, not content
		}

		var (
			content    string
			blockAttrs map[string]any
		)

		switch {
		case inFence:
			content = line
			blockAttrs = map[string]any{"code-block": map[string]any{"code-block": "plain"}}

			appendInlineLiteral(&blocks, content)
		case strings.HasPrefix(line, ">"):
			content, blockAttrs = parseBlockquoteLine(line)
			appendInline(&blocks, content)
		case headingRe.MatchString(line):
			m := headingRe.FindStringSubmatch(line)
			content = m[1]
			appendInlineHeading(&blocks, content)
		default:
			content, blockAttrs = parseListLine(line)
			appendInline(&blocks, content)
		}

		// Emit the trailing "\n" run. Per the documented examples, a final
		// plain paragraph line needs no trailing "\n" run at all (the API
		// implies it); every other line (non-final, or carrying block
		// attributes) needs an explicit one to encode the line break or the
		// block type.
		if last && len(blockAttrs) == 0 {
			continue
		}

		blocks = append(blocks, Block{Text: "\n", Attributes: blockAttrs})
	}

	return blocks
}

// parseBlockquoteLine strips a leading "> " (or bare ">") marker and detects
// a nested list marker inside the quoted content, returning the remaining
// text and the merged block attributes.
func parseBlockquoteLine(line string) (string, map[string]any) {
	rest := strings.TrimPrefix(line, ">")
	rest = strings.TrimPrefix(rest, " ")

	attrs := map[string]any{"blockquote": map[string]any{}}

	content, listAttrs := parseListLine(rest)
	if listAttrs != nil {
		attrs["list"] = listAttrs["list"]
	}

	return content, attrs
}

// parseListLine detects a leading bullet ("- "/"* ") or ordered ("1. ")
// marker and returns the remaining content and, if a marker was found, the
// block attributes for it.
func parseListLine(line string) (string, map[string]any) {
	switch {
	case strings.HasPrefix(line, "- "):
		return strings.TrimPrefix(line, "- "), map[string]any{"list": map[string]any{"list": "bullet"}}
	case strings.HasPrefix(line, "* "):
		return strings.TrimPrefix(line, "* "), map[string]any{"list": map[string]any{"list": "bullet"}}
	default:
		if m := orderedListRe.FindStringSubmatch(line); m != nil {
			return m[2], map[string]any{"list": map[string]any{"list": "ordered"}}
		}

		return line, nil
	}
}

// appendInline parses inline markdown (bold, italic, code, links, bare URLs)
// in content and appends the resulting runs to blocks. Empty content (e.g. a
// blank quoted line) produces no runs.
func appendInline(blocks *[]Block, content string) {
	if content == "" {
		return
	}

	*blocks = append(*blocks, parseInline(content)...)
}

// appendInlineLiteral appends content as a single literal run with no inline
// parsing, used for fenced code block lines.
func appendInlineLiteral(blocks *[]Block, content string) {
	if content == "" {
		return
	}

	*blocks = append(*blocks, Block{Text: content, Attributes: map[string]any{}})
}

// appendInlineHeading parses inline markdown in content and forces bold on
// every resulting run, since ClickUp has no heading attribute.
func appendInlineHeading(blocks *[]Block, content string) {
	if content == "" {
		return
	}

	runs := parseInline(content)
	for i := range runs {
		if runs[i].Attributes == nil {
			runs[i].Attributes = map[string]any{}
		}

		runs[i].Attributes["bold"] = true
	}

	*blocks = append(*blocks, runs...)
}

// parseInline scans a single line of text (block markers already stripped)
// for **bold**, *italic*/_italic_, `code`, [text](url) links, and bare URLs.
// Unrecognized syntax (including unmatched delimiters) passes through as
// plain text.
func parseInline(s string) []Block {
	var (
		runs []Block
		buf  strings.Builder
	)

	flush := func() {
		if buf.Len() == 0 {
			return
		}

		runs = append(runs, Block{Text: buf.String(), Attributes: map[string]any{}})
		buf.Reset()
	}

	i := 0
	for i < len(s) {
		switch {
		case s[i] == '`':
			if end := strings.IndexByte(s[i+1:], '`'); end >= 0 {
				flush()

				text := s[i+1 : i+1+end]
				runs = append(runs, Block{Text: text, Attributes: map[string]any{"code": true}})
				i = i + 1 + end + 1

				continue
			}

		case strings.HasPrefix(s[i:], "**"):
			if end := strings.Index(s[i+2:], "**"); end > 0 {
				flush()

				text := s[i+2 : i+2+end]
				runs = append(runs, Block{Text: text, Attributes: map[string]any{"bold": true}})
				i = i + 2 + end + 2

				continue
			}

		case s[i] == '*' && i+1 < len(s) && s[i+1] != ' ':
			if end := findClosing(s, i+1, '*'); end > 0 {
				flush()

				text := s[i+1 : end]
				runs = append(runs, Block{Text: text, Attributes: map[string]any{"italic": true}})
				i = end + 1

				continue
			}

		case s[i] == '_' && i+1 < len(s) && s[i+1] != ' ':
			if end := findClosing(s, i+1, '_'); end > 0 {
				flush()

				text := s[i+1 : end]
				runs = append(runs, Block{Text: text, Attributes: map[string]any{"italic": true}})
				i = end + 1

				continue
			}

		case s[i] == '[':
			if text, url, next, ok := parseLink(s, i); ok {
				flush()

				runs = append(runs, Block{Text: text, Attributes: map[string]any{"link": url}})
				i = next

				continue
			}

		case strings.HasPrefix(s[i:], "http://") || strings.HasPrefix(s[i:], "https://"):
			end := bareURLEnd(s, i)

			flush()

			url := s[i:end]
			runs = append(runs, Block{Text: url, Attributes: map[string]any{"link": url}})
			i = end

			continue
		}

		buf.WriteByte(s[i])
		i++
	}

	flush()

	return runs
}

// findClosing returns the index of the next occurrence of marker at or after
// from, requiring the character immediately before it not be a space (so
// "* word *" doesn't close on the trailing space), and content non-empty.
// Returns -1 if no valid closing marker is found.
func findClosing(s string, from int, marker byte) int {
	for j := from; j < len(s); j++ {
		if s[j] == marker && j > from && s[j-1] != ' ' {
			return j
		}
	}

	return -1
}

// parseLink attempts to parse a [text](url) link starting at index i (where
// s[i] == '['). It returns the link text, url, the index just past the
// closing ')', and whether a well-formed link was found.
func parseLink(s string, i int) (text, url string, next int, ok bool) {
	closeBracket := strings.IndexByte(s[i:], ']')
	if closeBracket < 0 {
		return "", "", 0, false
	}

	closeBracket += i

	if closeBracket+1 >= len(s) || s[closeBracket+1] != '(' {
		return "", "", 0, false
	}

	closeParen := strings.IndexByte(s[closeBracket+2:], ')')
	if closeParen < 0 {
		return "", "", 0, false
	}

	closeParen += closeBracket + 2

	return s[i+1 : closeBracket], s[closeBracket+2 : closeParen], closeParen + 1, true
}

// bareURLEnd finds the end of a bare URL starting at i: scans until
// whitespace or a trailing punctuation mark that's likely sentence
// punctuation rather than part of the URL.
func bareURLEnd(s string, i int) int {
	j := i
	for j < len(s) && !strings.ContainsRune(" \t\n", rune(s[j])) {
		j++
	}

	for j > i && strings.ContainsRune(".,;:!?)", rune(s[j-1])) {
		j--
	}

	return j
}

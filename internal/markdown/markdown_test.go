package markdown

import (
	"encoding/json"
	"testing"
)

func TestHasMarkdown(t *testing.T) {
	cases := []struct {
		name string
		text string
		want bool
	}{
		{"plain", "just some text", false},
		{"bold", "hello **world**", true},
		{"inline code", "run `go build`", true},
		{"fenced code", "```\ncode\n```", true},
		{"blockquote", "> quoted", true},
		{"bullet", "- item", true},
		{"star bullet", "* item", true},
		{"ordered", "1. first", true},
		{"link", "see [docs](https://example.com)", true},
		{"bare url only", "see https://example.com", false},
		{"lone underscore", "file_name.txt", false},
		{"lone asterisk", "5*3=15", false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := HasMarkdown(tc.text); got != tc.want {
				t.Errorf("HasMarkdown(%q) = %v, want %v", tc.text, got, tc.want)
			}
		})
	}
}

func TestToBlocks(t *testing.T) {
	cases := []struct {
		name string
		src  string
		want []Block
	}{
		{
			name: "plain paragraph",
			src:  "plain text",
			want: []Block{
				{Text: "plain text", Attributes: map[string]any{}},
			},
		},
		{
			name: "bold",
			src:  "**bold text**",
			want: []Block{
				{Text: "bold text", Attributes: map[string]any{"bold": true}},
			},
		},
		{
			name: "italic star",
			src:  "*italics*",
			want: []Block{
				{Text: "italics", Attributes: map[string]any{"italic": true}},
			},
		},
		{
			name: "italic underscore",
			src:  "_italics_",
			want: []Block{
				{Text: "italics", Attributes: map[string]any{"italic": true}},
			},
		},
		{
			name: "inline code",
			src:  "inline `code`",
			want: []Block{
				{Text: "inline ", Attributes: map[string]any{}},
				{Text: "code", Attributes: map[string]any{"code": true}},
			},
		},
		{
			name: "link",
			src:  "see [docs](https://example.com)",
			want: []Block{
				{Text: "see ", Attributes: map[string]any{}},
				{Text: "docs", Attributes: map[string]any{"link": "https://example.com"}},
			},
		},
		{
			name: "bare url",
			src:  "see https://example.com for more",
			want: []Block{
				{Text: "see ", Attributes: map[string]any{}},
				{Text: "https://example.com", Attributes: map[string]any{"link": "https://example.com"}},
				{Text: " for more", Attributes: map[string]any{}},
			},
		},
		{
			name: "bullet list",
			src:  "bulleted list\n- item one\n- item two",
			want: []Block{
				{Text: "bulleted list", Attributes: map[string]any{}},
				{Text: "\n", Attributes: nil},
				{Text: "item one", Attributes: map[string]any{}},
				{Text: "\n", Attributes: map[string]any{"list": map[string]any{"list": "bullet"}}},
				{Text: "item two", Attributes: map[string]any{}},
				{Text: "\n", Attributes: map[string]any{"list": map[string]any{"list": "bullet"}}},
			},
		},
		{
			name: "star bullet list",
			src:  "* item one",
			want: []Block{
				{Text: "item one", Attributes: map[string]any{}},
				{Text: "\n", Attributes: map[string]any{"list": map[string]any{"list": "bullet"}}},
			},
		},
		{
			name: "ordered list",
			src:  "1. first\n2. second",
			want: []Block{
				{Text: "first", Attributes: map[string]any{}},
				{Text: "\n", Attributes: map[string]any{"list": map[string]any{"list": "ordered"}}},
				{Text: "second", Attributes: map[string]any{}},
				{Text: "\n", Attributes: map[string]any{"list": map[string]any{"list": "ordered"}}},
			},
		},
		{
			name: "fenced code block",
			src:  "```\nline one\nline two\n```",
			want: []Block{
				{Text: "line one", Attributes: map[string]any{}},
				{Text: "\n", Attributes: map[string]any{"code-block": map[string]any{"code-block": "plain"}}},
				{Text: "line two", Attributes: map[string]any{}},
				{Text: "\n", Attributes: map[string]any{"code-block": map[string]any{"code-block": "plain"}}},
			},
		},
		{
			name: "single line blockquote",
			src:  "> quoted text",
			want: []Block{
				{Text: "quoted text", Attributes: map[string]any{}},
				{Text: "\n", Attributes: map[string]any{"blockquote": map[string]any{}}},
			},
		},
		{
			name: "multi-line blockquote with blank quoted line",
			src:  "> line one\n>\n> line two",
			want: []Block{
				{Text: "line one", Attributes: map[string]any{}},
				{Text: "\n", Attributes: map[string]any{"blockquote": map[string]any{}}},
				{Text: "\n", Attributes: map[string]any{"blockquote": map[string]any{}}},
				{Text: "line two", Attributes: map[string]any{}},
				{Text: "\n", Attributes: map[string]any{"blockquote": map[string]any{}}},
			},
		},
		{
			name: "heading becomes bold",
			src:  "## Section Title",
			want: []Block{
				{Text: "Section Title", Attributes: map[string]any{"bold": true}},
			},
		},
		{
			name: "unknown syntax passes through",
			src:  "some ~strikethrough~ text",
			want: []Block{
				{Text: "some ~strikethrough~ text", Attributes: map[string]any{}},
			},
		},
		{
			name: "unmatched bold marker passes through as text",
			src:  "**unterminated bold",
			want: []Block{
				{Text: "**unterminated bold", Attributes: map[string]any{}},
			},
		},
		{
			name: "the skill template",
			src: "> 🦄 **Fable 5.1 for Jeremy**\n" +
				">\n" +
				"> **Status:** In progress\n" +
				">\n" +
				"> Test of rich text: **bold**, `code`, [link](https://example.com).\n" +
				"> - bullet one\n" +
				"> - bullet two",
			want: []Block{
				{Text: "🦄 ", Attributes: map[string]any{}},
				{Text: "Fable 5.1 for Jeremy", Attributes: map[string]any{"bold": true}},
				{Text: "\n", Attributes: map[string]any{"blockquote": map[string]any{}}},
				{Text: "\n", Attributes: map[string]any{"blockquote": map[string]any{}}},
				{Text: "Status:", Attributes: map[string]any{"bold": true}},
				{Text: " In progress", Attributes: map[string]any{}},
				{Text: "\n", Attributes: map[string]any{"blockquote": map[string]any{}}},
				{Text: "\n", Attributes: map[string]any{"blockquote": map[string]any{}}},
				{Text: "Test of rich text: ", Attributes: map[string]any{}},
				{Text: "bold", Attributes: map[string]any{"bold": true}},
				{Text: ", ", Attributes: map[string]any{}},
				{Text: "code", Attributes: map[string]any{"code": true}},
				{Text: ", ", Attributes: map[string]any{}},
				{Text: "link", Attributes: map[string]any{"link": "https://example.com"}},
				{Text: ".", Attributes: map[string]any{}},
				{Text: "\n", Attributes: map[string]any{"blockquote": map[string]any{}}},
				{Text: "bullet one", Attributes: map[string]any{}},
				{Text: "\n", Attributes: map[string]any{"blockquote": map[string]any{}, "list": map[string]any{"list": "bullet"}}},
				{Text: "bullet two", Attributes: map[string]any{}},
				{Text: "\n", Attributes: map[string]any{"blockquote": map[string]any{}, "list": map[string]any{"list": "bullet"}}},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := ToBlocks(tc.src)

			gotJSON, err := json.Marshal(got)
			if err != nil {
				t.Fatalf("marshal got: %v", err)
			}

			wantJSON, err := json.Marshal(tc.want)
			if err != nil {
				t.Fatalf("marshal want: %v", err)
			}

			if string(gotJSON) != string(wantJSON) {
				t.Errorf("ToBlocks(%q) =\n%s\nwant\n%s", tc.src, gotJSON, wantJSON)
			}
		})
	}
}

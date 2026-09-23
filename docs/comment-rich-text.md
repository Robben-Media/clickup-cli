# Comment rich-text formatting

ClickUp does not parse markdown out of `comment_text`. To render bold,
italics, code, links, lists, code blocks, or blockquotes, a comment must
instead be sent as a `comment` array of rich-text blocks:
`{"text": "...", "attributes": {...}}`. Documented at
https://developer.clickup.com/docs/comment-formatting.md.

Implementation: `internal/markdown` converts a constrained subset of
markdown into this block array; `internal/cmd/comments.go` wires it into
`comments add/update/reply/add-list/add-view` with `--plain-text` /
`--markdown` overrides. `internal/clickup.Comment.Raw` exposes the raw
`comment` block array from API responses (`comments list --json` and
`comments replies --json`) for inspection.

## The blockquote attribute (undocumented)

ClickUp's public docs (as of 2026-09-22) cover `bold`, `italic`, `code`,
`link`, `list` (`bullet`/`ordered`/`checked`/`unchecked`/`toggled`), and
`code-block`, but not blockquotes. The attribute was confirmed empirically
by reading the raw `comment` block array of an existing comment that
renders as a blockquote (task `868kft8kg`, "GPT-5 Codex for Jeremy",
2026-09-11, `clickup-cli comments list 868kft8kg --json`):

```json
{
  "comment": [
    { "text": "🦄 ", "attributes": {} },
    { "text": "GPT-5 Codex for Jeremy", "attributes": { "bold": true } },
    { "text": "\n", "attributes": { "blockquote": {} } },
    { "text": "Status:", "attributes": { "bold": true } },
    { "text": " In progress", "attributes": {} },
    { "text": "\n", "attributes": { "blockquote": {} } },
    { "text": "Description and dates were still written as a 9 September review. ...", "attributes": {} },
    { "text": "\n", "attributes": { "blockquote": {} } }
  ]
}
```

The blockquote attribute is `{"blockquote": {}}`: an empty object on the
trailing `"\n"` run, the same shape family as `{"list": {"list": "bullet"}}`
and `{"code-block": {"code-block": "plain"}}`. It marks the whole preceding
line as part of a blockquote.

Nested list-inside-blockquote lines (e.g. `> - bullet one`) combine both
attributes on the same `"\n"` run: `{"blockquote": {}, "list": {"list":
"bullet"}}`. This combination isn't documented either; it was verified by
posting a live test comment to task 868kmugt0 with the skill template
below, reading the raw blocks back, and confirming ClickUp rendered a
bulleted list nested inside the blockquote (then deleting the test
comment).

```
> 🦄 **Fable 5.1 for Jeremy**
>
> **Status:** In progress
>
> Test of rich text: **bold**, `code`, [link](https://example.com).
> - bullet one
> - bullet two
```

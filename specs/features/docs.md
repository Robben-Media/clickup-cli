# Docs (v3)

## Overview

Implement the ClickUp Docs v3 API — search, fetch, create docs and manage pages within docs.

**Why**: Docs are ClickUp's native document system. CLI access enables automated documentation management, content creation, and page editing.

**Requires**: v3 base URL support and workspace ID configuration.

## API Endpoints

| Method | Path | Summary | Operation ID | Version |
|--------|------|---------|--------------|---------|
| GET | /api/v3/workspaces/{}/docs | Search for Docs | searchDocsPublic | v3 |
| GET | /api/v3/workspaces/{}/docs/{} | Fetch a Doc | getDocPublic | v3 |
| GET | /api/v3/workspaces/{}/docs/{}/page_listing | Fetch PageListing | getDocPageListingPublic | v3 |
| GET | /api/v3/workspaces/{}/docs/{}/pages | Fetch Pages | getDocPagesPublic | v3 |
| GET | /api/v3/workspaces/{}/docs/{}/pages/{} | Get page | getPagePublic | v3 |
| POST | /api/v3/workspaces/{}/docs | Create a Doc | createDocPublic | v3 |
| POST | /api/v3/workspaces/{}/docs/{}/pages | Create a Page | createPagePublic | v3 |
| PUT | /api/v3/workspaces/{}/docs/{}/pages/{} | Edit a Page | editPagePublic | v3 |

## User Stories

### US-001: Search Docs

**CLI Command:** `clickup docs search [--workspace <id>] [--query "..."]`

**JSON Output:**
```json
{"docs": [{"id": "doc_123", "name": "API Guide", "date_created": 1700000000000, "creator": {"id": 1}}]}
```

**Plain Output (TSV):** Headers: `ID	NAME	CREATED	CREATOR_ID`

### US-002: Get Doc

**CLI Command:** `clickup docs get <doc_id>`

### US-003: Page Listing

**CLI Command:** `clickup docs page-listing <doc_id>`

**Plain Output (TSV):** Headers: `PAGE_ID	TITLE	ORDER`

### US-004: Get Pages / Single Page

**CLI Commands:**
- `clickup docs pages <doc_id>` — all pages
- `clickup docs page <doc_id> <page_id>` — single page with content

### US-005: Create Doc

**CLI Command:** `clickup docs create [--workspace <id>] --name <name> [--parent-type <space|folder|list>] [--parent-id <id>]`

### US-006: Create Page

**CLI Command:** `clickup docs create-page <doc_id> --name <name> [--content "..."] [--content-format <md|html>]`

### US-007: Edit Page

**CLI Command:** `clickup docs edit-page <doc_id> <page_id> [--name "..."] [--content "..."] [--content-format <md|html>]`

## Request/Response Types

```go
type Doc struct {
    ID          string `json:"id"`
    Name        string `json:"name"`
    DateCreated int64  `json:"date_created,omitempty"`
    Creator     *User  `json:"creator,omitempty"`
}

type DocPage struct {
    ID      string `json:"id"`
    Name    string `json:"name"`
    Content string `json:"content,omitempty"`
    Order   int    `json:"order,omitempty"`
}

type DocsResponse struct {
    Docs []Doc `json:"docs"`
}

type DocPagesResponse struct {
    Pages []DocPage `json:"pages"`
}

type CreateDocRequest struct {
    Name       string `json:"name"`
    ParentType string `json:"parent_type,omitempty"`
    ParentID   string `json:"parent_id,omitempty"`
}

type CreatePageRequest struct {
    Name          string `json:"name"`
    Content       string `json:"content,omitempty"`
    ContentFormat string `json:"content_format,omitempty"` // "md" or "html"
}

type EditPageRequest struct {
    Name          string `json:"name,omitempty"`
    Content       string `json:"content,omitempty"`
    ContentFormat string `json:"content_format,omitempty"`
}
```

## Edge Cases

- Page content may be markdown or HTML depending on `content_format`
- Editing a page replaces entire content — no partial updates
- Page listing returns structure (ID + title) without content
- Doc creation can be scoped to space/folder/list via parent refs

## Feedback Loops

### Unit Tests
```go
func TestDocsService_Search(t *testing.T)     { /* search docs */ }
func TestDocsService_Get(t *testing.T)        { /* single doc */ }
func TestDocsService_PageListing(t *testing.T) { /* page listing */ }
func TestDocsService_Pages(t *testing.T)      { /* all pages */ }
func TestDocsService_Page(t *testing.T)       { /* single page */ }
func TestDocsService_Create(t *testing.T)     { /* create doc */ }
func TestDocsService_CreatePage(t *testing.T) { /* create page */ }
func TestDocsService_EditPage(t *testing.T)   { /* edit page */ }
```

## Technical Requirements

- New `DocsService` on `clickup.Client`
- All paths use `v3Path()` helper
- Edit page uses PUT (not PATCH like other v3 updates)
- Content format defaults to markdown if not specified

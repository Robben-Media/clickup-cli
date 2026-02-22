package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/builtbyrobben/clickup-cli/internal/clickup"
	"github.com/builtbyrobben/clickup-cli/internal/outfmt"
)

type DocsCmd struct {
	Search      DocsSearchCmd      `cmd:"" help:"Search for docs"`
	Get         DocsGetCmd         `cmd:"" help:"Get a doc by ID"`
	PageListing DocsPageListingCmd `cmd:"" help:"Get page listing for a doc"`
	Pages       DocsPagesCmd       `cmd:"" help:"Get all pages in a doc"`
	Page        DocsPageCmd        `cmd:"" help:"Get a single page"`
	Create      DocsCreateCmd      `cmd:"" help:"Create a new doc"`
	CreatePage  DocsCreatePageCmd  `cmd:"" help:"Create a page in a doc"`
	EditPage    DocsEditPageCmd    `cmd:"" help:"Edit a page"`
}

// DocsSearchCmd searches for docs.
type DocsSearchCmd struct {
	Query string `help:"Search query"`
}

func (cmd *DocsSearchCmd) Run(ctx context.Context) error {
	client, err := getClickUpClient(ctx)
	if err != nil {
		return err
	}

	result, err := client.Docs().Search(ctx, cmd.Query)
	if err != nil {
		return err
	}

	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(os.Stdout, result)
	}
	if outfmt.IsPlain(ctx) {
		headers := []string{"ID", "NAME", "CREATED", "CREATOR_ID"}
		rows := make([][]string, 0, len(result.Docs))
		for _, d := range result.Docs {
			creatorID := ""
			if d.Creator != nil {
				creatorID = fmt.Sprintf("%d", d.Creator.ID)
			}
			rows = append(rows, []string{d.ID, d.Name, fmt.Sprintf("%d", d.DateCreated), creatorID})
		}
		return outfmt.WritePlain(os.Stdout, headers, rows)
	}

	if len(result.Docs) == 0 {
		fmt.Fprintln(os.Stderr, "No docs found")
		return nil
	}

	fmt.Fprintf(os.Stderr, "Found %d docs\n\n", len(result.Docs))
	for _, d := range result.Docs {
		fmt.Printf("%s: %s\n", d.ID, d.Name)
	}

	return nil
}

// DocsGetCmd gets a single doc.
type DocsGetCmd struct {
	DocID string `arg:"" required:"" help:"Doc ID"`
}

func (cmd *DocsGetCmd) Run(ctx context.Context) error {
	client, err := getClickUpClient(ctx)
	if err != nil {
		return err
	}

	result, err := client.Docs().Get(ctx, cmd.DocID)
	if err != nil {
		return err
	}

	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(os.Stdout, result)
	}
	if outfmt.IsPlain(ctx) {
		headers := []string{"ID", "NAME", "CREATED"}
		rows := [][]string{{result.ID, result.Name, fmt.Sprintf("%d", result.DateCreated)}}
		return outfmt.WritePlain(os.Stdout, headers, rows)
	}

	fmt.Printf("ID: %s\n", result.ID)
	fmt.Printf("Name: %s\n", result.Name)
	if result.DateCreated > 0 {
		fmt.Printf("Created: %d\n", result.DateCreated)
	}
	if result.Creator != nil {
		fmt.Printf("Creator: %s\n", result.Creator.Username)
	}

	return nil
}

// DocsPageListingCmd gets the page listing for a doc.
type DocsPageListingCmd struct {
	DocID string `arg:"" required:"" help:"Doc ID"`
}

func (cmd *DocsPageListingCmd) Run(ctx context.Context) error {
	client, err := getClickUpClient(ctx)
	if err != nil {
		return err
	}

	result, err := client.Docs().GetPageListing(ctx, cmd.DocID)
	if err != nil {
		return err
	}

	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(os.Stdout, result)
	}
	if outfmt.IsPlain(ctx) {
		headers := []string{"PAGE_ID", "TITLE", "ORDER"}
		rows := make([][]string, 0, len(result.Pages))
		for _, p := range result.Pages {
			rows = append(rows, []string{p.ID, p.Name, fmt.Sprintf("%d", p.Order)})
		}
		return outfmt.WritePlain(os.Stdout, headers, rows)
	}

	if len(result.Pages) == 0 {
		fmt.Fprintln(os.Stderr, "No pages found")
		return nil
	}

	fmt.Fprintf(os.Stderr, "Found %d pages\n\n", len(result.Pages))
	for _, p := range result.Pages {
		fmt.Printf("%s: %s (order %d)\n", p.ID, p.Name, p.Order)
	}

	return nil
}

// DocsPagesCmd gets all pages in a doc.
type DocsPagesCmd struct {
	DocID string `arg:"" required:"" help:"Doc ID"`
}

func (cmd *DocsPagesCmd) Run(ctx context.Context) error {
	client, err := getClickUpClient(ctx)
	if err != nil {
		return err
	}

	result, err := client.Docs().GetPages(ctx, cmd.DocID)
	if err != nil {
		return err
	}

	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(os.Stdout, result)
	}
	if outfmt.IsPlain(ctx) {
		headers := []string{"PAGE_ID", "TITLE", "ORDER"}
		rows := make([][]string, 0, len(result.Pages))
		for _, p := range result.Pages {
			rows = append(rows, []string{p.ID, p.Name, fmt.Sprintf("%d", p.Order)})
		}
		return outfmt.WritePlain(os.Stdout, headers, rows)
	}

	if len(result.Pages) == 0 {
		fmt.Fprintln(os.Stderr, "No pages found")
		return nil
	}

	fmt.Fprintf(os.Stderr, "Found %d pages\n\n", len(result.Pages))
	for _, p := range result.Pages {
		fmt.Printf("%s: %s\n", p.ID, p.Name)
	}

	return nil
}

// DocsPageCmd gets a single page from a doc.
type DocsPageCmd struct {
	DocID  string `arg:"" required:"" help:"Doc ID"`
	PageID string `arg:"" required:"" help:"Page ID"`
}

func (cmd *DocsPageCmd) Run(ctx context.Context) error {
	client, err := getClickUpClient(ctx)
	if err != nil {
		return err
	}

	result, err := client.Docs().GetPage(ctx, cmd.DocID, cmd.PageID)
	if err != nil {
		return err
	}

	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(os.Stdout, result)
	}
	if outfmt.IsPlain(ctx) {
		headers := []string{"ID", "NAME", "CONTENT"}
		content := result.Content
		if len(content) > 100 {
			content = content[:97] + "..."
		}
		rows := [][]string{{result.ID, result.Name, content}}
		return outfmt.WritePlain(os.Stdout, headers, rows)
	}

	fmt.Printf("ID: %s\n", result.ID)
	fmt.Printf("Name: %s\n", result.Name)
	fmt.Printf("Order: %d\n", result.Order)
	if result.Content != "" {
		fmt.Printf("Content:\n%s\n", result.Content)
	}

	return nil
}

// DocsCreateCmd creates a new doc.
type DocsCreateCmd struct {
	Name       string `required:"" help:"Doc name"`
	ParentType string `help:"Parent type (space, folder, list)"`
	ParentID   string `help:"Parent ID"`
}

func (cmd *DocsCreateCmd) Run(ctx context.Context) error {
	client, err := getClickUpClient(ctx)
	if err != nil {
		return err
	}

	req := clickup.CreateDocRequest{
		Name:       cmd.Name,
		ParentType: cmd.ParentType,
		ParentID:   cmd.ParentID,
	}
	result, err := client.Docs().Create(ctx, req)
	if err != nil {
		return err
	}

	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(os.Stdout, result)
	}
	if outfmt.IsPlain(ctx) {
		headers := []string{"ID", "NAME"}
		rows := [][]string{{result.ID, result.Name}}
		return outfmt.WritePlain(os.Stdout, headers, rows)
	}

	fmt.Fprintf(os.Stderr, "Created doc\n\n")
	fmt.Printf("ID: %s\n", result.ID)
	fmt.Printf("Name: %s\n", result.Name)

	return nil
}

// DocsCreatePageCmd creates a page in a doc.
type DocsCreatePageCmd struct {
	DocID         string `arg:"" required:"" help:"Doc ID"`
	Name          string `required:"" help:"Page name"`
	Content       string `help:"Page content"`
	ContentFormat string `help:"Content format (md or html)"`
}

func (cmd *DocsCreatePageCmd) Run(ctx context.Context) error {
	client, err := getClickUpClient(ctx)
	if err != nil {
		return err
	}

	req := clickup.CreatePageRequest{
		Name:          cmd.Name,
		Content:       cmd.Content,
		ContentFormat: cmd.ContentFormat,
	}
	result, err := client.Docs().CreatePage(ctx, cmd.DocID, req)
	if err != nil {
		return err
	}

	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(os.Stdout, result)
	}
	if outfmt.IsPlain(ctx) {
		headers := []string{"ID", "NAME"}
		rows := [][]string{{result.ID, result.Name}}
		return outfmt.WritePlain(os.Stdout, headers, rows)
	}

	fmt.Fprintf(os.Stderr, "Created page\n\n")
	fmt.Printf("ID: %s\n", result.ID)
	fmt.Printf("Name: %s\n", result.Name)

	return nil
}

// DocsEditPageCmd edits a page in a doc.
type DocsEditPageCmd struct {
	DocID         string `arg:"" required:"" help:"Doc ID"`
	PageID        string `arg:"" required:"" help:"Page ID"`
	Name          string `help:"New page name"`
	Content       string `help:"New page content"`
	ContentFormat string `help:"Content format (md or html)"`
}

func (cmd *DocsEditPageCmd) Run(ctx context.Context) error {
	client, err := getClickUpClient(ctx)
	if err != nil {
		return err
	}

	req := clickup.EditPageRequest{
		Name:          cmd.Name,
		Content:       cmd.Content,
		ContentFormat: cmd.ContentFormat,
	}
	result, err := client.Docs().EditPage(ctx, cmd.DocID, cmd.PageID, req)
	if err != nil {
		return err
	}

	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(os.Stdout, result)
	}
	if outfmt.IsPlain(ctx) {
		headers := []string{"ID", "NAME"}
		rows := [][]string{{result.ID, result.Name}}
		return outfmt.WritePlain(os.Stdout, headers, rows)
	}

	fmt.Fprintf(os.Stderr, "Updated page\n\n")
	fmt.Printf("ID: %s\n", result.ID)
	fmt.Printf("Name: %s\n", result.Name)

	return nil
}

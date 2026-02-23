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
	Get         DocsGetCmd         `cmd:"" help:"Get a doc"`
	PageListing DocsPageListingCmd `cmd:"" help:"Get doc page listing"`
	Pages       DocsPagesCmd       `cmd:"" help:"Get all pages in a doc"`
	Page        DocsPageCmd        `cmd:"" help:"Get a single page"`
	Create      DocsCreateCmd      `cmd:"" help:"Create a doc"`
	CreatePage  DocsCreatePageCmd  `cmd:"" help:"Create a page in a doc"`
	EditPage    DocsEditPageCmd    `cmd:"" help:"Edit a page"`
}

type DocsSearchCmd struct {
	Query string `name:"query" short:"q" help:"Search query"`
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
		fmt.Printf("ID: %s\n", d.ID)
		fmt.Printf("  Name: %s\n", d.Name)
		fmt.Printf("  Created: %d\n\n", d.DateCreated)
	}

	return nil
}

type DocsGetCmd struct {
	DocID string `arg:"" help:"Doc ID" required:""`
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

	fmt.Printf("ID: %s\n", result.ID)
	fmt.Printf("Name: %s\n", result.Name)
	fmt.Printf("Created: %d\n", result.DateCreated)

	return nil
}

type DocsPageListingCmd struct {
	DocID string `arg:"" help:"Doc ID" required:""`
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
		fmt.Printf("ID: %s\n", p.ID)
		fmt.Printf("  Title: %s\n", p.Name)
		fmt.Printf("  Order: %d\n\n", p.Order)
	}

	return nil
}

type DocsPagesCmd struct {
	DocID string `arg:"" help:"Doc ID" required:""`
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
		fmt.Printf("ID: %s\n", p.ID)
		fmt.Printf("  Title: %s\n", p.Name)
		fmt.Printf("  Order: %d\n\n", p.Order)

		if p.Content != "" {
			fmt.Printf("  Content:\n%s\n\n", p.Content)
		}
	}

	return nil
}

type DocsPageCmd struct {
	DocID  string `arg:"" help:"Doc ID" required:""`
	PageID string `arg:"" help:"Page ID" required:""`
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

	fmt.Printf("ID: %s\n", result.ID)
	fmt.Printf("Name: %s\n", result.Name)
	fmt.Printf("Order: %d\n", result.Order)

	if result.Content != "" {
		fmt.Printf("\nContent:\n%s\n", result.Content)
	}

	return nil
}

type DocsCreateCmd struct {
	Name       string `name:"name" short:"n" help:"Doc name" required:""`
	ParentType string `name:"type" short:"t" help:"Parent type (space, folder, list)"`
	ParentID   string `name:"id" short:"i" help:"Parent ID"`
}

func (cmd *DocsCreateCmd) Run(ctx context.Context) error {
	client, err := getClickUpClient(ctx)
	if err != nil {
		return err
	}

	result, err := client.Docs().Create(ctx, clickup.CreateDocRequest{
		Name:       cmd.Name,
		ParentType: cmd.ParentType,
		ParentID:   cmd.ParentID,
	})
	if err != nil {
		return err
	}

	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(os.Stdout, result)
	}

	fmt.Fprintf(os.Stderr, "Created doc: %s\n", result.Name)
	fmt.Printf("ID: %s\n", result.ID)

	return nil
}

type DocsCreatePageCmd struct {
	DocID         string `arg:"" help:"Doc ID" required:""`
	Name          string `name:"name" short:"n" help:"Page name" required:""`
	Content       string `name:"content" short:"c" help:"Page content (markdown)"`
	ContentFormat string `name:"format" short:"f" help:"Content format (md or html)" default:"md"`
}

func (cmd *DocsCreatePageCmd) Run(ctx context.Context) error {
	client, err := getClickUpClient(ctx)
	if err != nil {
		return err
	}

	result, err := client.Docs().CreatePage(ctx, cmd.DocID, clickup.CreatePageRequest{
		Name:          cmd.Name,
		Content:       cmd.Content,
		ContentFormat: cmd.ContentFormat,
	})
	if err != nil {
		return err
	}

	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(os.Stdout, result)
	}

	fmt.Fprintf(os.Stderr, "Created page: %s\n", result.Name)
	fmt.Printf("ID: %s\n", result.ID)

	return nil
}

type DocsEditPageCmd struct {
	DocID         string `arg:"" help:"Doc ID" required:""`
	PageID        string `arg:"" help:"Page ID" required:""`
	Name          string `name:"name" short:"n" help:"New page name"`
	Content       string `name:"content" short:"c" help:"New page content (replaces entire content)"`
	ContentFormat string `name:"format" short:"f" help:"Content format (md or html)"`
}

func (cmd *DocsEditPageCmd) Run(ctx context.Context) error {
	client, err := getClickUpClient(ctx)
	if err != nil {
		return err
	}

	result, err := client.Docs().EditPage(ctx, cmd.DocID, cmd.PageID, clickup.EditPageRequest{
		Name:          cmd.Name,
		Content:       cmd.Content,
		ContentFormat: cmd.ContentFormat,
	})
	if err != nil {
		return err
	}

	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(os.Stdout, result)
	}

	fmt.Fprintf(os.Stderr, "Updated page: %s\n", result.Name)
	fmt.Printf("ID: %s\n", result.ID)

	return nil
}

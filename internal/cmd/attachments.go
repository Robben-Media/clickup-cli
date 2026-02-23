package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/builtbyrobben/clickup-cli/internal/outfmt"
)

type AttachmentsCmd struct {
	Upload AttachmentsUploadCmd `cmd:"" help:"Upload a file to a task (v2 API)"`
	List   AttachmentsListCmd   `cmd:"" help:"List attachments for a parent entity (v3 API)"`
	Create AttachmentsCreateCmd `cmd:"" help:"Upload a file to a parent entity (v3 API)"`
}

type AttachmentsUploadCmd struct {
	Task string `name:"task" help:"Task ID" required:""`
	File string `arg:"" help:"Path to the file to upload" required:""`
}

func (cmd *AttachmentsUploadCmd) Run(ctx context.Context) error {
	client, err := getClickUpClient(ctx)
	if err != nil {
		return err
	}

	result, err := client.Attachments().Upload(ctx, cmd.Task, cmd.File)
	if err != nil {
		return err
	}

	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(os.Stdout, result)
	}
	if outfmt.IsPlain(ctx) {
		headers := []string{"ID", "TITLE", "SIZE", "URL"}
		rows := [][]string{{result.ID, result.Title, fmt.Sprintf("%d", result.Size), result.URL}}
		return outfmt.WritePlain(os.Stdout, headers, rows)
	}

	fmt.Fprintf(os.Stderr, "Uploaded attachment: %s\n", result.Title)
	fmt.Printf("ID: %s\n", result.ID)
	fmt.Printf("URL: %s\n", result.URL)
	fmt.Printf("Size: %d bytes\n", result.Size)

	if result.Extension != "" {
		fmt.Printf("Extension: %s\n", result.Extension)
	}

	return nil
}

type AttachmentsListCmd struct {
	ParentType string `name:"type" short:"t" help:"Parent type (task, list, folder, space)" required:""`
	ParentID   string `name:"id" short:"i" help:"Parent ID" required:""`
}

func (cmd *AttachmentsListCmd) Run(ctx context.Context) error {
	client, err := getClickUpClient(ctx)
	if err != nil {
		return err
	}

	result, err := client.Attachments().List(ctx, cmd.ParentType, cmd.ParentID)
	if err != nil {
		return err
	}

	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(os.Stdout, result)
	}
	if outfmt.IsPlain(ctx) {
		headers := []string{"ID", "TITLE", "SIZE", "URL"}
		rows := make([][]string, 0, len(result.Attachments))
		for _, att := range result.Attachments {
			rows = append(rows, []string{att.ID, att.Title, fmt.Sprintf("%d", att.Size), att.URL})
		}
		return outfmt.WritePlain(os.Stdout, headers, rows)
	}

	if len(result.Attachments) == 0 {
		fmt.Fprintln(os.Stderr, "No attachments found")
		return nil
	}

	fmt.Fprintf(os.Stderr, "Found %d attachments\n\n", len(result.Attachments))

	for _, att := range result.Attachments {
		fmt.Printf("ID: %s\n", att.ID)
		fmt.Printf("  Title: %s\n", att.Title)
		fmt.Printf("  URL: %s\n", att.URL)
		fmt.Printf("  Size: %d bytes\n", att.Size)

		if att.Extension != "" {
			fmt.Printf("  Extension: %s\n", att.Extension)
		}

		fmt.Println()
	}

	return nil
}

type AttachmentsCreateCmd struct {
	ParentType string `name:"type" short:"t" help:"Parent type (task, list, folder, space)" required:""`
	ParentID   string `name:"id" short:"i" help:"Parent ID" required:""`
	File       string `arg:"" help:"Path to the file to upload" required:""`
}

func (cmd *AttachmentsCreateCmd) Run(ctx context.Context) error {
	client, err := getClickUpClient(ctx)
	if err != nil {
		return err
	}

	result, err := client.Attachments().Create(ctx, cmd.ParentType, cmd.ParentID, cmd.File)
	if err != nil {
		return err
	}

	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(os.Stdout, result)
	}
	if outfmt.IsPlain(ctx) {
		headers := []string{"ID", "TITLE", "SIZE", "URL"}
		rows := [][]string{{result.ID, result.Title, fmt.Sprintf("%d", result.Size), result.URL}}
		return outfmt.WritePlain(os.Stdout, headers, rows)
	}

	fmt.Fprintf(os.Stderr, "Created attachment: %s\n", result.Title)
	fmt.Printf("ID: %s\n", result.ID)
	fmt.Printf("URL: %s\n", result.URL)
	fmt.Printf("Size: %d bytes\n", result.Size)

	if result.Extension != "" {
		fmt.Printf("Extension: %s\n", result.Extension)
	}

	return nil
}

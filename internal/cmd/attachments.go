package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/builtbyrobben/clickup-cli/internal/outfmt"
)

type AttachmentsCmd struct {
	Upload AttachmentsUploadCmd `cmd:"" help:"Upload a file to a task (v2)"`
	List   AttachmentsListCmd   `cmd:"" help:"List attachments for an entity (v3)"`
	Create AttachmentsCreateCmd `cmd:"" help:"Upload a file to an entity (v3)"`
}

type AttachmentsUploadCmd struct {
	TaskID string `arg:"" required:"" help:"Task ID to attach file to"`
	File   string `arg:"" required:"" help:"Path to file to upload"`
}

func (cmd *AttachmentsUploadCmd) Run(ctx context.Context) error {
	client, err := getClickUpClient(ctx)
	if err != nil {
		return err
	}

	file, err := os.Open(cmd.File)
	if err != nil {
		return fmt.Errorf("open file: %w", err)
	}
	defer file.Close()

	result, err := client.Attachments().Upload(ctx, cmd.TaskID, file, cmd.File)
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

	fmt.Fprintf(os.Stderr, "File uploaded\n\n")
	fmt.Printf("ID: %s\n", result.ID)
	fmt.Printf("  Title: %s\n", result.Title)
	fmt.Printf("  Size: %d bytes\n", result.Size)
	fmt.Printf("  URL: %s\n", result.URL)

	return nil
}

type AttachmentsListCmd struct {
	ParentType string `required:"" help:"Parent type: task, list, folder, or space"`
	ParentID   string `required:"" help:"Parent ID"`
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
		fmt.Printf("  Size: %d bytes\n", att.Size)
		fmt.Printf("  URL: %s\n", att.URL)
		fmt.Println()
	}

	return nil
}

type AttachmentsCreateCmd struct {
	ParentType string `required:"" help:"Parent type: task, list, folder, or space"`
	ParentID   string `required:"" help:"Parent ID"`
	File       string `arg:"" required:"" help:"Path to file to upload"`
}

func (cmd *AttachmentsCreateCmd) Run(ctx context.Context) error {
	client, err := getClickUpClient(ctx)
	if err != nil {
		return err
	}

	file, err := os.Open(cmd.File)
	if err != nil {
		return fmt.Errorf("open file: %w", err)
	}
	defer file.Close()

	result, err := client.Attachments().Create(ctx, cmd.ParentType, cmd.ParentID, file, cmd.File)
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

	fmt.Fprintf(os.Stderr, "File uploaded\n\n")
	fmt.Printf("ID: %s\n", result.ID)
	fmt.Printf("  Title: %s\n", result.Title)
	fmt.Printf("  Size: %d bytes\n", result.Size)
	fmt.Printf("  URL: %s\n", result.URL)

	return nil
}

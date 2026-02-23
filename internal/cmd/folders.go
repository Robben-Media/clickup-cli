package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/builtbyrobben/clickup-cli/internal/clickup"
	"github.com/builtbyrobben/clickup-cli/internal/outfmt"
)

type FoldersCmd struct {
	Get          FoldersGetCmd          `cmd:"" help:"Get folder details"`
	Create       FoldersCreateCmd       `cmd:"" help:"Create a new folder"`
	Update       FoldersUpdateCmd       `cmd:"" help:"Update a folder"`
	Delete       FoldersDeleteCmd       `cmd:"" help:"Delete a folder"`
	FromTemplate FoldersFromTemplateCmd `cmd:"" help:"Create folder from template"`
}

type FoldersGetCmd struct {
	FolderID string `arg:"" required:"" help:"Folder ID"`
}

func (cmd *FoldersGetCmd) Run(ctx context.Context) error {
	client, err := getClickUpClient(ctx)
	if err != nil {
		return err
	}

	result, err := client.Folders().Get(ctx, cmd.FolderID)
	if err != nil {
		return err
	}

	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(os.Stdout, result)
	}
	if outfmt.IsPlain(ctx) {
		headers := []string{"ID", "NAME", "TASK_COUNT", "LIST_COUNT"}
		listCount := fmt.Sprintf("%d", len(result.Lists))
		rows := [][]string{{result.ID, result.Name, result.TaskCount, listCount}}
		return outfmt.WritePlain(os.Stdout, headers, rows)
	}

	fmt.Fprintf(os.Stderr, "Folder Details\n\n")
	fmt.Printf("ID: %s\n", result.ID)
	fmt.Printf("Name: %s\n", result.Name)
	fmt.Printf("Task Count: %s\n", result.TaskCount)

	if len(result.Lists) > 0 {
		fmt.Println("Lists:")

		for _, list := range result.Lists {
			fmt.Printf("  %s: %s\n", list.ID, list.Name)
		}
	}

	return nil
}

type FoldersCreateCmd struct {
	SpaceID string `arg:"" required:"" help:"Space ID"`
	Name    string `arg:"" required:"" help:"Folder name"`
}

func (cmd *FoldersCreateCmd) Run(ctx context.Context) error {
	client, err := getClickUpClient(ctx)
	if err != nil {
		return err
	}

	req := clickup.CreateFolderRequest{Name: cmd.Name}

	result, err := client.Folders().Create(ctx, cmd.SpaceID, req)
	if err != nil {
		return err
	}

	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(os.Stdout, result)
	}
	if outfmt.IsPlain(ctx) {
		headers := []string{"ID", "NAME", "SPACE_ID"}
		rows := [][]string{{result.ID, result.Name, result.Space.ID}}
		return outfmt.WritePlain(os.Stdout, headers, rows)
	}

	fmt.Fprintf(os.Stderr, "Folder created\n\n")
	fmt.Printf("ID: %s\n", result.ID)
	fmt.Printf("Name: %s\n", result.Name)
	fmt.Printf("Space: %s\n", result.Space.ID)

	return nil
}

type FoldersUpdateCmd struct {
	FolderID string `arg:"" required:"" help:"Folder ID"`
	Name     string `help:"New folder name"`
}

func (cmd *FoldersUpdateCmd) Run(ctx context.Context) error {
	client, err := getClickUpClient(ctx)
	if err != nil {
		return err
	}

	req := clickup.UpdateFolderRequest{Name: cmd.Name}

	result, err := client.Folders().Update(ctx, cmd.FolderID, req)
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

	fmt.Fprintf(os.Stderr, "Folder updated\n\n")
	fmt.Printf("ID: %s\n", result.ID)
	fmt.Printf("Name: %s\n", result.Name)

	return nil
}

type FoldersDeleteCmd struct {
	FolderID string `arg:"" required:"" help:"Folder ID"`
	Force    bool   `help:"Skip confirmation"`
}

func (cmd *FoldersDeleteCmd) Run(ctx context.Context) error {
	client, err := getClickUpClient(ctx)
	if err != nil {
		return err
	}

	if !cmd.Force && !outfmt.IsPlain(ctx) && !outfmt.IsJSON(ctx) {
		fmt.Fprintf(os.Stderr, "WARNING: Deleting a folder will move its lists to folderless.\n")
		fmt.Fprintf(os.Stderr, "Folder ID: %s\n\n", cmd.FolderID)
		fmt.Fprintf(os.Stderr, "Use --force to skip this confirmation.\n")

		return fmt.Errorf("operation cancelled: use --force to confirm destructive operation")
	}

	if err := client.Folders().Delete(ctx, cmd.FolderID); err != nil {
		return err
	}

	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(os.Stdout, map[string]string{
			"status":    "success",
			"message":   "Folder deleted",
			"folder_id": cmd.FolderID,
		})
	}
	if outfmt.IsPlain(ctx) {
		headers := []string{"STATUS", "FOLDER_ID"}
		rows := [][]string{{"success", cmd.FolderID}}
		return outfmt.WritePlain(os.Stdout, headers, rows)
	}

	fmt.Fprintf(os.Stderr, "Folder %s deleted\n", cmd.FolderID)

	return nil
}

type FoldersFromTemplateCmd struct {
	SpaceID    string `arg:"" required:"" help:"Space ID"`
	TemplateID string `arg:"" required:"" help:"Template ID"`
	Name       string `help:"New folder name (optional, uses template name if not set)"`
}

func (cmd *FoldersFromTemplateCmd) Run(ctx context.Context) error {
	client, err := getClickUpClient(ctx)
	if err != nil {
		return err
	}

	req := clickup.CreateFolderFromTemplateRequest{Name: cmd.Name}

	result, err := client.Folders().CreateFromTemplate(ctx, cmd.SpaceID, cmd.TemplateID, req)
	if err != nil {
		return err
	}

	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(os.Stdout, result)
	}
	if outfmt.IsPlain(ctx) {
		headers := []string{"ID", "NAME", "SPACE_ID"}
		rows := [][]string{{result.ID, result.Name, result.Space.ID}}
		return outfmt.WritePlain(os.Stdout, headers, rows)
	}

	fmt.Fprintf(os.Stderr, "Folder created from template\n\n")
	fmt.Printf("ID: %s\n", result.ID)
	fmt.Printf("Name: %s\n", result.Name)
	fmt.Printf("Space: %s\n", result.Space.ID)

	return nil
}

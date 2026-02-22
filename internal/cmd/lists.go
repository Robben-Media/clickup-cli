package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/builtbyrobben/clickup-cli/internal/clickup"
	"github.com/builtbyrobben/clickup-cli/internal/outfmt"
)

type ListsCmd struct {
	List ListsListCmd `cmd:"" help:"List all lists in a space or folder"`
}

type ListsListCmd struct {
	Space  string `help:"Space ID to list from (shows folders + folderless lists)"`
	Folder string `help:"Folder ID to list from"`
}

func (cmd *ListsListCmd) Run(ctx context.Context) error {
	client, err := getClickUpClient(ctx)
	if err != nil {
		return err
	}

	if cmd.Folder == "" && cmd.Space == "" {
		return fmt.Errorf("either --space or --folder is required")
	}

	// If folder is specified, list from folder
	if cmd.Folder != "" {
		return cmd.listByFolder(ctx, client)
	}

	// Otherwise list from space (folders + folderless lists)
	return cmd.listBySpace(ctx, client)
}

func (cmd *ListsListCmd) listByFolder(ctx context.Context, client *clickup.Client) error {
	result, err := client.Lists().ListByFolder(ctx, cmd.Folder)
	if err != nil {
		return err
	}

	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(os.Stdout, result)
	}
	if outfmt.IsPlain(ctx) {
		headers := []string{"ID", "NAME"}
		var rows [][]string
		for _, list := range result.Lists {
			rows = append(rows, []string{list.ID, list.Name})
		}
		return outfmt.WritePlain(os.Stdout, headers, rows)
	}

	if len(result.Lists) == 0 {
		fmt.Fprintln(os.Stderr, "No lists found")
		return nil
	}

	fmt.Fprintf(os.Stderr, "Found %d lists\n\n", len(result.Lists))

	for _, list := range result.Lists {
		fmt.Printf("ID: %s\n", list.ID)
		fmt.Printf("  Name: %s\n", list.Name)
		fmt.Println()
	}

	return nil
}

func (cmd *ListsListCmd) listBySpace(ctx context.Context, client *clickup.Client) error {
	// Get folders and their lists
	folders, err := client.Lists().ListFolders(ctx, cmd.Space)
	if err != nil {
		return err
	}

	// Get folderless lists
	folderless, err := client.Lists().ListFolderless(ctx, cmd.Space)
	if err != nil {
		return err
	}

	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(os.Stdout, map[string]any{
			"folders":          folders.Folders,
			"folderless_lists": folderless.Lists,
		})
	}
	if outfmt.IsPlain(ctx) {
		headers := []string{"ID", "NAME", "FOLDER"}
		var rows [][]string
		for _, folder := range folders.Folders {
			for _, list := range folder.Lists {
				rows = append(rows, []string{list.ID, list.Name, folder.Name})
			}
		}
		for _, list := range folderless.Lists {
			rows = append(rows, []string{list.ID, list.Name, ""})
		}
		return outfmt.WritePlain(os.Stdout, headers, rows)
	}

	if len(folders.Folders) == 0 && len(folderless.Lists) == 0 {
		fmt.Fprintln(os.Stderr, "No lists found")
		return nil
	}

	// Print folders with their lists
	for _, folder := range folders.Folders {
		fmt.Printf("Folder: %s (ID: %s)\n", folder.Name, folder.ID)

		for _, list := range folder.Lists {
			fmt.Printf("  List: %s (ID: %s)\n", list.Name, list.ID)
		}

		fmt.Println()
	}

	// Print folderless lists
	if len(folderless.Lists) > 0 {
		fmt.Println("Folderless Lists:")

		for _, list := range folderless.Lists {
			fmt.Printf("  ID: %s\n", list.ID)
			fmt.Printf("    Name: %s\n", list.Name)
		}

		fmt.Println()
	}

	return nil
}

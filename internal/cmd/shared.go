package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/builtbyrobben/clickup-cli/internal/clickup"
	"github.com/builtbyrobben/clickup-cli/internal/outfmt"
)

type SharedCmd struct {
	List SharedListCmd `cmd:"" help:"List shared hierarchy resources"`
}

type SharedListCmd struct {
	TeamID string `arg:"" required:"" help:"Team (workspace) ID"`
}

func (cmd *SharedListCmd) Run(ctx context.Context) error {
	client, err := getClickUpClient(ctx)
	if err != nil {
		return err
	}

	result, err := client.SharedHierarchy().List(ctx, cmd.TeamID)
	if err != nil {
		return err
	}

	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(os.Stdout, result)
	}
	if outfmt.IsPlain(ctx) {
		headers := []string{"TYPE", "ID", "NAME"}
		var rows [][]string
		for _, task := range result.Shared.Tasks {
			rows = append(rows, []string{"task", task.ID, task.Name})
		}
		for _, list := range result.Shared.Lists {
			rows = append(rows, []string{"list", list.ID, list.Name})
		}
		for _, folder := range result.Shared.Folders {
			rows = append(rows, []string{"folder", folder.ID, folder.Name})
		}
		return outfmt.WritePlain(os.Stdout, headers, rows)
	}

	// Human-readable output
	printSharedHierarchy(result)

	return nil
}

func printSharedHierarchy(result *clickup.SharedHierarchyResponse) {
	fmt.Fprintln(os.Stderr, "Shared Hierarchy")
	fmt.Fprintln(os.Stderr)

	if len(result.Shared.Tasks) > 0 {
		fmt.Fprintln(os.Stderr, "Tasks:")
		for _, task := range result.Shared.Tasks {
			fmt.Fprintf(os.Stderr, "  %s: %s\n", task.ID, task.Name)
		}
		fmt.Fprintln(os.Stderr)
	}

	if len(result.Shared.Lists) > 0 {
		fmt.Fprintln(os.Stderr, "Lists:")
		for _, list := range result.Shared.Lists {
			fmt.Fprintf(os.Stderr, "  %s: %s\n", list.ID, list.Name)
		}
		fmt.Fprintln(os.Stderr)
	}

	if len(result.Shared.Folders) > 0 {
		fmt.Fprintln(os.Stderr, "Folders:")
		for _, folder := range result.Shared.Folders {
			fmt.Fprintf(os.Stderr, "  %s: %s\n", folder.ID, folder.Name)
		}
		fmt.Fprintln(os.Stderr)
	}

	if len(result.Shared.Tasks) == 0 && len(result.Shared.Lists) == 0 && len(result.Shared.Folders) == 0 {
		fmt.Fprintln(os.Stderr, "No shared resources found")
	}
}

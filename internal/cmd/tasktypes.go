package cmd

import (
	"context"
	"fmt"
	"os"
	"strconv"

	"github.com/builtbyrobben/clickup-cli/internal/clickup"
	"github.com/builtbyrobben/clickup-cli/internal/outfmt"
)

type TaskTypesCmd struct {
	List TaskTypesListCmd `cmd:"" help:"List custom task types"`
}

type TaskTypesListCmd struct {
	TeamID string `arg:"" required:"" help:"Team (workspace) ID"`
}

func (cmd *TaskTypesListCmd) Run(ctx context.Context) error {
	client, err := getClickUpClient(ctx)
	if err != nil {
		return err
	}

	result, err := client.CustomTaskTypes().List(ctx, cmd.TeamID)
	if err != nil {
		return err
	}

	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(os.Stdout, result)
	}

	if outfmt.IsPlain(ctx) {
		headers := []string{"ID", "NAME", "NAME_PLURAL", "DESCRIPTION"}
		var rows [][]string
		for _, item := range result.CustomItems {
			rows = append(rows, []string{
				strconv.Itoa(item.ID),
				item.Name,
				item.NamePlural,
				item.Description,
			})
		}

		return outfmt.WritePlain(os.Stdout, headers, rows)
	}

	if len(result.CustomItems) == 0 {
		fmt.Fprintln(os.Stderr, "No custom task types found")
		return nil
	}

	fmt.Fprintf(os.Stderr, "Found %d custom task types\n\n", len(result.CustomItems))

	for _, item := range result.CustomItems {
		printCustomTaskType(&item)
	}

	return nil
}

func printCustomTaskType(item *clickup.CustomTaskType) {
	if item.Description != "" {
		fmt.Printf("%d: %s (%s) — %s\n", item.ID, item.Name, item.NamePlural, item.Description)
	} else {
		fmt.Printf("%d: %s (%s)\n", item.ID, item.Name, item.NamePlural)
	}
}

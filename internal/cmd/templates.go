package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/builtbyrobben/clickup-cli/internal/clickup"
	"github.com/builtbyrobben/clickup-cli/internal/outfmt"
)

type TemplatesCmd struct {
	List TemplatesListCmd `cmd:"" help:"List task templates"`
}

type TemplatesListCmd struct {
	TeamID string `arg:"" required:"" help:"Team (workspace) ID"`
	Page   int    `help:"Page number (0-indexed)" default:"0"`
}

func (cmd *TemplatesListCmd) Run(ctx context.Context) error {
	client, err := getClickUpClient(ctx)
	if err != nil {
		return err
	}

	result, err := client.Templates().List(ctx, cmd.TeamID, cmd.Page)
	if err != nil {
		return err
	}

	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(os.Stdout, result)
	}
	if outfmt.IsPlain(ctx) {
		headers := []string{"ID", "NAME"}
		var rows [][]string
		for _, template := range result.Templates {
			rows = append(rows, []string{template.ID, template.Name})
		}
		return outfmt.WritePlain(os.Stdout, headers, rows)
	}

	// Human-readable output
	printTemplates(result)

	return nil
}

func printTemplates(result *clickup.TaskTemplatesResponse) {
	fmt.Fprintln(os.Stderr, "Task Templates")
	fmt.Fprintln(os.Stderr)

	if len(result.Templates) == 0 {
		fmt.Fprintln(os.Stderr, "No templates found")
		return
	}

	for _, template := range result.Templates {
		fmt.Fprintf(os.Stderr, "  %s: %s\n", template.ID, template.Name)
	}
}

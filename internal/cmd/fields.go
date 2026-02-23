package cmd

import (
	"context"
	"fmt"
	"os"
	"strconv"

	"github.com/builtbyrobben/clickup-cli/internal/clickup"
	"github.com/builtbyrobben/clickup-cli/internal/outfmt"
)

type FieldsCmd struct {
	List   FieldsListCmd   `cmd:"" help:"List custom fields"`
	Set    FieldsSetCmd    `cmd:"" help:"Set a custom field value"`
	Remove FieldsRemoveCmd `cmd:"" help:"Remove a custom field value"`
}

type FieldsListCmd struct {
	ListID   string `help:"List ID"`
	FolderID string `help:"Folder ID"`
	SpaceID  string `help:"Space ID"`
	TeamID   string `help:"Team/Workspace ID"`
}

func (cmd *FieldsListCmd) Run(ctx context.Context) error {
	client, err := getClickUpClient(ctx)
	if err != nil {
		return err
	}

	// Exactly one scope is required
	count := 0
	if cmd.ListID != "" {
		count++
	}

	if cmd.FolderID != "" {
		count++
	}

	if cmd.SpaceID != "" {
		count++
	}

	if cmd.TeamID != "" {
		count++
	}

	if count != 1 {
		return fmt.Errorf("exactly one of --list, --folder, --space, or --team is required")
	}

	var result *clickup.CustomFieldsResponse

	switch {
	case cmd.ListID != "":
		result, err = client.CustomFields().ListByList(ctx, cmd.ListID)
	case cmd.FolderID != "":
		result, err = client.CustomFields().ListByFolder(ctx, cmd.FolderID)
	case cmd.SpaceID != "":
		result, err = client.CustomFields().ListBySpace(ctx, cmd.SpaceID)
	case cmd.TeamID != "":
		result, err = client.CustomFields().ListByTeam(ctx, cmd.TeamID)
	}

	if err != nil {
		return err
	}

	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(os.Stdout, result)
	}

	if outfmt.IsPlain(ctx) {
		headers := []string{"ID", "NAME", "TYPE", "REQUIRED"}
		var rows [][]string
		for _, field := range result.Fields {
			rows = append(rows, []string{
				field.ID,
				field.Name,
				field.Type,
				strconv.FormatBool(field.Required),
			})
		}

		return outfmt.WritePlain(os.Stdout, headers, rows)
	}

	if len(result.Fields) == 0 {
		fmt.Fprintln(os.Stderr, "No custom fields found")
		return nil
	}

	fmt.Fprintln(os.Stderr, "Custom Fields")

	for _, field := range result.Fields {
		req := "optional"
		if field.Required {
			req = "required"
		}

		fmt.Fprintf(os.Stderr, "  %s: %s (%s, %s)\n", field.ID, field.Name, field.Type, req)
	}

	return nil
}

type FieldsSetCmd struct {
	TaskID  string `arg:"" required:"" help:"Task ID"`
	FieldID string `required:"" help:"Custom field ID"`
	Value   string `arg:"" required:"" help:"Field value (format depends on field type)"`
}

func (cmd *FieldsSetCmd) Run(ctx context.Context) error {
	client, err := getClickUpClient(ctx)
	if err != nil {
		return err
	}

	// Pass the value as-is (string). For complex types (array, object),
	// the user would need to pass JSON, which we don't parse here.
	// This simple approach covers most common field types.
	if err := client.CustomFields().Set(ctx, cmd.TaskID, cmd.FieldID, cmd.Value); err != nil {
		return err
	}

	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(os.Stdout, map[string]string{
			"status":   "success",
			"task_id":  cmd.TaskID,
			"field_id": cmd.FieldID,
		})
	}

	if outfmt.IsPlain(ctx) {
		headers := []string{"STATUS", "TASK_ID", "FIELD_ID"}
		rows := [][]string{{"success", cmd.TaskID, cmd.FieldID}}

		return outfmt.WritePlain(os.Stdout, headers, rows)
	}

	fmt.Fprintf(os.Stderr, "Custom field %s set on task %s\n", cmd.FieldID, cmd.TaskID)

	return nil
}

type FieldsRemoveCmd struct {
	TaskID  string `arg:"" required:"" help:"Task ID"`
	FieldID string `required:"" help:"Custom field ID"`
}

func (cmd *FieldsRemoveCmd) Run(ctx context.Context) error {
	client, err := getClickUpClient(ctx)
	if err != nil {
		return err
	}

	if err := client.CustomFields().Remove(ctx, cmd.TaskID, cmd.FieldID); err != nil {
		return err
	}

	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(os.Stdout, map[string]string{
			"status":   "success",
			"message":  "Custom field value removed",
			"task_id":  cmd.TaskID,
			"field_id": cmd.FieldID,
		})
	}

	if outfmt.IsPlain(ctx) {
		headers := []string{"STATUS", "TASK_ID", "FIELD_ID"}
		rows := [][]string{{"success", cmd.TaskID, cmd.FieldID}}

		return outfmt.WritePlain(os.Stdout, headers, rows)
	}

	fmt.Fprintf(os.Stderr, "Custom field %s removed from task %s\n", cmd.FieldID, cmd.TaskID)

	return nil
}

package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"github.com/builtbyrobben/clickup-cli/internal/clickup"
	"github.com/builtbyrobben/clickup-cli/internal/outfmt"
)

type FieldsCmd struct {
	List   FieldsListCmd   `cmd:"" help:"List custom fields"`
	Set    FieldsSetCmd    `cmd:"" help:"Set a custom field value on a task"`
	Remove FieldsRemoveCmd `cmd:"" help:"Remove a custom field value from a task"`
}

type FieldsListCmd struct {
	ListID   string `help:"List ID to list fields from"`
	FolderID string `help:"Folder ID to list fields from"`
	SpaceID  string `help:"Space ID to list fields from"`
	TeamID   string `help:"Team/Workspace ID to list fields from"`
}

func (cmd *FieldsListCmd) Run(ctx context.Context) error {
	client, err := getClickUpClient(ctx)
	if err != nil {
		return err
	}

	// Count how many scope flags are set
	scopes := 0
	if cmd.ListID != "" {
		scopes++
	}

	if cmd.FolderID != "" {
		scopes++
	}

	if cmd.SpaceID != "" {
		scopes++
	}

	if cmd.TeamID != "" {
		scopes++
	}

	if scopes == 0 {
		return fmt.Errorf("exactly one scope flag required: --list, --folder, --space, or --team")
	}

	if scopes > 1 {
		return fmt.Errorf("only one scope flag allowed: --list, --folder, --space, or --team")
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
	fmt.Fprintln(os.Stderr, "")

	for _, field := range result.Fields {
		required := "optional"
		if field.Required {
			required = "required"
		}

		fmt.Fprintf(os.Stderr, "  %s: %s (%s, %s)\n", field.ID, field.Name, field.Type, required)
	}

	return nil
}

type FieldsSetCmd struct {
	TaskID  string `required:"" help:"Task ID to set the field on"`
	FieldID string `required:"" help:"Custom field ID"`
	Value   string `required:"" help:"Field value (JSON for complex types)"`
}

func (cmd *FieldsSetCmd) Run(ctx context.Context) error {
	client, err := getClickUpClient(ctx)
	if err != nil {
		return err
	}

	// Try to parse the value as JSON first, otherwise use as string
	var value interface{}
	if err := json.Unmarshal([]byte(cmd.Value), &value); err != nil {
		// Not valid JSON, use as string
		value = cmd.Value
	}

	if err := client.CustomFields().Set(ctx, cmd.TaskID, cmd.FieldID, value); err != nil {
		return err
	}

	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(os.Stdout, map[string]string{
			"status":   "success",
			"message":  "Custom field value set",
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
	TaskID  string `required:"" help:"Task ID to remove the field from"`
	FieldID string `required:"" help:"Custom field ID to remove"`
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

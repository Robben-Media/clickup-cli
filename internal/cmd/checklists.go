package cmd

import (
	"context"
	"fmt"
	"os"
	"strconv"

	"github.com/builtbyrobben/clickup-cli/internal/clickup"
	"github.com/builtbyrobben/clickup-cli/internal/outfmt"
)

type ChecklistsCmd struct {
	Create     ChecklistsCreateCmd     `cmd:"" help:"Create a new checklist on a task"`
	Update     ChecklistsUpdateCmd     `cmd:"" help:"Update a checklist"`
	Delete     ChecklistsDeleteCmd     `cmd:"" help:"Delete a checklist"`
	AddItem    ChecklistsAddItemCmd    `cmd:"" name:"add-item" help:"Add an item to a checklist"`
	UpdateItem ChecklistsUpdateItemCmd `cmd:"" name:"update-item" help:"Update a checklist item"`
	DeleteItem ChecklistsDeleteItemCmd `cmd:"" name:"delete-item" help:"Delete a checklist item"`
}

type ChecklistsCreateCmd struct {
	TaskID string `required:"" help:"Task ID to create checklist on"`
	Name   string `arg:"" required:"" help:"Checklist name"`
}

func (cmd *ChecklistsCreateCmd) Run(ctx context.Context) error {
	client, err := getClickUpClient(ctx)
	if err != nil {
		return err
	}

	req := clickup.CreateChecklistRequest{
		Name: cmd.Name,
	}

	result, err := client.Checklists().Create(ctx, cmd.TaskID, req)
	if err != nil {
		return err
	}

	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(os.Stdout, map[string]any{
			"checklist": result,
		})
	}

	if outfmt.IsPlain(ctx) {
		headers := []string{"ID", "NAME", "ITEM_COUNT"}
		rows := [][]string{{result.ID, result.Name, strconv.Itoa(len(result.Items))}}
		return outfmt.WritePlain(os.Stdout, headers, rows)
	}

	fmt.Fprintf(os.Stderr, "Checklist %q created (ID: %s)\n", result.Name, result.ID)

	return nil
}

type ChecklistsUpdateCmd struct {
	ChecklistID string `arg:"" required:"" help:"Checklist ID to update"`
	Name        string `help:"New checklist name"`
	Position    int    `help:"New position (0-indexed)"`
}

func (cmd *ChecklistsUpdateCmd) Run(ctx context.Context) error {
	client, err := getClickUpClient(ctx)
	if err != nil {
		return err
	}

	req := clickup.EditChecklistRequest{
		Name:     cmd.Name,
		Position: cmd.Position,
	}

	result, err := client.Checklists().Update(ctx, cmd.ChecklistID, req)
	if err != nil {
		return err
	}

	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(os.Stdout, map[string]any{
			"checklist": result,
		})
	}

	if outfmt.IsPlain(ctx) {
		headers := []string{"ID", "NAME", "ITEM_COUNT"}
		rows := [][]string{{result.ID, result.Name, strconv.Itoa(len(result.Items))}}
		return outfmt.WritePlain(os.Stdout, headers, rows)
	}

	fmt.Fprintf(os.Stderr, "Checklist %q updated\n", result.Name)

	return nil
}

type ChecklistsDeleteCmd struct {
	ChecklistID string `arg:"" required:"" help:"Checklist ID to delete"`
}

func (cmd *ChecklistsDeleteCmd) Run(ctx context.Context) error {
	client, err := getClickUpClient(ctx)
	if err != nil {
		return err
	}

	if err := client.Checklists().Delete(ctx, cmd.ChecklistID); err != nil {
		return err
	}

	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(os.Stdout, map[string]string{
			"status":       "success",
			"message":      "Checklist deleted",
			"checklist_id": cmd.ChecklistID,
		})
	}

	if outfmt.IsPlain(ctx) {
		headers := []string{"STATUS", "CHECKLIST_ID"}
		rows := [][]string{{"success", cmd.ChecklistID}}
		return outfmt.WritePlain(os.Stdout, headers, rows)
	}

	fmt.Fprintf(os.Stderr, "Checklist %s deleted\n", cmd.ChecklistID)

	return nil
}

type ChecklistsAddItemCmd struct {
	ChecklistID string `arg:"" required:"" help:"Checklist ID to add item to"`
	Name        string `arg:"" required:"" help:"Item name"`
	Assignee    int    `help:"User ID to assign the item to"`
}

func (cmd *ChecklistsAddItemCmd) Run(ctx context.Context) error {
	client, err := getClickUpClient(ctx)
	if err != nil {
		return err
	}

	req := clickup.CreateChecklistItemRequest{
		Name:     cmd.Name,
		Assignee: cmd.Assignee,
	}

	result, err := client.Checklists().AddItem(ctx, cmd.ChecklistID, req)
	if err != nil {
		return err
	}

	// Find the newly added item (it should be the last one)
	var newItem *clickup.ChecklistItem
	for i := range result.Items {
		if result.Items[i].Name == cmd.Name {
			newItem = &result.Items[i]

			break
		}
	}

	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(os.Stdout, map[string]any{
			"checklist": result,
		})
	}

	if outfmt.IsPlain(ctx) {
		headers := []string{"CHECKLIST_ID", "ITEM_ID", "NAME", "RESOLVED"}
		if newItem != nil {
			rows := [][]string{{result.ID, newItem.ID, newItem.Name, strconv.FormatBool(newItem.Resolved)}}
			return outfmt.WritePlain(os.Stdout, headers, rows)
		}
		// Fallback if item not found
		rows := [][]string{{result.ID, "", cmd.Name, "false"}}
		return outfmt.WritePlain(os.Stdout, headers, rows)
	}

	fmt.Fprintf(os.Stderr, "Item %q added to checklist %s\n", cmd.Name, cmd.ChecklistID)

	return nil
}

type ChecklistsUpdateItemCmd struct {
	ChecklistID string `arg:"" required:"" help:"Checklist ID containing the item"`
	ItemID      string `arg:"" required:"" help:"Item ID to update"`
	Name        string `help:"New item name"`
	Resolved    bool   `help:"Mark item as resolved"`
	Unresolved  bool   `help:"Mark item as unresolved"`
	Assignee    int    `help:"User ID to assign the item to"`
	Parent      string `help:"Parent item ID for nesting"`
}

func (cmd *ChecklistsUpdateItemCmd) Run(ctx context.Context) error {
	client, err := getClickUpClient(ctx)
	if err != nil {
		return err
	}

	req := clickup.EditChecklistItemRequest{
		Name:     cmd.Name,
		Assignee: cmd.Assignee,
		Parent:   cmd.Parent,
	}

	// Handle resolved flag (can be true, false, or unset)
	if cmd.Resolved && cmd.Unresolved {
		return fmt.Errorf("cannot specify both --resolved and --unresolved")
	}

	if cmd.Resolved {
		resolved := true
		req.Resolved = &resolved
	} else if cmd.Unresolved {
		resolved := false
		req.Resolved = &resolved
	}

	result, err := client.Checklists().UpdateItem(ctx, cmd.ChecklistID, cmd.ItemID, req)
	if err != nil {
		return err
	}

	// Find the updated item
	var updatedItem *clickup.ChecklistItem
	for i := range result.Items {
		if result.Items[i].ID == cmd.ItemID {
			updatedItem = &result.Items[i]

			break
		}
	}

	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(os.Stdout, map[string]any{
			"checklist": result,
		})
	}

	if outfmt.IsPlain(ctx) {
		headers := []string{"CHECKLIST_ID", "ITEM_ID", "NAME", "RESOLVED"}
		if updatedItem != nil {
			rows := [][]string{{result.ID, updatedItem.ID, updatedItem.Name, strconv.FormatBool(updatedItem.Resolved)}}
			return outfmt.WritePlain(os.Stdout, headers, rows)
		}
		// Fallback
		rows := [][]string{{result.ID, cmd.ItemID, "", ""}}
		return outfmt.WritePlain(os.Stdout, headers, rows)
	}

	fmt.Fprintf(os.Stderr, "Item %s updated in checklist %s\n", cmd.ItemID, cmd.ChecklistID)

	return nil
}

type ChecklistsDeleteItemCmd struct {
	ChecklistID string `arg:"" required:"" help:"Checklist ID containing the item"`
	ItemID      string `arg:"" required:"" help:"Item ID to delete"`
}

func (cmd *ChecklistsDeleteItemCmd) Run(ctx context.Context) error {
	client, err := getClickUpClient(ctx)
	if err != nil {
		return err
	}

	if err := client.Checklists().DeleteItem(ctx, cmd.ChecklistID, cmd.ItemID); err != nil {
		return err
	}

	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(os.Stdout, map[string]string{
			"status":       "success",
			"message":      "Checklist item deleted",
			"checklist_id": cmd.ChecklistID,
			"item_id":      cmd.ItemID,
		})
	}

	if outfmt.IsPlain(ctx) {
		headers := []string{"STATUS", "CHECKLIST_ID", "ITEM_ID"}
		rows := [][]string{{"success", cmd.ChecklistID, cmd.ItemID}}
		return outfmt.WritePlain(os.Stdout, headers, rows)
	}

	fmt.Fprintf(os.Stderr, "Item %s deleted from checklist %s\n", cmd.ItemID, cmd.ChecklistID)

	return nil
}

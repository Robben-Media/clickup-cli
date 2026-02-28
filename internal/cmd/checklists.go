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
	Create     ChecklistsCreateCmd     `cmd:"" help:"Create a checklist on a task"`
	Update     ChecklistsUpdateCmd     `cmd:"" help:"Update a checklist"`
	Delete     ChecklistsDeleteCmd     `cmd:"" help:"Delete a checklist"`
	AddItem    ChecklistsAddItemCmd    `cmd:"" help:"Add an item to a checklist"`
	UpdateItem ChecklistsUpdateItemCmd `cmd:"" help:"Update a checklist item"`
	DeleteItem ChecklistsDeleteItemCmd `cmd:"" help:"Delete a checklist item"`
}

type ChecklistsCreateCmd struct {
	TaskID string `required:"" help:"Task ID"`
	Name   string `arg:"" required:"" help:"Checklist name"`
}

func (cmd *ChecklistsCreateCmd) Run(ctx context.Context) error {
	client, err := getClickUpClient(ctx)
	if err != nil {
		return err
	}

	req := clickup.CreateChecklistRequest{Name: cmd.Name}

	result, err := client.Checklists().Create(ctx, cmd.TaskID, req)
	if err != nil {
		return err
	}

	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(os.Stdout, result)
	}

	if outfmt.IsPlain(ctx) {
		headers := []string{"ID", "NAME", "ITEM_COUNT"}
		rows := [][]string{{result.ID, result.Name, strconv.Itoa(len(result.Items))}}

		return outfmt.WritePlain(os.Stdout, headers, rows)
	}

	fmt.Fprintf(os.Stderr, "Checklist created (ID: %s)\n", result.ID)
	fmt.Printf("  Name: %s\n", result.Name)

	return nil
}

type ChecklistsUpdateCmd struct {
	ChecklistID string `arg:"" required:"" help:"Checklist ID"`
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
		return outfmt.WriteJSON(os.Stdout, result)
	}

	if outfmt.IsPlain(ctx) {
		headers := []string{"ID", "NAME", "POSITION"}
		rows := [][]string{{result.ID, result.Name, strconv.Itoa(result.OrderIndex)}}

		return outfmt.WritePlain(os.Stdout, headers, rows)
	}

	fmt.Fprintf(os.Stderr, "Checklist updated (ID: %s)\n", result.ID)
	fmt.Printf("  Name: %s\n", result.Name)

	return nil
}

type ChecklistsDeleteCmd struct {
	ChecklistID string `arg:"" required:"" help:"Checklist ID"`
	Force       bool   `help:"Skip confirmation"`
}

func (cmd *ChecklistsDeleteCmd) Run(ctx context.Context) error {
	client, err := getClickUpClient(ctx)
	if err != nil {
		return err
	}

	if !cmd.Force {
		fmt.Fprintf(os.Stderr, "Warning: This will permanently delete checklist %s and all its items\n", cmd.ChecklistID)
		fmt.Fprint(os.Stderr, "Use --force to confirm deletion\n")

		return fmt.Errorf("operation cancelled: use --force to confirm")
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
	ChecklistID string `arg:"" required:"" help:"Checklist ID"`
	Name        string `arg:"" required:"" help:"Item name"`
	Assignee    int    `help:"Assignee user ID"`
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

	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(os.Stdout, result)
	}

	if outfmt.IsPlain(ctx) {
		headers := []string{"CHECKLIST_ID", "ITEM_ID", "NAME", "RESOLVED"}
		var rows [][]string
		for _, item := range result.Items {
			rows = append(rows, []string{result.ID, item.ID, item.Name, strconv.FormatBool(item.Resolved)})
		}

		return outfmt.WritePlain(os.Stdout, headers, rows)
	}

	fmt.Fprintf(os.Stderr, "Item added to checklist %s\n", cmd.ChecklistID)

	if len(result.Items) > 0 {
		// Find the last item (just added)
		lastItem := result.Items[len(result.Items)-1]
		fmt.Printf("  Item ID: %s\n", lastItem.ID)
		fmt.Printf("  Name: %s\n", lastItem.Name)
	}

	return nil
}

type ChecklistsUpdateItemCmd struct {
	ChecklistID string `arg:"" required:"" help:"Checklist ID"`
	ItemID      string `arg:"" required:"" help:"Item ID"`
	Name        string `help:"New item name"`
	Resolved    *bool  `help:"Mark as resolved (true/false)"`
	Assignee    int    `help:"Assignee user ID"`
	Parent      string `help:"Parent item ID for nesting"`
}

func (cmd *ChecklistsUpdateItemCmd) Run(ctx context.Context) error {
	client, err := getClickUpClient(ctx)
	if err != nil {
		return err
	}

	req := clickup.EditChecklistItemRequest{
		Name:     cmd.Name,
		Resolved: cmd.Resolved,
		Assignee: cmd.Assignee,
		Parent:   cmd.Parent,
	}

	result, err := client.Checklists().UpdateItem(ctx, cmd.ChecklistID, cmd.ItemID, req)
	if err != nil {
		return err
	}

	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(os.Stdout, result)
	}

	if outfmt.IsPlain(ctx) {
		headers := []string{"CHECKLIST_ID", "ITEM_ID", "NAME", "RESOLVED"}
		var rows [][]string
		for _, item := range result.Items {
			rows = append(rows, []string{result.ID, item.ID, item.Name, strconv.FormatBool(item.Resolved)})
		}

		return outfmt.WritePlain(os.Stdout, headers, rows)
	}

	fmt.Fprintf(os.Stderr, "Checklist item updated\n")

	// Find the updated item
	for _, item := range result.Items {
		if item.ID == cmd.ItemID {
			fmt.Printf("  Item ID: %s\n", item.ID)
			fmt.Printf("  Name: %s\n", item.Name)
			fmt.Printf("  Resolved: %v\n", item.Resolved)

			break
		}
	}

	return nil
}

type ChecklistsDeleteItemCmd struct {
	ChecklistID string `arg:"" required:"" help:"Checklist ID"`
	ItemID      string `arg:"" required:"" help:"Item ID"`
	Force       bool   `help:"Skip confirmation"`
}

func (cmd *ChecklistsDeleteItemCmd) Run(ctx context.Context) error {
	client, err := getClickUpClient(ctx)
	if err != nil {
		return err
	}

	if !cmd.Force {
		fmt.Fprintf(os.Stderr, "Warning: This will permanently delete item %s from checklist %s\n", cmd.ItemID, cmd.ChecklistID)
		fmt.Fprint(os.Stderr, "Use --force to confirm deletion\n")

		return fmt.Errorf("operation cancelled: use --force to confirm")
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

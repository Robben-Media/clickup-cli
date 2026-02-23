package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/builtbyrobben/clickup-cli/internal/clickup"
	"github.com/builtbyrobben/clickup-cli/internal/outfmt"
)

type ViewsCmd struct {
	List   ViewsListCmd   `cmd:"" help:"List views"`
	Get    ViewsGetCmd    `cmd:"" help:"Get a view by ID"`
	Tasks  ViewsTasksCmd  `cmd:"" help:"Get tasks in a view"`
	Create ViewsCreateCmd `cmd:"" help:"Create a new view"`
	Update ViewsUpdateCmd `cmd:"" help:"Update a view"`
	Delete ViewsDeleteCmd `cmd:"" help:"Delete a view"`
}

type ViewsListCmd struct {
	Team   string `help:"List views in a workspace/team"`
	Space  string `help:"List views in a space"`
	Folder string `help:"List views in a folder"`
	List   string `help:"List views in a list"`
}

func (cmd *ViewsListCmd) Run(ctx context.Context) error {
	client, err := getClickUpClient(ctx)
	if err != nil {
		return err
	}

	// Validate exactly one scope is provided
	scopes := 0
	if cmd.Team != "" {
		scopes++
	}

	if cmd.Space != "" {
		scopes++
	}

	if cmd.Folder != "" {
		scopes++
	}

	if cmd.List != "" {
		scopes++
	}

	if scopes != 1 {
		return fmt.Errorf("exactly one of --team, --space, --folder, or --list is required")
	}

	var result *clickup.ViewsResponse

	switch {
	case cmd.Team != "":
		result, err = client.Views().ListByTeam(ctx, cmd.Team)
	case cmd.Space != "":
		result, err = client.Views().ListBySpace(ctx, cmd.Space)
	case cmd.Folder != "":
		result, err = client.Views().ListByFolder(ctx, cmd.Folder)
	case cmd.List != "":
		result, err = client.Views().ListByList(ctx, cmd.List)
	}

	if err != nil {
		return err
	}

	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(os.Stdout, result)
	}

	if outfmt.IsPlain(ctx) {
		headers := []string{"ID", "NAME", "TYPE", "PARENT_TYPE", "PARENT_ID"}
		var rows [][]string

		for _, view := range result.Views {
			rows = append(rows, []string{
				view.ID,
				view.Name,
				view.Type,
				fmt.Sprintf("%d", view.Parent.Type),
				view.Parent.ID,
			})
		}

		return outfmt.WritePlain(os.Stdout, headers, rows)
	}

	// Human-readable output
	if len(result.Views) == 0 {
		fmt.Fprintln(os.Stderr, "No views found")

		return nil
	}

	fmt.Fprintf(os.Stderr, "Found %d views:\n\n", len(result.Views))

	for _, view := range result.Views {
		protected := ""
		if view.Protected {
			protected = " (protected)"
		}

		fmt.Printf("  %s (%s)%s\n", view.Name, view.Type, protected)
		fmt.Printf("    ID: %s\n", view.ID)
		fmt.Printf("    Parent: type=%d id=%s\n", view.Parent.Type, view.Parent.ID)
	}

	return nil
}

type ViewsGetCmd struct {
	ViewID string `arg:"" required:"" help:"View ID"`
}

func (cmd *ViewsGetCmd) Run(ctx context.Context) error {
	client, err := getClickUpClient(ctx)
	if err != nil {
		return err
	}

	result, err := client.Views().Get(ctx, cmd.ViewID)
	if err != nil {
		return err
	}

	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(os.Stdout, result)
	}

	if outfmt.IsPlain(ctx) {
		headers := []string{"ID", "NAME", "TYPE", "PROTECTED"}
		protected := "false"
		if result.Protected {
			protected = "true"
		}

		rows := [][]string{{result.ID, result.Name, result.Type, protected}}

		return outfmt.WritePlain(os.Stdout, headers, rows)
	}

	// Human-readable output
	protected := ""
	if result.Protected {
		protected = " (protected)"
	}

	fmt.Fprintf(os.Stderr, "View: %s (%s)%s\n", result.Name, result.Type, protected)
	fmt.Printf("  ID: %s\n", result.ID)
	fmt.Printf("  Parent: type=%d id=%s\n", result.Parent.Type, result.Parent.ID)

	return nil
}

type ViewsTasksCmd struct {
	ViewID string `arg:"" required:"" help:"View ID"`
	Page   int    `help:"Page number for pagination"`
}

func (cmd *ViewsTasksCmd) Run(ctx context.Context) error {
	client, err := getClickUpClient(ctx)
	if err != nil {
		return err
	}

	result, err := client.Views().Tasks(ctx, cmd.ViewID, cmd.Page)
	if err != nil {
		return err
	}

	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(os.Stdout, result)
	}

	if outfmt.IsPlain(ctx) {
		headers := []string{"ID", "NAME", "STATUS", "PRIORITY", "URL"}
		var rows [][]string

		for _, task := range result.Tasks {
			priority := ""
			if task.Priority != nil {
				priority = task.Priority.Name
			}

			rows = append(rows, []string{
				task.ID,
				task.Name,
				task.Status.Status,
				priority,
				task.URL,
			})
		}

		return outfmt.WritePlain(os.Stdout, headers, rows)
	}

	// Human-readable output
	if len(result.Tasks) == 0 {
		fmt.Fprintln(os.Stderr, "No tasks found in this view")

		return nil
	}

	fmt.Fprintf(os.Stderr, "Found %d tasks in view:\n\n", len(result.Tasks))

	for _, task := range result.Tasks {
		fmt.Printf("  %s (%s)\n", task.Name, task.Status.Status)
		fmt.Printf("    ID: %s\n", task.ID)

		if task.Priority != nil {
			fmt.Printf("    Priority: %s\n", task.Priority.Name)
		}

		fmt.Printf("    URL: %s\n", task.URL)
	}

	return nil
}

type ViewsCreateCmd struct {
	Team   string `help:"Create view in a workspace/team"`
	Space  string `help:"Create view in a space"`
	Folder string `help:"Create view in a folder"`
	List   string `help:"Create view in a list"`
	Name   string `arg:"" required:"" help:"View name"`
	Type   string `required:"" help:"View type (list, board, calendar, gantt, activity, map, workload, table)"`
}

func (cmd *ViewsCreateCmd) Run(ctx context.Context) error {
	client, err := getClickUpClient(ctx)
	if err != nil {
		return err
	}

	// Validate exactly one scope is provided
	scopes := 0
	if cmd.Team != "" {
		scopes++
	}

	if cmd.Space != "" {
		scopes++
	}

	if cmd.Folder != "" {
		scopes++
	}

	if cmd.List != "" {
		scopes++
	}

	if scopes != 1 {
		return fmt.Errorf("exactly one of --team, --space, --folder, or --list is required")
	}

	req := clickup.CreateViewRequest{
		Name: cmd.Name,
		Type: cmd.Type,
	}

	var result *clickup.View

	switch {
	case cmd.Team != "":
		result, err = client.Views().CreateInTeam(ctx, cmd.Team, req)
	case cmd.Space != "":
		result, err = client.Views().CreateInSpace(ctx, cmd.Space, req)
	case cmd.Folder != "":
		result, err = client.Views().CreateInFolder(ctx, cmd.Folder, req)
	case cmd.List != "":
		result, err = client.Views().CreateInList(ctx, cmd.List, req)
	}

	if err != nil {
		return err
	}

	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(os.Stdout, result)
	}

	if outfmt.IsPlain(ctx) {
		headers := []string{"ID", "NAME", "TYPE"}
		rows := [][]string{{result.ID, result.Name, result.Type}}

		return outfmt.WritePlain(os.Stdout, headers, rows)
	}

	fmt.Fprintf(os.Stderr, "View created\n")
	fmt.Printf("  ID: %s\n", result.ID)
	fmt.Printf("  Name: %s\n", result.Name)
	fmt.Printf("  Type: %s\n", result.Type)

	return nil
}

type ViewsUpdateCmd struct {
	ViewID string `arg:"" required:"" help:"View ID"`
	Name   string `help:"New view name"`
}

func (cmd *ViewsUpdateCmd) Run(ctx context.Context) error {
	client, err := getClickUpClient(ctx)
	if err != nil {
		return err
	}

	req := clickup.UpdateViewRequest{
		Name: cmd.Name,
	}

	result, err := client.Views().Update(ctx, cmd.ViewID, req)
	if err != nil {
		return err
	}

	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(os.Stdout, result)
	}

	if outfmt.IsPlain(ctx) {
		headers := []string{"ID", "NAME", "TYPE"}
		rows := [][]string{{result.ID, result.Name, result.Type}}

		return outfmt.WritePlain(os.Stdout, headers, rows)
	}

	fmt.Fprintf(os.Stderr, "View updated\n")
	fmt.Printf("  ID: %s\n", result.ID)
	fmt.Printf("  Name: %s\n", result.Name)
	fmt.Printf("  Type: %s\n", result.Type)

	return nil
}

type ViewsDeleteCmd struct {
	ViewID string `arg:"" required:"" help:"View ID"`
	Force  bool   `help:"Skip confirmation"`
}

func (cmd *ViewsDeleteCmd) Run(ctx context.Context) error {
	client, err := getClickUpClient(ctx)
	if err != nil {
		return err
	}

	if !cmd.Force {
		fmt.Fprintf(os.Stderr, "Warning: This will permanently delete view %s\n", cmd.ViewID)
		fmt.Fprint(os.Stderr, "Use --force to confirm deletion\n")

		return fmt.Errorf("operation cancelled: use --force to confirm")
	}

	if err := client.Views().Delete(ctx, cmd.ViewID); err != nil {
		return err
	}

	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(os.Stdout, map[string]string{
			"status":  "success",
			"message": "View deleted",
			"view_id": cmd.ViewID,
		})
	}

	if outfmt.IsPlain(ctx) {
		headers := []string{"STATUS", "VIEW_ID"}
		rows := [][]string{{"success", cmd.ViewID}}

		return outfmt.WritePlain(os.Stdout, headers, rows)
	}

	fmt.Fprintf(os.Stderr, "View %s deleted\n", cmd.ViewID)

	return nil
}

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
	Team   string `help:"Team (workspace) ID"`
	Space  string `help:"Space ID"`
	Folder string `help:"Folder ID"`
	List   string `help:"List ID"`
}

func (cmd *ViewsListCmd) Run(ctx context.Context) error {
	// Count how many scope flags are set
	scopeCount := 0
	if cmd.Team != "" {
		scopeCount++
	}
	if cmd.Space != "" {
		scopeCount++
	}
	if cmd.Folder != "" {
		scopeCount++
	}
	if cmd.List != "" {
		scopeCount++
	}

	if scopeCount == 0 {
		return fmt.Errorf("exactly one scope flag required: --team, --space, --folder, or --list")
	}

	if scopeCount > 1 {
		return fmt.Errorf("only one scope flag allowed: --team, --space, --folder, or --list")
	}

	client, err := getClickUpClient(ctx)
	if err != nil {
		return err
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

	if len(result.Views) == 0 {
		fmt.Fprintln(os.Stderr, "No views found")
		return nil
	}

	fmt.Fprintf(os.Stderr, "Found %d views\n\n", len(result.Views))

	for _, view := range result.Views {
		printView(&view)
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
		rows := [][]string{{
			result.ID,
			result.Name,
			result.Type,
			fmt.Sprintf("%t", result.Protected),
		}}
		return outfmt.WritePlain(os.Stdout, headers, rows)
	}

	printViewDetail(result)

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

	result, err := client.Views().GetTasks(ctx, cmd.ViewID, cmd.Page)
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
			rows = append(rows, []string{task.ID, task.Name, task.Status.Status, priority, task.URL})
		}
		return outfmt.WritePlain(os.Stdout, headers, rows)
	}

	if len(result.Tasks) == 0 {
		fmt.Fprintln(os.Stderr, "No tasks found in view")
		return nil
	}

	fmt.Fprintf(os.Stderr, "Found %d tasks in view\n\n", len(result.Tasks))

	for _, task := range result.Tasks {
		printTask(&task)
	}

	return nil
}

type ViewsCreateCmd struct {
	Name   string `arg:"" required:"" help:"View name"`
	Type   string `required:"" help:"View type: list, board, calendar, gantt, activity, map, workload, table"`
	Team   string `help:"Team (workspace) ID"`
	Space  string `help:"Space ID"`
	Folder string `help:"Folder ID"`
	List   string `help:"List ID"`
}

func (cmd *ViewsCreateCmd) Run(ctx context.Context) error {
	// Count how many scope flags are set
	scopeCount := 0
	if cmd.Team != "" {
		scopeCount++
	}
	if cmd.Space != "" {
		scopeCount++
	}
	if cmd.Folder != "" {
		scopeCount++
	}
	if cmd.List != "" {
		scopeCount++
	}

	if scopeCount == 0 {
		return fmt.Errorf("exactly one scope flag required: --team, --space, --folder, or --list")
	}

	if scopeCount > 1 {
		return fmt.Errorf("only one scope flag allowed: --team, --space, --folder, or --list")
	}

	client, err := getClickUpClient(ctx)
	if err != nil {
		return err
	}

	req := clickup.CreateViewRequest{
		Name: cmd.Name,
		Type: cmd.Type,
	}

	var result *clickup.View

	switch {
	case cmd.Team != "":
		result, err = client.Views().CreateByTeam(ctx, cmd.Team, req)
	case cmd.Space != "":
		result, err = client.Views().CreateBySpace(ctx, cmd.Space, req)
	case cmd.Folder != "":
		result, err = client.Views().CreateByFolder(ctx, cmd.Folder, req)
	case cmd.List != "":
		result, err = client.Views().CreateByList(ctx, cmd.List, req)
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

	fmt.Fprintf(os.Stderr, "Created view\n\n")
	printViewDetail(result)

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

	fmt.Fprintf(os.Stderr, "Updated view\n\n")
	printViewDetail(result)

	return nil
}

type ViewsDeleteCmd struct {
	ViewID string `arg:"" required:"" help:"View ID"`
}

func (cmd *ViewsDeleteCmd) Run(ctx context.Context) error {
	client, err := getClickUpClient(ctx)
	if err != nil {
		return err
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

func printView(view *clickup.View) {
	fmt.Printf("ID: %s\n", view.ID)
	fmt.Printf("  Name: %s\n", view.Name)
	fmt.Printf("  Type: %s\n", view.Type)
	fmt.Printf("  Protected: %t\n", view.Protected)
	fmt.Println()
}

func printViewDetail(view *clickup.View) {
	fmt.Printf("ID: %s\n", view.ID)
	fmt.Printf("Name: %s\n", view.Name)
	fmt.Printf("Type: %s\n", view.Type)
	fmt.Printf("Protected: %t\n", view.Protected)
	fmt.Printf("Parent ID: %s\n", view.Parent.ID)
	fmt.Printf("Parent Type: %d\n", view.Parent.Type)
}

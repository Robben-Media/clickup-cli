package cmd

import (
	"context"
	"fmt"
	"os"
	"strconv"

	"github.com/builtbyrobben/clickup-cli/internal/clickup"
	"github.com/builtbyrobben/clickup-cli/internal/outfmt"
)

type ListsCmd struct {
	List         ListsListCmd         `cmd:"" help:"List all lists in a space or folder"`
	Get          ListsGetCmd          `cmd:"" help:"Get list details"`
	Create       ListsCreateCmd       `cmd:"" help:"Create a new list"`
	Update       ListsUpdateCmd       `cmd:"" help:"Update a list"`
	Delete       ListsDeleteCmd       `cmd:"" help:"Delete a list"`
	FromTemplate ListsFromTemplateCmd `cmd:"" help:"Create list from template"`
	AddTask      ListsAddTaskCmd      `cmd:"" help:"Add task to list"`
	RemoveTask   ListsRemoveTaskCmd   `cmd:"" help:"Remove task from list"`
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

type ListsGetCmd struct {
	ListID string `arg:"" required:"" help:"List ID"`
}

func (cmd *ListsGetCmd) Run(ctx context.Context) error {
	client, err := getClickUpClient(ctx)
	if err != nil {
		return err
	}

	result, err := client.Lists().Get(ctx, cmd.ListID)
	if err != nil {
		return err
	}

	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(os.Stdout, result)
	}
	if outfmt.IsPlain(ctx) {
		headers := []string{"ID", "NAME", "TASK_COUNT", "FOLDER", "SPACE"}
		rows := [][]string{{
			result.ID,
			result.Name,
			strconv.Itoa(result.TaskCount),
			result.Folder.Name,
			result.Space.ID,
		}}
		return outfmt.WritePlain(os.Stdout, headers, rows)
	}

	fmt.Fprintf(os.Stderr, "List Details\n\n")
	fmt.Printf("ID: %s\n", result.ID)
	fmt.Printf("Name: %s\n", result.Name)
	fmt.Printf("Task Count: %d\n", result.TaskCount)
	fmt.Printf("Folder: %s (%s)\n", result.Folder.Name, result.Folder.ID)
	fmt.Printf("Space ID: %s\n", result.Space.ID)

	return nil
}

type ListsCreateCmd struct {
	Name     string `arg:"" required:"" help:"List name"`
	Folder   string `help:"Folder ID to create list in (required unless --space is set)"`
	Space    string `help:"Space ID for folderless list (required unless --folder is set)"`
	Content  string `help:"List description"`
	DueDate  int64  `help:"Due date in milliseconds"`
	Priority int    `help:"Priority (1-4)"`
	Assignee int    `help:"Assignee user ID"`
}

func (cmd *ListsCreateCmd) Run(ctx context.Context) error {
	client, err := getClickUpClient(ctx)
	if err != nil {
		return err
	}

	if cmd.Folder == "" && cmd.Space == "" {
		return fmt.Errorf("either --folder or --space is required")
	}

	if cmd.Folder != "" && cmd.Space != "" {
		return fmt.Errorf("only one of --folder or --space can be set, not both")
	}

	req := clickup.CreateListRequest{
		Name:     cmd.Name,
		Content:  cmd.Content,
		DueDate:  cmd.DueDate,
		Priority: cmd.Priority,
		Assignee: cmd.Assignee,
	}

	var result *clickup.ListDetail

	if cmd.Folder != "" {
		result, err = client.Lists().CreateInFolder(ctx, cmd.Folder, req)
	} else {
		result, err = client.Lists().CreateFolderless(ctx, cmd.Space, req)
	}

	if err != nil {
		return err
	}

	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(os.Stdout, result)
	}
	if outfmt.IsPlain(ctx) {
		headers := []string{"ID", "NAME", "TASK_COUNT"}
		rows := [][]string{{result.ID, result.Name, strconv.Itoa(result.TaskCount)}}
		return outfmt.WritePlain(os.Stdout, headers, rows)
	}

	fmt.Fprintf(os.Stderr, "List created\n\n")
	fmt.Printf("ID: %s\n", result.ID)
	fmt.Printf("Name: %s\n", result.Name)

	return nil
}

type ListsUpdateCmd struct {
	ListID        string `arg:"" required:"" help:"List ID"`
	Name          string `help:"New list name"`
	Content       string `help:"List description"`
	DueDate       int64  `help:"Due date in milliseconds"`
	Priority      int    `help:"Priority (1-4)"`
	Assignee      int    `help:"Assignee user ID"`
	UnsetAssignee bool   `help:"Remove assignee from list"`
}

func (cmd *ListsUpdateCmd) Run(ctx context.Context) error {
	client, err := getClickUpClient(ctx)
	if err != nil {
		return err
	}

	req := clickup.UpdateListRequest{
		Name:          cmd.Name,
		Content:       cmd.Content,
		DueDate:       cmd.DueDate,
		Priority:      cmd.Priority,
		Assignee:      cmd.Assignee,
		UnsetAssignee: cmd.UnsetAssignee,
	}

	result, err := client.Lists().Update(ctx, cmd.ListID, req)
	if err != nil {
		return err
	}

	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(os.Stdout, result)
	}
	if outfmt.IsPlain(ctx) {
		headers := []string{"ID", "NAME", "TASK_COUNT"}
		rows := [][]string{{result.ID, result.Name, strconv.Itoa(result.TaskCount)}}
		return outfmt.WritePlain(os.Stdout, headers, rows)
	}

	fmt.Fprintf(os.Stderr, "List updated\n\n")
	fmt.Printf("ID: %s\n", result.ID)
	fmt.Printf("Name: %s\n", result.Name)

	return nil
}

type ListsDeleteCmd struct {
	ListID string `arg:"" required:"" help:"List ID"`
	Force  bool   `help:"Skip confirmation"`
}

func (cmd *ListsDeleteCmd) Run(ctx context.Context) error {
	client, err := getClickUpClient(ctx)
	if err != nil {
		return err
	}

	if !cmd.Force && !outfmt.IsPlain(ctx) && !outfmt.IsJSON(ctx) {
		fmt.Fprintf(os.Stderr, "WARNING: Deleting a list will delete all tasks in it.\n")
		fmt.Fprintf(os.Stderr, "List ID: %s\n\n", cmd.ListID)
		fmt.Fprintf(os.Stderr, "Use --force to skip this confirmation.\n")

		return fmt.Errorf("operation cancelled: use --force to confirm destructive operation")
	}

	if err := client.Lists().Delete(ctx, cmd.ListID); err != nil {
		return err
	}

	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(os.Stdout, map[string]string{
			"status":  "success",
			"message": "List deleted",
			"list_id": cmd.ListID,
		})
	}
	if outfmt.IsPlain(ctx) {
		headers := []string{"STATUS", "LIST_ID"}
		rows := [][]string{{"success", cmd.ListID}}
		return outfmt.WritePlain(os.Stdout, headers, rows)
	}

	fmt.Fprintf(os.Stderr, "List %s deleted\n", cmd.ListID)

	return nil
}

type ListsFromTemplateCmd struct {
	TemplateID string `arg:"" required:"" help:"Template ID"`
	Name       string `arg:"" required:"" help:"New list name"`
	Folder     string `help:"Folder ID to create list in (required unless --space is set)"`
	Space      string `help:"Space ID for folderless list (required unless --folder is set)"`
}

func (cmd *ListsFromTemplateCmd) Run(ctx context.Context) error {
	client, err := getClickUpClient(ctx)
	if err != nil {
		return err
	}

	if cmd.Folder == "" && cmd.Space == "" {
		return fmt.Errorf("either --folder or --space is required")
	}

	if cmd.Folder != "" && cmd.Space != "" {
		return fmt.Errorf("only one of --folder or --space can be set, not both")
	}

	req := clickup.CreateListFromTemplateRequest{
		Name: cmd.Name,
	}

	var result *clickup.ListDetail

	if cmd.Folder != "" {
		result, err = client.Lists().CreateFromTemplateInFolder(ctx, cmd.Folder, cmd.TemplateID, req)
	} else {
		result, err = client.Lists().CreateFromTemplateInSpace(ctx, cmd.Space, cmd.TemplateID, req)
	}

	if err != nil {
		return err
	}

	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(os.Stdout, result)
	}
	if outfmt.IsPlain(ctx) {
		headers := []string{"ID", "NAME"}
		rows := [][]string{{result.ID, result.Name}}
		return outfmt.WritePlain(os.Stdout, headers, rows)
	}

	fmt.Fprintf(os.Stderr, "List created from template\n\n")
	fmt.Printf("ID: %s\n", result.ID)
	fmt.Printf("Name: %s\n", result.Name)

	return nil
}

type ListsAddTaskCmd struct {
	ListID string `arg:"" required:"" help:"List ID"`
	TaskID string `arg:"" required:"" help:"Task ID"`
}

func (cmd *ListsAddTaskCmd) Run(ctx context.Context) error {
	client, err := getClickUpClient(ctx)
	if err != nil {
		return err
	}

	if err := client.Lists().AddTask(ctx, cmd.ListID, cmd.TaskID); err != nil {
		return err
	}

	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(os.Stdout, map[string]string{
			"status":  "success",
			"message": "Task added to list",
			"list_id": cmd.ListID,
			"task_id": cmd.TaskID,
		})
	}
	if outfmt.IsPlain(ctx) {
		headers := []string{"STATUS", "LIST_ID", "TASK_ID"}
		rows := [][]string{{"success", cmd.ListID, cmd.TaskID}}
		return outfmt.WritePlain(os.Stdout, headers, rows)
	}

	fmt.Fprintf(os.Stderr, "Task %s added to list %s\n", cmd.TaskID, cmd.ListID)

	return nil
}

type ListsRemoveTaskCmd struct {
	ListID string `arg:"" required:"" help:"List ID"`
	TaskID string `arg:"" required:"" help:"Task ID"`
}

func (cmd *ListsRemoveTaskCmd) Run(ctx context.Context) error {
	client, err := getClickUpClient(ctx)
	if err != nil {
		return err
	}

	if err := client.Lists().RemoveTask(ctx, cmd.ListID, cmd.TaskID); err != nil {
		return err
	}

	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(os.Stdout, map[string]string{
			"status":  "success",
			"message": "Task removed from list",
			"list_id": cmd.ListID,
			"task_id": cmd.TaskID,
		})
	}
	if outfmt.IsPlain(ctx) {
		headers := []string{"STATUS", "LIST_ID", "TASK_ID"}
		rows := [][]string{{"success", cmd.ListID, cmd.TaskID}}
		return outfmt.WritePlain(os.Stdout, headers, rows)
	}

	fmt.Fprintf(os.Stderr, "Task %s removed from list %s\n", cmd.TaskID, cmd.ListID)

	return nil
}

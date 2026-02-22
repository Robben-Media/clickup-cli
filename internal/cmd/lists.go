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
	AddTask      ListsAddTaskCmd      `cmd:"" help:"Add a task to a list"`
	RemoveTask   ListsRemoveTaskCmd   `cmd:"" help:"Remove a task from a list"`
	FromTemplate ListsFromTemplateCmd `cmd:"" help:"Create a list from a template"`
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
	ListID string `arg:"" help:"List ID"`
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
		folderName := ""
		if result.Folder.Name != "" {
			folderName = result.Folder.Name
		}
		rows := [][]string{{result.ID, result.Name, strconv.Itoa(result.TaskCount), folderName, result.Space.Name}}
		return outfmt.WritePlain(os.Stdout, headers, rows)
	}

	fmt.Printf("ID: %s\n", result.ID)
	fmt.Printf("Name: %s\n", result.Name)
	fmt.Printf("Task Count: %d\n", result.TaskCount)
	if result.Folder.ID != "" {
		fmt.Printf("Folder: %s (%s)\n", result.Folder.Name, result.Folder.ID)
	}
	if result.Space.ID != "" {
		fmt.Printf("Space: %s (%s)\n", result.Space.Name, result.Space.ID)
	}

	return nil
}

type ListsCreateCmd struct {
	Name     string `arg:"" help:"List name"`
	Folder   string `help:"Folder ID to create list in (mutually exclusive with --space)"`
	Space    string `help:"Space ID to create folderless list in (mutually exclusive with --folder)"`
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
		return fmt.Errorf("--folder and --space are mutually exclusive")
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
		rows := [][]string{{result.ID, result.Name, "0"}}
		return outfmt.WritePlain(os.Stdout, headers, rows)
	}

	fmt.Fprintf(os.Stderr, "List created\n\n")
	fmt.Printf("ID: %s\n", result.ID)
	fmt.Printf("Name: %s\n", result.Name)

	return nil
}

type ListsUpdateCmd struct {
	ListID        string `arg:"" help:"List ID"`
	Name          string `help:"New list name"`
	Content       string `help:"New description"`
	DueDate       int64  `help:"New due date in milliseconds"`
	Priority      int    `help:"New priority (1-4)"`
	Assignee      int    `help:"New assignee user ID"`
	UnsetAssignee bool   `help:"Remove assignee"`
	UnsetDueDate  bool   `help:"Remove due date"`
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
		UnsetDueDate:  cmd.UnsetDueDate,
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
	ListID string `arg:"" help:"List ID"`
}

func (cmd *ListsDeleteCmd) Run(ctx context.Context) error {
	client, err := getClickUpClient(ctx)
	if err != nil {
		return err
	}

	err = client.Lists().Delete(ctx, cmd.ListID)
	if err != nil {
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

	fmt.Printf("List %s deleted\n", cmd.ListID)

	return nil
}

type ListsAddTaskCmd struct {
	ListID string `arg:"" help:"List ID"`
	TaskID string `arg:"" help:"Task ID"`
}

func (cmd *ListsAddTaskCmd) Run(ctx context.Context) error {
	client, err := getClickUpClient(ctx)
	if err != nil {
		return err
	}

	err = client.Lists().AddTask(ctx, cmd.ListID, cmd.TaskID)
	if err != nil {
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

	fmt.Printf("Task %s added to list %s\n", cmd.TaskID, cmd.ListID)

	return nil
}

type ListsRemoveTaskCmd struct {
	ListID string `arg:"" help:"List ID"`
	TaskID string `arg:"" help:"Task ID"`
}

func (cmd *ListsRemoveTaskCmd) Run(ctx context.Context) error {
	client, err := getClickUpClient(ctx)
	if err != nil {
		return err
	}

	err = client.Lists().RemoveTask(ctx, cmd.ListID, cmd.TaskID)
	if err != nil {
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

	fmt.Printf("Task %s removed from list %s\n", cmd.TaskID, cmd.ListID)

	return nil
}

type ListsFromTemplateCmd struct {
	TemplateID string `arg:"" help:"Template ID"`
	Folder     string `help:"Folder ID to create list in (mutually exclusive with --space)"`
	Space      string `help:"Space ID to create folderless list in (mutually exclusive with --folder)"`
	Name       string `help:"Name for the new list"`
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
		return fmt.Errorf("--folder and --space are mutually exclusive")
	}

	req := clickup.CreateListFromTemplateRequest{
		Name: cmd.Name,
	}

	var result *clickup.ListDetail

	if cmd.Folder != "" {
		result, err = client.Lists().FromTemplateInFolder(ctx, cmd.Folder, cmd.TemplateID, req)
	} else {
		result, err = client.Lists().FromTemplateInSpace(ctx, cmd.Space, cmd.TemplateID, req)
	}

	if err != nil {
		return err
	}

	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(os.Stdout, result)
	}
	if outfmt.IsPlain(ctx) {
		headers := []string{"ID", "NAME", "TASK_COUNT"}
		rows := [][]string{{result.ID, result.Name, "0"}}
		return outfmt.WritePlain(os.Stdout, headers, rows)
	}

	fmt.Fprintf(os.Stderr, "List created from template\n\n")
	fmt.Printf("ID: %s\n", result.ID)
	fmt.Printf("Name: %s\n", result.Name)

	return nil
}

package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/builtbyrobben/clickup-cli/internal/clickup"
	"github.com/builtbyrobben/clickup-cli/internal/outfmt"
)

type TasksCmd struct {
	List   TasksListCmd   `cmd:"" help:"List tasks in a list"`
	Get    TasksGetCmd    `cmd:"" help:"Get a task by ID"`
	Create TasksCreateCmd `cmd:"" help:"Create a new task"`
	Update TasksUpdateCmd `cmd:"" help:"Update a task"`
	Delete TasksDeleteCmd `cmd:"" help:"Delete a task"`
}

type TasksListCmd struct {
	List     string `required:"" help:"List ID to fetch tasks from"`
	Status   string `help:"Filter by status (e.g. open, closed)"`
	Assignee string `help:"Filter by assignee name or ID"`
}

func (cmd *TasksListCmd) Run(ctx context.Context) error {
	client, err := getClickUpClient()
	if err != nil {
		return err
	}

	result, err := client.Tasks().List(ctx, cmd.List, cmd.Status, cmd.Assignee)
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
		fmt.Fprintln(os.Stderr, "No tasks found")
		return nil
	}

	fmt.Fprintf(os.Stderr, "Found %d tasks\n\n", len(result.Tasks))

	for _, task := range result.Tasks {
		printTask(&task)
	}

	return nil
}

type TasksGetCmd struct {
	TaskID string `arg:"" required:"" help:"Task ID"`
}

func (cmd *TasksGetCmd) Run(ctx context.Context) error {
	client, err := getClickUpClient()
	if err != nil {
		return err
	}

	result, err := client.Tasks().Get(ctx, cmd.TaskID)
	if err != nil {
		return err
	}

	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(os.Stdout, result)
	}
	if outfmt.IsPlain(ctx) {
		headers := []string{"ID", "NAME", "STATUS", "PRIORITY", "DUE_DATE", "URL"}
		priority := ""
		if result.Priority != nil {
			priority = result.Priority.Name
		}
		rows := [][]string{{result.ID, result.Name, result.Status.Status, priority, result.DueDate, result.URL}}
		return outfmt.WritePlain(os.Stdout, headers, rows)
	}

	printTaskDetail(result)

	return nil
}

type TasksCreateCmd struct {
	ListID   string `arg:"" required:"" help:"List ID to create task in"`
	Name     string `arg:"" required:"" help:"Task name"`
	Assignee int    `help:"Assign to user ID"`
	Priority *int   `help:"Priority (1=urgent, 2=high, 3=normal, 4=low)"`
	Due      string `help:"Due date (unix timestamp in milliseconds)"`
}

func (cmd *TasksCreateCmd) Run(ctx context.Context) error {
	client, err := getClickUpClient()
	if err != nil {
		return err
	}

	req := clickup.CreateTaskRequest{
		Name:    cmd.Name,
		DueDate: cmd.Due,
	}

	if cmd.Assignee != 0 {
		req.Assignees = []int{cmd.Assignee}
	}

	if cmd.Priority != nil {
		req.Priority = cmd.Priority
	}

	result, err := client.Tasks().Create(ctx, cmd.ListID, req)
	if err != nil {
		return err
	}

	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(os.Stdout, result)
	}
	if outfmt.IsPlain(ctx) {
		headers := []string{"ID", "NAME", "STATUS", "URL"}
		rows := [][]string{{result.ID, result.Name, result.Status.Status, result.URL}}
		return outfmt.WritePlain(os.Stdout, headers, rows)
	}

	fmt.Fprintf(os.Stderr, "Created task\n\n")
	printTaskDetail(result)

	return nil
}

type TasksUpdateCmd struct {
	TaskID   string `arg:"" required:"" help:"Task ID"`
	Status   string `help:"New status"`
	Name     string `help:"New name"`
	Assignee int    `help:"Assign to user ID (adds assignee)"`
	Unassign int    `help:"Unassign user ID (removes assignee)"`
	Priority *int   `help:"New priority (1=urgent, 2=high, 3=normal, 4=low)"`
}

func (cmd *TasksUpdateCmd) Run(ctx context.Context) error {
	client, err := getClickUpClient()
	if err != nil {
		return err
	}

	req := clickup.UpdateTaskRequest{
		Name:   cmd.Name,
		Status: cmd.Status,
	}

	if cmd.Assignee != 0 || cmd.Unassign != 0 {
		req.Assignees = &clickup.TaskAssigneesUpdate{}
		if cmd.Assignee != 0 {
			req.Assignees.Add = []int{cmd.Assignee}
		}
		if cmd.Unassign != 0 {
			req.Assignees.Rem = []int{cmd.Unassign}
		}
	}

	if cmd.Priority != nil {
		req.Priority = cmd.Priority
	}

	result, err := client.Tasks().Update(ctx, cmd.TaskID, req)
	if err != nil {
		return err
	}

	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(os.Stdout, result)
	}
	if outfmt.IsPlain(ctx) {
		headers := []string{"ID", "NAME", "STATUS", "URL"}
		rows := [][]string{{result.ID, result.Name, result.Status.Status, result.URL}}
		return outfmt.WritePlain(os.Stdout, headers, rows)
	}

	fmt.Fprintf(os.Stderr, "Updated task\n\n")
	printTaskDetail(result)

	return nil
}

type TasksDeleteCmd struct {
	TaskID string `arg:"" required:"" help:"Task ID"`
}

func (cmd *TasksDeleteCmd) Run(ctx context.Context) error {
	client, err := getClickUpClient()
	if err != nil {
		return err
	}

	if err := client.Tasks().Delete(ctx, cmd.TaskID); err != nil {
		return err
	}

	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(os.Stdout, map[string]string{
			"status":  "success",
			"message": "Task deleted",
			"task_id": cmd.TaskID,
		})
	}
	if outfmt.IsPlain(ctx) {
		headers := []string{"STATUS", "TASK_ID"}
		rows := [][]string{{"success", cmd.TaskID}}
		return outfmt.WritePlain(os.Stdout, headers, rows)
	}

	fmt.Fprintf(os.Stderr, "Task %s deleted\n", cmd.TaskID)

	return nil
}

func printTask(task *clickup.Task) {
	fmt.Printf("ID: %s\n", task.ID)
	fmt.Printf("  Name: %s\n", task.Name)
	fmt.Printf("  Status: %s\n", task.Status.Status)

	if task.Priority != nil {
		fmt.Printf("  Priority: %s\n", task.Priority.Name)
	}

	if task.URL != "" {
		fmt.Printf("  URL: %s\n", task.URL)
	}

	fmt.Println()
}

func printTaskDetail(task *clickup.Task) {
	fmt.Printf("ID: %s\n", task.ID)
	fmt.Printf("Name: %s\n", task.Name)
	fmt.Printf("Status: %s\n", task.Status.Status)

	if task.Priority != nil {
		fmt.Printf("Priority: %s\n", task.Priority.Name)
	}

	if task.Description != "" {
		fmt.Printf("Description: %s\n", task.Description)
	}

	if task.DueDate != "" {
		fmt.Printf("Due Date: %s\n", task.DueDate)
	}

	if len(task.Assignees) > 0 {
		fmt.Print("Assignees:")

		for _, a := range task.Assignees {
			fmt.Printf(" %s", a.Username)
		}

		fmt.Println()
	}

	if task.URL != "" {
		fmt.Printf("URL: %s\n", task.URL)
	}
}

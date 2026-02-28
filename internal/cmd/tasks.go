package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/builtbyrobben/clickup-cli/internal/clickup"
	"github.com/builtbyrobben/clickup-cli/internal/outfmt"
)

type TasksCmd struct {
	List             TasksListCmd             `cmd:"" help:"List tasks in a list"`
	Get              TasksGetCmd              `cmd:"" help:"Get a task by ID"`
	Create           TasksCreateCmd           `cmd:"" help:"Create a new task"`
	Update           TasksUpdateCmd           `cmd:"" help:"Update a task"`
	Delete           TasksDeleteCmd           `cmd:"" help:"Delete a task"`
	Search           TasksSearchCmd           `cmd:"" help:"Search tasks across workspace"`
	TimeInStatus     TasksTimeInStatusCmd     `cmd:"" help:"Get time-in-status for a task"`
	BulkTimeInStatus TasksBulkTimeInStatusCmd `cmd:"" help:"Get time-in-status for multiple tasks"`
	Merge            TasksMergeCmd            `cmd:"" help:"Merge tasks into one"`
	Move             TasksMoveCmd             `cmd:"" help:"Move a task to a different list"`
	FromTemplate     TasksFromTemplateCmd     `cmd:"" help:"Create a task from a template"`
}

type TasksListCmd struct {
	List     string `required:"" help:"List ID to fetch tasks from"`
	Status   string `help:"Filter by status (e.g. open, closed)"`
	Assignee string `help:"Filter by assignee name or ID"`
}

func (cmd *TasksListCmd) Run(ctx context.Context) error {
	client, err := getClickUpClient(ctx)
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
	client, err := getClickUpClient(ctx)
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
	client, err := getClickUpClient(ctx)
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
	client, err := getClickUpClient(ctx)
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
	client, err := getClickUpClient(ctx)
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

// TasksSearchCmd searches for tasks across a workspace.
type TasksSearchCmd struct {
	TeamID        string   `required:"" help:"Team ID to search within"`
	Status        []string `help:"Filter by status (can be repeated)"`
	Assignee      []int    `help:"Filter by assignee user ID (can be repeated)"`
	Tag           []string `help:"Filter by tag (can be repeated)"`
	DueDateGt     int64    `help:"Due date greater than (unix ms)"`
	DueDateLt     int64    `help:"Due date less than (unix ms)"`
	IncludeClosed bool     `help:"Include closed tasks"`
	Page          int      `help:"Page number (0-indexed)"`
	OrderBy       string   `help:"Order by field (e.g. due_date, created)"`
}

func (cmd *TasksSearchCmd) Run(ctx context.Context) error {
	client, err := getClickUpClient(ctx)
	if err != nil {
		return err
	}

	params := clickup.FilteredTeamTasksParams{
		Page:          cmd.Page,
		OrderBy:       cmd.OrderBy,
		Statuses:      cmd.Status,
		Assignees:     cmd.Assignee,
		Tags:          cmd.Tag,
		DueDateGt:     cmd.DueDateGt,
		DueDateLt:     cmd.DueDateLt,
		IncludeClosed: cmd.IncludeClosed,
	}

	result, err := client.Tasks().Search(ctx, cmd.TeamID, params)
	if err != nil {
		return err
	}

	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(os.Stdout, result)
	}
	if outfmt.IsPlain(ctx) {
		headers := []string{"ID", "NAME", "STATUS", "PRIORITY", "LIST", "URL"}
		rows := make([][]string, 0, len(result.Tasks))
		for _, task := range result.Tasks {
			priority := ""
			if task.Priority != nil {
				priority = task.Priority.Name
			}
			rows = append(rows, []string{task.ID, task.Name, task.Status.Status, priority, task.List.Name, task.URL})
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

// TasksTimeInStatusCmd gets time-in-status for a single task.
type TasksTimeInStatusCmd struct {
	TaskID string `arg:"" required:"" help:"Task ID"`
}

func (cmd *TasksTimeInStatusCmd) Run(ctx context.Context) error {
	client, err := getClickUpClient(ctx)
	if err != nil {
		return err
	}

	result, err := client.Tasks().TimeInStatus(ctx, cmd.TaskID)
	if err != nil {
		return err
	}

	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(os.Stdout, result)
	}
	if outfmt.IsPlain(ctx) {
		headers := []string{"STATUS", "MINUTES", "CURRENT"}
		var rows [][]string

		for _, s := range result.StatusHistory {
			rows = append(rows, []string{s.Status, formatMinutes(s.TotalTime.ByMinute), "false"})
		}

		if result.CurrentStatus != nil {
			rows = append(rows, []string{result.CurrentStatus.Status, formatMinutes(result.CurrentStatus.TotalTime.ByMinute), "true"})
		}

		return outfmt.WritePlain(os.Stdout, headers, rows)
	}

	fmt.Fprintf(os.Stderr, "Time in Status for task %s\n\n", cmd.TaskID)

	for _, s := range result.StatusHistory {
		fmt.Printf("  %s: %s\n", s.Status, formatMinutesToDuration(s.TotalTime.ByMinute))
	}

	if result.CurrentStatus != nil {
		fmt.Printf("  %s: %s (current)\n", result.CurrentStatus.Status, formatMinutesToDuration(result.CurrentStatus.TotalTime.ByMinute))
	}

	return nil
}

// TasksBulkTimeInStatusCmd gets time-in-status for multiple tasks.
type TasksBulkTimeInStatusCmd struct {
	TaskIDs []string `arg:"" required:"" help:"Task IDs"`
}

func (cmd *TasksBulkTimeInStatusCmd) Run(ctx context.Context) error {
	client, err := getClickUpClient(ctx)
	if err != nil {
		return err
	}

	result, err := client.Tasks().BulkTimeInStatus(ctx, cmd.TaskIDs)
	if err != nil {
		return err
	}

	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(os.Stdout, result)
	}
	if outfmt.IsPlain(ctx) {
		headers := []string{"TASK_ID", "STATUS", "MINUTES", "CURRENT"}
		var rows [][]string

		for taskID, data := range result {
			for _, s := range data.StatusHistory {
				rows = append(rows, []string{taskID, s.Status, formatMinutes(s.TotalTime.ByMinute), "false"})
			}

			if data.CurrentStatus != nil {
				rows = append(rows, []string{taskID, data.CurrentStatus.Status, formatMinutes(data.CurrentStatus.TotalTime.ByMinute), "true"})
			}
		}

		return outfmt.WritePlain(os.Stdout, headers, rows)
	}

	for taskID, data := range result {
		fmt.Fprintf(os.Stderr, "Task %s:\n", taskID)

		for _, s := range data.StatusHistory {
			fmt.Printf("  %s: %s\n", s.Status, formatMinutesToDuration(s.TotalTime.ByMinute))
		}

		if data.CurrentStatus != nil {
			fmt.Printf("  %s: %s (current)\n", data.CurrentStatus.Status, formatMinutesToDuration(data.CurrentStatus.TotalTime.ByMinute))
		}

		fmt.Println()
	}

	return nil
}

// TasksMergeCmd merges tasks into one.
type TasksMergeCmd struct {
	TargetTaskID  string   `arg:"" required:"" help:"Target task ID to merge into"`
	SourceTaskIDs []string `required:"" help:"Source task IDs to merge (comma-separated or repeated --source)"`
}

func (cmd *TasksMergeCmd) Run(ctx context.Context) error {
	client, err := getClickUpClient(ctx)
	if err != nil {
		return err
	}

	if len(cmd.SourceTaskIDs) == 0 {
		return fmt.Errorf("at least one source task ID is required")
	}

	result, err := client.Tasks().Merge(ctx, cmd.TargetTaskID, cmd.SourceTaskIDs)
	if err != nil {
		return err
	}

	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(os.Stdout, map[string]string{
			"status":      "success",
			"message":     "Tasks merged",
			"target_id":   cmd.TargetTaskID,
			"merged_into": result.ID,
		})
	}
	if outfmt.IsPlain(ctx) {
		headers := []string{"STATUS", "TARGET_ID", "MERGED_INTO"}
		rows := [][]string{{"success", cmd.TargetTaskID, result.ID}}
		return outfmt.WritePlain(os.Stdout, headers, rows)
	}

	fmt.Fprintf(os.Stderr, "Tasks merged into %s\n", result.ID)

	return nil
}

// TasksMoveCmd moves a task to a different list.
type TasksMoveCmd struct {
	TaskID string `arg:"" required:"" help:"Task ID to move"`
	ListID string `required:"" help:"Target list ID"`
}

func (cmd *TasksMoveCmd) Run(ctx context.Context) error {
	client, err := getClickUpClient(ctx)
	if err != nil {
		return err
	}

	result, err := client.Tasks().Move(ctx, cmd.TaskID, cmd.ListID)
	if err != nil {
		return err
	}

	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(os.Stdout, result)
	}
	if outfmt.IsPlain(ctx) {
		headers := []string{"STATUS", "TASK_ID", "LIST_ID"}
		rows := [][]string{{result.Status, result.TaskID, result.ListID}}
		return outfmt.WritePlain(os.Stdout, headers, rows)
	}

	fmt.Fprintf(os.Stderr, "Task %s moved to list %s\n", cmd.TaskID, cmd.ListID)

	return nil
}

// TasksFromTemplateCmd creates a task from a template.
type TasksFromTemplateCmd struct {
	ListID     string `arg:"" required:"" help:"List ID to create task in"`
	TemplateID string `arg:"" required:"" help:"Template ID"`
	Name       string `help:"Override task name"`
}

func (cmd *TasksFromTemplateCmd) Run(ctx context.Context) error {
	client, err := getClickUpClient(ctx)
	if err != nil {
		return err
	}

	req := clickup.CreateTaskFromTemplateRequest{
		Name: cmd.Name,
	}

	result, err := client.Tasks().CreateFromTemplate(ctx, cmd.ListID, cmd.TemplateID, req)
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

	fmt.Fprintf(os.Stderr, "Created task from template\n\n")
	printTaskDetail(result)

	return nil
}

// formatMinutes formats minutes as a string.
func formatMinutes(minutes int64) string {
	return fmt.Sprintf("%d", minutes)
}

// formatMinutesToDuration formats minutes into a human-readable duration.
func formatMinutesToDuration(minutes int64) string {
	hours := minutes / 60
	mins := minutes % 60

	return fmt.Sprintf("%dh %dm", hours, mins)
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

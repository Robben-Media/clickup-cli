package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/builtbyrobben/clickup-cli/internal/clickup"
	"github.com/builtbyrobben/clickup-cli/internal/outfmt"
)

type GoalsCmd struct {
	List            GoalsListCmd            `cmd:"" help:"List goals"`
	Get             GoalsGetCmd             `cmd:"" help:"Get a goal by ID"`
	Create          GoalsCreateCmd          `cmd:"" help:"Create a goal"`
	Update          GoalsUpdateCmd          `cmd:"" help:"Update a goal"`
	Delete          GoalsDeleteCmd          `cmd:"" help:"Delete a goal"`
	AddKeyResult    GoalsAddKeyResultCmd    `cmd:"" name:"add-key-result" help:"Add a key result to a goal"`
	UpdateKeyResult GoalsUpdateKeyResultCmd `cmd:"" name:"update-key-result" help:"Update a key result"`
	DeleteKeyResult GoalsDeleteKeyResultCmd `cmd:"" name:"delete-key-result" help:"Delete a key result"`
}

type GoalsListCmd struct {
	Team string `required:"" help:"Team (workspace) ID"`
}

func (cmd *GoalsListCmd) Run(ctx context.Context) error {
	client, err := getClickUpClient(ctx)
	if err != nil {
		return err
	}

	result, err := client.Goals().List(ctx, cmd.Team)
	if err != nil {
		return err
	}

	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(os.Stdout, result)
	}
	if outfmt.IsPlain(ctx) {
		headers := []string{"ID", "NAME", "PERCENT_COMPLETE", "DUE_DATE"}
		var rows [][]string
		for _, g := range result.Goals {
			rows = append(rows, []string{
				g.ID,
				g.Name,
				fmt.Sprintf("%d", g.PercentCompleted),
				g.DueDate,
			})
		}
		return outfmt.WritePlain(os.Stdout, headers, rows)
	}

	if len(result.Goals) == 0 {
		fmt.Fprintln(os.Stderr, "No goals found")
		return nil
	}

	fmt.Fprintf(os.Stderr, "Found %d goals\n\n", len(result.Goals))

	for _, g := range result.Goals {
		printGoal(&g)
	}

	return nil
}

type GoalsGetCmd struct {
	GoalID string `arg:"" required:"" help:"Goal ID"`
}

func (cmd *GoalsGetCmd) Run(ctx context.Context) error {
	client, err := getClickUpClient(ctx)
	if err != nil {
		return err
	}

	result, err := client.Goals().Get(ctx, cmd.GoalID)
	if err != nil {
		return err
	}

	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(os.Stdout, result)
	}
	if outfmt.IsPlain(ctx) {
		headers := []string{"ID", "NAME", "PERCENT_COMPLETE", "KEY_RESULT_COUNT"}
		krCount := len(result.KeyResults)
		rows := [][]string{{result.ID, result.Name, fmt.Sprintf("%d", result.PercentCompleted), fmt.Sprintf("%d", krCount)}}
		return outfmt.WritePlain(os.Stdout, headers, rows)
	}

	printGoalDetail(result)

	return nil
}

type GoalsCreateCmd struct {
	Team        string   `arg:"" required:"" help:"Team (workspace) ID"`
	Name        string   `arg:"" required:"" help:"Goal name"`
	DueDate     int64    `help:"Due date (unix timestamp in milliseconds)"`
	Description string   `help:"Goal description"`
	Owners      []string `help:"Owner user IDs (comma-separated or multiple flags)"`
	Color       string   `help:"Goal color (hex, e.g., #FF0000)"`
}

func (cmd *GoalsCreateCmd) Run(ctx context.Context) error {
	client, err := getClickUpClient(ctx)
	if err != nil {
		return err
	}

	req := clickup.CreateGoalRequest{
		Name:        cmd.Name,
		DueDate:     cmd.DueDate,
		Description: cmd.Description,
		Color:       cmd.Color,
	}

	for _, owner := range cmd.Owners {
		var id int
		if _, parseErr := fmt.Sscanf(owner, "%d", &id); parseErr == nil {
			req.Owners = append(req.Owners, id)
		}
	}

	result, err := client.Goals().Create(ctx, cmd.Team, req)
	if err != nil {
		return err
	}

	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(os.Stdout, result)
	}
	if outfmt.IsPlain(ctx) {
		headers := []string{"ID", "NAME", "DUE_DATE"}
		rows := [][]string{{result.ID, result.Name, result.DueDate}}
		return outfmt.WritePlain(os.Stdout, headers, rows)
	}

	fmt.Fprintf(os.Stderr, "Created goal\n\n")
	printGoalDetail(result)

	return nil
}

type GoalsUpdateCmd struct {
	GoalID      string   `arg:"" required:"" help:"Goal ID"`
	Name        string   `help:"New name"`
	Description string   `help:"New description"`
	DueDate     int64    `help:"New due date (unix timestamp in milliseconds)"`
	Color       string   `help:"New color (hex)"`
	AddOwners   []string `help:"User IDs to add as owners"`
	RemOwners   []string `help:"User IDs to remove from owners"`
}

func (cmd *GoalsUpdateCmd) Run(ctx context.Context) error {
	client, err := getClickUpClient(ctx)
	if err != nil {
		return err
	}

	req := clickup.UpdateGoalRequest{
		Name:        cmd.Name,
		Description: cmd.Description,
		DueDate:     cmd.DueDate,
		Color:       cmd.Color,
	}

	for _, owner := range cmd.AddOwners {
		var id int
		if _, parseErr := fmt.Sscanf(owner, "%d", &id); parseErr == nil {
			req.AddOwners = append(req.AddOwners, id)
		}
	}

	for _, owner := range cmd.RemOwners {
		var id int
		if _, parseErr := fmt.Sscanf(owner, "%d", &id); parseErr == nil {
			req.RemOwners = append(req.RemOwners, id)
		}
	}

	result, err := client.Goals().Update(ctx, cmd.GoalID, req)
	if err != nil {
		return err
	}

	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(os.Stdout, result)
	}
	if outfmt.IsPlain(ctx) {
		headers := []string{"ID", "NAME", "PERCENT_COMPLETE"}
		rows := [][]string{{result.ID, result.Name, fmt.Sprintf("%d", result.PercentCompleted)}}
		return outfmt.WritePlain(os.Stdout, headers, rows)
	}

	fmt.Fprintf(os.Stderr, "Updated goal\n\n")
	printGoalDetail(result)

	return nil
}

type GoalsDeleteCmd struct {
	GoalID string `arg:"" required:"" help:"Goal ID"`
}

func (cmd *GoalsDeleteCmd) Run(ctx context.Context) error {
	client, err := getClickUpClient(ctx)
	if err != nil {
		return err
	}

	if err := client.Goals().Delete(ctx, cmd.GoalID); err != nil {
		return err
	}

	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(os.Stdout, map[string]string{
			"status":  "success",
			"message": "Goal deleted",
			"goal_id": cmd.GoalID,
		})
	}
	if outfmt.IsPlain(ctx) {
		headers := []string{"STATUS", "GOAL_ID"}
		rows := [][]string{{"success", cmd.GoalID}}
		return outfmt.WritePlain(os.Stdout, headers, rows)
	}

	fmt.Fprintf(os.Stderr, "Goal %s deleted\n", cmd.GoalID)

	return nil
}

type GoalsAddKeyResultCmd struct {
	GoalID     string   `arg:"" required:"" help:"Goal ID"`
	Name       string   `required:"" help:"Key result name"`
	Type       string   `required:"" help:"Key result type: number, currency, boolean, percentage, automatic"`
	StepsStart int      `help:"Start value for number/currency/percentage types"`
	StepsEnd   int      `help:"End value for number/currency/percentage types"`
	Unit       string   `help:"Unit of measurement"`
	Owners     []string `help:"Owner user IDs"`
}

func (cmd *GoalsAddKeyResultCmd) Run(ctx context.Context) error {
	client, err := getClickUpClient(ctx)
	if err != nil {
		return err
	}

	req := clickup.CreateKeyResultRequest{
		Name:       cmd.Name,
		Type:       cmd.Type,
		StepsStart: cmd.StepsStart,
		StepsEnd:   cmd.StepsEnd,
		Unit:       cmd.Unit,
	}

	for _, owner := range cmd.Owners {
		var id int
		if _, parseErr := fmt.Sscanf(owner, "%d", &id); parseErr == nil {
			req.Owners = append(req.Owners, id)
		}
	}

	result, err := client.Goals().CreateKeyResult(ctx, cmd.GoalID, req)
	if err != nil {
		return err
	}

	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(os.Stdout, result)
	}
	if outfmt.IsPlain(ctx) {
		headers := []string{"ID", "NAME", "TYPE", "PROGRESS"}
		progress := fmt.Sprintf("%d/%d", result.StepsCurrent, result.StepsEnd)
		rows := [][]string{{result.ID, result.Name, result.Type, progress}}
		return outfmt.WritePlain(os.Stdout, headers, rows)
	}

	fmt.Fprintf(os.Stderr, "Created key result\n\n")
	printKeyResult(result)

	return nil
}

type GoalsUpdateKeyResultCmd struct {
	KeyResultID  string `arg:"" required:"" help:"Key result ID"`
	StepsCurrent int    `help:"Current progress value"`
	Note         string `help:"Note about the update"`
}

func (cmd *GoalsUpdateKeyResultCmd) Run(ctx context.Context) error {
	client, err := getClickUpClient(ctx)
	if err != nil {
		return err
	}

	req := clickup.EditKeyResultRequest{
		StepsCurrent: cmd.StepsCurrent,
		Note:         cmd.Note,
	}

	result, err := client.Goals().EditKeyResult(ctx, cmd.KeyResultID, req)
	if err != nil {
		return err
	}

	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(os.Stdout, result)
	}
	if outfmt.IsPlain(ctx) {
		headers := []string{"ID", "NAME", "STEPS_CURRENT", "STEPS_END"}
		rows := [][]string{{result.ID, result.Name, fmt.Sprintf("%d", result.StepsCurrent), fmt.Sprintf("%d", result.StepsEnd)}}
		return outfmt.WritePlain(os.Stdout, headers, rows)
	}

	fmt.Fprintf(os.Stderr, "Updated key result\n\n")
	printKeyResult(result)

	return nil
}

type GoalsDeleteKeyResultCmd struct {
	KeyResultID string `arg:"" required:"" help:"Key result ID"`
}

func (cmd *GoalsDeleteKeyResultCmd) Run(ctx context.Context) error {
	client, err := getClickUpClient(ctx)
	if err != nil {
		return err
	}

	if err := client.Goals().DeleteKeyResult(ctx, cmd.KeyResultID); err != nil {
		return err
	}

	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(os.Stdout, map[string]string{
			"status":        "success",
			"message":       "Key result deleted",
			"key_result_id": cmd.KeyResultID,
		})
	}
	if outfmt.IsPlain(ctx) {
		headers := []string{"STATUS", "KEY_RESULT_ID"}
		rows := [][]string{{"success", cmd.KeyResultID}}
		return outfmt.WritePlain(os.Stdout, headers, rows)
	}

	fmt.Fprintf(os.Stderr, "Key result %s deleted\n", cmd.KeyResultID)

	return nil
}

func printGoal(g *clickup.Goal) {
	fmt.Printf("ID: %s\n", g.ID)
	fmt.Printf("  Name: %s\n", g.Name)
	fmt.Printf("  Progress: %d%%\n", g.PercentCompleted)

	if g.DueDate != "" {
		fmt.Printf("  Due Date: %s\n", g.DueDate)
	}

	fmt.Println()
}

func printGoalDetail(g *clickup.Goal) {
	fmt.Printf("ID: %s\n", g.ID)
	fmt.Printf("Name: %s\n", g.Name)
	fmt.Printf("Progress: %d%%\n", g.PercentCompleted)

	if g.Description != "" {
		fmt.Printf("Description: %s\n", g.Description)
	}

	if g.DueDate != "" {
		fmt.Printf("Due Date: %s\n", g.DueDate)
	}

	if g.Color != "" {
		fmt.Printf("Color: %s\n", g.Color)
	}

	if len(g.Owners) > 0 {
		names := make([]string, 0, len(g.Owners))
		for _, o := range g.Owners {
			names = append(names, o.Username)
		}
		fmt.Printf("Owners: %s\n", strings.Join(names, ", "))
	}

	if len(g.KeyResults) > 0 {
		fmt.Println("Key Results:")
		for _, kr := range g.KeyResults {
			fmt.Printf("  - %s: %d/%d (%s)\n", kr.Name, kr.StepsCurrent, kr.StepsEnd, kr.Type)
		}
	}
}

func printKeyResult(kr *clickup.KeyResult) {
	fmt.Printf("ID: %s\n", kr.ID)
	fmt.Printf("Name: %s\n", kr.Name)
	fmt.Printf("Type: %s\n", kr.Type)
	fmt.Printf("Progress: %d / %d\n", kr.StepsCurrent, kr.StepsEnd)

	if kr.Unit != "" {
		fmt.Printf("Unit: %s\n", kr.Unit)
	}

	if kr.Note != "" {
		fmt.Printf("Note: %s\n", kr.Note)
	}
}

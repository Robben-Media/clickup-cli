package cmd

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/builtbyrobben/clickup-cli/internal/clickup"
	"github.com/builtbyrobben/clickup-cli/internal/outfmt"
)

type GoalsCmd struct {
	List            GoalsListCmd            `cmd:"" help:"List goals"`
	Get             GoalsGetCmd             `cmd:"" help:"Get goal details"`
	Create          GoalsCreateCmd          `cmd:"" help:"Create a goal"`
	Update          GoalsUpdateCmd          `cmd:"" help:"Update a goal"`
	Delete          GoalsDeleteCmd          `cmd:"" help:"Delete a goal"`
	AddKeyResult    GoalsAddKeyResultCmd    `cmd:"" help:"Add key result to goal"`
	UpdateKeyResult GoalsUpdateKeyResultCmd `cmd:"" help:"Update key result progress"`
	DeleteKeyResult GoalsDeleteKeyResultCmd `cmd:"" help:"Delete a key result"`
}

type GoalsListCmd struct {
	TeamID           string `arg:"" required:"" help:"Workspace/Team ID"`
	IncludeCompleted bool   `help:"Include completed goals"`
}

func (cmd *GoalsListCmd) Run(ctx context.Context) error {
	client, err := getClickUpClient(ctx)
	if err != nil {
		return err
	}

	result, err := client.Goals().List(ctx, cmd.TeamID, cmd.IncludeCompleted)
	if err != nil {
		return err
	}

	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(os.Stdout, result)
	}

	if outfmt.IsPlain(ctx) {
		headers := []string{"ID", "NAME", "PERCENT_COMPLETE", "DUE_DATE", "KEY_RESULT_COUNT"}
		var rows [][]string

		for _, goal := range result.Goals {
			rows = append(rows, []string{
				goal.ID,
				goal.Name,
				strconv.Itoa(goal.PercentCompleted),
				goal.DueDate,
				strconv.Itoa(len(goal.KeyResults)),
			})
		}

		return outfmt.WritePlain(os.Stdout, headers, rows)
	}

	// Human-readable output
	if len(result.Goals) == 0 {
		fmt.Fprintln(os.Stderr, "No goals found")

		return nil
	}

	fmt.Fprintf(os.Stderr, "Found %d goals:\n\n", len(result.Goals))

	for _, goal := range result.Goals {
		fmt.Fprintf(os.Stderr, "  %s: %s (%d%% complete)\n", goal.ID, goal.Name, goal.PercentCompleted)

		if goal.DueDate != "" {
			fmt.Fprintf(os.Stderr, "    Due: %s\n", goal.DueDate)
		}

		if goal.Description != "" {
			fmt.Fprintf(os.Stderr, "    Description: %s\n", goal.Description)
		}

		if len(goal.KeyResults) > 0 {
			fmt.Fprintf(os.Stderr, "    Key Results: %d\n", len(goal.KeyResults))
		}
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
		headers := []string{"ID", "NAME", "PERCENT_COMPLETE", "DUE_DATE", "KEY_RESULT_COUNT"}
		rows := [][]string{{
			result.ID,
			result.Name,
			strconv.Itoa(result.PercentCompleted),
			result.DueDate,
			strconv.Itoa(len(result.KeyResults)),
		}}

		return outfmt.WritePlain(os.Stdout, headers, rows)
	}

	// Human-readable output
	fmt.Fprintf(os.Stderr, "Goal: %s\n\n", result.Name)
	fmt.Printf("  ID: %s\n", result.ID)
	fmt.Printf("  Progress: %d%%\n", result.PercentCompleted)

	if result.DueDate != "" {
		fmt.Printf("  Due: %s\n", result.DueDate)
	}

	if result.Description != "" {
		fmt.Printf("  Description: %s\n", result.Description)
	}

	if result.Color != "" {
		fmt.Printf("  Color: %s\n", result.Color)
	}

	if len(result.Owners) > 0 {
		var owners []string

		for _, o := range result.Owners {
			owners = append(owners, o.Username)
		}

		fmt.Printf("  Owners: %s\n", strings.Join(owners, ", "))
	}

	if len(result.KeyResults) > 0 {
		fmt.Printf("\n  Key Results (%d):\n", len(result.KeyResults))

		for _, kr := range result.KeyResults {
			fmt.Printf("    - %s [%s]\n", kr.Name, kr.Type)
			fmt.Printf("      Progress: %d / %d (%d%%)\n", kr.StepsCurrent, kr.StepsEnd, calculateProgress(kr.StepsCurrent, kr.StepsEnd, kr.StepsStart))

			if kr.Unit != "" {
				fmt.Printf("      Unit: %s\n", kr.Unit)
			}

			if kr.Note != "" {
				fmt.Printf("      Note: %s\n", kr.Note)
			}
		}
	}

	return nil
}

func calculateProgress(current, end, start int) int {
	if end == start {
		return 0
	}

	return (current - start) * 100 / (end - start)
}

type GoalsCreateCmd struct {
	TeamID      string `arg:"" required:"" help:"Workspace/Team ID"`
	Name        string `arg:"" required:"" help:"Goal name"`
	DueDate     int64  `help:"Due date in milliseconds"`
	Description string `help:"Goal description"`
	Owners      string `help:"Comma-separated list of owner user IDs"`
	Color       string `name:"goal-color" help:"Goal color (hex)"`
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

	if cmd.Owners != "" {
		ownerStrs := strings.Split(cmd.Owners, ",")

		owners := make([]int, 0, len(ownerStrs))

		for _, o := range ownerStrs {
			id, parseErr := strconv.Atoi(strings.TrimSpace(o))
			if parseErr != nil {
				return fmt.Errorf("invalid owner ID %q: %w", o, parseErr)
			}

			owners = append(owners, id)
		}

		req.Owners = owners
	}

	result, err := client.Goals().Create(ctx, cmd.TeamID, req)
	if err != nil {
		return err
	}

	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(os.Stdout, result)
	}

	if outfmt.IsPlain(ctx) {
		headers := []string{"ID", "NAME", "PERCENT_COMPLETE"}
		rows := [][]string{{result.ID, result.Name, strconv.Itoa(result.PercentCompleted)}}

		return outfmt.WritePlain(os.Stdout, headers, rows)
	}

	fmt.Fprintf(os.Stderr, "Goal created\n")
	fmt.Printf("  ID: %s\n", result.ID)
	fmt.Printf("  Name: %s\n", result.Name)

	if result.DueDate != "" {
		fmt.Printf("  Due: %s\n", result.DueDate)
	}

	return nil
}

type GoalsUpdateCmd struct {
	GoalID       string `arg:"" required:"" help:"Goal ID"`
	Name         string `help:"New goal name"`
	Description  string `help:"New goal description"`
	DueDate      int64  `help:"New due date in milliseconds"`
	Color        string `name:"goal-color" help:"New goal color (hex)"`
	AddOwners    string `help:"Comma-separated list of owner IDs to add"`
	RemoveOwners string `help:"Comma-separated list of owner IDs to remove"`
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

	if cmd.AddOwners != "" {
		ownerStrs := strings.Split(cmd.AddOwners, ",")

		owners := make([]int, 0, len(ownerStrs))

		for _, o := range ownerStrs {
			id, parseErr := strconv.Atoi(strings.TrimSpace(o))
			if parseErr != nil {
				return fmt.Errorf("invalid add_owner ID %q: %w", o, parseErr)
			}

			owners = append(owners, id)
		}

		req.AddOwners = owners
	}

	if cmd.RemoveOwners != "" {
		ownerStrs := strings.Split(cmd.RemoveOwners, ",")

		owners := make([]int, 0, len(ownerStrs))

		for _, o := range ownerStrs {
			id, parseErr := strconv.Atoi(strings.TrimSpace(o))
			if parseErr != nil {
				return fmt.Errorf("invalid remove_owner ID %q: %w", o, parseErr)
			}

			owners = append(owners, id)
		}

		req.RemOwners = owners
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
		rows := [][]string{{result.ID, result.Name, strconv.Itoa(result.PercentCompleted)}}

		return outfmt.WritePlain(os.Stdout, headers, rows)
	}

	fmt.Fprintf(os.Stderr, "Goal updated\n")
	fmt.Printf("  ID: %s\n", result.ID)
	fmt.Printf("  Name: %s\n", result.Name)
	fmt.Printf("  Progress: %d%%\n", result.PercentCompleted)

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

	if !forceEnabled(ctx) {
		fmt.Fprintf(os.Stderr, "Warning: This will permanently delete goal %s and all its key results\n", cmd.GoalID)
		fmt.Fprint(os.Stderr, "Use --force to confirm deletion\n")

		return fmt.Errorf("operation cancelled: use --force to confirm")
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
	GoalID     string `arg:"" required:"" help:"Goal ID"`
	Name       string `required:"" help:"Key result name"`
	Type       string `required:"" help:"Key result type (number, currency, boolean, percentage, automatic)"`
	StepsStart int    `help:"Start value for number/currency/percentage types"`
	StepsEnd   int    `help:"End value for number/currency/percentage types"`
	Unit       string `help:"Unit of measurement"`
	Owners     string `help:"Comma-separated list of owner user IDs"`
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

	if cmd.Owners != "" {
		ownerStrs := strings.Split(cmd.Owners, ",")

		owners := make([]int, 0, len(ownerStrs))

		for _, o := range ownerStrs {
			id, parseErr := strconv.Atoi(strings.TrimSpace(o))
			if parseErr != nil {
				return fmt.Errorf("invalid owner ID %q: %w", o, parseErr)
			}

			owners = append(owners, id)
		}

		req.Owners = owners
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
		rows := [][]string{{
			result.ID,
			result.Name,
			result.Type,
			fmt.Sprintf("%d/%d", result.StepsCurrent, result.StepsEnd),
		}}

		return outfmt.WritePlain(os.Stdout, headers, rows)
	}

	fmt.Fprintf(os.Stderr, "Key result created\n")
	fmt.Printf("  ID: %s\n", result.ID)
	fmt.Printf("  Name: %s\n", result.Name)
	fmt.Printf("  Type: %s\n", result.Type)
	fmt.Printf("  Range: %d - %d\n", result.StepsStart, result.StepsEnd)

	return nil
}

type GoalsUpdateKeyResultCmd struct {
	KeyResultID  string `arg:"" required:"" help:"Key result ID"`
	StepsCurrent int    `help:"Current progress value"`
	Note         string `help:"Note about progress"`
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

	result, err := client.Goals().UpdateKeyResult(ctx, cmd.KeyResultID, req)
	if err != nil {
		return err
	}

	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(os.Stdout, result)
	}

	if outfmt.IsPlain(ctx) {
		headers := []string{"ID", "NAME", "TYPE", "PROGRESS"}
		rows := [][]string{{
			result.ID,
			result.Name,
			result.Type,
			fmt.Sprintf("%d/%d", result.StepsCurrent, result.StepsEnd),
		}}

		return outfmt.WritePlain(os.Stdout, headers, rows)
	}

	fmt.Fprintf(os.Stderr, "Key result updated\n")
	fmt.Printf("  ID: %s\n", result.ID)
	fmt.Printf("  Name: %s\n", result.Name)
	fmt.Printf("  Progress: %d / %d\n", result.StepsCurrent, result.StepsEnd)

	if result.Note != "" {
		fmt.Printf("  Note: %s\n", result.Note)
	}

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

	if !forceEnabled(ctx) {
		fmt.Fprintf(os.Stderr, "Warning: This will permanently delete key result %s\n", cmd.KeyResultID)
		fmt.Fprint(os.Stderr, "Use --force to confirm deletion\n")

		return fmt.Errorf("operation cancelled: use --force to confirm")
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

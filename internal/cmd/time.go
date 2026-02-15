package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/builtbyrobben/clickup-cli/internal/outfmt"
)

type TimeCmd struct {
	Log  TimeLogCmd  `cmd:"" help:"Log time to a task"`
	List TimeListCmd `cmd:"" help:"List time entries for a task"`
}

type TimeLogCmd struct {
	TaskID     string `arg:"" required:"" help:"Task ID"`
	DurationMs int64  `arg:"" required:"" help:"Duration in milliseconds"`
}

func (cmd *TimeLogCmd) Run(ctx context.Context) error {
	client, err := getClickUpClient()
	if err != nil {
		return err
	}

	teamID, err := getTeamID()
	if err != nil {
		return err
	}

	result, err := client.Time().Log(ctx, teamID, cmd.TaskID, cmd.DurationMs)
	if err != nil {
		return err
	}

	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(os.Stdout, result)
	}

	fmt.Fprintf(os.Stderr, "Time logged (ID: %s)\n", result.ID)
	fmt.Printf("Duration: %s ms\n", result.Duration)

	return nil
}

type TimeListCmd struct {
	TaskID string `arg:"" required:"" help:"Task ID"`
}

func (cmd *TimeListCmd) Run(ctx context.Context) error {
	client, err := getClickUpClient()
	if err != nil {
		return err
	}

	teamID, err := getTeamID()
	if err != nil {
		return err
	}

	result, err := client.Time().List(ctx, teamID, cmd.TaskID)
	if err != nil {
		return err
	}

	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(os.Stdout, result)
	}

	if len(result.Data) == 0 {
		fmt.Fprintln(os.Stderr, "No time entries found")
		return nil
	}

	fmt.Fprintf(os.Stderr, "Found %d time entries\n\n", len(result.Data))

	for _, entry := range result.Data {
		fmt.Printf("ID: %s\n", entry.ID)
		fmt.Printf("  Duration: %s ms\n", entry.Duration)

		if entry.Start != "" {
			fmt.Printf("  Start: %s\n", entry.Start)
		}

		if entry.End != "" {
			fmt.Printf("  End: %s\n", entry.End)
		}

		fmt.Println()
	}

	return nil
}

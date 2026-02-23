package cmd

import (
	"context"
	"fmt"
	"os"
	"strconv"

	"github.com/builtbyrobben/clickup-cli/internal/clickup"
	"github.com/builtbyrobben/clickup-cli/internal/outfmt"
)

type TimeLegacyCmd struct {
	List   TimeLegacyListCmd   `cmd:"" help:"List tracked time for a task"`
	Track  TimeLegacyTrackCmd  `cmd:"" help:"Track time on a task"`
	Update TimeLegacyUpdateCmd `cmd:"" help:"Update tracked time"`
	Delete TimeLegacyDeleteCmd `cmd:"" help:"Delete tracked time"`
}

type TimeLegacyListCmd struct {
	TaskID        string `arg:"" required:"" help:"Task ID"`
	CustomTaskIDs bool   `help:"Treat task ID as custom task ID"`
	TeamID        string `help:"Team ID (required when using custom task IDs)"`
}

func (cmd *TimeLegacyListCmd) Run(ctx context.Context) error {
	client, err := getClickUpClient(ctx)
	if err != nil {
		return err
	}

	result, err := client.LegacyTime().List(ctx, cmd.TaskID, cmd.CustomTaskIDs, cmd.TeamID)
	if err != nil {
		return err
	}

	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(os.Stdout, result)
	}

	if outfmt.IsPlain(ctx) {
		headers := []string{"ID", "START", "END", "TIME_MS", "SOURCE"}
		var rows [][]string
		for _, interval := range result.Data {
			rows = append(rows, []string{
				interval.ID,
				strconv.FormatInt(interval.Start, 10),
				strconv.FormatInt(interval.End, 10),
				strconv.FormatInt(interval.Time, 10),
				interval.Source,
			})
		}

		return outfmt.WritePlain(os.Stdout, headers, rows)
	}

	if len(result.Data) == 0 {
		fmt.Fprintln(os.Stderr, "No tracked time found")
		return nil
	}

	fmt.Fprintf(os.Stderr, "Tracked time for task %s\n\n", cmd.TaskID)

	for _, interval := range result.Data {
		printLegacyTimeInterval(&interval)
	}

	return nil
}

type TimeLegacyTrackCmd struct {
	TaskID string `arg:"" required:"" help:"Task ID"`
	Time   int64  `required:"" help:"Duration in milliseconds"`
	Start  int64  `help:"Start timestamp in milliseconds"`
	End    int64  `help:"End timestamp in milliseconds"`
}

func (cmd *TimeLegacyTrackCmd) Run(ctx context.Context) error {
	client, err := getClickUpClient(ctx)
	if err != nil {
		return err
	}

	req := clickup.TrackTimeRequest{
		Time:  cmd.Time,
		Start: cmd.Start,
		End:   cmd.End,
	}

	result, err := client.LegacyTime().Track(ctx, cmd.TaskID, req)
	if err != nil {
		return err
	}

	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(os.Stdout, result)
	}

	if outfmt.IsPlain(ctx) {
		headers := []string{"ID"}
		rows := [][]string{{result.ID}}

		return outfmt.WritePlain(os.Stdout, headers, rows)
	}

	fmt.Fprintf(os.Stderr, "Time tracked successfully\n")
	fmt.Printf("Interval ID: %s\n", result.ID)

	return nil
}

type TimeLegacyUpdateCmd struct {
	TaskID     string `arg:"" required:"" help:"Task ID"`
	IntervalID string `arg:"" required:"" help:"Interval ID"`
	Time       int64  `help:"Duration in milliseconds"`
	Start      int64  `help:"Start timestamp in milliseconds"`
	End        int64  `help:"End timestamp in milliseconds"`
}

func (cmd *TimeLegacyUpdateCmd) Run(ctx context.Context) error {
	client, err := getClickUpClient(ctx)
	if err != nil {
		return err
	}

	req := clickup.EditTimeRequest{
		Time:  cmd.Time,
		Start: cmd.Start,
		End:   cmd.End,
	}

	if err := client.LegacyTime().Edit(ctx, cmd.TaskID, cmd.IntervalID, req); err != nil {
		return err
	}

	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(os.Stdout, map[string]string{"status": "success", "interval_id": cmd.IntervalID})
	}

	if outfmt.IsPlain(ctx) {
		headers := []string{"STATUS", "INTERVAL_ID"}
		rows := [][]string{{"success", cmd.IntervalID}}

		return outfmt.WritePlain(os.Stdout, headers, rows)
	}

	fmt.Fprintf(os.Stderr, "Time interval updated\n")
	fmt.Printf("Interval ID: %s\n", cmd.IntervalID)

	return nil
}

type TimeLegacyDeleteCmd struct {
	TaskID     string `arg:"" required:"" help:"Task ID"`
	IntervalID string `arg:"" required:"" help:"Interval ID"`
}

func (cmd *TimeLegacyDeleteCmd) Run(ctx context.Context) error {
	client, err := getClickUpClient(ctx)
	if err != nil {
		return err
	}

	if err := client.LegacyTime().Delete(ctx, cmd.TaskID, cmd.IntervalID); err != nil {
		return err
	}

	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(os.Stdout, map[string]string{"status": "success", "message": "Time interval deleted"})
	}

	if outfmt.IsPlain(ctx) {
		headers := []string{"STATUS", "INTERVAL_ID"}
		rows := [][]string{{"success", cmd.IntervalID}}

		return outfmt.WritePlain(os.Stdout, headers, rows)
	}

	fmt.Fprintf(os.Stderr, "Time interval %s deleted\n", cmd.IntervalID)

	return nil
}

func printLegacyTimeInterval(interval *clickup.LegacyTimeInterval) {
	fmt.Printf("ID: %s\n", interval.ID)
	fmt.Printf("  Duration: %s\n", formatLegacyDuration(interval.Time))
	fmt.Printf("  Start: %s\n", formatLegacyTimestamp(interval.Start))
	fmt.Printf("  End: %s\n", formatLegacyTimestamp(interval.End))

	if interval.Source != "" {
		fmt.Printf("  Source: %s\n", interval.Source)
	}

	fmt.Println()
}

func formatLegacyDuration(ms int64) string {
	seconds := ms / 1000
	minutes := seconds / 60
	hours := minutes / 60
	minutes %= 60

	return fmt.Sprintf("%dh %dm", hours, minutes)
}

func formatLegacyTimestamp(ms int64) string {
	seconds := ms / 1000
	return fmt.Sprintf("<timestamp:%d>", seconds)
}

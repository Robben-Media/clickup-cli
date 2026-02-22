package cmd

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/builtbyrobben/clickup-cli/internal/clickup"
	"github.com/builtbyrobben/clickup-cli/internal/outfmt"
)

type TimeCmd struct {
	Log        TimeLogCmd        `cmd:"" help:"Log time to a task"`
	List       TimeListCmd       `cmd:"" help:"List time entries for a task"`
	Get        TimeGetCmd        `cmd:"" help:"Get a single time entry"`
	Current    TimeCurrentCmd    `cmd:"" help:"Get currently running timer"`
	Start      TimeStartCmd      `cmd:"" help:"Start a timer"`
	Stop       TimeStopCmd       `cmd:"" help:"Stop the running timer"`
	Update     TimeUpdateCmd     `cmd:"" help:"Update a time entry"`
	Delete     TimeDeleteCmd     `cmd:"" help:"Delete a time entry"`
	History    TimeHistoryCmd    `cmd:"" help:"Get change history for a time entry"`
	Tags       TimeTagsCmd       `cmd:"" help:"List all time entry tags"`
	AddTags    TimeAddTagsCmd    `cmd:"" help:"Add tags to time entries"`
	RemoveTags TimeRemoveTagsCmd `cmd:"" help:"Remove tags from time entries"`
	RenameTag  TimeRenameTagCmd  `cmd:"" help:"Rename a time entry tag"`
}

type TimeLogCmd struct {
	TaskID     string `arg:"" required:"" help:"Task ID"`
	DurationMs int64  `arg:"" required:"" help:"Duration in milliseconds"`
	Start      string `help:"Start time as Unix ms timestamp, or 'now' (default: now)" default:"now"`
}

func (cmd *TimeLogCmd) Run(ctx context.Context) error {
	client, err := getClickUpClient(ctx)
	if err != nil {
		return err
	}

	teamID, err := getTeamID()
	if err != nil {
		return err
	}

	var startMs int64
	if cmd.Start == "" || cmd.Start == "now" {
		startMs = time.Now().UnixMilli()
	} else {
		startMs, err = strconv.ParseInt(cmd.Start, 10, 64)
		if err != nil {
			return fmt.Errorf("invalid start timestamp %q: %w", cmd.Start, err)
		}
	}

	result, err := client.Time().Log(ctx, teamID, cmd.TaskID, cmd.DurationMs, startMs)
	if err != nil {
		return err
	}

	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(os.Stdout, result)
	}
	if outfmt.IsPlain(ctx) {
		headers := []string{"ID", "DURATION", "START", "END"}
		rows := [][]string{{result.ID.String(), result.Duration.String(), result.Start.String(), result.End.String()}}
		return outfmt.WritePlain(os.Stdout, headers, rows)
	}

	fmt.Fprintf(os.Stderr, "Time logged (ID: %s)\n", result.ID)
	fmt.Printf("Duration: %s ms\n", result.Duration)

	return nil
}

type TimeListCmd struct {
	TaskID string `arg:"" required:"" help:"Task ID"`
}

func (cmd *TimeListCmd) Run(ctx context.Context) error {
	client, err := getClickUpClient(ctx)
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
	if outfmt.IsPlain(ctx) {
		headers := []string{"ID", "DURATION", "START", "END"}
		var rows [][]string
		for _, entry := range result.Data {
			rows = append(rows, []string{entry.ID.String(), entry.Duration.String(), entry.Start.String(), entry.End.String()})
		}
		return outfmt.WritePlain(os.Stdout, headers, rows)
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

type TimeGetCmd struct {
	EntryID string `arg:"" required:"" help:"Time entry ID"`
}

func (cmd *TimeGetCmd) Run(ctx context.Context) error {
	client, err := getClickUpClient(ctx)
	if err != nil {
		return err
	}

	teamID, err := getTeamID()
	if err != nil {
		return err
	}

	result, err := client.Time().Get(ctx, teamID, cmd.EntryID)
	if err != nil {
		return err
	}

	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(os.Stdout, result)
	}
	if outfmt.IsPlain(ctx) {
		headers := []string{"ID", "TASK", "USER", "DURATION", "START", "END", "DESCRIPTION"}
		rows := [][]string{{
			result.ID.String(),
			result.Task.ID,
			result.User.Username,
			result.Duration.String(),
			result.Start.String(),
			result.End.String(),
			result.Description,
		}}
		return outfmt.WritePlain(os.Stdout, headers, rows)
	}

	fmt.Printf("Time Entry %s\n", result.ID)

	if result.Task.ID != "" {
		fmt.Printf("  Task: %s (%s)\n", result.Task.Name, result.Task.ID)
	}

	fmt.Printf("  User: %s\n", result.User.Username)
	fmt.Printf("  Duration: %s ms\n", result.Duration)
	fmt.Printf("  Start: %s\n", result.Start)
	fmt.Printf("  End: %s\n", result.End)
	fmt.Printf("  Billable: %v\n", result.Billable)

	if result.Description != "" {
		fmt.Printf("  Description: %s\n", result.Description)
	}

	if len(result.Tags) > 0 {
		tagNames := make([]string, len(result.Tags))
		for i, tag := range result.Tags {
			tagNames[i] = tag.Name
		}

		fmt.Printf("  Tags: %s\n", strings.Join(tagNames, ", "))
	}

	return nil
}

type TimeCurrentCmd struct{}

func (cmd *TimeCurrentCmd) Run(ctx context.Context) error {
	client, err := getClickUpClient(ctx)
	if err != nil {
		return err
	}

	teamID, err := getTeamID()
	if err != nil {
		return err
	}

	result, err := client.Time().Current(ctx, teamID)
	if err != nil {
		return err
	}

	if result == nil || result.ID == "" {
		if outfmt.IsJSON(ctx) {
			return outfmt.WriteJSON(os.Stdout, map[string]any{"running": false})
		}
		if outfmt.IsPlain(ctx) {
			headers := []string{"RUNNING"}
			return outfmt.WritePlain(os.Stdout, headers, [][]string{{"false"}})
		}

		fmt.Fprintln(os.Stderr, "No timer running")

		return nil
	}

	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(os.Stdout, result)
	}
	if outfmt.IsPlain(ctx) {
		headers := []string{"ID", "TASK", "START", "DESCRIPTION"}
		rows := [][]string{{
			result.ID.String(),
			result.Task.ID,
			result.Start.String(),
			result.Description,
		}}
		return outfmt.WritePlain(os.Stdout, headers, rows)
	}

	fmt.Fprintln(os.Stderr, "Running Timer")
	fmt.Printf("  ID: %s\n", result.ID)

	if result.Task.ID != "" {
		fmt.Printf("  Task: %s (%s)\n", result.Task.Name, result.Task.ID)
	}

	fmt.Printf("  Start: %s\n", result.Start)

	if result.Description != "" {
		fmt.Printf("  Description: %s\n", result.Description)
	}

	return nil
}

type TimeStartCmd struct {
	TaskID      string   `help:"Task ID to associate with the timer"`
	Description string   `help:"Description for the time entry"`
	Billable    bool     `help:"Mark as billable"`
	Tags        []string `help:"Tags to apply (comma-separated or multiple flags)"`
}

func (cmd *TimeStartCmd) Run(ctx context.Context) error {
	client, err := getClickUpClient(ctx)
	if err != nil {
		return err
	}

	teamID, err := getTeamID()
	if err != nil {
		return err
	}

	req := clickup.StartTimeEntryRequest{
		TaskID:      cmd.TaskID,
		Description: cmd.Description,
		Billable:    cmd.Billable,
	}

	if len(cmd.Tags) > 0 {
		req.Tags = make([]clickup.Tag, len(cmd.Tags))
		for i, tag := range cmd.Tags {
			req.Tags[i] = clickup.Tag{Name: tag}
		}
	}

	result, err := client.Time().Start(ctx, teamID, req)
	if err != nil {
		return err
	}

	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(os.Stdout, result)
	}
	if outfmt.IsPlain(ctx) {
		headers := []string{"ID", "TASK", "START", "DESCRIPTION"}
		rows := [][]string{{
			result.ID.String(),
			result.Task.ID,
			result.Start.String(),
			result.Description,
		}}
		return outfmt.WritePlain(os.Stdout, headers, rows)
	}

	fmt.Fprintf(os.Stderr, "Timer started (ID: %s)\n", result.ID)

	if result.Task.ID != "" {
		fmt.Fprintf(os.Stderr, "Task: %s (%s)\n", result.Task.Name, result.Task.ID)
	}

	return nil
}

type TimeStopCmd struct{}

func (cmd *TimeStopCmd) Run(ctx context.Context) error {
	client, err := getClickUpClient(ctx)
	if err != nil {
		return err
	}

	teamID, err := getTeamID()
	if err != nil {
		return err
	}

	result, err := client.Time().Stop(ctx, teamID)
	if err != nil {
		return err
	}

	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(os.Stdout, result)
	}
	if outfmt.IsPlain(ctx) {
		headers := []string{"ID", "TASK", "DURATION", "START", "END"}
		rows := [][]string{{
			result.ID.String(),
			result.Task.ID,
			result.Duration.String(),
			result.Start.String(),
			result.End.String(),
		}}
		return outfmt.WritePlain(os.Stdout, headers, rows)
	}

	fmt.Fprintf(os.Stderr, "Timer stopped (ID: %s)\n", result.ID)
	fmt.Printf("Duration: %s ms\n", result.Duration)

	return nil
}

type TimeUpdateCmd struct {
	EntryID     string `arg:"" required:"" help:"Time entry ID"`
	Description string `help:"New description"`
	Duration    int64  `help:"New duration in milliseconds"`
	Start       int64  `help:"New start time in milliseconds"`
	End         int64  `help:"New end time in milliseconds"`
	Billable    *bool  `help:"Set billable status"`
	TagAction   string `help:"Tag action: 'add' or 'remove'"`
	Tags        string `help:"Comma-separated tag names"`
}

func (cmd *TimeUpdateCmd) Run(ctx context.Context) error {
	client, err := getClickUpClient(ctx)
	if err != nil {
		return err
	}

	teamID, err := getTeamID()
	if err != nil {
		return err
	}

	req := clickup.UpdateTimeEntryRequest{
		Description: cmd.Description,
		Duration:    cmd.Duration,
		Start:       cmd.Start,
		End:         cmd.End,
		Billable:    cmd.Billable,
		TagAction:   cmd.TagAction,
	}

	if cmd.Tags != "" {
		tagNames := strings.Split(cmd.Tags, ",")
		req.Tags = make([]clickup.Tag, len(tagNames))
		for i, name := range tagNames {
			req.Tags[i] = clickup.Tag{Name: strings.TrimSpace(name)}
		}
	}

	result, err := client.Time().Update(ctx, teamID, cmd.EntryID, req)
	if err != nil {
		return err
	}

	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(os.Stdout, result)
	}
	if outfmt.IsPlain(ctx) {
		headers := []string{"ID", "DURATION", "DESCRIPTION"}
		rows := [][]string{{
			result.ID.String(),
			result.Duration.String(),
			result.Description,
		}}
		return outfmt.WritePlain(os.Stdout, headers, rows)
	}

	fmt.Fprintf(os.Stderr, "Time entry updated (ID: %s)\n", result.ID)

	return nil
}

type TimeDeleteCmd struct {
	EntryID string `arg:"" required:"" help:"Time entry ID"`
}

func (cmd *TimeDeleteCmd) Run(ctx context.Context) error {
	client, err := getClickUpClient(ctx)
	if err != nil {
		return err
	}

	teamID, err := getTeamID()
	if err != nil {
		return err
	}

	if err := client.Time().Delete(ctx, teamID, cmd.EntryID); err != nil {
		return err
	}

	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(os.Stdout, map[string]string{"status": "success", "entry_id": cmd.EntryID})
	}
	if outfmt.IsPlain(ctx) {
		headers := []string{"STATUS", "ENTRY_ID"}
		return outfmt.WritePlain(os.Stdout, headers, [][]string{{"success", cmd.EntryID}})
	}

	fmt.Fprintf(os.Stderr, "Time entry %s deleted\n", cmd.EntryID)

	return nil
}

type TimeHistoryCmd struct {
	EntryID string `arg:"" required:"" help:"Time entry ID"`
}

func (cmd *TimeHistoryCmd) Run(ctx context.Context) error {
	client, err := getClickUpClient(ctx)
	if err != nil {
		return err
	}

	teamID, err := getTeamID()
	if err != nil {
		return err
	}

	result, err := client.Time().History(ctx, teamID, cmd.EntryID)
	if err != nil {
		return err
	}

	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(os.Stdout, result)
	}
	if outfmt.IsPlain(ctx) {
		headers := []string{"ID", "FIELD", "BEFORE", "AFTER", "DATE", "USER"}
		var rows [][]string
		for _, item := range result.Data {
			rows = append(rows, []string{
				item.ID,
				item.Field,
				item.Before,
				item.After,
				item.Date,
				item.User.Username,
			})
		}
		return outfmt.WritePlain(os.Stdout, headers, rows)
	}

	if len(result.Data) == 0 {
		fmt.Fprintln(os.Stderr, "No history found")
		return nil
	}

	fmt.Fprintf(os.Stderr, "History for time entry %s\n\n", cmd.EntryID)

	for _, item := range result.Data {
		fmt.Printf("  %s: %s -> %s (by %s)\n", item.Field, item.Before, item.After, item.User.Username)
	}

	return nil
}

type TimeTagsCmd struct{}

func (cmd *TimeTagsCmd) Run(ctx context.Context) error {
	client, err := getClickUpClient(ctx)
	if err != nil {
		return err
	}

	teamID, err := getTeamID()
	if err != nil {
		return err
	}

	result, err := client.Time().ListTags(ctx, teamID)
	if err != nil {
		return err
	}

	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(os.Stdout, result)
	}
	if outfmt.IsPlain(ctx) {
		headers := []string{"NAME"}
		var rows [][]string
		for _, tag := range result.Data {
			rows = append(rows, []string{tag.Name})
		}
		return outfmt.WritePlain(os.Stdout, headers, rows)
	}

	if len(result.Data) == 0 {
		fmt.Fprintln(os.Stderr, "No tags found")
		return nil
	}

	fmt.Fprintf(os.Stderr, "Found %d tags\n\n", len(result.Data))

	for _, tag := range result.Data {
		fmt.Printf("  %s\n", tag.Name)
	}

	return nil
}

type TimeAddTagsCmd struct {
	EntryIDs string `required:"" help:"Comma-separated time entry IDs"`
	Tags     string `required:"" help:"Comma-separated tag names"`
}

func (cmd *TimeAddTagsCmd) Run(ctx context.Context) error {
	client, err := getClickUpClient(ctx)
	if err != nil {
		return err
	}

	teamID, err := getTeamID()
	if err != nil {
		return err
	}

	entryIDs := strings.Split(cmd.EntryIDs, ",")
	tagNames := strings.Split(cmd.Tags, ",")

	tags := make([]clickup.Tag, len(tagNames))
	for i, name := range tagNames {
		tags[i] = clickup.Tag{Name: strings.TrimSpace(name)}
	}

	req := clickup.TimeEntryTagsRequest{
		TimeEntryIDs: entryIDs,
		Tags:         tags,
	}

	if err := client.Time().AddTags(ctx, teamID, req); err != nil {
		return err
	}

	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(os.Stdout, map[string]string{"status": "success"})
	}
	if outfmt.IsPlain(ctx) {
		headers := []string{"STATUS"}
		return outfmt.WritePlain(os.Stdout, headers, [][]string{{"success"}})
	}

	fmt.Fprintf(os.Stderr, "Tags added to %d time entries\n", len(entryIDs))

	return nil
}

type TimeRemoveTagsCmd struct {
	EntryIDs string `required:"" help:"Comma-separated time entry IDs"`
	Tags     string `required:"" help:"Comma-separated tag names"`
}

func (cmd *TimeRemoveTagsCmd) Run(ctx context.Context) error {
	client, err := getClickUpClient(ctx)
	if err != nil {
		return err
	}

	teamID, err := getTeamID()
	if err != nil {
		return err
	}

	entryIDs := strings.Split(cmd.EntryIDs, ",")
	tagNames := strings.Split(cmd.Tags, ",")

	tags := make([]clickup.Tag, len(tagNames))
	for i, name := range tagNames {
		tags[i] = clickup.Tag{Name: strings.TrimSpace(name)}
	}

	req := clickup.TimeEntryTagsRequest{
		TimeEntryIDs: entryIDs,
		Tags:         tags,
	}

	if err := client.Time().RemoveTags(ctx, teamID, req); err != nil {
		return err
	}

	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(os.Stdout, map[string]string{"status": "success"})
	}
	if outfmt.IsPlain(ctx) {
		headers := []string{"STATUS"}
		return outfmt.WritePlain(os.Stdout, headers, [][]string{{"success"}})
	}

	fmt.Fprintf(os.Stderr, "Tags removed from %d time entries\n", len(entryIDs))

	return nil
}

type TimeRenameTagCmd struct {
	OldName string `required:"" help:"Current tag name"`
	NewName string `required:"" help:"New tag name"`
}

func (cmd *TimeRenameTagCmd) Run(ctx context.Context) error {
	client, err := getClickUpClient(ctx)
	if err != nil {
		return err
	}

	teamID, err := getTeamID()
	if err != nil {
		return err
	}

	req := clickup.RenameTimeEntryTagRequest{
		Name:    cmd.OldName,
		NewName: cmd.NewName,
	}

	if err := client.Time().RenameTag(ctx, teamID, req); err != nil {
		return err
	}

	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(os.Stdout, map[string]string{"status": "success"})
	}
	if outfmt.IsPlain(ctx) {
		headers := []string{"STATUS", "OLD_NAME", "NEW_NAME"}
		return outfmt.WritePlain(os.Stdout, headers, [][]string{{"success", cmd.OldName, cmd.NewName}})
	}

	fmt.Fprintf(os.Stderr, "Tag renamed: %s -> %s\n", cmd.OldName, cmd.NewName)

	return nil
}

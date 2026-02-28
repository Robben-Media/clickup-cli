package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/builtbyrobben/clickup-cli/internal/clickup"
	"github.com/builtbyrobben/clickup-cli/internal/outfmt"
)

type RelationshipsCmd struct {
	AddDep    RelationshipsAddDepCmd    `cmd:"" help:"Add a task dependency"`
	RemoveDep RelationshipsRemoveDepCmd `cmd:"" help:"Remove a task dependency"`
	Link      RelationshipsLinkCmd      `cmd:"" help:"Link two tasks"`
	Unlink    RelationshipsUnlinkCmd    `cmd:"" help:"Unlink two tasks"`
}

type RelationshipsAddDepCmd struct {
	TaskID       string `arg:"" required:"" help:"Task ID"`
	DependsOn    string `help:"Task ID that this task depends on (this task waits for other)"`
	DependencyOf string `help:"Task ID that depends on this task (this task blocks other)"`
}

func (cmd *RelationshipsAddDepCmd) Run(ctx context.Context) error {
	client, err := getClickUpClient(ctx)
	if err != nil {
		return err
	}

	if cmd.DependsOn == "" && cmd.DependencyOf == "" {
		return fmt.Errorf("either --depends-on or --dependency-of must be specified")
	}

	req := clickup.AddDependencyRequest{
		DependsOn:    cmd.DependsOn,
		DependencyOf: cmd.DependencyOf,
	}

	if err := client.Relationships().AddDependency(ctx, cmd.TaskID, req); err != nil {
		return err
	}

	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(os.Stdout, map[string]string{
			"status":  "success",
			"message": "Dependency added",
			"task_id": cmd.TaskID,
		})
	}

	if outfmt.IsPlain(ctx) {
		headers := []string{"STATUS", "TASK_ID"}
		var other string
		if cmd.DependsOn != "" {
			other = cmd.DependsOn
		} else {
			other = cmd.DependencyOf
		}

		rows := [][]string{{"success", cmd.TaskID + " -> " + other}}

		return outfmt.WritePlain(os.Stdout, headers, rows)
	}

	if cmd.DependsOn != "" {
		fmt.Fprintf(os.Stderr, "Dependency added: task %s now depends on %s\n", cmd.TaskID, cmd.DependsOn)
	} else {
		fmt.Fprintf(os.Stderr, "Dependency added: task %s now blocks %s\n", cmd.TaskID, cmd.DependencyOf)
	}

	return nil
}

type RelationshipsRemoveDepCmd struct {
	TaskID       string `arg:"" required:"" help:"Task ID"`
	DependsOn    string `help:"Task ID to remove as a dependency (this task was waiting for other)"`
	DependencyOf string `help:"Task ID to remove as dependent (this task was blocking other)"`
}

func (cmd *RelationshipsRemoveDepCmd) Run(ctx context.Context) error {
	client, err := getClickUpClient(ctx)
	if err != nil {
		return err
	}

	if cmd.DependsOn == "" && cmd.DependencyOf == "" {
		return fmt.Errorf("either --depends-on or --dependency-of must be specified")
	}

	req := clickup.AddDependencyRequest{
		DependsOn:    cmd.DependsOn,
		DependencyOf: cmd.DependencyOf,
	}

	if err := client.Relationships().DeleteDependency(ctx, cmd.TaskID, req); err != nil {
		return err
	}

	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(os.Stdout, map[string]string{
			"status":  "success",
			"message": "Dependency removed",
			"task_id": cmd.TaskID,
		})
	}

	if outfmt.IsPlain(ctx) {
		headers := []string{"STATUS", "TASK_ID"}
		rows := [][]string{{"success", cmd.TaskID}}

		return outfmt.WritePlain(os.Stdout, headers, rows)
	}

	fmt.Fprintf(os.Stderr, "Dependency removed from task %s\n", cmd.TaskID)

	return nil
}

type RelationshipsLinkCmd struct {
	TaskID       string `arg:"" required:"" help:"Task ID"`
	LinkedTaskID string `arg:"" required:"" help:"Task ID to link to"`
}

func (cmd *RelationshipsLinkCmd) Run(ctx context.Context) error {
	client, err := getClickUpClient(ctx)
	if err != nil {
		return err
	}

	if err := client.Relationships().AddLink(ctx, cmd.TaskID, cmd.LinkedTaskID); err != nil {
		return err
	}

	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(os.Stdout, map[string]string{
			"status":    "success",
			"message":   "Task link added",
			"task_id":   cmd.TaskID,
			"linked_to": cmd.LinkedTaskID,
		})
	}

	if outfmt.IsPlain(ctx) {
		headers := []string{"STATUS", "TASK_ID", "LINKED_TO"}
		rows := [][]string{{"success", cmd.TaskID, cmd.LinkedTaskID}}

		return outfmt.WritePlain(os.Stdout, headers, rows)
	}

	fmt.Fprintf(os.Stderr, "Tasks linked: %s <-> %s\n", cmd.TaskID, cmd.LinkedTaskID)

	return nil
}

type RelationshipsUnlinkCmd struct {
	TaskID       string `arg:"" required:"" help:"Task ID"`
	LinkedTaskID string `arg:"" required:"" help:"Task ID to unlink from"`
}

func (cmd *RelationshipsUnlinkCmd) Run(ctx context.Context) error {
	client, err := getClickUpClient(ctx)
	if err != nil {
		return err
	}

	if err := client.Relationships().DeleteLink(ctx, cmd.TaskID, cmd.LinkedTaskID); err != nil {
		return err
	}

	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(os.Stdout, map[string]string{
			"status":   "success",
			"message":  "Task link removed",
			"task_id":  cmd.TaskID,
			"unlinked": cmd.LinkedTaskID,
		})
	}

	if outfmt.IsPlain(ctx) {
		headers := []string{"STATUS", "TASK_ID", "UNLINKED"}
		rows := [][]string{{"success", cmd.TaskID, cmd.LinkedTaskID}}

		return outfmt.WritePlain(os.Stdout, headers, rows)
	}

	fmt.Fprintf(os.Stderr, "Tasks unlinked: %s <-> %s\n", cmd.TaskID, cmd.LinkedTaskID)

	return nil
}

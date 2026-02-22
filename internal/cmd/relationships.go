package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/builtbyrobben/clickup-cli/internal/clickup"
	"github.com/builtbyrobben/clickup-cli/internal/outfmt"
)

type RelationshipsCmd struct {
	AddDep    RelationshipsAddDepCmd    `cmd:"" name:"add-dep" help:"Add a dependency to a task"`
	RemoveDep RelationshipsRemoveDepCmd `cmd:"" name:"remove-dep" help:"Remove a dependency from a task"`
	Link      RelationshipsLinkCmd      `cmd:"" help:"Link two tasks"`
	Unlink    RelationshipsUnlinkCmd    `cmd:"" help:"Unlink two tasks"`
}

type RelationshipsAddDepCmd struct {
	TaskID    string `arg:"" required:"" help:"Task ID to add dependency to"`
	DependsOn string `help:"Task ID that this task depends on (this task waits for the other)"`
	Blocking  string `help:"Task ID that this task blocks (this task blocks the other)"`
}

func (cmd *RelationshipsAddDepCmd) Run(ctx context.Context) error {
	client, err := getClickUpClient(ctx)
	if err != nil {
		return err
	}

	req := clickup.AddDependencyRequest{
		DependsOn:    cmd.DependsOn,
		DependencyOf: cmd.Blocking,
	}

	if err := client.Relationships().AddDependency(ctx, cmd.TaskID, req); err != nil {
		return err
	}

	var relatedTaskID string
	var relationType string

	if cmd.DependsOn != "" {
		relatedTaskID = cmd.DependsOn
		relationType = "depends_on"
	} else {
		relatedTaskID = cmd.Blocking
		relationType = "blocks"
	}

	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(os.Stdout, map[string]string{
			"status":        "success",
			"message":       "Dependency added",
			"task_id":       cmd.TaskID,
			"relation_type": relationType,
			"related_task":  relatedTaskID,
		})
	}

	if outfmt.IsPlain(ctx) {
		headers := []string{"STATUS", "TASK_ID", "RELATION_TYPE", "RELATED_TASK"}
		rows := [][]string{{"success", cmd.TaskID, relationType, relatedTaskID}}
		return outfmt.WritePlain(os.Stdout, headers, rows)
	}

	fmt.Fprintf(os.Stderr, "Dependency added: task %s %s task %s\n", cmd.TaskID, relationType, relatedTaskID)

	return nil
}

type RelationshipsRemoveDepCmd struct {
	TaskID    string `arg:"" required:"" help:"Task ID to remove dependency from"`
	DependsOn string `help:"Task ID to stop waiting on"`
	Blocking  string `help:"Task ID to stop blocking"`
}

func (cmd *RelationshipsRemoveDepCmd) Run(ctx context.Context) error {
	client, err := getClickUpClient(ctx)
	if err != nil {
		return err
	}

	req := clickup.AddDependencyRequest{
		DependsOn:    cmd.DependsOn,
		DependencyOf: cmd.Blocking,
	}

	if err := client.Relationships().RemoveDependency(ctx, cmd.TaskID, req); err != nil {
		return err
	}

	var relatedTaskID string
	var relationType string

	if cmd.DependsOn != "" {
		relatedTaskID = cmd.DependsOn
		relationType = "depends_on"
	} else {
		relatedTaskID = cmd.Blocking
		relationType = "blocks"
	}

	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(os.Stdout, map[string]string{
			"status":        "success",
			"message":       "Dependency removed",
			"task_id":       cmd.TaskID,
			"relation_type": relationType,
			"related_task":  relatedTaskID,
		})
	}

	if outfmt.IsPlain(ctx) {
		headers := []string{"STATUS", "TASK_ID", "RELATION_TYPE", "RELATED_TASK"}
		rows := [][]string{{"success", cmd.TaskID, relationType, relatedTaskID}}
		return outfmt.WritePlain(os.Stdout, headers, rows)
	}

	fmt.Fprintf(os.Stderr, "Dependency removed: task %s no longer %s task %s\n", cmd.TaskID, relationType, relatedTaskID)

	return nil
}

type RelationshipsLinkCmd struct {
	TaskID       string `arg:"" required:"" help:"First task ID"`
	LinkedTaskID string `arg:"" required:"" help:"Task ID to link to"`
}

func (cmd *RelationshipsLinkCmd) Run(ctx context.Context) error {
	client, err := getClickUpClient(ctx)
	if err != nil {
		return err
	}

	if err := client.Relationships().Link(ctx, cmd.TaskID, cmd.LinkedTaskID); err != nil {
		return err
	}

	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(os.Stdout, map[string]string{
			"status":         "success",
			"message":        "Task link added",
			"task_id":        cmd.TaskID,
			"linked_task_id": cmd.LinkedTaskID,
		})
	}

	if outfmt.IsPlain(ctx) {
		headers := []string{"STATUS", "TASK_ID", "LINKED_TO"}
		rows := [][]string{{"success", cmd.TaskID, cmd.LinkedTaskID}}
		return outfmt.WritePlain(os.Stdout, headers, rows)
	}

	fmt.Fprintf(os.Stderr, "Task link added: %s <-> %s\n", cmd.TaskID, cmd.LinkedTaskID)

	return nil
}

type RelationshipsUnlinkCmd struct {
	TaskID       string `arg:"" required:"" help:"First task ID"`
	LinkedTaskID string `arg:"" required:"" help:"Task ID to unlink from"`
}

func (cmd *RelationshipsUnlinkCmd) Run(ctx context.Context) error {
	client, err := getClickUpClient(ctx)
	if err != nil {
		return err
	}

	if err := client.Relationships().Unlink(ctx, cmd.TaskID, cmd.LinkedTaskID); err != nil {
		return err
	}

	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(os.Stdout, map[string]string{
			"status":         "success",
			"message":        "Task link removed",
			"task_id":        cmd.TaskID,
			"linked_task_id": cmd.LinkedTaskID,
		})
	}

	if outfmt.IsPlain(ctx) {
		headers := []string{"STATUS", "TASK_ID", "LINKED_TO"}
		rows := [][]string{{"success", cmd.TaskID, cmd.LinkedTaskID}}
		return outfmt.WritePlain(os.Stdout, headers, rows)
	}

	fmt.Fprintf(os.Stderr, "Task link removed: %s <-> %s\n", cmd.TaskID, cmd.LinkedTaskID)

	return nil
}

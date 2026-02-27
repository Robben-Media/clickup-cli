package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/builtbyrobben/clickup-cli/internal/clickup"
	"github.com/builtbyrobben/clickup-cli/internal/outfmt"
)

type TagsCmd struct {
	List   TagsListCmd   `cmd:"" help:"List tags in a space"`
	Create TagsCreateCmd `cmd:"" help:"Create a tag in a space"`
	Update TagsUpdateCmd `cmd:"" help:"Update a tag in a space"`
	Delete TagsDeleteCmd `cmd:"" help:"Delete a tag from a space"`
	Add    TagsAddCmd    `cmd:"" help:"Add a tag to a task"`
	Remove TagsRemoveCmd `cmd:"" help:"Remove a tag from a task"`
}

type TagsListCmd struct {
	SpaceID string `required:"" help:"Space ID"`
}

func (cmd *TagsListCmd) Run(ctx context.Context) error {
	client, err := getClickUpClient(ctx)
	if err != nil {
		return err
	}

	result, err := client.Tags().List(ctx, cmd.SpaceID)
	if err != nil {
		return err
	}

	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(os.Stdout, result)
	}

	if outfmt.IsPlain(ctx) {
		headers := []string{"NAME", "FG_COLOR", "BG_COLOR"}
		var rows [][]string
		for _, tag := range result.Tags {
			rows = append(rows, []string{tag.Name, tag.TagFg, tag.TagBg})
		}

		return outfmt.WritePlain(os.Stdout, headers, rows)
	}

	if len(result.Tags) == 0 {
		fmt.Fprintln(os.Stderr, "No tags found in this space")
		return nil
	}

	fmt.Fprintf(os.Stderr, "Tags in space %s\n\n", cmd.SpaceID)

	for _, tag := range result.Tags {
		if tag.TagBg != "" {
			fmt.Fprintf(os.Stderr, "  %s (bg: %s)\n", tag.Name, tag.TagBg)
		} else {
			fmt.Fprintf(os.Stderr, "  %s\n", tag.Name)
		}
	}

	return nil
}

type TagsCreateCmd struct {
	SpaceID string `required:"" help:"Space ID"`
	Name    string `arg:"" required:"" help:"Tag name"`
	Bg      string `help:"Background color (hex, e.g., #f44336)"`
	Fg      string `help:"Foreground color (hex, e.g., #ffffff)"`
}

func (cmd *TagsCreateCmd) Run(ctx context.Context) error {
	client, err := getClickUpClient(ctx)
	if err != nil {
		return err
	}

	req := clickup.CreateSpaceTagRequest{
		Tag: clickup.SpaceTag{
			Name:  cmd.Name,
			TagFg: cmd.Fg,
			TagBg: cmd.Bg,
		},
	}

	if err := client.Tags().Create(ctx, cmd.SpaceID, req); err != nil {
		return err
	}

	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(os.Stdout, map[string]string{
			"status":  "success",
			"message": "Tag created",
			"name":    cmd.Name,
		})
	}

	if outfmt.IsPlain(ctx) {
		headers := []string{"STATUS", "NAME"}
		rows := [][]string{{"success", cmd.Name}}

		return outfmt.WritePlain(os.Stdout, headers, rows)
	}

	fmt.Fprintf(os.Stderr, "Tag '%s' created\n", cmd.Name)

	return nil
}

type TagsUpdateCmd struct {
	SpaceID string `required:"" help:"Space ID"`
	Name    string `arg:"" required:"" help:"Current tag name"`
	NewName string `help:"New tag name"`
	Bg      string `help:"Background color (hex, e.g., #f44336)"`
	Fg      string `help:"Foreground color (hex, e.g., #ffffff)"`
}

func (cmd *TagsUpdateCmd) Run(ctx context.Context) error {
	client, err := getClickUpClient(ctx)
	if err != nil {
		return err
	}

	// Get current tag to preserve unspecified fields
	result, err := client.Tags().List(ctx, cmd.SpaceID)
	if err != nil {
		return err
	}

	var currentTag *clickup.SpaceTag
	for i := range result.Tags {
		if result.Tags[i].Name == cmd.Name {
			currentTag = &result.Tags[i]
			break
		}
	}

	if currentTag == nil {
		return fmt.Errorf("tag '%s' not found", cmd.Name)
	}

	// Build update request, preserving existing values
	tag := clickup.SpaceTag{}
	if cmd.NewName != "" {
		tag.Name = cmd.NewName
	} else {
		tag.Name = cmd.Name
	}

	if cmd.Fg != "" {
		tag.TagFg = cmd.Fg
	} else {
		tag.TagFg = currentTag.TagFg
	}

	if cmd.Bg != "" {
		tag.TagBg = cmd.Bg
	} else {
		tag.TagBg = currentTag.TagBg
	}

	req := clickup.EditSpaceTagRequest{Tag: tag}

	if err := client.Tags().Update(ctx, cmd.SpaceID, cmd.Name, req); err != nil {
		return err
	}

	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(os.Stdout, map[string]string{
			"status":  "success",
			"message": "Tag updated",
			"name":    tag.Name,
		})
	}

	if outfmt.IsPlain(ctx) {
		headers := []string{"STATUS", "NAME"}
		rows := [][]string{{"success", tag.Name}}

		return outfmt.WritePlain(os.Stdout, headers, rows)
	}

	fmt.Fprintf(os.Stderr, "Tag '%s' updated\n", cmd.Name)

	return nil
}

type TagsDeleteCmd struct {
	SpaceID string `required:"" help:"Space ID"`
	Name    string `arg:"" required:"" help:"Tag name"`
}

func (cmd *TagsDeleteCmd) Run(ctx context.Context) error {
	client, err := getClickUpClient(ctx)
	if err != nil {
		return err
	}

	if !forceEnabled(ctx) {
		fmt.Fprintf(os.Stderr, "Warning: This will permanently delete tag '%s' from space %s\n", cmd.Name, cmd.SpaceID)
		fmt.Fprint(os.Stderr, "Use --force to confirm deletion\n")

		return fmt.Errorf("operation cancelled: use --force to confirm")
	}

	if err := client.Tags().Delete(ctx, cmd.SpaceID, cmd.Name); err != nil {
		return err
	}

	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(os.Stdout, map[string]string{
			"status":  "success",
			"message": "Tag deleted",
			"name":    cmd.Name,
		})
	}

	if outfmt.IsPlain(ctx) {
		headers := []string{"STATUS", "NAME"}
		rows := [][]string{{"success", cmd.Name}}

		return outfmt.WritePlain(os.Stdout, headers, rows)
	}

	fmt.Fprintf(os.Stderr, "Tag '%s' deleted\n", cmd.Name)

	return nil
}

type TagsAddCmd struct {
	TaskID string `required:"" help:"Task ID"`
	Name   string `arg:"" required:"" help:"Tag name"`
}

func (cmd *TagsAddCmd) Run(ctx context.Context) error {
	client, err := getClickUpClient(ctx)
	if err != nil {
		return err
	}

	if err := client.Tags().AddToTask(ctx, cmd.TaskID, cmd.Name); err != nil {
		return err
	}

	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(os.Stdout, map[string]string{
			"status":  "success",
			"message": "Tag added to task",
			"task_id": cmd.TaskID,
			"tag":     cmd.Name,
		})
	}

	if outfmt.IsPlain(ctx) {
		headers := []string{"STATUS", "TASK_ID", "TAG"}
		rows := [][]string{{"success", cmd.TaskID, cmd.Name}}

		return outfmt.WritePlain(os.Stdout, headers, rows)
	}

	fmt.Fprintf(os.Stderr, "Tag '%s' added to task %s\n", cmd.Name, cmd.TaskID)

	return nil
}

type TagsRemoveCmd struct {
	TaskID string `required:"" help:"Task ID"`
	Name   string `arg:"" required:"" help:"Tag name"`
}

func (cmd *TagsRemoveCmd) Run(ctx context.Context) error {
	client, err := getClickUpClient(ctx)
	if err != nil {
		return err
	}

	if err := client.Tags().RemoveFromTask(ctx, cmd.TaskID, cmd.Name); err != nil {
		return err
	}

	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(os.Stdout, map[string]string{
			"status":  "success",
			"message": "Tag removed from task",
			"task_id": cmd.TaskID,
			"tag":     cmd.Name,
		})
	}

	if outfmt.IsPlain(ctx) {
		headers := []string{"STATUS", "TASK_ID", "TAG"}
		rows := [][]string{{"success", cmd.TaskID, cmd.Name}}

		return outfmt.WritePlain(os.Stdout, headers, rows)
	}

	fmt.Fprintf(os.Stderr, "Tag '%s' removed from task %s\n", cmd.Name, cmd.TaskID)

	return nil
}

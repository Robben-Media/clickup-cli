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
	Create TagsCreateCmd `cmd:"" help:"Create a new tag in a space"`
	Update TagsUpdateCmd `cmd:"" help:"Update a tag in a space"`
	Delete TagsDeleteCmd `cmd:"" help:"Delete a tag from a space"`
	Add    TagsAddCmd    `cmd:"" help:"Add a tag to a task"`
	Remove TagsRemoveCmd `cmd:"" help:"Remove a tag from a task"`
}

type TagsListCmd struct {
	SpaceID string `required:"" help:"Space ID to list tags from"`
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
		fmt.Fprintln(os.Stderr, "No tags found")
		return nil
	}

	fmt.Fprintf(os.Stderr, "Tags in space %s\n\n", cmd.SpaceID)

	for _, tag := range result.Tags {
		if tag.TagBg != "" {
			fmt.Printf("  %s (bg: %s)\n", tag.Name, tag.TagBg)
		} else {
			fmt.Printf("  %s\n", tag.Name)
		}
	}

	return nil
}

type TagsCreateCmd struct {
	SpaceID string `arg:"" required:"" help:"Space ID to create tag in"`
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
			TagBg: cmd.Bg,
			TagFg: cmd.Fg,
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

	fmt.Fprintf(os.Stderr, "Tag %q created\n", cmd.Name)

	return nil
}

type TagsUpdateCmd struct {
	SpaceID string `arg:"" required:"" help:"Space ID containing the tag"`
	Name    string `arg:"" required:"" help:"Current tag name"`
	NewName string `help:"New tag name"`
	Bg      string `help:"New background color (hex, e.g., #f44336)"`
	Fg      string `help:"New foreground color (hex, e.g., #ffffff)"`
}

func (cmd *TagsUpdateCmd) Run(ctx context.Context) error {
	client, err := getClickUpClient(ctx)
	if err != nil {
		return err
	}

	// Use the current name if no new name provided
	name := cmd.NewName
	if name == "" {
		name = cmd.Name
	}

	req := clickup.EditSpaceTagRequest{
		Tag: clickup.SpaceTag{
			Name:  name,
			TagBg: cmd.Bg,
			TagFg: cmd.Fg,
		},
	}

	if err := client.Tags().Update(ctx, cmd.SpaceID, cmd.Name, req); err != nil {
		return err
	}

	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(os.Stdout, map[string]string{
			"status":  "success",
			"message": "Tag updated",
			"name":    name,
		})
	}
	if outfmt.IsPlain(ctx) {
		headers := []string{"STATUS", "NAME"}
		rows := [][]string{{"success", name}}
		return outfmt.WritePlain(os.Stdout, headers, rows)
	}

	fmt.Fprintf(os.Stderr, "Tag %q updated\n", name)

	return nil
}

type TagsDeleteCmd struct {
	SpaceID string `arg:"" required:"" help:"Space ID containing the tag"`
	Name    string `arg:"" required:"" help:"Tag name to delete"`
}

func (cmd *TagsDeleteCmd) Run(ctx context.Context) error {
	client, err := getClickUpClient(ctx)
	if err != nil {
		return err
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

	fmt.Fprintf(os.Stderr, "Tag %q deleted\n", cmd.Name)

	return nil
}

type TagsAddCmd struct {
	TaskID  string `arg:"" required:"" help:"Task ID to add tag to"`
	TagName string `arg:"" required:"" help:"Tag name to add"`
}

func (cmd *TagsAddCmd) Run(ctx context.Context) error {
	client, err := getClickUpClient(ctx)
	if err != nil {
		return err
	}

	if err := client.Tags().AddToTask(ctx, cmd.TaskID, cmd.TagName); err != nil {
		return err
	}

	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(os.Stdout, map[string]string{
			"status":   "success",
			"message":  "Tag added to task",
			"task_id":  cmd.TaskID,
			"tag_name": cmd.TagName,
		})
	}
	if outfmt.IsPlain(ctx) {
		headers := []string{"STATUS", "TASK_ID", "TAG_NAME"}
		rows := [][]string{{"success", cmd.TaskID, cmd.TagName}}
		return outfmt.WritePlain(os.Stdout, headers, rows)
	}

	fmt.Fprintf(os.Stderr, "Tag %q added to task %s\n", cmd.TagName, cmd.TaskID)

	return nil
}

type TagsRemoveCmd struct {
	TaskID  string `arg:"" required:"" help:"Task ID to remove tag from"`
	TagName string `arg:"" required:"" help:"Tag name to remove"`
}

func (cmd *TagsRemoveCmd) Run(ctx context.Context) error {
	client, err := getClickUpClient(ctx)
	if err != nil {
		return err
	}

	if err := client.Tags().RemoveFromTask(ctx, cmd.TaskID, cmd.TagName); err != nil {
		return err
	}

	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(os.Stdout, map[string]string{
			"status":   "success",
			"message":  "Tag removed from task",
			"task_id":  cmd.TaskID,
			"tag_name": cmd.TagName,
		})
	}
	if outfmt.IsPlain(ctx) {
		headers := []string{"STATUS", "TASK_ID", "TAG_NAME"}
		rows := [][]string{{"success", cmd.TaskID, cmd.TagName}}
		return outfmt.WritePlain(os.Stdout, headers, rows)
	}

	fmt.Fprintf(os.Stderr, "Tag %q removed from task %s\n", cmd.TagName, cmd.TaskID)

	return nil
}

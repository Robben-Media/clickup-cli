package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/builtbyrobben/clickup-cli/internal/outfmt"
)

type CommentsCmd struct {
	List CommentsListCmd `cmd:"" help:"List comments on a task"`
	Add  CommentsAddCmd  `cmd:"" help:"Add a comment to a task"`
}

type CommentsListCmd struct {
	TaskID string `arg:"" required:"" help:"Task ID"`
}

func (cmd *CommentsListCmd) Run(ctx context.Context) error {
	client, err := getClickUpClient(ctx)
	if err != nil {
		return err
	}

	result, err := client.Comments().List(ctx, cmd.TaskID)
	if err != nil {
		return err
	}

	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(os.Stdout, result)
	}
	if outfmt.IsPlain(ctx) {
		headers := []string{"ID", "USER", "TEXT", "DATE"}
		var rows [][]string
		for _, comment := range result.Comments {
			rows = append(rows, []string{comment.ID.String(), comment.User.Username, comment.Text, comment.Date})
		}
		return outfmt.WritePlain(os.Stdout, headers, rows)
	}

	if len(result.Comments) == 0 {
		fmt.Fprintln(os.Stderr, "No comments found")
		return nil
	}

	fmt.Fprintf(os.Stderr, "Found %d comments\n\n", len(result.Comments))

	for _, comment := range result.Comments {
		fmt.Printf("ID: %s\n", comment.ID)
		fmt.Printf("  User: %s\n", comment.User.Username)
		fmt.Printf("  Text: %s\n", comment.Text)

		if comment.Date != "" {
			fmt.Printf("  Date: %s\n", comment.Date)
		}

		fmt.Println()
	}

	return nil
}

type CommentsAddCmd struct {
	TaskID string `arg:"" required:"" help:"Task ID"`
	Text   string `arg:"" required:"" help:"Comment text"`
}

func (cmd *CommentsAddCmd) Run(ctx context.Context) error {
	client, err := getClickUpClient(ctx)
	if err != nil {
		return err
	}

	result, err := client.Comments().Add(ctx, cmd.TaskID, cmd.Text)
	if err != nil {
		return err
	}

	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(os.Stdout, result)
	}
	if outfmt.IsPlain(ctx) {
		headers := []string{"ID"}
		rows := [][]string{{result.ID.String()}}
		return outfmt.WritePlain(os.Stdout, headers, rows)
	}

	fmt.Fprintf(os.Stderr, "Comment added (ID: %s)\n", result.ID)

	return nil
}

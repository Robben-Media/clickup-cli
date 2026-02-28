package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/builtbyrobben/clickup-cli/internal/clickup"
	"github.com/builtbyrobben/clickup-cli/internal/outfmt"
)

type CommentsCmd struct {
	List         CommentsListCmd         `cmd:"" help:"List comments on a task"`
	Add          CommentsAddCmd          `cmd:"" help:"Add a comment to a task"`
	Delete       CommentsDeleteCmd       `cmd:"" help:"Delete a comment"`
	Update       CommentsUpdateCmd       `cmd:"" help:"Update a comment"`
	Replies      CommentsRepliesCmd      `cmd:"" help:"List threaded replies to a comment"`
	Reply        CommentsReplyCmd        `cmd:"" help:"Create a threaded reply to a comment"`
	ListComments CommentsListCommentsCmd `cmd:"" help:"List comments on a list" aliases:"list-comments"`
	AddList      CommentsAddListCmd      `cmd:"" help:"Add a comment to a list" aliases:"add-list"`
	ViewComments CommentsViewCommentsCmd `cmd:"" help:"List comments on a view" aliases:"view-comments"`
	AddView      CommentsAddViewCmd      `cmd:"" help:"Add a comment to a view" aliases:"add-view"`
	Subtypes     CommentsSubtypesCmd     `cmd:"" help:"Get post subtype IDs (v3 API)"`
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

type CommentsDeleteCmd struct {
	CommentID string `arg:"" required:"" help:"Comment ID"`
}

func (cmd *CommentsDeleteCmd) Run(ctx context.Context) error {
	client, err := getClickUpClient(ctx)
	if err != nil {
		return err
	}

	if !forceEnabled(ctx) {
		fmt.Fprintf(os.Stderr, "Warning: This will permanently delete comment %s\n", cmd.CommentID)
		fmt.Fprint(os.Stderr, "Use --force to confirm deletion\n")
		return nil
	}

	if err := client.Comments().Delete(ctx, cmd.CommentID); err != nil {
		return err
	}

	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(os.Stdout, map[string]string{"status": "success", "comment_id": cmd.CommentID})
	}
	if outfmt.IsPlain(ctx) {
		headers := []string{"STATUS", "COMMENT_ID"}
		rows := [][]string{{"success", cmd.CommentID}}
		return outfmt.WritePlain(os.Stdout, headers, rows)
	}

	fmt.Fprintf(os.Stderr, "Comment %s deleted\n", cmd.CommentID)

	return nil
}

type CommentsUpdateCmd struct {
	CommentID string `arg:"" required:"" help:"Comment ID"`
	Text      string `help:"New comment text"`
	Resolved  bool   `help:"Mark as resolved"`
	Assignee  int    `help:"Reassign to user ID"`
}

func (cmd *CommentsUpdateCmd) Run(ctx context.Context) error {
	client, err := getClickUpClient(ctx)
	if err != nil {
		return err
	}

	req := clickup.UpdateCommentRequest{
		CommentText: cmd.Text,
		Assignee:    cmd.Assignee,
	}

	if cmd.Resolved {
		resolved := true
		req.Resolved = &resolved
	}

	if err := client.Comments().Update(ctx, cmd.CommentID, req); err != nil {
		return err
	}

	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(os.Stdout, map[string]string{"status": "success", "comment_id": cmd.CommentID})
	}
	if outfmt.IsPlain(ctx) {
		headers := []string{"STATUS", "COMMENT_ID"}
		rows := [][]string{{"success", cmd.CommentID}}
		return outfmt.WritePlain(os.Stdout, headers, rows)
	}

	fmt.Fprintf(os.Stderr, "Comment %s updated\n", cmd.CommentID)

	return nil
}

type CommentsRepliesCmd struct {
	CommentID string `arg:"" required:"" help:"Comment ID"`
}

func (cmd *CommentsRepliesCmd) Run(ctx context.Context) error {
	client, err := getClickUpClient(ctx)
	if err != nil {
		return err
	}

	result, err := client.Comments().Replies(ctx, cmd.CommentID)
	if err != nil {
		return err
	}

	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(os.Stdout, result)
	}
	if outfmt.IsPlain(ctx) {
		headers := []string{"ID", "USER", "DATE", "TEXT"}
		var rows [][]string
		for _, comment := range result.Comments {
			rows = append(rows, []string{comment.ID.String(), comment.User.Username, comment.Date, comment.Text})
		}
		return outfmt.WritePlain(os.Stdout, headers, rows)
	}

	if len(result.Comments) == 0 {
		fmt.Fprintf(os.Stderr, "No replies to comment %s\n", cmd.CommentID)
		return nil
	}

	fmt.Fprintf(os.Stderr, "Replies to comment %s\n\n", cmd.CommentID)

	for _, comment := range result.Comments {
		fmt.Printf("ID: %s\n", comment.ID)
		fmt.Printf("  User: %s\n", comment.User.Username)
		fmt.Printf("  Date: %s\n", comment.Date)
		fmt.Printf("  Text: %s\n", comment.Text)
		fmt.Println()
	}

	return nil
}

type CommentsReplyCmd struct {
	CommentID string `arg:"" required:"" help:"Comment ID"`
	Text      string `arg:"" required:"" help:"Reply text"`
}

func (cmd *CommentsReplyCmd) Run(ctx context.Context) error {
	client, err := getClickUpClient(ctx)
	if err != nil {
		return err
	}

	result, err := client.Comments().Reply(ctx, cmd.CommentID, cmd.Text)
	if err != nil {
		return err
	}

	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(os.Stdout, result)
	}
	if outfmt.IsPlain(ctx) {
		headers := []string{"ID", "TEXT"}
		rows := [][]string{{result.ID.String(), result.Text}}
		return outfmt.WritePlain(os.Stdout, headers, rows)
	}

	fmt.Fprintln(os.Stderr, "Reply created")
	fmt.Fprintf(os.Stderr, "ID: %s\n", result.ID)
	fmt.Fprintf(os.Stderr, "Text: %s\n", result.Text)

	return nil
}

type CommentsListCommentsCmd struct {
	ListID string `arg:"" required:"" help:"List ID"`
}

func (cmd *CommentsListCommentsCmd) Run(ctx context.Context) error {
	client, err := getClickUpClient(ctx)
	if err != nil {
		return err
	}

	result, err := client.Comments().ListComments(ctx, cmd.ListID)
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
		fmt.Fprintf(os.Stderr, "No comments on list %s\n", cmd.ListID)
		return nil
	}

	fmt.Fprintf(os.Stderr, "Found %d comments on list %s\n\n", len(result.Comments), cmd.ListID)

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

type CommentsAddListCmd struct {
	ListID   string `arg:"" required:"" help:"List ID"`
	Text     string `arg:"" required:"" help:"Comment text"`
	Assignee int    `help:"Assign to user ID"`
}

func (cmd *CommentsAddListCmd) Run(ctx context.Context) error {
	client, err := getClickUpClient(ctx)
	if err != nil {
		return err
	}

	req := clickup.CreateListCommentRequest{
		CommentText: cmd.Text,
		Assignee:    cmd.Assignee,
	}

	result, err := client.Comments().AddList(ctx, cmd.ListID, req)
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

	fmt.Fprintf(os.Stderr, "Comment added to list (ID: %s)\n", result.ID)

	return nil
}

type CommentsViewCommentsCmd struct {
	ViewID  string `arg:"" required:"" help:"View ID"`
	Start   int    `help:"Pagination offset"`
	StartID string `help:"Pagination cursor"`
}

func (cmd *CommentsViewCommentsCmd) Run(ctx context.Context) error {
	client, err := getClickUpClient(ctx)
	if err != nil {
		return err
	}

	result, err := client.Comments().ViewComments(ctx, cmd.ViewID, cmd.Start, cmd.StartID)
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
		fmt.Fprintf(os.Stderr, "No comments on view %s\n", cmd.ViewID)
		return nil
	}

	fmt.Fprintf(os.Stderr, "Found %d comments on view %s\n\n", len(result.Comments), cmd.ViewID)

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

type CommentsAddViewCmd struct {
	ViewID   string `arg:"" required:"" help:"View ID"`
	Text     string `arg:"" required:"" help:"Comment text"`
	Assignee int    `help:"Assign to user ID"`
}

func (cmd *CommentsAddViewCmd) Run(ctx context.Context) error {
	client, err := getClickUpClient(ctx)
	if err != nil {
		return err
	}

	req := clickup.CreateViewCommentRequest{
		CommentText: cmd.Text,
		Assignee:    cmd.Assignee,
	}

	result, err := client.Comments().AddView(ctx, cmd.ViewID, req)
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

	fmt.Fprintf(os.Stderr, "Comment added to view (ID: %s)\n", result.ID)

	return nil
}

type CommentsSubtypesCmd struct {
	TypeID string `arg:"" required:"" help:"Type ID"`
}

func (cmd *CommentsSubtypesCmd) Run(ctx context.Context) error {
	client, err := getClickUpClient(ctx)
	if err != nil {
		return err
	}

	result, err := client.Comments().Subtypes(ctx, cmd.TypeID)
	if err != nil {
		return err
	}

	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(os.Stdout, result)
	}
	if outfmt.IsPlain(ctx) {
		headers := []string{"ID", "NAME"}
		var rows [][]string
		for _, subtype := range result.Subtypes {
			rows = append(rows, []string{subtype.ID, subtype.Name})
		}
		return outfmt.WritePlain(os.Stdout, headers, rows)
	}

	if len(result.Subtypes) == 0 {
		fmt.Fprintf(os.Stderr, "No subtypes for type %s\n", cmd.TypeID)
		return nil
	}

	fmt.Fprintf(os.Stderr, "Post Subtypes for type %s\n\n", cmd.TypeID)

	for _, subtype := range result.Subtypes {
		fmt.Printf("  %s: %s\n", subtype.ID, subtype.Name)
	}

	return nil
}

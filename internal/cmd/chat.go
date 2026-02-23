package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/builtbyrobben/clickup-cli/internal/clickup"
	"github.com/builtbyrobben/clickup-cli/internal/outfmt"
)

type ChatCmd struct {
	Channels         ChatChannelsCmd         `cmd:"" help:"List all chat channels"`
	Channel          ChatChannelCmd          `cmd:"" help:"Get channel details"`
	ChannelFollowers ChatChannelFollowersCmd `cmd:"" help:"Get channel followers"`
	ChannelMembers   ChatChannelMembersCmd   `cmd:"" help:"Get channel members"`
	CreateChannel    ChatCreateChannelCmd    `cmd:"" help:"Create a channel"`
	CreateDM         ChatCreateDMCmd         `cmd:"" help:"Create a direct message channel"`
	CreateLocChannel ChatCreateLocChannelCmd `cmd:"" help:"Create a location-based channel"`
	UpdateChannel    ChatUpdateChannelCmd    `cmd:"" help:"Update a channel"`
	DeleteChannel    ChatDeleteChannelCmd    `cmd:"" help:"Delete a channel"`
	Messages         ChatMessagesCmd         `cmd:"" help:"List channel messages"`
	Send             ChatSendCmd             `cmd:"" help:"Send a message"`
	UpdateMessage    ChatUpdateMessageCmd    `cmd:"" help:"Update a message"`
	DeleteMessage    ChatDeleteMessageCmd    `cmd:"" help:"Delete a message"`
	Reactions        ChatReactionsCmd        `cmd:"" help:"List message reactions"`
	React            ChatReactCmd            `cmd:"" help:"Add a reaction"`
	Unreact          ChatUnreactCmd          `cmd:"" help:"Remove a reaction"`
	Replies          ChatRepliesCmd          `cmd:"" help:"List message replies"`
	Reply            ChatReplyCmd            `cmd:"" help:"Reply to a message"`
	TaggedUsers      ChatTaggedUsersCmd      `cmd:"" help:"Get tagged users in a message"`
}

type ChatChannelsCmd struct{}

func (cmd *ChatChannelsCmd) Run(ctx context.Context) error {
	client, err := getClickUpClient(ctx)
	if err != nil {
		return err
	}

	result, err := client.Chat().ListChannels(ctx)
	if err != nil {
		return err
	}

	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(os.Stdout, result)
	}
	if outfmt.IsPlain(ctx) {
		headers := []string{"ID", "NAME", "TYPE", "MEMBER_COUNT"}
		rows := make([][]string, 0, len(result.Channels))
		for _, ch := range result.Channels {
			rows = append(rows, []string{ch.ID, ch.Name, ch.Type, fmt.Sprintf("%d", ch.MemberCount)})
		}
		return outfmt.WritePlain(os.Stdout, headers, rows)
	}

	if len(result.Channels) == 0 {
		fmt.Fprintln(os.Stderr, "No channels found")
		return nil
	}

	fmt.Fprintf(os.Stderr, "Found %d channels\n\n", len(result.Channels))

	for _, ch := range result.Channels {
		fmt.Printf("ID: %s\n", ch.ID)
		fmt.Printf("  Name: %s\n", ch.Name)
		fmt.Printf("  Type: %s\n", ch.Type)
		fmt.Printf("  Members: %d\n\n", ch.MemberCount)
	}

	return nil
}

type ChatChannelCmd struct {
	ChannelID string `arg:"" help:"Channel ID" required:""`
}

func (cmd *ChatChannelCmd) Run(ctx context.Context) error {
	client, err := getClickUpClient(ctx)
	if err != nil {
		return err
	}

	result, err := client.Chat().GetChannel(ctx, cmd.ChannelID)
	if err != nil {
		return err
	}

	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(os.Stdout, result)
	}
	if outfmt.IsPlain(ctx) {
		headers := []string{"ID", "NAME", "TYPE", "MEMBER_COUNT"}
		rows := [][]string{{result.ID, result.Name, result.Type, fmt.Sprintf("%d", result.MemberCount)}}
		return outfmt.WritePlain(os.Stdout, headers, rows)
	}

	fmt.Printf("ID: %s\n", result.ID)
	fmt.Printf("Name: %s\n", result.Name)
	fmt.Printf("Type: %s\n", result.Type)
	fmt.Printf("Members: %d\n", result.MemberCount)

	return nil
}

type ChatChannelFollowersCmd struct {
	ChannelID string `arg:"" help:"Channel ID" required:""`
}

func (cmd *ChatChannelFollowersCmd) Run(ctx context.Context) error {
	client, err := getClickUpClient(ctx)
	if err != nil {
		return err
	}

	result, err := client.Chat().GetFollowers(ctx, cmd.ChannelID)
	if err != nil {
		return err
	}

	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(os.Stdout, result)
	}
	if outfmt.IsPlain(ctx) {
		headers := []string{"ID", "USERNAME", "EMAIL"}
		rows := make([][]string, 0, len(result.Members))
		for _, u := range result.Members {
			rows = append(rows, []string{fmt.Sprintf("%d", u.ID), u.Username, u.Email})
		}
		return outfmt.WritePlain(os.Stdout, headers, rows)
	}

	if len(result.Members) == 0 {
		fmt.Fprintln(os.Stderr, "No followers found")
		return nil
	}

	fmt.Fprintf(os.Stderr, "Found %d followers\n\n", len(result.Members))

	for _, u := range result.Members {
		fmt.Printf("ID: %d\n", u.ID)
		fmt.Printf("  Username: %s\n", u.Username)

		if u.Email != "" {
			fmt.Printf("  Email: %s\n", u.Email)
		}

		fmt.Println()
	}

	return nil
}

type ChatChannelMembersCmd struct {
	ChannelID string `arg:"" help:"Channel ID" required:""`
}

func (cmd *ChatChannelMembersCmd) Run(ctx context.Context) error {
	client, err := getClickUpClient(ctx)
	if err != nil {
		return err
	}

	result, err := client.Chat().GetMembers(ctx, cmd.ChannelID)
	if err != nil {
		return err
	}

	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(os.Stdout, result)
	}
	if outfmt.IsPlain(ctx) {
		headers := []string{"ID", "USERNAME", "EMAIL"}
		rows := make([][]string, 0, len(result.Members))
		for _, u := range result.Members {
			rows = append(rows, []string{fmt.Sprintf("%d", u.ID), u.Username, u.Email})
		}
		return outfmt.WritePlain(os.Stdout, headers, rows)
	}

	if len(result.Members) == 0 {
		fmt.Fprintln(os.Stderr, "No members found")
		return nil
	}

	fmt.Fprintf(os.Stderr, "Found %d members\n\n", len(result.Members))

	for _, u := range result.Members {
		fmt.Printf("ID: %d\n", u.ID)
		fmt.Printf("  Username: %s\n", u.Username)

		if u.Email != "" {
			fmt.Printf("  Email: %s\n", u.Email)
		}

		fmt.Println()
	}

	return nil
}

type ChatCreateChannelCmd struct {
	Name string `name:"name" short:"n" help:"Channel name" required:""`
}

func (cmd *ChatCreateChannelCmd) Run(ctx context.Context) error {
	client, err := getClickUpClient(ctx)
	if err != nil {
		return err
	}

	result, err := client.Chat().CreateChannel(ctx, clickup.CreateChatChannelRequest{Name: cmd.Name})
	if err != nil {
		return err
	}

	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(os.Stdout, result)
	}

	fmt.Fprintf(os.Stderr, "Created channel: %s\n", result.Name)
	fmt.Printf("ID: %s\n", result.ID)

	return nil
}

type ChatCreateDMCmd struct {
	Members string `name:"members" short:"m" help:"Comma-separated user IDs" required:""`
}

func (cmd *ChatCreateDMCmd) Run(ctx context.Context) error {
	client, err := getClickUpClient(ctx)
	if err != nil {
		return err
	}

	memberIDs := strings.Split(cmd.Members, ",")

	result, err := client.Chat().CreateDM(ctx, clickup.CreateDMRequest{Members: memberIDs})
	if err != nil {
		return err
	}

	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(os.Stdout, result)
	}

	fmt.Fprintf(os.Stderr, "Created direct message channel\n")
	fmt.Printf("ID: %s\n", result.ID)

	return nil
}

type ChatCreateLocChannelCmd struct {
	Name       string `name:"name" short:"n" help:"Channel name" required:""`
	ParentType string `name:"type" short:"t" help:"Parent type (space, folder, list)" required:""`
	ParentID   string `name:"id" short:"i" help:"Parent ID" required:""`
}

func (cmd *ChatCreateLocChannelCmd) Run(ctx context.Context) error {
	client, err := getClickUpClient(ctx)
	if err != nil {
		return err
	}

	result, err := client.Chat().CreateLocationChannel(ctx, clickup.CreateLocationChannelRequest{
		Name:       cmd.Name,
		ParentType: cmd.ParentType,
		ParentID:   cmd.ParentID,
	})
	if err != nil {
		return err
	}

	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(os.Stdout, result)
	}

	fmt.Fprintf(os.Stderr, "Created location channel: %s\n", result.Name)
	fmt.Printf("ID: %s\n", result.ID)

	return nil
}

type ChatUpdateChannelCmd struct {
	ChannelID string `arg:"" help:"Channel ID" required:""`
	Name      string `name:"name" short:"n" help:"New channel name"`
}

func (cmd *ChatUpdateChannelCmd) Run(ctx context.Context) error {
	client, err := getClickUpClient(ctx)
	if err != nil {
		return err
	}

	result, err := client.Chat().UpdateChannel(ctx, cmd.ChannelID, clickup.UpdateChannelRequest{Name: cmd.Name})
	if err != nil {
		return err
	}

	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(os.Stdout, result)
	}

	fmt.Fprintf(os.Stderr, "Updated channel: %s\n", result.Name)

	return nil
}

type ChatDeleteChannelCmd struct {
	ChannelID string `arg:"" help:"Channel ID" required:""`
	Force     bool   `name:"force" short:"f" help:"Skip confirmation"`
}

func (cmd *ChatDeleteChannelCmd) Run(ctx context.Context) error {
	if !cmd.Force {
		fmt.Fprintf(os.Stderr, "Warning: This will delete channel %s and all its messages.\n", cmd.ChannelID)
		fmt.Fprint(os.Stderr, "Use --force to confirm deletion.\n")

		return nil
	}

	client, err := getClickUpClient(ctx)
	if err != nil {
		return err
	}

	err = client.Chat().DeleteChannel(ctx, cmd.ChannelID)
	if err != nil {
		return err
	}

	fmt.Fprintf(os.Stderr, "Deleted channel: %s\n", cmd.ChannelID)

	return nil
}

type ChatMessagesCmd struct {
	ChannelID string `arg:"" help:"Channel ID" required:""`
	Limit     int    `name:"limit" short:"l" help:"Maximum messages to return"`
	Cursor    string `name:"cursor" short:"c" help:"Pagination cursor"`
}

func (cmd *ChatMessagesCmd) Run(ctx context.Context) error {
	client, err := getClickUpClient(ctx)
	if err != nil {
		return err
	}

	result, err := client.Chat().ListMessages(ctx, cmd.ChannelID, cmd.Limit, cmd.Cursor)
	if err != nil {
		return err
	}

	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(os.Stdout, result)
	}
	if outfmt.IsPlain(ctx) {
		headers := []string{"ID", "USER_ID", "TYPE", "DATE", "CONTENT", "REPLIES"}
		rows := make([][]string, 0, len(result.Data))
		for _, msg := range result.Data {
			rows = append(rows, []string{msg.ID, msg.UserID, msg.Type, string(msg.DateCreated), msg.Content, fmt.Sprintf("%d", msg.RepliesCount)})
		}
		return outfmt.WritePlain(os.Stdout, headers, rows)
	}

	if len(result.Data) == 0 {
		fmt.Fprintln(os.Stderr, "No messages found")
		return nil
	}

	fmt.Fprintf(os.Stderr, "Found %d messages\n\n", len(result.Data))

	for _, msg := range result.Data {
		fmt.Printf("ID: %s\n", msg.ID)
		fmt.Printf("  User: %s\n", msg.UserID)
		fmt.Printf("  Type: %s\n", msg.Type)
		fmt.Printf("  Content: %s\n", msg.Content)
		fmt.Printf("  Replies: %d\n\n", msg.RepliesCount)
	}

	if result.Pagination != nil && result.Pagination.NextPageToken != "" {
		fmt.Fprintf(os.Stderr, "Next page cursor: %s\n", result.Pagination.NextPageToken)
	}

	return nil
}

type ChatSendCmd struct {
	ChannelID string `arg:"" help:"Channel ID" required:""`
	Text      string `name:"text" short:"t" help:"Message text" required:""`
}

func (cmd *ChatSendCmd) Run(ctx context.Context) error {
	client, err := getClickUpClient(ctx)
	if err != nil {
		return err
	}

	result, err := client.Chat().SendMessage(ctx, cmd.ChannelID, clickup.SendMessageRequest{Content: cmd.Text})
	if err != nil {
		return err
	}

	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(os.Stdout, result)
	}

	fmt.Fprintf(os.Stderr, "Sent message\n")
	fmt.Printf("ID: %s\n", result.ID)

	return nil
}

type ChatUpdateMessageCmd struct {
	MessageID string `arg:"" help:"Message ID" required:""`
	Text      string `name:"text" short:"t" help:"New message text" required:""`
}

func (cmd *ChatUpdateMessageCmd) Run(ctx context.Context) error {
	client, err := getClickUpClient(ctx)
	if err != nil {
		return err
	}

	result, err := client.Chat().UpdateMessage(ctx, cmd.MessageID, clickup.UpdateMessageRequest{Content: cmd.Text})
	if err != nil {
		return err
	}

	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(os.Stdout, result)
	}

	fmt.Fprintf(os.Stderr, "Updated message: %s\n", result.ID)

	return nil
}

type ChatDeleteMessageCmd struct {
	MessageID string `arg:"" help:"Message ID" required:""`
	Force     bool   `name:"force" short:"f" help:"Skip confirmation"`
}

func (cmd *ChatDeleteMessageCmd) Run(ctx context.Context) error {
	if !cmd.Force {
		fmt.Fprintf(os.Stderr, "Warning: This will delete message %s.\n", cmd.MessageID)
		fmt.Fprint(os.Stderr, "Use --force to confirm deletion.\n")

		return nil
	}

	client, err := getClickUpClient(ctx)
	if err != nil {
		return err
	}

	err = client.Chat().DeleteMessage(ctx, cmd.MessageID)
	if err != nil {
		return err
	}

	fmt.Fprintf(os.Stderr, "Deleted message: %s\n", cmd.MessageID)

	return nil
}

type ChatReactionsCmd struct {
	MessageID string `arg:"" help:"Message ID" required:""`
}

func (cmd *ChatReactionsCmd) Run(ctx context.Context) error {
	client, err := getClickUpClient(ctx)
	if err != nil {
		return err
	}

	result, err := client.Chat().ListReactions(ctx, cmd.MessageID)
	if err != nil {
		return err
	}

	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(os.Stdout, result)
	}
	if outfmt.IsPlain(ctx) {
		headers := []string{"ID", "USER_ID", "EMOJI", "DATE"}
		rows := make([][]string, 0, len(result.Reactions))
		for _, r := range result.Reactions {
			rows = append(rows, []string{r.ID, r.UserID, r.Reaction, string(r.DateCreated)})
		}
		return outfmt.WritePlain(os.Stdout, headers, rows)
	}

	if len(result.Reactions) == 0 {
		fmt.Fprintln(os.Stderr, "No reactions found")
		return nil
	}

	fmt.Fprintf(os.Stderr, "Found %d reactions\n\n", len(result.Reactions))

	for _, r := range result.Reactions {
		fmt.Printf("ID: %s\n", r.ID)
		fmt.Printf("  User: %s\n", r.UserID)
		fmt.Printf("  Emoji: %s\n\n", r.Reaction)
	}

	return nil
}

type ChatReactCmd struct {
	MessageID string `arg:"" help:"Message ID" required:""`
	Emoji     string `name:"emoji" short:"e" help:"Emoji reaction" required:""`
}

func (cmd *ChatReactCmd) Run(ctx context.Context) error {
	client, err := getClickUpClient(ctx)
	if err != nil {
		return err
	}

	result, err := client.Chat().CreateReaction(ctx, cmd.MessageID, clickup.CreateReactionRequest{Reaction: cmd.Emoji})
	if err != nil {
		return err
	}

	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(os.Stdout, result)
	}

	fmt.Fprintf(os.Stderr, "Added reaction: %s\n", result.Reaction)
	fmt.Printf("ID: %s\n", result.ID)

	return nil
}

type ChatUnreactCmd struct {
	MessageID  string `arg:"" help:"Message ID" required:""`
	ReactionID string `arg:"" help:"Reaction ID" required:""`
}

func (cmd *ChatUnreactCmd) Run(ctx context.Context) error {
	client, err := getClickUpClient(ctx)
	if err != nil {
		return err
	}

	err = client.Chat().DeleteReaction(ctx, cmd.MessageID, cmd.ReactionID)
	if err != nil {
		return err
	}

	fmt.Fprintf(os.Stderr, "Removed reaction: %s\n", cmd.ReactionID)

	return nil
}

type ChatRepliesCmd struct {
	MessageID string `arg:"" help:"Message ID" required:""`
}

func (cmd *ChatRepliesCmd) Run(ctx context.Context) error {
	client, err := getClickUpClient(ctx)
	if err != nil {
		return err
	}

	result, err := client.Chat().ListReplies(ctx, cmd.MessageID)
	if err != nil {
		return err
	}

	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(os.Stdout, result)
	}
	if outfmt.IsPlain(ctx) {
		headers := []string{"ID", "USER_ID", "CONTENT"}
		rows := make([][]string, 0, len(result.Data))
		for _, r := range result.Data {
			rows = append(rows, []string{r.ID, r.UserID, r.Content})
		}
		return outfmt.WritePlain(os.Stdout, headers, rows)
	}

	if len(result.Data) == 0 {
		fmt.Fprintln(os.Stderr, "No replies found")
		return nil
	}

	fmt.Fprintf(os.Stderr, "Found %d replies\n\n", len(result.Data))

	for _, r := range result.Data {
		fmt.Printf("ID: %s\n", r.ID)
		fmt.Printf("  User: %s\n", r.UserID)
		fmt.Printf("  Content: %s\n\n", r.Content)
	}

	return nil
}

type ChatReplyCmd struct {
	MessageID string `arg:"" help:"Message ID" required:""`
	Text      string `name:"text" short:"t" help:"Reply text" required:""`
}

func (cmd *ChatReplyCmd) Run(ctx context.Context) error {
	client, err := getClickUpClient(ctx)
	if err != nil {
		return err
	}

	result, err := client.Chat().CreateReply(ctx, cmd.MessageID, clickup.SendMessageRequest{Content: cmd.Text})
	if err != nil {
		return err
	}

	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(os.Stdout, result)
	}

	fmt.Fprintf(os.Stderr, "Created reply\n")
	fmt.Printf("ID: %s\n", result.ID)

	return nil
}

type ChatTaggedUsersCmd struct {
	MessageID string `arg:"" help:"Message ID" required:""`
}

func (cmd *ChatTaggedUsersCmd) Run(ctx context.Context) error {
	client, err := getClickUpClient(ctx)
	if err != nil {
		return err
	}

	result, err := client.Chat().GetTaggedUsers(ctx, cmd.MessageID)
	if err != nil {
		return err
	}

	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(os.Stdout, result)
	}
	if outfmt.IsPlain(ctx) {
		headers := []string{"ID", "USERNAME", "EMAIL"}
		rows := make([][]string, 0, len(result.Users))
		for _, u := range result.Users {
			rows = append(rows, []string{fmt.Sprintf("%d", u.ID), u.Username, u.Email})
		}
		return outfmt.WritePlain(os.Stdout, headers, rows)
	}

	if len(result.Users) == 0 {
		fmt.Fprintln(os.Stderr, "No tagged users found")
		return nil
	}

	fmt.Fprintf(os.Stderr, "Found %d tagged users\n\n", len(result.Users))

	for _, u := range result.Users {
		fmt.Printf("ID: %d\n", u.ID)
		fmt.Printf("  Username: %s\n", u.Username)

		if u.Email != "" {
			fmt.Printf("  Email: %s\n", u.Email)
		}

		fmt.Println()
	}

	return nil
}

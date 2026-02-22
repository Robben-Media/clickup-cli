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
	Channels         ChatChannelsCmd         `cmd:"" help:"List chat channels"`
	Channel          ChatChannelCmd          `cmd:"" help:"Get a channel by ID"`
	ChannelFollowers ChatChannelFollowersCmd `cmd:"" help:"Get channel followers"`
	ChannelMembers   ChatChannelMembersCmd   `cmd:"" help:"Get channel members"`
	Messages         ChatMessagesCmd         `cmd:"" help:"List messages in a channel"`
	CreateChannel    ChatCreateChannelCmd    `cmd:"" help:"Create a new channel"`
	CreateDM         ChatCreateDMCmd         `cmd:"" help:"Create a direct message channel"`
	CreateLocation   ChatCreateLocationCmd   `cmd:"" help:"Create a channel in a location"`
	UpdateChannel    ChatUpdateChannelCmd    `cmd:"" help:"Update a channel"`
	DeleteChannel    ChatDeleteChannelCmd    `cmd:"" help:"Delete a channel"`
	Send             ChatSendCmd             `cmd:"" help:"Send a message to a channel"`
	Reactions        ChatReactionsCmd        `cmd:"" help:"List reactions on a message"`
	Replies          ChatRepliesCmd          `cmd:"" help:"List replies to a message"`
	TaggedUsers      ChatTaggedUsersCmd      `cmd:"" help:"Get users tagged in a message"`
	React            ChatReactCmd            `cmd:"" help:"Add a reaction to a message"`
	Reply            ChatReplyCmd            `cmd:"" help:"Reply to a message"`
	UpdateMessage    ChatUpdateMessageCmd    `cmd:"" help:"Update a message"`
	DeleteMessage    ChatDeleteMessageCmd    `cmd:"" help:"Delete a message"`
	DeleteReaction   ChatDeleteReactionCmd   `cmd:"" help:"Remove a reaction"`
}

// ChatChannelsCmd lists all chat channels.
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
		fmt.Printf("%s: %s (%s, %d members)\n", ch.ID, ch.Name, ch.Type, ch.MemberCount)
	}

	return nil
}

// ChatChannelCmd gets a single channel.
type ChatChannelCmd struct {
	ChannelID string `arg:"" required:"" help:"Channel ID"`
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

// ChatChannelFollowersCmd gets channel followers.
type ChatChannelFollowersCmd struct {
	ChannelID string `arg:"" required:"" help:"Channel ID"`
}

func (cmd *ChatChannelFollowersCmd) Run(ctx context.Context) error {
	client, err := getClickUpClient(ctx)
	if err != nil {
		return err
	}

	result, err := client.Chat().GetChannelFollowers(ctx, cmd.ChannelID)
	if err != nil {
		return err
	}

	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(os.Stdout, result)
	}
	if outfmt.IsPlain(ctx) {
		headers := []string{"ID", "USERNAME"}
		rows := make([][]string, 0, len(result.Users))
		for _, u := range result.Users {
			rows = append(rows, []string{fmt.Sprintf("%d", u.ID), u.Username})
		}
		return outfmt.WritePlain(os.Stdout, headers, rows)
	}

	if len(result.Users) == 0 {
		fmt.Fprintln(os.Stderr, "No followers found")
		return nil
	}

	fmt.Fprintf(os.Stderr, "Found %d followers\n\n", len(result.Users))
	for _, u := range result.Users {
		fmt.Printf("%d: %s\n", u.ID, u.Username)
	}

	return nil
}

// ChatChannelMembersCmd gets channel members.
type ChatChannelMembersCmd struct {
	ChannelID string `arg:"" required:"" help:"Channel ID"`
}

func (cmd *ChatChannelMembersCmd) Run(ctx context.Context) error {
	client, err := getClickUpClient(ctx)
	if err != nil {
		return err
	}

	result, err := client.Chat().GetChannelMembers(ctx, cmd.ChannelID)
	if err != nil {
		return err
	}

	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(os.Stdout, result)
	}
	if outfmt.IsPlain(ctx) {
		headers := []string{"ID", "USERNAME"}
		rows := make([][]string, 0, len(result.Users))
		for _, u := range result.Users {
			rows = append(rows, []string{fmt.Sprintf("%d", u.ID), u.Username})
		}
		return outfmt.WritePlain(os.Stdout, headers, rows)
	}

	if len(result.Users) == 0 {
		fmt.Fprintln(os.Stderr, "No members found")
		return nil
	}

	fmt.Fprintf(os.Stderr, "Found %d members\n\n", len(result.Users))
	for _, u := range result.Users {
		fmt.Printf("%d: %s\n", u.ID, u.Username)
	}

	return nil
}

// ChatMessagesCmd lists messages in a channel.
type ChatMessagesCmd struct {
	ChannelID string `arg:"" required:"" help:"Channel ID"`
	Limit     int    `help:"Maximum number of messages to return"`
	Cursor    string `help:"Pagination cursor"`
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
		for _, m := range result.Data {
			content := m.Content
			if len(content) > 50 {
				content = content[:47] + "..."
			}
			rows = append(rows, []string{m.ID, m.UserID, m.Type, string(m.DateCreated), content, fmt.Sprintf("%d", m.RepliesCount)})
		}
		return outfmt.WritePlain(os.Stdout, headers, rows)
	}

	if len(result.Data) == 0 {
		fmt.Fprintln(os.Stderr, "No messages found")
		return nil
	}

	fmt.Fprintf(os.Stderr, "Found %d messages\n\n", len(result.Data))
	for _, m := range result.Data {
		fmt.Printf("ID: %s\n", m.ID)
		fmt.Printf("  User: %s\n", m.UserID)
		fmt.Printf("  Date: %s\n", m.DateCreated)
		fmt.Printf("  Content: %s\n", m.Content)
		if m.RepliesCount > 0 {
			fmt.Printf("  Replies: %d\n", m.RepliesCount)
		}
		fmt.Println()
	}

	if result.Pagination != nil && result.Pagination.NextPageToken != "" {
		fmt.Fprintf(os.Stderr, "Next page cursor: %s\n", result.Pagination.NextPageToken)
	}

	return nil
}

// ChatCreateChannelCmd creates a new channel.
type ChatCreateChannelCmd struct {
	Name string `required:"" help:"Channel name"`
}

func (cmd *ChatCreateChannelCmd) Run(ctx context.Context) error {
	client, err := getClickUpClient(ctx)
	if err != nil {
		return err
	}

	req := clickup.CreateChatChannelRequest{Name: cmd.Name}
	result, err := client.Chat().CreateChannel(ctx, req)
	if err != nil {
		return err
	}

	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(os.Stdout, result)
	}
	if outfmt.IsPlain(ctx) {
		headers := []string{"ID", "NAME", "TYPE"}
		rows := [][]string{{result.ID, result.Name, result.Type}}
		return outfmt.WritePlain(os.Stdout, headers, rows)
	}

	fmt.Fprintf(os.Stderr, "Created channel\n\n")
	fmt.Printf("ID: %s\n", result.ID)
	fmt.Printf("Name: %s\n", result.Name)
	fmt.Printf("Type: %s\n", result.Type)

	return nil
}

// ChatCreateDMCmd creates a direct message channel.
type ChatCreateDMCmd struct {
	Members string `required:"" help:"Comma-separated user IDs"`
}

func (cmd *ChatCreateDMCmd) Run(ctx context.Context) error {
	client, err := getClickUpClient(ctx)
	if err != nil {
		return err
	}

	members := strings.Split(cmd.Members, ",")
	for i, m := range members {
		members[i] = strings.TrimSpace(m)
	}

	req := clickup.CreateDMRequest{Members: members}
	result, err := client.Chat().CreateDirectMessage(ctx, req)
	if err != nil {
		return err
	}

	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(os.Stdout, result)
	}
	if outfmt.IsPlain(ctx) {
		headers := []string{"ID", "TYPE"}
		rows := [][]string{{result.ID, result.Type}}
		return outfmt.WritePlain(os.Stdout, headers, rows)
	}

	fmt.Fprintf(os.Stderr, "Created direct message channel\n\n")
	fmt.Printf("ID: %s\n", result.ID)
	fmt.Printf("Type: %s\n", result.Type)

	return nil
}

// ChatCreateLocationCmd creates a channel in a location.
type ChatCreateLocationCmd struct {
	Name       string `required:"" help:"Channel name"`
	ParentType string `required:"" help:"Parent type (space, folder, list)"`
	ParentID   string `required:"" help:"Parent ID"`
}

func (cmd *ChatCreateLocationCmd) Run(ctx context.Context) error {
	client, err := getClickUpClient(ctx)
	if err != nil {
		return err
	}

	req := clickup.CreateLocationChannelRequest{
		Name:       cmd.Name,
		ParentType: cmd.ParentType,
		ParentID:   cmd.ParentID,
	}
	result, err := client.Chat().CreateLocationChannel(ctx, req)
	if err != nil {
		return err
	}

	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(os.Stdout, result)
	}
	if outfmt.IsPlain(ctx) {
		headers := []string{"ID", "NAME", "TYPE"}
		rows := [][]string{{result.ID, result.Name, result.Type}}
		return outfmt.WritePlain(os.Stdout, headers, rows)
	}

	fmt.Fprintf(os.Stderr, "Created location channel\n\n")
	fmt.Printf("ID: %s\n", result.ID)
	fmt.Printf("Name: %s\n", result.Name)
	fmt.Printf("Type: %s\n", result.Type)

	return nil
}

// ChatUpdateChannelCmd updates a channel.
type ChatUpdateChannelCmd struct {
	ChannelID string `arg:"" required:"" help:"Channel ID"`
	Name      string `help:"New channel name"`
}

func (cmd *ChatUpdateChannelCmd) Run(ctx context.Context) error {
	client, err := getClickUpClient(ctx)
	if err != nil {
		return err
	}

	req := clickup.UpdateChannelRequest{Name: cmd.Name}
	result, err := client.Chat().UpdateChannel(ctx, cmd.ChannelID, req)
	if err != nil {
		return err
	}

	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(os.Stdout, result)
	}
	if outfmt.IsPlain(ctx) {
		headers := []string{"ID", "NAME", "TYPE"}
		rows := [][]string{{result.ID, result.Name, result.Type}}
		return outfmt.WritePlain(os.Stdout, headers, rows)
	}

	fmt.Fprintf(os.Stderr, "Updated channel\n\n")
	fmt.Printf("ID: %s\n", result.ID)
	fmt.Printf("Name: %s\n", result.Name)

	return nil
}

// ChatDeleteChannelCmd deletes a channel.
type ChatDeleteChannelCmd struct {
	ChannelID string `arg:"" required:"" help:"Channel ID"`
}

func (cmd *ChatDeleteChannelCmd) Run(ctx context.Context) error {
	client, err := getClickUpClient(ctx)
	if err != nil {
		return err
	}

	if err := client.Chat().DeleteChannel(ctx, cmd.ChannelID); err != nil {
		return err
	}

	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(os.Stdout, map[string]string{
			"status":     "success",
			"message":    "Channel deleted",
			"channel_id": cmd.ChannelID,
		})
	}
	if outfmt.IsPlain(ctx) {
		headers := []string{"STATUS", "CHANNEL_ID"}
		rows := [][]string{{"success", cmd.ChannelID}}
		return outfmt.WritePlain(os.Stdout, headers, rows)
	}

	fmt.Fprintf(os.Stderr, "Channel %s deleted\n", cmd.ChannelID)

	return nil
}

// ChatSendCmd sends a message to a channel.
type ChatSendCmd struct {
	ChannelID string `arg:"" required:"" help:"Channel ID"`
	Text      string `required:"" help:"Message content"`
}

func (cmd *ChatSendCmd) Run(ctx context.Context) error {
	client, err := getClickUpClient(ctx)
	if err != nil {
		return err
	}

	req := clickup.SendMessageRequest{Content: cmd.Text}
	result, err := client.Chat().SendMessage(ctx, cmd.ChannelID, req)
	if err != nil {
		return err
	}

	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(os.Stdout, result)
	}
	if outfmt.IsPlain(ctx) {
		headers := []string{"ID", "CHANNEL_ID", "DATE"}
		rows := [][]string{{result.ID, result.ParentChannel, string(result.DateCreated)}}
		return outfmt.WritePlain(os.Stdout, headers, rows)
	}

	fmt.Fprintf(os.Stderr, "Message sent\n\n")
	fmt.Printf("ID: %s\n", result.ID)
	fmt.Printf("Channel: %s\n", result.ParentChannel)

	return nil
}

// ChatReactionsCmd lists reactions on a message.
type ChatReactionsCmd struct {
	MessageID string `arg:"" required:"" help:"Message ID"`
}

func (cmd *ChatReactionsCmd) Run(ctx context.Context) error {
	client, err := getClickUpClient(ctx)
	if err != nil {
		return err
	}

	result, err := client.Chat().GetReactions(ctx, cmd.MessageID)
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
		fmt.Printf("%s: %s by user %s\n", r.ID, r.Reaction, r.UserID)
	}

	return nil
}

// ChatRepliesCmd lists replies to a message.
type ChatRepliesCmd struct {
	MessageID string `arg:"" required:"" help:"Message ID"`
}

func (cmd *ChatRepliesCmd) Run(ctx context.Context) error {
	client, err := getClickUpClient(ctx)
	if err != nil {
		return err
	}

	result, err := client.Chat().GetReplies(ctx, cmd.MessageID)
	if err != nil {
		return err
	}

	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(os.Stdout, result)
	}
	if outfmt.IsPlain(ctx) {
		headers := []string{"ID", "USER_ID", "DATE", "CONTENT"}
		rows := make([][]string, 0, len(result.Data))
		for _, m := range result.Data {
			content := m.Content
			if len(content) > 50 {
				content = content[:47] + "..."
			}
			rows = append(rows, []string{m.ID, m.UserID, string(m.DateCreated), content})
		}
		return outfmt.WritePlain(os.Stdout, headers, rows)
	}

	if len(result.Data) == 0 {
		fmt.Fprintln(os.Stderr, "No replies found")
		return nil
	}

	fmt.Fprintf(os.Stderr, "Found %d replies\n\n", len(result.Data))
	for _, m := range result.Data {
		fmt.Printf("ID: %s\n", m.ID)
		fmt.Printf("  User: %s\n", m.UserID)
		fmt.Printf("  Date: %s\n", m.DateCreated)
		fmt.Printf("  Content: %s\n\n", m.Content)
	}

	return nil
}

// ChatTaggedUsersCmd gets users tagged in a message.
type ChatTaggedUsersCmd struct {
	MessageID string `arg:"" required:"" help:"Message ID"`
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
		headers := []string{"ID", "USERNAME"}
		rows := make([][]string, 0, len(result.Users))
		for _, u := range result.Users {
			rows = append(rows, []string{fmt.Sprintf("%d", u.ID), u.Username})
		}
		return outfmt.WritePlain(os.Stdout, headers, rows)
	}

	if len(result.Users) == 0 {
		fmt.Fprintln(os.Stderr, "No tagged users found")
		return nil
	}

	fmt.Fprintf(os.Stderr, "Found %d tagged users\n\n", len(result.Users))
	for _, u := range result.Users {
		fmt.Printf("%d: %s\n", u.ID, u.Username)
	}

	return nil
}

// ChatReactCmd adds a reaction to a message.
type ChatReactCmd struct {
	MessageID string `arg:"" required:"" help:"Message ID"`
	Emoji     string `required:"" help:"Emoji reaction"`
}

func (cmd *ChatReactCmd) Run(ctx context.Context) error {
	client, err := getClickUpClient(ctx)
	if err != nil {
		return err
	}

	req := clickup.CreateReactionRequest{Reaction: cmd.Emoji}
	result, err := client.Chat().CreateReaction(ctx, cmd.MessageID, req)
	if err != nil {
		return err
	}

	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(os.Stdout, result)
	}
	if outfmt.IsPlain(ctx) {
		headers := []string{"ID", "MESSAGE_ID", "EMOJI"}
		rows := [][]string{{result.ID, result.MessageID, result.Reaction}}
		return outfmt.WritePlain(os.Stdout, headers, rows)
	}

	fmt.Fprintf(os.Stderr, "Reaction added\n\n")
	fmt.Printf("ID: %s\n", result.ID)
	fmt.Printf("Emoji: %s\n", result.Reaction)

	return nil
}

// ChatReplyCmd creates a reply to a message.
type ChatReplyCmd struct {
	MessageID string `arg:"" required:"" help:"Message ID"`
	Text      string `required:"" help:"Reply content"`
}

func (cmd *ChatReplyCmd) Run(ctx context.Context) error {
	client, err := getClickUpClient(ctx)
	if err != nil {
		return err
	}

	req := clickup.SendMessageRequest{Content: cmd.Text}
	result, err := client.Chat().CreateReply(ctx, cmd.MessageID, req)
	if err != nil {
		return err
	}

	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(os.Stdout, result)
	}
	if outfmt.IsPlain(ctx) {
		headers := []string{"ID", "PARENT_MESSAGE", "DATE"}
		rows := [][]string{{result.ID, result.ParentMessage, string(result.DateCreated)}}
		return outfmt.WritePlain(os.Stdout, headers, rows)
	}

	fmt.Fprintf(os.Stderr, "Reply sent\n\n")
	fmt.Printf("ID: %s\n", result.ID)
	fmt.Printf("Parent: %s\n", result.ParentMessage)

	return nil
}

// ChatUpdateMessageCmd updates a message.
type ChatUpdateMessageCmd struct {
	MessageID string `arg:"" required:"" help:"Message ID"`
	Text      string `help:"New message content"`
}

func (cmd *ChatUpdateMessageCmd) Run(ctx context.Context) error {
	client, err := getClickUpClient(ctx)
	if err != nil {
		return err
	}

	req := clickup.UpdateMessageRequest{Content: cmd.Text}
	result, err := client.Chat().UpdateMessage(ctx, cmd.MessageID, req)
	if err != nil {
		return err
	}

	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(os.Stdout, result)
	}
	if outfmt.IsPlain(ctx) {
		headers := []string{"ID", "DATE"}
		rows := [][]string{{result.ID, string(result.DateUpdated)}}
		return outfmt.WritePlain(os.Stdout, headers, rows)
	}

	fmt.Fprintf(os.Stderr, "Message updated\n\n")
	fmt.Printf("ID: %s\n", result.ID)

	return nil
}

// ChatDeleteMessageCmd deletes a message.
type ChatDeleteMessageCmd struct {
	MessageID string `arg:"" required:"" help:"Message ID"`
}

func (cmd *ChatDeleteMessageCmd) Run(ctx context.Context) error {
	client, err := getClickUpClient(ctx)
	if err != nil {
		return err
	}

	if err := client.Chat().DeleteMessage(ctx, cmd.MessageID); err != nil {
		return err
	}

	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(os.Stdout, map[string]string{
			"status":     "success",
			"message":    "Message deleted",
			"message_id": cmd.MessageID,
		})
	}
	if outfmt.IsPlain(ctx) {
		headers := []string{"STATUS", "MESSAGE_ID"}
		rows := [][]string{{"success", cmd.MessageID}}
		return outfmt.WritePlain(os.Stdout, headers, rows)
	}

	fmt.Fprintf(os.Stderr, "Message %s deleted\n", cmd.MessageID)

	return nil
}

// ChatDeleteReactionCmd removes a reaction.
type ChatDeleteReactionCmd struct {
	MessageID  string `arg:"" required:"" help:"Message ID"`
	ReactionID string `arg:"" required:"" help:"Reaction ID"`
}

func (cmd *ChatDeleteReactionCmd) Run(ctx context.Context) error {
	client, err := getClickUpClient(ctx)
	if err != nil {
		return err
	}

	if err := client.Chat().DeleteReaction(ctx, cmd.MessageID, cmd.ReactionID); err != nil {
		return err
	}

	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(os.Stdout, map[string]string{
			"status":      "success",
			"message":     "Reaction deleted",
			"reaction_id": cmd.ReactionID,
		})
	}
	if outfmt.IsPlain(ctx) {
		headers := []string{"STATUS", "REACTION_ID"}
		rows := [][]string{{"success", cmd.ReactionID}}
		return outfmt.WritePlain(os.Stdout, headers, rows)
	}

	fmt.Fprintf(os.Stderr, "Reaction %s removed\n", cmd.ReactionID)

	return nil
}

package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/builtbyrobben/clickup-cli/internal/clickup"
	"github.com/builtbyrobben/clickup-cli/internal/outfmt"
)

type UsersCmd struct {
	Get    UsersGetCmd    `cmd:"" help:"Get a user by ID"`
	Invite UsersInviteCmd `cmd:"" help:"Invite a user to workspace by email"`
	Update UsersUpdateCmd `cmd:"" help:"Update a user on workspace"`
	Remove UsersRemoveCmd `cmd:"" help:"Remove a user from workspace"`
}

type UsersGetCmd struct {
	TeamID string `required:"" help:"Team/workspace ID"`
	UserID string `arg:"" required:"" help:"User ID"`
}

func (cmd *UsersGetCmd) Run(ctx context.Context) error {
	client, err := getClickUpClient(ctx)
	if err != nil {
		return err
	}

	result, err := client.Users().Get(ctx, cmd.TeamID, cmd.UserID)
	if err != nil {
		return err
	}

	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(os.Stdout, result)
	}

	if outfmt.IsPlain(ctx) {
		headers := []string{"ID", "USERNAME", "EMAIL", "ROLE"}
		rows := [][]string{{
			fmt.Sprintf("%d", result.User.ID),
			result.User.Username,
			result.User.Email,
			fmt.Sprintf("%d", result.User.Role),
		}}
		return outfmt.WritePlain(os.Stdout, headers, rows)
	}

	fmt.Printf("User %d\n", result.User.ID)
	fmt.Printf("  Username: %s\n", result.User.Username)
	fmt.Printf("  Email: %s\n", result.User.Email)
	fmt.Printf("  Role: %s (%d)\n", roleName(result.User.Role), result.User.Role)

	return nil
}

type UsersInviteCmd struct {
	TeamID string `required:"" help:"Team/workspace ID"`
	Email  string `arg:"" required:"" help:"Email address to invite"`
	Admin  bool   `help:"Invite as admin"`
}

func (cmd *UsersInviteCmd) Run(ctx context.Context) error {
	client, err := getClickUpClient(ctx)
	if err != nil {
		return err
	}

	req := clickup.InviteUserRequest{
		Email: cmd.Email,
		Admin: cmd.Admin,
	}

	result, err := client.Users().Invite(ctx, cmd.TeamID, req)
	if err != nil {
		return err
	}

	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(os.Stdout, result)
	}

	if outfmt.IsPlain(ctx) {
		headers := []string{"ID", "EMAIL", "USERNAME"}
		username := ""
		if result.User.Username != "" {
			username = result.User.Username
		}
		rows := [][]string{{
			fmt.Sprintf("%d", result.User.ID),
			result.User.Email,
			username,
		}}
		return outfmt.WritePlain(os.Stdout, headers, rows)
	}

	fmt.Fprintf(os.Stderr, "Invited user\n\n")
	fmt.Printf("User %d\n", result.User.ID)
	fmt.Printf("  Email: %s\n", result.User.Email)

	if result.User.Username != "" {
		fmt.Printf("  Username: %s\n", result.User.Username)
	}

	return nil
}

type UsersUpdateCmd struct {
	TeamID   string `required:"" help:"Team/workspace ID"`
	UserID   string `arg:"" required:"" help:"User ID"`
	Username string `help:"New username"`
	Admin    bool   `help:"Set as admin"`
}

func (cmd *UsersUpdateCmd) Run(ctx context.Context) error {
	client, err := getClickUpClient(ctx)
	if err != nil {
		return err
	}

	req := clickup.EditUserRequest{
		Username: cmd.Username,
		Admin:    cmd.Admin,
	}

	result, err := client.Users().Update(ctx, cmd.TeamID, cmd.UserID, req)
	if err != nil {
		return err
	}

	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(os.Stdout, result)
	}

	if outfmt.IsPlain(ctx) {
		headers := []string{"ID", "USERNAME", "EMAIL", "ROLE"}
		rows := [][]string{{
			fmt.Sprintf("%d", result.User.ID),
			result.User.Username,
			result.User.Email,
			fmt.Sprintf("%d", result.User.Role),
		}}
		return outfmt.WritePlain(os.Stdout, headers, rows)
	}

	fmt.Fprintf(os.Stderr, "Updated user\n\n")
	fmt.Printf("User %d\n", result.User.ID)
	fmt.Printf("  Username: %s\n", result.User.Username)
	fmt.Printf("  Email: %s\n", result.User.Email)
	fmt.Printf("  Role: %s (%d)\n", roleName(result.User.Role), result.User.Role)

	return nil
}

type UsersRemoveCmd struct {
	TeamID string `required:"" help:"Team/workspace ID"`
	UserID string `arg:"" required:"" help:"User ID"`
}

func (cmd *UsersRemoveCmd) Run(ctx context.Context) error {
	client, err := getClickUpClient(ctx)
	if err != nil {
		return err
	}

	if err := client.Users().Remove(ctx, cmd.TeamID, cmd.UserID); err != nil {
		return err
	}

	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(os.Stdout, map[string]string{
			"status":  "success",
			"message": "User removed",
			"user_id": cmd.UserID,
		})
	}

	if outfmt.IsPlain(ctx) {
		headers := []string{"STATUS", "USER_ID"}
		rows := [][]string{{"success", cmd.UserID}}
		return outfmt.WritePlain(os.Stdout, headers, rows)
	}

	fmt.Fprintf(os.Stderr, "User %s removed\n", cmd.UserID)

	return nil
}

func roleName(role int) string {
	switch role {
	case 1:
		return "Owner"
	case 2:
		return "Admin"
	case 3:
		return "Member"
	case 4:
		return "Guest"
	default:
		return "Unknown"
	}
}

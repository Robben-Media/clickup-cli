package cmd

import (
	"context"
	"fmt"
	"os"
	"strconv"

	"github.com/builtbyrobben/clickup-cli/internal/clickup"
	"github.com/builtbyrobben/clickup-cli/internal/outfmt"
)

type UsersCmd struct {
	Get    UsersGetCmd    `cmd:"" help:"Get user details"`
	Invite UsersInviteCmd `cmd:"" help:"Invite user to workspace"`
	Update UsersUpdateCmd `cmd:"" help:"Update user role"`
	Remove UsersRemoveCmd `cmd:"" help:"Remove user from workspace"`
}

type UsersGetCmd struct {
	TeamID string `arg:"" required:"" help:"Team (workspace) ID"`
	UserID int    `arg:"" required:"" help:"User ID"`
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
			strconv.Itoa(result.ID),
			result.Username,
			result.Email,
			formatRole(result.Role),
		}}
		return outfmt.WritePlain(os.Stdout, headers, rows)
	}

	printUserDetail(result)
	return nil
}

type UsersInviteCmd struct {
	TeamID string `arg:"" required:"" help:"Team (workspace) ID"`
	Email  string `required:"" help:"Email address to invite"`
	Admin  bool   `help:"Grant admin role"`
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
		headers := []string{"ID", "USERNAME", "EMAIL", "ROLE"}
		rows := [][]string{{
			strconv.Itoa(result.ID),
			result.Username,
			result.Email,
			formatRole(result.Role),
		}}
		return outfmt.WritePlain(os.Stdout, headers, rows)
	}

	fmt.Fprintf(os.Stderr, "User invited successfully\n\n")
	printUserDetail(result)
	return nil
}

type UsersUpdateCmd struct {
	TeamID   string `arg:"" required:"" help:"Team (workspace) ID"`
	UserID   int    `arg:"" required:"" help:"User ID"`
	Username string `help:"New username"`
	Admin    bool   `help:"Grant admin role (set to false to remove admin)"`
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
			strconv.Itoa(result.ID),
			result.Username,
			result.Email,
			formatRole(result.Role),
		}}
		return outfmt.WritePlain(os.Stdout, headers, rows)
	}

	fmt.Fprintf(os.Stderr, "User updated successfully\n\n")
	printUserDetail(result)
	return nil
}

type UsersRemoveCmd struct {
	TeamID string `arg:"" required:"" help:"Team (workspace) ID"`
	UserID int    `arg:"" required:"" help:"User ID"`
	Force  bool   `help:"Skip confirmation"`
}

func (cmd *UsersRemoveCmd) Run(ctx context.Context) error {
	if !cmd.Force {
		fmt.Fprintf(os.Stderr, "Warning: This will remove user %d from workspace %s.\n", cmd.UserID, cmd.TeamID)
		fmt.Fprint(os.Stderr, "Use --force to confirm.\n")
		return nil
	}

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
			"user_id": strconv.Itoa(cmd.UserID),
		})
	}

	if outfmt.IsPlain(ctx) {
		headers := []string{"STATUS", "USER_ID"}
		rows := [][]string{{"success", strconv.Itoa(cmd.UserID)}}
		return outfmt.WritePlain(os.Stdout, headers, rows)
	}

	fmt.Fprintf(os.Stderr, "User %d removed from workspace.\n", cmd.UserID)
	return nil
}

func printUserDetail(user *clickup.UserDetail) {
	fmt.Fprintf(os.Stderr, "User %d\n", user.ID)
	fmt.Printf("  Username: %s\n", user.Username)
	fmt.Printf("  Email: %s\n", user.Email)
	fmt.Printf("  Role: %s (%d)\n", roleName(user.Role), user.Role)

	if user.InvitedBy != nil {
		fmt.Printf("  Invited by: %s (%d)\n", user.InvitedBy.Username, user.InvitedBy.ID)
	}
}

func formatRole(role int) string {
	return fmt.Sprintf("%d", role)
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

package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/builtbyrobben/clickup-cli/internal/outfmt"
)

type WorkspacesCmd struct {
	List  WorkspacesListCmd  `cmd:"" help:"List all authorized workspaces"`
	Plan  WorkspacesPlanCmd  `cmd:"" help:"Get workspace plan"`
	Seats WorkspacesSeatsCmd `cmd:"" help:"Get workspace seat usage"`
}

type WorkspacesListCmd struct{}

func (cmd *WorkspacesListCmd) Run(ctx context.Context) error {
	client, err := getClickUpClient(ctx)
	if err != nil {
		return err
	}

	result, err := client.Workspaces().List(ctx)
	if err != nil {
		return err
	}

	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(os.Stdout, result)
	}
	if outfmt.IsPlain(ctx) {
		headers := []string{"ID", "NAME", "MEMBER_COUNT"}
		var rows [][]string
		for _, team := range result.Teams {
			rows = append(rows, []string{team.ID, team.Name, fmt.Sprintf("%d", len(team.Members))})
		}
		return outfmt.WritePlain(os.Stdout, headers, rows)
	}

	if len(result.Teams) == 0 {
		fmt.Fprintln(os.Stderr, "No workspaces found")
		return nil
	}

	fmt.Fprintf(os.Stderr, "Workspaces\n\n")

	for _, team := range result.Teams {
		fmt.Printf("  %s: %s (%d members)\n", team.ID, team.Name, len(team.Members))
	}

	return nil
}

type WorkspacesPlanCmd struct {
	TeamID string `required:"" help:"Team ID (workspace ID)"`
}

func (cmd *WorkspacesPlanCmd) Run(ctx context.Context) error {
	client, err := getClickUpClient(ctx)
	if err != nil {
		return err
	}

	result, err := client.Workspaces().Plan(ctx, cmd.TeamID)
	if err != nil {
		return err
	}

	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(os.Stdout, result)
	}
	if outfmt.IsPlain(ctx) {
		headers := []string{"TEAM_ID", "PLAN_ID", "PLAN_NAME"}
		rows := [][]string{{result.TeamID, fmt.Sprintf("%d", result.PlanID), result.PlanName}}
		return outfmt.WritePlain(os.Stdout, headers, rows)
	}

	fmt.Fprintf(os.Stderr, "Workspace Plan\n\n")
	fmt.Printf("  Team ID: %s\n", result.TeamID)
	fmt.Printf("  Plan ID: %d\n", result.PlanID)
	fmt.Printf("  Plan Name: %s\n", result.PlanName)

	return nil
}

type WorkspacesSeatsCmd struct {
	TeamID string `required:"" help:"Team ID (workspace ID)"`
}

func (cmd *WorkspacesSeatsCmd) Run(ctx context.Context) error {
	client, err := getClickUpClient(ctx)
	if err != nil {
		return err
	}

	result, err := client.Workspaces().Seats(ctx, cmd.TeamID)
	if err != nil {
		return err
	}

	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(os.Stdout, result)
	}
	if outfmt.IsPlain(ctx) {
		headers := []string{"TYPE", "FILLED", "TOTAL", "EMPTY"}
		var rows [][]string
		rows = append(rows, []string{
			"members",
			fmt.Sprintf("%d", result.Members.FilledSeats),
			fmt.Sprintf("%d", result.Members.TotalSeats),
			fmt.Sprintf("%d", result.Members.EmptySeats),
		})
		rows = append(rows, []string{
			"guests",
			fmt.Sprintf("%d", result.Guests.FilledSeats),
			fmt.Sprintf("%d", result.Guests.TotalSeats),
			fmt.Sprintf("%d", result.Guests.EmptySeats),
		})
		return outfmt.WritePlain(os.Stdout, headers, rows)
	}

	fmt.Fprintf(os.Stderr, "Workspace Seats\n\n")
	fmt.Printf("  Members:\n")
	fmt.Printf("    Filled: %d\n", result.Members.FilledSeats)
	fmt.Printf("    Total: %d\n", result.Members.TotalSeats)
	fmt.Printf("    Empty: %d\n", result.Members.EmptySeats)
	fmt.Printf("  Guests:\n")
	fmt.Printf("    Filled: %d\n", result.Guests.FilledSeats)
	fmt.Printf("    Total: %d\n", result.Guests.TotalSeats)
	fmt.Printf("    Empty: %d\n", result.Guests.EmptySeats)

	return nil
}

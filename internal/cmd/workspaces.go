package cmd

import (
	"context"
	"fmt"
	"os"
	"strconv"

	"github.com/builtbyrobben/clickup-cli/internal/clickup"
	"github.com/builtbyrobben/clickup-cli/internal/outfmt"
)

type WorkspacesCmd struct {
	List  WorkspacesListCmd  `cmd:"" help:"List workspaces (teams)"`
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
			rows = append(rows, []string{team.ID, team.Name, strconv.Itoa(len(team.Members))})
		}
		return outfmt.WritePlain(os.Stdout, headers, rows)
	}

	if len(result.Teams) == 0 {
		fmt.Fprintln(os.Stderr, "No workspaces found")
		return nil
	}

	fmt.Fprintf(os.Stderr, "Found %d workspaces\n\n", len(result.Teams))

	for _, team := range result.Teams {
		printWorkspace(&team)
	}

	return nil
}

type WorkspacesPlanCmd struct {
	TeamID string `arg:"" required:"" help:"Team (workspace) ID"`
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
		rows := [][]string{{result.TeamID, strconv.Itoa(result.PlanID), result.PlanName}}
		return outfmt.WritePlain(os.Stdout, headers, rows)
	}

	fmt.Fprintf(os.Stderr, "Workspace Plan\n\n")
	fmt.Printf("Team ID: %s\n", result.TeamID)
	fmt.Printf("Plan: %s (ID: %d)\n", result.PlanName, result.PlanID)

	return nil
}

type WorkspacesSeatsCmd struct {
	TeamID string `arg:"" required:"" help:"Team (workspace) ID"`
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
		rows := make([][]string, 0, 2)
		rows = append(rows, []string{
			"members",
			strconv.Itoa(result.Members.FilledSeats),
			strconv.Itoa(result.Members.TotalSeats),
			strconv.Itoa(result.Members.EmptySeats),
		})
		rows = append(rows, []string{
			"guests",
			strconv.Itoa(result.Guests.FilledSeats),
			strconv.Itoa(result.Guests.TotalSeats),
			strconv.Itoa(result.Guests.EmptySeats),
		})
		return outfmt.WritePlain(os.Stdout, headers, rows)
	}

	fmt.Fprintf(os.Stderr, "Workspace Seat Usage\n\n")
	fmt.Printf("Members: %d / %d (%d empty)\n",
		result.Members.FilledSeats,
		result.Members.TotalSeats,
		result.Members.EmptySeats)
	fmt.Printf("Guests:  %d / %d (%d empty)\n",
		result.Guests.FilledSeats,
		result.Guests.TotalSeats,
		result.Guests.EmptySeats)

	return nil
}

func printWorkspace(team *clickup.Workspace) {
	fmt.Printf("%s: %s (%d members)\n", team.ID, team.Name, len(team.Members))
}

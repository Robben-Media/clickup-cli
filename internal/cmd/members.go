package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/builtbyrobben/clickup-cli/internal/outfmt"
)

type MembersCmd struct {
	List MembersListCmd `cmd:"" help:"List all team members"`
}

type MembersListCmd struct{}

func (cmd *MembersListCmd) Run(ctx context.Context) error {
	client, err := getClickUpClient(ctx)
	if err != nil {
		return err
	}

	teamID, err := getTeamID()
	if err != nil {
		return err
	}

	result, err := client.Members().List(ctx, teamID)
	if err != nil {
		return err
	}

	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(os.Stdout, result)
	}
	if outfmt.IsPlain(ctx) {
		headers := []string{"ID", "USERNAME", "EMAIL"}
		var rows [][]string
		for _, member := range result.Members {
			rows = append(rows, []string{fmt.Sprintf("%d", member.User.ID), member.User.Username, member.User.Email})
		}
		return outfmt.WritePlain(os.Stdout, headers, rows)
	}

	if len(result.Members) == 0 {
		fmt.Fprintln(os.Stderr, "No members found")
		return nil
	}

	fmt.Fprintf(os.Stderr, "Found %d members\n\n", len(result.Members))

	for _, member := range result.Members {
		fmt.Printf("ID: %d\n", member.User.ID)
		fmt.Printf("  Username: %s\n", member.User.Username)

		if member.User.Email != "" {
			fmt.Printf("  Email: %s\n", member.User.Email)
		}

		fmt.Println()
	}

	return nil
}

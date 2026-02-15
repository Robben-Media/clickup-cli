package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/builtbyrobben/clickup-cli/internal/outfmt"
)

type SpacesCmd struct {
	List SpacesListCmd `cmd:"" help:"List all spaces in the team"`
}

type SpacesListCmd struct{}

func (cmd *SpacesListCmd) Run(ctx context.Context) error {
	client, err := getClickUpClient()
	if err != nil {
		return err
	}

	teamID, err := getTeamID()
	if err != nil {
		return err
	}

	result, err := client.Spaces().List(ctx, teamID)
	if err != nil {
		return err
	}

	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(os.Stdout, result)
	}

	if len(result.Spaces) == 0 {
		fmt.Fprintln(os.Stderr, "No spaces found")
		return nil
	}

	fmt.Fprintf(os.Stderr, "Found %d spaces\n\n", len(result.Spaces))

	for _, space := range result.Spaces {
		fmt.Printf("ID: %s\n", space.ID)
		fmt.Printf("  Name: %s\n", space.Name)
		fmt.Println()
	}

	return nil
}

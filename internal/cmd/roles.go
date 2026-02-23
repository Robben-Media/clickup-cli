package cmd

import (
	"context"
	"fmt"
	"os"
	"strconv"

	"github.com/builtbyrobben/clickup-cli/internal/clickup"
	"github.com/builtbyrobben/clickup-cli/internal/outfmt"
)

type RolesCmd struct {
	List RolesListCmd `cmd:"" help:"List custom roles"`
}

type RolesListCmd struct {
	TeamID string `arg:"" required:"" help:"Team (workspace) ID"`
}

func (cmd *RolesListCmd) Run(ctx context.Context) error {
	client, err := getClickUpClient(ctx)
	if err != nil {
		return err
	}

	result, err := client.Roles().List(ctx, cmd.TeamID)
	if err != nil {
		return err
	}

	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(os.Stdout, result)
	}
	if outfmt.IsPlain(ctx) {
		headers := []string{"ID", "NAME", "PERMISSION_COUNT"}
		var rows [][]string
		for _, role := range result.CustomRoles {
			rows = append(rows, []string{
				strconv.Itoa(role.ID),
				role.Name,
				strconv.Itoa(len(role.Permissions)),
			})
		}
		return outfmt.WritePlain(os.Stdout, headers, rows)
	}

	if len(result.CustomRoles) == 0 {
		fmt.Fprintln(os.Stderr, "No custom roles found")
		return nil
	}

	fmt.Fprintf(os.Stderr, "Found %d custom roles\n\n", len(result.CustomRoles))

	for _, role := range result.CustomRoles {
		printCustomRole(&role)
	}

	return nil
}

func printCustomRole(role *clickup.CustomRole) {
	fmt.Printf("ID: %d\n", role.ID)
	fmt.Printf("  Name: %s\n", role.Name)
	fmt.Printf("  Permissions: %d\n", len(role.Permissions))
	fmt.Println()
}

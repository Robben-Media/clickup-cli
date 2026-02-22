package cmd

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/builtbyrobben/clickup-cli/internal/clickup"
	"github.com/builtbyrobben/clickup-cli/internal/outfmt"
)

type SpacesCmd struct {
	List   SpacesListCmd   `cmd:"" help:"List all spaces in the team"`
	Get    SpacesGetCmd    `cmd:"" help:"Get space details"`
	Create SpacesCreateCmd `cmd:"" help:"Create a new space"`
	Update SpacesUpdateCmd `cmd:"" help:"Update a space"`
	Delete SpacesDeleteCmd `cmd:"" help:"Delete a space"`
}

type SpacesListCmd struct{}

func (cmd *SpacesListCmd) Run(ctx context.Context) error {
	client, err := getClickUpClient(ctx)
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
	if outfmt.IsPlain(ctx) {
		headers := []string{"ID", "NAME"}
		var rows [][]string
		for _, space := range result.Spaces {
			rows = append(rows, []string{space.ID, space.Name})
		}
		return outfmt.WritePlain(os.Stdout, headers, rows)
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

type SpacesGetCmd struct {
	SpaceID string `arg:"" required:"" help:"Space ID"`
}

func (cmd *SpacesGetCmd) Run(ctx context.Context) error {
	client, err := getClickUpClient(ctx)
	if err != nil {
		return err
	}

	result, err := client.Spaces().Get(ctx, cmd.SpaceID)
	if err != nil {
		return err
	}

	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(os.Stdout, result)
	}
	if outfmt.IsPlain(ctx) {
		headers := []string{"ID", "NAME", "PRIVATE", "STATUS_COUNT"}
		statusCount := strconv.Itoa(len(result.Statuses))
		rows := [][]string{{result.ID, result.Name, strconv.FormatBool(result.Private), statusCount}}
		return outfmt.WritePlain(os.Stdout, headers, rows)
	}

	fmt.Fprintf(os.Stderr, "Space Details\n\n")
	fmt.Printf("ID: %s\n", result.ID)
	fmt.Printf("Name: %s\n", result.Name)
	fmt.Printf("Private: %t\n", result.Private)

	if len(result.Statuses) > 0 {
		var statuses []string
		for _, s := range result.Statuses {
			statuses = append(statuses, s.Status)
		}

		fmt.Printf("Statuses: %s\n", strings.Join(statuses, ", "))
	}

	if result.Features.DueDates.Enabled || result.Features.TimeTracking.Enabled ||
		result.Features.Tags.Enabled || result.Features.Checklists.Enabled {
		var features []string
		if result.Features.DueDates.Enabled {
			features = append(features, "due_dates")
		}

		if result.Features.TimeTracking.Enabled {
			features = append(features, "time_tracking")
		}

		if result.Features.Tags.Enabled {
			features = append(features, "tags")
		}

		if result.Features.Checklists.Enabled {
			features = append(features, "checklists")
		}

		fmt.Printf("Features: %s\n", strings.Join(features, ", "))
	}

	return nil
}

type SpacesCreateCmd struct {
	TeamID            string `arg:"" required:"" help:"Team (workspace) ID"`
	Name              string `arg:"" required:"" help:"Space name"`
	Private           bool   `help:"Make space private"`
	Color             string `help:"Space color (hex)"`
	MultipleAssignees bool   `help:"Enable multiple assignees"`
}

func (cmd *SpacesCreateCmd) Run(ctx context.Context) error {
	client, err := getClickUpClient(ctx)
	if err != nil {
		return err
	}

	req := clickup.CreateSpaceRequest{
		Name:              cmd.Name,
		MultipleAssignees: cmd.MultipleAssignees,
	}

	result, err := client.Spaces().Create(ctx, cmd.TeamID, req)
	if err != nil {
		return err
	}

	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(os.Stdout, result)
	}
	if outfmt.IsPlain(ctx) {
		headers := []string{"ID", "NAME", "PRIVATE"}
		rows := [][]string{{result.ID, result.Name, strconv.FormatBool(result.Private)}}
		return outfmt.WritePlain(os.Stdout, headers, rows)
	}

	fmt.Fprintf(os.Stderr, "Space created\n\n")
	fmt.Printf("ID: %s\n", result.ID)
	fmt.Printf("Name: %s\n", result.Name)
	fmt.Printf("Private: %t\n", result.Private)

	return nil
}

type SpacesUpdateCmd struct {
	SpaceID           string `arg:"" required:"" help:"Space ID"`
	Name              string `help:"New space name"`
	Color             string `help:"Space color (hex)"`
	Private           bool   `help:"Make space private"`
	Public            bool   `help:"Make space public"`
	MultipleAssignees bool   `help:"Enable multiple assignees"`
}

func (cmd *SpacesUpdateCmd) Run(ctx context.Context) error {
	client, err := getClickUpClient(ctx)
	if err != nil {
		return err
	}

	req := clickup.UpdateSpaceRequest{
		Name:  cmd.Name,
		Color: cmd.Color,
	}

	if cmd.Private {
		private := true
		req.Private = &private
	} else if cmd.Public {
		private := false
		req.Private = &private
	}

	if cmd.MultipleAssignees {
		ma := true
		req.MultipleAssignees = &ma
	}

	result, err := client.Spaces().Update(ctx, cmd.SpaceID, req)
	if err != nil {
		return err
	}

	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(os.Stdout, result)
	}
	if outfmt.IsPlain(ctx) {
		headers := []string{"ID", "NAME", "PRIVATE"}
		rows := [][]string{{result.ID, result.Name, strconv.FormatBool(result.Private)}}
		return outfmt.WritePlain(os.Stdout, headers, rows)
	}

	fmt.Fprintf(os.Stderr, "Space updated\n\n")
	fmt.Printf("ID: %s\n", result.ID)
	fmt.Printf("Name: %s\n", result.Name)
	fmt.Printf("Private: %t\n", result.Private)

	return nil
}

type SpacesDeleteCmd struct {
	SpaceID string `arg:"" required:"" help:"Space ID"`
	Force   bool   `help:"Skip confirmation"`
}

func (cmd *SpacesDeleteCmd) Run(ctx context.Context) error {
	client, err := getClickUpClient(ctx)
	if err != nil {
		return err
	}

	if !cmd.Force && !outfmt.IsPlain(ctx) && !outfmt.IsJSON(ctx) {
		fmt.Fprintf(os.Stderr, "WARNING: Deleting a space will delete all folders, lists, and tasks within it.\n")
		fmt.Fprintf(os.Stderr, "Space ID: %s\n\n", cmd.SpaceID)
		fmt.Fprintf(os.Stderr, "Use --force to skip this confirmation.\n")

		return fmt.Errorf("operation cancelled: use --force to confirm destructive operation")
	}

	if err := client.Spaces().Delete(ctx, cmd.SpaceID); err != nil {
		return err
	}

	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(os.Stdout, map[string]string{
			"status":   "success",
			"message":  "Space deleted",
			"space_id": cmd.SpaceID,
		})
	}
	if outfmt.IsPlain(ctx) {
		headers := []string{"STATUS", "SPACE_ID"}
		rows := [][]string{{"success", cmd.SpaceID}}
		return outfmt.WritePlain(os.Stdout, headers, rows)
	}

	fmt.Fprintf(os.Stderr, "Space %s deleted\n", cmd.SpaceID)

	return nil
}

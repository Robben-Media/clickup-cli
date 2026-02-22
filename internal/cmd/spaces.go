package cmd

import (
	"context"
	"fmt"
	"os"
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
		statusCount := len(result.Statuses)
		rows := [][]string{{result.ID, result.Name, fmt.Sprintf("%t", result.Private), fmt.Sprintf("%d", statusCount)}}

		return outfmt.WritePlain(os.Stdout, headers, rows)
	}

	fmt.Printf("ID: %s\n", result.ID)
	fmt.Printf("Name: %s\n", result.Name)
	fmt.Printf("Private: %t\n", result.Private)

	if len(result.Statuses) > 0 {
		statusNames := make([]string, 0, len(result.Statuses))
		for _, s := range result.Statuses {
			statusNames = append(statusNames, s.Status)
		}
		fmt.Printf("Statuses: %s\n", strings.Join(statusNames, ", "))
	}

	var enabledFeatures []string
	if result.Features.DueDates.Enabled {
		enabledFeatures = append(enabledFeatures, "due_dates")
	}

	if result.Features.TimeTracking.Enabled {
		enabledFeatures = append(enabledFeatures, "time_tracking")
	}

	if result.Features.Tags.Enabled {
		enabledFeatures = append(enabledFeatures, "tags")
	}

	if result.Features.Checklists.Enabled {
		enabledFeatures = append(enabledFeatures, "checklists")
	}

	if len(enabledFeatures) > 0 {
		fmt.Printf("Features: %s\n", strings.Join(enabledFeatures, ", "))
	}

	return nil
}

type SpacesCreateCmd struct {
	TeamID            string `arg:"" required:"" help:"Team ID (workspace)"`
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
		rows := [][]string{{result.ID, result.Name, fmt.Sprintf("%t", result.Private)}}

		return outfmt.WritePlain(os.Stdout, headers, rows)
	}

	fmt.Printf("Space created\n")
	fmt.Printf("ID: %s\n", result.ID)
	fmt.Printf("Name: %s\n", result.Name)
	fmt.Printf("Private: %t\n", result.Private)

	return nil
}

type SpacesUpdateCmd struct {
	SpaceID           string `arg:"" required:"" help:"Space ID"`
	Name              string `help:"New space name"`
	Color             string `help:"Space color (hex)"`
	Private           *bool  `help:"Make space private (true/false)"`
	MultipleAssignees *bool  `help:"Enable multiple assignees (true/false)"`
}

func (cmd *SpacesUpdateCmd) Run(ctx context.Context) error {
	client, err := getClickUpClient(ctx)
	if err != nil {
		return err
	}

	req := clickup.UpdateSpaceRequest{
		Name:              cmd.Name,
		Color:             cmd.Color,
		Private:           cmd.Private,
		MultipleAssignees: cmd.MultipleAssignees,
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
		rows := [][]string{{result.ID, result.Name, fmt.Sprintf("%t", result.Private)}}

		return outfmt.WritePlain(os.Stdout, headers, rows)
	}

	fmt.Printf("Space updated\n")
	fmt.Printf("ID: %s\n", result.ID)
	fmt.Printf("Name: %s\n", result.Name)
	fmt.Printf("Private: %t\n", result.Private)

	return nil
}

type SpacesDeleteCmd struct {
	SpaceID string `arg:"" required:"" help:"Space ID"`
}

func (cmd *SpacesDeleteCmd) Run(ctx context.Context) error {
	client, err := getClickUpClient(ctx)
	if err != nil {
		return err
	}

	err = client.Spaces().Delete(ctx, cmd.SpaceID)
	if err != nil {
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

	fmt.Printf("Space %s deleted\n", cmd.SpaceID)

	return nil
}

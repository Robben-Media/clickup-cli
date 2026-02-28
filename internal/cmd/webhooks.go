package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/builtbyrobben/clickup-cli/internal/clickup"
	"github.com/builtbyrobben/clickup-cli/internal/outfmt"
)

type WebhooksCmd struct {
	List   WebhooksListCmd   `cmd:"" help:"List webhooks"`
	Create WebhooksCreateCmd `cmd:"" help:"Create a webhook"`
	Update WebhooksUpdateCmd `cmd:"" help:"Update a webhook"`
	Delete WebhooksDeleteCmd `cmd:"" help:"Delete a webhook"`
}

type WebhooksListCmd struct {
	TeamID string `arg:"" required:"" help:"Workspace/Team ID"`
}

func (cmd *WebhooksListCmd) Run(ctx context.Context) error {
	client, err := getClickUpClient(ctx)
	if err != nil {
		return err
	}

	result, err := client.Webhooks().List(ctx, cmd.TeamID)
	if err != nil {
		return err
	}

	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(os.Stdout, result)
	}

	if outfmt.IsPlain(ctx) {
		headers := []string{"ID", "ENDPOINT", "STATUS", "EVENTS"}
		var rows [][]string

		for _, webhook := range result.Webhooks {
			rows = append(rows, []string{
				webhook.ID,
				webhook.Endpoint,
				webhook.Status,
				strings.Join(webhook.Events, ","),
			})
		}

		return outfmt.WritePlain(os.Stdout, headers, rows)
	}

	// Human-readable output
	if len(result.Webhooks) == 0 {
		fmt.Fprintln(os.Stderr, "No webhooks found")

		return nil
	}

	fmt.Fprintf(os.Stderr, "Found %d webhooks:\n\n", len(result.Webhooks))

	for _, webhook := range result.Webhooks {
		fmt.Printf("  %s\n", webhook.Endpoint)
		fmt.Printf("    ID: %s\n", webhook.ID)
		fmt.Printf("    Status: %s\n", webhook.Status)
		fmt.Printf("    Events: %s\n", strings.Join(webhook.Events, ", "))
	}

	return nil
}

type WebhooksCreateCmd struct {
	TeamID   string `arg:"" required:"" help:"Workspace/Team ID"`
	Endpoint string `required:"" help:"Webhook endpoint URL"`
	Events   string `required:"" help:"Comma-separated list of events"`
	Space    string `help:"Scope to space ID"`
	Folder   string `help:"Scope to folder ID"`
	List     string `help:"Scope to list ID"`
	Task     string `help:"Scope to task ID"`
}

func (cmd *WebhooksCreateCmd) Run(ctx context.Context) error {
	client, err := getClickUpClient(ctx)
	if err != nil {
		return err
	}

	req := clickup.CreateWebhookRequest{
		Endpoint: cmd.Endpoint,
		Events:   strings.Split(cmd.Events, ","),
		SpaceID:  cmd.Space,
		FolderID: cmd.Folder,
		ListID:   cmd.List,
		TaskID:   cmd.Task,
	}

	result, err := client.Webhooks().Create(ctx, cmd.TeamID, req)
	if err != nil {
		return err
	}

	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(os.Stdout, result)
	}

	if outfmt.IsPlain(ctx) {
		headers := []string{"ID", "ENDPOINT", "STATUS"}
		rows := [][]string{{result.ID, result.Endpoint, result.Status}}

		return outfmt.WritePlain(os.Stdout, headers, rows)
	}

	fmt.Fprintf(os.Stderr, "Webhook created\n")
	fmt.Printf("  ID: %s\n", result.ID)
	fmt.Printf("  Endpoint: %s\n", result.Endpoint)
	fmt.Printf("  Status: %s\n", result.Status)
	fmt.Printf("  Events: %s\n", strings.Join(result.Events, ", "))

	return nil
}

type WebhooksUpdateCmd struct {
	WebhookID string `arg:"" required:"" help:"Webhook ID"`
	Endpoint  string `help:"New endpoint URL"`
	Events    string `help:"Comma-separated list of events"`
	Status    string `help:"Webhook status (active or inactive)"`
}

func (cmd *WebhooksUpdateCmd) Run(ctx context.Context) error {
	client, err := getClickUpClient(ctx)
	if err != nil {
		return err
	}

	req := clickup.UpdateWebhookRequest{
		Endpoint: cmd.Endpoint,
		Status:   cmd.Status,
	}

	if cmd.Events != "" {
		req.Events = strings.Split(cmd.Events, ",")
	}

	result, err := client.Webhooks().Update(ctx, cmd.WebhookID, req)
	if err != nil {
		return err
	}

	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(os.Stdout, result)
	}

	if outfmt.IsPlain(ctx) {
		headers := []string{"ID", "ENDPOINT", "STATUS"}
		rows := [][]string{{result.ID, result.Endpoint, result.Status}}

		return outfmt.WritePlain(os.Stdout, headers, rows)
	}

	fmt.Fprintf(os.Stderr, "Webhook updated\n")
	fmt.Printf("  ID: %s\n", result.ID)
	fmt.Printf("  Endpoint: %s\n", result.Endpoint)
	fmt.Printf("  Status: %s\n", result.Status)

	return nil
}

type WebhooksDeleteCmd struct {
	WebhookID string `arg:"" required:"" help:"Webhook ID"`
}

func (cmd *WebhooksDeleteCmd) Run(ctx context.Context) error {
	client, err := getClickUpClient(ctx)
	if err != nil {
		return err
	}

	if !forceEnabled(ctx) {
		fmt.Fprintf(os.Stderr, "Warning: This will permanently delete webhook %s\n", cmd.WebhookID)
		fmt.Fprint(os.Stderr, "Use --force to confirm deletion\n")

		return fmt.Errorf("operation cancelled: use --force to confirm")
	}

	if err := client.Webhooks().Delete(ctx, cmd.WebhookID); err != nil {
		return err
	}

	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(os.Stdout, map[string]string{
			"status":     "success",
			"message":    "Webhook deleted",
			"webhook_id": cmd.WebhookID,
		})
	}

	if outfmt.IsPlain(ctx) {
		headers := []string{"STATUS", "WEBHOOK_ID"}
		rows := [][]string{{"success", cmd.WebhookID}}

		return outfmt.WritePlain(os.Stdout, headers, rows)
	}

	fmt.Fprintf(os.Stderr, "Webhook %s deleted\n", cmd.WebhookID)

	return nil
}

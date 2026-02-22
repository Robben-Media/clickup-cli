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
	Team string `required:"" help:"Team (workspace) ID"`
}

func (cmd *WebhooksListCmd) Run(ctx context.Context) error {
	client, err := getClickUpClient(ctx)
	if err != nil {
		return err
	}

	result, err := client.Webhooks().List(ctx, cmd.Team)
	if err != nil {
		return err
	}

	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(os.Stdout, result)
	}
	if outfmt.IsPlain(ctx) {
		headers := []string{"ID", "ENDPOINT", "STATUS", "EVENTS"}
		var rows [][]string
		for _, wh := range result.Webhooks {
			rows = append(rows, []string{
				wh.ID,
				wh.Endpoint,
				wh.Status,
				strings.Join(wh.Events, ","),
			})
		}
		return outfmt.WritePlain(os.Stdout, headers, rows)
	}

	if len(result.Webhooks) == 0 {
		fmt.Fprintln(os.Stderr, "No webhooks found")
		return nil
	}

	fmt.Fprintf(os.Stderr, "Found %d webhooks\n\n", len(result.Webhooks))

	for _, wh := range result.Webhooks {
		printWebhook(&wh)
	}

	return nil
}

type WebhooksCreateCmd struct {
	Team     string `required:"" help:"Team (workspace) ID"`
	Endpoint string `required:"" help:"Webhook endpoint URL"`
	Events   string `required:"" help:"Comma-separated list of events (e.g., taskCreated,taskUpdated)"`
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

	events := strings.Split(cmd.Events, ",")
	for i, e := range events {
		events[i] = strings.TrimSpace(e)
	}

	req := clickup.CreateWebhookRequest{
		Endpoint: cmd.Endpoint,
		Events:   events,
		SpaceID:  cmd.Space,
		FolderID: cmd.Folder,
		ListID:   cmd.List,
		TaskID:   cmd.Task,
	}

	result, err := client.Webhooks().Create(ctx, cmd.Team, req)
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

	fmt.Fprintf(os.Stderr, "Created webhook\n\n")
	printWebhookDetail(result)

	return nil
}

type WebhooksUpdateCmd struct {
	WebhookID string `arg:"" required:"" help:"Webhook ID"`
	Endpoint  string `help:"New endpoint URL"`
	Events    string `help:"Comma-separated list of events"`
	Status    string `help:"Webhook status: active or inactive"`
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
		events := strings.Split(cmd.Events, ",")
		for i, e := range events {
			events[i] = strings.TrimSpace(e)
		}
		req.Events = events
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

	fmt.Fprintf(os.Stderr, "Updated webhook\n\n")
	printWebhookDetail(result)

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

func printWebhook(wh *clickup.Webhook) {
	fmt.Printf("ID: %s\n", wh.ID)
	fmt.Printf("  Endpoint: %s\n", wh.Endpoint)
	fmt.Printf("  Status: %s\n", wh.Status)
	fmt.Printf("  Events: %s\n", strings.Join(wh.Events, ", "))
	fmt.Println()
}

func printWebhookDetail(wh *clickup.Webhook) {
	fmt.Printf("ID: %s\n", wh.ID)
	fmt.Printf("Endpoint: %s\n", wh.Endpoint)
	fmt.Printf("Status: %s\n", wh.Status)
	fmt.Printf("Events: %s\n", strings.Join(wh.Events, ", "))

	if wh.SpaceID != "" {
		fmt.Printf("Space ID: %s\n", wh.SpaceID)
	}

	if wh.FolderID != "" {
		fmt.Printf("Folder ID: %s\n", wh.FolderID)
	}

	if wh.ListID != "" {
		fmt.Printf("List ID: %s\n", wh.ListID)
	}

	if wh.TaskID != "" {
		fmt.Printf("Task ID: %s\n", wh.TaskID)
	}
}

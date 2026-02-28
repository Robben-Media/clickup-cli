package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/builtbyrobben/clickup-cli/internal/clickup"
	"github.com/builtbyrobben/clickup-cli/internal/outfmt"
)

type AuditLogsCmd struct {
	Query AuditLogsQueryCmd `cmd:"" help:"Query audit logs"`
}

type AuditLogsQueryCmd struct {
	StartDate int64  `help:"Start date (Unix timestamp in milliseconds)"`
	EndDate   int64  `help:"End date (Unix timestamp in milliseconds)"`
	EventType string `help:"Filter by event type"`
	UserID    string `help:"Filter by user ID"`
	Limit     int    `help:"Maximum number of results" default:"100"`
}

func (cmd *AuditLogsQueryCmd) Run(ctx context.Context) error {
	client, err := getClickUpClient(ctx)
	if err != nil {
		return err
	}

	req := clickup.AuditLogQuery{
		StartDate: cmd.StartDate,
		EndDate:   cmd.EndDate,
		EventType: cmd.EventType,
		UserID:    cmd.UserID,
		Limit:     cmd.Limit,
	}

	result, err := client.AuditLogs().Query(ctx, req)
	if err != nil {
		return err
	}

	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(os.Stdout, result)
	}

	if outfmt.IsPlain(ctx) {
		headers := []string{"ID", "EVENT_TYPE", "USER_ID", "TIMESTAMP", "RESOURCE_TYPE", "RESOURCE_ID"}
		var rows [][]string
		for _, entry := range result.AuditLogs {
			rows = append(rows, []string{
				entry.ID,
				entry.EventType,
				entry.UserID,
				string(entry.Timestamp),
				entry.ResourceType,
				entry.ResourceID,
			})
		}

		return outfmt.WritePlain(os.Stdout, headers, rows)
	}

	if len(result.AuditLogs) == 0 {
		fmt.Fprintln(os.Stderr, "No audit logs found")
		return nil
	}

	fmt.Fprintf(os.Stderr, "Found %d audit logs\n\n", len(result.AuditLogs))

	for _, entry := range result.AuditLogs {
		printAuditLogEntry(&entry)
	}

	return nil
}

func printAuditLogEntry(entry *clickup.AuditLogEntry) {
	fmt.Printf("%s: %s by user %s at %s\n", entry.ID, entry.EventType, entry.UserID, string(entry.Timestamp))
	fmt.Printf("  Resource: %s %s\n", entry.ResourceType, entry.ResourceID)
	fmt.Println()
}

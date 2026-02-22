package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/builtbyrobben/clickup-cli/internal/clickup"
	"github.com/builtbyrobben/clickup-cli/internal/config"
	"github.com/builtbyrobben/clickup-cli/internal/secrets"
)

func getClickUpClient(ctx context.Context) (*clickup.Client, error) {
	// Resolve workspace ID: context (flag) > env var > config
	workspaceID := getWorkspaceIDFromContext(ctx)

	if workspaceID == "" {
		workspaceID = os.Getenv("CLICKUP_WORKSPACE_ID")
	}

	if workspaceID == "" {
		var err error

		workspaceID, err = config.GetWorkspaceID()
		if err != nil {
			return nil, fmt.Errorf("read workspace config: %w", err)
		}
	}

	// 1. Check env var for API key
	if key := os.Getenv("CLICKUP_API_KEY"); key != "" {
		return clickup.NewClient(key, clickup.WithWorkspaceID(workspaceID)), nil
	}

	// 2. Check keyring
	store, err := secrets.OpenDefault()
	if err != nil {
		return nil, fmt.Errorf("open credential store: %w", err)
	}

	key, err := store.GetAPIKey()
	if err != nil {
		return nil, fmt.Errorf("no credentials found; run: clickup-cli auth set-key --stdin")
	}

	return clickup.NewClient(key, clickup.WithWorkspaceID(workspaceID)), nil
}

func getTeamID() (string, error) {
	// 1. Check env var
	if id := os.Getenv("CLICKUP_TEAM_ID"); id != "" {
		return id, nil
	}

	// 2. Check config file
	id, err := config.GetTeamID()
	if err != nil {
		return "", fmt.Errorf("read config: %w", err)
	}

	if id != "" {
		return id, nil
	}

	// 3. Error with instructions
	return "", fmt.Errorf("no team ID configured; run: clickup-cli auth set-team <TEAM_ID>")
}

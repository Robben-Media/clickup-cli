package cmd

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"golang.org/x/term"

	"github.com/builtbyrobben/clickup-cli/internal/config"
	"github.com/builtbyrobben/clickup-cli/internal/outfmt"
	"github.com/builtbyrobben/clickup-cli/internal/secrets"
)

type AuthCmd struct {
	SetKey  AuthSetKeyCmd  `cmd:"" help:"Set API key (uses --stdin by default)"`
	SetTeam AuthSetTeamCmd `cmd:"" help:"Set ClickUp Team ID"`
	Status  AuthStatusCmd  `cmd:"" help:"Show authentication status"`
	Remove  AuthRemoveCmd  `cmd:"" help:"Remove stored credentials"`
}

type AuthSetKeyCmd struct {
	Stdin bool   `help:"Read API key from stdin (default: true)" default:"true"`
	Key   string `arg:"" optional:"" help:"API key (discouraged; exposes in shell history)"`
}

func (cmd *AuthSetKeyCmd) Run(ctx context.Context) error {
	var apiKey string

	// Priority: argument > stdin
	switch {
	case cmd.Key != "":
		// Warn about shell history exposure
		fmt.Fprintln(os.Stderr, "Warning: passing keys as arguments exposes them in shell history. Use --stdin instead.")
		apiKey = strings.TrimSpace(cmd.Key)
	case term.IsTerminal(int(os.Stdin.Fd())):
		// Interactive prompt
		fmt.Fprint(os.Stderr, "Enter API key: ")

		byteKey, err := term.ReadPassword(int(os.Stdin.Fd()))
		fmt.Fprintln(os.Stderr) // New line after password input

		if err != nil {
			return fmt.Errorf("read API key: %w", err)
		}

		apiKey = strings.TrimSpace(string(byteKey))
	default:
		// Read from stdin (piped)
		byteKey, err := io.ReadAll(os.Stdin)
		if err != nil {
			return fmt.Errorf("read API key from stdin: %w", err)
		}

		apiKey = strings.TrimSpace(string(byteKey))
	}

	if apiKey == "" {
		return fmt.Errorf("API key cannot be empty")
	}

	store, err := secrets.OpenDefault()
	if err != nil {
		return fmt.Errorf("open credential store: %w", err)
	}

	if err := store.SetAPIKey(apiKey); err != nil {
		return fmt.Errorf("store API key: %w", err)
	}

	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(os.Stdout, map[string]string{
			"status":  "success",
			"message": "API key stored in keyring",
		})
	}
	if outfmt.IsPlain(ctx) {
		return outfmt.WritePlain(os.Stdout, []string{"STATUS", "MESSAGE"}, [][]string{{"success", "API key stored in keyring"}})
	}

	fmt.Fprintln(os.Stderr, "API key stored in keyring")

	return nil
}

type AuthSetTeamCmd struct {
	TeamID string `arg:"" required:"" help:"ClickUp Team ID"`
}

func (cmd *AuthSetTeamCmd) Run(ctx context.Context) error {
	if cmd.TeamID == "" {
		return fmt.Errorf("team ID cannot be empty")
	}

	if err := config.SetTeamID(cmd.TeamID); err != nil {
		return fmt.Errorf("store team ID: %w", err)
	}

	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(os.Stdout, map[string]string{
			"status":  "success",
			"message": "Team ID stored in config",
			"team_id": cmd.TeamID,
		})
	}
	if outfmt.IsPlain(ctx) {
		return outfmt.WritePlain(os.Stdout, []string{"STATUS", "TEAM_ID"}, [][]string{{"success", cmd.TeamID}})
	}

	fmt.Fprintf(os.Stderr, "Team ID %s stored in config\n", cmd.TeamID)

	return nil
}

type AuthStatusCmd struct{}

func (cmd *AuthStatusCmd) Run(ctx context.Context) error {
	store, err := secrets.OpenDefault()
	if err != nil {
		return fmt.Errorf("open credential store: %w", err)
	}

	hasKey, err := store.HasKey()
	if err != nil {
		return fmt.Errorf("check API key: %w", err)
	}

	// Check environment variable overrides
	envKey := os.Getenv("CLICKUP_API_KEY")
	envOverride := envKey != ""

	envTeamID := os.Getenv("CLICKUP_TEAM_ID")
	cfgTeamID, _ := config.GetTeamID()

	teamID := envTeamID
	teamSource := "env"

	if teamID == "" {
		teamID = cfgTeamID
		teamSource = "config"
	}

	status := map[string]any{
		"has_key":         hasKey,
		"env_override":    envOverride,
		"storage_backend": "keyring",
		"has_team_id":     teamID != "",
	}

	if teamID != "" {
		status["team_id"] = teamID
		status["team_id_source"] = teamSource
	}

	if hasKey && !envOverride {
		// Show redacted key
		key, err := store.GetAPIKey()
		if err == nil && len(key) > 8 {
			status["key_redacted"] = key[:4] + "..." + key[len(key)-4:]
		}
	}

	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(os.Stdout, status)
	}
	if outfmt.IsPlain(ctx) {
		headers := []string{"HAS_KEY", "ENV_OVERRIDE", "STORAGE", "HAS_TEAM_ID", "TEAM_ID", "TEAM_SOURCE"}
		ts := teamSource
		if teamID == "" {
			ts = ""
		}
		rows := [][]string{{
			fmt.Sprintf("%t", hasKey),
			fmt.Sprintf("%t", envOverride),
			"keyring",
			fmt.Sprintf("%t", teamID != ""),
			teamID,
			ts,
		}}
		return outfmt.WritePlain(os.Stdout, headers, rows)
	}

	// Human-readable output
	fmt.Fprintf(os.Stdout, "Storage: %s\n", status["storage_backend"])

	switch {
	case envOverride:
		fmt.Fprintln(os.Stdout, "API Key: Using CLICKUP_API_KEY environment variable")
	case hasKey:
		fmt.Fprintln(os.Stdout, "API Key: Authenticated")

		if redacted, ok := status["key_redacted"].(string); ok {
			fmt.Fprintf(os.Stdout, "Key: %s\n", redacted)
		}
	default:
		fmt.Fprintln(os.Stdout, "API Key: Not authenticated")
		fmt.Fprintln(os.Stderr, "Run: clickup-cli auth set-key --stdin")
	}

	if teamID != "" {
		fmt.Fprintf(os.Stdout, "Team ID: %s (source: %s)\n", teamID, teamSource)
	} else {
		fmt.Fprintln(os.Stdout, "Team ID: Not configured")
		fmt.Fprintln(os.Stderr, "Run: clickup-cli auth set-team <TEAM_ID>")
	}

	return nil
}

type AuthRemoveCmd struct{}

func (cmd *AuthRemoveCmd) Run(ctx context.Context) error {
	store, err := secrets.OpenDefault()
	if err != nil {
		return fmt.Errorf("open credential store: %w", err)
	}

	if err := store.DeleteAPIKey(); err != nil {
		return fmt.Errorf("remove API key: %w", err)
	}

	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(os.Stdout, map[string]string{
			"status":  "success",
			"message": "API key removed",
		})
	}
	if outfmt.IsPlain(ctx) {
		return outfmt.WritePlain(os.Stdout, []string{"STATUS", "MESSAGE"}, [][]string{{"success", "API key removed"}})
	}

	fmt.Fprintln(os.Stderr, "API key removed")

	return nil
}

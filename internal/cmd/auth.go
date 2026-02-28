package cmd

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"golang.org/x/term"

	"github.com/builtbyrobben/clickup-cli/internal/clickup"
	"github.com/builtbyrobben/clickup-cli/internal/config"
	"github.com/builtbyrobben/clickup-cli/internal/outfmt"
	"github.com/builtbyrobben/clickup-cli/internal/secrets"
)

type AuthCmd struct {
	SetKey       AuthSetKeyCmd       `cmd:"" help:"Set API key (uses --stdin by default)"`
	SetTeam      AuthSetTeamCmd      `cmd:"" help:"Set ClickUp Team ID"`
	SetWorkspace AuthSetWorkspaceCmd `cmd:"" help:"Set ClickUp Workspace ID for v3 API"`
	Status       AuthStatusCmd       `cmd:"" help:"Show authentication status"`
	Remove       AuthRemoveCmd       `cmd:"" help:"Remove stored credentials"`
	Whoami       AuthWhoamiCmd       `cmd:"" help:"Get the currently authorized user"`
	Token        AuthTokenCmd        `cmd:"" help:"Exchange OAuth authorization code for access token"`
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

type AuthSetWorkspaceCmd struct {
	WorkspaceID string `arg:"" required:"" help:"ClickUp Workspace ID for v3 API"`
}

func (cmd *AuthSetWorkspaceCmd) Run(ctx context.Context) error {
	if cmd.WorkspaceID == "" {
		return fmt.Errorf("workspace ID cannot be empty")
	}

	if err := config.SetWorkspaceID(cmd.WorkspaceID); err != nil {
		return fmt.Errorf("store workspace ID: %w", err)
	}

	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(os.Stdout, map[string]string{
			"status":       "success",
			"message":      "Workspace ID stored in config",
			"workspace_id": cmd.WorkspaceID,
		})
	}
	if outfmt.IsPlain(ctx) {
		return outfmt.WritePlain(os.Stdout, []string{"STATUS", "WORKSPACE_ID"}, [][]string{{"success", cmd.WorkspaceID}})
	}

	fmt.Fprintf(os.Stderr, "Workspace ID %s stored in config\n", cmd.WorkspaceID)

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

	envWorkspaceID := os.Getenv("CLICKUP_WORKSPACE_ID")
	cfgWorkspaceID, _ := config.GetWorkspaceID()

	workspaceID := envWorkspaceID
	workspaceSource := "env"

	if workspaceID == "" {
		workspaceID = cfgWorkspaceID
		workspaceSource = "config"
	}

	status := map[string]any{
		"has_key":          hasKey,
		"env_override":     envOverride,
		"storage_backend":  "keyring",
		"has_team_id":      teamID != "",
		"has_workspace_id": workspaceID != "",
	}

	if teamID != "" {
		status["team_id"] = teamID
		status["team_id_source"] = teamSource
	}

	if workspaceID != "" {
		status["workspace_id"] = workspaceID
		status["workspace_id_source"] = workspaceSource
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
		headers := []string{"HAS_KEY", "ENV_OVERRIDE", "STORAGE", "HAS_TEAM_ID", "TEAM_ID", "TEAM_SOURCE", "HAS_WORKSPACE_ID", "WORKSPACE_ID", "WORKSPACE_SOURCE"}
		ts := teamSource
		if teamID == "" {
			ts = ""
		}
		ws := workspaceSource
		if workspaceID == "" {
			ws = ""
		}
		rows := [][]string{{
			fmt.Sprintf("%t", hasKey),
			fmt.Sprintf("%t", envOverride),
			"keyring",
			fmt.Sprintf("%t", teamID != ""),
			teamID,
			ts,
			fmt.Sprintf("%t", workspaceID != ""),
			workspaceID,
			ws,
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

	if workspaceID != "" {
		fmt.Fprintf(os.Stdout, "Workspace ID: %s (source: %s)\n", workspaceID, workspaceSource)
	} else {
		fmt.Fprintln(os.Stdout, "Workspace ID: Not configured (optional, required for v3 API)")
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

type AuthWhoamiCmd struct{}

func (cmd *AuthWhoamiCmd) Run(ctx context.Context) error {
	client, err := getClickUpClient(ctx)
	if err != nil {
		return err
	}

	result, err := client.Auth().Whoami(ctx)
	if err != nil {
		return err
	}

	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(os.Stdout, result)
	}
	if outfmt.IsPlain(ctx) {
		headers := []string{"ID", "USERNAME", "EMAIL"}
		rows := [][]string{{
			fmt.Sprintf("%d", result.User.ID),
			result.User.Username,
			result.User.Email,
		}}
		return outfmt.WritePlain(os.Stdout, headers, rows)
	}

	fmt.Fprintf(os.Stderr, "Authenticated as:\n")
	fmt.Printf("  ID: %d\n", result.User.ID)
	fmt.Printf("  Username: %s\n", result.User.Username)
	fmt.Printf("  Email: %s\n", result.User.Email)

	if result.User.Color != "" {
		fmt.Printf("  Color: %s\n", result.User.Color)
	}

	if result.User.ProfilePicture != "" {
		fmt.Printf("  Profile Picture: %s\n", result.User.ProfilePicture)
	}

	return nil
}

type AuthTokenCmd struct {
	ClientID     string `required:"" help:"OAuth client ID"`
	ClientSecret string `required:"" help:"OAuth client secret"`
	Code         string `required:"" help:"OAuth authorization code"`
}

func (cmd *AuthTokenCmd) Run(ctx context.Context) error {
	client, err := getClickUpClient(ctx)
	if err != nil {
		return err
	}

	result, err := client.Auth().Token(ctx, clickup.OAuthTokenRequest{
		ClientID:     cmd.ClientID,
		ClientSecret: cmd.ClientSecret,
		Code:         cmd.Code,
	})
	if err != nil {
		return err
	}

	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(os.Stdout, result)
	}
	if outfmt.IsPlain(ctx) {
		headers := []string{"ACCESS_TOKEN", "TOKEN_TYPE"}
		rows := [][]string{{result.AccessToken, result.TokenType}}
		return outfmt.WritePlain(os.Stdout, headers, rows)
	}

	fmt.Fprintf(os.Stderr, "Token exchange successful\n")
	fmt.Printf("  Access Token: %s\n", result.AccessToken)
	fmt.Printf("  Token Type: %s\n", result.TokenType)

	return nil
}

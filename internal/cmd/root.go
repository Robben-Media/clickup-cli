package cmd

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"

	"github.com/alecthomas/kong"

	"github.com/builtbyrobben/clickup-cli/internal/errfmt"
	"github.com/builtbyrobben/clickup-cli/internal/outfmt"
)

type RootFlags struct {
	Color     string `help:"Color output: auto|always|never" default:"${color}"`
	JSON      bool   `help:"Output JSON to stdout (best for scripting)" default:"${json}"`
	Plain     bool   `help:"Output stable, parseable text to stdout (TSV; no colors)" default:"${plain}"`
	Force     bool   `help:"Skip confirmations for destructive commands"`
	NoInput   bool   `help:"Never prompt; fail instead (useful for CI)"`
	Verbose   bool   `help:"Enable verbose logging"`
	Workspace string `help:"Workspace ID for v3 API calls (required for Chat, Docs, etc.)" default:"${workspace}"`
}

type CLI struct {
	RootFlags `embed:""`

	Version    kong.VersionFlag `help:"Print version and exit"`
	Auth       AuthCmd          `cmd:"" help:"Auth and credentials"`
	Workspaces WorkspacesCmd    `cmd:"" help:"Workspace operations"`
	Tasks      TasksCmd         `cmd:"" help:"Task operations"`
	Spaces     SpacesCmd        `cmd:"" help:"Space operations"`
	Lists      ListsCmd         `cmd:"" help:"List operations"`
	Members    MembersCmd       `cmd:"" help:"Team member operations"`
	Comments   CommentsCmd      `cmd:"" help:"Comment operations"`
	Time       TimeCmd          `cmd:"" help:"Time tracking"`
	Groups     GroupsCmd        `cmd:"" help:"User group operations"`
	Roles      RolesCmd         `cmd:"" help:"Custom role operations"`
	VersionCmd VersionCmd       `cmd:"" name:"version" help:"Print version"`
}

type exitPanic struct{ code int }

type workspaceIDKey struct{}

func withWorkspaceID(ctx context.Context, workspaceID string) context.Context {
	return context.WithValue(ctx, workspaceIDKey{}, workspaceID)
}

func getWorkspaceIDFromContext(ctx context.Context) string {
	if v, ok := ctx.Value(workspaceIDKey{}).(string); ok {
		return v
	}
	return ""
}

func Execute(args []string) (err error) {
	parser, cli, err := newParser(helpDescription())
	if err != nil {
		return err
	}

	defer func() {
		if r := recover(); r != nil {
			if ep, ok := r.(exitPanic); ok {
				if ep.code == 0 {
					err = nil
					return
				}
				err = &ExitError{Code: ep.code, Err: errors.New("exited")}
				return
			}
			panic(r)
		}
	}()

	kctx, err := parser.Parse(args)
	if err != nil {
		parsedErr := wrapParseError(err)
		_, _ = fmt.Fprintln(os.Stderr, errfmt.Format(parsedErr))
		return parsedErr
	}

	logLevel := slog.LevelWarn
	if cli.Verbose {
		logLevel = slog.LevelDebug
	}

	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: logLevel,
	})))

	mode, err := outfmt.FromFlags(cli.JSON, cli.Plain)
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, errfmt.Format(err))
		return newUsageError(err)
	}

	ctx := context.Background()
	ctx = outfmt.WithMode(ctx, mode)
	ctx = withWorkspaceID(ctx, cli.Workspace)

	kctx.BindTo(ctx, (*context.Context)(nil))
	kctx.Bind(&cli.RootFlags)

	err = kctx.Run()
	if err == nil {
		return nil
	}

	_, _ = fmt.Fprintln(os.Stderr, errfmt.Format(err))

	return err
}

func wrapParseError(err error) error {
	if err == nil {
		return nil
	}

	var parseErr *kong.ParseError
	if errors.As(err, &parseErr) {
		return &ExitError{Code: 2, Err: parseErr}
	}

	return err
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}

	return fallback
}

func boolString(v bool) string {
	if v {
		return "true"
	}

	return "false"
}

func newParser(description string) (*kong.Kong, *CLI, error) {
	envMode := outfmt.FromEnv("CLICKUP_CLI")
	vars := kong.Vars{
		"color":     envOr("CLICKUP_CLI_COLOR", "auto"),
		"json":      boolString(envMode.JSON),
		"plain":     boolString(envMode.Plain),
		"version":   VersionString(),
		"workspace": envOr("CLICKUP_WORKSPACE_ID", ""),
	}

	cli := &CLI{}

	parser, err := kong.New(
		cli,
		kong.Name("clickup-cli"),
		kong.Description(description),
		kong.ConfigureHelp(kong.HelpOptions{
			Compact: true,
			Summary: true,
		}),
		kong.Vars(vars),
		kong.Writers(os.Stdout, os.Stderr),
		kong.Exit(func(code int) { panic(exitPanic{code: code}) }),
	)
	if err != nil {
		return nil, nil, err
	}

	return parser, cli, nil
}

func helpDescription() string {
	return "ClickUp CLI - Project management from the command line"
}

// newUsageError wraps errors in a way main() can map to exit code 2.
func newUsageError(err error) error {
	if err == nil {
		return nil
	}

	return &ExitError{Code: 2, Err: err}
}

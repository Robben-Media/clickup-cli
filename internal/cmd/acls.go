package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/builtbyrobben/clickup-cli/internal/clickup"
	"github.com/builtbyrobben/clickup-cli/internal/outfmt"
)

type ACLsCmd struct {
	Update ACLsUpdateCmd `cmd:"" help:"Update access control settings"`
}

type ACLsUpdateCmd struct {
	ObjectType string `name:"type" short:"t" required:"" help:"Object type (space, folder, list)"`
	ObjectID   string `name:"id" short:"i" required:"" help:"Object ID"`
	Private    bool   `help:"Make object private"`
	Public     bool   `help:"Make object public"`
	Sharing    string `help:"Sharing mode (open or closed)"`
}

func (cmd *ACLsUpdateCmd) Run(ctx context.Context) error {
	client, err := getClickUpClient(ctx)
	if err != nil {
		return err
	}

	req := clickup.UpdateACLRequest{
		Sharing: cmd.Sharing,
	}

	// Handle private/public flags
	if cmd.Private && cmd.Public {
		return fmt.Errorf("cannot specify both --private and --public")
	}

	if cmd.Private {
		privateTrue := true
		req.Private = &privateTrue
	} else if cmd.Public {
		privateFalse := false
		req.Private = &privateFalse
	}

	if err := client.ACLs().Update(ctx, cmd.ObjectType, cmd.ObjectID, req); err != nil {
		return err
	}

	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(os.Stdout, map[string]string{"status": "success", "message": "ACL updated"})
	}

	if outfmt.IsPlain(ctx) {
		headers := []string{"STATUS", "OBJECT_TYPE", "OBJECT_ID"}
		rows := [][]string{{"success", cmd.ObjectType, cmd.ObjectID}}

		return outfmt.WritePlain(os.Stdout, headers, rows)
	}

	fmt.Fprintf(os.Stderr, "ACL updated for %s %s\n", cmd.ObjectType, cmd.ObjectID)

	return nil
}

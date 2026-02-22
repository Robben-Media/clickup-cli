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

type GroupsCmd struct {
	List   GroupsListCmd   `cmd:"" help:"List user groups"`
	Create GroupsCreateCmd `cmd:"" help:"Create a user group"`
	Update GroupsUpdateCmd `cmd:"" help:"Update a user group"`
	Delete GroupsDeleteCmd `cmd:"" help:"Delete a user group"`
}

type GroupsListCmd struct{}

func (cmd *GroupsListCmd) Run(ctx context.Context) error {
	client, err := getClickUpClient(ctx)
	if err != nil {
		return err
	}

	result, err := client.UserGroups().List(ctx)
	if err != nil {
		return err
	}

	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(os.Stdout, result)
	}
	if outfmt.IsPlain(ctx) {
		headers := []string{"ID", "NAME", "MEMBER_COUNT"}
		var rows [][]string
		for _, group := range result.Groups {
			rows = append(rows, []string{group.ID, group.Name, strconv.Itoa(len(group.Members))})
		}
		return outfmt.WritePlain(os.Stdout, headers, rows)
	}

	if len(result.Groups) == 0 {
		fmt.Fprintln(os.Stderr, "No user groups found")
		return nil
	}

	fmt.Fprintf(os.Stderr, "Found %d user groups\n\n", len(result.Groups))

	for _, group := range result.Groups {
		printUserGroup(&group)
	}

	return nil
}

type GroupsCreateCmd struct {
	TeamID  string `arg:"" required:"" help:"Team (workspace) ID"`
	Name    string `arg:"" required:"" help:"Group name"`
	Members string `help:"Comma-separated user IDs to add"`
}

func (cmd *GroupsCreateCmd) Run(ctx context.Context) error {
	client, err := getClickUpClient(ctx)
	if err != nil {
		return err
	}

	req := clickup.CreateUserGroupRequest{
		Name: cmd.Name,
	}

	if cmd.Members != "" {
		memberIDs, parseErr := parseMemberIDs(cmd.Members)
		if parseErr != nil {
			return parseErr
		}
		req.Members = memberIDs
	}

	result, err := client.UserGroups().Create(ctx, cmd.TeamID, req)
	if err != nil {
		return err
	}

	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(os.Stdout, result)
	}
	if outfmt.IsPlain(ctx) {
		headers := []string{"ID", "NAME", "MEMBER_COUNT"}
		rows := [][]string{{result.ID, result.Name, strconv.Itoa(len(result.Members))}}
		return outfmt.WritePlain(os.Stdout, headers, rows)
	}

	fmt.Fprintf(os.Stderr, "Created user group\n\n")
	printUserGroupDetail(result)

	return nil
}

type GroupsUpdateCmd struct {
	GroupID       string `arg:"" required:"" help:"Group ID"`
	Name          string `help:"New group name"`
	AddMembers    string `help:"Comma-separated user IDs to add"`
	RemoveMembers string `help:"Comma-separated user IDs to remove"`
}

func (cmd *GroupsUpdateCmd) Run(ctx context.Context) error {
	client, err := getClickUpClient(ctx)
	if err != nil {
		return err
	}

	req := clickup.UpdateUserGroupRequest{
		Name: cmd.Name,
	}

	if cmd.AddMembers != "" || cmd.RemoveMembers != "" {
		req.Members = &clickup.UserGroupMembersUpdate{}

		if cmd.AddMembers != "" {
			memberIDs, parseErr := parseMemberIDs(cmd.AddMembers)
			if parseErr != nil {
				return parseErr
			}
			req.Members.Add = memberIDs
		}

		if cmd.RemoveMembers != "" {
			memberIDs, parseErr := parseMemberIDs(cmd.RemoveMembers)
			if parseErr != nil {
				return parseErr
			}
			req.Members.Rem = memberIDs
		}
	}

	result, err := client.UserGroups().Update(ctx, cmd.GroupID, req)
	if err != nil {
		return err
	}

	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(os.Stdout, result)
	}
	if outfmt.IsPlain(ctx) {
		headers := []string{"ID", "NAME", "MEMBER_COUNT"}
		rows := [][]string{{result.ID, result.Name, strconv.Itoa(len(result.Members))}}
		return outfmt.WritePlain(os.Stdout, headers, rows)
	}

	fmt.Fprintf(os.Stderr, "Updated user group\n\n")
	printUserGroupDetail(result)

	return nil
}

type GroupsDeleteCmd struct {
	GroupID string `arg:"" required:"" help:"Group ID"`
}

func (cmd *GroupsDeleteCmd) Run(ctx context.Context) error {
	client, err := getClickUpClient(ctx)
	if err != nil {
		return err
	}

	if err := client.UserGroups().Delete(ctx, cmd.GroupID); err != nil {
		return err
	}

	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(os.Stdout, map[string]string{
			"status":   "success",
			"message":  "User group deleted",
			"group_id": cmd.GroupID,
		})
	}
	if outfmt.IsPlain(ctx) {
		headers := []string{"STATUS", "GROUP_ID"}
		rows := [][]string{{"success", cmd.GroupID}}
		return outfmt.WritePlain(os.Stdout, headers, rows)
	}

	fmt.Fprintf(os.Stderr, "User group %s deleted\n", cmd.GroupID)

	return nil
}

func parseMemberIDs(s string) ([]int, error) {
	parts := strings.Split(s, ",")
	ids := make([]int, 0, len(parts))

	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}

		id, err := strconv.Atoi(p)
		if err != nil {
			return nil, fmt.Errorf("invalid user ID %q: %w", p, err)
		}

		ids = append(ids, id)
	}

	return ids, nil
}

func printUserGroup(group *clickup.UserGroup) {
	fmt.Printf("ID: %s\n", group.ID)
	fmt.Printf("  Name: %s\n", group.Name)
	fmt.Printf("  Members: %d\n", len(group.Members))
	fmt.Println()
}

func printUserGroupDetail(group *clickup.UserGroup) {
	fmt.Printf("ID: %s\n", group.ID)
	fmt.Printf("Name: %s\n", group.Name)

	if len(group.Members) > 0 {
		fmt.Print("Members:")

		for _, m := range group.Members {
			fmt.Printf(" %s", m.Username)
		}

		fmt.Println()
	}
}

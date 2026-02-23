package cmd

import (
	"context"
	"fmt"
	"os"
	"strconv"

	"github.com/builtbyrobben/clickup-cli/internal/clickup"
	"github.com/builtbyrobben/clickup-cli/internal/outfmt"
)

type GuestsCmd struct {
	Get              GuestsGetCmd              `cmd:"" help:"Get a guest by ID"`
	Invite           GuestsInviteCmd           `cmd:"" help:"Invite a guest to the workspace"`
	Update           GuestsUpdateCmd           `cmd:"" help:"Update guest permissions"`
	Remove           GuestsRemoveCmd           `cmd:"" help:"Remove a guest from the workspace"`
	AddToTask        GuestsAddToTaskCmd        `cmd:"" help:"Add a guest to a task"`
	RemoveFromTask   GuestsRemoveFromTaskCmd   `cmd:"" help:"Remove a guest from a task"`
	AddToList        GuestsAddToListCmd        `cmd:"" help:"Add a guest to a list"`
	RemoveFromList   GuestsRemoveFromListCmd   `cmd:"" help:"Remove a guest from a list"`
	AddToFolder      GuestsAddToFolderCmd      `cmd:"" help:"Add a guest to a folder"`
	RemoveFromFolder GuestsRemoveFromFolderCmd `cmd:"" help:"Remove a guest from a folder"`
}

type GuestsGetCmd struct {
	TeamID  string `arg:"" required:"" help:"Team (workspace) ID"`
	GuestID int    `arg:"" required:"" help:"Guest ID"`
}

func (cmd *GuestsGetCmd) Run(ctx context.Context) error {
	client, err := getClickUpClient(ctx)
	if err != nil {
		return err
	}

	result, err := client.Guests().Get(ctx, cmd.TeamID, cmd.GuestID)
	if err != nil {
		return err
	}

	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(os.Stdout, result)
	}
	if outfmt.IsPlain(ctx) {
		headers := []string{"ID", "USERNAME", "EMAIL", "TASKS", "LISTS", "FOLDERS"}
		rows := [][]string{{
			strconv.Itoa(result.ID),
			result.Username,
			result.Email,
			strconv.Itoa(result.TasksCount),
			strconv.Itoa(result.ListsCount),
			strconv.Itoa(result.FoldersCount),
		}}
		return outfmt.WritePlain(os.Stdout, headers, rows)
	}

	printGuestDetail(result)

	return nil
}

type GuestsInviteCmd struct {
	TeamID              string `arg:"" required:"" help:"Team (workspace) ID"`
	Email               string `arg:"" required:"" help:"Guest email address"`
	CanEditTags         bool   `help:"Allow guest to edit tags"`
	CanSeeTimeSpent     bool   `help:"Allow guest to see time spent"`
	CanSeeTimeEstimated bool   `help:"Allow guest to see time estimated"`
}

func (cmd *GuestsInviteCmd) Run(ctx context.Context) error {
	client, err := getClickUpClient(ctx)
	if err != nil {
		return err
	}

	req := clickup.InviteGuestRequest{
		Email:               cmd.Email,
		CanEditTags:         cmd.CanEditTags,
		CanSeeTimeSpent:     cmd.CanSeeTimeSpent,
		CanSeeTimeEstimated: cmd.CanSeeTimeEstimated,
	}

	result, err := client.Guests().Invite(ctx, cmd.TeamID, req)
	if err != nil {
		return err
	}

	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(os.Stdout, result)
	}
	if outfmt.IsPlain(ctx) {
		headers := []string{"ID", "USERNAME", "EMAIL"}
		rows := [][]string{{
			strconv.Itoa(result.ID),
			result.Username,
			result.Email,
		}}
		return outfmt.WritePlain(os.Stdout, headers, rows)
	}

	fmt.Fprintf(os.Stderr, "Invited guest to workspace\n\n")
	printGuestDetail(result)

	return nil
}

type GuestsUpdateCmd struct {
	TeamID              string `arg:"" required:"" help:"Team (workspace) ID"`
	GuestID             int    `arg:"" required:"" help:"Guest ID"`
	CanEditTags         *bool  `help:"Allow guest to edit tags (true/false)"`
	CanSeeTimeSpent     *bool  `help:"Allow guest to see time spent (true/false)"`
	CanSeeTimeEstimated *bool  `help:"Allow guest to see time estimated (true/false)"`
}

func (cmd *GuestsUpdateCmd) Run(ctx context.Context) error {
	client, err := getClickUpClient(ctx)
	if err != nil {
		return err
	}

	req := clickup.EditGuestRequest{
		CanEditTags:         cmd.CanEditTags,
		CanSeeTimeSpent:     cmd.CanSeeTimeSpent,
		CanSeeTimeEstimated: cmd.CanSeeTimeEstimated,
	}

	result, err := client.Guests().Update(ctx, cmd.TeamID, cmd.GuestID, req)
	if err != nil {
		return err
	}

	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(os.Stdout, result)
	}
	if outfmt.IsPlain(ctx) {
		headers := []string{"ID", "USERNAME", "EMAIL"}
		rows := [][]string{{
			strconv.Itoa(result.ID),
			result.Username,
			result.Email,
		}}
		return outfmt.WritePlain(os.Stdout, headers, rows)
	}

	fmt.Fprintf(os.Stderr, "Updated guest permissions\n\n")
	printGuestDetail(result)

	return nil
}

type GuestsRemoveCmd struct {
	TeamID  string `arg:"" required:"" help:"Team (workspace) ID"`
	GuestID int    `arg:"" required:"" help:"Guest ID"`
}

func (cmd *GuestsRemoveCmd) Run(ctx context.Context) error {
	client, err := getClickUpClient(ctx)
	if err != nil {
		return err
	}

	if err := client.Guests().Remove(ctx, cmd.TeamID, cmd.GuestID); err != nil {
		return err
	}

	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(os.Stdout, map[string]string{
			"status":   "success",
			"message":  "Guest removed from workspace",
			"guest_id": strconv.Itoa(cmd.GuestID),
		})
	}
	if outfmt.IsPlain(ctx) {
		headers := []string{"STATUS", "GUEST_ID"}
		rows := [][]string{{"success", strconv.Itoa(cmd.GuestID)}}
		return outfmt.WritePlain(os.Stdout, headers, rows)
	}

	fmt.Fprintf(os.Stderr, "Guest %d removed from workspace\n", cmd.GuestID)

	return nil
}

type GuestsAddToTaskCmd struct {
	TaskID          string `arg:"" required:"" help:"Task ID"`
	GuestID         int    `arg:"" required:"" help:"Guest ID"`
	PermissionLevel string `short:"p" required:"" help:"Permission level: read, comment, edit, create"`
}

func (cmd *GuestsAddToTaskCmd) Run(ctx context.Context) error {
	client, err := getClickUpClient(ctx)
	if err != nil {
		return err
	}

	result, err := client.Guests().AddToTask(ctx, cmd.TaskID, cmd.GuestID, cmd.PermissionLevel)
	if err != nil {
		return err
	}

	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(os.Stdout, result)
	}
	if outfmt.IsPlain(ctx) {
		headers := []string{"ID", "USERNAME", "TASKS"}
		rows := [][]string{{
			strconv.Itoa(result.ID),
			result.Username,
			strconv.Itoa(result.TasksCount),
		}}
		return outfmt.WritePlain(os.Stdout, headers, rows)
	}

	fmt.Fprintf(os.Stderr, "Added guest to task with %s permission\n\n", cmd.PermissionLevel)
	printGuestDetail(result)

	return nil
}

type GuestsRemoveFromTaskCmd struct {
	TaskID  string `arg:"" required:"" help:"Task ID"`
	GuestID int    `arg:"" required:"" help:"Guest ID"`
}

func (cmd *GuestsRemoveFromTaskCmd) Run(ctx context.Context) error {
	client, err := getClickUpClient(ctx)
	if err != nil {
		return err
	}

	if err := client.Guests().RemoveFromTask(ctx, cmd.TaskID, cmd.GuestID); err != nil {
		return err
	}

	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(os.Stdout, map[string]string{
			"status":   "success",
			"message":  "Guest removed from task",
			"task_id":  cmd.TaskID,
			"guest_id": strconv.Itoa(cmd.GuestID),
		})
	}
	if outfmt.IsPlain(ctx) {
		headers := []string{"STATUS", "TASK_ID", "GUEST_ID"}
		rows := [][]string{{"success", cmd.TaskID, strconv.Itoa(cmd.GuestID)}}
		return outfmt.WritePlain(os.Stdout, headers, rows)
	}

	fmt.Fprintf(os.Stderr, "Guest %d removed from task %s\n", cmd.GuestID, cmd.TaskID)

	return nil
}

type GuestsAddToListCmd struct {
	ListID          string `arg:"" required:"" help:"List ID"`
	GuestID         int    `arg:"" required:"" help:"Guest ID"`
	PermissionLevel string `short:"p" required:"" help:"Permission level: read, comment, edit, create"`
}

func (cmd *GuestsAddToListCmd) Run(ctx context.Context) error {
	client, err := getClickUpClient(ctx)
	if err != nil {
		return err
	}

	result, err := client.Guests().AddToList(ctx, cmd.ListID, cmd.GuestID, cmd.PermissionLevel)
	if err != nil {
		return err
	}

	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(os.Stdout, result)
	}
	if outfmt.IsPlain(ctx) {
		headers := []string{"ID", "USERNAME", "LISTS"}
		rows := [][]string{{
			strconv.Itoa(result.ID),
			result.Username,
			strconv.Itoa(result.ListsCount),
		}}
		return outfmt.WritePlain(os.Stdout, headers, rows)
	}

	fmt.Fprintf(os.Stderr, "Added guest to list with %s permission\n\n", cmd.PermissionLevel)
	printGuestDetail(result)

	return nil
}

type GuestsRemoveFromListCmd struct {
	ListID  string `arg:"" required:"" help:"List ID"`
	GuestID int    `arg:"" required:"" help:"Guest ID"`
}

func (cmd *GuestsRemoveFromListCmd) Run(ctx context.Context) error {
	client, err := getClickUpClient(ctx)
	if err != nil {
		return err
	}

	if err := client.Guests().RemoveFromList(ctx, cmd.ListID, cmd.GuestID); err != nil {
		return err
	}

	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(os.Stdout, map[string]string{
			"status":   "success",
			"message":  "Guest removed from list",
			"list_id":  cmd.ListID,
			"guest_id": strconv.Itoa(cmd.GuestID),
		})
	}
	if outfmt.IsPlain(ctx) {
		headers := []string{"STATUS", "LIST_ID", "GUEST_ID"}
		rows := [][]string{{"success", cmd.ListID, strconv.Itoa(cmd.GuestID)}}
		return outfmt.WritePlain(os.Stdout, headers, rows)
	}

	fmt.Fprintf(os.Stderr, "Guest %d removed from list %s\n", cmd.GuestID, cmd.ListID)

	return nil
}

type GuestsAddToFolderCmd struct {
	FolderID        string `arg:"" required:"" help:"Folder ID"`
	GuestID         int    `arg:"" required:"" help:"Guest ID"`
	PermissionLevel string `short:"p" required:"" help:"Permission level: read, comment, edit, create"`
}

func (cmd *GuestsAddToFolderCmd) Run(ctx context.Context) error {
	client, err := getClickUpClient(ctx)
	if err != nil {
		return err
	}

	result, err := client.Guests().AddToFolder(ctx, cmd.FolderID, cmd.GuestID, cmd.PermissionLevel)
	if err != nil {
		return err
	}

	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(os.Stdout, result)
	}
	if outfmt.IsPlain(ctx) {
		headers := []string{"ID", "USERNAME", "FOLDERS"}
		rows := [][]string{{
			strconv.Itoa(result.ID),
			result.Username,
			strconv.Itoa(result.FoldersCount),
		}}
		return outfmt.WritePlain(os.Stdout, headers, rows)
	}

	fmt.Fprintf(os.Stderr, "Added guest to folder with %s permission\n\n", cmd.PermissionLevel)
	printGuestDetail(result)

	return nil
}

type GuestsRemoveFromFolderCmd struct {
	FolderID string `arg:"" required:"" help:"Folder ID"`
	GuestID  int    `arg:"" required:"" help:"Guest ID"`
}

func (cmd *GuestsRemoveFromFolderCmd) Run(ctx context.Context) error {
	client, err := getClickUpClient(ctx)
	if err != nil {
		return err
	}

	if err := client.Guests().RemoveFromFolder(ctx, cmd.FolderID, cmd.GuestID); err != nil {
		return err
	}

	if outfmt.IsJSON(ctx) {
		return outfmt.WriteJSON(os.Stdout, map[string]string{
			"status":    "success",
			"message":   "Guest removed from folder",
			"folder_id": cmd.FolderID,
			"guest_id":  strconv.Itoa(cmd.GuestID),
		})
	}
	if outfmt.IsPlain(ctx) {
		headers := []string{"STATUS", "FOLDER_ID", "GUEST_ID"}
		rows := [][]string{{"success", cmd.FolderID, strconv.Itoa(cmd.GuestID)}}
		return outfmt.WritePlain(os.Stdout, headers, rows)
	}

	fmt.Fprintf(os.Stderr, "Guest %d removed from folder %s\n", cmd.GuestID, cmd.FolderID)

	return nil
}

func printGuestDetail(guest *clickup.Guest) {
	fmt.Printf("ID: %d\n", guest.ID)
	fmt.Printf("Username: %s\n", guest.Username)
	fmt.Printf("Email: %s\n", guest.Email)

	if guest.TasksCount > 0 || guest.ListsCount > 0 || guest.FoldersCount > 0 {
		fmt.Printf("Access: %d tasks, %d lists, %d folders\n",
			guest.TasksCount, guest.ListsCount, guest.FoldersCount)
	}
}

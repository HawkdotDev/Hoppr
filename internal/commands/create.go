package commands

import (
	"context"
	"fmt"
	"io"

	"hoppr/internal/core"
	"hoppr/internal/domain"
	"hoppr/internal/ui"
)

type CreateCommand struct {
	service *core.ProjectService
}

func NewCreateCommand(service *core.ProjectService) *CreateCommand {
	return &CreateCommand{service: service}
}

func (c *CreateCommand) Name() string        { return "create" }
func (c *CreateCommand) Aliases() []string   { return []string{"new"} }
func (c *CreateCommand) Synopsis() string    { return "Create a new empty list" }
func (c *CreateCommand) Usage() string       { return "hop create <list>" }

func (c *CreateCommand) Execute(ctx context.Context, args []string, out io.Writer, errOut io.Writer) int {
	if len(args) < 1 {
		ui.ErrorMsg(errOut, "Please provide a list name.")
		fmt.Fprintf(errOut, "  %s %s\n", ui.Gray("Usage:"), ui.Cyan(c.Usage()))
		return domain.ExitUsage
	}

	listName := args[0]
	err := c.service.CreateList(ctx, listName)
	if err != nil {
		ui.ErrorMsg(errOut, "%v", err)
		return domain.ExitFailure
	}

	ui.SuccessMsg(out, "Created project list %s", ui.Violet(fmt.Sprintf("[%s]", listName)))
	return domain.ExitSuccess
}

package commands

import (
	"context"
	"fmt"
	"io"

	"hoppr/internal/core"
	"hoppr/internal/domain"
	"hoppr/internal/ui"
)

type RemoveCommand struct {
	service *core.ProjectService
}

func NewRemoveCommand(service *core.ProjectService) *RemoveCommand {
	return &RemoveCommand{service: service}
}

func (c *RemoveCommand) Name() string        { return "remove" }
func (c *RemoveCommand) Aliases() []string   { return []string{"rm", "del", "delete"} }
func (c *RemoveCommand) Synopsis() string    { return "Remove a project from a list" }
func (c *RemoveCommand) Usage() string       { return "hop remove <name> [list]" }

func (c *RemoveCommand) Execute(ctx context.Context, args []string, out io.Writer, errOut io.Writer) int {
	if len(args) < 1 {
		ui.ErrorMsg(errOut, "Please provide a project name.")
		fmt.Fprintf(errOut, "  %s %s\n", ui.Gray("Usage:"), ui.Cyan(c.Usage()))
		return domain.ExitUsage
	}

	name := args[0]
	list := ""
	if len(args) >= 2 {
		list = args[1]
	}

	finalName, finalList, err := c.service.RemoveProject(ctx, name, list)
	if err != nil {
		ui.ErrorMsg(errOut, "%v", err)
		if err == domain.ErrProjectNotFound || err == domain.ErrListNotFound {
			return domain.ExitNotFound
		}
		return domain.ExitFailure
	}

	ui.SuccessMsg(out, "Removed project %s from list %s",
		ui.Cyan(ui.Bold(finalName)),
		ui.Violet(fmt.Sprintf("[%s]", finalList)),
	)
	return domain.ExitSuccess
}

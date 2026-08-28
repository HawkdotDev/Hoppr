package commands

import (
	"context"
	"fmt"
	"io"

	"hoppr/internal/core"
	"hoppr/internal/domain"
	"hoppr/internal/ui"
)

type DropCommand struct {
	service *core.ProjectService
}

func NewDropCommand(service *core.ProjectService) *DropCommand {
	return &DropCommand{service: service}
}

func (c *DropCommand) Name() string        { return "drop" }
func (c *DropCommand) Aliases() []string   { return []string{"delete-list", "rmdir"} }
func (c *DropCommand) Synopsis() string    { return "Delete an entire list" }
func (c *DropCommand) Usage() string       { return "hop drop <list>" }

func (c *DropCommand) Execute(ctx context.Context, args []string, out io.Writer, errOut io.Writer) int {
	if len(args) < 1 {
		ui.ErrorMsg(errOut, "Please provide a list name to drop.")
		fmt.Fprintf(errOut, "  %s %s\n", ui.Gray("Usage:"), ui.Cyan(c.Usage()))
		return domain.ExitUsage
	}

	listName := args[0]
	err := c.service.DropList(ctx, listName)
	if err != nil {
		ui.ErrorMsg(errOut, "%v", err)
		if err == domain.ErrListNotFound {
			return domain.ExitNotFound
		}
		return domain.ExitFailure
	}

	ui.SuccessMsg(out, "Deleted project list %s", ui.Violet(fmt.Sprintf("[%s]", listName)))
	return domain.ExitSuccess
}

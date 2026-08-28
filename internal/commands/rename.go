package commands

import (
	"context"
	"fmt"
	"io"

	"hoppr/internal/core"
	"hoppr/internal/domain"
	"hoppr/internal/ui"
)

type RenameCommand struct {
	service *core.ProjectService
}

func NewRenameCommand(service *core.ProjectService) *RenameCommand {
	return &RenameCommand{service: service}
}

func (c *RenameCommand) Name() string        { return "rename" }
func (c *RenameCommand) Aliases() []string   { return []string{"mv"} }
func (c *RenameCommand) Synopsis() string    { return "Rename a list" }
func (c *RenameCommand) Usage() string       { return "hop rename <list> <new>" }

func (c *RenameCommand) Execute(ctx context.Context, args []string, out io.Writer, errOut io.Writer) int {
	if len(args) < 2 {
		fmt.Fprintln(errOut, ui.Red("Error: Please provide current list name and new list name."))
		fmt.Fprintf(errOut, "Usage: %s\n", c.Usage())
		return domain.ExitUsage
	}

	oldName := args[0]
	newName := args[1]

	err := c.service.RenameList(ctx, oldName, newName)
	if err != nil {
		fmt.Fprintf(errOut, ui.Red("Error: %v\n"), err)
		if err == domain.ErrListNotFound {
			return domain.ExitNotFound
		}
		return domain.ExitFailure
	}

	fmt.Fprintf(out, "Renamed list '%s' to '%s'.\n", ui.Bold(oldName), ui.Bold(newName))
	return domain.ExitSuccess
}

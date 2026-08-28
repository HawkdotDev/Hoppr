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
		fmt.Fprintln(errOut, ui.Red("Error: Please provide a project name."))
		fmt.Fprintf(errOut, "Usage: %s\n", c.Usage())
		return domain.ExitUsage
	}

	name := args[0]
	list := ""
	if len(args) >= 2 {
		list = args[1]
	}

	finalName, finalList, err := c.service.RemoveProject(ctx, name, list)
	if err != nil {
		fmt.Fprintf(errOut, ui.Red("Error: %v\n"), err)
		if err == domain.ErrProjectNotFound || err == domain.ErrListNotFound {
			return domain.ExitNotFound
		}
		return domain.ExitFailure
	}

	fmt.Fprintf(out, "Removed project '%s' from list '%s'.\n", ui.Bold(finalName), ui.Yellow(finalList))
	return domain.ExitSuccess
}

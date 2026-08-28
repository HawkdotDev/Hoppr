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
		fmt.Fprintln(errOut, ui.Red("Error: Please provide a list name."))
		fmt.Fprintf(errOut, "Usage: %s\n", c.Usage())
		return domain.ExitUsage
	}

	listName := args[0]
	err := c.service.CreateList(ctx, listName)
	if err != nil {
		fmt.Fprintf(errOut, ui.Red("Error: %v\n"), err)
		return domain.ExitFailure
	}

	fmt.Fprintf(out, "Created empty list '%s'.\n", ui.Bold(listName))
	return domain.ExitSuccess
}

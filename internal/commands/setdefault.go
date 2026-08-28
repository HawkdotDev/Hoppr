package commands

import (
	"context"
	"fmt"
	"io"

	"hoppr/internal/core"
	"hoppr/internal/domain"
	"hoppr/internal/ui"
)

type SetDefaultCommand struct {
	service *core.ProjectService
}

func NewSetDefaultCommand(service *core.ProjectService) *SetDefaultCommand {
	return &SetDefaultCommand{service: service}
}

func (c *SetDefaultCommand) Name() string        { return "setdefault" }
func (c *SetDefaultCommand) Aliases() []string   { return []string{"default", "use"} }
func (c *SetDefaultCommand) Synopsis() string    { return "Set the default list" }
func (c *SetDefaultCommand) Usage() string       { return "hop setdefault <list>" }

func (c *SetDefaultCommand) Execute(ctx context.Context, args []string, out io.Writer, errOut io.Writer) int {
	if len(args) < 1 {
		fmt.Fprintln(errOut, ui.Red("Error: Please provide a list name to set as default."))
		fmt.Fprintf(errOut, "Usage: %s\n", c.Usage())
		return domain.ExitUsage
	}

	listName := args[0]
	err := c.service.SetDefaultList(ctx, listName)
	if err != nil {
		fmt.Fprintf(errOut, ui.Red("Error: %v\n"), err)
		return domain.ExitFailure
	}

	fmt.Fprintf(out, "Set '%s' as the default list.\n", ui.Bold(listName))
	return domain.ExitSuccess
}

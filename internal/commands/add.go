package commands

import (
	"context"
	"fmt"
	"io"

	"hoppr/internal/core"
	"hoppr/internal/domain"
	"hoppr/internal/ui"
)

type AddCommand struct {
	service *core.ProjectService
}

func NewAddCommand(service *core.ProjectService) *AddCommand {
	return &AddCommand{service: service}
}

func (c *AddCommand) Name() string        { return "add" }
func (c *AddCommand) Aliases() []string   { return []string{"a"} }
func (c *AddCommand) Synopsis() string    { return "Save current directory as a project" }
func (c *AddCommand) Usage() string       { return "hop add <name> [list]" }

func (c *AddCommand) Execute(ctx context.Context, args []string, out io.Writer, errOut io.Writer) int {
	if len(args) < 1 {
		fmt.Fprintln(errOut, ui.Red("Error: Please provide a project name (or '.' for current directory name)."))
		fmt.Fprintf(errOut, "Usage: %s\n", c.Usage())
		return domain.ExitUsage
	}

	name := args[0]
	list := ""
	if len(args) >= 2 {
		list = args[1]
	}

	finalName, finalList, path, err := c.service.AddProject(ctx, name, list)
	if err != nil {
		fmt.Fprintf(errOut, ui.Red("Error: %v\n"), err)
		return domain.ExitFailure
	}

	fmt.Fprintf(out, "Added project '%s' -> %s [%s]\n", ui.Bold(finalName), ui.Cyan(path), ui.Yellow(finalList))
	return domain.ExitSuccess
}

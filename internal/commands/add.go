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
		ui.ErrorMsg(errOut, "Please provide a project name (or '.' for current directory name).")
		fmt.Fprintf(errOut, "  %s %s\n", ui.Gray("Usage:"), ui.Cyan(c.Usage()))
		return domain.ExitUsage
	}

	name := args[0]
	list := ""
	if len(args) >= 2 {
		list = args[1]
	}

	finalName, finalList, path, err := c.service.AddProject(ctx, name, list)
	if err != nil {
		ui.ErrorMsg(errOut, "%v", err)
		return domain.ExitFailure
	}

	ui.SuccessMsg(out, "Saved project %s %s %s %s",
		ui.Cyan(ui.Bold(finalName)),
		ui.Gray(ui.IconArrow),
		ui.Gray(path),
		ui.Violet(fmt.Sprintf("[%s]", finalList)),
	)
	return domain.ExitSuccess
}

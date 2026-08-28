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
func (c *SetDefaultCommand) Aliases() []string   { return []string{"use", "default", "set-default"} }
func (c *SetDefaultCommand) Synopsis() string    { return "Set the default list" }
func (c *SetDefaultCommand) Usage() string       { return "hop setdefault <list>" }

func (c *SetDefaultCommand) Execute(ctx context.Context, args []string, out io.Writer, errOut io.Writer) int {
	if len(args) < 1 {
		ui.ErrorMsg(errOut, "Please provide a list name.")
		fmt.Fprintf(errOut, "  %s %s\n", ui.Gray("Usage:"), ui.Cyan(c.Usage()))
		return domain.ExitUsage
	}

	listName := args[0]
	err := c.service.SetDefaultList(ctx, listName)
	if err != nil {
		ui.ErrorMsg(errOut, "%v", err)
		return domain.ExitFailure
	}

	ui.SuccessMsg(out, "Active default list set to %s %s",
		ui.Violet(fmt.Sprintf("[%s]", listName)),
		ui.Amber("(★ default)"),
	)
	return domain.ExitSuccess
}

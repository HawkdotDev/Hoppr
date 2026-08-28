package commands

import (
	"context"
	"fmt"
	"io"

	"hoppr/internal/core"
	"hoppr/internal/domain"
	"hoppr/internal/ui"
)

type ImportCommand struct {
	service *core.ProjectService
}

func NewImportCommand(service *core.ProjectService) *ImportCommand {
	return &ImportCommand{service: service}
}

func (c *ImportCommand) Name() string        { return "import" }
func (c *ImportCommand) Aliases() []string   { return []string{"scan", "load"} }
func (c *ImportCommand) Synopsis() string    { return "Bulk-import a folder's subfolders as projects" }
func (c *ImportCommand) Usage() string       { return "hop import <list> <folder>" }

func (c *ImportCommand) Execute(ctx context.Context, args []string, out io.Writer, errOut io.Writer) int {
	if len(args) < 2 {
		fmt.Fprintln(errOut, ui.Red("Error: Please provide target list and folder path (use '.' for defaults)."))
		fmt.Fprintf(errOut, "Usage: %s\n", c.Usage())
		return domain.ExitUsage
	}

	list := args[0]
	folder := args[1]

	count, finalList, finalFolder, err := c.service.Import(ctx, list, folder)
	if err != nil {
		fmt.Fprintf(errOut, ui.Red("Error: %v\n"), err)
		return domain.ExitFailure
	}

	fmt.Fprintf(out, "Successfully imported %s projects from '%s' into list '%s'.\n",
		ui.Bold(fmt.Sprintf("%d", count)),
		ui.Cyan(finalFolder),
		ui.Yellow(finalList),
	)
	return domain.ExitSuccess
}

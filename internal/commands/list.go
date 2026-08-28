package commands

import (
	"context"
	"fmt"
	"io"
	"strings"

	"hoppr/internal/core"
	"hoppr/internal/domain"
	"hoppr/internal/ui"
)

type ListCommand struct {
	service *core.ProjectService
}

func NewListCommand(service *core.ProjectService) *ListCommand {
	return &ListCommand{service: service}
}

func (c *ListCommand) Name() string        { return "list" }
func (c *ListCommand) Aliases() []string   { return []string{"ls", "l"} }
func (c *ListCommand) Synopsis() string    { return "Show all lists, or projects in a specific list" }
func (c *ListCommand) Usage() string       { return "hop list [list] [--plain] [--json]" }

func (c *ListCommand) Execute(ctx context.Context, args []string, out io.Writer, errOut io.Writer) int {
	isPlain := false
	isJSON := false
	var targetList string

	for _, arg := range args {
		switch arg {
		case "--plain", "-p":
			isPlain = true
		case "--json":
			isJSON = true
		default:
			if !strings.HasPrefix(arg, "-") && targetList == "" {
				targetList = arg
			}
		}
	}

	if targetList != "" && targetList != "." {
		projects, isDefault, err := c.service.GetList(ctx, targetList)
		if err != nil {
			fmt.Fprintf(errOut, ui.Red("Error: %v\n"), err)
			return domain.ExitNotFound
		}

		if isJSON {
			_ = ui.RenderJSON(out, projects)
			return domain.ExitSuccess
		}
		if isPlain {
			ui.RenderPlain(out, projects)
			return domain.ExitSuccess
		}

		ui.RenderProjectsTable(out, targetList, projects, isDefault)
		return domain.ExitSuccess
	}

	// List all
	lists, defaultList, err := c.service.GetAllLists(ctx)
	if err != nil {
		fmt.Fprintf(errOut, ui.Red("Error: %v\n"), err)
		return domain.ExitFailure
	}

	if isJSON {
		_ = ui.RenderJSON(out, lists)
		return domain.ExitSuccess
	}

	if isPlain {
		// Output all projects in flat tab-separated format
		flatProjects := make(map[string]string)
		for _, projects := range lists {
			for name, path := range projects {
				flatProjects[name] = path
			}
		}
		ui.RenderPlain(out, flatProjects)
		return domain.ExitSuccess
	}

	ui.RenderAllLists(out, lists, defaultList)
	return domain.ExitSuccess
}

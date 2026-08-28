package commands

import (
	"context"
	"fmt"
	"io"
	"slices"
	"strings"

	"hoppr/internal/core"
	"hoppr/internal/domain"
)

// GetPathCommand is the fast-path query used by the shell jumper wrapper.
type GetPathCommand struct {
	service *core.ProjectService
}

func NewGetPathCommand(service *core.ProjectService) *GetPathCommand {
	return &GetPathCommand{service: service}
}

func (c *GetPathCommand) Name() string        { return "_get_path" }
func (c *GetPathCommand) Aliases() []string   { return nil }
func (c *GetPathCommand) Synopsis() string    { return "Internal routine to return project path" }
func (c *GetPathCommand) Usage() string       { return "hop _get_path <name> [list]" }

func (c *GetPathCommand) Execute(ctx context.Context, args []string, out io.Writer, errOut io.Writer) int {
	if len(args) < 1 {
		return domain.ExitUsage
	}

	name := args[0]
	list := ""
	if len(args) >= 2 {
		list = args[1]
	}

	path, err := c.service.GetPath(ctx, name, list)
	if err != nil {
		return domain.ExitNotFound
	}

	// Output clean path to stdout without trailing newline for shell consumption
	fmt.Fprint(out, path)
	return domain.ExitSuccess
}

// GetEditorCommand is the fast query to output the configured editor.
type GetEditorCommand struct {
	service *core.ProjectService
}

func NewGetEditorCommand(service *core.ProjectService) *GetEditorCommand {
	return &GetEditorCommand{service: service}
}

func (c *GetEditorCommand) Name() string        { return "_get_editor" }
func (c *GetEditorCommand) Aliases() []string   { return nil }
func (c *GetEditorCommand) Synopsis() string    { return "Internal routine to return editor executable" }
func (c *GetEditorCommand) Usage() string       { return "hop _get_editor" }

func (c *GetEditorCommand) Execute(ctx context.Context, args []string, out io.Writer, errOut io.Writer) int {
	editor, err := c.service.GetEditor(ctx)
	if err != nil {
		return domain.ExitFailure
	}
	fmt.Fprint(out, editor)
	return domain.ExitSuccess
}

// CompleteCommand provides dynamic completions for shell tabs.
type CompleteCommand struct {
	service *core.ProjectService
}

func NewCompleteCommand(service *core.ProjectService) *CompleteCommand {
	return &CompleteCommand{service: service}
}

func (c *CompleteCommand) Name() string        { return "_complete" }
func (c *CompleteCommand) Aliases() []string   { return nil }
func (c *CompleteCommand) Synopsis() string    { return "Internal routine for shell completion" }
func (c *CompleteCommand) Usage() string       { return "hop _complete" }

func (c *CompleteCommand) Execute(ctx context.Context, args []string, out io.Writer, errOut io.Writer) int {
	lists, _, err := c.service.GetAllLists(ctx)
	if err != nil {
		return domain.ExitFailure
	}

	commands := []string{
		"add", "remove", "rm", "del", "list", "ls",
		"create", "new", "drop", "rename", "mv",
		"setdefault", "default", "use", "import", "scan",
		"doctor", "check", "--help", "--version",
	}

	candidates := make([]string, 0, len(commands)+len(lists)*4)
	candidates = append(candidates, commands...)

	for listName, projects := range lists {
		candidates = append(candidates, listName)
		for projName := range projects {
			candidates = append(candidates, projName)
		}
	}
	slices.Sort(candidates)
	candidates = slices.Compact(candidates)

	prefix := ""
	if len(args) > 0 {
		prefix = strings.ToLower(args[0])
	}

	for _, cand := range candidates {
		if prefix == "" || strings.HasPrefix(strings.ToLower(cand), prefix) {
			fmt.Fprintln(out, cand)
		}
	}
	return domain.ExitSuccess
}

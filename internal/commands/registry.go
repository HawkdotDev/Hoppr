package commands

import (
	"context"
	"fmt"
	"io"
	"slices"
	"strings"
	"text/tabwriter"

	"hoppr/internal/core"
	"hoppr/internal/domain"
	"hoppr/internal/storage"
	"hoppr/internal/ui"
	"hoppr/internal/version"
)

// Registry manages and dispatches CLI commands.
type Registry struct {
	commands map[string]Command
	aliases  map[string]string
}

// NewRegistry initializes a Registry with all standard subcommands.
func NewRegistry(store storage.StorageEngine) *Registry {
	projectService := core.NewProjectService(store)
	doctorService := core.NewDoctorService(store)

	r := &Registry{
		commands: make(map[string]Command),
		aliases:  make(map[string]string),
	}

	r.Register(NewAddCommand(projectService))
	r.Register(NewRemoveCommand(projectService))
	r.Register(NewListCommand(projectService))
	r.Register(NewCreateCommand(projectService))
	r.Register(NewDropCommand(projectService))
	r.Register(NewRenameCommand(projectService))
	r.Register(NewSetDefaultCommand(projectService))
	r.Register(NewImportCommand(projectService))
	r.Register(NewDoctorCommand(doctorService))
	r.Register(NewGetPathCommand(projectService))
	r.Register(NewGetEditorCommand(projectService))
	r.Register(NewCompleteCommand(projectService))

	return r
}

// Register adds a command to the registry.
func (r *Registry) Register(cmd Command) {
	r.commands[cmd.Name()] = cmd
	for _, alias := range cmd.Aliases() {
		r.aliases[alias] = cmd.Name()
	}
}

// Dispatch executes the matched command or returns a usage error.
func (r *Registry) Dispatch(ctx context.Context, args []string, out io.Writer, errOut io.Writer) int {
	if len(args) == 0 {
		r.PrintUsage(out)
		return domain.ExitSuccess
	}

	cmdName := args[0]
	cmdArgs := args[1:]

	// Fast paths for help and version flags
	switch cmdName {
	case "--help", "-h", "help":
		r.PrintUsage(out)
		return domain.ExitSuccess
	case "--version", "-v", "version":
		r.PrintVersion(out)
		return domain.ExitSuccess
	}

	// Check canonical command or alias
	if aliasTarget, ok := r.aliases[cmdName]; ok {
		cmdName = aliasTarget
	}

	if cmd, exists := r.commands[cmdName]; exists {
		return cmd.Execute(ctx, cmdArgs, out, errOut)
	}

	fmt.Fprintf(errOut, ui.Red("Error: Unknown command '%s'\n"), cmdName)
	r.PrintUsage(errOut)
	return domain.ExitUsage
}

// PrintUsage prints the main CLI help menu.
func (r *Registry) PrintUsage(w io.Writer) {
	PrintDirectUsage(w)
}

// PrintDirectUsage prints the main CLI help menu directly without requiring a registry instance.
func PrintDirectUsage(w io.Writer) {
	fmt.Fprintf(w, "%s - The fast lane to your favourite projects\n\n", ui.Bold("Hoppr"))
	fmt.Fprintf(w, "%s:\n  hop <command> [arguments]\n\n", ui.Bold("Usage"))

	fmt.Fprintf(w, "%s:\n", ui.Bold("Project Commands"))
	tw := tabwriter.NewWriter(w, 0, 0, 3, ' ', 0)
	fmt.Fprintf(tw, "  %s\t%s\n", ui.Cyan("add <name> [list]"), "Save current directory as a project (use '.' for defaults)")
	fmt.Fprintf(tw, "  %s\t%s\n", ui.Cyan("remove <name> [list]"), "Remove a project from a list")
	tw.Flush()

	fmt.Fprintf(w, "\n%s:\n", ui.Bold("List Commands"))
	tw = tabwriter.NewWriter(w, 0, 0, 3, ' ', 0)
	fmt.Fprintf(tw, "  %s\t%s\n", ui.Cyan("create <list>"), "Create a new empty list")
	fmt.Fprintf(tw, "  %s\t%s\n", ui.Cyan("list [list]"), "Show all lists or projects in a specific list (--plain, --json)")
	fmt.Fprintf(tw, "  %s\t%s\n", ui.Cyan("import <list> <folder>"), "Bulk-import a folder's subfolders as projects")
	fmt.Fprintf(tw, "  %s\t%s\n", ui.Cyan("drop <list>"), "Delete an entire list")
	fmt.Fprintf(tw, "  %s\t%s\n", ui.Cyan("rename <list> <new>"), "Rename a list")
	fmt.Fprintf(tw, "  %s\t%s\n", ui.Cyan("setdefault <list>"), "Set the default list")
	tw.Flush()

	fmt.Fprintf(w, "\n%s:\n", ui.Bold("Diagnostics & Utilities"))
	tw = tabwriter.NewWriter(w, 0, 0, 3, ' ', 0)
	fmt.Fprintf(tw, "  %s\t%s\n", ui.Cyan("doctor"), "Inspect environment health and validate saved paths")
	fmt.Fprintf(tw, "  %s\t%s\n", ui.Cyan("--version, -v"), "Show version and build metadata")
	fmt.Fprintf(tw, "  %s\t%s\n", ui.Cyan("--help, -h"), "Show this help message")
	tw.Flush()

	fmt.Fprintf(w, "\n%s:\n", ui.Bold("Shorthand '.' Resolution"))
	fmt.Fprintf(w, "  %s resolves to current directory\n", ui.Yellow("<folder> ."))
	fmt.Fprintf(w, "  %s resolves to current folder's basename\n", ui.Yellow("<name>   ."))
	fmt.Fprintf(w, "  %s resolves to the active default list\n", ui.Yellow("<list>   ."))
}

// PrintVersion outputs version details.
func (r *Registry) PrintVersion(w io.Writer) {
	fmt.Fprintf(w, "hop version %s (commit: %s, built: %s)\n",
		ui.Bold(version.Version),
		version.Commit,
		version.BuildDate,
	)
}

// GetVisibleCommandNames returns all non-hidden command names.
func (r *Registry) GetVisibleCommandNames() []string {
	names := make([]string, 0)
	for name := range r.commands {
		if !strings.HasPrefix(name, "_") {
			names = append(names, name)
		}
	}
	slices.Sort(names)
	return names
}

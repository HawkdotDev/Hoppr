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
	updateService := core.NewUpdateService()

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
	r.Register(NewUpdateCommand(updateService))
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

	// Calculate visible candidates for fuzzy match
	candidates := r.GetVisibleCommandNames()
	for alias := range r.aliases {
		if !strings.HasPrefix(alias, "_") {
			candidates = append(candidates, alias)
		}
	}

	suggestion := ui.FindClosestMatch(cmdName, candidates, 3)
	if suggestion != "" {
		ui.ErrorMsg(errOut, "Unknown command '%s'. Did you mean '%s'?", ui.Bold(cmdName), ui.Cyan(ui.Bold(suggestion)))
	} else {
		ui.ErrorMsg(errOut, "Unknown command '%s'", cmdName)
	}
	fmt.Fprintf(errOut, "\n  %s Run '%s' to see all available commands.\n\n", ui.Cyan("💡 Tip:"), ui.Cyan("hop --help"))
	return domain.ExitUsage
}

// PrintUsage prints the main CLI help menu.
func (r *Registry) PrintUsage(w io.Writer) {
	PrintDirectUsage(w)
}

// PrintDirectUsage prints the main CLI help menu directly without requiring a registry instance.
func PrintDirectUsage(w io.Writer) {
	fmt.Fprintf(w, "\n%s %s — %s\n\n",
		ui.Violet(ui.IconZap),
		ui.Bold("Hoppr"),
		ui.Gray("The fast lane to your favourite projects"),
	)

	fmt.Fprintf(w, "%s\n", ui.Bold("Usage:"))
	fmt.Fprintf(w, "  %s %s\n", ui.Cyan("hop"), ui.Gray("<command> [arguments]"))
	fmt.Fprintf(w, "  %s %s\n\n", ui.Cyan("hop"), ui.Gray("<project>               Jump directly into project"))

	fmt.Fprintf(w, "%s\n", ui.Bold("Project Shortcuts:"))
	tw := tabwriter.NewWriter(w, 0, 0, 3, ' ', 0)
	fmt.Fprintf(tw, "  %s\t%s\n", ui.Cyan("add <name> [list]"), ui.Gray("Save current directory (use '.' for defaults)"))
	fmt.Fprintf(tw, "  %s\t%s\n", ui.Cyan("remove <name> [list]"), ui.Gray("Remove a project shortcut"))
	tw.Flush()

	fmt.Fprintf(w, "\n%s\n", ui.Bold("List Management:"))
	tw = tabwriter.NewWriter(w, 0, 0, 3, ' ', 0)
	fmt.Fprintf(tw, "  %s\t%s\n", ui.Cyan("list [list]"), ui.Gray("Display saved projects (--plain, --json)"))
	fmt.Fprintf(tw, "  %s\t%s\n", ui.Cyan("create <list>"), ui.Gray("Create a new empty project list"))
	fmt.Fprintf(tw, "  %s\t%s\n", ui.Cyan("import <list> <folder>"), ui.Gray("Bulk-import subfolders as projects"))
	fmt.Fprintf(tw, "  %s\t%s\n", ui.Cyan("drop <list>"), ui.Gray("Delete an entire project list"))
	fmt.Fprintf(tw, "  %s\t%s\n", ui.Cyan("rename <list> <new>"), ui.Gray("Rename an existing list"))
	fmt.Fprintf(tw, "  %s\t%s\n", ui.Cyan("setdefault <list>"), ui.Gray("Change the active default list"))
	tw.Flush()

	fmt.Fprintf(w, "\n%s\n", ui.Bold("Diagnostics & System:"))
	tw = tabwriter.NewWriter(w, 0, 0, 3, ' ', 0)
	fmt.Fprintf(tw, "  %s\t%s\n", ui.Cyan("doctor"), ui.Gray("Inspect environment and check for broken paths"))
	fmt.Fprintf(tw, "  %s\t%s\n", ui.Cyan("update"), ui.Gray("Upgrade Hoppr to the latest release (self-update)"))
	fmt.Fprintf(tw, "  %s\t%s\n", ui.Cyan("--version, -v"), ui.Gray("Show version and build metadata"))
	fmt.Fprintf(tw, "  %s\t%s\n", ui.Cyan("--help, -h"), ui.Gray("Show this help message"))
	tw.Flush()

	fmt.Fprintf(w, "\n%s\n", ui.Bold("Shorthand '.' Resolution:"))
	fmt.Fprintf(w, "  %s  resolves to current folder's name\n", ui.Amber("<name>   ."))
	fmt.Fprintf(w, "  %s  resolves to current directory path\n", ui.Amber("<folder> ."))
	fmt.Fprintf(w, "  %s  resolves to the active default list\n\n", ui.Amber("<list>   ."))
}

// PrintVersion outputs version details.
func (r *Registry) PrintVersion(w io.Writer) {
	fmt.Fprintf(w, "%s %s %s (commit: %s, built: %s)\n",
		ui.Violet(ui.IconZap),
		ui.Bold("hop"),
		ui.Cyan(version.FormattedVersion()),
		ui.Gray(version.Commit),
		ui.Gray(version.BuildDate),
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

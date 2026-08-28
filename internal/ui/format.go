package ui

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"slices"
	"text/tabwriter"
)

// Message Helpers
func SuccessMsg(w io.Writer, format string, a ...any) {
	icon := Green(IconCheck)
	msg := fmt.Sprintf(format, a...)
	fmt.Fprintf(w, "  %s  %s\n", icon, msg)
}

func ErrorMsg(w io.Writer, format string, a ...any) {
	icon := Red(IconCross)
	msg := fmt.Sprintf(format, a...)
	fmt.Fprintf(w, "  %s  %s\n", icon, Red(msg))
}

func WarnMsg(w io.Writer, format string, a ...any) {
	icon := Yellow(IconWarn)
	msg := fmt.Sprintf(format, a...)
	fmt.Fprintf(w, "  %s  %s\n", icon, Yellow(msg))
}

func InfoMsg(w io.Writer, format string, a ...any) {
	icon := Cyan(IconInfo)
	msg := fmt.Sprintf(format, a...)
	fmt.Fprintf(w, "  %s  %s\n", icon, msg)
}

// RenderProjectsTable formats projects inside a list cleanly with tree styling and health checks.
func RenderProjectsTable(w io.Writer, listName string, projects map[string]string, isDefault bool) {
	// List Header
	countStr := Dim(fmt.Sprintf("(%d projects)", len(projects)))
	if len(projects) == 1 {
		countStr = Dim("(1 project)")
	}

	header := fmt.Sprintf("%s [%s] %s", Violet(IconZap), Bold(listName), countStr)
	if isDefault {
		header = fmt.Sprintf("%s [%s] %s %s", Violet(IconZap), Bold(listName), Amber("★ default"), countStr)
	}
	fmt.Fprintln(w, header)

	if len(projects) == 0 {
		fmt.Fprintln(w, Dim("    (no projects saved in this list)"))
		return
	}

	// Sort project names deterministically
	names := make([]string, 0, len(projects))
	for name := range projects {
		names = append(names, name)
	}
	slices.Sort(names)

	for i, name := range names {
		path := projects[name]
		branch := Gray(IconBranch)
		if i == len(names)-1 {
			branch = Gray(IconLast)
		}

		// Check if path exists on disk
		status := ""
		if _, err := os.Stat(path); os.IsNotExist(err) {
			status = " " + Red("[broken path]")
		}

		fmt.Fprintf(w, "    %s %s  %s %s%s\n",
			branch,
			Cyan(Bold(name)),
			Gray(IconArrow),
			Gray(path),
			status,
		)
	}
}

// RenderAllLists displays all project lists grouped with their projects.
func RenderAllLists(w io.Writer, lists map[string]map[string]string, defaultList string) {
	if len(lists) == 0 {
		fmt.Fprintln(w, Dim("  No lists found. Run 'hop add .' to get started!"))
		return
	}

	listNames := make([]string, 0, len(lists))
	for name := range lists {
		listNames = append(listNames, name)
	}
	slices.Sort(listNames)

	for i, listName := range listNames {
		RenderProjectsTable(w, listName, lists[listName], listName == defaultList)
		if i < len(listNames)-1 {
			fmt.Fprintln(w)
		}
	}
}

// RenderPlain outputs projects in simple tab-separated format for machine piping and scripting.
func RenderPlain(w io.Writer, projects map[string]string) {
	names := make([]string, 0, len(projects))
	for name := range projects {
		names = append(names, name)
	}
	slices.Sort(names)

	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	for _, name := range names {
		fmt.Fprintf(tw, "%s\t%s\n", name, projects[name])
	}
	tw.Flush()
}

// RenderJSON outputs structured data directly to writer.
func RenderJSON(w io.Writer, data any) error {
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	return encoder.Encode(data)
}

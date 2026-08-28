package ui

import (
	"encoding/json"
	"fmt"
	"io"
	"slices"
	"text/tabwriter"
)

// RenderProjectsTable formats projects inside a list cleanly with tabwriter.
func RenderProjectsTable(w io.Writer, listName string, projects map[string]string, isDefault bool) {
	header := fmt.Sprintf("[%s]", listName)
	if isDefault {
		header = fmt.Sprintf("[%s] %s", listName, Dim("(default)"))
	}
	fmt.Fprintln(w, Cyan(header))

	if len(projects) == 0 {
		fmt.Fprintln(w, Dim("  (no projects saved in this list)"))
		return
	}

	// Sort project names deterministically
	names := make([]string, 0, len(projects))
	for name := range projects {
		names = append(names, name)
	}
	slices.Sort(names)

	tw := tabwriter.NewWriter(w, 0, 0, 3, ' ', 0)
	for _, name := range names {
		fmt.Fprintf(tw, "  %s\t: %s\n", Bold(name), Dim(projects[name]))
	}
	tw.Flush()
}

// RenderAllLists displays all project lists grouped with their projects.
func RenderAllLists(w io.Writer, lists map[string]map[string]string, defaultList string) {
	if len(lists) == 0 {
		fmt.Fprintln(w, Dim("No lists found."))
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

// RenderPlain outputs projects in simple tab-separated format for machine piping.
func RenderPlain(w io.Writer, projects map[string]string) {
	names := make([]string, 0, len(projects))
	for name := range projects {
		names = append(names, name)
	}
	slices.Sort(names)

	for _, name := range names {
		fmt.Fprintf(w, "%s\t%s\n", name, projects[name])
	}
}

// RenderJSON outputs structured data directly to writer.
func RenderJSON(w io.Writer, data any) error {
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	return encoder.Encode(data)
}

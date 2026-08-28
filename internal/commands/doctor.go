package commands

import (
	"context"
	"fmt"
	"io"

	"hoppr/internal/core"
	"hoppr/internal/domain"
	"hoppr/internal/ui"
)

type DoctorCommand struct {
	doctor *core.DoctorService
}

func NewDoctorCommand(doctor *core.DoctorService) *DoctorCommand {
	return &DoctorCommand{doctor: doctor}
}

func (c *DoctorCommand) Name() string        { return "doctor" }
func (c *DoctorCommand) Aliases() []string   { return []string{"check", "diag"} }
func (c *DoctorCommand) Synopsis() string    { return "Inspect environment and validate project paths" }
func (c *DoctorCommand) Usage() string       { return "hop doctor" }

func (c *DoctorCommand) Execute(ctx context.Context, args []string, out io.Writer, errOut io.Writer) int {
	fmt.Fprintln(out, ui.Bold("Running Hoppr diagnostics..."))
	fmt.Fprintln(out)

	results, err := c.doctor.RunDiagnostics(ctx)
	if err != nil {
		fmt.Fprintf(errOut, ui.Red("Doctor error: %v\n"), err)
		return domain.ExitFailure
	}

	allPassed := true
	for _, res := range results {
		if res.Passed {
			fmt.Fprintf(out, "  %s  %s: %s\n", ui.Green("✓"), ui.Bold(res.Title), ui.Dim(res.Message))
		} else {
			allPassed = false
			fmt.Fprintf(out, "  %s  %s: %s\n", ui.Red("✗"), ui.Bold(res.Title), ui.Yellow(res.Message))
		}
	}

	fmt.Fprintln(out)
	if allPassed {
		fmt.Fprintln(out, ui.Green("All checks passed! Your environment is healthy."))
		return domain.ExitSuccess
	}

	fmt.Fprintln(out, ui.Yellow("Some issues were found in your environment. Check the items above."))
	return domain.ExitFailure
}

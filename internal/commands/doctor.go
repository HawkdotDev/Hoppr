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
	fmt.Fprintf(out, "\n%s %s\n\n", ui.Violet(ui.IconZap), ui.Bold("Hoppr Environment Diagnostics"))

	results, err := c.doctor.RunDiagnostics(ctx)
	if err != nil {
		ui.ErrorMsg(errOut, "Diagnostics failed: %v", err)
		return domain.ExitFailure
	}

	passedCount := 0
	failedCount := 0

	for _, res := range results {
		if res.Passed {
			passedCount++
			fmt.Fprintf(out, "  %s  %-20s %s\n",
				ui.Green(ui.IconCheck),
				ui.Bold(res.Title),
				ui.Gray(res.Message),
			)
		} else {
			failedCount++
			fmt.Fprintf(out, "  %s  %-20s %s\n",
				ui.Red(ui.IconCross),
				ui.Bold(res.Title),
				ui.Red(res.Message),
			)
		}
	}

	fmt.Fprintln(out)
	divider := ui.Gray("───────────────────────────────────────────────")
	fmt.Fprintln(out, divider)

	if failedCount == 0 {
		fmt.Fprintf(out, "  %s %s (%d checks passed)\n\n",
			ui.Green(ui.IconCheck),
			ui.Green(ui.Bold("All checks passed! Your environment is healthy.")),
			passedCount,
		)
		return domain.ExitSuccess
	}

	fmt.Fprintf(out, "  %s %s (%d passed, %d issues)\n",
		ui.Yellow(ui.IconWarn),
		ui.Yellow(ui.Bold("Some issues were detected in your configuration.")),
		passedCount,
		failedCount,
	)
	fmt.Fprintf(out, "  %s Run '%s' to update or '%s' to clean up.\n\n",
		ui.Cyan("💡 Tip:"),
		ui.Cyan("hop add ."),
		ui.Cyan("hop remove <name>"),
	)
	return domain.ExitFailure
}

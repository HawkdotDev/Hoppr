package commands

import (
	"context"
	"fmt"
	"io"

	"hoppr/internal/core"
	"hoppr/internal/domain"
	"hoppr/internal/ui"
	"hoppr/internal/version"
)

type UpdateCommand struct {
	updater *core.UpdateService
}

func NewUpdateCommand(updater *core.UpdateService) *UpdateCommand {
	return &UpdateCommand{updater: updater}
}

func (c *UpdateCommand) Name() string        { return "update" }
func (c *UpdateCommand) Aliases() []string   { return []string{"upgrade", "self-update"} }
func (c *UpdateCommand) Synopsis() string    { return "Check for updates and upgrade Hoppr to the latest version" }
func (c *UpdateCommand) Usage() string       { return "hop update" }

func (c *UpdateCommand) Execute(ctx context.Context, args []string, out io.Writer, errOut io.Writer) int {
	fmt.Fprintf(out, "\n%s %s\n\n", ui.Violet(ui.IconZap), ui.Bold("Checking for Hoppr updates..."))

	latestTag, hasUpdate, err := c.updater.CheckUpdate(ctx)
	if err != nil {
		ui.ErrorMsg(errOut, "Update check failed: %v", err)
		return domain.ExitFailure
	}

	current := "v" + version.Version
	if !hasUpdate {
		ui.SuccessMsg(out, "Hoppr is already on the latest version (%s).", ui.Emerald(ui.Bold(current)))
		fmt.Fprintln(out)
		return domain.ExitSuccess
	}

	fmt.Fprintf(out, "  %s  New version available: %s %s %s\n\n",
		ui.Amber(ui.IconInfo),
		ui.Gray(current),
		ui.Gray(ui.IconArrow),
		ui.Emerald(ui.Bold(latestTag)),
	)

	err = c.updater.DownloadAndApplyUpdate(ctx, latestTag, func(step string) {
		fmt.Fprintf(out, "  %s  %s\n", ui.Cyan("●"), step)
	})

	if err != nil {
		fmt.Fprintln(errOut)
		ui.ErrorMsg(errOut, "Update failed: %v", err)
		return domain.ExitFailure
	}

	fmt.Fprintln(out)
	ui.SuccessMsg(out, "Successfully updated Hoppr to %s! 🚀", ui.Emerald(ui.Bold(latestTag)))
	fmt.Fprintf(out, "  %s Run '%s' to verify your installation.\n\n", ui.Cyan("💡 Tip:"), ui.Cyan("hop doctor"))
	return domain.ExitSuccess
}

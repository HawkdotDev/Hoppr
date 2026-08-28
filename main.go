package main

import (
	"context"
	"fmt"
	"os"

	"hoppr/internal/commands"
	"hoppr/internal/domain"
	"hoppr/internal/storage"
	"hoppr/internal/ui"
	"hoppr/internal/version"
)

func main() {
	// Zero-I/O Fast Path for instant terminal response (<0.5ms)
	if len(os.Args) < 2 {
		commands.PrintDirectUsage(os.Stdout)
		os.Exit(domain.ExitSuccess)
	}

	switch os.Args[1] {
	case "--version", "-v", "version":
		fmt.Fprintf(os.Stdout, "%s %s %s (commit: %s, built: %s)\n",
			ui.Violet(ui.IconZap),
			ui.Bold("hop"),
			ui.Cyan("v"+version.Version),
			ui.Gray(version.Commit),
			ui.Gray(version.BuildDate),
		)
		os.Exit(domain.ExitSuccess)
	case "--help", "-h", "help":
		commands.PrintDirectUsage(os.Stdout)
		os.Exit(domain.ExitSuccess)
	}

	// Initialize Storage Engine with atomic write and cross-process lock capabilities
	store, err := storage.NewJSONStorage("")
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: %v\n", ui.Red("Storage Error"), err)
		os.Exit(domain.ExitFailure)
	}

	// Initialize Command Dispatcher Registry
	registry := commands.NewRegistry(store)

	ctx := context.Background()
	exitCode := registry.Dispatch(ctx, os.Args[1:], os.Stdout, os.Stderr)
	os.Exit(exitCode)
}
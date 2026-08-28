package commands

import (
	"context"
	"io"
)

// Command represents an executable CLI subcommand.
type Command interface {
	Name() string
	Aliases() []string
	Synopsis() string
	Usage() string
	Execute(ctx context.Context, args []string, out io.Writer, errOut io.Writer) int
}

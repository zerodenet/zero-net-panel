package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

// NewVersionCommand prints build metadata.
func NewVersionCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Print version information",
		RunE: func(cmd *cobra.Command, args []string) error {
			out := cmd.OutOrStdout()
			fmt.Fprintf(out, "version: %s\ncommit: %s\nbuildDate: %s\n", Version, Commit, BuildDate)
			return nil
		},
	}
	return cmd
}

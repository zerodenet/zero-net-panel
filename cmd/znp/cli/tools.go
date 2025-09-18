package cli

import "github.com/spf13/cobra"

func NewToolsCommand(opts *GlobalOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tools",
		Short: "Operations and maintenance helpers",
	}

	cmd.AddCommand(
		NewToolsCheckConfigCommand(opts),
	)

	return cmd
}

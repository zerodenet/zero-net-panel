package cli

import (
	"context"

	"github.com/spf13/cobra"
)

func NewRootCommand() *cobra.Command {
	opts := &GlobalOptions{}

	cmd := &cobra.Command{
		Use:           "znp",
		Short:         "Zero Network Panel control CLI",
		Long:          "Zero Network Panel command line interface for managing API, gRPC and maintenance tasks.",
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	cmd.PersistentFlags().StringVarP(&opts.ConfigFile, "config", "f", "etc/znp-api.yaml", "Path to configuration file")

	cmd.AddCommand(
		NewServeCommand(opts),
		NewMigrateCommand(opts),
		NewToolsCommand(opts),
		NewVersionCommand(),
		NewInstallCommand(opts),
	)

	return cmd
}

func Execute(ctx context.Context) error {
	root := NewRootCommand()
	root.SetContext(ctx)
	return root.Execute()
}

package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

func NewToolsCheckConfigCommand(opts *GlobalOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "check-config",
		Short: "Validate configuration file and print a summary",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := loadConfig(opts.ConfigFile)
			if err != nil {
				return err
			}

			cmd.Println("Configuration file validated successfully.")
			cmd.Println(fmt.Sprintf("Service: %s", cfg.Project.Name))
			cmd.Println(fmt.Sprintf("HTTP: %s:%d", cfg.Host, cfg.Port))
			cmd.Println(fmt.Sprintf("Admin route prefix: %s", cfg.Admin.RoutePrefix))
			if cfg.GRPC.Enabled() {
				cmd.Println(fmt.Sprintf("gRPC: %s", cfg.GRPC.ListenOn))
			} else {
				cmd.Println("gRPC: disabled")
			}
			if cfg.Database.IsEmpty() {
				cmd.Println("Database: not configured")
			} else {
				cmd.Println(fmt.Sprintf("Database: %s", cfg.Database.Driver))
			}
			return nil
		},
	}

	return cmd
}

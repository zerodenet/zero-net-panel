package cli

import (
	"github.com/spf13/cobra"

	"github.com/zero-net-panel/zero-net-panel/internal/bootstrap"
)

type serveOptions struct {
	autoMigrate   bool
	seedDemo      bool
	disableGRPC   bool
	targetVersion uint64
}

func NewServeCommand(opts *GlobalOptions) *cobra.Command {
	serveOpts := &serveOptions{autoMigrate: true}

	cmd := &cobra.Command{
		Use:   "serve",
		Short: "Run HTTP and gRPC services as a single unit",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := loadConfig(opts.ConfigFile)
			if err != nil {
				return err
			}

			if serveOpts.disableGRPC {
				cfg.GRPC.SetEnabled(false)
			}

			if serveOpts.autoMigrate || serveOpts.seedDemo {
				if _, err := bootstrap.PrepareDatabase(cmd.Context(), cfg, bootstrap.DatabaseOptions{
					AutoMigrate:   serveOpts.autoMigrate,
					SeedDemo:      serveOpts.seedDemo,
					TargetVersion: serveOpts.targetVersion,
				}); err != nil {
					return err
				}
			}

			return RunServices(cmd.Context(), cfg)
		},
	}

	cmd.Flags().BoolVar(&serveOpts.autoMigrate, "auto-migrate", serveOpts.autoMigrate, "Run database migrations before starting services")
	cmd.Flags().BoolVar(&serveOpts.seedDemo, "seed-demo", serveOpts.seedDemo, "Seed demonstration data after migrations")
	cmd.Flags().BoolVar(&serveOpts.disableGRPC, "disable-grpc", serveOpts.disableGRPC, "Disable the integrated gRPC server")
	cmd.Flags().Uint64Var(&serveOpts.targetVersion, "migrate-to", serveOpts.targetVersion, "Apply migrations up to the specified version before starting services (0 = latest)")

	return cmd
}

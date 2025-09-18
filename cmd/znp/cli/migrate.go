package cli

import (
	"github.com/spf13/cobra"

	"github.com/zero-net-panel/zero-net-panel/internal/bootstrap"
)

func NewMigrateCommand(opts *GlobalOptions) *cobra.Command {
	apply := true
	var seedDemo bool
	var targetVersion uint64

	cmd := &cobra.Command{
		Use:   "migrate",
		Short: "Run database migrations and optional seed tasks",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := loadConfig(opts.ConfigFile)
			if err != nil {
				return err
			}

			if !apply && !seedDemo {
				cmd.Println("Nothing to do: enable --apply and/or --seed-demo to run tasks.")
				return nil
			}

			return bootstrap.PrepareDatabase(cmd.Context(), cfg, bootstrap.DatabaseOptions{
				AutoMigrate:   apply,
				SeedDemo:      seedDemo,
				TargetVersion: targetVersion,
			})
		},
	}

	cmd.Flags().BoolVar(&apply, "apply", apply, "Apply database schema migrations")
	cmd.Flags().BoolVar(&seedDemo, "seed-demo", seedDemo, "Seed demonstration data after migrations")
	cmd.Flags().Uint64Var(&targetVersion, "to", targetVersion, "Run migrations up to the specified version (0 = latest)")

	return cmd
}

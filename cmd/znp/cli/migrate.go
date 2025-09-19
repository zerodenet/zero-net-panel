package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/zero-net-panel/zero-net-panel/internal/bootstrap"
)

func NewMigrateCommand(opts *GlobalOptions) *cobra.Command {
	apply := true
	var seedDemo bool
	var rollback bool
	var targetVersion uint64

	cmd := &cobra.Command{
		Use:   "migrate",
		Short: "Run database migrations, rollbacks, and optional seed tasks",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := loadConfig(opts.ConfigFile)
			if err != nil {
				return err
			}

			if rollback && !apply {
				return fmt.Errorf("--rollback requires --apply to run schema migrations")
			}
			if rollback && targetVersion == 0 {
				return fmt.Errorf("--rollback requires --to to specify the desired version")
			}

			if !apply && !seedDemo {
				cmd.Println("Nothing to do: enable --apply and/or --seed-demo to run tasks.")
				return nil
			}

			result, err := bootstrap.PrepareDatabase(cmd.Context(), cfg, bootstrap.DatabaseOptions{
				AutoMigrate:   apply,
				SeedDemo:      seedDemo,
				TargetVersion: targetVersion,
				AllowRollback: rollback,
			})
			if err != nil {
				return err
			}

			if apply {
				if len(result.AppliedVersions) == 0 && len(result.RolledBackVersions) == 0 {
					cmd.Printf("Schema state unchanged (current=%d, target=%d).\n", result.AfterVersion, result.TargetVersion)
				} else if len(result.RolledBackVersions) > 0 {
					cmd.Printf("Rollback complete (before=%d, after=%d, target=%d).\n", result.BeforeVersion, result.AfterVersion, result.TargetVersion)
				} else {
					cmd.Printf("Migrations applied (before=%d, after=%d, target=%d).\n", result.BeforeVersion, result.AfterVersion, result.TargetVersion)
				}
			}

			return nil
		},
	}

	cmd.Flags().BoolVar(&apply, "apply", apply, "Apply database schema migrations")
	cmd.Flags().BoolVar(&seedDemo, "seed-demo", seedDemo, "Seed demonstration data after migrations")
	cmd.Flags().BoolVar(&rollback, "rollback", rollback, "Rollback schema to the specified version (requires --to)")
	cmd.Flags().Uint64Var(&targetVersion, "to", targetVersion, "Run migrations up to the specified version (0 = latest)")

	return cmd
}

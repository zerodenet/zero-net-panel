package cli

import (
	"fmt"
	"strings"

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

			var issues []string
			if cfg.Database.IsEmpty() {
				issues = append(issues, "database driver/dsn is required")
			} else {
				switch strings.ToLower(strings.TrimSpace(cfg.Database.Driver)) {
				case "mysql", "postgres", "sqlite":
				default:
					issues = append(issues, fmt.Sprintf("unsupported database driver %q (expected mysql/postgres/sqlite)", cfg.Database.Driver))
				}
			}

			switch strings.ToLower(strings.TrimSpace(cfg.Cache.Provider)) {
			case "memory":
			case "redis":
				if strings.TrimSpace(cfg.Cache.Redis.Host) == "" {
					issues = append(issues, "cache provider redis requires Cache.Redis.Host")
				}
			case "":
				issues = append(issues, "cache provider is required (memory or redis)")
			default:
				issues = append(issues, fmt.Sprintf("unsupported cache provider %q (use memory or redis)", cfg.Cache.Provider))
			}

			if strings.TrimSpace(cfg.Kernel.DefaultProtocol) == "" {
				issues = append(issues, "kernel.defaultProtocol is required (http or grpc)")
			}
			if strings.EqualFold(cfg.Kernel.DefaultProtocol, "http") && strings.TrimSpace(cfg.Kernel.HTTP.BaseURL) == "" && strings.TrimSpace(cfg.Kernel.GRPC.Endpoint) == "" {
				issues = append(issues, "kernel HTTP base URL is empty; set Kernel.HTTP.BaseURL or switch defaultProtocol")
			}

			if strings.TrimSpace(cfg.Auth.AccessSecret) == "" || strings.TrimSpace(cfg.Auth.RefreshSecret) == "" {
				issues = append(issues, "auth secrets are required")
			}

			if len(issues) > 0 {
				return fmt.Errorf("configuration validation failed:\n- %s", strings.Join(issues, "\n- "))
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
			if len(cfg.Admin.Access.AllowCIDRs) > 0 {
				cmd.Println(fmt.Sprintf("Admin allowlist CIDRs: %s", strings.Join(cfg.Admin.Access.AllowCIDRs, ",")))
			} else {
				cmd.Println("Admin allowlist CIDRs: none")
			}
			if cfg.Admin.Access.RateLimitPerMinute > 0 {
				cmd.Println(fmt.Sprintf("Admin rate limit: %d req/min (burst=%d)", cfg.Admin.Access.RateLimitPerMinute, cfg.Admin.Access.Burst))
			} else {
				cmd.Println("Admin rate limit: disabled")
			}
			if len(cfg.Webhook.AllowCIDRs) > 0 {
				cmd.Println(fmt.Sprintf("Webhook allowlist CIDRs: %s", strings.Join(cfg.Webhook.AllowCIDRs, ",")))
			} else {
				cmd.Println("Webhook allowlist CIDRs: none")
			}
			if cfg.Webhook.SharedToken != "" {
				cmd.Println("Webhook shared token: set")
			} else {
				cmd.Println("Webhook shared token: not set")
			}
			if cfg.Webhook.Stripe.SigningSecret != "" {
				cmd.Println(fmt.Sprintf("Webhook Stripe signature: enabled (tolerance=%ds)", cfg.Webhook.Stripe.ToleranceSeconds))
			} else {
				cmd.Println("Webhook Stripe signature: disabled")
			}
			return nil
		},
	}

	return cmd
}

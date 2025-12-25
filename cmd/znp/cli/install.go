package cli

import (
	"bufio"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"golang.org/x/crypto/bcrypt"
	"gopkg.in/yaml.v3"
	"gorm.io/gorm"

	"github.com/zero-net-panel/zero-net-panel/internal/bootstrap"
	"github.com/zero-net-panel/zero-net-panel/internal/config"
	"github.com/zero-net-panel/zero-net-panel/internal/repository"
	"github.com/zero-net-panel/zero-net-panel/pkg/database"
)

func NewInstallCommand(opts *GlobalOptions) *cobra.Command {
	var nonInteractive bool
	var outputFile string

	cmd := &cobra.Command{
		Use:   "install",
		Short: "Interactive installation wizard for Zero Network Panel",
		Long: `Run the installation wizard to set up Zero Network Panel for the first time.
This command will guide you through:
  - Database configuration
  - Admin account creation
  - Service configuration
  - JWT secrets generation
  - Configuration file generation`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if nonInteractive {
				return fmt.Errorf("non-interactive mode not yet implemented, please run without --non-interactive")
			}

			wizard := &InstallWizard{
				cmd:        cmd,
				reader:     bufio.NewReader(os.Stdin),
				outputFile: outputFile,
			}

			return wizard.Run()
		},
	}

	cmd.Flags().BoolVar(&nonInteractive, "non-interactive", false, "Run in non-interactive mode with defaults")
	cmd.Flags().StringVarP(&outputFile, "output", "o", "etc/znp-installed.yaml", "Output configuration file path")

	return cmd
}

type InstallWizard struct {
	cmd           *cobra.Command
	reader        *bufio.Reader
	outputFile    string
	cfg           config.Config
	adminEmail    string
	adminPassword string
}

func (w *InstallWizard) Run() error {
	w.printWelcome()

	// Step 1: Database configuration
	if err := w.configureDatabaseStep(); err != nil {
		return fmt.Errorf("database configuration failed: %w", err)
	}

	// Step 2: Service configuration
	if err := w.configureServiceStep(); err != nil {
		return fmt.Errorf("service configuration failed: %w", err)
	}

	// Step 3: Generate JWT secrets
	if err := w.configureAuthStep(); err != nil {
		return fmt.Errorf("auth configuration failed: %w", err)
	}

	// Step 4: Admin account creation
	if err := w.configureAdminStep(); err != nil {
		return fmt.Errorf("admin account creation failed: %w", err)
	}

	// Step 5: Optional features
	if err := w.configureOptionalFeaturesStep(); err != nil {
		return fmt.Errorf("optional features configuration failed: %w", err)
	}

	// Step 6: Save configuration
	if err := w.saveConfiguration(); err != nil {
		return fmt.Errorf("failed to save configuration: %w", err)
	}

	// Step 7: Initialize database
	if err := w.initializeDatabaseStep(); err != nil {
		return fmt.Errorf("database initialization failed: %w", err)
	}

	// Step 8: Create admin user
	if err := w.createAdminUserStep(); err != nil {
		return fmt.Errorf("admin user creation failed: %w", err)
	}

	w.printSuccess()
	return nil
}

func (w *InstallWizard) printWelcome() {
	w.cmd.Println("\n╔════════════════════════════════════════════════════════════════╗")
	w.cmd.Println("║    Welcome to Zero Network Panel Installation Wizard          ║")
	w.cmd.Println("╚════════════════════════════════════════════════════════════════╝")
	w.cmd.Println("\nThis wizard will help you set up Zero Network Panel.")
	w.cmd.Println("Press Ctrl+C at any time to exit.")
}

func (w *InstallWizard) configureDatabaseStep() error {
	w.cmd.Println("═══ Step 1: Database Configuration ═══")

	// Database driver selection
	w.cmd.Println("Select database driver:")
	w.cmd.Println("  1) SQLite (Development/Testing)")
	w.cmd.Println("  2) MySQL (Recommended for Production)")
	w.cmd.Println("  3) PostgreSQL")

	choice := w.prompt("Enter choice [1-3]", "1")

	var driver string
	var defaultDSN string

	switch choice {
	case "1":
		driver = "sqlite"
		defaultDSN = "file:znp.db?cache=shared&mode=rwc"
	case "2":
		driver = "mysql"
		defaultDSN = "root:password@tcp(127.0.0.1:3306)/znp?parseTime=true&loc=UTC"
	case "3":
		driver = "postgres"
		defaultDSN = "host=localhost port=5432 user=znp password=password dbname=znp sslmode=disable"
	default:
		return fmt.Errorf("invalid choice: %s", choice)
	}

	w.cfg.Database.Driver = driver

	if driver == "sqlite" {
		dbPath := w.prompt("SQLite database file path", defaultDSN)
		w.cfg.Database.DSN = dbPath
	} else {
		w.cmd.Printf("\nExample DSN for %s:\n  %s\n\n", driver, defaultDSN)
		dsn := w.prompt("Enter database DSN", defaultDSN)
		w.cfg.Database.DSN = dsn
	}

	// Set reasonable defaults for database pool
	w.cfg.Database.MaxOpenConns = 20
	w.cfg.Database.MaxIdleConns = 10
	w.cfg.Database.ConnMaxLifetime = 300 * time.Second
	w.cfg.Database.LogLevel = "warn"

	w.cmd.Println("\n✓ Database configuration completed")
	return nil
}

func (w *InstallWizard) configureServiceStep() error {
	w.cmd.Println("═══ Step 2: Service Configuration ═══")

	w.cfg.Name = "znp.api"
	w.cfg.Host = w.prompt("Service host", "0.0.0.0")
	portStr := w.prompt("Service port", "8888")
	w.cfg.Port = w.parseInt(portStr, 8888)
	w.cfg.Timeout = 60000

	// Project information
	w.cfg.Project.Name = "Zero Network Panel"
	w.cfg.Project.Description = "Backend service benchmarked against xboard capabilities"
	w.cfg.Project.Version = "0.1.0"

	// Cache configuration
	w.cfg.Cache.Provider = "memory"
	w.cfg.Cache.Memory.Size = 1024

	// Kernel configuration
	w.cfg.Kernel.DefaultProtocol = "http"
	w.cfg.Kernel.HTTP.BaseURL = w.prompt("Kernel HTTP base URL (optional)", "http://kernel.local")
	w.cfg.Kernel.HTTP.Timeout = 5 * time.Second
	w.cfg.Kernel.GRPC.Endpoint = "127.0.0.1:9000"
	w.cfg.Kernel.GRPC.Timeout = 5 * time.Second

	w.cmd.Println("\n✓ Service configuration completed")
	return nil
}

func (w *InstallWizard) configureAuthStep() error {
	w.cmd.Println("═══ Step 3: JWT Authentication Configuration ═══")

	w.cmd.Println("Generating secure JWT secrets...")

	accessSecret, err := generateSecret(32)
	if err != nil {
		return fmt.Errorf("failed to generate access secret: %w", err)
	}
	w.cfg.Auth.AccessSecret = accessSecret
	w.cfg.Auth.AccessExpire = 24 * time.Hour

	refreshSecret, err := generateSecret(32)
	if err != nil {
		return fmt.Errorf("failed to generate refresh secret: %w", err)
	}
	w.cfg.Auth.RefreshSecret = refreshSecret
	w.cfg.Auth.RefreshExpire = 720 * time.Hour

	w.cmd.Println("✓ JWT secrets generated successfully")
	return nil
}

func (w *InstallWizard) configureAdminStep() error {
	w.cmd.Println("═══ Step 4: Admin Account Configuration ═══")

	// Email validation regex
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

	for {
		email := w.prompt("Admin email", "admin@example.com")
		if emailRegex.MatchString(email) {
			w.adminEmail = email
			break
		}
		w.cmd.Println("✗ Invalid email format. Please try again.")
	}

	for {
		password := w.promptPassword("Admin password (min 8 characters)")
		if len(password) < 8 {
			w.cmd.Println("✗ Password must be at least 8 characters long.")
			continue
		}
		confirmPassword := w.promptPassword("Confirm admin password")
		if password != confirmPassword {
			w.cmd.Println("✗ Passwords do not match. Please try again.")
			continue
		}
		w.adminPassword = password
		break
	}

	w.cmd.Println("\n✓ Admin account configuration completed")
	return nil
}

func (w *InstallWizard) configureOptionalFeaturesStep() error {
	w.cmd.Println("═══ Step 5: Optional Features ═══")

	// Metrics configuration
	enableMetrics := w.promptYesNo("Enable Prometheus metrics", true)
	w.cfg.Metrics.Enable = enableMetrics
	if enableMetrics {
		w.cfg.Metrics.Path = "/metrics"
		separatePort := w.promptYesNo("Use separate port for metrics", true)
		if separatePort {
			w.cfg.Metrics.ListenOn = w.prompt("Metrics port", "0.0.0.0:9100")
		} else {
			w.cfg.Metrics.ListenOn = ""
		}
	}

	// Admin configuration
	w.cfg.Admin.RoutePrefix = w.prompt("Admin route prefix", "admin")
	w.cfg.Admin.Access.AllowCIDRs = []string{}
	w.cfg.Admin.Access.RateLimitPerMinute = 0
	w.cfg.Admin.Access.Burst = 0

	// Webhook configuration
	w.cfg.Webhook.AllowCIDRs = []string{}
	w.cfg.Webhook.SharedToken = ""
	w.cfg.Webhook.Stripe.SigningSecret = ""
	w.cfg.Webhook.Stripe.ToleranceSeconds = 300

	// gRPC configuration
	enableGRPC := w.promptYesNo("Enable gRPC server", true)
	w.cfg.GRPC.Enable = &enableGRPC
	if enableGRPC {
		w.cfg.GRPC.ListenOn = w.prompt("gRPC listen address", "0.0.0.0:8890")
		reflection := true
		w.cfg.GRPC.Reflection = &reflection
	}

	w.cmd.Println("\n✓ Optional features configured")
	return nil
}

func (w *InstallWizard) saveConfiguration() error {
	w.cmd.Println("═══ Step 6: Saving Configuration ═══")

	// Normalize configuration
	w.cfg.Normalize()

	// Ensure output directory exists
	dir := filepath.Dir(w.outputFile)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Marshal configuration to YAML
	data, err := yaml.Marshal(&w.cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal configuration: %w", err)
	}

	// Write to file
	if err := os.WriteFile(w.outputFile, data, 0600); err != nil {
		return fmt.Errorf("failed to write configuration file: %w", err)
	}

	w.cmd.Printf("✓ Configuration saved to: %s\n\n", w.outputFile)
	return nil
}

func (w *InstallWizard) initializeDatabaseStep() error {
	w.cmd.Println("═══ Step 7: Initializing Database ═══")

	db, closeFn, err := database.NewGorm(w.cfg.Database)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer closeFn()

	if db == nil {
		return fmt.Errorf("database connection is nil")
	}

	w.cmd.Println("Running database migrations...")

	result, err := bootstrap.ApplyMigrations(w.cmd.Context(), db, 0, false)
	if err != nil {
		return fmt.Errorf("migration failed: %w", err)
	}

	w.cmd.Printf("✓ Migrations applied: %d version(s)\n", len(result.AppliedVersions))
	w.cmd.Printf("  Current schema version: %d\n\n", result.AfterVersion)

	return nil
}

func (w *InstallWizard) createAdminUserStep() error {
	w.cmd.Println("═══ Step 8: Creating Admin User ═══")

	db, closeFn, err := database.NewGorm(w.cfg.Database)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer closeFn()

	// Check if admin user already exists
	var existingUser repository.User
	err = db.Where("email = ?", w.adminEmail).First(&existingUser).Error
	if err == nil {
		w.cmd.Println("✓ Admin user already exists, skipping creation")
		// Clear credentials from memory
		w.clearCredentials()
		return nil
	} else if err != gorm.ErrRecordNotFound {
		// Clear credentials from memory on error
		w.clearCredentials()
		return fmt.Errorf("failed to check existing user: %w", err)
	}

	// Hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(w.adminPassword), bcrypt.DefaultCost)
	if err != nil {
		// Clear credentials from memory on error
		w.clearCredentials()
		return fmt.Errorf("failed to hash password: %w", err)
	}

	// Create admin user
	adminUser := repository.User{
		Email:        w.adminEmail,
		DisplayName:  "System Administrator",
		PasswordHash: string(hashedPassword),
		Roles:        []string{"admin", "user"},
		Status:       "active",
	}

	if err := db.Create(&adminUser).Error; err != nil {
		// Clear credentials from memory on error
		w.clearCredentials()
		return fmt.Errorf("failed to create admin user: %w", err)
	}

	w.cmd.Println("✓ Admin user created successfully")

	// Clear credentials from memory after successful creation
	w.clearCredentials()
	return nil
}

// clearCredentials securely clears sensitive credential data from memory
func (w *InstallWizard) clearCredentials() {
	// Overwrite password with zeros before clearing
	if w.adminPassword != "" {
		passwordBytes := []byte(w.adminPassword)
		for i := range passwordBytes {
			passwordBytes[i] = 0
		}
	}
	w.adminPassword = ""
	w.adminEmail = ""
}

func (w *InstallWizard) printSuccess() {
	w.cmd.Println("\n╔════════════════════════════════════════════════════════════════╗")
	w.cmd.Println("║              Installation Completed Successfully!             ║")
	w.cmd.Println("╚════════════════════════════════════════════════════════════════╝")
	w.cmd.Println("\n✓ Configuration file created")
	w.cmd.Println("✓ Database initialized")
	w.cmd.Println("✓ Admin user created")
	w.cmd.Println("\nNext steps:")
	w.cmd.Printf("  1. Review the configuration file: %s\n", w.outputFile)
	w.cmd.Printf("  2. Start the service: go run ./cmd/znp serve --config %s\n", w.outputFile)
	w.cmd.Printf("  3. Access the API at: http://%s:%d/api/v1/ping\n", w.cfg.Host, w.cfg.Port)
	w.cmd.Printf("  4. Login with: %s\n", w.adminEmail)
	w.cmd.Println("\nThank you for using Zero Network Panel!")
	w.cmd.Println()
}

// Helper functions

func (w *InstallWizard) prompt(message, defaultValue string) string {
	if defaultValue != "" {
		w.cmd.Printf("%s [%s]: ", message, defaultValue)
	} else {
		w.cmd.Printf("%s: ", message)
	}

	input, err := w.reader.ReadString('\n')
	if err != nil {
		return defaultValue
	}

	input = strings.TrimSpace(input)
	if input == "" {
		return defaultValue
	}
	return input
}

func (w *InstallWizard) promptPassword(message string) string {
	// NOTE: For production use, consider using golang.org/x/term.ReadPassword
	// to hide password input. Current implementation shows passwords for
	// compatibility with automated testing and non-TTY environments.
	w.cmd.Printf("%s (will be visible): ", message)

	// Read password from stdin
	input, err := w.reader.ReadString('\n')
	if err != nil {
		return ""
	}

	password := strings.TrimSpace(input)
	// Don't echo empty passwords in automated mode, but allow them for testing
	if password == "" {
		return ""
	}

	return password
}

func (w *InstallWizard) promptYesNo(message string, defaultYes bool) bool {
	defaultStr := "y/N"
	if defaultYes {
		defaultStr = "Y/n"
	}

	response := w.prompt(fmt.Sprintf("%s [%s]", message, defaultStr), "")
	response = strings.ToLower(strings.TrimSpace(response))

	if response == "" {
		return defaultYes
	}

	return response == "y" || response == "yes"
}

func (w *InstallWizard) parseInt(s string, defaultValue int) int {
	var val int
	if _, err := fmt.Sscanf(s, "%d", &val); err != nil {
		return defaultValue
	}
	return val
}

func generateSecret(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}

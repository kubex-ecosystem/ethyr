// Package cli provides the daemon command for background service operations
package cli

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	// "github.com/kubex-ecosystem/kbx_be/internal/app/daemon"
	"github.com/spf13/cobra"
)

var (
	gobeURL             string
	gobeAPIKey          string
	autoScheduleEnabled bool
	scheduleCron        string
	notifyChannels      []string
	healthCheckInterval time.Duration
)

// NewDaemonCommand creates the daemon command
func NewDaemonCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "daemon",
		Short: "Start kubexGnyx as background daemon with GoBE integration",
		Long: `Start the kubexGnyx as a background daemon service that integrates with GoBE backend.

The daemon provides:
• Automatic repository analysis scheduling
• Integration with KubeX AI Squad system
• Discord/WhatsApp/Email notifications
• Health monitoring and reporting
• Meta-recursivity coordination with lookatni/gnyx

Examples:
  kubexGnyx daemon --gobe-url=http://localhost:3000 --gobe-api-key=abc123
  kubexGnyx daemon --auto-schedule --schedule-cron="0 2 * * *"
  kubexGnyx daemon --notify-channels=discord,email`,
		RunE: runDaemon,
	}

	// GoBE Integration flags
	cmd.Flags().StringVar(&gobeURL, "gobe-url",
		getEnvOrDefault("GOBE_URL", "http://localhost:3000"),
		"GoBE backend URL")
	cmd.Flags().StringVar(&gobeAPIKey, "gobe-api-key",
		os.Getenv("GOBE_API_KEY"),
		"GoBE API key for authentication")

	// Scheduling flags
	cmd.Flags().BoolVar(&autoScheduleEnabled, "auto-schedule", false,
		"Enable automatic repository analysis scheduling")
	cmd.Flags().StringVar(&scheduleCron, "schedule-cron", "0 2 * * *",
		"Cron expression for automatic scheduling (default: daily at 2 AM)")

	// Notification flags
	cmd.Flags().StringSliceVar(&notifyChannels, "notify-channels",
		[]string{"discord"},
		"Notification channels (discord,email,webhook)")

	// Health monitoring flags
	cmd.Flags().DurationVar(&healthCheckInterval, "health-interval",
		5*time.Minute,
		"Health check interval")

	return cmd
}

func runDaemon(cmd *cobra.Command, args []string) error {
	// Validate required flags
	if gobeAPIKey == "" {
		return fmt.Errorf("--gobe-api-key is required (or set GOBE_API_KEY env var)")
	}

	// Setup graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Wait for shutdown signal
	<-sigChan
	fmt.Fprintln(os.Stderr, "Received shutdown signal, stopping daemon...")

	return nil
}

func printDaemonInfo(config any /* daemon.DaemonConfig */) {
	_ = config
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

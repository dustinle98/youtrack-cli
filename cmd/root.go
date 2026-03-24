package cmd

import (
	"fmt"
	"os"

	"github.com/dustinle98/youtrack-cli/internal/api"
	"github.com/dustinle98/youtrack-cli/internal/config"
	"github.com/dustinle98/youtrack-cli/internal/output"
	"github.com/spf13/cobra"
)

// Version is set at build time via ldflags.
var Version = "dev"

// Global flags
var (
	jsonFlag   bool
	quietFlag  bool
	fieldsFlag string
)

// Shared instances (initialized in PersistentPreRun)
var (
	apiClient *api.Client
	cfg       *config.Config
	out       *output.Formatter
)

var rootCmd = &cobra.Command{
	Use:   "ytc",
	Short: "YouTrack CLI — fast command-line client for YouTrack",
	Long: `ytc is a fast CLI for interacting with JetBrains YouTrack.
Optimized for AI agents and human developers alike.

Quick start:
  ytc auth login              # Set up authentication
  ytc issue list -p DEMO      # List issues in a project
  ytc issue view DEMO-123     # View issue details
  ytc search "bug in login"   # Search issues`,
	Version: Version,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Always initialize formatter
		out = output.New(jsonFlag, quietFlag, fieldsFlag)

		// Skip auth check for auth commands
		if cmd.Name() == "login" || cmd.Name() == "status" || cmd.Name() == "help" || cmd.Name() == "ytc" {
			return nil
		}
		// Also skip for root command itself
		if cmd.Parent() == nil || cmd.Parent().Name() == "auth" {
			return nil
		}

		var err error
		cfg, err = config.Load()
		if err != nil {
			return fmt.Errorf("loading config: %w", err)
		}

		if err := cfg.Validate(); err != nil {
			return err
		}

		apiClient = api.NewClient(cfg.URL, cfg.Token, cfg.VerifySSL)
		return nil
	},
	SilenceUsage: true,
}

// Execute runs the root command.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().BoolVar(&jsonFlag, "json", false, "Output as JSON")
	rootCmd.PersistentFlags().BoolVarP(&quietFlag, "quiet", "q", false, "Minimal output (IDs only)")
	rootCmd.PersistentFlags().StringVar(&fieldsFlag, "fields", "", "Comma-separated list of fields to include in JSON output")
}

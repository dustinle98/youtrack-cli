package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/dustinle98/youtrack-cli/internal/api"
	"github.com/dustinle98/youtrack-cli/internal/config"
	"github.com/spf13/cobra"
)

var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "Manage authentication",
}

var authLoginCmd = &cobra.Command{
	Use:   "login",
	Short: "Log in to a YouTrack instance",
	Long:  `Interactively set up authentication by providing your YouTrack URL and API token.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		reader := bufio.NewReader(os.Stdin)

		// Load existing config for defaults
		existing, _ := config.Load()

		// Prompt for URL
		defaultURL := ""
		if existing != nil && existing.URL != "" {
			defaultURL = existing.URL
		}
		if defaultURL != "" {
			fmt.Printf("YouTrack URL [%s]: ", defaultURL)
		} else {
			fmt.Print("YouTrack URL (e.g. https://myteam.youtrack.cloud): ")
		}
		urlInput, _ := reader.ReadString('\n')
		urlInput = strings.TrimSpace(urlInput)
		if urlInput == "" {
			urlInput = defaultURL
		}
		if urlInput == "" {
			return fmt.Errorf("URL is required")
		}

		// Prompt for token
		fmt.Print("API Token (from YouTrack → Profile → Authentication): ")
		tokenInput, _ := reader.ReadString('\n')
		tokenInput = strings.TrimSpace(tokenInput)
		if tokenInput == "" {
			return fmt.Errorf("API token is required")
		}

		// Test connection
		fmt.Print("Testing connection... ")
		client := api.NewClient(urlInput, tokenInput, true)
		user, err := client.TestConnection()
		if err != nil {
			fmt.Println("✗")
			return fmt.Errorf("connection failed: %w", err)
		}
		fmt.Printf("✓ Logged in as %s\n", user)

		// Save config
		newCfg := &config.Config{
			URL:       urlInput,
			Token:     tokenInput,
			VerifySSL: true,
		}
		if existing != nil && existing.DefaultProject != "" {
			newCfg.DefaultProject = existing.DefaultProject
		}
		if err := config.Save(newCfg); err != nil {
			return fmt.Errorf("saving config: %w", err)
		}
		fmt.Printf("Config saved to %s\n", config.ConfigFilePath())
		return nil
	},
}

var authStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show current authentication status",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return fmt.Errorf("loading config: %w", err)
		}

		if cfg.URL == "" || cfg.Token == "" {
			fmt.Println("Not authenticated. Run 'ytc auth login' to set up.")
			return nil
		}

		fmt.Printf("URL:   %s\n", cfg.URL)
		fmt.Printf("Token: %s...%s\n", cfg.Token[:4], cfg.Token[len(cfg.Token)-4:])

		// Test connection
		fmt.Print("Status: ")
		client := api.NewClient(cfg.URL, cfg.Token, cfg.VerifySSL)
		user, err := client.TestConnection()
		if err != nil {
			fmt.Printf("✗ %s\n", err)
			return nil
		}
		fmt.Printf("✓ Connected as %s\n", user)
		return nil
	},
}

func init() {
	authCmd.AddCommand(authLoginCmd)
	authCmd.AddCommand(authStatusCmd)
	rootCmd.AddCommand(authCmd)
}

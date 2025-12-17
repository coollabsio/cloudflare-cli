package cmd

import (
	"context"
	"fmt"

	"github.com/heyandras/cfdns/internal/client"
	"github.com/heyandras/cfdns/internal/config"
	"github.com/spf13/cobra"
)

var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "Authentication commands",
}

var authVerifyCmd = &cobra.Command{
	Use:   "verify",
	Short: "Verify API credentials",
	Long:  `Verify that the configured API credentials are valid and can access the Cloudflare API.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if !cfg.HasCredentials() {
			return fmt.Errorf(`no credentials configured

Set one of the following:
  Environment variable: CLOUDFLARE_API_TOKEN
  or
  Environment variables: CLOUDFLARE_API_KEY + CLOUDFLARE_API_EMAIL
  or
  Config file at ~/.cloudflare/config.yaml with:
    api_token: your-token-here`)
		}

		c, err := client.New(cfg)
		if err != nil {
			return err
		}

		ctx := context.Background()
		if err := c.VerifyToken(ctx); err != nil {
			return err
		}

		out.WriteSuccess(fmt.Sprintf("Authentication successful (using %s)", cfg.AuthMethod()))
		return nil
	},
}

var authSaveCmd = &cobra.Command{
	Use:   "save <token>",
	Short: "Save API token to config file",
	Long: `Save an API token to the config file (~/.cloudflare/config.yaml).

Example:
  cfdns auth save YOUR_API_TOKEN`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		token := args[0]

		// Create a new config with the token
		newCfg := &config.Config{
			APIToken: token,
		}

		// Optionally verify the token first
		c, err := client.New(newCfg)
		if err != nil {
			return err
		}

		ctx := context.Background()
		if err := c.VerifyToken(ctx); err != nil {
			return fmt.Errorf("token verification failed: %w", err)
		}

		// Save to config file
		configPath := cfgFile
		if configPath == "" {
			configPath = config.DefaultConfigPath()
		}

		if err := newCfg.Save(configPath); err != nil {
			return fmt.Errorf("failed to save config: %w", err)
		}

		out.WriteSuccess(fmt.Sprintf("Token saved to %s", configPath))
		return nil
	},
}

func init() {
	rootCmd.AddCommand(authCmd)
	authCmd.AddCommand(authVerifyCmd)
	authCmd.AddCommand(authSaveCmd)
}

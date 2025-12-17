package cmd

import (
	"os"

	"github.com/heyandras/cf/internal/config"
	"github.com/heyandras/cf/internal/output"
	"github.com/spf13/cobra"
)

var (
	cfgFile      string
	outputFormat string
	cfg          *config.Config
	out          *output.Writer
)

// rootCmd represents the base command
var rootCmd = &cobra.Command{
	Use:   "cf",
	Short: "Cloudflare DNS CLI",
	Long: `A command-line tool for managing Cloudflare DNS records.

Configure authentication using environment variables:
  CLOUDFLARE_API_TOKEN (recommended)
  or
  CLOUDFLARE_API_KEY + CLOUDFLARE_API_EMAIL

Or create a config file at ~/.cloudflare/config.yaml:
  api_token: your-token-here`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		var err error
		cfg, err = config.Load(cfgFile)
		if err != nil {
			return err
		}

		// Determine output format: flag > config > default
		format := output.FormatTable
		if cfg.OutputFormat == "json" {
			format = output.FormatJSON
		}
		// Command-line flag overrides config
		if cmd.Flags().Changed("output") {
			if outputFormat == "json" {
				format = output.FormatJSON
			} else {
				format = output.FormatTable
			}
		}
		out = output.NewWriter(format)
		return nil
	},
}

// Execute runs the root command
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is ~/.cloudflare/config.yaml)")
	rootCmd.PersistentFlags().StringVarP(&outputFormat, "output", "o", "table", "output format (table, json)")
}

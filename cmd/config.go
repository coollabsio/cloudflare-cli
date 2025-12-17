package cmd

import (
	"fmt"

	"github.com/heyandras/cf/internal/config"
	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Configuration commands",
}

var configSetCmd = &cobra.Command{
	Use:   "set <key> <value>",
	Short: "Set a config value",
	Long: `Set a configuration value.

Available keys:
  output_format  - Default output format (table, json)

Examples:
  cf config set output_format json
  cf config set output_format table`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		key := args[0]
		value := args[1]

		// Load existing config
		configPath := cfgFile
		if configPath == "" {
			configPath = config.DefaultConfigPath()
		}

		existingCfg, _ := config.Load(configPath)
		if existingCfg == nil {
			existingCfg = &config.Config{}
		}

		switch key {
		case "output_format":
			if value != "table" && value != "json" {
				return fmt.Errorf("invalid output_format: %s (must be 'table' or 'json')", value)
			}
			existingCfg.OutputFormat = value
		default:
			return fmt.Errorf("unknown config key: %s", key)
		}

		if err := existingCfg.Save(configPath); err != nil {
			return fmt.Errorf("failed to save config: %w", err)
		}

		out.WriteSuccess(fmt.Sprintf("Set %s = %s", key, value))
		return nil
	},
}

var configGetCmd = &cobra.Command{
	Use:   "get <key>",
	Short: "Get a config value",
	Long: `Get a configuration value.

Available keys:
  output_format  - Default output format

Examples:
  cf config get output_format`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		key := args[0]

		switch key {
		case "output_format":
			value := cfg.OutputFormat
			if value == "" {
				value = "table"
			}
			fmt.Println(value)
		default:
			return fmt.Errorf("unknown config key: %s", key)
		}

		return nil
	},
}

var configListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all config values",
	RunE: func(cmd *cobra.Command, args []string) error {
		outputFormat := cfg.OutputFormat
		if outputFormat == "" {
			outputFormat = "table (default)"
		}

		headers := []string{"Key", "Value"}
		rows := [][]string{
			{"output_format", outputFormat},
		}
		return out.WriteTable(headers, rows)
	},
}

func init() {
	rootCmd.AddCommand(configCmd)
	configCmd.AddCommand(configSetCmd)
	configCmd.AddCommand(configGetCmd)
	configCmd.AddCommand(configListCmd)
}

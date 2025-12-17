package cmd

import (
	"fmt"

	"github.com/coollabsio/cf/internal/version"
	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number",
	Long:  `Display the current version of cf.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("cf version %s\n", version.GetVersion())
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}

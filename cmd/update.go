package cmd

import (
	"context"
	"fmt"

	"github.com/coollabsio/cf/internal/version"
	"github.com/creativeprojects/go-selfupdate"
	goversion "github.com/hashicorp/go-version"
	"github.com/spf13/cobra"
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update cf to the latest version",
	Long:  `Check for and download the latest version of cf from GitHub releases.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		currentVersion := version.GetVersion()
		fmt.Printf("Current version: %s\n", currentVersion)
		fmt.Println("Checking for updates...")

		ctx := context.Background()
		latest, found, err := selfupdate.DetectLatest(ctx, selfupdate.ParseSlug("coollabsio/cf"))
		if err != nil {
			return fmt.Errorf("failed to detect latest version: %w", err)
		}
		if !found {
			return fmt.Errorf("no release found for this platform")
		}

		// Compare versions (skip for dev builds)
		if currentVersion != "dev" {
			current, err := goversion.NewVersion(currentVersion)
			if err != nil {
				return fmt.Errorf("failed to parse current version: %w", err)
			}

			latestVersion, err := goversion.NewVersion(latest.Version())
			if err != nil {
				return fmt.Errorf("failed to parse latest version: %w", err)
			}

			if !latestVersion.GreaterThan(current) {
				fmt.Printf("You are already on the latest version (%s)\n", currentVersion)
				return nil
			}
		}

		fmt.Printf("Updating to version %s...\n", latest.Version())

		exe, err := selfupdate.ExecutablePath()
		if err != nil {
			return fmt.Errorf("failed to get executable path: %w", err)
		}

		if err := selfupdate.UpdateTo(ctx, latest.AssetURL, latest.AssetName, exe); err != nil {
			return fmt.Errorf("failed to update: %w", err)
		}

		fmt.Printf("Successfully updated to version %s\n", latest.Version())
		return nil
	},
}

func init() {
	rootCmd.AddCommand(updateCmd)
}

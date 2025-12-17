package version

import (
	"context"
	"fmt"
	"time"

	"github.com/creativeprojects/go-selfupdate"
	goversion "github.com/hashicorp/go-version"
)

// Version is set at build time via ldflags
var Version = "dev"

// updateMessage holds the result of an async version check
var updateMessage chan string

// GetVersion returns the current version
func GetVersion() string {
	return Version
}

// StartUpdateCheck begins an async check for newer versions.
// Call PrintUpdateMessage after command execution to display any update notification.
func StartUpdateCheck() {
	updateMessage = make(chan string, 1)

	// Skip check for dev builds
	if Version == "dev" {
		close(updateMessage)
		return
	}

	go func() {
		defer close(updateMessage)

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		latest, found, err := selfupdate.DetectLatest(ctx, selfupdate.ParseSlug("coollabsio/cf"))
		if err != nil || !found {
			return
		}

		currentVersion, err := goversion.NewVersion(Version)
		if err != nil {
			return
		}

		latestVersion, err := goversion.NewVersion(latest.Version())
		if err != nil {
			return
		}

		if latestVersion.GreaterThan(currentVersion) {
			updateMessage <- fmt.Sprintf("\nA new version (%s) is available. Update with: cf update\n", latest.Version())
		}
	}()
}

// PrintUpdateMessage waits for the async version check to complete and prints any update notification.
// This should be called after the command has finished executing.
func PrintUpdateMessage() {
	if updateMessage == nil {
		return
	}

	// Wait for the check to complete (with the 5s timeout from the goroutine)
	if msg, ok := <-updateMessage; ok && msg != "" {
		fmt.Print(msg)
	}
}

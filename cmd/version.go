package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var (
	Version = "development" //nolint: gochecknoglobals
	Commit  = "0.0.0"       //nolint: gochecknoglobals
)

func versionCommand() *cobra.Command {
	versionCmd := cobra.Command{
		Use:   "version",
		Short: "Displays the build version",
		Run:   versionHandler,
	}

	return &versionCmd
}

func versionHandler(_ *cobra.Command, _ []string) {
	logs.Info().Msgf("%s Version: %s Commit: %s", os.Args[0], Version, Commit)
}

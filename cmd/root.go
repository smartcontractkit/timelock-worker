package cmd

import (
	"os"

	"github.com/rs/zerolog"
	"github.com/smartcontractkit/timelock-worker/pkg/logger"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	rootCmd = &cobra.Command{
		Use:   "timelock-worker",
		Short: "Pull and execute scheduled transactions from Timelock contract",
	}

	logs             *zerolog.Logger
	logLevel, output string
)

func Execute() {
	if err := configureRootCmd(); err != nil {
		os.Exit(1)
	}

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func configureRootCmd() error {
	cobra.OnInitialize(initConfig)

	rootCmd.AddCommand(
		versionCommand(),
		startCommand(),
	)

	// Global flags
	rootCmd.PersistentFlags().StringVarP(&logLevel, "log-level", "l", "info", "set logging level (trace, debug, info, warn, error, fatal)")
	if err := viper.BindPFlag("log-level", rootCmd.PersistentFlags().Lookup("log-level")); err != nil {
		return err
	}

	rootCmd.PersistentFlags().StringVarP(&output, "output", "o", "human", "set logging output (human, json)")
	if err := viper.BindPFlag("output", rootCmd.PersistentFlags().Lookup("output")); err != nil {
		return err
	}

	return nil
}

func initConfig() {
	// Output hardcoded to JSON
	logs = logger.Logger(viper.GetString("log-level"), viper.GetString("output"))
	logs.Debug().Msgf("initialized Logger")
}

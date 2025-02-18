package cmd

import (
	"log"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap"

	"github.com/smartcontractkit/timelock-worker/pkg/logger"
)

var (
	rootCmd = &cobra.Command{
		Use:   "timelock-worker",
		Short: "Pull and execute scheduled transactions from Timelock contract",
	}
	logs     *zap.Logger
	logLevel string
	output   string
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
	rootCmd.PersistentFlags().StringVarP(&logLevel, "log-level", "l", "info", "set logging level (debug, info)")
	if err := viper.BindPFlag("log-level", rootCmd.PersistentFlags().Lookup("log-level")); err != nil {
		return err
	}
	if err := viper.BindEnv("log-level", "LOGLEVEL"); err != nil {
		return err
	}

	rootCmd.PersistentFlags().StringVarP(&output, "output", "o", "human", "set logging output (human, json)")
	if err := viper.BindPFlag("output", rootCmd.PersistentFlags().Lookup("output")); err != nil {
		return err
	}
	if err := viper.BindEnv("output", "OUTPUT"); err != nil {
		return err
	}

	return nil
}

func initConfig() {
	var err error
	logs, err = logger.NewLogger(viper.GetString("log-level"), viper.GetString("output"))
	if err != nil {
		log.Fatalf("unable to create logger: %s", err)
	}

	logs.Debug("initialized Logger")
}

package cmd

import (
	"context"
	"math/big"

	"github.com/smartcontractkit/timelock-worker/pkg/cli"
	"github.com/smartcontractkit/timelock-worker/pkg/timelock"
	"github.com/spf13/cobra"
)

func startCommand() *cobra.Command {
	var (
		startCmd = cobra.Command{
			Use:   "start",
			Short: "Starts the Timelock Worker daemon",
			Run:   startHandler,
		}

		nodeURL, privateKey, timelockAddress, callProxyAddress string
		fromBlock, pollPeriod                                  int64
	)

	// Initialize timelock-worker configuration.
	// Precedence: flags > env variables > timelock.env file.
	timelockConf, err := cli.NewTimelockCLI()
	if err != nil {
		logs.Fatal().Msgf("error initializing configuration: %s", err.Error())
	}
	// Set healthStatus to error on startup.
	// Will be set to "ok" once the rpc connection and subscription is successful.
	timelock.SetHealthStatus(timelock.HealthStatusError)

	startCmd.Flags().StringVarP(&nodeURL, "node-url", "n", timelockConf.NodeURL, "RPC Endpoint for the target blockchain")
	startCmd.Flags().StringVarP(&timelockAddress, "timelock-address", "a", timelockConf.TimelockAddress, "Address of the target Timelock contract")
	startCmd.Flags().StringVarP(&callProxyAddress, "call-proxy-address", "f", timelockConf.CallProxyAddress, "Address of the target CallProxyAddress contract")
	startCmd.Flags().StringVarP(&privateKey, "private-key", "k", timelockConf.PrivateKey, "Private key used to execute transactions")
	startCmd.Flags().Int64Var(&fromBlock, "from-block", timelockConf.FromBlock, "Start watching from this block")
	startCmd.Flags().Int64Var(&pollPeriod, "poll-period", timelockConf.PollPeriod, "Poll period in seconds")

	return &startCmd
}

func startHandler(cmd *cobra.Command, _ []string) {
	// Use this ctx as the base context.
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go startHTTPHealthServer()
	startTimelock(ctx, cmd)
}

func startTimelock(ctx context.Context, cmd *cobra.Command) {

	nodeURL, err := cmd.Flags().GetString("node-url")
	if err != nil {
		logs.Fatal().Msgf("value of node-url not set: %s", err.Error())
	}

	timelockAddress, err := cmd.Flags().GetString("timelock-address")
	if err != nil {
		logs.Fatal().Msgf("value of timelock-address not set: %s", err.Error())
	}

	callProxyAddress, err := cmd.Flags().GetString("call-proxy-address")
	if err != nil {
		logs.Fatal().Msgf("value of call-proxy-address not set: %s", err.Error())
	}

	privateKey, err := cmd.Flags().GetString("private-key")
	if err != nil {
		logs.Fatal().Msgf("value of private-key not set: %s", err.Error())
	}

	fromBlock, err := cmd.Flags().GetInt64("from-block")
	if err != nil {
		logs.Fatal().Msgf("value of from-block not set: %s", err.Error())
	}

	pollPeriod, err := cmd.Flags().GetInt64("poll-period")
	if err != nil {
		logs.Fatal().Msgf("value of poll-period not set: %s", err.Error())
	}

	tWorker, err := timelock.NewTimelockWorker(nodeURL, timelockAddress, callProxyAddress, privateKey, big.NewInt(fromBlock), pollPeriod, logs)
	if err != nil {
		logs.Fatal().Msgf("error creating the timelock-worker: %s", err.Error())
	}

	if err := tWorker.Listen(ctx); err != nil {
		logs.Fatal().Msgf("error while starting timelock-worker: %s", err.Error())
	}

	logs.Info().Msg("shutting down timelock-worker")
}

func startHTTPHealthServer() {
	timelock.StartHTTPHealthServer()
}

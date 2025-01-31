package cmd

import (
	"context"
	"math/big"

	"github.com/spf13/cobra"

	"github.com/smartcontractkit/timelock-worker/pkg/cli"
	"github.com/smartcontractkit/timelock-worker/pkg/timelock"
)

func startCommand() *cobra.Command {
	var (
		startCmd = cobra.Command{
			Use:   "start",
			Short: "Starts the Timelock Worker daemon",
			Run:   startHandler,
		}

		nodeURL, privateKey, timelockAddress, callProxyAddress string
		fromBlock, pollPeriod, eventListenerPollPeriod         int64
		eventListenerPollSize                                  uint64
		dryRun                                                 bool
	)

	// Initialize timelock-worker configuration.
	// Precedence: flags > env variables > timelock.env file.
	timelockConf, err := cli.NewTimelockCLI()
	if err != nil {
		logs.Sugar().Fatalf("error initializing configuration: %s", err.Error())
	}
	// Set liveStatus to OK on startup.
	// Set readyStatus to Error on startup.
	// Will be set to OK once the rpc connection and subscription is successful.
	timelock.SetLiveStatus(timelock.HealthStatusOK)
	timelock.SetReadyStatus(timelock.HealthStatusError)

	startCmd.Flags().StringVarP(&nodeURL, "node-url", "n", timelockConf.NodeURL, "RPC Endpoint for the target blockchain")
	startCmd.Flags().StringVarP(&timelockAddress, "timelock-address", "a", timelockConf.TimelockAddress, "Address of the target Timelock contract")
	startCmd.Flags().StringVarP(&callProxyAddress, "call-proxy-address", "f", timelockConf.CallProxyAddress, "Address of the target CallProxyAddress contract")
	startCmd.Flags().StringVarP(&privateKey, "private-key", "k", timelockConf.PrivateKey, "Private key used to execute transactions")
	startCmd.Flags().Int64Var(&fromBlock, "from-block", timelockConf.FromBlock, "Start watching from this block")
	startCmd.Flags().Int64Var(&pollPeriod, "poll-period", timelockConf.PollPeriod, "Poll period in seconds")
	startCmd.Flags().Int64Var(&eventListenerPollPeriod, "event-listener-poll-period", timelockConf.EventListenerPollPeriod, "Event Listener poll period in seconds")
	startCmd.Flags().Uint64Var(&eventListenerPollSize, "event-listener-poll-size", timelockConf.EventListenerPollSize, "Number of entries to fetch when polling logs")
	startCmd.Flags().BoolVar(&dryRun, "dry-run", timelockConf.DryRun, "Enable \"dry run\" mode -- monitor events but don't trigger any calls")

	return &startCmd
}

func startHandler(cmd *cobra.Command, _ []string) {
	go startHTTPHealthServer()
	go startMetricsServer()
	startTimelock(cmd)
}

func startTimelock(cmd *cobra.Command) {
	slog := logs.Sugar()

	nodeURL, err := cmd.Flags().GetString("node-url")
	if err != nil {
		slog.Fatalf("value of node-url not set: %s", err.Error())
	}

	timelockAddress, err := cmd.Flags().GetString("timelock-address")
	if err != nil {
		slog.Fatalf("value of timelock-address not set: %s", err.Error())
	}

	callProxyAddress, err := cmd.Flags().GetString("call-proxy-address")
	if err != nil {
		slog.Fatalf("value of call-proxy-address not set: %s", err.Error())
	}

	privateKey, err := cmd.Flags().GetString("private-key")
	if err != nil {
		slog.Fatalf("value of private-key not set: %s", err.Error())
	}

	fromBlock, err := cmd.Flags().GetInt64("from-block")
	if err != nil {
		slog.Fatalf("value of from-block not set: %s", err.Error())
	}

	pollPeriod, err := cmd.Flags().GetInt64("poll-period")
	if err != nil {
		slog.Fatalf("value of poll-period not set: %s", err.Error())
	}

	eventListenerPollPeriod, err := cmd.Flags().GetInt64("event-listener-poll-period")
	if err != nil {
		slog.Fatalf("value of poll-period not set: %s", err.Error())
	}

	eventListenerPollSize, err := cmd.Flags().GetUint64("event-listener-poll-size")
	if err != nil {
		slog.Fatalf("value of event-listener-poll-size not set: %s", err.Error())
	}

	dryRun, err := cmd.Flags().GetBool("dry-run")
	if err != nil {
		slog.Fatalf("value of dry-run not set: %s", err.Error())
	}

	tWorker, err := timelock.NewTimelockWorker(nodeURL, timelockAddress, callProxyAddress, privateKey,
		big.NewInt(fromBlock), pollPeriod, eventListenerPollPeriod, eventListenerPollSize, dryRun, slog)
	if err != nil {
		slog.Fatalf("error creating the timelock-worker: %s", err.Error())
	}

	if err := tWorker.Listen(context.Background()); err != nil {
		slog.Fatalf("error while starting timelock-worker: %s", err.Error())
	}

	slog.Infof("shutting down timelock-worker")
}

func startHTTPHealthServer() {
	timelock.StartHTTPHealthServer(logs.Sugar())
}

func startMetricsServer() {
	timelock.StartMetricsServer(logs.Sugar())
}

package cmd

import (
	"context"
	"fmt"
	"math/big"
	"net/http"
	"time"
	"sync/atomic"

	"github.com/smartcontractkit/timelock-worker/pkg/cli"
	"github.com/smartcontractkit/timelock-worker/pkg/timelock"
	"github.com/spf13/cobra"
)

var healthStatus atomic.Value

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
	healthStatus.Store("Error")

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

	if err := tWorker.Listen(ctx, &healthStatus); err != nil {
		logs.Fatal().Msgf("error while starting timelock-worker: %s", err.Error())
	}

	logs.Info().Msg("shutting down timelock-worker")

}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	status := healthStatus.Load().(string)
	if status == "OK" {
		w.Write([]byte("OK"))
 } else {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Error"))
 }
}

// starts a http server, serving the healthz endpoint.
func startHTTPHealthServer() {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", healthHandler)

	server := &http.Server{
		Addr:         ":8080",
		Handler:      mux,
		ReadTimeout:  5 * time.Second,  // Set your desired read timeout
		WriteTimeout: 10 * time.Second, // Set your desired write timeout
		IdleTimeout:  15 * time.Second, // Set your desired idle timeout
	}

	fmt.Println("Server listening on :8080")
	if err := server.ListenAndServe(); err != nil {
		fmt.Println(err)
	}
}

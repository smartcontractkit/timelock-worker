package cmd

import (
	"fmt"
	"os"
	"net/http"
	"github.com/spf13/cobra"
)

func serverCommand() *cobra.Command {
	var (
		serverCmd = cobra.Command{
			Use:   "healthcheck",
			Short: "Serves the Timelock Worker healthcheck endpoint",
			Run:   startServer,
		}

		// nodeURL, privateKey, timelockAddress, callProxyAddress string
		// fromBlock, pollPeriod                                  int64
	)

	// Initialize timelock-worker configuration.
	// Precedence: flags > env variables > timelock.env file.
	// timelockConf, err := cli.NewTimelockCLI()
	// if err != nil {
	// 	logs.Fatal().Msgf("error initializing configuration: %s", err.Error())
	// }

	// startCmd.Flags().StringVarP(&nodeURL, "node-url", "n", timelockConf.NodeURL, "RPC Endpoint for the target blockchain")
	// startCmd.Flags().StringVarP(&timelockAddress, "timelock-address", "a", timelockConf.TimelockAddress, "Address of the target Timelock contract")
	// startCmd.Flags().StringVarP(&callProxyAddress, "call-proxy-address", "f", timelockConf.CallProxyAddress, "Address of the target CallProxyAddress contract")
	// startCmd.Flags().StringVarP(&privateKey, "private-key", "k", timelockConf.PrivateKey, "Private key used to execute transactions")
	// startCmd.Flags().Int64Var(&fromBlock, "from-block", timelockConf.FromBlock, "Start watching from this block")
	// startCmd.Flags().Int64Var(&pollPeriod, "poll-period", timelockConf.PollPeriod, "Poll period in seconds")

	return &serverCmd
}

func startServer(cmd *cobra.Command, _ []string) {
	http.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "OK")
	})

	port := 8080
	addr := fmt.Sprintf(":%d", port)

	fmt.Printf("Starting HTTP server on port %d...\n", port)
	if err := http.ListenAndServe(addr, nil); err != nil {
		fmt.Println("Error starting HTTP server:", err)
		os.Exit(1)
	}
}

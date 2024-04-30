package timelock

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"math/big"
	"net/url"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/rs/zerolog"
	"github.com/smartcontractkit/timelock-worker/pkg/timelock/contract"
)

var (
	wg      sync.WaitGroup
	tWorker *Worker
)

// Worker represents a worker instance.
// address is an array of addresses as expected by ethereum.FilterQuery,
// but it's enforced only to one address in the logic.
type Worker struct {
	ethClient       *ethclient.Client
	contract        *contract.Timelock
	executeContract *contract.Timelock
	abi             *abi.ABI
	address         []common.Address
	fromBlock       *big.Int
	pollPeriod      int64
	logger          *zerolog.Logger
	privateKey      *ecdsa.PrivateKey
	scheduler
}

// NewTimelockWorker initializes and returns a timelockWorker.
// It's a singleton, so further executions will retrieve the same timelockWorker.
func NewTimelockWorker(nodeURL, timelockAddress, callProxyAddress, privateKey string, fromBlock *big.Int, pollPeriod int64, logger *zerolog.Logger) (*Worker, error) {
	// Sanity check on each provided variable before allocating more resources.
	u, err := url.ParseRequestURI(nodeURL)
	if err != nil {
		return nil, err
	}

	if u.Scheme == "http" || u.Scheme == "https" {
		return nil, fmt.Errorf("only ws or wss are valid options to suscribe to events: nodeURL using %s", u.Scheme)
	}

	if !common.IsHexAddress(timelockAddress) {
		return nil, fmt.Errorf("timelock address provided is not valid: %s", timelockAddress)
	}

	if !common.IsHexAddress(callProxyAddress) {
		return nil, fmt.Errorf("call proxy address provided is not valid: %s", callProxyAddress)
	}

	if pollPeriod <= 0 {
		return nil, fmt.Errorf("poll-period must be a positive non-zero integer: got %d", pollPeriod)
	}

	if fromBlock.Int64() < big.NewInt(0).Int64() {
		return nil, fmt.Errorf("from block can't be a negative number (minimum value 0): got %d", fromBlock.Int64())
	}

	if _, err := crypto.HexToECDSA(privateKey); err != nil {
		return nil, fmt.Errorf("the provided private key is not valid: got %s", privateKey)
	}

	// All variables provided are correct, start allocating new structures.
	client, err := rpc.Dial(nodeURL)
	if err != nil {
		return nil, err
	}

	ethClient := ethclient.NewClient(client)

	timelockABI, err := contract.TimelockMetaData.GetAbi()
	if err != nil {
		return nil, err
	}

	// The contract ABI give grants capabilities such as parsing events and accessing to fields.
	// As NewTimelock only accepts one contract, hardcode it to address[0].
	timelockContract, err := contract.NewTimelock(common.HexToAddress(timelockAddress), ethClient)
	if err != nil {
		return nil, err
	}

	// The execute contract is the call proxy contract, which is the one that actually executes the transaction.
	// It's not the same as the timelock contract, so it has to be initialized separately.
	executeContract, err := contract.NewTimelock(common.HexToAddress(callProxyAddress), ethClient)
	if err != nil {
		return nil, err
	}

	privateKeyECDSA, err := crypto.HexToECDSA(privateKey)
	if err != nil {
		return nil, err
	}

	tWorker = &Worker{
		ethClient:       ethClient,
		contract:        timelockContract,
		executeContract: executeContract,
		abi:             timelockABI,
		address:         []common.Address{common.HexToAddress(timelockAddress)},
		fromBlock:       fromBlock,
		pollPeriod:      pollPeriod,
		logger:          logger,
		privateKey:      privateKeyECDSA,
		scheduler:       *newScheduler(time.Duration(pollPeriod) * time.Second),
	}

	return tWorker, nil
}

// Listen is the main function of a Timelock Worker, it subscribes to events using the ethClient
// and targeting the contract address set.
func (tw *Worker) Listen(ctx context.Context) error {
	c, cancel := context.WithCancel(ctx)
	defer cancel()

	// Log timelock-worker configuration.
	tw.startLog()

	// Handle OS signals to properly stop/kill the process.
	stopCh := make(chan string)
	defer close(stopCh)
	go handleOSSignal(stopCh)

	// Run the scheduler to add/del operations in a thread-safe way.
	go tw.runScheduler(c)

	// Retrieve historical logs.
	history, err := tw.retrieveHistoricalLogs(c)
	if err != nil {
		tw.logger.Error().Err(err).Msg("failed to subscribe and process logs.")
		return err
	}

	// Create a subscription and retrieve new logs asynchronously.
	logCh := make(chan types.Log)
	defer close(logCh)

	if err := tw.subscribeNewLogs(c, logCh); err != nil {
		tw.logger.Error().Err(err).Msg("failed to subscribe and process logs.")
		return err
	}

	// Main goroutine; processes old and new logs and handles cancellation.
	tw.processLogs(c, history, logCh, stopCh, cancel)

	// Block until all goroutines are done.
	wg.Wait()

	tw.dumpOperationStore(time.Now)

	return nil
}

func (tw *Worker) setupFilterQuery() ethereum.FilterQuery {
	return ethereum.FilterQuery{
		Addresses: tw.address,
		FromBlock: tw.fromBlock,
	}
}

// subscribeAndProcessLogs creates a subscription and processes logs based on a filter query.
func (tw *Worker) subscribeNewLogs(ctx context.Context, logCh chan types.Log) error {
	query := tw.setupFilterQuery()

	// SubscribeFilterLogs creates an asynchronous subscription to the events.
	// It receives all the new events.
	sub, err := tw.ethClient.SubscribeFilterLogs(ctx, query, logCh)
	if err != nil {
		tw.logger.Error().Msgf("unexpected error while creating subscription: %s", err.Error())
		return err
	}
	defer sub.Unsubscribe()

	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case err := <-sub.Err():
				// Check if the error is not nil, because sub.Unsubscribe will
				// signal the channel sub.Err() to close it, leading to false nil errors.
				if err != nil {
					tw.logger.Warn().Msgf("subscription error: %s", err.Error())
					SetReadyStatus(HealthStatusError)
					sub.Unsubscribe()

					success := false
					for try := range maxSubRetries {
						tw.logger.Warn().Msgf("trying to re-create subscription: %v/%v retry.", try+1, maxSubRetries)
						sub, err = tw.ethClient.SubscribeFilterLogs(ctx, query, logCh)
						if err == nil {
							tw.logger.Info().Msg("subscription successfully recreated.")
							SetReadyStatus(HealthStatusOK)
							success = true
							break
						}

						time.Sleep(time.Second * time.Duration(try))
					}

					if !success {
						tw.logger.Error().Msg("failed to recreate subscription after retries: shutting down timelock-worker.")
						wg.Done()
						return
					}
				}

			case <-ctx.Done():
				tw.logger.Debug().Msgf("shutting down subscription")
				SetReadyStatus(HealthStatusError)
				return
			}
		}
	}()

	return nil
}

func (tw *Worker) retrieveHistoricalLogs(ctx context.Context) (chan types.Log, error) {
	query := tw.setupFilterQuery()
	logCh := make(chan types.Log)

	// FilterLogs starts filtering the logs at fromBlock, gathering the historical data.
	// This is needed to guarantee that all the events are gathered, even in scenarios where the worker crashes.
	filter, err := tw.ethClient.FilterLogs(ctx, query)
	if err != nil {
		return nil, err
	}

	// Process incoming historical logs in a separate goroutine.
	wg.Add(1)
	go func() {
		defer wg.Done()
		for _, log := range filter {
			select {
			case logCh <- log:
				tw.logger.Debug().Msgf("processing historical log: %v\n", log)
			case <-ctx.Done():
				tw.logger.Debug().Msg("shutting down the retrieval of old logs in incomplete status")
				return
			}
		}
		tw.logger.Debug().Msg("retrieved correctly all historical events")
		close(logCh)
	}()

	return logCh, nil
}

func (tw *Worker) processLogs(ctx context.Context, oldLog, newLog chan types.Log, stopCh chan string, cancel context.CancelFunc) {
	// This is the goroutine watching over the subscribed and historical logs.
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case log := <-newLog:
				if err := tw.handleLog(ctx, log); err != nil {
					tw.logger.Error().Msgf("error processing new log: %v\n", log)
				}

			case log := <-oldLog:
				if err := tw.handleLog(ctx, log); err != nil {
					tw.logger.Error().Msgf("error processing historical log: %v\n", log)
				}

			case signal := <-stopCh:
				tw.logger.Info().Msgf("received OS signal %s", signal)
				SetReadyStatus(HealthStatusError)
				cancel()
				return
			}
		}
	}()
}

func (tw *Worker) handleLog(ctx context.Context, log types.Log) error {
	// Ignore logs with no topics.
	if len(log.Topics) == 0 {
		return nil
	}

	// Decode the log into an event using the ABI exposed in Timelock.go
	event, err := tw.abi.EventByID(log.Topics[0])
	if err != nil {
		return err
	}

	if event == nil {
		return fmt.Errorf("event is null")
	}

	switch event.Name {
	// A CallScheduled event should be added to an scheduler only if it's not already done
	// and it's a valid Operation.
	case eventCallScheduled:
		cs, err := tw.contract.ParseCallScheduled(log)
		if err != nil {
			return err
		}

		if !isDone(ctx, tw.contract, cs.Id) && isOperation(ctx, tw.contract, cs.Id) {
			tw.logger.Info().Hex(fieldTXHash, cs.Raw.TxHash[:]).Uint64(fieldBlockNumber, cs.Raw.BlockNumber).Msgf("%s received", eventCallScheduled)
			tw.addToScheduler(cs)
		}

		// A CallExecuted which is in Done status should delete the task in the scheduler store.
	case eventCallExecuted:
		cs, err := tw.contract.ParseCallExecuted(log)
		if err != nil {
			return err
		}

		if isDone(ctx, tw.contract, cs.Id) {
			tw.logger.Info().Hex(fieldTXHash, cs.Raw.TxHash[:]).Uint64(fieldBlockNumber, cs.Raw.BlockNumber).Msgf("%s received, skipping operation", eventCallExecuted)
			tw.delFromScheduler(cs.Id)
		}

		// A Cancelled which is in Done status should delete the task in the scheduler store.
	case eventCancelled:
		cs, err := tw.contract.ParseCancelled(log)
		if err != nil {
			return err
		}

		if isDone(ctx, tw.contract, cs.Id) {
			tw.logger.Info().Hex(fieldTXHash, cs.Raw.TxHash[:]).Uint64(fieldBlockNumber, cs.Raw.BlockNumber).Msgf("%s received, cancelling operation", eventCancelled)
			tw.delFromScheduler(cs.Id)
		}
	}

	return nil
}

func (tw *Worker) startLog() {
	tw.logger.Info().Msgf("timelock-worker started")
	tw.logger.Info().Msgf("\tTimelock contract address: %v", tw.address[0])

	wallet, err := privateKeyToAddress(tw.privateKey)
	if err != nil {
		tw.logger.Fatal().Msgf("\tEOA address: unable to determine")
	}

	tw.logger.Info().Msgf("\tEOA address: %v", wallet)
	tw.logger.Info().Msgf("\tStarting from block: %v", tw.fromBlock)
	tw.logger.Info().Msgf("\tPoll Period: %v", time.Duration(tw.pollPeriod*int64(time.Second)).String())
}

// handleOSSignal handles SIGINT and SIGTERM OS signals, and signals the stopCh.
func handleOSSignal(stopCh chan string) {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	defer close(sigCh)

	// Block and wait until a system signal happens.
	signal := <-sigCh

	// In the future SIGHUP can be used to reload configuration.
	switch signal {
	case syscall.SIGINT:
		stopCh <- syscall.SIGINT.String()
	case syscall.SIGTERM:
		stopCh <- syscall.SIGTERM.String()
	}
}

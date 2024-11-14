package timelock

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"math/big"
	"net/url"
	"os/signal"
	"slices"
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

	"github.com/smartcontractkit/timelock-worker/pkg/isclosed"
	"github.com/smartcontractkit/timelock-worker/pkg/timelock/contract"
)

// Worker represents a worker instance.
// address is an array of addresses as expected by ethereum.FilterQuery,
// but it's enforced only to one address in the logic.
type Worker struct {
	ethClient          *ethclient.Client
	contract           *contract.Timelock
	executeContract    *contract.Timelock
	abi                *abi.ABI
	address            []common.Address
	fromBlock          *big.Int
	pollPeriod         int64
	listenerPollPeriod int64
	logger             *zerolog.Logger
	privateKey         *ecdsa.PrivateKey
	scheduler
}

var httpSchemes = []string{"http", "https"}

var validNodeUrlSchemes = []string{"http", "https", "ws", "wss"}

// NewTimelockWorker initializes and returns a timelockWorker.
// It's a singleton, so further executions will retrieve the same timelockWorker.
func NewTimelockWorker(
	nodeURL, timelockAddress, callProxyAddress, privateKey string, fromBlock *big.Int,
	pollPeriod int64, listenerPollPeriod int64, logger *zerolog.Logger,
) (*Worker, error) {
	// Sanity check on each provided variable before allocating more resources.
	u, err := url.ParseRequestURI(nodeURL)
	if err != nil {
		return nil, err
	}

	if !slices.Contains(validNodeUrlSchemes, u.Scheme) {
		return nil, fmt.Errorf("invalid node URL: %s (accepted schemes are: %v)", nodeURL, validNodeUrlSchemes)
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

	if slices.Contains(httpSchemes, u.Scheme) && listenerPollPeriod <= 0 {
		return nil, fmt.Errorf("event-listener-poll-period must be a positive non-zero integer: got %d", listenerPollPeriod)
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

	tWorker := &Worker{
		ethClient:          ethClient,
		contract:           timelockContract,
		executeContract:    executeContract,
		abi:                timelockABI,
		address:            []common.Address{common.HexToAddress(timelockAddress)},
		fromBlock:          fromBlock,
		pollPeriod:         pollPeriod,
		listenerPollPeriod: listenerPollPeriod,
		logger:             logger,
		privateKey:         privateKeyECDSA,
		scheduler:          *newScheduler(time.Duration(pollPeriod) * time.Second),
	}

	return tWorker, nil
}

// Listen is the main function of a Timelock Worker.
// It handles the retrieval of old and new events, contexts and cancellations.
func (tw *Worker) Listen(ctx context.Context) error {
	ctxwc, cancel := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)

	// Log timelock-worker configuration.
	tw.startLog()

	// Run the scheduler to add/del operations in a thread-safe way.
	schedulingDone := tw.runScheduler(ctxwc)

	// Retrieve historical logs.
	historyDone, historyCh, err := tw.retrieveHistoricalLogs(ctxwc)
	if err != nil {
		tw.logger.Error().Err(err).Msg("failed to retrieve historical logs.")
		return err
	}

	// Retrieve logs asynchronously.
	newDone, logCh, err := tw.retrieveNewLogs(ctxwc)
	if err != nil {
		tw.logger.Error().Err(err).Msg("failed to subscribe and process new logs.")
		return err
	}

	// Main goroutine; processes old and new logs and handles cancellation.
	processingDone := tw.processLogs(ctxwc, historyCh, logCh)

	// Block until the context is done or until processing is completed.
	// This cover the two cases where timelock-worker can exit:
	// - A signal to stop timelock-worker was received.
	// - The subscription errored out and wasn't recovered.
	select {
	case <-ctxwc.Done():
	case <-processingDone:
		cancel()
	}

	tw.logger.Info().Msg("shutting down timelock-worker")
	tw.logger.Info().Msg("dumping operation store")
	tw.dumpOperationStore(time.Now)

	// Wait for all goroutines to finish.
	shutdownCtx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	<-isclosed.All(shutdownCtx, schedulingDone, historyDone, newDone, processingDone) //nolint:contextcheck

	return nil
}

// setupFilterQuery returns an ethereum.FilterQuery initialized to watch the Timelock contract.
func (tw *Worker) setupFilterQuery(fromBlock *big.Int) ethereum.FilterQuery {
	return ethereum.FilterQuery{
		Addresses: tw.address,
		FromBlock: fromBlock,
	}
}

// retrieveNewLogs returns a "control channel" and a "logs channels". The logs channel is where
// new log events will be asynchronously pushed to.
//
// The actual retrieveal is performed by either `subscribeNewLogs`, if the node connection
// supports subscriptions, or `pollNewLogs` otherwise. In practice, the ethclient library
// simply checks if the given node URL is "http(s)" or not.
func (tw *Worker) retrieveNewLogs(ctx context.Context) (<-chan struct{}, <-chan types.Log, error) {
	if tw.ethClient.Client().SupportsSubscriptions() {
		return tw.subscribeNewLogs(ctx)
	} else {
		return tw.pollNewLogs(ctx)
	}
}

// subscribeNewLogs subscribes to a Timelock contract and emit logs through the channel it returns.
func (tw *Worker) subscribeNewLogs(ctx context.Context) (<-chan struct{}, <-chan types.Log, error) {
	query := tw.setupFilterQuery(tw.fromBlock)
	logCh := make(chan types.Log)
	done := make(chan struct{})

	// SubscribeFilterLogs creates an asynchronous subscription to the events.
	// It receives all the new events.
	sub, err := tw.ethClient.SubscribeFilterLogs(ctx, query, logCh)
	if err != nil {
		tw.logger.Error().Msgf("unexpected error while creating subscription: %s", err.Error())
		return nil, nil, err
	}

	go func() {
		defer close(done)
		defer close(logCh)
		defer sub.Unsubscribe()

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

	return done, logCh, nil
}

// pollNewLogs periodically retrieves logs from the Timelock and emit them through the channel it returns.
func (tw *Worker) pollNewLogs(ctx context.Context) (<-chan struct{}, <-chan types.Log, error) {
	lastBlock := tw.fromBlock
	logCh := make(chan types.Log)
	done := make(chan struct{})

	go func() {
		defer close(done)
		defer close(logCh)

		tw.logger.Debug().Msgf("polling for new logs every %d seconds", tw.listenerPollPeriod)
		ticker := time.NewTicker(time.Duration(tw.listenerPollPeriod) * time.Second)
		defer ticker.Stop()

		for {
			lastBlock = tw.fetchAndDispatchLogs(ctx, logCh, lastBlock)

			select {
			case <-ticker.C:
				continue
			case <-ctx.Done():
				tw.logger.Debug().Msg("context done; stopping pollNewLogs")
				SetReadyStatus(HealthStatusError)
				return
			}
		}
	}()

	return done, logCh, nil
}

// retrieveHistoricalLogs returns a types.Log channel and retrieves all the historical events of a given contract.
// Once all the logs have been sent into the channel the function returns and the channel is closed.
func (tw *Worker) retrieveHistoricalLogs(ctx context.Context) (<-chan struct{}, <-chan types.Log, error) {
	query := tw.setupFilterQuery(tw.fromBlock)
	logCh := make(chan types.Log)
	done := make(chan struct{})

	// FIXME(gustavogama-cll): find a more elegant solution to skip retrieving historical
	// logs when using the "poll-based" event listener
	if !tw.ethClient.Client().SupportsSubscriptions() {
		tw.logger.Debug().Msgf("node does not support subscriptions; skipping historical log")
		close(done)
		close(logCh)
		return done, logCh, nil
	}

	// FilterLogs starts filtering the logs at fromBlock, gathering the historical data.
	// This is needed to guarantee that all the events are gathered, even in scenarios where the worker crashes.
	filter, err := tw.ethClient.FilterLogs(ctx, query)
	if err != nil {
		return nil, nil, err
	}

	// Process incoming historical logs in a separate goroutine.
	go func() {
		defer close(done)
		defer close(logCh)

		for _, log := range filter {
			select {
			case logCh <- log:
				tw.logger.Debug().Msgf("processing historical log: %+v\n", log)
			case <-ctx.Done():
				tw.logger.Debug().Msg("stopped while processing historical logs: incomplete retrieval.")
				return
			}
		}
		tw.logger.Debug().Msg("retrieved correctly all historical logs")
	}()

	return done, logCh, nil
}

func (tw *Worker) fetchAndDispatchLogs(ctx context.Context, logCh chan types.Log, lastBlock *big.Int) *big.Int {
	query := tw.setupFilterQuery(lastBlock)
	logs, err := tw.ethClient.FilterLogs(ctx, query)
	if err != nil {
		tw.logger.Error().Err(err).Msg("unable to fetch logs from eth client")
		SetReadyStatus(HealthStatusError) // FIXME(gustavogama-cll): wait for N errors before setting status
		return lastBlock
	}
	tw.logger.Debug().Msgf("fetched %d log entries starting from block %d", len(logs), lastBlock)
	SetReadyStatus(HealthStatusOK)

	for _, log := range logs {
		lastBlock = new(big.Int).SetUint64(max(lastBlock.Uint64(), log.BlockNumber+1))
		select {
		case logCh <- log:
			tw.logger.Debug().Interface("log", log).Msg("dispatching log")
		case <-ctx.Done():
			tw.logger.Debug().Msg("stopped while dispatching logs: incomplete retrieval.")
			break
		}
	}

	return lastBlock
}

// processLogs is implemented as a fan-in for all the logs channels, merging all the data and handling logs sequentially.
// This function is thread safe.
func (tw *Worker) processLogs(ctx context.Context, oldLog, newLog <-chan types.Log) <-chan struct{} {
	var (
		done, newDone, oldDone = make(chan struct{}), make(chan struct{}), make(chan struct{})
		ctxwc, cancel          = context.WithCancel(ctx)
	)

	// Cancel the context and shutdown the processing routine if no more logs are available.
	go func() {
		defer cancel()
		<-isclosed.All(ctxwc, oldDone, newDone)
	}()

	// This is the goroutine watching over the subscribed and historical logs.
	go func() {
		defer close(done)

		for {
			select {
			case log, open := <-newLog:
				if !open {
					close(newDone)
					newLog = nil
					continue
				}

				if err := tw.handleLog(ctxwc, log); err != nil {
					tw.logger.Error().Msgf("error processing new log: %v\n", log)
				}

			case log, open := <-oldLog:
				if !open {
					close(oldDone)
					oldLog = nil
					continue
				}

				if err := tw.handleLog(ctxwc, log); err != nil {
					tw.logger.Error().Msgf("error processing historical log: %v\n", log)
				}

			case <-ctxwc.Done():
				tw.logger.Info().Msgf("cancelled processing logs")
				SetReadyStatus(HealthStatusError)
				return
			}
		}
	}()

	return done
}

// handleLog handles the logic of parsing every event, its type and actions associated to each one.
// CallScheduled events have to be added to the scheduler.
// CallExecuted and CallCanceled signals an event that has to be removed from the scheduler.
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
	default:
		tw.logger.Info().Str("event", event.Name).Msgf("discarding event")
	}

	return nil
}

// startLog prints the timelock-worker configuration.
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
	tw.logger.Info().Msgf("\tEvent Listener Poll Period: %v", time.Duration(tw.listenerPollPeriod*int64(time.Second)).String())
}

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
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/gagliardetto/solana-go/rpc/ws"
	contracts "github.com/smartcontractkit/ccip-owner-contracts/gethwrappers"
	"go.uber.org/zap"

	"github.com/smartcontractkit/timelock-worker/pkg/isclosed"
)

// WorkerSolana represents a solana worker instance
type WorkerSolana struct {
	SolanaClient       *rpc.Client
	SolanaWSClient     *ws.Client
	timelockProgramKey *solana.PublicKey
	fromBlock          *big.Int
	pollPeriod         int64
	listenerPollPeriod int64
	pollSize           uint64
	dryRun             bool
	logger             *zap.SugaredLogger
	privateKey         *solana.PrivateKey
	scheduler          Scheduler
}

// NewTimelockWorkerSolana initializes and returns a timelockWorker.
// It's a singleton, so further executions will retrieve the same timelockWorker.
func NewTimelockWorkerSolana(
	nodeURL, timelockAddress, callProxyAddress, privateKey string, fromBlock *big.Int,
	pollPeriod int64, listenerPollPeriod int64, pollSize uint64, dryRun bool, logger *zap.SugaredLogger,
) (*WorkerSolana, error) {
	// Sanity check on each provided variable before allocating more resources.
	u, err := url.ParseRequestURI(nodeURL)
	if err != nil {
		return nil, err
	}

	if !slices.Contains(validNodeUrlSchemes, u.Scheme) {
		return nil, fmt.Errorf("invalid node URL: %s (accepted schemes are: %v)", nodeURL, validNodeUrlSchemes)
	}

	timelockPubKey, err := solana.PublicKeyFromBase58(timelockAddress)
	if err != nil {
		return nil, fmt.Errorf("timelock addresses provided is not valid: %s", timelockAddress)
	}

	if pollPeriod <= 0 {
		return nil, fmt.Errorf("poll-period must be a positive non-zero integer: got %d", pollPeriod)
	}

	if slices.Contains(httpSchemes, u.Scheme) && listenerPollPeriod <= 0 {
		return nil, fmt.Errorf("event-listener-poll-period must be a positive non-zero integer: got %d", listenerPollPeriod)
	}

	if slices.Contains(httpSchemes, u.Scheme) && pollSize == 0 {
		return nil, fmt.Errorf("event-listener-poll-size must be a positive non-zero integer: got %d", pollSize)
	}

	if fromBlock.Int64() < big.NewInt(0).Int64() {
		return nil, fmt.Errorf("from block can't be a negative number (minimum value 0): got %d", fromBlock.Int64())
	}

	if _, err := crypto.HexToECDSA(privateKey); err != nil {
		return nil, fmt.Errorf("the provided private key is not valid: got %s", privateKey)
	}

	// All variables provided are correct, start allocating new structures.
	client := rpc.New(nodeURL)

	// The contract ABI give grants capabilities such as parsing events and accessing to fields.
	// As NewTimelock only accepts one contract, hardcode it to addresses[0].
	timelockContract, err := contracts.NewRBACTimelock(common.HexToAddress(timelockAddress), ethClient)
	if err != nil {
		return nil, err
	}

	// The execute contract is the call proxy contract, which is the one that actually executes the transaction.
	// It's not the same as the timelock contract, so it has to be initialized separately.
	executeContract, err := contracts.NewRBACTimelock(common.HexToAddress(callProxyAddress), ethClient)
	if err != nil {
		return nil, err
	}

	privateKeyECDSA, err := crypto.HexToECDSA(privateKey)
	if err != nil {
		return nil, err
	}

	tWorker := &WorkerSolana{
		ethClient:          ethClient,
		contract:           timelockContract,
		executeContract:    executeContract,
		abi:                timelockABI,
		timelockProgramKey: timelockPubKey,
		fromBlock:          fromBlock,
		pollPeriod:         pollPeriod,
		listenerPollPeriod: listenerPollPeriod,
		pollSize:           pollSize,
		dryRun:             dryRun,
		logger:             logger,
		privateKey:         privateKeyECDSA,
	}

	if dryRun {
		tWorker.scheduler = newNopScheduler(logger)
	} else {
		tWorker.scheduler = newScheduler(time.Duration(pollPeriod)*time.Second, logger, tWorker.execute)
	}

	return tWorker, nil
}

// Listen is the main function of a Timelock WorkerEVM.
// It handles the retrieval of old and new events, contexts and cancellations.
func (tw *WorkerEVM) Listen(ctx context.Context) error {
	ctxwc, cancel := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)

	// Log timelock-worker configuration.
	tw.startLog()

	// Run the scheduler to add/del operations in a thread-safe way.
	schedulingDone := tw.scheduler.runScheduler(ctxwc)

	// Retrieve historical logs.
	historyDone, historyCh, err := tw.retrieveHistoricalLogs(ctxwc)
	if err != nil {
		tw.logger.With("error", err).Error("failed to retrieve historical logs.")
		return err
	}

	// Retrieve logs asynchronously.
	newDone, logCh, err := tw.retrieveNewLogs(ctxwc)
	if err != nil {
		tw.logger.With("error", err).Error("failed to subscribe and process new logs.")
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

	tw.logger.Info("shutting down timelock-worker")
	tw.logger.Info("dumping operation store")
	tw.scheduler.dumpOperationStore(time.Now)

	// Wait for all goroutines to finish.
	shutdownCtx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	<-isclosed.All(shutdownCtx, schedulingDone, historyDone, newDone, processingDone) //nolint:contextcheck

	return nil
}

// setupFilterQuery returns an ethereum.FilterQuery initialized to watch the Timelock contract.
func (tw *WorkerEVM) setupFilterQuery(fromBlock, toBlock *big.Int) ethereum.FilterQuery {
	return ethereum.FilterQuery{
		Addresses: tw.addresses,
		FromBlock: fromBlock,
		ToBlock:   toBlock,
		Topics:    [][]common.Hash{},
	}
}

// retrieveNewLogs returns a "control channel" and a "logs channels". The logs channel is where
// new log events will be asynchronously pushed to.
//
// The actual retrieval is performed by either `subscribeNewLogs`, if the node connection
// supports subscriptions, or `pollNewLogs` otherwise. In practice, the ethclient library
// simply checks if the given node URL is "http(s)" or not.
func (tw *WorkerEVM) retrieveNewLogs(ctx context.Context) (<-chan struct{}, <-chan types.Log, error) {
	if tw.ethClient.Client().SupportsSubscriptions() {
		return tw.subscribeNewLogs(ctx)
	}

	return tw.pollNewLogs(ctx)
}

// subscribeNewLogs subscribes to a Timelock contract and emit logs through the channel it returns.
func (tw *WorkerEVM) subscribeNewLogs(ctx context.Context) (<-chan struct{}, <-chan types.Log, error) {
	query := tw.setupFilterQuery(tw.fromBlock, nil)
	logCh := make(chan types.Log)
	done := make(chan struct{})

	// SubscribeFilterLogs creates an asynchronous subscription to the events.
	// It receives all the new events.
	sub, err := tw.ethClient.SubscribeFilterLogs(ctx, query, logCh)
	if err != nil {
		tw.logger.Errorf("unexpected error while creating subscription: %s", err.Error())
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
					tw.logger.Warnf("subscription error: %s", err.Error())
					SetReadyStatus(HealthStatusError)
					sub.Unsubscribe()

					success := false
					for try := range maxSubRetries {
						tw.logger.Warnf("trying to re-create subscription: %v/%v retry.", try+1, maxSubRetries)
						sub, err = tw.ethClient.SubscribeFilterLogs(ctx, query, logCh)
						if err == nil {
							tw.logger.Info("subscription successfully recreated.")
							SetReadyStatus(HealthStatusOK)
							success = true

							break
						}

						time.Sleep(time.Second * time.Duration(try))
					}

					if !success {
						tw.logger.Error("failed to recreate subscription after retries: shutting down timelock-worker.")
						return
					}
				}

			case <-ctx.Done():
				tw.logger.Debug("shutting down subscription")
				SetReadyStatus(HealthStatusError)

				return
			}
		}
	}()

	return done, logCh, nil
}

// pollNewLogs periodically retrieves logs from the Timelock and emit them through the channel it returns.
func (tw *WorkerEVM) pollNewLogs(ctx context.Context) (<-chan struct{}, <-chan types.Log, error) {
	lastBlock := tw.fromBlock
	logCh := make(chan types.Log)
	done := make(chan struct{})

	go func() {
		defer close(done)
		defer close(logCh)

		tw.logger.Debugf("polling for new logs every %d seconds", tw.listenerPollPeriod)
		ticker := time.NewTicker(time.Duration(tw.listenerPollPeriod) * time.Second)
		defer ticker.Stop()

		for {
			lastBlock = tw.fetchAndDispatchLogs(ctx, logCh, lastBlock, nil)

			select {
			case <-ticker.C:
				continue
			case <-ctx.Done():
				tw.logger.Debug("context done; stopping pollNewLogs")
				SetReadyStatus(HealthStatusError)

				return
			}
		}
	}()

	return done, logCh, nil
}

// retrieveHistoricalLogs returns a types.Log channel and retrieves all the historical events of a given contract.
// Once all the logs have been sent into the channel the function returns and the channel is closed.
func (tw *WorkerEVM) retrieveHistoricalLogs(ctx context.Context) (<-chan struct{}, <-chan types.Log, error) {
	query := tw.setupFilterQuery(tw.fromBlock, nil)
	logCh := make(chan types.Log)
	done := make(chan struct{})

	// FIXME(gustavogama-cll): find a more elegant solution to skip retrieving historical
	// logs when using the "poll-based" event listener
	if !tw.ethClient.Client().SupportsSubscriptions() {
		tw.logger.Debug("node does not support subscriptions; skipping historical log")
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
				tw.logger.Debugf("processing historical log: %+v\n", log)
			case <-ctx.Done():
				tw.logger.Debug("stopped while processing historical logs: incomplete retrieval.")
				return
			}
		}
		tw.logger.Debug("retrieved correctly all historical logs")
	}()

	return done, logCh, nil
}

func (tw *WorkerEVM) fetchAndDispatchLogs(
	ctx context.Context, logCh chan types.Log, fromBlock, currentChainBlock *big.Int,
) *big.Int {
	if currentChainBlock == nil {
		blockNumber, err := Retry(ctx, func(rctx context.Context) (uint64, error) {
			return tw.ethClient.BlockNumber(ctx)
		})
		if err != nil {
			tw.logger.With("error", err).Error("unable to fetch current block number from eth client")
			return currentChainBlock
		}
		currentChainBlock = new(big.Int).SetUint64(blockNumber)
	}
	toBlock := new(big.Int).SetUint64(min(currentChainBlock.Uint64(), fromBlock.Uint64()+tw.pollSize))

	query := tw.setupFilterQuery(fromBlock, toBlock)
	tw.logger.Debugf("fetching logs from block %v to block %v", query.FromBlock, query.ToBlock)
	logs, err := Retry(ctx, func(rctx context.Context) ([]types.Log, error) {
		return tw.ethClient.FilterLogs(rctx, query)
	})
	if err != nil {
		tw.logger.With("error", err).Error("unable to fetch logs from eth client")
		SetReadyStatus(HealthStatusError) // FIXME(gustavogama-cll): wait for N errors before setting status

		return fromBlock
	}

	tw.logger.Debugf("fetched %d log entries from block %d to block %d", len(logs), fromBlock, toBlock)
	SetReadyStatus(HealthStatusOK)

	for _, log := range logs {
		select {
		case logCh <- log:
			tw.logger.With("log", log).Debug("dispatching log")
		case <-ctx.Done():
			tw.logger.Debug("stopped while dispatching logs: incomplete retrieval.")
			return toBlock
		}
	}

	if toBlock.Cmp(currentChainBlock) < 0 {
		// we haven't reached the current block; re-run same procedure with
		// the 'toBlock` as the start block
		return tw.fetchAndDispatchLogs(ctx, logCh, toBlock, currentChainBlock)
	}

	return toBlock
}

// processLogs is implemented as a fan-in for all the logs channels, merging all the data and handling logs sequentially.
// This function is thread safe.
func (tw *WorkerEVM) processLogs(ctx context.Context, oldLog, newLog <-chan types.Log) <-chan struct{} {
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
					tw.logger.Errorf("error processing new log: %v\n", log)
				}

			case log, open := <-oldLog:
				if !open {
					close(oldDone)
					oldLog = nil

					continue
				}

				if err := tw.handleLog(ctxwc, log); err != nil {
					tw.logger.Errorf("error processing historical log: %v\n", log)
				}

			case <-ctxwc.Done():
				tw.logger.Info("cancelled processing logs")
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
func (tw *WorkerEVM) handleLog(ctx context.Context, log types.Log) error {
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
	case eventCallScheduled:
		err = tw.handleEventScheduled(ctx, log)
	case eventCallExecuted:
		err = tw.handleEventExecuted(ctx, log)
	case eventCancelled:
		err = tw.handleEventCancelled(ctx, log)
	default:
		tw.logger.With("event", event.Name).Info("discarding event")
	}

	return err
}

// A CallScheduled event should be added to an scheduler only if it's not already done
// and it's a valid Operation.
func (tw *WorkerEVM) handleEventScheduled(ctx context.Context, log types.Log) error {
	cs, err := tw.contract.ParseCallScheduled(log)
	if err != nil {
		return fmt.Errorf("failed to parse CallScheduled log: %w", err)
	}

	logger := tw.logger.With(fieldTXHash, fmt.Sprintf("%x", cs.Raw.TxHash[:])).
		With(fieldBlockNumber, cs.Raw.BlockNumber).
		With(operationID, fmt.Sprintf("%x", cs.Id))

	isDone, err := isDone(ctx, tw.contract, cs.Id)
	if err != nil {
		return fmt.Errorf("timelock.isDone call failed (operation id: %x)", cs.Id)
	}

	if !isDone {
		isOp, err := isOperation(ctx, tw.contract, cs.Id)
		if err != nil {
			return fmt.Errorf("timelock.isOperation call failed (operation id: %x)", cs.Id)
		}

		if isOp {
			logger.Infof("%s received", eventCallScheduled)
			tw.scheduler.addToScheduler(cs)
		} else {
			logger.Warn("invalid operation")
		}
	}

	return nil
}

// A CallExecuted which is in Done status should delete the task in the scheduler store.
func (tw *WorkerEVM) handleEventExecuted(ctx context.Context, log types.Log) error {
	cs, err := tw.contract.ParseCallExecuted(log)
	if err != nil {
		return fmt.Errorf("failed to parse CallExecuted log: %w", err)
	}

	logger := tw.logger.With(fieldTXHash, fmt.Sprintf("%x", cs.Raw.TxHash[:])).
		With(fieldBlockNumber, cs.Raw.BlockNumber).
		With(operationID, fmt.Sprintf("%x", cs.Id))

	isDone, err := isDone(ctx, tw.contract, cs.Id)
	if err != nil {
		return fmt.Errorf("timelock.isDone call failed (operation id: %x)", cs.Id)
	}

	if isDone {
		logger.Infof("%s received, deleting operation from scheduler", eventCallExecuted)
		tw.scheduler.delFromScheduler(cs.Id)
	} else {
		logger.Warn("operation not done; skipping deletion from scheduler")
	}

	return nil
}

// A Cancelled which is in Done status should delete the task in the scheduler store.
func (tw *WorkerEVM) handleEventCancelled(_ context.Context, log types.Log) error {
	cs, err := tw.contract.ParseCancelled(log)
	if err != nil {
		return fmt.Errorf("failed to parse Cancelled log: %w", err)
	}

	tw.logger.With(fieldTXHash, fmt.Sprintf("%x", cs.Raw.TxHash[:])).
		With(fieldBlockNumber, cs.Raw.BlockNumber).
		With(operationID, fmt.Sprintf("%x", cs.Id)).
		Infof("%s received, cancelling operation", eventCancelled)
	tw.scheduler.delFromScheduler(cs.Id)

	return nil
}

// startLog prints the timelock-worker configuration.
func (tw *WorkerEVM) startLog() {
	tw.logger.Info("timelock-worker started")
	tw.logger.Infof("\tTimelock contract addresses: %v", tw.addresses[0])

	wallet, err := privateKeyToAddress(tw.privateKey)
	if err != nil {
		tw.logger.Fatal("\tEOA addresses: unable to determine")
	}

	tw.logger.Infof("\tEOA addresses: %v", wallet)
	tw.logger.Infof("\tStarting from block: %v", tw.fromBlock)
	tw.logger.Infof("\tPoll Period: %v", time.Duration(tw.pollPeriod*int64(time.Second)).String())
	tw.logger.Infof("\tEvent Listener Poll Period: %v", time.Duration(tw.listenerPollPeriod*int64(time.Second)).String())
	tw.logger.Infof("\tEvent Listener Poll # Logs%v", tw.pollSize)
}

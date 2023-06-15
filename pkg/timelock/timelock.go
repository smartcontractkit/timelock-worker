package timelock

import (
	"context"
	"crypto/ecdsa"
	"math/big"
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
	ABI             *abi.ABI
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
		ABI:             timelockABI,
		address:         []common.Address{common.HexToAddress(timelockAddress)},
		fromBlock:       fromBlock,
		pollPeriod:      pollPeriod,
		logger:          logger,
		privateKey:      privateKeyECDSA,
		scheduler:       *newScheduler(defaultSchedulerDelay),
	}

	return tWorker, nil
}

// Listen is the main function of a Timelock Worker, it subscribes to events using the ethClient
// and targeting the contract address set.
func (tw *Worker) Listen(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	stopCh := make(chan string)
	logCh := make(chan types.Log)
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM)

	tw.updateSchedulerDelay(time.Duration(tw.pollPeriod) * time.Second)
	go tw.runScheduler(ctx)
	tw.startLog()

	// Shutdown gracefully based on OS interrupts.
	go func() {
		for {
			handleOSSignal(<-sigCh, stopCh)
		}
	}()

	/*
		The subscription and filtering can be performed via methods provided by the contract, such as
		FilterCallScheduled and WatchCallScheduled.
		However, this limits the future extensibility of timelock-worker, and does not provide
		any other advantage over using a subscription via the ethereum client.
	*/

	// FilterQuery to be feed to the subscription and FilterLogs.
	query := ethereum.FilterQuery{
		Addresses: tw.address,
		FromBlock: tw.fromBlock,
		// Topics: {{common.Hash{0x01d5062a}}},
	}

	// Create the new subscription with the predefined query.
	sub, err := tw.ethClient.SubscribeFilterLogs(ctx, query, logCh)
	if err != nil {
		return err
	}

	// Read events by FilterLogs. This method calls eth_getLogs under the hood.
	filter, err := tw.ethClient.FilterLogs(ctx, query)
	if err != nil {
		return err
	}

	go func() {
		for _, l := range filter {
			logCh <- l
		}
	}()

	// This is the goroutine watching over the subscription.
	// We want wg.Done() to cancel the whole execution, so don't add more than 1 to wg.
	// Also, when receiving an event that creates an error, skip the event and
	// continue processing the rest, as an external operator can cancel the faulty event.
	loop := true
	wg.Add(1)
	go func() {
		for loop {
			select {
			case log := <-logCh:
				// Decode the log into an event using the ABI exposed in Timelock.go
				event, err := tw.ABI.EventByID(log.Topics[0])
				if err != nil {
					continue
				}

				if event == nil {
					continue
				}

				switch event.Name {
				// A CallScheduled event should be added to an scheduler only if it's not already done
				// and it's a valid Operation.
				case eventCallScheduled:
					cs, err := tw.contract.ParseCallScheduled(log)
					if err != nil {
						continue
					}

					if !isDone(ctx, tw.contract, cs.Id) && isOperation(ctx, tw.contract, cs.Id) {
						tw.logger.Info().Hex(fieldTXHash, cs.Raw.TxHash[:]).Uint64(fieldBlockNumber, cs.Raw.BlockNumber).Msgf("%s received", eventCallScheduled)
						tw.addToScheduler(cs)
					}

					// A CallExecuted which is in Done status should delete the task in the scheduler store.
				case eventCallExecuted:
					cs, err := tw.contract.ParseCallExecuted(log)
					if err != nil {
						continue
					}

					if isDone(ctx, tw.contract, cs.Id) {
						tw.logger.Info().Hex(fieldTXHash, cs.Raw.TxHash[:]).Uint64(fieldBlockNumber, cs.Raw.BlockNumber).Msgf("%s received, skipping operation", eventCallExecuted)
						tw.delFromScheduler(cs.Id)
					}

					// A Cancelled which is in Done status should delete the task in the scheduler store.
				case eventCancelled:
					cs, err := tw.contract.ParseCancelled(log)
					if err != nil {
						continue
					}

					if isDone(ctx, tw.contract, cs.Id) {
						tw.logger.Info().Hex(fieldTXHash, cs.Raw.TxHash[:]).Uint64(fieldBlockNumber, cs.Raw.BlockNumber).Msgf("%s received, cancelling operation", eventCancelled)
						tw.delFromScheduler(cs.Id)
					}
				}

			case err := <-sub.Err():
				// Check if the error is not nil, because sub.Unsubscribe will
				// signal the channel sub.Err() to close it, leading to false nil errors.
				if err != nil {
					tw.logger.Info().Msgf("subscription: %s", err.Error())
					loop = false
				}

			case signal := <-stopCh:
				tw.logger.Info().Msgf("received OS signal %s", signal)
				loop = false
			}
		}
		wg.Done()
	}()
	wg.Wait()

	// Close in this specific order to avoid runtime panics,
	// or memory leaks.
	defer close(sigCh)
	defer close(stopCh)
	defer close(logCh)
	defer sub.Unsubscribe()
	defer ctx.Done()

	tw.dumpOperationStore()

	return nil
}

// handleOSSignal handles SIGINT and SIGTERM OS signals, and signals the stopCh.
func handleOSSignal(signal os.Signal, stopCh chan string) {
	switch signal {
	// TODO: When receiving a SIGHUP, a demon should
	// reload its own configuration.
	// case syscall.SIGHUP:
	// SIGINT and SIGTERM are different cases in case at some point
	// they have to be handled separately to add extra logic.
	case syscall.SIGINT:
		stopCh <- syscall.SIGINT.String()
	case syscall.SIGTERM:
		stopCh <- syscall.SIGTERM.String()
	}
}

func (tw *Worker) startLog() {
	tw.logger.Info().Msgf("timelock-worker started")
	tw.logger.Info().Msgf("\tTimelock contract address: %v", tw.address[0])

	wallet, err := privateKeyToAddress(tw.privateKey)
	if err != nil {
		tw.logger.Info().Msgf("\tWallet address: unable to determine")
	}
	tw.logger.Info().Msgf("\tWallet address: %v", wallet)

	tw.logger.Info().Msgf("\tStarting from block: %v", tw.fromBlock)
	tw.logger.Info().Msgf("\tPoll Period: %v seconds", tw.pollPeriod)
}

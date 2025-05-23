package timelock

import (
	"context"
	"fmt"
	"math/big"
	"net/url"
	"os/signal"
	"slices"
	"syscall"
	"time"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"

	"go.uber.org/zap"
)

// WorkerSolana represents a solana worker instance. It fetches periodically the latest signatures
// and transactions from the Solana RPC node and dispatches them to the scheduler.
type WorkerSolana struct {
	solanaClient       *rpc.Client
	timelockProgramKey solana.PublicKey
	pollPeriod         int64
	listenerPollPeriod int64
	pollSize           uint64
	dryRun             bool
	logger             *zap.SugaredLogger
	privateKey         solana.PrivateKey
	lastSignature      *solana.Signature // last signature processed
	scheduler          Scheduler
}

// NewTimelockWorkerSolana initializes and returns a timelockWorker.
// It's a singleton, so further executions will retrieve the same timelockWorker.
func NewTimelockWorkerSolana(
	nodeURL, timelockAddress, callProxyAddress, privateKey string, fromBlock *big.Int,
	pollPeriod int64, listenerPollPeriod int64, pollSize uint64, dryRun bool, logger *zap.SugaredLogger,
) (*WorkerSolana, error) {
	var privateKeySolana solana.PrivateKey
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

	if privateKeySolana, err = solana.PrivateKeyFromBase58(privateKey); err != nil {
		return nil, fmt.Errorf("the provided private key is not valid: got %s", privateKey)
	}

	// All variables provided are correct, start allocating new structures.
	client := rpc.New(nodeURL)

	tWorker := &WorkerSolana{
		solanaClient:       client,
		timelockProgramKey: timelockPubKey,
		pollPeriod:         pollPeriod,
		listenerPollPeriod: listenerPollPeriod,
		pollSize:           pollSize,
		dryRun:             dryRun,
		logger:             logger,
		privateKey:         privateKeySolana,
	}

	if dryRun {
		tWorker.scheduler = nil // TODO: add solana nopScheduler implementation
	} else {
		tWorker.scheduler = nil // TODO: add solana Scheduler implementation
	}

	return tWorker, nil
}

// Listen is the main function of a Timelock WorkerSolana.
// It handles the retrieval of old and new events, contexts and cancellations.
func (w *WorkerSolana) Listen(ctx context.Context) error {
	ctxwc, _ := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)

	// Log timelock-worker configuration.
	w.startLog()

	// Run the scheduler to add/del operations in a thread-safe way.
	_ = w.scheduler.runScheduler(ctxwc)

	panic("solana log processing not implemented yet")
	//// Retrieve logs asynchronously.
	//pollDone, txCh := w.pollNewSignatures(ctxwc)
	//
	//// Start processing transactions
	//procDone := w.processTransactions(ctx, txCh)
	//// Block until the context is done or until processing is completed.
	//// This cover the two cases where timelock-worker can exit:
	//// - A signal to stop timelock-worker was received.
	//// - The subscription errored out and wasn't recovered.
	//select {
	//case <-ctxwc.Done():
	//case <-procDone:
	//	cancel()
	//}
	//
	//w.logger.Info("shutting down timelock-worker")
	//w.logger.Info("dumping operation store")
	//w.scheduler.dumpOperationStore(time.Now)
	//
	//// Wait for all goroutines to finish.
	//shutdownCtx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	//defer cancel()
	//
	//<-isclosed.All(shutdownCtx, schedulingDone, pollDone, procDone) //nolint:contextcheck
	//
	//return nil
}

// startLog prints the timelock-worker configuration.
func (w *WorkerSolana) startLog() {
	w.logger.Info("timelock-worker started [solana]")
	w.logger.Infof("\tTimelock prorgam addresses: %v", w.timelockProgramKey.String())

	wallet := w.privateKey.PublicKey()

	w.logger.Infof("\tsolana account address: %v", wallet)
	w.logger.Infof("\tPoll Period: %v", time.Duration(w.pollPeriod*int64(time.Second)).String())
	w.logger.Infof("\tEvent Listener Poll Period: %v", time.Duration(w.listenerPollPeriod*int64(time.Second)).String())
	w.logger.Infof("\tEvent Listener Poll # Logs%v", w.pollSize)
}

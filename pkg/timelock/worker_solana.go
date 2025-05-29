package timelock

import (
	"context"
	"fmt"
	"net/url"
	"os/signal"
	"slices"
	"syscall"
	"time"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/gagliardetto/solana-go/rpc/jsonrpc"

	"github.com/smartcontractkit/mcms/sdk"
	solanasdk "github.com/smartcontractkit/mcms/sdk/solana"
	"go.uber.org/zap"

	"github.com/smartcontractkit/timelock-worker/pkg/isclosed"
)

// WorkerSolana represents a solana worker instance. It fetches periodically the latest signatures
// and transactions from the Solana RPC node and dispatches them to the scheduler.
type WorkerSolana struct {
	solanaClient        *rpc.Client
	timelockProgramKey  solana.PublicKey
	instanceSeed        solanasdk.PDASeed // instanceSeed is the seed used to derive the timelock instance.
	timelockFullAddress string            // timelockFullAddress is the full address of the timelock program <programID>.<instanceSeed>
	pollPeriod          int64
	listenerPollPeriod  int64
	pollSize            int
	dryRun              bool
	inspector           sdk.TimelockInspector // inspector is used to query timelock state
	logger              *zap.SugaredLogger
	privateKey          solana.PrivateKey
	lastSignature       *solana.Signature // last signature processed
	scheduler           Scheduler
}

// NewTimelockWorkerSolana initializes and returns a timelockWorker.
func NewTimelockWorkerSolana(
	nodeURL,
	timelockAddress,
	privateKey string,
	pollPeriod int64,
	listenerPollPeriod int64,
	pollSize int,
	dryRun bool,
	logger *zap.SugaredLogger,
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

	timelockPubKey, instanceSeed, err := solanasdk.ParseContractAddress(timelockAddress)
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
	inspector := solanasdk.NewTimelockInspector(client)

	tWorker := &WorkerSolana{
		solanaClient:        client,
		instanceSeed:        instanceSeed,
		timelockFullAddress: timelockAddress,
		timelockProgramKey:  timelockPubKey,
		pollPeriod:          pollPeriod,
		listenerPollPeriod:  listenerPollPeriod,
		pollSize:            pollSize,
		inspector:           inspector,
		dryRun:              dryRun,
		logger:              logger,
		privateKey:          privateKeySolana,
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
	ctxwc, cancel := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)

	// Log timelock-worker configuration.
	w.startLog()

	// Run the scheduler to add/del operations in a thread-safe way.
	//schedulingDone = w.scheduler.runScheduler(ctxwc)

	//// Retrieve logs asynchronously.
	pollDone, txCh := w.pollNewSignatures(ctxwc)

	// Start processing transactions
	procDone := w.processTransactions(ctxwc, txCh)
	// Block until the context is done or until processing is completed.
	// This cover the two cases where timelock-worker can exit:
	// - A signal to stop timelock-worker was received.
	// - The subscription errored out and wasn't recovered.
	select {
	case <-ctxwc.Done():
	case <-procDone:
		cancel()
	}

	w.logger.Info("shutting down timelock-worker")
	w.logger.Info("dumping operation store")
	// TODO: re-add when scheduler is implemented
	//w.scheduler.dumpOperationStore(time.Now)

	// Wait for all goroutines to finish.
	shutdownCtx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	<-isclosed.All(shutdownCtx, pollDone, procDone) //nolint:contextcheck

	return nil
}

// pollNewSignatures continuously fetches only new signatures touching your program,
// then loads each confirmed transaction so you can parse its LogMessages.
func (w *WorkerSolana) pollNewSignatures(
	ctx context.Context,
) (<-chan struct{}, <-chan *rpc.TransactionWithMeta) {
	done := make(chan struct{})
	txCh := make(chan *rpc.TransactionWithMeta, w.pollSize)

	go func() {
		defer func() {
			w.logger.Info("pollNewSignatures exiting")
			close(done)
			close(txCh)
		}()

		w.logger.Infof(
			"starting pollNewSignatures: program=%s pollPeriod=%ds pollSize=%d",
			w.timelockProgramKey,
			w.pollPeriod,
			w.pollSize,
		)

		ticker := time.NewTicker(time.Duration(w.pollPeriod) * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				w.logger.Info("context cancelled")
				return

			case <-ticker.C:
				w.logger.Infof("new pollNewSignatures tick: %s", time.Now().Format(time.RFC3339))

				var until solana.Signature
				if w.lastSignature != nil {
					until = *w.lastSignature
				}

				// Always try to fetch signatures
				sigs, err := w.solanaClient.GetSignaturesForAddressWithOpts(
					ctx,
					w.timelockProgramKey,
					&rpc.GetSignaturesForAddressOpts{
						Limit: &w.pollSize,
						Until: until,
					},
				)
				if err != nil {
					w.logger.Errorf("GetSignaturesForAddress failed: %v", err)
					continue
				}
				if len(sigs) == 0 {
					if w.lastSignature == nil {
						w.logger.Warn("no existing sigs to anchor")
					} else {
						w.logger.Debug("no new sigs")
					}

					continue
				}

				w.logger.Infof("found %d new signatures", len(sigs))

				// Build JSON-RPC batch requests (oldest→newest)
				requests := make(jsonrpc.RPCRequests, len(sigs))
				for i, info := range sigs {
					j := len(sigs) - 1 - i // reverse to oldest first
					requests[j] = jsonrpc.NewRequest(
						"getTransaction",
						info.Signature,
						&rpc.GetTransactionOpts{Encoding: solana.EncodingBase58},
					)
				}

				// Call getTransaction in batch
				resps, err := w.solanaClient.RPCCallBatch(ctx, requests)
				if err != nil {
					w.logger.Errorf("RPCCallBatch failed: %v", err)
					continue
				}

				for i, resp := range resps {
					sigInfo := sigs[len(sigs)-1-i]
					if resp.Error != nil {
						w.logger.Warnf("tx %s failed: %v", sigInfo.Signature, resp.Error)
						continue
					}

					tx := new(rpc.TransactionWithMeta)
					if err := resp.GetObject(&tx); err != nil {
						w.logger.Warnf("decode %s failed: %v", sigInfo.Signature, err)
						continue
					}

					w.lastSignature = &sigInfo.Signature
					txCh <- tx
				}
			}
		}
	}()

	return done, txCh
}

func (w *WorkerSolana) processTransactions(ctx context.Context, txChannel <-chan *rpc.TransactionWithMeta) <-chan struct{} {
	var (
		done, newDone, oldDone = make(chan struct{}), make(chan struct{}), make(chan struct{})
		ctxwc, cancel          = context.WithCancel(ctx)
	)

	// Cancel the context and shutdown the processing routine if no more logs are available.
	go func() {
		defer cancel()
		<-isclosed.All(ctxwc, oldDone, newDone)
	}()

	// This is the goroutine watching over the polled txs in solana
	go func() {
		defer close(done)

		for {
			select {
			case log, open := <-txChannel:
				if !open {
					close(newDone)
					txChannel = nil

					continue
				}

				if err := w.handleTx(ctxwc, log); err != nil {
					w.logger.Errorf("error processing new log: %v\n", log)
				}

			case <-ctxwc.Done():
				w.logger.Info("cancelled processing logs")
				SetReadyStatus(HealthStatusError)

				return
			}
		}
	}()

	return done
}

// handleTx handles the logic of parsing every solana tx, it looks through the tx logs and
// parses the events emitted by the timelock program. Any other events are ignored.
func (w *WorkerSolana) handleTx(ctx context.Context, tx *rpc.TransactionWithMeta) error {
	// ignore tx with no logs
	if len(tx.Meta.LogMessages) == 0 {
		return nil
	}

	timelockEvent, err := ParseTimelockEvents(w.logger, tx)
	if err != nil {
		return fmt.Errorf("failed to parse timelock events: %w", err)
	}

	for _, scheduledEvent := range timelockEvent.Scheduled {
		w.logger.Debugf("found event scheduled: %s", scheduledEvent.ID)
		err = w.handleEventScheduled(ctx, scheduledEvent)
		if err != nil {
			w.logger.Errorf("error handling scheduled event: %v", err)
			w.logger.Warnf("skipping event scheduled due to failure in checking operation state: %s", scheduledEvent)

			continue
		}
	}
	for _, executedEvent := range timelockEvent.Executed {
		w.logger.Debugf("found event executed: %s", executedEvent.ID)
		err = w.handleEventExecuted(ctx, executedEvent)
		if err != nil {
			w.logger.Errorf("error handling executed event: %v", err)
			w.logger.Warnf("skipping event executed due to failure in checking operation state: %s", executedEvent)

			continue
		}
	}
	for _, bypasserEvent := range timelockEvent.Cancelled {
		w.logger.Debugf("found event cancelled: %s", bypasserEvent.ID)
		w.handleEventCancelled(ctx, bypasserEvent)
	}

	return nil
}

// handleEventCancelled checks if the operation is cancelled and deletes it from the scheduler if it is.
func (w *WorkerSolana) handleEventCancelled(_ context.Context, event Cancelled) {
	w.logger.With(operationID, fmt.Sprintf("%x", event.ID)).
		Infof("%s received, cancelling operation", eventCancelled)

	// TODO: add scheduler call once scheduler is implemented
	//w.scheduler.delFromScheduler(event.ID)
}

// handleEventExecuted checks if the operation is done and deletes it from the scheduler if it is.
func (w *WorkerSolana) handleEventExecuted(ctx context.Context, event CallExecuted) error {
	logger := w.logger.With(eventIndex, fmt.Sprintf("%x", event.Index)).
		With(eventTarget, event.Target.String()).
		With(operationID, fmt.Sprintf("%x", event.ID))

	isDone, err := w.inspector.IsOperationDone(ctx, w.timelockFullAddress, event.ID)
	if err != nil {
		return fmt.Errorf("timelock.isOperationDone call failed (operation id: %x): %w", event.ID, err)
	}
	if isDone {
		logger.Infof("%s received, deleting operation from scheduler", eventCallExecuted)
		// TODO: add scheduler call once scheduler is implemented
		//w.scheduler.delFromScheduler(event.ID)
	} else {
		logger.Warn("operation not done; skipping deletion from scheduler")
	}

	return nil
}

// handleEventScheduled checks if the operation is already scheduled and adds it to the scheduler if it is not.
func (w *WorkerSolana) handleEventScheduled(ctx context.Context, event CallScheduled) error {
	logger := w.logger.With(eventIndex, fmt.Sprintf("%x", event.Index)).
		With(eventTarget, event.Target.String()).
		With(operationID, fmt.Sprintf("%x", event.ID))

	isDone, err := w.inspector.IsOperationDone(ctx, w.timelockFullAddress, event.ID)
	if err != nil {
		return fmt.Errorf("timelock.isOperationDone call failed (operation id: %x): %w", event.ID, err)
	}
	if !isDone {
		isOp, err := w.inspector.IsOperation(ctx, w.timelockFullAddress, event.ID)
		if err != nil {
			return fmt.Errorf("timelock.isOperation call failed (operation id: %x)", event.ID)
		}

		if isOp {
			logger.Infof("%s received", eventCallScheduled)
			// TODO: add scheduler call once scheduler is implemented
			// w.scheduler.addToScheduler(cs)
		} else {
			logger.Warn("invalid operation")
		}
	}

	return nil
}

// startLog prints the timelock-worker configuration.
func (w *WorkerSolana) startLog() {
	w.logger.Info("timelock-worker started [solana]")
	w.logger.Infof("\tTimelock program addresses: %v", w.timelockProgramKey.String())

	wallet := w.privateKey.PublicKey()

	w.logger.Infof("\tSolana account address: %v", wallet)
	w.logger.Infof("\tPoll Period: %v", time.Duration(w.pollPeriod*int64(time.Second)).String())
	w.logger.Infof("\tEvent Listener Poll Period: %v", time.Duration(w.listenerPollPeriod*int64(time.Second)).String())
	w.logger.Infof("\tEvent Listener Poll #Logs: %v", w.pollSize)
}

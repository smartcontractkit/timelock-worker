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
	commitmentType      rpc.CommitmentType
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
	commitmentType rpc.CommitmentType,
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
		commitmentType:      commitmentType,
	}

	if dryRun {
		tWorker.scheduler = newNopScheduler(logger)
	} else {
		tWorker.scheduler = newScheduler(time.Duration(pollPeriod)*time.Second, logger, func(context.Context, []TimelockCallScheduled) {})
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
	schedulingDone := w.scheduler.runScheduler(ctxwc)

	//// Retrieve logs asynchronously.
	pollDone, txCh := w.startPolling(ctxwc)

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

	w.scheduler.dumpOperationStore(time.Now)

	// Wait for all goroutines to finish.
	shutdownCtx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	<-isclosed.All(shutdownCtx, schedulingDone, pollDone, procDone) //nolint:contextcheck

	return nil
}

func (w *WorkerSolana) startPolling(ctx context.Context) (<-chan struct{}, <-chan *rpc.TransactionWithMeta) {
	done := make(chan struct{})
	txCh := make(chan *rpc.TransactionWithMeta, w.pollSize)
	sigCh := make(chan solana.Signature, 100)

	go w.pollSignatures(ctx, sigCh, done)
	go w.loadTransactions(ctx, sigCh, txCh)

	return done, txCh
}

// pollSignatures periodically polls new signatures from the given timelock program and sends them to the tx fetching channel.
func (w *WorkerSolana) pollSignatures(ctx context.Context, sigCh chan<- solana.Signature, done chan struct{}) {
	ticker := time.NewTicker(time.Duration(w.pollPeriod) * time.Second)
	defer ticker.Stop()
	defer close(sigCh)
	defer close(done)

	w.logger.Infow("starting pollSignatures", "program", w.timelockProgramKey)

	for {
		select {
		case <-ctx.Done():
			w.logger.Info("pollSignatures context cancelled")
			return
		case <-ticker.C:
			var until solana.Signature
			if w.lastSignature != nil {
				until = *w.lastSignature
			}

			sigs, err := w.solanaClient.GetSignaturesForAddressWithOpts(
				ctx,
				w.timelockProgramKey,
				&rpc.GetSignaturesForAddressOpts{
					Limit:      &w.pollSize,
					Until:      until,
					Commitment: w.commitmentType,
				},
			)
			if err != nil {
				w.logger.Errorf("GetSignaturesForAddress failed: %v", err)
				continue
			}
			if len(sigs) == 0 {
				continue
			}

			slices.Reverse(sigs)
			for _, info := range sigs {
				select {
				case sigCh <- info.Signature:
				case <-ctx.Done():
					return
				}
				w.lastSignature = &info.Signature
			}
		}
	}
}

// loadTransactions tries getting a tx multiple times.
func (w *WorkerSolana) loadTransactions(ctx context.Context, sigCh <-chan solana.Signature, txCh chan<- *rpc.TransactionWithMeta) {
	for {
		select {
		case <-ctx.Done():
			w.logger.Info("loadTransactions context cancelled")
			return
		case sig, ok := <-sigCh:
			if !ok {
				return
			}
			tx, err := w.retryGetTransaction(ctx, sig)
			if err != nil {
				w.logger.Warnf("getTransaction %s failed permanently: %v", sig, err)
				continue
			}
			select {
			case txCh <- tx:
			case <-ctx.Done():
				return
			}
		}
	}
}

// rpcErrorOrDefault returns rpc error.
func rpcErrorOrDefault(err error, resps jsonrpc.RPCResponses) error {
	if len(resps) > 0 && resps[0].Error != nil {
		return resps[0].Error
	}

	return err
}

func (w *WorkerSolana) retryGetTransaction(ctx context.Context, sig solana.Signature) (*rpc.TransactionWithMeta, error) {
	var tx *rpc.TransactionWithMeta

	operation := func() error {
		req := jsonrpc.NewRequest("getTransaction", sig, &rpc.GetTransactionOpts{Encoding: solana.EncodingBase58, Commitment: w.commitmentType})
		resps, err := w.solanaClient.RPCCallBatch(ctx, jsonrpc.RPCRequests{req})
		if err != nil || len(resps) == 0 || resps[0].Error != nil {
			return rpcErrorOrDefault(err, resps)
		}

		tx = new(rpc.TransactionWithMeta)
		if err := resps[0].GetObject(&tx); err != nil {
			return err
		}

		return nil // success
	}

	tx, err := Retry(ctx, func(ctx context.Context) (*rpc.TransactionWithMeta, error) {
		err := operation()
		if err != nil {
			w.logger.Warnw("Retryable error", "err", err)
		}

		return tx, err
	})
	if err != nil {
		w.logger.Errorw("Retry failed", "err", err)
		return nil, err
	}

	return tx, nil
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
			case tx, open := <-txChannel:
				if !open {
					close(newDone)
					txChannel = nil

					continue
				}

				if err := w.handleTx(ctxwc, tx); err != nil {
					w.logger.Errorf("error processing new tx: %w %v\n", err, tx)
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

	timelockEvent, err := ParseTimelockEvents(tx)
	if err != nil {
		return fmt.Errorf("failed to parse timelock events: %w", err)
	}

	for _, scheduledEvent := range timelockEvent.Scheduled {
		w.logger.Debugf("found event scheduled: %s", scheduledEvent.ID)
		err = w.handleEventScheduled(ctx, scheduledEvent)
		if err != nil {
			w.logger.Errorf("error handling scheduled event: %v continuing with next event...", err)
			continue
		}
	}
	for _, executedEvent := range timelockEvent.Executed {
		w.logger.Debugf("found event executed: %s", executedEvent.ID)
		err = w.handleEventExecuted(ctx, executedEvent)
		if err != nil {
			w.logger.Errorf("error handling executed event: %v continuing with next event...", err)
			continue
		}
	}
	for _, cancellerEvent := range timelockEvent.Cancelled {
		w.logger.Debugf("found event cancelled: %s continuing with next event...", cancellerEvent.ID)
		w.handleEventCancelled(ctx, cancellerEvent)
	}

	return nil
}

// handleEventCancelled checks if the operation is cancelled and deletes it from the scheduler if it is.
func (w *WorkerSolana) handleEventCancelled(_ context.Context, event SolanaTimelockCallCancelledEvent) {
	w.logger.With(operationID, fmt.Sprintf("%x", event.ID)).
		Infow("event received, cancelling operation", "event type", eventCancelled)

	w.scheduler.delFromScheduler(event.ID)
}

// handleEventExecuted checks if the operation is done and deletes it from the scheduler if it is.
func (w *WorkerSolana) handleEventExecuted(ctx context.Context, event SolanaTimelockCallExecutedEvent) error {
	logger := w.logger.With(eventIndex, fmt.Sprintf("%x", event.Index)).
		With(eventTarget, event.Target.String()).
		With(operationID, fmt.Sprintf("%x", event.ID))

	isDone, err := w.inspector.IsOperationDone(ctx, w.timelockFullAddress, event.ID)
	if err != nil {
		return fmt.Errorf("timelock.isOperationDone call failed (operation id: %x): %w", event.ID, err)
	}
	if isDone {
		logger.Infow("event received, deleting operation from scheduler", "event type ", eventCallExecuted)

		w.scheduler.delFromScheduler(event.ID)
	} else {
		logger.Warn("operation not done; skipping deletion from scheduler")
	}

	return nil
}

// handleEventScheduled checks if the operation is already scheduled and adds it to the scheduler if it is not.
func (w *WorkerSolana) handleEventScheduled(ctx context.Context, event SolanaTimelockCallScheduledEvent) error {
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
			logger.Infow("event received", "event type", eventCallScheduled)
			w.scheduler.addToScheduler(&solanaTimelockCallScheduled{callScheduledEvent: event})
		} else {
			logger.Warn("invalid operation")
		}
	}

	return nil
}

// startLog prints the timelock-worker configuration.
func (w *WorkerSolana) startLog() {
	w.logger.Info("timelock-worker started [solana]")
	w.logger.Infow("\tTimelock program:", "program address", w.timelockProgramKey.String())

	wallet := w.privateKey.PublicKey()

	w.logger.Infow("\tSolana account address", "wallet address", wallet)
	w.logger.Infow("\tPoll Period", "period", time.Duration(w.pollPeriod*int64(time.Second)).String())
	w.logger.Infow("\tEvent Listener Poll Period", "period", time.Duration(w.listenerPollPeriod*int64(time.Second)).String())
	w.logger.Infow("\tEvent Listener Poll #Logs", "period", w.pollSize)
}

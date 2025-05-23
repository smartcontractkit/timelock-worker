package timelock

import (
	"context"
	"strings"

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

	"github.com/smartcontractkit/timelock-worker/pkg/isclosed"
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
	ctxwc, cancel := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)

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

// ptrInt is a helper to take an int and return *int
func ptrInt(i int) *int { return &i }

// pollNewSignatures continuously fetches only new signatures touching your program,
// then loads each confirmed transaction so you can parse its LogMessages.
func (w *WorkerSolana) pollNewSignatures(
	ctx context.Context,
) (<-chan struct{}, <-chan *rpc.TransactionWithMeta) {
	done := make(chan struct{})
	txCh := make(chan *rpc.TransactionWithMeta)

	go func() {
		w.logger.Infof(
			"starting pollNewSignatures: program=%s pollPeriod=%ds pollSize=%d",
			w.timelockProgramKey,
			w.pollPeriod,
			w.pollSize,
		)
		defer func() {
			w.logger.Info("pollNewSignatures exiting")
			close(done)
			close(txCh)
		}()

		// 1) First run: anchor to the very latest signature so we don't replay history
		if w.lastSignature == nil {
			w.logger.Debug("anchoring to latest signature (first run)")
			docs, err := w.solanaClient.GetSignaturesForAddressWithOpts(
				ctx,
				w.timelockProgramKey,
				&rpc.GetSignaturesForAddressOpts{
					Limit: ptrInt(1),
				},
			)
			if err != nil {
				w.logger.Errorf("failed to anchor to latest signature: %v", err)
			} else if len(docs) == 0 {
				w.logger.Warn("no existing signatures found to anchor to")
			} else {
				sig0 := docs[0].Signature
				w.lastSignature = &sig0
				w.logger.Infof("anchored to latest signature: %s", sig0)
			}
		}

		ticker := time.NewTicker(time.Duration(w.pollPeriod) * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				w.logger.Info("context canceled; stopping pollNewSignatures")
				return

			case <-ticker.C:
				// 2) Fetch everything newer than lastSignature
				var until solana.Signature
				if w.lastSignature != nil {
					until = *w.lastSignature
				}
				w.logger.Debugf("polling new signatures since: %s (limit %d)", until, w.pollSize)

				limit := int(w.pollSize)
				sigs, err := w.solanaClient.GetSignaturesForAddressWithOpts(
					ctx,
					w.timelockProgramKey,
					&rpc.GetSignaturesForAddressOpts{
						Limit: ptrInt(limit),
						Until: until,
					},
				)
				if err != nil {
					w.logger.Errorf("GetSignaturesForAddressWithOpts failed: %v", err)
					continue
				}
				if len(sigs) == 0 {
					w.logger.Debug("no new signatures found")
					continue
				}
				w.logger.Infof("found %d new signatures", len(sigs))

				// 3) RPC returns newest→oldest; process oldest→newest
				for i := len(sigs) - 1; i >= 0; i-- {
					sig := sigs[i].Signature
					w.logger.Debugf("processing signature: %s", sig)
					w.lastSignature = &sig

					// 4) Fetch the confirmed transaction (with logs)
					tx, err := w.solanaClient.GetConfirmedTransactionWithOpts(
						ctx,
						sig,
						&rpc.GetTransactionOpts{
							Encoding: solana.EncodingBase58,
						},
					)
					if err != nil {
						w.logger.Warnf("GetConfirmedTransaction(%s) failed: %v", sig, err)
						continue
					}
					w.logger.Debugf("fetched transaction for signature: %s", sig)
					txCh <- tx
				}
			}
		}
	}()

	return done, txCh
}

// processTransactions reads parsed transactions and dispatches events to the scheduler
func (w *WorkerSolana) processTransactions(
	ctx context.Context,
	txCh <-chan *rpc.TransactionWithMeta,
) <-chan struct{} {
	done := make(chan struct{})

	go func() {
		defer close(done)

		for {
			select {
			case <-ctx.Done():
				return
			case tx, ok := <-txCh:
				if !ok {
					return
				}

				for _, msg := range tx.Meta.LogMessages {
					// Scheduled
					if isCallScheduledLog(msg) {
						ev, err := parseCallScheduled(msg)
						if err != nil {
							w.logger.Warnf("parseCallScheduled failed: %v", err)
							continue
						}
						w.logger.Infof("scheduling op %s at %s", ev.ID, ev.ETA)
						w.scheduler.addToScheduler(ev)

						// Executed
					} else if isCallExecutedLog(msg) {
						ev, err := parseCallExecuted(msg)
						if err != nil {
							w.logger.Warnf("parseCallExecuted failed: %v", err)
							continue
						}
						w.logger.Infof("executed op %s", ev.ID)
						w.scheduler.delFromScheduler(ev.ID)

						// Cancelled
					} else if isCallCancelledLog(msg) {
						ev, err := parseCallCancelled(msg)
						if err != nil {
							w.logger.Warnf("parseCallCancelled failed: %v", err)
							continue
						}
						w.logger.Infof("cancelled op %s", ev.ID)
						w.scheduler.delFromScheduler(ev.ID)
					}
				}
			}
		}
	}()

	return done
}

// Helpers for detecting and parsing on-chain logs
func isCallScheduledLog(msg string) bool {
	return strings.HasPrefix(msg, "Program log: EVENT:CallScheduled")
}
func isCallExecutedLog(msg string) bool {
	return strings.HasPrefix(msg, "Program log: EVENT:CallExecuted")
}
func isCallCancelledLog(msg string) bool {
	return strings.HasPrefix(msg, "Program log: EVENT:CallCancelled")
}

// startLog prints the timelock-worker configuration.
func (w *WorkerSolana) startLog() {
	w.logger.Info("timelock-worker started [solana]")
	w.logger.Infof("\tTimelock prorgam addresses: %v", w.timelockProgramKey.String())

	wallet := w.privateKey.PublicKey()

	w.logger.Infof("\taccount address: %v", wallet)
	w.logger.Infof("\tPoll Period: %v", time.Duration(w.pollPeriod*int64(time.Second)).String())
	w.logger.Infof("\tEvent Listener Poll Period: %v", time.Duration(w.listenerPollPeriod*int64(time.Second)).String())
	w.logger.Infof("\tEvent Listener Poll # Logs%v", w.pollSize)
}

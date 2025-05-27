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
	"github.com/gagliardetto/solana-go/rpc/jsonrpc"
	timelockgb "github.com/smartcontractkit/chainlink-ccip/chains/solana/gobindings/timelock"
	"go.uber.org/zap"
)

// TimelockEvents groups all of the decoded events from one tx.
type TimelockEvents struct {
	Scheduled []timelockgb.CallScheduled
	Executed  []timelockgb.CallExecuted
	Cancelled []timelockgb.CallCancelled
}

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

// ptrInt is a helper to take an int and return *int
func ptrInt(i int) *int { return &i }

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

		// 1) Anchor on first run
		if w.lastSignature == nil {
			signaturesResp, err := w.solanaClient.GetSignaturesForAddressWithOpts(
				ctx,
				w.timelockProgramKey,
				&rpc.GetSignaturesForAddressOpts{Limit: ptrInt(1)},
			)
			if err != nil {
				w.logger.Errorf("anchor failed: %v", err)
			} else if len(signaturesResp) > 0 {
				w.lastSignature = &signaturesResp[0].Signature
				w.logger.Infof("anchored to latest: %s", *w.lastSignature)
			} else {
				w.logger.Warn("no existing sigs to anchor")
			}
		}

		ticker := time.NewTicker(time.Duration(w.pollPeriod) * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				w.logger.Info("context cancelled")
				return

			case <-ticker.C:
				// 2) fetch sigs
				var until solana.Signature
				if w.lastSignature != nil {
					until = *w.lastSignature
				}
				sigs, err := w.solanaClient.GetSignaturesForAddressWithOpts(
					ctx,
					w.timelockProgramKey,
					&rpc.GetSignaturesForAddressOpts{
						Limit: ptrInt(int(w.pollSize)),
						Until: until,
					},
				)
				if err != nil {
					w.logger.Errorf("GetSignaturesForAddress failed: %v", err)
					continue
				}
				if len(sigs) == 0 {
					w.logger.Debug("no new sigs")
					continue
				}
				w.logger.Infof("found %d new signatures", len(sigs))

				// 3) build JSON-RPC batch requests (oldest→newest)
				requests := make(jsonrpc.RPCRequests, len(sigs))
				for i, info := range sigs {
					// reverse so index 0 = oldest
					j := len(sigs) - 1 - i
					requests[j] = jsonrpc.NewRequest(
						"getTransaction",
						info.Signature,
						&rpc.GetTransactionOpts{Encoding: solana.EncodingBase58},
					)
				}

				// 4) fire them all at once
				resps, err := w.solanaClient.RPCCallBatch(ctx, requests)
				if err != nil {
					w.logger.Errorf("RPCCallBatch failed: %v", err)
					continue
				}

				// 5) emit results in order
				for i, resp := range resps {
					// match back to sigs reversed
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

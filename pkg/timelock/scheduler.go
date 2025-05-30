package timelock

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"math/big"
	"os"
	"slices"
	"sync"
	"time"

	eth "github.com/ethereum/go-ethereum/common"
	"go.uber.org/zap"
)

type operationKey [32]byte

func operationKeyFromHex(hex string) operationKey {
	return operationKey(eth.HexToHash(hex))
}

type TimelockCallScheduled interface {
	Id() operationKey
	Index() int
	BlockNumber() *big.Int
	TxHash() string
}

type Scheduler interface {
	runScheduler(ctx context.Context) <-chan struct{}
	addToScheduler(op TimelockCallScheduled)
	delFromScheduler(op operationKey)
	dumpOperationStore(now func() time.Time)
}

type executeFn func(context.Context, []TimelockCallScheduled)

// Scheduler represents a scheduler with an in memory store.
// Whenever accesing the map the mutex should be Locked, to prevent
// any race condition.
type scheduler struct {
	mu        sync.Mutex
	ticker    *time.Ticker
	add       chan TimelockCallScheduled
	del       chan operationKey
	store     map[operationKey][]TimelockCallScheduled
	busy      bool
	logger    *zap.SugaredLogger
	executeFn executeFn
}

// newScheduler returns a new initialized scheduler.
func newScheduler(tick time.Duration, logger *zap.SugaredLogger, executeFn executeFn) *scheduler {
	s := &scheduler{
		ticker:    time.NewTicker(tick),
		add:       make(chan TimelockCallScheduled),
		del:       make(chan operationKey),
		store:     make(map[operationKey][]TimelockCallScheduled),
		busy:      false,
		logger:    logger,
		executeFn: executeFn,
	}

	return s
}

// runScheduler starts the scheduler.
// ticker.C will signal the scheduler every N seconds, where N is
// defined in the scheduler.ticker field
// add and del are the channels to manage the store, we
// call them this way so no process is allowd to add/delete from
// the store, which could cause race conditions like adding/deleting
// while the operation is being executed.
func (tw *scheduler) runScheduler(ctx context.Context) <-chan struct{} {
	done := make(chan struct{})

	go func() {
		defer close(done)
		for {
			select {
			case <-tw.ticker.C:
				if len(tw.store) <= 0 {
					tw.logger.Debug("new scheduler tick: no operations in store")
					continue
				}

				if !tw.isSchedulerBusy() {
					tw.logger.Debug("new scheduler tick: operations in store")
					tw.setSchedulerBusy()
					for _, op := range tw.store {
						tw.executeFn(ctx, op)
					}
					tw.setSchedulerFree()
				} else {
					tw.logger.Debug("new scheduler tick: scheduler is busy, skipping until next tick")
				}

			case op := <-tw.add:
				tw.mu.Lock()
				for len(tw.store[op.Id()]) <= op.Index() {
					tw.store[op.Id()] = append(tw.store[op.Id()], op)
				}
				tw.store[op.Id()][op.Index()] = op
				tw.mu.Unlock()
				tw.logger.Debugf("scheduled operation: %x", op.Id())

			case op := <-tw.del:
				if _, ok := tw.store[op]; ok {
					tw.mu.Lock()
					delete(tw.store, op)
					tw.mu.Unlock()
					tw.logger.Debugf("de-scheduled operation: %x", op)
				}

			case <-ctx.Done():
				tw.logger.Debug("shutting down scheduler")
				return
			}
		}
	}()

	return done
}

// updateSchedulerDelay updates the internal ticker delay, so it can be reconfigured while running.
func (tw *scheduler) updateSchedulerDelay(t time.Duration) {
	if t <= 0 {
		tw.logger.Debugf("internal min delay not changed, invalid duration: %v", t.String())
		return
	}

	tw.ticker.Reset(t)
	tw.logger.Debugf("internal min delay changed to %v", t.String())
}

// addToScheduler adds a new CallSchedule operation safely to the store.
func (tw *scheduler) addToScheduler(op TimelockCallScheduled) {
	tw.logger.Debugf("scheduling operation: %x", op.Id())
	tw.add <- op
}

// delFromScheduler deletes an operation safely from the store.
func (tw *scheduler) delFromScheduler(op operationKey) {
	tw.logger.Debugf("de-scheduling operation: %x", op)
	tw.del <- op
}

func (tw *scheduler) setSchedulerBusy() {
	tw.logger.Debugf("setting scheduler busy")
	tw.mu.Lock()
	tw.busy = true
	tw.mu.Unlock()
}

func (tw *scheduler) setSchedulerFree() {
	tw.logger.Debugf("setting scheduler free")
	tw.mu.Lock()
	tw.busy = false
	tw.mu.Unlock()
}

func (tw *scheduler) isSchedulerBusy() bool {
	return tw.busy
}

// dumpOperationStore dumps to the logger and to the log file the current scheduled unexecuted operations.
// maps in go don't guarantee order, so that's why we have to find the earliest block.
func (tw *scheduler) dumpOperationStore(now func() time.Time) {
	if len(tw.store) <= 0 {
		tw.logger.Info("no operations to dump")
		return
	}

	f, err := os.Create(logPath + logFile)
	if err != nil {
		tw.logger.Fatalf("unable to create %s: %s", logPath+logFile, err.Error())
	}
	defer f.Close()

	tw.logger.Infof("generating logs with pending operations in %s", logPath+logFile)

	// Get the earliest block from all the operations stored by sorting them.
	blocks := make([]*big.Int, 0)
	for _, op := range tw.store {
		blocks = append(blocks, op[0].BlockNumber())
	}
	slices.SortFunc(blocks, func(a, b *big.Int) int { return a.Cmp(b) })

	w := bufio.NewWriter(f)

	writeOperationStore(w, tw.logger, tw.store, blocks[0], now)

	w.Flush()
}

type storeRecord struct {
	Block *big.Int
	OpKey operationKey
	Ops   []TimelockCallScheduled
}

// writeOperationStore writes the operations to the writer.
func writeOperationStore(
	w io.Writer,
	logger *zap.SugaredLogger,
	store map[operationKey][]TimelockCallScheduled,
	earliest *big.Int,
	now func() time.Time,
) {
	var (
		err error
		op  TimelockCallScheduled
		msg string
	)

	_, err = fmt.Fprintf(w, "Process stopped at %v\n", now().In(time.UTC))
	if err != nil {
		logger.Fatalf("error writing to buffer: %s", err.Error())
	}

	// order the store records by block number
	storeRecords := make([]storeRecord, 0)
	for opID, ops := range store {
		if len(ops) <= 0 {
			continue
		}
		storeRecords = append(storeRecords, storeRecord{
			Block: ops[0].BlockNumber(),
			OpKey: opID,
			Ops:   ops,
		})
	}
	slices.SortFunc(storeRecords, func(a, b storeRecord) int { return a.Block.Cmp(b.Block) })

	for _, record := range storeRecords {
		op = record.Ops[0]

		if op.BlockNumber().Cmp(earliest) == 0 {
			logLine := fmt.Sprintf("earliest unexecuted CallSchedule. Use this block number when "+
				"spinning up the service again, with the environment variable or in timelock.env as FROM_BLOCK=%v, "+
				"or using the flag --from-block=%v", op.BlockNumber(), op.BlockNumber())
			logger.With(fieldTXHash, op.TxHash()).With(fieldBlockNumber, op.BlockNumber()).Info(logLine)
			msg = toEarliestRecord(op)
		} else {
			logger.With(fieldTXHash, op.TxHash()).With(fieldBlockNumber, op.BlockNumber()).Info("CallSchedule pending")
			msg = toSubsequentRecord(op)
		}

		_, err = fmt.Fprint(w, msg)
		if err != nil {
			logger.Fatalf("error writing to buffer: %s", err.Error())
		}
	}
}

// toEarliestRecord returns a string with the earliest record.
func toEarliestRecord(op TimelockCallScheduled) string {
	tmpl := "Earliest CallSchedule pending ID: %x\tBlock Number: %v\n" +
		"\tUse this block number to ensure all pending operations are properly executed.  " +
		"\tSet it as environment variable or in timelock.env with FROM_BLOCK=%v, or as a flag with --from-block=%v\n"

	return fmt.Sprintf(tmpl, op.Id(), op.BlockNumber(), op.BlockNumber(), op.BlockNumber())
}

// toSubsequentRecord returns a string for use with each subsequent record sent to a writer.
func toSubsequentRecord(op TimelockCallScheduled) string {
	return fmt.Sprintf("CallSchedule pending ID: %x\tBlock Number: %v\n", op.Id(), op.BlockNumber())
}

// ----- nop scheduler -----
// nopScheduler implements the Scheduler interface but doesn't not effectively trigger any operations.
type nopScheduler struct {
	logger *zap.SugaredLogger
}

func newNopScheduler(logger *zap.SugaredLogger) *nopScheduler {
	return &nopScheduler{logger: logger}
}

func (s *nopScheduler) runScheduler(ctx context.Context) <-chan struct{} {
	s.logger.Info("nop.runScheduler")
	ch := make(chan struct{})

	go func() {
		<-ctx.Done()
		close(ch)
	}()

	return ch
}

func (s *nopScheduler) addToScheduler(op TimelockCallScheduled) {
	s.logger.With("op", op).Info("nop.addToScheduler")
}

func (s *nopScheduler) delFromScheduler(key operationKey) {
	s.logger.With("key", key).Info("nop.delFromScheduler")
}

func (s *nopScheduler) dumpOperationStore(now func() time.Time) {
	s.logger.Info("nop.dumpOperationStore")
}

package timelock

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"maps"
	"os"
	"slices"
	"sort"
	"sync"
	"time"

	"github.com/rs/zerolog"

	"github.com/smartcontractkit/timelock-worker/pkg/timelock/contract"
)

type operationKey [32]byte

type Scheduler interface {
	runScheduler(ctx context.Context) <-chan struct{}
	addToScheduler(op *contract.TimelockCallScheduled)
	delFromScheduler(op operationKey)
	dumpOperationStore(now func() time.Time)
}

type executeFn func(context.Context, []*contract.TimelockCallScheduled)

// Scheduler represents a scheduler with an in memory store.
// Whenever accesing the map the mutex should be Locked, to prevent
// any race condition.
type scheduler struct {
	mu        sync.Mutex
	ticker    *time.Ticker
	add       chan *contract.TimelockCallScheduled
	del       chan operationKey
	store     map[operationKey][]*contract.TimelockCallScheduled
	busy      bool
	logger    *zerolog.Logger
	executeFn executeFn
}

// newScheduler returns a new initialized scheduler.
func newScheduler(tick time.Duration, logger *zerolog.Logger, executeFn executeFn) *scheduler {
	s := &scheduler{
		ticker:    time.NewTicker(tick),
		add:       make(chan *contract.TimelockCallScheduled),
		del:       make(chan operationKey),
		store:     make(map[operationKey][]*contract.TimelockCallScheduled),
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
func (s *scheduler) runScheduler(ctx context.Context) <-chan struct{} {
	done := make(chan struct{})

	go func() {
		defer close(done)
		for {
			select {
			case <-s.ticker.C:
				s.executeOperations(ctx)

			case op := <-s.add:
				s.addOperation(op)

			case opKey := <-s.del:
				s.delOperation(opKey)

			case <-ctx.Done():
				s.logger.Debug().Msgf("shutting down scheduler")
				return
			}
		}
	}()

	return done
}

func (s *scheduler) executeOperations(ctx context.Context) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if len(s.store) <= 0 {
		s.logger.Debug().Msgf("scheduler.executingOperations: no operations in store")
		return
	}

	if s.busy {
		s.logger.Debug().Msgf("scheduler.executeOperations: scheduler is busy, skipping until next tick")
		return
	}
	s.busy = true

	store := maps.Clone(s.store)
	go func() {
		s.logger.Debug().Msgf("scheduler.executeOperations: %d operations in store", len(store))
		for _, op := range store {
			s.executeFn(ctx, op)
		}

		s.mu.Lock()
		defer s.mu.Unlock()
		s.busy = false
	}()
}

func (s *scheduler) addOperation(op *contract.TimelockCallScheduled) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if len(s.store[op.Id]) <= int(op.Index.Int64()) {
		s.store[op.Id] = append(s.store[op.Id], op)
	}
	s.store[op.Id][op.Index.Int64()] = op
	s.logger.Debug().Msgf("scheduled operation: %x", op.Id)
}

func (s *scheduler) delOperation(key operationKey) {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, ok := s.store[key]
	if !ok {
		s.logger.Warn().Msgf("operation not found in scheduler: %x", key)
		return
	}

	delete(s.store, key)
	s.logger.Debug().Msgf("de-scheduled operation: %x", key)
}

// updateSchedulerDelay updates the internal ticker delay, so it can be reconfigured while running.
func (s *scheduler) updateSchedulerDelay(t time.Duration) {
	if t <= 0 {
		s.logger.Debug().Msgf("internal min delay not changed, invalid duration: %v", t.String())
		return
	}

	s.ticker.Reset(t)
	s.logger.Debug().Msgf("internal min delay changed to %v", t.String())
}

// addToScheduler adds a new CallSchedule operation safely to the store.
func (s *scheduler) addToScheduler(op *contract.TimelockCallScheduled) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.logger.Debug().Msgf("scheduling operation: %x", op.Id)
	s.add <- op
	s.logger.Debug().Msgf("operations in scheduler: %v", len(s.store))
}

// delFromScheduler deletes an operation safely from the store.
func (s *scheduler) delFromScheduler(op operationKey) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.logger.Debug().Msgf("de-scheduling operation: %x", op)
	s.del <- op
	s.logger.Debug().Msgf("operations in scheduler: %v", len(s.store))
}

// dumpOperationStore dumps to the logger and to the log file the current scheduled unexecuted operations.
// maps in go don't guarantee order, so that's why we have to find the earliest block.
func (s *scheduler) dumpOperationStore(now func() time.Time) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if len(s.store) <= 0 {
		s.logger.Info().Msgf("no operations to dump")
		return
	}

	f, err := os.Create(logPath + logFile)
	if err != nil {
		s.logger.Fatal().Msgf("unable to create %s: %s", logPath+logFile, err.Error())
	}
	defer f.Close()

	s.logger.Info().Msgf("generating logs with pending operations in %s", logPath+logFile)

	// Get the earliest block from all the operations stored by sorting them.
	blocks := make([]uint64, 0)
	for _, op := range s.store {
		blocks = append(blocks, op[0].Raw.BlockNumber)
	}
	slices.Sort(blocks)

	w := bufio.NewWriter(f)

	writeOperationStore(w, s.logger, s.store, blocks[0], now)

	w.Flush()
}

type storeRecord struct {
	Block uint64
	OpKey operationKey
	Ops   []*contract.TimelockCallScheduled
}

// writeOperationStore writes the operations to the writer.
func writeOperationStore(
	w io.Writer,
	logger *zerolog.Logger,
	store map[operationKey][]*contract.TimelockCallScheduled,
	earliest uint64,
	now func() time.Time,
) {
	var (
		err error
		op  *contract.TimelockCallScheduled
		msg string
	)

	_, err = fmt.Fprintf(w, "Process stopped at %v\n", now().In(time.UTC))
	if err != nil {
		logger.Fatal().Msgf("error writing to buffer: %s", err.Error())
	}

	// order the store records by block number
	storeRecords := make([]storeRecord, 0)
	for opID, ops := range store {
		if len(ops) <= 0 {
			continue
		}
		storeRecords = append(storeRecords, storeRecord{
			Block: ops[0].Raw.BlockNumber,
			OpKey: opID,
			Ops:   ops,
		})
	}
	sort.Slice(storeRecords, func(i, j int) bool {
		return storeRecords[i].Block < storeRecords[j].Block
	})

	for _, record := range storeRecords {
		op = record.Ops[0]

		if op.Raw.BlockNumber == earliest {
			logLine := fmt.Sprintf("earliest unexecuted CallSchedule. Use this block number when "+
				"spinning up the service again, with the environment variable or in timelock.env as FROM_BLOCK=%v, "+
				"or using the flag --from-block=%v", op.Raw.BlockNumber, op.Raw.BlockNumber)
			logger.Info().Hex(fieldTXHash, op.Raw.TxHash[:]).Uint64(fieldBlockNumber, op.Raw.BlockNumber).Msg(logLine)
			msg = toEarliestRecord(op)
		} else {
			logger.Info().Hex(fieldTXHash, op.Raw.TxHash[:]).Uint64(fieldBlockNumber, op.Raw.BlockNumber).Msgf("CallSchedule pending")
			msg = toSubsequentRecord(op)
		}

		_, err = fmt.Fprint(w, msg)
		if err != nil {
			logger.Fatal().Msgf("error writing to buffer: %s", err.Error())
		}
	}
}

// toEarliestRecord returns a string with the earliest record.
func toEarliestRecord(op *contract.TimelockCallScheduled) string {
	tmpl := "Earliest CallSchedule pending ID: %x\tBlock Number: %v\n" +
		"\tUse this block number to ensure all pending operations are properly executed.  " +
		"\tSet it as environment variable or in timelock.env with FROM_BLOCK=%v, or as a flag with --from-block=%v\n"

	return fmt.Sprintf(tmpl, op.Id, op.Raw.BlockNumber, op.Raw.BlockNumber, op.Raw.BlockNumber)
}

// toSubsequentRecord returns a string for use with each subsequent record sent to a writer.
func toSubsequentRecord(op *contract.TimelockCallScheduled) string {
	return fmt.Sprintf("CallSchedule pending ID: %x\tBlock Number: %v\n", op.Id, op.Raw.BlockNumber)
}

// ----- nop scheduler -----
// nopScheduler implements the Scheduler interface but doesn't not effectively trigger any operations.
type nopScheduler struct {
	logger *zerolog.Logger
}

func newNopScheduler(logger *zerolog.Logger) *nopScheduler {
	return &nopScheduler{logger: logger}
}

func (s *nopScheduler) runScheduler(ctx context.Context) <-chan struct{} {
	s.logger.Info().Msg("nop.runScheduler")
	ch := make(chan struct{})

	go func() {
		<-ctx.Done()
		close(ch)
	}()

	return ch
}

func (s *nopScheduler) addToScheduler(op *contract.TimelockCallScheduled) {
	s.logger.Info().Any("op", op).Msg("nop.addToScheduler")
}

func (s *nopScheduler) delFromScheduler(key operationKey) {
	s.logger.Info().Any("key", key).Msg("nop.delFromScheduler")
}

func (s *nopScheduler) dumpOperationStore(now func() time.Time) {
	s.logger.Info().Msg("nop.dumpOperationStore")
}

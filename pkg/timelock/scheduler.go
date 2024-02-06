package timelock

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"sort"
	"sync"
	"time"

	"github.com/rs/zerolog"
	"github.com/smartcontractkit/timelock-worker/pkg/timelock/contract"
)

type operationKey [32]byte

// Scheduler represents a scheduler with an in memory store.
// Whenever accesing the map the mutex should be Locked, to prevent
// any race condition.
type scheduler struct {
	mu     sync.Mutex
	ticker *time.Ticker
	add    chan *contract.TimelockCallScheduled
	del    chan operationKey
	store  map[operationKey][]*contract.TimelockCallScheduled
	busy   bool
}

// newScheduler returns a new initialized scheduler.
func newScheduler(tick time.Duration) *scheduler {
	s := &scheduler{
		ticker: time.NewTicker(tick),
		add:    make(chan *contract.TimelockCallScheduled),
		del:    make(chan operationKey),
		store:  make(map[operationKey][]*contract.TimelockCallScheduled),
		busy:   false,
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
func (tw *Worker) runScheduler(ctx context.Context) {
	for {
		select {
		case <-tw.ticker.C:
			if len(tw.store) <= 0 {
				tw.logger.Debug().Msgf("new scheduler tick: no operations in store")
				continue
			}

			if !tw.isSchedulerBusy() {
				tw.logger.Debug().Msgf("new scheduler tick: operations in store")
				tw.setSchedulerBusy()
				for _, op := range tw.store {
					tw.execute(ctx, op)
				}
				tw.setSchedulerFree()
			} else {
				tw.logger.Debug().Msgf("new scheduler tick: scheduler is busy, skipping until next tick")
			}

		case op := <-tw.add:
			tw.mu.Lock()
			if len(tw.store[op.Id]) <= int(op.Index.Int64()) {
				tw.store[op.Id] = append(tw.store[op.Id], op)
			}
			tw.store[op.Id][op.Index.Int64()] = op
			tw.mu.Unlock()
			tw.logger.Debug().Msgf("scheduled operation: %x", op.Id)

		case op := <-tw.del:
			if _, ok := tw.store[op]; ok {
				tw.mu.Lock()
				delete(tw.store, op)
				tw.mu.Unlock()
				tw.logger.Debug().Msgf("de-scheduled operation: %x", op)
			}
		}
	}
}

// updateSchedulerDelay updates the internal ticker delay, so it can be reconfigured while running.
func (tw *Worker) updateSchedulerDelay(t time.Duration) {
	if t > 0 {
		tw.ticker.Reset(t)
		tw.logger.Debug().Msgf("internal min delay changed to %v", t.String())
	} else {
		tw.logger.Debug().Msgf("internal min delay not changed, invalid duration: %v", t.String())
	}
}

// addToScheduler adds a new CallSchedule operation safely to the store.
func (tw *Worker) addToScheduler(op *contract.TimelockCallScheduled) {
	tw.logger.Debug().Msgf("scheduling operation: %x", op.Id)
	tw.add <- op
	tw.logger.Debug().Msgf("operations in scheduler: %v", len(tw.store))
}

// delFromScheduler deletes an operation safely from the store.
func (tw *Worker) delFromScheduler(op operationKey) {
	tw.logger.Debug().Msgf("de-scheduling operation: %v", op)
	tw.del <- op
	tw.logger.Debug().Msgf("operations in scheduler: %v", len(tw.store))
}

func (tw *Worker) setSchedulerBusy() {
	tw.logger.Debug().Msgf("setting scheduler busy")
	tw.mu.Lock()
	tw.busy = true
	tw.mu.Unlock()
}

func (tw *Worker) setSchedulerFree() {
	tw.logger.Debug().Msgf("setting scheduler free")
	tw.mu.Lock()
	tw.busy = false
	tw.mu.Unlock()
}

func (tw *Worker) isSchedulerBusy() bool {
	return tw.busy
}

// dumpOperationStore dumps to the logger and to the log file the current scheduled unexecuted operations.
// maps in go don't guarantee order, so that's why we have to find the earliest block.
func (tw *Worker) dumpOperationStore(now func() time.Time) {
	if len(tw.store) > 0 {
		f, err := os.Create(logPath + logFile)
		if err != nil {
			tw.logger.Fatal().Msgf("unable to create %s: %s", logPath+logFile, err.Error())
		}
		defer f.Close()

		tw.logger.Info().Msgf("generating logs with pending operations in %s", logPath+logFile)

		// Get the earliest block from all the operations stored by sorting them.
		blocks := make([]int, 0)
		for _, op := range tw.store {
			blocks = append(blocks, int(op[0].Raw.BlockNumber))
		}
		sort.Ints(blocks)

		w := bufio.NewWriter(f)

		writeOperationStore(w, tw.logger, tw.store, blocks[0], now)

		w.Flush()
	}
}

// writeOperationStore writes the operations to the writer.
func writeOperationStore(
	w io.Writer,
	logger *zerolog.Logger,
	store map[operationKey][]*contract.TimelockCallScheduled,
	earliest int,
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

	for _, record := range store {
		op = record[0]

		if int(op.Raw.BlockNumber) == earliest {
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

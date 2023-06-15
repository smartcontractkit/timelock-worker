package timelock

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"sort"
	"sync"
	"time"

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
	tw.logger.Debug().Msgf("de-scheduling operation: %x", op)
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
func (tw *Worker) dumpOperationStore() {
	if len(tw.store) > 0 {
		f, err := os.Create(timelockLogPath + timelockLogFile)
		if err != nil {
			tw.logger.Fatal().Msgf("unable to create %s: %s", timelockLogPath+timelockLogFile, err.Error())
		}
		defer f.Close()

		tw.logger.Info().Msgf("generating logs with pending operations in %s", timelockLogPath+timelockLogFile)

		w := bufio.NewWriter(f)
		_, err = fmt.Fprintf(w, "Process stopped at %v\n", time.Now().In(time.UTC))
		if err != nil {
			tw.logger.Fatal().Msgf("error writing to buffer: %s", err.Error())
		}

		// Get the earliest block from all the operations stored by sorting them.
		blocks := make([]int, 0)
		for _, op := range tw.store {
			blocks = append(blocks, int(op[0].Raw.BlockNumber))
		}
		sort.Ints(blocks)

		for _, op := range tw.store {
			if int(op[0].Raw.BlockNumber) == blocks[0] {
				tw.logger.Info().Hex(fieldTXHash, op[0].Raw.TxHash[:]).Uint64(fieldBlockNumber, op[0].Raw.BlockNumber).Msgf("earliest unexecuted CallSchedule. Use this block number when spinning up the service again, with the environment variable or in timelock.env as FROM_BLOCK=%v, or using the flag --from-block=%v", op[0].Raw.BlockNumber, op[0].Raw.BlockNumber)
				_, err = fmt.Fprintf(w, "Earliest CallSchedule pending ID: %x\tBlock Number: %v\n\tUse this block number to ensure all pending operations are properly executed.\n\tSet it as environment variable or in timelock.env with FROM_BLOCK=%v, or as a flag with --from-block=%v\n", op[0].Id, op[0].Raw.BlockNumber, op[0].Raw.BlockNumber, op[0].Raw.BlockNumber)
				if err != nil {
					tw.logger.Fatal().Msgf("error writing to buffer: %s", err.Error())
				}
			} else {
				_, err = fmt.Fprintf(w, "CallSchedule pending ID: %x\tBlock Number: %v\n", op[0].Id, op[0].Raw.BlockNumber)
				tw.logger.Info().Hex(fieldTXHash, op[0].Raw.TxHash[:]).Uint64(fieldBlockNumber, op[0].Raw.BlockNumber).Msgf("CallSchedule pending")
				if err != nil {
					tw.logger.Fatal().Msgf("error writing to buffer: %s", err.Error())
				}
			}
		}

		w.Flush()
	}
}

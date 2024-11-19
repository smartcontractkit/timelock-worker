package timelock

import (
	"context"
	"fmt"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/rs/zerolog"
	"github.com/smartcontractkit/timelock-worker/pkg/timelock/contract"
	"github.com/stretchr/testify/assert"
)

func Test_newScheduler(t *testing.T) {
	logger := zerolog.Nop()
	execFn := func(context.Context, []*contract.TimelockCallScheduled) {}
	tScheduler := newTestScheduler()

	type args struct {
		tick time.Duration
	}
	tests := []struct {
		name string
		args args
		want *scheduler
	}{
		{
			name: "New scheduler",
			args: args{
				tick: 10 * time.Second,
			},
			want: tScheduler,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := newScheduler(tt.args.tick, &logger, execFn)
			if reflect.TypeOf(got) != reflect.TypeOf(tt.want) {
				t.Errorf("newScheduler() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_scheduler_updateSchedulerDelay(t *testing.T) {
	tScheduler := newTestScheduler()

	// Should never fail
	tScheduler.updateSchedulerDelay(1 * time.Second)
	tScheduler.updateSchedulerDelay(-1 * time.Second)
	tScheduler.updateSchedulerDelay(0 * time.Second)
}

func Test_scheduler_isSchedulerBusy(t *testing.T) {
	tScheduler := newTestScheduler()

	isBusy := tScheduler.isSchedulerBusy()
	assert.Equal(t, false, isBusy, "scheduler should be busy by default")

	tScheduler.setSchedulerBusy()
	isBusy = tScheduler.isSchedulerBusy()
	assert.Equal(t, true, isBusy, "scheduler should be busy after setSchedulerBusy()")

	tScheduler.setSchedulerFree()
	isBusy = tScheduler.isSchedulerBusy()
	assert.Equal(t, false, isBusy, "scheduler shouldn't be busy after setSchedulerFree()")
}

func Test_scheduler_setSchedulerBusy(t *testing.T) {
	tScheduler := newTestScheduler()

	tScheduler.setSchedulerBusy()
	isBusy := tScheduler.isSchedulerBusy()
	assert.Equal(t, true, isBusy, "scheduler should be busy after setSchedulerBusy()")
}

func Test_scheduler_setSchedulerFree(t *testing.T) {
	logger := zerolog.Nop()
	execFn := func(context.Context, []*contract.TimelockCallScheduled) {}
	tScheduler := newScheduler(10 * time.Second, &logger, execFn)

	tScheduler.setSchedulerFree()
	isBusy := tScheduler.isSchedulerBusy()
	assert.Equal(t, false, isBusy, "scheduler shouldn't be busy after setSchedulerFree()")
}

// Test_dumpOperationStore tests the dumpOperationStore method and ensures that it writes the correct contents to the file.
func Test_dumpOperationStore(t *testing.T) {
	var (
		fName         = logPath + logFile
		earliestBlock = 42
		opKeys        = generateOpKeys(t, []string{"1", "2"})

		earliest = &contract.TimelockCallScheduled{
			Id: opKeys[0],
			Raw: types.Log{
				TxHash:      common.HexToHash("txn-1"),
				BlockNumber: uint64(earliestBlock),
			},
		}

		following = &contract.TimelockCallScheduled{
			Id: opKeys[1],
			Raw: types.Log{
				TxHash:      common.HexToHash("txn-2"),
				BlockNumber: uint64(earliestBlock + 1),
			},
		}

		store = map[operationKey][]*contract.TimelockCallScheduled{
			opKeys[0]: {earliest},
			opKeys[1]: {following},
		}

		scheduler = scheduler{
			store: store,
			logger: func(l zerolog.Logger) *zerolog.Logger { return &l }(zerolog.Nop()),
		}
	)

	defer os.Remove(fName)

	// setup fake time
	dateString := "2021-11-22"
	date, err := time.Parse("2006-01-02", dateString)
	assert.NoError(t, err)

	nowFunc := func() time.Time {
		return date
	}

	wantPrefix := fmt.Sprintf("Process stopped at %v\n", nowFunc().In(time.UTC))

	// Write the store to the file.
	scheduler.dumpOperationStore(nowFunc)

	// Read the file and compare the contents.
	gotRead, err := os.ReadFile(fName)
	assert.NoError(t, err)

	// Assert that the contents of the file match the expected contents.
	var wantRead []byte
	wantRead = append(wantRead, []byte(wantPrefix)...)
	wantRead = append(wantRead, []byte(toEarliestRecord(earliest))...)
	wantRead = append(wantRead, []byte(toSubsequentRecord(following))...)
	assert.Equal(t, wantRead, gotRead)
}

// ----- helpers -----

func newTestScheduler() *scheduler {
	logger := zerolog.Nop()
	execFn := func(context.Context, []*contract.TimelockCallScheduled) {}
	return newScheduler(10 * time.Second, &logger, execFn)
}

// generateOpKeys generates a slice of operation keys from a slice of strings.
func generateOpKeys(t *testing.T, in []string) [][32]byte {
	t.Helper()

	opKeys := make([][32]byte, 0, len(in))
	for _, id := range in {
		padding := strings.Repeat("0", 32-len(id))
		padded := fmt.Sprintf("%s%s", padding, id)
		var key [32]byte
		copy(key[:], padded)
		opKeys = append(opKeys, key)
	}
	return opKeys
}

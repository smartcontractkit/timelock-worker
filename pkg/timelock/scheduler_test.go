package timelock

import (
	"context"
	"fmt"
	"math/big"
	"math/rand"
	"os"
	"reflect"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/google/go-cmp/cmp"
	"github.com/samber/lo"
	contracts "github.com/smartcontractkit/ccip-owner-contracts/gethwrappers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func Test_newScheduler(t *testing.T) {
	logger := zap.NewNop().Sugar()
	execFn := func(context.Context, []*contracts.RBACTimelockCallScheduled) {}
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
			got := newScheduler(tt.args.tick, logger, execFn)
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
	logger := zap.NewNop().Sugar()
	execFn := func(context.Context, []*contracts.RBACTimelockCallScheduled) {}
	tScheduler := newScheduler(10*time.Second, logger, execFn)

	tScheduler.setSchedulerFree()
	isBusy := tScheduler.isSchedulerBusy()
	assert.Equal(t, false, isBusy, "scheduler shouldn't be busy after setSchedulerFree()")
}

// Test_dumpOperationStore tests the dumpOperationStore method and ensures that it writes the correct contents to the file.
func Test_dumpOperationStore(t *testing.T) {
	var (
		fName         = logPath + logFile
		logger        = zap.NewNop().Sugar()
		earliestBlock = 42
		opKeys        = generateOpKeys(t, []string{"1", "2"})

		earliest = &contracts.RBACTimelockCallScheduled{
			Id: opKeys[0],
			Raw: types.Log{
				TxHash:      common.HexToHash("txn-1"),
				BlockNumber: uint64(earliestBlock),
			},
		}

		following = &contracts.RBACTimelockCallScheduled{
			Id: opKeys[1],
			Raw: types.Log{
				TxHash:      common.HexToHash("txn-2"),
				BlockNumber: uint64(earliestBlock + 1),
			},
		}

		store = map[operationKey][]*contracts.RBACTimelockCallScheduled{
			opKeys[0]: {earliest},
			opKeys[1]: {following},
		}

		scheduler = scheduler{
			store:  store,
			logger: logger,
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

func Test_scheduler_concurrency(t *testing.T) {
	const numOps = 100
	logger := zap.NewNop().Sugar()
	ctx, cancel := context.WithCancel(context.Background())

	executedOps := map[int]uint16{} // {numericOpId: executionCount}
	executedCh := make(chan operationKey)
	execFn := func(ctx context.Context, ops []*contracts.RBACTimelockCallScheduled) {
		for _, op := range ops {
			opNum := int(opIDToNum(t, op.Id))
			executedOps[opNum] = executedOps[opNum] + 1
			go func() {
				time.Sleep(time.Duration(1+rand.Intn(50)) * time.Millisecond)
				executedCh <- op.Id
			}()
		}
	}

	// run scheduler
	testScheduler := newScheduler(10*time.Millisecond, logger, execFn)
	_ = testScheduler.runScheduler(ctx)

	// run mock event listener
	go runMockEventListener(t, ctx, cancel, testScheduler, executedCh, numOps)

	// wait for all operations to be executed
	<-ctx.Done()

	require.GreaterOrEqual(t, len(executedOps), numOps)
	executedIDs := lo.Keys(executedOps)
	slices.Sort(executedIDs)
	require.Empty(t, cmp.Diff(lo.Range(100), executedIDs[:numOps]))
}

// ----- helpers -----

func newTestScheduler() *scheduler {
	logger := zap.NewNop().Sugar()
	execFn := func(context.Context, []*contracts.RBACTimelockCallScheduled) {}
	return newScheduler(10*time.Second, logger, execFn)
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

func runMockEventListener(
	t *testing.T,
	ctx context.Context,
	cancel context.CancelFunc,
	testScheduler *scheduler,
	executedCh <-chan operationKey,
	lastOpID int16,
) {
	t.Helper()

	opNum := int64(0)

	ticker := time.NewTicker(15 * time.Millisecond)
	for {
		select {
		case <-ticker.C:
			op := &contracts.RBACTimelockCallScheduled{Id: opID(uint16(opNum)), Index: big.NewInt(0)}
			opNum += 1
			testScheduler.addToScheduler(op)

		case executedOpID := <-executedCh:
			testScheduler.delFromScheduler(executedOpID)
			if opIDToNum(t, executedOpID) == lastOpID {
				cancel()
			}

		case <-ctx.Done():
			return
		}
	}
}

func opID(n uint16) [32]byte {
	id := [32]byte{}
	id[31] = byte(n)
	id[30] = byte(n >> 8)
	return id
}

func opIDToNum(t *testing.T, opID [32]byte) int16 {
	t.Helper()
	opNum, ok := new(big.Int).SetString(fmt.Sprintf("%x", opID), 16)
	require.True(t, ok)
	return int16(opNum.Uint64())
}

package timelock

import (
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func Test_newScheduler(t *testing.T) {
	tScheduler := newScheduler(10 * time.Second)

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
			got := newScheduler(tt.args.tick)
			if reflect.TypeOf(got) != reflect.TypeOf(tt.want) {
				t.Errorf("newScheduler() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestWorker_updateSchedulerDelay(t *testing.T) {
	testWorker, _ := NewTimelockWorker(testNodeURL, testTimelockAddress, testCallProxyAddress, testPrivateKey, testFromBlock, int64(testPollPeriod), testLogger)

	// Should never fail
	testWorker.updateSchedulerDelay(1 * time.Second)
	testWorker.updateSchedulerDelay(-1 * time.Second)
	testWorker.updateSchedulerDelay(0 * time.Second)
}

func TestWorker_isSchedulerBusy(t *testing.T) {
	testWorker, _ := NewTimelockWorker(testNodeURL, testTimelockAddress, testCallProxyAddress, testPrivateKey, testFromBlock, int64(testPollPeriod), testLogger)

	isBusy := testWorker.isSchedulerBusy()
	assert.Equal(t, false, isBusy, "scheduler should be busy by default")

	testWorker.setSchedulerBusy()
	isBusy = testWorker.isSchedulerBusy()
	assert.Equal(t, true, isBusy, "scheduler should be busy after setSchedulerBusy()")

	testWorker.setSchedulerFree()
	isBusy = testWorker.isSchedulerBusy()
	assert.Equal(t, false, isBusy, "scheduler shouldn't be busy after setSchedulerFree()")
}

func TestWorker_setSchedulerBusy(t *testing.T) {
	testWorker, _ := NewTimelockWorker(testNodeURL, testTimelockAddress, testCallProxyAddress, testPrivateKey, testFromBlock, int64(testPollPeriod), testLogger)

	testWorker.setSchedulerBusy()
	isBusy := testWorker.isSchedulerBusy()
	assert.Equal(t, true, isBusy, "scheduler should be busy after setSchedulerBusy()")
}

func TestWorker_setSchedulerFree(t *testing.T) {
	testWorker, _ := NewTimelockWorker(testNodeURL, testTimelockAddress, testCallProxyAddress, testPrivateKey, testFromBlock, int64(testPollPeriod), testLogger)

	testWorker.setSchedulerFree()
	isBusy := testWorker.isSchedulerBusy()
	assert.Equal(t, false, isBusy, "scheduler shouldn't be busy after setSchedulerFree()")
}

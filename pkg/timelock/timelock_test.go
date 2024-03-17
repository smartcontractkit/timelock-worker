package timelock

import (
	"math/big"
	"os"
	"reflect"
	"sync"
	"syscall"
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
)

func TestNewTimelockWorker(t *testing.T) {
	testWorker, _ := NewTimelockWorker(testNodeURL, testTimelockAddress, testCallProxyAddress, testPrivateKey, testFromBlock, int64(testPollPeriod), testLogger)

	if testWorker == nil {
		t.Skipf("testWorker is nil, are the env variables in const_test.go set?")
	}

	type args struct {
		nodeURL          string
		timelockAddress  string
		callProxyAddress string
		privateKey       string
		fromBlock        *big.Int
		pollPeriod       int64
		logger           *zerolog.Logger
	}
	tests := []struct {
		name    string
		args    args
		want    *Worker
		wantErr bool
	}{
		{
			name: "NewTimelockWorker new instance is created (success)",
			args: args{
				nodeURL:          testNodeURL,
				timelockAddress:  testTimelockAddress,
				callProxyAddress: testCallProxyAddress,
				privateKey:       testPrivateKey,
				fromBlock:        testFromBlock,
				pollPeriod:       int64(testPollPeriod),
				logger:           testLogger,
			},
			want:    testWorker,
			wantErr: false,
		},
		{
			name: "NewTimelockWorker bad rpc provided (fail)",
			args: args{
				nodeURL:          "wss://bad/rpc",
				timelockAddress:  testTimelockAddress,
				callProxyAddress: testCallProxyAddress,
				privateKey:       testPrivateKey,
				fromBlock:        testFromBlock,
				pollPeriod:       int64(testPollPeriod),
				logger:           testLogger,
			},
			want:    testWorker,
			wantErr: true,
		},
		{
			name: "NewTimelockWorker bad rpc protocol provided (fail)",
			args: args{
				nodeURL:          "https://bad/protocol",
				timelockAddress:  testTimelockAddress,
				callProxyAddress: testCallProxyAddress,
				privateKey:       testPrivateKey,
				fromBlock:        testFromBlock,
				pollPeriod:       int64(testPollPeriod),
				logger:           testLogger,
			},
			want:    testWorker,
			wantErr: true,
		},
		{
			name: "NewTimelockWorker bad timelock address provided (fail)",
			args: args{
				nodeURL:          testNodeURL,
				timelockAddress:  "0x1234",
				callProxyAddress: testCallProxyAddress,
				privateKey:       testPrivateKey,
				fromBlock:        testFromBlock,
				pollPeriod:       int64(testPollPeriod),
				logger:           testLogger,
			},
			want:    testWorker,
			wantErr: true,
		},
		{
			name: "NewTimelockWorker bad call proxy address provided (fail)",
			args: args{
				nodeURL:          testNodeURL,
				timelockAddress:  testTimelockAddress,
				callProxyAddress: "0x1234",
				privateKey:       testPrivateKey,
				fromBlock:        testFromBlock,
				pollPeriod:       int64(testPollPeriod),
				logger:           testLogger,
			},
			want:    testWorker,
			wantErr: true,
		},
		{
			name: "NewTimelockWorker bad private key provided (fail)",
			args: args{
				nodeURL:          testNodeURL,
				timelockAddress:  testTimelockAddress,
				callProxyAddress: testCallProxyAddress,
				privateKey:       "0123456789",
				fromBlock:        testFromBlock,
				pollPeriod:       int64(testPollPeriod),
				logger:           testLogger,
			},
			want:    testWorker,
			wantErr: true,
		},
		{
			name: "NewTimelockWorker bad negative from block provided (fail)",
			args: args{
				nodeURL:          testNodeURL,
				timelockAddress:  testTimelockAddress,
				callProxyAddress: testCallProxyAddress,
				privateKey:       testPrivateKey,
				fromBlock:        big.NewInt(-1),
				pollPeriod:       int64(testPollPeriod),
				logger:           testLogger,
			},
			want:    testWorker,
			wantErr: true,
		},
		{
			name: "NewTimelockWorker bad poll period provided (fail)",
			args: args{
				nodeURL:          testNodeURL,
				timelockAddress:  testTimelockAddress,
				callProxyAddress: testCallProxyAddress,
				privateKey:       testPrivateKey,
				fromBlock:        testFromBlock,
				pollPeriod:       0,
				logger:           testLogger,
			},
			want:    testWorker,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewTimelockWorker(tt.args.nodeURL, tt.args.timelockAddress, tt.args.callProxyAddress, tt.args.privateKey, tt.args.fromBlock, tt.args.pollPeriod, tt.args.logger)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewTimelockWorker() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !(reflect.TypeOf(tt.want) == reflect.TypeOf(got)) {
				t.Errorf("NewTimelockWorker() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_handleOSSignal(t *testing.T) {
	stopCh := make(chan string)
	sigCh := make(chan os.Signal, 1)
	var testWg sync.WaitGroup

	testWg.Add(1)
	go func() {
		for {
			handleOSSignal(<-sigCh, stopCh)
		}
	}()

	sigCh <- syscall.SIGTERM
	assert.Equal(t, <-stopCh, "terminated", "send SIGTERM, receive terminated")

	sigCh <- syscall.SIGINT
	assert.Equal(t, <-stopCh, "interrupt", "send SIGINT, receive interrupt")
}

func TestWorker_startLog(t *testing.T) {
	testWorker, _ := NewTimelockWorker(testNodeURL, testTimelockAddress, testCallProxyAddress, testPrivateKey, testFromBlock, int64(testPollPeriod), testLogger)

	if testWorker == nil {
		t.Skipf("testWorker is nil, are the env variables in const_test.go set?")
	}

	tests := []struct {
		name string
	}{
		{
			"Show logs",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testWorker.startLog()
		})
	}
}

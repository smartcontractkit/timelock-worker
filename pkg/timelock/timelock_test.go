package timelock

import (
	"math/big"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestTimelockWorker(
	t *testing.T, nodeURL, timelockAddress, callProxyAddress, privateKey string, fromBlock *big.Int,
	pollPeriod int64, eventListenerPollPeriod int64, dryRun bool, logger *zerolog.Logger,
) *Worker {
	assert.NotEmpty(t, nodeURL, "nodeURL is empty. Are environment variabes in const_test.go set?")
	assert.NotEmpty(t, timelockAddress, "nodeURL is empty. Are environment variabes in const_test.go set?")
	assert.NotEmpty(t, callProxyAddress, "callProxyAddress is empty. Are environment variabes in const_test.go set?")
	assert.NotEmpty(t, privateKey, "privateKey is empty. Are environment variabes in const_test.go set?")
	assert.NotNil(t, fromBlock, "fromBlock is nil. Are environment variabes in const_test.go set?")
	assert.NotEmpty(t, pollPeriod, "pollPeriod is empty. Are environment variabes in const_test.go set?")
	assert.NotNil(t, logger, "logger is nil. Are environment variabes in const_test.go set?")

	tw, err := NewTimelockWorker(nodeURL, timelockAddress, callProxyAddress, privateKey, fromBlock,
		pollPeriod, eventListenerPollPeriod, dryRun, logger)
	assert.NoError(t, err)
	assert.NotNil(t, tw)

	return tw
}

func TestNewTimelockWorker(t *testing.T) {
	svr := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, req *http.Request) {
		writer.Write([]byte("Ok"))
	}))
	defer svr.Close()

	type argsT struct {
		nodeURL                 string
		timelockAddress         string
		callProxyAddress        string
		privateKey              string
		fromBlock               *big.Int
		pollPeriod              int64
		eventListenerPollPeriod int64
		dryRun                  bool
		logger                  zerolog.Logger
	}
	defaultArgs := argsT{
		nodeURL:                 svr.URL,
		timelockAddress:         "0x0000000000000000000000000000000000000001",
		callProxyAddress:        "0x0000000000000000000000000000000000000002",
		privateKey:              "1921763610b80b2b147ca676c775c172ba6c037f6ba135792bed6bc458c660f0",
		fromBlock:               big.NewInt(1),
		pollPeriod:              900,
		eventListenerPollPeriod: 60,
		dryRun:                  false,
		logger:                  zerolog.Nop(),
	}

	tests := []struct {
		name    string
		setup   func(*argsT)
		wantErr string
	}{
		{
			name:  "success",
			setup: func(*argsT) {},
		},
		{
			name:    "failure - invalid host in node url",
			setup:   func(a *argsT) { a.nodeURL = "wss://invalid.host/rpc" },
			wantErr: "no such host",
		},
		{
			name:    "failure - invalid url scheme in node url",
			setup:   func(a *argsT) { a.nodeURL = "invalid://localhost/rpc" },
			wantErr: "invalid node URL: invalid://localhost/rpc (accepted schemes are: [http https ws wss])",
		},
		{
			name:    "failure - bad timelock address",
			setup:   func(a *argsT) { a.timelockAddress = "invalid" },
			wantErr: "timelock address provided is not valid: invalid",
		},
		{
			name:    "failure - bad call proxy address",
			setup:   func(a *argsT) { a.callProxyAddress = "invalid" },
			wantErr: "call proxy address provided is not valid: invalid",
		},
		{
			name:    "failure - bad private key",
			setup:   func(a *argsT) { a.privateKey = "invalid" },
			wantErr: "the provided private key is not valid: got invalid",
		},
		{
			name:    "failure - bad from block",
			setup:   func(a *argsT) { a.fromBlock = big.NewInt(-1) },
			wantErr: "from block can't be a negative number (minimum value 0): got -1",
		},
		{
			name:    "failure - bad poll period",
			setup:   func(a *argsT) { a.pollPeriod = -1 },
			wantErr: "poll-period must be a positive non-zero integer: got -1",
		},
		{
			name:    "failure - bad event listener poll period",
			setup:   func(a *argsT) { a.eventListenerPollPeriod = -1 },
			wantErr: "event-listener-poll-period must be a positive non-zero integer: got -1",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			args := defaultArgs
			tt.setup(&args)

			got, err := NewTimelockWorker(args.nodeURL, args.timelockAddress, args.callProxyAddress,
				args.privateKey, args.fromBlock, args.pollPeriod, args.eventListenerPollPeriod,
				args.dryRun, &args.logger)

			if tt.wantErr == "" {
				require.NoError(t, err)
				require.IsType(t, &Worker{}, got)
			} else {
				require.ErrorContains(t, err, tt.wantErr)
			}
		})
	}
}

func TestWorker_startLog(t *testing.T) {
	testWorker := newTestTimelockWorker(t, testNodeURL, testTimelockAddress, testCallProxyAddress, testPrivateKey,
		testFromBlock, int64(testPollPeriod), int64(testEventListenerPollPeriod), testDryRun, testLogger)

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

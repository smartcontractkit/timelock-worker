// new_worker_solana_test.go
package timelock

import (
	"context"
	"math/big"
	"testing"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"
)

func TestNewTimelockWorkerSolana(t *testing.T) {
	// Prepare one valid keypair for tests
	priv, err := solana.NewRandomPrivateKey()
	require.NoError(t, err)
	privStr := priv.String()
	pubStr := priv.PublicKey().String()

	// Dummy callProxyAddress (unused in constructor)
	callProxy := pubStr
	// fromBlock is unused by NewTimelockWorkerSolana; just pass zero
	fromBlock := big.NewInt(0)

	logger := zap.NewNop().Sugar()

	tests := []struct {
		name                 string
		nodeURL              string
		timelockAddress      string
		callProxyAddress     string
		privateKey           string
		fromBlock            *big.Int
		pollPeriod           int64
		listenerPollPeriod   int64
		pollSize             uint64
		dryRun               bool
		wantErr              bool
		errContainsSubstring string
	}{
		{
			name:                 "invalid URL syntax",
			nodeURL:              "://not-a-url",
			timelockAddress:      pubStr,
			callProxyAddress:     callProxy,
			privateKey:           privStr,
			fromBlock:            fromBlock,
			pollPeriod:           1,
			listenerPollPeriod:   1,
			pollSize:             1,
			dryRun:               true,
			wantErr:              true,
			errContainsSubstring: "parse \"://not-a-url\"",
		},
		{
			name:                 "unsupported scheme",
			nodeURL:              "ftp://example.com",
			timelockAddress:      pubStr,
			callProxyAddress:     callProxy,
			privateKey:           privStr,
			fromBlock:            fromBlock,
			pollPeriod:           1,
			listenerPollPeriod:   1,
			pollSize:             1,
			dryRun:               false,
			wantErr:              true,
			errContainsSubstring: "invalid node URL",
		},
		{
			name:                 "invalid timelock address",
			nodeURL:              "http://example.com",
			timelockAddress:      "notBase58",
			callProxyAddress:     callProxy,
			privateKey:           privStr,
			fromBlock:            fromBlock,
			pollPeriod:           1,
			listenerPollPeriod:   1,
			pollSize:             1,
			dryRun:               false,
			wantErr:              true,
			errContainsSubstring: "timelock addresses provided is not valid",
		},
		{
			name:                 "negative pollPeriod",
			nodeURL:              "http://example.com",
			timelockAddress:      pubStr,
			callProxyAddress:     callProxy,
			privateKey:           privStr,
			fromBlock:            fromBlock,
			pollPeriod:           0,
			listenerPollPeriod:   1,
			pollSize:             1,
			dryRun:               false,
			wantErr:              true,
			errContainsSubstring: "poll-period must be a positive non-zero integer",
		},
		{
			name:                 "zero listenerPollPeriod on HTTP",
			nodeURL:              "https://example.com",
			timelockAddress:      pubStr,
			callProxyAddress:     callProxy,
			privateKey:           privStr,
			fromBlock:            fromBlock,
			pollPeriod:           1,
			listenerPollPeriod:   0,
			pollSize:             1,
			dryRun:               false,
			wantErr:              true,
			errContainsSubstring: "event-listener-poll-period must be a positive non-zero integer",
		},
		{
			name:                 "zero pollSize on HTTP",
			nodeURL:              "http://example.com",
			timelockAddress:      pubStr,
			callProxyAddress:     callProxy,
			privateKey:           privStr,
			fromBlock:            fromBlock,
			pollPeriod:           1,
			listenerPollPeriod:   1,
			pollSize:             0,
			dryRun:               false,
			wantErr:              true,
			errContainsSubstring: "event-listener-poll-size must be a positive non-zero integer",
		},
		{
			name:               "valid HTTP parameters",
			nodeURL:            "https://example.com",
			timelockAddress:    pubStr,
			callProxyAddress:   callProxy,
			privateKey:         privStr,
			fromBlock:          fromBlock,
			pollPeriod:         5,
			listenerPollPeriod: 5,
			pollSize:           10,
			dryRun:             true,
			wantErr:            false,
		},
		{
			name:               "valid WS parameters ignore HTTP-only checks",
			nodeURL:            "ws://example.com",
			timelockAddress:    pubStr,
			callProxyAddress:   callProxy,
			privateKey:         privStr,
			fromBlock:          fromBlock,
			pollPeriod:         2,
			listenerPollPeriod: 0,
			pollSize:           0,
			dryRun:             false,
			wantErr:            false,
		},
		{
			name:                 "invalid private key",
			nodeURL:              "http://example.com",
			timelockAddress:      pubStr,
			callProxyAddress:     callProxy,
			privateKey:           "notBase58",
			fromBlock:            fromBlock,
			pollPeriod:           1,
			listenerPollPeriod:   1,
			pollSize:             1,
			dryRun:               false,
			wantErr:              true,
			errContainsSubstring: "the provided private key is not valid",
		},
	}

	for _, tc := range tests {
		tc := tc // capture range variable
		t.Run(tc.name, func(t *testing.T) {
			w, err := NewTimelockWorkerSolana(
				tc.nodeURL,
				tc.timelockAddress,
				tc.callProxyAddress,
				tc.privateKey,
				tc.fromBlock,
				tc.pollPeriod,
				tc.listenerPollPeriod,
				tc.pollSize,
				tc.dryRun,
				logger,
			)
			if tc.wantErr {
				require.Error(t, err)
				if tc.errContainsSubstring != "" {
					require.Contains(t, err.Error(), tc.errContainsSubstring)
				}
				require.Nil(t, w)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, w)

			// verify fields were set correctly
			require.Equal(t, tc.pollPeriod, w.pollPeriod)
			require.Equal(t, tc.listenerPollPeriod, w.listenerPollPeriod)
			require.Equal(t, tc.pollSize, w.pollSize)
			require.Equal(t, tc.dryRun, w.dryRun)

			// timelockProgramKey should parse back to the same PublicKey
			expectedPk, pkErr := solana.PublicKeyFromBase58(tc.timelockAddress)
			require.NoError(t, pkErr)
			require.True(t, expectedPk.Equals(w.timelockProgramKey))

			// privateKey should round-trip through Base58
			require.Equal(t, privStr, w.privateKey.String())

			// scheduler is still nil (TODO)
			require.Nil(t, w.scheduler)
		})
	}
}

func TestListen_PanicAndLogs(t *testing.T) {
	// set up an observer to capture logs at Info level
	core, recorded := observer.New(zap.InfoLevel)
	logger := zap.New(core).Sugar()

	// dummy solana client and key
	client := rpc.New("http://example.com")
	programKey, err := solana.PublicKeyFromBase58("11111111111111111111111111111111")
	require.NoError(t, err)
	priv, err := solana.NewRandomPrivateKey()
	require.NoError(t, err)

	// build WorkerSolana with fake scheduler

	w := &WorkerSolana{
		solanaClient:       client,
		timelockProgramKey: programKey,
		pollPeriod:         3,
		listenerPollPeriod: 4,
		pollSize:           5,
		dryRun:             false,
		logger:             logger,
		privateKey:         priv,
		lastSignature:      nil,
	}

	// Calling Listen should panic for now
	// TODO: remove once log polling and scheduler are implemented
	require.Panics(t, func() {
		_ = w.Listen(context.Background())
	})

	// Inspect captured logs
	logs := recorded.All()
	require.GreaterOrEqual(t, len(logs), 5, "expected at least 5 log entries")

	// 1st log: worker start
	require.Equal(t, "timelock-worker started [solana]", logs[0].Message)

	// 2nd log: program address line
	require.Contains(t, logs[1].Message, w.timelockProgramKey.String())

	// 3rd log: solana account address
	require.Contains(t, logs[2].Message, "solana account address:")

	// 4th log: poll period
	require.Contains(t, logs[3].Message, "Poll Period:")

	// 5th log: listener poll period or poll size
	require.Contains(t, logs[4].Message, "Event Listener Poll Period:")
}

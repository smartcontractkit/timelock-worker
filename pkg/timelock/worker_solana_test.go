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
	// prepare keypair
	priv, err := solana.NewRandomPrivateKey()
	require.NoError(t, err)
	privStr := priv.String()
	pubStr := priv.PublicKey().String()

	fromBlock := big.NewInt(0)
	logger := zap.NewNop().Sugar()

	tests := []struct {
		name            string
		nodeURL         string
		timelockAddress string
		privateKey      string
		fromBlock       *big.Int
		pollPeriod      int64
		listenerPoll    int64
		pollSize        uint64
		dryRun          bool
		errContains     string
		want            func(t *testing.T, w *WorkerSolana, err error)
	}{
		// happy path
		{
			name:            "valid HTTP",
			nodeURL:         "https://example.com",
			timelockAddress: pubStr,
			privateKey:      privStr,
			fromBlock:       fromBlock,
			pollPeriod:      5,
			listenerPoll:    5,
			pollSize:        10,
			dryRun:          true,
			want: func(t *testing.T, w *WorkerSolana, err error) {
				require.NoError(t, err)
				require.NotNil(t, w)
				require.Equal(t, int64(5), w.pollPeriod)
				require.Equal(t, int64(5), w.listenerPollPeriod)
				require.Equal(t, uint64(10), w.pollSize)
				require.True(t, w.dryRun)
				expectedPk, _ := solana.PublicKeyFromBase58(pubStr)
				require.True(t, expectedPk.Equals(w.timelockProgramKey))
				require.Equal(t, privStr, w.privateKey.String())
				require.Nil(t, w.scheduler)
			},
		},
		// error cases
		{
			name:            "invalid URL syntax",
			nodeURL:         "://bad",
			timelockAddress: pubStr,
			privateKey:      privStr,
			fromBlock:       fromBlock,
			pollPeriod:      1,
			listenerPoll:    1,
			pollSize:        1,
			dryRun:          false,
			errContains:     "parse \"://bad\"",
			want: func(t *testing.T, w *WorkerSolana, err error) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "parse \"://bad\": missing protocol scheme")
				require.Nil(t, w)
			},
		},
		{
			name:            "unsupported scheme",
			nodeURL:         "ftp://example.com",
			timelockAddress: pubStr,
			privateKey:      privStr,
			fromBlock:       fromBlock,
			pollPeriod:      1,
			listenerPoll:    1,
			pollSize:        1,
			dryRun:          false,
			errContains:     "invalid node URL",
			want: func(t *testing.T, w *WorkerSolana, err error) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "invalid node URL")
				require.Nil(t, w)
			},
		},
		{
			name:            "invalid timelock address",
			nodeURL:         "http://example.com",
			timelockAddress: "notBase58",
			privateKey:      privStr,
			fromBlock:       fromBlock,
			pollPeriod:      1,
			listenerPoll:    1,
			pollSize:        1,
			dryRun:          false,
			errContains:     "timelock addresses provided is not valid",
			want: func(t *testing.T, w *WorkerSolana, err error) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "timelock addresses provided is not valid")
				require.Nil(t, w)
			},
		},
		{
			name:            "negative pollPeriod",
			nodeURL:         "http://example.com",
			timelockAddress: pubStr,
			privateKey:      privStr,
			fromBlock:       fromBlock,
			pollPeriod:      0,
			listenerPoll:    1,
			pollSize:        1,
			dryRun:          false,
			errContains:     "poll-period must be a positive non-zero integer",
			want: func(t *testing.T, w *WorkerSolana, err error) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "poll-period must be a positive non-zero integer")
				require.Nil(t, w)
			},
		},
		{
			name:            "zero listenerPoll on HTTP",
			nodeURL:         "https://example.com",
			timelockAddress: pubStr,
			privateKey:      privStr,
			fromBlock:       fromBlock,
			pollPeriod:      1,
			listenerPoll:    0,
			pollSize:        1,
			dryRun:          false,
			errContains:     "event-listener-poll-period must be a positive non-zero integer",
			want: func(t *testing.T, w *WorkerSolana, err error) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "event-listener-poll-period must be a positive non-zero integer")
				require.Nil(t, w)
			},
		},
		{
			name:            "zero pollSize on HTTP",
			nodeURL:         "http://example.com",
			timelockAddress: pubStr,
			privateKey:      privStr,
			fromBlock:       fromBlock,
			pollPeriod:      1,
			listenerPoll:    1,
			pollSize:        0,
			dryRun:          false,
			errContains:     "event-listener-poll-size must be a positive non-zero integer",
			want: func(t *testing.T, w *WorkerSolana, err error) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "event-listener-poll-size must be a positive non-zero integer")
				require.Nil(t, w)
			},
		},
		{
			name:            "invalid private key",
			nodeURL:         "http://example.com",
			timelockAddress: pubStr,
			privateKey:      "notBase58",
			fromBlock:       fromBlock,
			pollPeriod:      1,
			listenerPoll:    1,
			pollSize:        1,
			dryRun:          false,
			errContains:     "the provided private key is not valid",
			want: func(t *testing.T, w *WorkerSolana, err error) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "the provided private key is not valid")
				require.Nil(t, w)
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			w, err := NewTimelockWorkerSolana(
				tc.nodeURL,
				tc.timelockAddress,
				tc.privateKey,
				tc.fromBlock,
				tc.pollPeriod,
				tc.listenerPoll,
				tc.pollSize,
				tc.dryRun,
				logger,
			)
			tc.want(t, w, err)
		})
	}
}

func TestListen_PanicAndLogs(t *testing.T) {
	core, recorded := observer.New(zap.InfoLevel)
	logger := zap.New(core).Sugar()

	client := rpc.New("http://example.com")
	programKey, err := solana.PublicKeyFromBase58("11111111111111111111111111111111")
	require.NoError(t, err)
	priv, err := solana.NewRandomPrivateKey()
	require.NoError(t, err)

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

	require.Panics(t, func() {
		_ = w.Listen(context.Background())
	})

	logs := recorded.All()
	require.GreaterOrEqual(t, len(logs), 5)
	require.Equal(t, "timelock-worker started [solana]", logs[0].Message)
	require.Contains(t, logs[1].Message, w.timelockProgramKey.String())
	require.Contains(t, logs[2].Message, "Solana account address:")
	require.Contains(t, logs[3].Message, "Poll Period:")
	require.Contains(t, logs[4].Message, "Event Listener Poll Period:")
}

package integration

import (
	"context"
	"math/big"
	"strings"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/samber/lo"
	contracts "github.com/smartcontractkit/ccip-owner-contracts/gethwrappers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"

	"github.com/smartcontractkit/timelock-worker/pkg/timelock"
	timelockTests "github.com/smartcontractkit/timelock-worker/tests"
)

func (s *integrationTestSuite) TestTimelockWorkerListen() {
	ctx, cancel := context.WithCancel(s.Ctx)
	defer cancel()

	account := NewTestAccount(s.T())
	s.Logf("new account created: %v", account)

	_, err := s.GethContainer.CreateAccount(ctx, account.HexAddress, account.HexPrivateKey, 1)
	s.Require().NoError(err)

	backend := NewRPCBackend(s.T(), ctx, s.GethContainer.HTTPConnStr(s.T(), ctx))
	transactor := s.KeyedTransactor(account.PrivateKey, nil)

	expectedEvents := []string{
		"RoleAdminChanged", // <--+
		"RoleAdminChanged", //    |
		"RoleAdminChanged", //    |
		"RoleAdminChanged", //    |
		"RoleAdminChanged", //    |
		"RoleGranted",      //    |-- contract deployment
		"RoleGranted",      //    |
		"RoleGranted",      //    |
		"RoleGranted",      //    |
		"MinDelayChange",   // <--+
		"MinDelayChange",   // <----- updateDelay call
	}

	tests := []struct {
		name string
		url  string
	}{
		{name: "http connection", url: s.GethContainer.HTTPConnStr(s.T(), ctx)},
		{name: "websocket connection", url: s.GethContainer.WSConnStr(s.T(), ctx)},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			sctx, cancel := context.WithCancel(ctx)
			defer cancel()

			logger, logs := timelockTests.NewTestLogger()

			timelockAddress, _, _, timelockContract := DeployTimelock(s.T(), ctx, transactor, backend,
				account.Address, big.NewInt(1))
			callProxyAddress, _, _, _ := DeployCallProxy(s.T(), ctx, transactor, backend, timelockAddress)

			go runTimelockWorker(s.T(), sctx, tt.url, timelockAddress.String(), callProxyAddress.String(),
				account.HexPrivateKey, big.NewInt(0), int64(60), int64(1), uint64(10), true, logger)

			UpdateDelay(s.T(), ctx, transactor, backend, timelockContract, big.NewInt(10))

			s.EventuallyWithT(func(collect *assert.CollectT) {
				logEntries := logs.FilterMessage("discarding event").All()
				events := lo.Map(logEntries, func(e observer.LoggedEntry, _ int) string { return e.Context[0].String })
				assert.ElementsMatch(collect, events, expectedEvents)
			}, 5*time.Second, 200*time.Millisecond)
		})
	}
}

func (s *integrationTestSuite) TestTimelockWorkerDryRun() {
	ctx, cancel := context.WithCancel(s.Ctx)
	defer cancel()

	account := NewTestAccount(s.T())
	s.Logf("new account created: %v", account)

	_, err := s.GethContainer.CreateAccount(ctx, account.HexAddress, account.HexPrivateKey, 1)
	s.Require().NoError(err)

	gethURL := s.GethContainer.HTTPConnStr(s.T(), ctx)
	backend := NewRPCBackend(s.T(), ctx, gethURL)

	transactor := s.KeyedTransactor(account.PrivateKey, nil)

	calls := []contracts.RBACTimelockCall{{
		Target: common.HexToAddress("0x000000000000000000000000000000000000000"),
		Value:  big.NewInt(1),
		Data:   hexutil.MustDecode("0x0123456789abcdef"),
	}}

	tests := []struct {
		name   string
		dryRun bool
		assert func(t *testing.T, logs *observer.ObservedLogs)
	}{
		{
			name:   "dry run enabled",
			dryRun: true,
			assert: func(t *testing.T, logs *observer.ObservedLogs) {
				t.Helper()
				s.Require().EventuallyWithT(func(t *assert.CollectT) {
					assertLogMessage(t, logs, "CallScheduled received")
					assertLogMessage(t, logs, "nop.addToScheduler")
				}, 2*time.Second, 100*time.Millisecond)
			},
		},
		{
			name:   "dry run disabled",
			dryRun: false,
			assert: func(t *testing.T, logs *observer.ObservedLogs) {
				t.Helper()
				s.Require().EventuallyWithT(func(t *assert.CollectT) {
					assertLogMessage(t, logs, "scheduling operation: 371141ec10c0cc52996bed94240931136172d0b46bdc4bceaea1ef76675c1237")
					assertLogMessage(t, logs, "scheduled operation: 371141ec10c0cc52996bed94240931136172d0b46bdc4bceaea1ef76675c1237")
				}, 2*time.Second, 100*time.Millisecond, logMessages(logs))
			},
		},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			tctx, cancel := context.WithCancel(ctx)
			defer cancel()

			logger, logs := timelockTests.NewTestLogger()

			timelockAddress, _, _, timelockContract := DeployTimelock(s.T(), tctx, transactor, backend,
				account.Address, big.NewInt(1))
			callProxyAddress, _, _, _ := DeployCallProxy(s.T(), tctx, transactor, backend, timelockAddress)

			go runTimelockWorker(s.T(), tctx, gethURL, timelockAddress.String(), callProxyAddress.String(),
				account.HexPrivateKey, big.NewInt(0), int64(1), int64(1), uint64(10), tt.dryRun, logger)

			ScheduleBatch(s.T(), tctx, transactor, backend, timelockContract, calls, [32]byte{}, [32]byte{}, big.NewInt(1))

			tt.assert(s.T(), logs)
		})
	}
}

func (s *integrationTestSuite) TestTimelockWorkerCancelledEvent() {
	// --- arrange ---
	ctx, cancel := context.WithCancel(s.Ctx)
	defer cancel()

	account := NewTestAccount(s.T())
	_, err := s.GethContainer.CreateAccount(ctx, account.HexAddress, account.HexPrivateKey, 1)
	s.Require().NoError(err)
	s.Logf("new account created: %v", account)

	gethURL := s.GethContainer.HTTPConnStr(s.T(), ctx)
	backend := NewRPCBackend(s.T(), ctx, gethURL)
	transactor := s.KeyedTransactor(account.PrivateKey, nil)
	logger, logs := timelockTests.NewTestLogger()

	timelockAddress, _, _, timelockContract := DeployTimelock(s.T(), ctx, transactor, backend,
		account.Address, big.NewInt(1))
	callProxyAddress, _, _, _ := DeployCallProxy(s.T(), ctx, transactor, backend, timelockAddress)

	go runTimelockWorker(s.T(), ctx, gethURL, timelockAddress.String(), callProxyAddress.String(),
		account.HexPrivateKey, big.NewInt(0), int64(1), int64(1), uint64(10), false, logger)

	calls := []contracts.RBACTimelockCall{{
		Target: common.HexToAddress("0x000000000000000000000000000000000000000"),
		Value:  big.NewInt(1),
		Data:   hexutil.MustDecode("0x0123456789abcdef"),
	}}
	predecessor := common.Hash{}
	salt := common.Hash{}
	operationID := common.HexToHash("371141ec10c0cc52996bed94240931136172d0b46bdc4bceaea1ef76675c1237")

	// --- act ---
	ScheduleBatch(s.T(), ctx, transactor, backend, timelockContract, calls, predecessor, salt, big.NewInt(1))
	CancelBatch(s.T(), ctx, transactor, backend, timelockContract, operationID)

	// --- assert ---
	s.Require().EventuallyWithT(func(collect *assert.CollectT) {
		assertLogMessage(collect, logs, "Cancelled received, cancelling operation")
		assertLogMessage(collect, logs, "de-scheduling operation: 371141ec10c0cc52996bed94240931136172d0b46bdc4bceaea1ef76675c1237")
		assertLogMessage(collect, logs, "de-scheduled operation: 371141ec10c0cc52996bed94240931136172d0b46bdc4bceaea1ef76675c1237")
	}, 3*time.Second, 100*time.Millisecond, logMessages(logs))
}

func (s *integrationTestSuite) TestTimelockWorkerPollSize() {
	// --- arrange ---
	ctx, cancel := context.WithCancel(s.Ctx)
	defer cancel()

	account := NewTestAccount(s.T())
	_, err := s.GethContainer.CreateAccount(ctx, account.HexAddress, account.HexPrivateKey, 1)
	s.Require().NoError(err)
	s.Logf("new account created: %v", account)

	gethURL := s.GethContainer.HTTPConnStr(s.T(), ctx)
	backend := NewRPCBackend(s.T(), ctx, gethURL)
	transactor := s.KeyedTransactor(account.PrivateKey, nil)
	logger, logs := timelockTests.NewTestLogger()

	timelockAddress, _, _, _ := DeployTimelock(s.T(), ctx, transactor, backend,
		account.Address, big.NewInt(1))
	callProxyAddress, _, _, _ := DeployCallProxy(s.T(), ctx, transactor, backend, timelockAddress)

	time.Sleep(1 * time.Second) // wait for a few blocks before starting the timelock worker service

	// --- act ---
	go runTimelockWorker(s.T(), ctx, gethURL, timelockAddress.String(), callProxyAddress.String(),
		account.HexPrivateKey, big.NewInt(0), int64(1), int64(1), uint64(2), false, logger)

	// --- assert ---
	s.Require().EventuallyWithT(func(collect *assert.CollectT) {
		assertLogMessage(collect, logs, "fetching logs from block 0 to block 1")
		assertLogMessage(collect, logs, "fetching logs from block 2 to block 3")
		assertLogMessage(collect, logs, "fetching logs from block 4 to block 5")
	}, 2*time.Second, 100*time.Millisecond, logMessages(logs))
}

// ----- helpers -----

func runTimelockWorker(
	t *testing.T, ctx context.Context, nodeURL, timelockAddress, callProxyAddress, privateKey string,
	fromBlock *big.Int, pollPeriod int64, listenerPollPeriod int64, listenerPollSize uint64,
	dryRun bool, logger *zap.Logger,
) {
	t.Logf("TimelockWorker.Listen(%v, %v, %v, %v, %v, %v, %v, %v)", nodeURL, timelockAddress,
		callProxyAddress, privateKey, fromBlock, pollPeriod, listenerPollPeriod, listenerPollSize)
	timelockWorker, err := timelock.NewTimelockWorkerEVM(nodeURL, timelockAddress,
		callProxyAddress, privateKey, fromBlock, pollPeriod, listenerPollPeriod, listenerPollSize, dryRun, logger.Sugar())
	require.NoError(t, err)
	require.NotNil(t, timelockWorker)

	err = timelockWorker.Listen(ctx)
	require.NoError(t, err)
}

func assertLogMessage(t assert.TestingT, logs *observer.ObservedLogs, message string) {
	assert.Equal(t, logs.FilterMessage(message).Len(), 1)
}

func logMessages(logs *observer.ObservedLogs) string {
	m := make([]string, 0, logs.Len())
	for _, entry := range logs.All() {
		m = append(m, entry.Message)
	}
	return "LOGS:\n" + strings.Join(m, "\n")
}

package integration

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"regexp"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/rs/zerolog"
	"github.com/samber/lo"
	contracts "github.com/smartcontractkit/ccip-owner-contracts/gethwrappers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

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

	gethURL := s.GethContainer.HTTPConnStr(s.T(), ctx)
	backend := NewRPCBackend(s.T(), ctx, gethURL)

	transactor := s.KeyedTransactor(account.PrivateKey, nil)

	tests := []struct {
		name string
		url  string
	}{
		{name: "http connection", url: gethURL},
		{name: "websocket connection", url: s.GethContainer.WSConnStr(s.T(), ctx)},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			sctx, cancel := context.WithCancel(ctx)
			defer cancel()

			logger := timelockTests.NewTestLogger(zerolog.Nop()) // "zerolog.TestWriter{T: t, Frame: 6}" when debugging

			timelockAddress, _, _, timelockContract := DeployTimelock(s.T(), ctx, transactor, backend,
				account.Address, big.NewInt(1))
			callProxyAddress, _, _, _ := DeployCallProxy(s.T(), ctx, transactor, backend, timelockAddress)

			go runTimelockWorker(s.T(), sctx, tt.url, timelockAddress.String(), callProxyAddress.String(),
				account.HexPrivateKey, big.NewInt(0), int64(60), int64(1), true, logger.Logger())

			UpdateDelay(s.T(), ctx, transactor, backend, timelockContract, big.NewInt(10))

			assertCapturedLogMessages(s.T(), logger)
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

	tests := []struct {
		name   string
		dryRun bool
		assert func(t *testing.T, logger timelockTests.TestLogger)
	}{
		{
			name:   "dry run enabled",
			dryRun: true,
			assert: func(t *testing.T, logger timelockTests.TestLogger) {
				messages := []string{
					`"message":"CallScheduled received"`,
					`"message":"nop.addToScheduler"`,
				}
				s.Require().EventuallyWithT(func(t *assert.CollectT) {
					for _, message := range messages {
						s.Assert().True(containsMatchingMessage(logger, regexp.MustCompile(message)))
					}
				}, 2*time.Second, 100*time.Millisecond)
			},
		},
		{
			name:   "dry run disabled",
			dryRun: false,
			assert: func(t *testing.T, logger timelockTests.TestLogger) {
				messages := []string{
					`"message":"scheduling operation: 371141ec10c0cc52996bed94240931136172d0b46bdc4bceaea1ef76675c1237"`,
					`"message":"scheduled operation: 371141ec10c0cc52996bed94240931136172d0b46bdc4bceaea1ef76675c1237"`,
				}
				s.Require().EventuallyWithT(func(t *assert.CollectT) {
					for _, message := range messages {
						s.Assert().True(containsMatchingMessage(logger, regexp.MustCompile(message)))
					}
				}, 2*time.Second, 100*time.Millisecond)
			},
		},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			tctx, cancel := context.WithCancel(ctx)
			defer cancel()

			logger := timelockTests.NewTestLogger(zerolog.Nop()) // "zerolog.TestWriter{T: t, Frame: 6}" when debugging

			timelockAddress, _, _, timelockContract := DeployTimelock(s.T(), tctx, transactor, backend,
				account.Address, big.NewInt(1))
			callProxyAddress, _, _, _ := DeployCallProxy(s.T(), tctx, transactor, backend, timelockAddress)

			go runTimelockWorker(s.T(), tctx, gethURL, timelockAddress.String(), callProxyAddress.String(),
				account.HexPrivateKey, big.NewInt(0), int64(1), int64(1), tt.dryRun, logger.Logger())

			calls := []contracts.RBACTimelockCall{{
				Target: common.HexToAddress("0x000000000000000000000000000000000000000"),
				Value:  big.NewInt(1),
				Data:   hexutil.MustDecode("0x0123456789abcdef"),
			}}
			ScheduleBatch(s.T(), tctx, transactor, backend, timelockContract, calls, [32]byte{}, [32]byte{}, big.NewInt(1))

			tt.assert(s.T(), logger)
		})
	}
}

// ----- helpers -----

func runTimelockWorker(
	t *testing.T, ctx context.Context, nodeURL, timelockAddress, callProxyAddress, privateKey string,
	fromBlock *big.Int, pollPeriod int64, listenerPollPeriod int64, dryRun bool, logger *zerolog.Logger,
) {
	t.Logf("TimelockWorker.Listen(%v, %v, %v, %v, %v, %v, %v)", nodeURL, timelockAddress,
		callProxyAddress, privateKey, fromBlock, pollPeriod, listenerPollPeriod)
	timelockWorker, err := timelock.NewTimelockWorker(nodeURL, timelockAddress,
		callProxyAddress, privateKey, fromBlock, pollPeriod, listenerPollPeriod, dryRun, logger)
	require.NoError(t, err)
	require.NotNil(t, timelockWorker)

	err = timelockWorker.Listen(ctx)
	require.NoError(t, err)
}

func assertCapturedLogMessages(t *testing.T, logger timelockTests.TestLogger) {
	t.Helper()
	require.EventuallyWithT(t, func(collect *assert.CollectT) {
		messages := selectDiscardingEventMessages(t, logger)
		assert.ElementsMatch(collect, messages, []string{
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
		})
	}, 5*time.Second, 200*time.Millisecond)
}

func selectMatchingMessage(logger timelockTests.TestLogger, pattern *regexp.Regexp) []string {
	return lo.Filter(logger.Messages(), func(loggedMessage string, _ int) bool {
		return pattern.MatchString(loggedMessage)
	})
}

func containsMatchingMessage(logger timelockTests.TestLogger, pattern *regexp.Regexp) bool {
	return len(selectMatchingMessage(logger, pattern)) > 0
}

func selectDiscardingEventMessages(t *testing.T, logger timelockTests.TestLogger) []string {
	t.Helper()

	messages := selectMatchingMessage(logger, regexp.MustCompile(`"message":"discarding event"`))
	return lo.Map(messages, func(message string, _ int) string {
		parsedEntry := struct{ Event string }{}
		err := json.Unmarshal([]byte(message), &parsedEntry)
		require.NoError(t, err)
		return parsedEntry.Event
	})
}

func assertJSONSubset(t assert.TestingT, expected string, actual string) bool {
	var expectedJSONAsInterface, actualJSONAsInterface interface{}

	if err := json.Unmarshal([]byte(expected), &expectedJSONAsInterface); err != nil {
		return assert.Fail(t, fmt.Sprintf("Expected value ('%s') is not valid json.\nJSON parsing error: '%s'", expected, err.Error()))
	}

	if err := json.Unmarshal([]byte(actual), &actualJSONAsInterface); err != nil {
		return assert.Fail(t, fmt.Sprintf("Input ('%s') needs to be valid json.\nJSON parsing error: '%s'", actual, err.Error()))
	}

	return assert.Subset(t, expectedJSONAsInterface, actualJSONAsInterface)
}

func requireJSONSubset(t require.TestingT, expected string, actual string) {
	if assertJSONSubset(t, expected, actual) {
		return
	}
	t.FailNow()
}

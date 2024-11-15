package integration

import (
	"context"
	"encoding/json"
	"math/big"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/timelock-worker/pkg/timelock"
	timelockTests "github.com/smartcontractkit/timelock-worker/tests"
)

func (s *integrationTestSuite) TestTimelockWorkerListen() {
	ctx, cancel := context.WithCancel(s.Ctx)
	defer cancel()

	account1 := NewTestAccount(s.T())
	s.Logf("new account created: %v", account1)

	_, err := s.GethContainer.CreateAccount(ctx, account1.hexAddress, account1.hexPrivateKey, 1)
	s.Require().NoError(err)

	// create geth rpc client
	gethURL := s.GethContainer.HTTPConnStr(s.T(), ctx)
	client, err := ethclient.DialContext(ctx, gethURL)
	s.Require().NoError(err)

	transactor := s.KeyedTransactor(account1.privateKey, nil)

	tests := []struct {
		name string
		url  string
	}{
		{name: "http connection", url: gethURL},
		{name: "websocket connection", url: s.GethContainer.WSConnStr(s.T(), ctx)},
	}
	for _, tt := range tests {
		s.Run(tt.name, func(t *testing.T) {
			sctx, cancel := context.WithCancel(ctx)
			defer cancel()

			logger := timelockTests.NewTestLogger(zerolog.Nop()) // "zerolog.TestWriter{T: t, Frame: 6}" when debugging

			timelockAddress, _, _, timelockContract := s.DeployTimelock(ctx, transactor, client, account1.address)
			callProxyAddress, _, _, _ := s.DeployCallProxy(ctx, transactor, client, timelockAddress)

			go runTimelockWorker(s.T(), sctx, tt.url, timelockAddress.String(), callProxyAddress.String(),
				account1.hexPrivateKey, big.NewInt(0), int64(60), int64(1), logger.Logger())

			s.UpdateDelay(ctx, transactor, client, timelockContract, big.NewInt(10))

			assertCapturedLogMessages(s.T(), logger)
		})
	}
}

// ----- helpers -----

func runTimelockWorker(
	t *testing.T, ctx context.Context, nodeURL, timelockAddress, callProxyAddress, privateKey string,
	fromBlock *big.Int, pollPeriod int64, listenerPollPeriod int64, logger *zerolog.Logger,
) {
	t.Logf("TimelockWorker.Listen(%v, %v, %v, %v, %v, %v, %v)", nodeURL, timelockAddress,
		callProxyAddress, privateKey, fromBlock, pollPeriod, listenerPollPeriod)
	timelockWorker, err := timelock.NewTimelockWorker(nodeURL, timelockAddress,
		callProxyAddress, privateKey, fromBlock, pollPeriod, listenerPollPeriod, logger)
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

func selectDiscardingEventMessages(t *testing.T, logger timelockTests.TestLogger) []string {
	t.Helper()

	selectedEvents := []string{}

	for _, entry := range logger.Messages() {
		parsedEntry := struct {
			Message string
			Event   string
		}{}
		err := json.Unmarshal([]byte(entry), &parsedEntry)
		require.NoError(t, err)

		if parsedEntry.Message == "discarding event" {
			selectedEvents = append(selectedEvents, parsedEntry.Event)
		}
	}

	return selectedEvents
}

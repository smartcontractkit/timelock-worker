package solana

import (
	"context"
	"time"

	"github.com/gagliardetto/solana-go"
	"github.com/samber/lo"
	e2eutils "github.com/smartcontractkit/mcms/e2e/utils/solana"
	solanasdk "github.com/smartcontractkit/mcms/sdk/solana"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap/zaptest/observer"

	timelockTests "github.com/smartcontractkit/timelock-worker/tests"
)

// TestTimelockWorkerListen tests the Solana timelock worker's ability to listen for events and process them correctly.
func (s *solanaIntegrationTestSuite) TestTimelockWorkerListen() {

	ctx, cancel := context.WithCancel(s.Ctx)
	defer cancel()
	e2eutils.FundAccounts(s.T(), ctx, []solana.PublicKey{s.TestPrivateKey.PublicKey()}, 1, s.solanaClient)
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
	sctx, cancel := context.WithCancel(ctx)
	defer cancel()

	logger, logs := timelockTests.NewTestLogger()
	instanceIDSeed := solanasdk.PDASeed([32]byte{'t', 'i', 'm', 'e', 'l', 'o', 'c', 'k'})
	s.initializeTimelockInstance(instanceIDSeed, time.Second*1)
	contractID := solanasdk.ContractAddress(s.TimelockProgramID, instanceIDSeed)
	go runTimelockWorkerSolana(s.T(),
		sctx,
		s.solanaBlockchain.Nodes[0].HostHTTPUrl,
		contractID,
		s.TestPrivateKey.String(),
		int64(60),
		int64(1),
		10,
		true,
		logger)
	predecessor := [32]byte{}
	salt := [32]byte{123}
	s.scheduleTestIx(instanceIDSeed, predecessor, salt)

	s.EventuallyWithT(func(collect *assert.CollectT) {
		logEntries := logs.FilterMessage("discarding event").All()
		events := lo.Map(logEntries, func(e observer.LoggedEntry, _ int) string { return e.Context[0].String })
		assert.ElementsMatch(collect, events, expectedEvents)
	}, 5*time.Second, 200*time.Millisecond)
}

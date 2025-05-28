package solana

import (
	"context"
	"time"

	"github.com/gagliardetto/solana-go"
	e2eutils "github.com/smartcontractkit/mcms/e2e/utils/solana"
	solanasdk "github.com/smartcontractkit/mcms/sdk/solana"
	"github.com/stretchr/testify/assert"

	timelockTests "github.com/smartcontractkit/timelock-worker/tests"
)

// TestTimelockWorkerListen tests the Solana timelock worker's ability to listen for events and process them correctly.
func (s *solanaIntegrationTestSuite) TestTimelockWorkerListen() {

	ctx, cancel := context.WithCancel(s.Ctx)
	defer cancel()
	e2eutils.FundAccounts(s.T(), ctx, []solana.PublicKey{s.TestPrivateKey.PublicKey()}, 1, s.solanaClient)

	sctx, cancel := context.WithCancel(ctx)
	defer cancel()

	logger, logs := timelockTests.NewTestLogger()
	instanceIDSeed := solanasdk.PDASeed([32]byte{'t', 'i', 'm', 'e', 'l', 'o', 'c', 'k'})
	s.initializeTimelockInstance(instanceIDSeed, time.Second*1)
	contractID := solanasdk.ContractAddress(s.TimelockProgramID, instanceIDSeed)

	predecessor := [32]byte{}
	salt := [32]byte{123}
	s.scheduleTestIx(instanceIDSeed, predecessor, salt)

	go runTimelockWorkerSolana(s.T(),
		sctx,
		s.solanaBlockchain.Nodes[0].HostHTTPUrl,
		contractID,
		s.TestPrivateKey.String(),
		int64(1),
		int64(1),
		10,
		true,
		logger)

	s.EventuallyWithT(func(collect *assert.CollectT) {
		logEntries := logs.All()
		if !assert.GreaterOrEqual(collect, len(logEntries), 11, "Expected at least 12 log entries") {
			return
		}
		assert.Greater(collect, len(logEntries), 11, "Expected more than 8 log entries, got %d", len(logEntries))
		assert.Equal(collect, logEntries[0].Message, "timelock-worker started [solana]")
		assert.Equal(collect, logEntries[1].Message, "\tTimelock program addresses: DoajfR5tK24xVw51fWcawUZWhAXD8yrBJVacc13neVQA")
		assert.Equal(collect, logEntries[2].Message, "\taccount address: 9n1pyVGGo6V4mpiSDMVay5As9NurEkY283wwRk1Kto2C")
		assert.Equal(collect, logEntries[3].Message, "\tPoll Period: 1s")
		assert.Equal(collect, logEntries[4].Message, "\tEvent Listener Poll Period: 1s")
		assert.Equal(collect, logEntries[5].Message, "\tEvent Listener Poll #Logs: 10")
		assert.Equal(collect, logEntries[6].Message, "starting pollNewSignatures: program=DoajfR5tK24xVw51fWcawUZWhAXD8yrBJVacc13neVQA pollPeriod=1s pollSize=10")
		assert.Contains(collect, logEntries[7].Message, "new pollNewSignatures tick")
		assert.Equal(collect, logEntries[8].Message, "found 7 new signatures")
		assert.Contains(collect, logEntries[9].Message, "found event scheduled:")
		assert.Contains(collect, logEntries[10].Message, "CallScheduled received")
	}, 20*time.Second, 200*time.Millisecond)
}

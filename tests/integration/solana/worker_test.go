package solana

import (
	"context"
	"time"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	e2eutils "github.com/smartcontractkit/mcms/e2e/utils/solana"
	solanasdk "github.com/smartcontractkit/mcms/sdk/solana"
	"github.com/stretchr/testify/assert"

	timelockTests "github.com/smartcontractkit/timelock-worker/tests"
)

// TestTimelockWorkerListen tests the Solana timelock worker's ability to listen for events and process them correctly.
func (s *solanaIntegrationTestSuite) TestTimelockWorkerListen_ScheduleOP() {

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
		rpc.CommitmentConfirmed,
		logger)

	s.EventuallyWithT(func(collect *assert.CollectT) {
		logEntries := logs.All()
		if !assert.GreaterOrEqual(collect, len(logEntries), 11, "Expected at least 12 log entries") {
			return
		}

		assert.Equal(collect, logEntries[0].Message, "timelock-worker started [solana]")
		assert.Equal(collect, logEntries[1].Message, "\tTimelock program:")
		assert.Equal(collect, logEntries[2].Message, "\tSolana account address")
		assert.Equal(collect, logEntries[3].Message, "\tPoll Period")
		assert.Equal(collect, logEntries[4].Message, "\tEvent Listener Poll Period")
		assert.Equal(collect, logEntries[5].Message, "\tEvent Listener Poll #Logs")
		assert.Equal(collect, logEntries[6].Message, "nop.runScheduler")
		assert.Equal(collect, logEntries[7].Message, "starting pollSignatures")
		assert.Contains(collect, logEntries[8].Message, "found event scheduled:")
		assert.Contains(collect, logEntries[9].Message, "event received")
		assert.Contains(collect, logEntries[10].Message, "nop.addToScheduler")
	}, 20*time.Second, 200*time.Millisecond)
}

func (s *solanaIntegrationTestSuite) TestTimelockWorkerListen_CancelScheduledOP() {
	ctx, cancel := context.WithCancel(s.Ctx)
	defer cancel()
	e2eutils.FundAccounts(s.T(), ctx, []solana.PublicKey{s.TestPrivateKey.PublicKey()}, 1, s.solanaClient)

	sctx, cancel := context.WithCancel(ctx)
	defer cancel()

	logger, logs := timelockTests.NewTestLogger()
	instanceIDSeed := solanasdk.PDASeed([32]byte{'c', 'a', 'n', 'c', 'e', 'l', 't', 'e', 's', 't'})
	s.initializeTimelockInstance(instanceIDSeed, time.Second*1)
	contractID := solanasdk.ContractAddress(s.TimelockProgramID, instanceIDSeed)

	predecessor := [32]byte{}
	salt := [32]byte{42}
	_, operationID := s.scheduleTestIx(instanceIDSeed, predecessor, salt)

	// Cancel the operation
	s.cancelScheduledIx(instanceIDSeed, operationID)

	go runTimelockWorkerSolana(s.T(),
		sctx,
		s.solanaBlockchain.Nodes[0].HostHTTPUrl,
		contractID,
		s.TestPrivateKey.String(),
		int64(1),
		int64(1),
		10,
		true,
		rpc.CommitmentConfirmed,
		logger,
	)

	s.EventuallyWithT(func(collect *assert.CollectT) {
		logEntries := logs.All()
		if !assert.GreaterOrEqual(collect, len(logEntries), 12, "Expected at least 12 log entries") {
			return
		}

		assert.Equal(collect, "timelock-worker started [solana]", logEntries[0].Message)
		assert.Equal(collect, "starting pollSignatures", logEntries[7].Message)
		assert.Contains(collect, logEntries[8].Message, "found event scheduled:")
		assert.Contains(collect, logEntries[9].Message, "error handling scheduled event")
		assert.Contains(collect, logEntries[10].Message, "found event cancelled")
		assert.Contains(collect, logEntries[11].Message, "event received, cancelling operation")
		assert.Contains(collect, logEntries[12].Message, "nop.delFromScheduler")
	}, 25*time.Second, 200*time.Millisecond)
}

func (s *solanaIntegrationTestSuite) TestTimelockWorkerListen_ExecutedOP() {
	ctx, cancel := context.WithCancel(s.Ctx)
	defer cancel()
	e2eutils.FundAccounts(s.T(), ctx, []solana.PublicKey{s.TestPrivateKey.PublicKey()}, 1, s.solanaClient)

	sctx, cancel := context.WithCancel(ctx)
	defer cancel()

	logger, logs := timelockTests.NewTestLogger()
	instanceIDSeed := solanasdk.PDASeed([32]byte{'e', 'x', 'e', 'c', 't', 'e', 's', 't'})

	s.initializeTimelockInstance(instanceIDSeed, time.Second*1)
	contractID := solanasdk.ContractAddress(s.TimelockProgramID, instanceIDSeed)

	predecessor := [32]byte{}
	salt := [32]byte{77}
	ix, operationID := s.scheduleTestIx(instanceIDSeed, predecessor, salt)

	time.Sleep(5 * time.Second)
	// Execute the operation
	s.executeScheduledIx(instanceIDSeed, operationID, ix)

	go runTimelockWorkerSolana(s.T(),
		sctx,
		s.solanaBlockchain.Nodes[0].HostHTTPUrl,
		contractID,
		s.TestPrivateKey.String(),
		int64(1),
		int64(1),
		10,
		true,
		rpc.CommitmentConfirmed,
		logger,
	)

	s.EventuallyWithT(func(collect *assert.CollectT) {
		logEntries := logs.All()
		if !assert.GreaterOrEqual(collect, len(logEntries), 1, "Expected at least 13 log entries") {
			return
		}

		assert.Equal(collect, "timelock-worker started [solana]", logEntries[0].Message)
		assert.Equal(collect, "starting pollSignatures", logEntries[7].Message)
		assert.Contains(collect, logEntries[8].Message, "found event scheduled:")
		assert.Contains(collect, logEntries[9].Message, "found event executed")
		assert.Contains(collect, logEntries[10].Message, "event received, deleting operation from scheduler")
		assert.Contains(collect, logEntries[11].Message, "nop.delFromScheduler")
	}, 25*time.Second, 200*time.Millisecond)
}

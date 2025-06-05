package solana

import (
	"context"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	cpistub "github.com/smartcontractkit/chainlink-ccip/chains/solana/gobindings/external_program_cpi_stub"
	rmnremotebindings "github.com/smartcontractkit/chainlink-ccip/chains/solana/gobindings/rmn_remote"
	timelockbindings "github.com/smartcontractkit/chainlink-ccip/chains/solana/gobindings/timelock"
	solanacommon "github.com/smartcontractkit/chainlink-ccip/chains/solana/utils/common"
	"github.com/smartcontractkit/mcms"
	e2eutils "github.com/smartcontractkit/mcms/e2e/utils/solana"
	solanasdk "github.com/smartcontractkit/mcms/sdk/solana"
	mcmstypes "github.com/smartcontractkit/mcms/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	timelockTests "github.com/smartcontractkit/timelock-worker/tests"
	evmtests "github.com/smartcontractkit/timelock-worker/tests/integration/evm"
)

// TestTimelockWorkerListen tests the Solana timelock worker's ability to listen for scheduled and canceled operations.
func (s *solanaIntegrationTestSuite) TestTimelockWorkerListen() {
	s.withTimeout(45 * time.Second)

	// --- Arrange shared setup ---
	e2eutils.FundAccounts(s.T(), s.Ctx, []solana.PublicKey{s.TestPrivateKey.PublicKey()}, 1, s.solanaClient)

	logger, logs := timelockTests.NewTestLogger()
	instanceIDSeed := solanasdk.PDASeed([32]byte{'t', 'i', 'm', 'e', 'l', 'o', 'c', 'k'})
	s.initializeTimelockInstance(instanceIDSeed, 1*time.Second)
	contractID := solanasdk.ContractAddress(s.TimelockProgramID, instanceIDSeed)

	predecessor := [32]byte{}
	salt1 := [32]byte{123}
	salt2 := [32]byte{42}

	go runTimelockWorkerSolana(
		s.T(),
		s.Ctx,
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

	// --- Test scheduling operation ---
	var opID [32]byte

	{
		_, opID = s.scheduleTestIx(instanceIDSeed, predecessor, salt1)
		_, _ = s.scheduleTestIx(instanceIDSeed, predecessor, salt2)

		s.EventuallyWithT(func(collect *assert.CollectT) {
			assert.Equal(collect, logs.FilterMessage("timelock-worker started [solana]").Len(), 1)
			assert.GreaterOrEqual(collect, logs.FilterMessageSnippet("found event scheduled:").Len(), 2)
			assert.GreaterOrEqual(collect, logs.FilterMessageSnippet("event received").Len(), 2)
		}, 10*time.Second, 200*time.Millisecond, logMessages(logs))
	}

	// --- Test canceling the first operation ---
	{
		s.cancelScheduledIx(instanceIDSeed, opID)

		s.EventuallyWithT(func(collect *assert.CollectT) {
			assert.GreaterOrEqual(collect, logs.FilterMessageSnippet("found event cancelled").Len(), 1)
			assert.GreaterOrEqual(collect, logs.FilterMessageSnippet("event received, cancelling operation").Len(), 1)
			assert.GreaterOrEqual(collect, logs.FilterMessageSnippet("nop.delFromScheduler").Len(), 1)
		}, 10*time.Second, 200*time.Millisecond, logMessages(logs))
	}

}

// TestTimelockWorkerExecute tests the Solana timelock worker's ability to execute scheduled operations.
func (s *solanaIntegrationTestSuite) TestTimelockWorkerExecute() {
	// --- arrange ---
	s.withTimeout(30 * time.Second) // FIXME: configure solana-test-validator with shorter slots and faster finalization
	s.T().Setenv("MCMS_SOLANA_MAX_RETRIES", "20")

	e2eutils.FundAccounts(s.T(), s.Ctx, []solana.PublicKey{s.TestPrivateKey.PublicKey()}, 1, s.solanaClient)

	logger, logs := timelockTests.NewTestLogger()
	signer := evmtests.NewTestAccount(s.T())
	pdaSeed := solanasdk.PDASeed(common.BytesToHash([]byte("test-timelock-worker-execute\x00\x00\x00\x00")))
	mcmAddress := solanasdk.ContractAddress(s.McmProgramID, pdaSeed)
	timelockAddress := solanasdk.ContractAddress(s.TimelockProgramID, pdaSeed)
	mcmSignerPDA, _ := solanasdk.FindSignerPDA(s.McmProgramID, pdaSeed)
	timelockSignerPDA, _ := solanasdk.FindTimelockSignerPDA(s.TimelockProgramID, pdaSeed)
	u8ValuePDA, _, _ := solana.FindProgramAddress([][]byte{[]byte("u8_value")}, s.StubProgramID)

	e2eutils.FundAccounts(s.T(), s.Ctx, []solana.PublicKey{mcmSignerPDA, timelockSignerPDA}, 1, s.solanaClient)
	s.initializeMcm(pdaSeed, signer.Address)
	s.initializeTimelockInstance(pdaSeed, time.Second*1)
	s.assignRoleToAccounts(pdaSeed, []solana.PublicKey{mcmSignerPDA}, timelockbindings.Proposer_Role)
	s.assignRoleToAccounts(pdaSeed, []solana.PublicKey{mcmSignerPDA}, timelockbindings.Bypasser_Role)
	s.initializeCPIStub()
	s.initializeRmnRemote()
	s.transferOwnershipRmnRemote(pdaSeed, signer)

	initialStubValue := readCPIStubU8Value(s.Ctx, s.T(), s.solanaClient, u8ValuePDA)

	proposal := solanaProposalWithStubMutInstruction(s.T(), mcmAddress, timelockAddress, s.getMcmOpCount(pdaSeed),
		s.StubProgramID, s.RmnRemoteProgramID, s.ProposerAccessController, s.CancellerAccessController,
		s.BypasserAccessController)
	s.scheduleProposal(proposal, signer.PrivateKey, s.TestPrivateKey)
	s.Log("scheduled proposal")

	// --- act ---
	go runTimelockWorkerSolana(s.T(), s.Ctx, s.solanaBlockchain.Nodes[0].HostHTTPUrl, timelockAddress,
		s.TestPrivateKey.String(), int64(1), int64(1), 10, false, rpc.CommitmentConfirmed, logger)
	s.Log("timelock worker started")

	// --- assert ---
	s.EventuallyWithT(func(collect *assert.CollectT) {
		logEntries := logs.All()
		if !assert.GreaterOrEqual(collect, len(logEntries), 11, "Expected at least 11 log entries") {
			return
		}
		assert.Equal(collect, logs.FilterMessage("added transaction 0 to mcms batch operation b6b1c03b04ff100d1b1e76e3a8cc336c81da06d0d2e5082ca9d50bba261d03c0").Len(), 1)
		assert.Equal(collect, logs.FilterMessageSnippet("execute operation b6b1c03b04ff100d1b1e76e3a8cc336c81da06d0d2e5082ca9d50bba261d03c0 success").Len(), 1)
		assert.Equal(collect, logs.FilterMessage("de-scheduled operation: b6b1c03b04ff100d1b1e76e3a8cc336c81da06d0d2e5082ca9d50bba261d03c0").Len(), 1)
		assert.Equal(collect, logs.FilterMessage("found event executed: 0xb6b1c03b04ff100d1b1e76e3a8cc336c81da06d0d2e5082ca9d50bba261d03c0 (index: 0)").Len(), 1)

		assert.Equal(collect, logs.FilterMessage("added transaction 0 to mcms batch operation ebbdec1a849b9b85194254e17af9dbbbc7b366645327e928e0bc5a6b87fbcf0b").Len(), 1)
		assert.Equal(collect, logs.FilterMessage("added transaction 1 to mcms batch operation ebbdec1a849b9b85194254e17af9dbbbc7b366645327e928e0bc5a6b87fbcf0b").Len(), 1)
		assert.Equal(collect, logs.FilterMessageSnippet("execute operation ebbdec1a849b9b85194254e17af9dbbbc7b366645327e928e0bc5a6b87fbcf0b success").Len(), 1)
		assert.Equal(collect, logs.FilterMessage("de-scheduled operation: ebbdec1a849b9b85194254e17af9dbbbc7b366645327e928e0bc5a6b87fbcf0b").Len(), 1)
		assert.Equal(collect, logs.FilterMessage("found event executed: 0xebbdec1a849b9b85194254e17af9dbbbc7b366645327e928e0bc5a6b87fbcf0b (index: 0)").Len(), 1)
		assert.Equal(collect, logs.FilterMessage("found event executed: 0xebbdec1a849b9b85194254e17af9dbbbc7b366645327e928e0bc5a6b87fbcf0b (index: 1)").Len(), 1)

		finalStubValue := readCPIStubU8Value(s.Ctx, s.T(), s.solanaClient, u8ValuePDA)
		assert.Equal(collect, initialStubValue+1, finalStubValue)
	}, 10*time.Second, 1000*time.Millisecond, logMessages(logs))
}

func solanaProposalWithStubMutInstruction(
	t *testing.T, mcmAddress, timelockAddress string, opCount uint64,
	cpiStubProgramID solana.PublicKey, rmnRemoteProgramID solana.PublicKey,
	proposerAccessController, bypasserAccessController, cancellerAccessController solana.PublicKey,
) *mcms.TimelockProposal {
	t.Helper()

	// FIXME: add multiple batch operations, with multiple instructions, with overlapping accounts
	// (to ensure the code that retrieves the instructions from the operations PDA works correctly)

	mcmProgramID, mcmSeed, err := solanasdk.ParseContractAddress(mcmAddress)
	require.NoError(t, err)
	timelockProgramID, timelockSeed, err := solanasdk.ParseContractAddress(timelockAddress)
	require.NoError(t, err)
	timelockSignerPDA, err := solanasdk.FindTimelockSignerPDA(timelockProgramID, timelockSeed)
	require.NoError(t, err)

	// cpi stub account mut transaction
	u8ValuePDA, _, err := solana.FindProgramAddress([][]byte{[]byte("u8_value")}, cpiStubProgramID)
	require.NoError(t, err)
	accountMutInstruction, err := cpistub.NewAccountMutInstruction(u8ValuePDA, timelockSignerPDA, solana.SystemProgramID).ValidateAndBuild()
	require.NoError(t, err)
	accountMutTransaction, err := solanasdk.NewTransactionFromInstruction(accountMutInstruction, "CPIStub", nil)
	require.NoError(t, err)

	// rmn remote curse transaction
	curseSubject := rmnremotebindings.CurseSubject{Value: [16]byte{'s', 'u', 'b', 'j', 'e', 'c', 't'}}
	configPDA, _, _ := solana.FindProgramAddress([][]byte{[]byte("config")}, rmnRemoteProgramID)
	cursesPDA, _, _ := solana.FindProgramAddress([][]byte{[]byte("curses")}, rmnRemoteProgramID)
	verifyNotCursedInstruction, err := rmnremotebindings.NewVerifyNotCursedInstruction(curseSubject, cursesPDA,
		configPDA).ValidateAndBuild()
	require.NoError(t, err)
	verifyNotCursedTransaction, err := solanasdk.NewTransactionFromInstruction(verifyNotCursedInstruction, "RMNRemote", nil)
	require.NoError(t, err)

	curseInstruction, err := rmnremotebindings.NewCurseInstruction(curseSubject, configPDA, timelockSignerPDA,
		cursesPDA, solana.SystemProgramID).ValidateAndBuild()
	require.NoError(t, err)
	curseTransaction, err := solanasdk.NewTransactionFromInstruction(curseInstruction, "RMNRemote", nil)
	require.NoError(t, err)

	chainMetadata, err := solanasdk.NewChainMetadata(opCount, mcmProgramID, mcmSeed,
		proposerAccessController, cancellerAccessController, bypasserAccessController)
	require.NoError(t, err)

	proposal, err := mcms.NewTimelockProposalBuilder().
		SetValidUntil(uint32(2051222400)). // 2035-01-01T12:00:00 UTC
		SetDescription("proposal to test the timelock proposal converter").
		SetOverridePreviousRoot(true).
		SetVersion("v1").
		SetDelay(mcmstypes.NewDuration(1*time.Second)).
		SetAction(mcmstypes.TimelockActionSchedule).
		SetChainMetadata(map[mcmstypes.ChainSelector]mcmstypes.ChainMetadata{solChainSelector: chainMetadata}).
		AddTimelockAddress(solChainSelector, timelockAddress).
		AddOperation(mcmstypes.BatchOperation{
			ChainSelector: solChainSelector,
			Transactions:  []mcmstypes.Transaction{accountMutTransaction},
		}).
		AddOperation(mcmstypes.BatchOperation{
			ChainSelector: solChainSelector,
			Transactions:  []mcmstypes.Transaction{verifyNotCursedTransaction, curseTransaction},
		}).
		Build()
	require.NoError(t, err)

	return proposal
}

func readCPIStubU8Value(ctx context.Context, t *testing.T, client *rpc.Client, pda solana.PublicKey) uint8 {
	t.Helper()
	var account cpistub.Value
	err := solanacommon.GetAccountDataBorshInto(ctx, client, pda, rpc.CommitmentConfirmed, &account)
	require.NoError(t, err)

	return account.Value
}

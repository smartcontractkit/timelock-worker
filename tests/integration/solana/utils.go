package solana

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"strings"
	"testing"
	"time"

	eth "github.com/ethereum/go-ethereum/common"
	bin "github.com/gagliardetto/binary"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/programs/system"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"

	chainsel "github.com/smartcontractkit/chain-selectors"
	"github.com/smartcontractkit/chainlink-ccip/chains/solana/contracts/tests/testutils"
	"github.com/smartcontractkit/chainlink-ccip/chains/solana/gobindings/access_controller"
	cpistubbindings "github.com/smartcontractkit/chainlink-ccip/chains/solana/gobindings/external_program_cpi_stub"
	mcmbindings "github.com/smartcontractkit/chainlink-ccip/chains/solana/gobindings/mcm"
	rmnremotebindings "github.com/smartcontractkit/chainlink-ccip/chains/solana/gobindings/rmn_remote"
	timelockbindings "github.com/smartcontractkit/chainlink-ccip/chains/solana/gobindings/timelock"
	"github.com/smartcontractkit/chainlink-ccip/chains/solana/utils/accesscontroller"
	"github.com/smartcontractkit/chainlink-ccip/chains/solana/utils/common"
	timelockutils "github.com/smartcontractkit/chainlink-ccip/chains/solana/utils/timelock"
	"github.com/smartcontractkit/mcms"
	mcmssdk "github.com/smartcontractkit/mcms/sdk"
	solanasdk "github.com/smartcontractkit/mcms/sdk/solana"
	mcmstypes "github.com/smartcontractkit/mcms/types"

	"github.com/smartcontractkit/timelock-worker/pkg/timelock"
	evmtests "github.com/smartcontractkit/timelock-worker/tests/integration/evm"
)

var solChainSelector = mcmstypes.ChainSelector(chainsel.SOLANA_DEVNET.Selector)

func (s *solanaIntegrationTestSuite) scheduleProposal(
	proposal *mcms.TimelockProposal, signerKey *ecdsa.PrivateKey, auth solana.PrivateKey,
) {
	converters := map[mcmstypes.ChainSelector]mcmssdk.TimelockConverter{solChainSelector: solanasdk.TimelockConverter{}}
	mcmProposal, _, err := proposal.Convert(s.Ctx, converters)
	s.Require().NoError(err)

	inspectors := map[mcmstypes.ChainSelector]mcmssdk.Inspector{solChainSelector: solanasdk.NewInspector(s.solanaClient)}
	encoders, err := mcmProposal.GetEncoders()
	s.Require().NoError(err)
	encoder := encoders[solChainSelector]
	executor := solanasdk.NewExecutor(encoder.(*solanasdk.Encoder), s.solanaClient, auth)
	executors := map[mcmstypes.ChainSelector]mcmssdk.Executor{solChainSelector: executor}

	signable, err := mcms.NewSignable(&mcmProposal, inspectors)
	s.Require().NoError(err)
	s.Require().NotNil(signable)
	_, err = signable.SignAndAppend(mcms.NewPrivateKeySigner(signerKey))
	s.Require().NoError(err)

	executable, err := mcms.NewExecutable(&mcmProposal, executors)
	s.Require().NoError(err)

	signature, err := executable.SetRoot(s.Ctx, solChainSelector)
	s.Require().NoError(err)
	_, err = solana.SignatureFromBase58(signature.Hash)
	s.Require().NoError(err)
	s.Logf("mcm.SetRoot()")

	for i := range mcmProposal.Operations {
		s.Logf("mcm.Execute(%d)", i)
		signature, err = executable.Execute(s.Ctx, i)
		s.Require().NoError(err)
		_, err = solana.SignatureFromBase58(signature.Hash)
		s.Require().NoError(err)
	}
}

// getBatchAddAccessIxs returns a slice of instructions to batch add access for multiple addresses to a specific role in the Solana timelock instance.
func (s *solanaIntegrationTestSuite) getBatchAddAccessIxs(
	ctx context.Context, timelockID [32]byte, roleAcAccount solana.PublicKey, role timelockbindings.Role,
	addresses []solana.PublicKey, authority solana.PrivateKey, chunkSize int,
) []solana.Instruction {
	var ac access_controller.AccessController
	err := common.GetAccountDataBorshInto(ctx, s.solanaClient, roleAcAccount, rpc.CommitmentConfirmed, &ac)
	s.Require().NoError(err)

	ixs := []solana.Instruction{}
	for i := 0; i < len(addresses); i += chunkSize {
		end := i + chunkSize
		if end > len(addresses) {
			end = len(addresses)
		}
		chunk := addresses[i:end]
		pda, err := solanasdk.FindTimelockConfigPDA(s.TimelockProgramID, timelockID)
		s.Require().NoError(err)

		ix := timelockbindings.NewBatchAddAccessInstruction(
			timelockID,
			role,
			pda,
			s.AccessControllerProgramID,
			roleAcAccount,
			authority.PublicKey(),
		)
		for _, address := range chunk {
			ix.Append(solana.Meta(address))
		}
		vIx, err := ix.ValidateAndBuild()
		s.Require().NoError(err)

		ixs = append(ixs, vIx)
	}

	return ixs
}

// assignRoleToAccounts assigns the specified role to the provided accounts in the Solana timelock instance.
func (s *solanaIntegrationTestSuite) assignRoleToAccounts(
	pdaSeed solanasdk.PDASeed, accounts []solana.PublicKey, role timelockbindings.Role,
) {
	instructions := s.getBatchAddAccessIxs(s.Ctx, pdaSeed, s.RoleMap[role].AccessController.PublicKey(),
		role, accounts, s.TestPrivateKey, 1)
	testutils.SendAndConfirm(s.Ctx, s.T(), s.solanaClient, instructions, s.TestPrivateKey, rpc.CommitmentConfirmed)
}

func (s *solanaIntegrationTestSuite) initializeAccessController(auth solana.PrivateKey) {
	_, s.RoleMap = timelockutils.TestRoleAccounts(2)
	s.ProposerAccessController = s.RoleMap[timelockbindings.Proposer_Role].AccessController.PublicKey()
	s.ExecutorAccessController = s.RoleMap[timelockbindings.Executor_Role].AccessController.PublicKey()
	s.CancellerAccessController = s.RoleMap[timelockbindings.Canceller_Role].AccessController.PublicKey()
	s.BypasserAccessController = s.RoleMap[timelockbindings.Bypasser_Role].AccessController.PublicKey()

	for _, data := range s.RoleMap {
		initAccIxs := s.getInitAccessControllersIxs(s.Ctx, data.AccessController.PublicKey(), auth)

		testutils.SendAndConfirm(s.Ctx, s.T(), s.solanaClient, initAccIxs, auth, rpc.CommitmentConfirmed, common.AddSigners(data.AccessController))

		var ac access_controller.AccessController
		err := common.GetAccountDataBorshInto(s.Ctx, s.solanaClient, data.AccessController.PublicKey(), rpc.CommitmentConfirmed, &ac)
		s.Require().NoError(err)
	}
}

// initializeTimelock initializes a new timelock instance on Solana with the given PDA seed and minimum delay.
// also assigns to the admin account all the roles defined in the RoleMap.
func (s *solanaIntegrationTestSuite) initializeTimelock(pdaSeed solanasdk.PDASeed, minDelay time.Duration) {
	timelockbindings.SetProgramID(s.TimelockProgramID)
	access_controller.SetProgramID(s.AccessControllerProgramID)
	admin := s.TestPrivateKey

	s.initializeAccessController(admin)

	s.Run("init timelock", func() {
		configPDA, err2 := solanasdk.FindTimelockConfigPDA(s.TimelockProgramID, pdaSeed)
		s.Require().NoError(err2)

		initTimelockIx, err3 := timelockbindings.NewInitializeInstruction(
			pdaSeed,
			uint64(minDelay.Seconds()),
			configPDA,
			admin.PublicKey(),
			solana.SystemProgramID,
			s.TimelockProgramID,
			s.getProgramDataAddress(s.TimelockProgramID),
			s.AccessControllerProgramID,
			s.ProposerAccessController,
			s.ExecutorAccessController,
			s.CancellerAccessController,
			s.BypasserAccessController,
		).ValidateAndBuild()
		s.Require().NoError(err3)

		testutils.SendAndConfirm(s.Ctx, s.T(), s.solanaClient, []solana.Instruction{initTimelockIx}, admin, rpc.CommitmentConfirmed)

		var configAccount timelockbindings.Config
		err := common.GetAccountDataBorshInto(s.Ctx, s.solanaClient, configPDA, rpc.CommitmentConfirmed, &configAccount)
		s.Require().NoError(err, "Failed to get timelock config account data")

		s.Require().Equal(admin.PublicKey(), configAccount.Owner, "Owner doesn't match")
		s.Require().Equal(uint64(minDelay.Seconds()), configAccount.MinDelay, "MinDelay doesn't match")
		s.Require().Equal(s.ProposerAccessController, configAccount.ProposerRoleAccessController, "ProposerRoleAccessController doesn't match")
		s.Require().Equal(s.ExecutorAccessController, configAccount.ExecutorRoleAccessController, "ExecutorRoleAccessController doesn't match")
		s.Require().Equal(s.CancellerAccessController, configAccount.CancellerRoleAccessController, "CancellerRoleAccessController doesn't match")
		s.Require().Equal(s.BypasserAccessController, configAccount.BypasserRoleAccessController, "BypasserRoleAccessController doesn't match")
	})

	s.Run("setup roles", func() {
		for _, roleAccounts := range s.RoleMap {
			accounts := make([]solana.PublicKey, 0, len(roleAccounts.Accounts))
			for _, account := range roleAccounts.Accounts {
				accounts = append(accounts, account.PublicKey())
			}

			s.assignRoleToAccounts(pdaSeed, accounts, roleAccounts.Role)

			for _, account := range roleAccounts.Accounts {
				found, err := accesscontroller.HasAccess(s.Ctx, s.solanaClient, roleAccounts.AccessController.PublicKey(),
					account.PublicKey(), rpc.CommitmentConfirmed)
				s.Require().NoError(err)
				s.Require().True(found, "Account %s not found in %s AccessList", account.PublicKey(), roleAccounts.Role)
			}
		}
	})
}

// initializeMcm initializes a new mcm instance on Solana with the given PDA seed.
func (s *solanaIntegrationTestSuite) initializeMcm(pdaSeed [32]byte, signer eth.Address) {
	mcmbindings.SetProgramID(s.McmProgramID)
	auth, err := solana.PrivateKeyFromBase58(privateKey)
	s.Require().NoError(err)

	configPDA, err := solanasdk.FindConfigPDA(s.McmProgramID, pdaSeed)
	s.Require().NoError(err)
	rootMetadataPDA, err := solanasdk.FindRootMetadataPDA(s.McmProgramID, pdaSeed)
	s.Require().NoError(err)
	expiringRootAndOpCountPDA, err := solanasdk.FindExpiringRootAndOpCountPDA(s.McmProgramID, pdaSeed)
	s.Require().NoError(err)
	chainSelector := chainsel.SOLANA_DEVNET.Selector

	ix, err := mcmbindings.NewInitializeInstruction(chainSelector, pdaSeed, configPDA, auth.PublicKey(),
		solana.SystemProgramID, s.McmProgramID, s.getProgramDataAddress(s.McmProgramID), rootMetadataPDA,
		expiringRootAndOpCountPDA).ValidateAndBuild()
	s.Require().NoError(err)

	_, err = sendAndConfirm(s.Ctx, s.T(), s.solanaClient, auth, []solana.Instruction{ix})
	s.Require().NoError(err)

	// get config and validate
	var configAccount mcmbindings.MultisigConfig
	err = common.GetAccountDataBorshInto(s.Ctx, s.solanaClient, configPDA, rpc.CommitmentConfirmed, &configAccount)
	s.Require().NoError(err)

	s.Require().Equal(chainSelector, configAccount.ChainId)
	s.Require().Equal(auth.PublicKey(), configAccount.Owner)

	mcmConfig := mcmstypes.Config{Quorum: 1, Signers: []eth.Address{signer}}
	configurer := solanasdk.NewConfigurer(s.solanaClient, auth, mcmstypes.ChainSelector(chainSelector))
	_, err = configurer.SetConfig(s.Ctx, solanasdk.ContractAddress(s.McmProgramID, pdaSeed), &mcmConfig, true)
	s.Require().NoError(err)
}

// getInitAccessControllersIxs returns the instructions to initialize access controllers on Solana.
func (s *solanaIntegrationTestSuite) getInitAccessControllersIxs(ctx context.Context, roleAcAccount solana.PublicKey, authority solana.PrivateKey) []solana.Instruction {
	ixs := []solana.Instruction{}

	dataSize := uint64(8 + 32 + 32 + ((32 * 64) + 8)) // discriminator + owner + proposed owner + access_list (64 max addresses + length)
	rentExemption, err := s.solanaClient.GetMinimumBalanceForRentExemption(ctx, dataSize, rpc.CommitmentConfirmed)
	s.Require().NoError(err)

	createAccountInstruction, err := system.NewCreateAccountInstruction(rentExemption, dataSize,
		s.AccessControllerProgramID, authority.PublicKey(), roleAcAccount).ValidateAndBuild()
	s.Require().NoError(err)
	ixs = append(ixs, createAccountInstruction)

	initializeInstruction, err := access_controller.NewInitializeInstruction(roleAcAccount,
		authority.PublicKey()).ValidateAndBuild()
	s.Require().NoError(err)
	ixs = append(ixs, initializeInstruction)

	return ixs
}

func (s *solanaIntegrationTestSuite) initializeCPIStub() {
	cpistubbindings.SetProgramID(s.StubProgramID)

	admin, err := solana.PrivateKeyFromBase58(privateKey)
	s.Require().NoError(err)

	valuePDA, _, err := solana.FindProgramAddress([][]byte{[]byte("u8_value")}, s.StubProgramID)
	s.Require().NoError(err)

	instruction, err := cpistubbindings.NewInitializeInstruction(valuePDA, admin.PublicKey(), solana.SystemProgramID).
		ValidateAndBuild()
	s.Require().NoError(err)

	_, err = sendAndConfirm(s.Ctx, s.T(), s.solanaClient, admin, []solana.Instruction{instruction})
	s.Require().NoError(err)
}

func (s *solanaIntegrationTestSuite) initializeRmnRemote() {
	rmnremotebindings.SetProgramID(s.RmnRemoteProgramID)

	auth, err := solana.PrivateKeyFromBase58(privateKey)
	s.Require().NoError(err)
	configPDA, _, err := solana.FindProgramAddress([][]byte{[]byte("config")}, s.RmnRemoteProgramID)
	s.Require().NoError(err)
	cursesPDA, _, err := solana.FindProgramAddress([][]byte{[]byte("curses")}, s.RmnRemoteProgramID)
	s.Require().NoError(err)
	// s.Logf("ACCOUNTS:\n%v\n%v\n%v\n%v\n%v\n%v\n", configPDA, cursesPDA, auth.PublicKey(),
	// 	solana.SystemProgramID, s.RmnRemoteProgramID, s.getProgramDataAddress(s.RmnRemoteProgramID))

	instruction, err := rmnremotebindings.NewInitializeInstruction(configPDA, cursesPDA, auth.PublicKey(),
		solana.SystemProgramID, s.RmnRemoteProgramID, s.getProgramDataAddress(s.RmnRemoteProgramID)).ValidateAndBuild()
	s.Require().NoError(err)

	_, err = sendAndConfirm(s.Ctx, s.T(), s.solanaClient, auth, []solana.Instruction{instruction})
	s.Require().NoError(err)
}

func (s *solanaIntegrationTestSuite) transferOwnershipRmnRemote(seed solanasdk.PDASeed, signer evmtests.TestAccount) {
	configPDA, _, err := solana.FindProgramAddress([][]byte{[]byte("config")}, s.RmnRemoteProgramID)
	s.Require().NoError(err)
	cursesPDA, _, err := solana.FindProgramAddress([][]byte{[]byte("curses")}, s.RmnRemoteProgramID)
	s.Require().NoError(err)
	timelockSignerPDA, err := solanasdk.FindTimelockSignerPDA(s.TimelockProgramID, seed)
	s.Require().NoError(err)

	transferInstruction, err := rmnremotebindings.NewTransferOwnershipInstruction(timelockSignerPDA, configPDA,
		cursesPDA, s.TestPrivateKey.PublicKey()).ValidateAndBuild()
	s.Require().NoError(err)

	acceptInstruction, err := rmnremotebindings.NewAcceptOwnershipInstruction(configPDA,
		timelockSignerPDA).ValidateAndBuild()
	s.Require().NoError(err)

	s.transferOwnership(transferInstruction, acceptInstruction, seed, signer)
	s.Logf("transferred ownership of RMNRemote to timelock")
}

func (s *solanaIntegrationTestSuite) transferOwnership(
	transferInstruction, acceptInstruction solana.Instruction, seed solanasdk.PDASeed, signer evmtests.TestAccount,
) {
	// --- transfer ---
	_, err := sendAndConfirm(s.Ctx, s.T(), s.solanaClient, s.TestPrivateKey, []solana.Instruction{transferInstruction})
	s.Require().NoError(err)

	// --- accept ---
	acceptOwnershipTransaction, err := solanasdk.NewTransactionFromInstruction(acceptInstruction, "", nil)
	s.Require().NoError(err)

	chainMetadata, err := solanasdk.NewChainMetadata(s.getMcmOpCount(seed), s.McmProgramID, seed,
		s.ProposerAccessController, s.CancellerAccessController, s.BypasserAccessController)
	s.Require().NoError(err)
	timelockAddress := solanasdk.ContractAddress(s.TimelockProgramID, seed)

	proposal, err := mcms.NewTimelockProposalBuilder().
		SetValidUntil(uint32(2051222400)). // 2035-01-01T12:00:00 UTC
		SetDescription("proposal to accept ownership").
		SetOverridePreviousRoot(true).
		SetVersion("v1").
		SetAction(mcmstypes.TimelockActionBypass).
		SetChainMetadata(map[mcmstypes.ChainSelector]mcmstypes.ChainMetadata{solChainSelector: chainMetadata}).
		AddTimelockAddress(solChainSelector, timelockAddress).
		AddOperation(mcmstypes.BatchOperation{
			ChainSelector: solChainSelector,
			Transactions:  []mcmstypes.Transaction{acceptOwnershipTransaction},
		}).
		Build()
	s.Require().NoError(err)

	s.scheduleProposal(proposal, signer.PrivateKey, s.TestPrivateKey)
}

func (s *solanaIntegrationTestSuite) getMcmOpCount(seed solanasdk.PDASeed) uint64 {
	inspector := solanasdk.NewInspector(s.solanaClient)
	opCount, err := inspector.GetOpCount(s.Ctx, solanasdk.ContractAddress(s.McmProgramID, seed))
	s.Require().NoError(err)

	return opCount
}

// runTimelockWorkerSolana runs the Solana timelock worker with the provided parameters.
func runTimelockWorkerSolana(
	t *testing.T,
	ctx context.Context,
	nodeURL, timelockAddress, privateKey string,
	pollPeriod, listenerPollPeriod int64,
	listenerPollSize int,
	dryRun bool,
	commitmentType rpc.CommitmentType,
	logger *zap.Logger,
) {
	t.Helper()

	t.Logf("TimelockWorker.Listen(%v, %v, %v, %v, %v, %v, %v, %v)", nodeURL, timelockAddress, privateKey, pollPeriod, listenerPollPeriod, listenerPollSize, dryRun, commitmentType)
	timelockWorker, err := timelock.NewTimelockWorkerSolana(nodeURL, timelockAddress, privateKey, pollPeriod, listenerPollPeriod, listenerPollSize, dryRun, commitmentType, logger.Sugar())
	require.NoError(t, err)
	require.NotNil(t, timelockWorker)

	err = timelockWorker.Listen(ctx)
	require.NoError(t, err)
}

func (s *solanaIntegrationTestSuite) scheduleTestIx(
	timelockID solanasdk.PDASeed,
	predecessor,
	salt [32]byte,
) (solana.Instruction, [32]byte) {
	cpistubbindings.SetProgramID(s.StubProgramID)
	timelockbindings.SetProgramID(s.TimelockProgramID)
	ixStub, err := cpistubbindings.NewU8InstructionDataInstruction(uint8(10)).ValidateAndBuild()
	s.Require().NoError(err, "Failed to create instruction data instruction")
	ixData, err := ixStub.Data()
	s.Require().NoError(err)
	accounts := make([]timelockbindings.InstructionAccount, len(ixStub.Accounts()))
	for i, account := range ixStub.Accounts() {
		accounts[i] = timelockbindings.InstructionAccount{
			Pubkey:     account.PublicKey,
			IsSigner:   account.IsSigner,
			IsWritable: account.IsWritable,
		}
	}
	opInstructions := []timelockbindings.InstructionData{{Data: ixData, ProgramId: ixStub.ProgramID(), Accounts: accounts}}
	operationID, err := solanasdk.HashOperation(opInstructions, predecessor, salt)
	s.Require().NoError(err)
	operationPDA, err := solanasdk.FindTimelockOperationPDA(s.TimelockProgramID, timelockID, operationID)
	s.Require().NoError(err)
	configPDA, err := solanasdk.FindTimelockConfigPDA(s.TimelockProgramID, timelockID)
	s.Require().NoError(err)
	proposerAC := s.RoleMap[timelockbindings.Proposer_Role].AccessController.PublicKey()
	ixs := []solana.Instruction{}
	initOpIx, err := timelockbindings.NewInitializeOperationInstruction(
		timelockID,
		operationID,
		predecessor,
		salt,
		uint32(len(opInstructions)), // #nosec G115
		operationPDA,
		configPDA,
		proposerAC,
		s.TestPrivateKey.PublicKey(),
		solana.SystemProgramID,
	).ValidateAndBuild()
	s.Require().NoError(err)
	ixs = append(ixs, initOpIx)
	for i, ixData := range opInstructions {
		initializeIx, errAppend := timelockbindings.NewInitializeInstructionInstruction(
			timelockID,
			operationID,
			ixData.ProgramId,
			ixData.Accounts,
			operationPDA,
			configPDA,
			proposerAC,
			s.TestPrivateKey.PublicKey(),
			solana.SystemProgramID,
		).ValidateAndBuild()
		s.Require().NoError(errAppend)
		ixs = append(ixs, initializeIx)

		rawData := ixData.Data
		offset := 0

		for offset < len(rawData) {
			end := offset + solanasdk.AppendIxDataChunkSize
			if end > len(rawData) {
				end = len(rawData)
			}
			chunk := rawData[offset:end]

			appendIx, appendErr := timelockbindings.NewAppendInstructionDataInstruction(
				timelockID,
				operationID,
				uint32(i), // #nosec G115
				chunk,
				operationPDA,
				configPDA,
				proposerAC,
				s.TestPrivateKey.PublicKey(),
				solana.SystemProgramID,
			).ValidateAndBuild()
			s.Require().NoError(appendErr)
			ixs = append(ixs, appendIx)
			offset = end
		}
	}
	// Finalize Operation
	finOpIx, err := timelockbindings.NewFinalizeOperationInstruction(
		timelockID,
		operationID,
		operationPDA,
		configPDA,
		proposerAC,
		s.TestPrivateKey.PublicKey(),
	).ValidateAndBuild()
	s.Require().NoError(err)
	ixs = append(ixs, finOpIx)
	testutils.SendAndConfirm(s.Ctx, s.T(), s.solanaClient, ixs, s.TestPrivateKey, rpc.CommitmentConfirmed)

	// Schedule the operation
	scheduleIx, err := timelockbindings.NewScheduleBatchInstruction(
		timelockID,
		operationID,
		1,
		operationPDA,
		configPDA,
		s.RoleMap[timelockbindings.Proposer_Role].AccessController.PublicKey(),
		s.TestPrivateKey.PublicKey(),
	).ValidateAndBuild()
	s.Require().NoError(err)
	res := testutils.SendAndConfirm(s.Ctx, s.T(), s.solanaClient, []solana.Instruction{scheduleIx}, s.TestPrivateKey, rpc.CommitmentConfirmed)
	s.Require().Nil(res.Meta.Err, "Transaction failed in program execution: %v", res.Meta.Err)
	var txFinalized *rpc.GetTransactionResult
	tx, err := res.Transaction.GetTransaction()
	s.Require().NoError(err, "Failed to decode transaction")
	s.Require().NotNil(tx, "Decoded transaction is nil")
	s.Require().NotEmpty(tx.Signatures, "Transaction contains no signatures")

	sig := tx.Signatures[0]
	s.EventuallyWithT(func(c *assert.CollectT) {
		var err error
		txFinalized, err = s.solanaClient.GetTransaction(
			s.Ctx,
			sig,
			&rpc.GetTransactionOpts{
				Encoding:   solana.EncodingBase64,
				Commitment: rpc.CommitmentFinalized,
			},
		)
		assert.NoError(c, err)
		assert.NotNil(c, txFinalized)
	}, 30*time.Second, 500*time.Millisecond)

	return ixStub, operationID
}

func (s *solanaIntegrationTestSuite) cancelScheduledIx(
	timelockID solanasdk.PDASeed,
	operationID [32]byte,
) {
	timelock2.SetProgramID(s.TimelockProgramID)

	operationPDA, err := solanasdk.FindTimelockOperationPDA(s.TimelockProgramID, timelockID, operationID)
	s.Require().NoError(err)

	configPDA, err := solanasdk.FindTimelockConfigPDA(s.TimelockProgramID, timelockID)
	s.Require().NoError(err)

	proposerAC := s.RoleMap[timelock2.Proposer_Role].AccessController.PublicKey()

	cancelIx, err := timelock2.NewCancelInstruction(
		timelockID,
		operationID,
		operationPDA,
		configPDA,
		proposerAC,
		s.TestPrivateKey.PublicKey(),
	).ValidateAndBuild()
	s.Require().NoError(err)

	res := testutils.SendAndConfirm(s.Ctx, s.T(), s.solanaClient, []solana.Instruction{cancelIx}, s.TestPrivateKey, rpc.CommitmentConfirmed)
	s.Require().Nil(res.Meta.Err, "Cancellation transaction failed: %v", res.Meta.Err)
}

func (s *solanaIntegrationTestSuite) executeScheduledIx(
	timelockID solanasdk.PDASeed,
	operationID [32]byte,
	scheduledIx solana.Instruction,
) {
	timelock2.SetProgramID(s.TimelockProgramID)

	operationPDA, err := solanasdk.FindTimelockOperationPDA(s.TimelockProgramID, timelockID, operationID)
	s.Require().NoError(err)

	predecessorOp := solana.PublicKey{} // empty if predecessor is [32]byte{}
	configPDA, err := solanasdk.FindTimelockConfigPDA(s.TimelockProgramID, timelockID)
	s.Require().NoError(err)

	timelockSignerPDA, err := solanasdk.FindTimelockSignerPDA(s.TimelockProgramID, timelockID)
	s.Require().NoError(err)

	proposerAC := s.RoleMap[timelock2.Proposer_Role].AccessController.PublicKey()

	// Build ExecuteBatch instruction with remaining accounts
	executeIxBuilder := timelock2.NewExecuteBatchInstruction(
		timelockID,
		operationID,
		operationPDA,
		predecessorOp,
		configPDA,
		timelockSignerPDA,
		proposerAC,
		s.TestPrivateKey.PublicKey(),
	)

	executeIxBuilder.Append(&solana.AccountMeta{
		PublicKey: s.StubProgramID,
	})

	executeIx, err := executeIxBuilder.ValidateAndBuild()
	s.Require().NoError(err)

	res := testutils.SendAndConfirm(s.Ctx, s.T(), s.solanaClient, []solana.Instruction{executeIx}, s.TestPrivateKey, rpc.CommitmentConfirmed)
	s.Require().Nil(res.Meta.Err, "Execution transaction failed: %v", res.Meta.Err)
}

func (s *solanaIntegrationTestSuite) getProgramDataAddress(programID solana.PublicKey) solana.PublicKey {
	opts := &rpc.GetAccountInfoOpts{Commitment: rpc.CommitmentConfirmed}
	data, accErr := s.solanaClient.GetAccountInfoWithOpts(s.Ctx, programID, opts)
	s.Require().NoError(accErr)

	var programData struct {
		DataType uint32
		Address  solana.PublicKey
	}
	err := bin.UnmarshalBorsh(&programData, data.Bytes())
	s.Require().NoError(err)

	return programData.Address
}

func sendAndConfirm(
	ctx context.Context, t *testing.T, client *rpc.Client, txPayer solana.PrivateKey, instructions []solana.Instruction,
) (solana.Signature, error) {
	t.Helper()

	hashResult, err := client.GetLatestBlockhash(ctx, rpc.CommitmentConfirmed)
	require.NoError(t, err)

	tx, err := solana.NewTransaction(instructions, hashResult.Value.Blockhash, solana.TransactionPayer(txPayer.PublicKey()))
	require.NoError(t, err)

	signers := map[solana.PublicKey]solana.PrivateKey{txPayer.PublicKey(): txPayer}
	_, err = tx.Sign(func(pub solana.PublicKey) *solana.PrivateKey {
		priv, ok := signers[pub]
		require.True(t, ok)

		return &priv
	})
	require.NoError(t, err)

	txOpts := rpc.TransactionOpts{SkipPreflight: false, PreflightCommitment: rpc.CommitmentConfirmed}
	signature := solana.Signature{}
	require.EventuallyWithT(t, func(collect *assert.CollectT) {
		var err error
		signature, err = client.SendTransactionWithOpts(ctx, tx, txOpts)
		assert.NoError(t, err)
	}, 1*time.Second, 50*time.Millisecond)

	// wait for the transaction to be confirmed or finalized
	ticker := time.NewTicker(50 * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			require.NoError(t, fmt.Errorf("unable to confirm transaction within timeout"))
		case <-ticker.C:
			statusRes, sigErr := client.GetSignatureStatuses(ctx, true, signature)
			require.NoError(t, sigErr)
			require.NotNil(t, statusRes)
			require.GreaterOrEqual(t, len(statusRes.Value), 1)

			if statusRes.Value[0] != nil &&
				(statusRes.Value[0].ConfirmationStatus == rpc.ConfirmationStatusConfirmed ||
					statusRes.Value[0].ConfirmationStatus == rpc.ConfirmationStatusFinalized) {
				return signature, nil
			}
		}
	}
}

func logMessages(logs *observer.ObservedLogs) string {
	entries := lo.Map(logs.All(), func(l observer.LoggedEntry, _ int) string { return logEntryString(l) })
	return fmt.Sprintf("LOG MESSAGES:\n%v\n", strings.Join(entries, "\n"))
}

func logEntryString(logEntry observer.LoggedEntry) string {
	output := logEntry.Message
	for key, value := range logEntry.ContextMap() {
		output += fmt.Sprintf(" %v=%v", key, value)
	}

	return output
}

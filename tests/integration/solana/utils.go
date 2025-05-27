package solana

import (
	"context"
	"testing"
	"time"

	bin "github.com/gagliardetto/binary"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/programs/system"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/smartcontractkit/chainlink-ccip/chains/solana/contracts/tests/testutils"
	"github.com/smartcontractkit/chainlink-ccip/chains/solana/gobindings/access_controller"
	"github.com/smartcontractkit/chainlink-ccip/chains/solana/gobindings/external_program_cpi_stub"
	timelock2 "github.com/smartcontractkit/chainlink-ccip/chains/solana/gobindings/timelock"
	"github.com/smartcontractkit/chainlink-ccip/chains/solana/utils/accesscontroller"
	"github.com/smartcontractkit/chainlink-ccip/chains/solana/utils/common"
	solanasdk "github.com/smartcontractkit/mcms/sdk/solana"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/smartcontractkit/timelock-worker/pkg/timelock"
)

// getBatchAddAccessIxs returns a slice of instructions to batch add access for multiple addresses to a specific role in the Solana timelock instance.
func (s *solanaIntegrationTestSuite) getBatchAddAccessIxs(
	ctx context.Context, timelockID [32]byte, roleAcAccount solana.PublicKey, role timelock2.Role,
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

		ix := timelock2.NewBatchAddAccessInstruction(
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

// AssignRoleToAccounts assigns the specified role to the provided accounts in the Solana timelock instance.
func (s *solanaIntegrationTestSuite) AssignRoleToAccounts(
	ctx context.Context, pdaSeed solanasdk.PDASeed, auth solana.PrivateKey,
	accounts []solana.PublicKey, role timelock2.Role,
) {
	instructions := s.getBatchAddAccessIxs(ctx, pdaSeed, s.RoleMap[role].AccessController.PublicKey(),
		role, accounts, auth, 1)
	testutils.SendAndConfirm(ctx, s.T(), s.solanaClient, instructions, auth, rpc.CommitmentConfirmed)
}

// initializeTimelockInstance initializes a new timelock instance on Solana with the given PDA seed and minimum delay.
// also assigns to the admin account all the roles defined in the RoleMap.
func (s *solanaIntegrationTestSuite) initializeTimelockInstance(pdaSeed solanasdk.PDASeed, minDelay time.Duration) {
	t := s.T()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	t.Cleanup(cancel)
	timelock2.SetProgramID(s.TimelockProgramID)
	access_controller.SetProgramID(s.AccessControllerProgramID)
	admin := s.TestPrivateKey
	t.Run("init access controller", func(t *testing.T) {
		for _, data := range s.RoleMap {
			initAccIxs := s.getInitAccessControllersIxs(ctx, data.AccessController.PublicKey(), admin)

			testutils.SendAndConfirm(ctx, t, s.solanaClient, initAccIxs, admin, rpc.CommitmentConfirmed, common.AddSigners(data.AccessController))

			var ac access_controller.AccessController
			err := common.GetAccountDataBorshInto(ctx, s.solanaClient, data.AccessController.PublicKey(), rpc.CommitmentConfirmed, &ac)
			s.Require().NoError(err, "Failed to get access controller account data")
		}
	})
	s.Run("init timelock", func() {
		data, accErr := s.solanaClient.GetAccountInfoWithOpts(ctx, s.TimelockProgramID, &rpc.GetAccountInfoOpts{
			Commitment: rpc.CommitmentConfirmed,
		})
		s.Require().NoError(accErr)

		var programData struct {
			DataType uint32
			Address  solana.PublicKey
		}
		s.Require().NoError(bin.UnmarshalBorsh(&programData, data.Bytes()))

		pda, err2 := solanasdk.FindTimelockConfigPDA(s.TimelockProgramID, pdaSeed)
		s.Require().NoError(err2)

		initTimelockIx, err3 := timelock2.NewInitializeInstruction(
			pdaSeed,
			uint64(minDelay.Seconds()),
			pda,
			admin.PublicKey(),
			solana.SystemProgramID,
			s.TimelockProgramID,
			programData.Address,
			s.AccessControllerProgramID,
			s.RoleMap[timelock2.Proposer_Role].AccessController.PublicKey(),
			s.RoleMap[timelock2.Executor_Role].AccessController.PublicKey(),
			s.RoleMap[timelock2.Canceller_Role].AccessController.PublicKey(),
			s.RoleMap[timelock2.Bypasser_Role].AccessController.PublicKey(),
		).ValidateAndBuild()
		s.Require().NoError(err3)

		testutils.SendAndConfirm(ctx, s.T(), s.solanaClient, []solana.Instruction{initTimelockIx}, admin, rpc.CommitmentConfirmed)

		var configAccount timelock2.Config
		err := common.GetAccountDataBorshInto(ctx, s.solanaClient, pda, rpc.CommitmentConfirmed, &configAccount)
		s.Require().NoError(err, "Failed to get timelock config account data")

		s.Require().Equal(admin.PublicKey(), configAccount.Owner, "Owner doesn't match")
		s.Require().Equal(uint64(minDelay.Seconds()), configAccount.MinDelay, "MinDelay doesn't match")
		s.Require().Equal(s.RoleMap[timelock2.Proposer_Role].AccessController.PublicKey(), configAccount.ProposerRoleAccessController, "ProposerRoleAccessController doesn't match")
		s.Require().Equal(s.RoleMap[timelock2.Executor_Role].AccessController.PublicKey(), configAccount.ExecutorRoleAccessController, "ExecutorRoleAccessController doesn't match")
		s.Require().Equal(s.RoleMap[timelock2.Canceller_Role].AccessController.PublicKey(), configAccount.CancellerRoleAccessController, "CancellerRoleAccessController doesn't match")
		s.Require().Equal(s.RoleMap[timelock2.Bypasser_Role].AccessController.PublicKey(), configAccount.BypasserRoleAccessController, "BypasserRoleAccessController doesn't match")
	})
	s.Run("setup roles", func() {
		for _, roleAccounts := range s.RoleMap {
			accounts := make([]solana.PublicKey, 0, len(roleAccounts.Accounts))
			for _, account := range roleAccounts.Accounts {
				accounts = append(accounts, account.PublicKey())
			}

			s.AssignRoleToAccounts(ctx, pdaSeed, admin, accounts, roleAccounts.Role)

			for _, account := range roleAccounts.Accounts {
				found, err := accesscontroller.HasAccess(ctx, s.solanaClient, roleAccounts.AccessController.PublicKey(),
					account.PublicKey(), rpc.CommitmentConfirmed)
				s.Require().NoError(err)
				s.Require().True(found, "Account %s not found in %s AccessList", account.PublicKey(), roleAccounts.Role)
			}
		}
	})
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

// runTimelockWorkerSolana runs the Solana timelock worker with the provided parameters.
func runTimelockWorkerSolana(
	t *testing.T,
	ctx context.Context,
	nodeURL, timelockAddress, privateKey string,
	pollPeriod, listenerPollPeriod int64,
	listenerPollSize int,
	dryRun bool,
	logger *zap.Logger,
) {
	t.Helper()

	t.Logf("TimelockWorker.Listen(%v, %v, %v, %v, %v, %v)", nodeURL, timelockAddress, privateKey, pollPeriod, listenerPollPeriod, listenerPollSize)
	timelockWorker, err := timelock.NewTimelockWorkerSolana(nodeURL, timelockAddress, privateKey, pollPeriod, listenerPollPeriod, listenerPollSize, dryRun, logger.Sugar())
	require.NoError(t, err)
	require.NotNil(t, timelockWorker)

	err = timelockWorker.Listen(ctx)
	require.NoError(t, err)
}

func (s *solanaIntegrationTestSuite) scheduleTestIx(
	timelockID solanasdk.PDASeed,
	predecessor,
	salt [32]byte) (solana.Instruction, [32]byte) {
	external_program_cpi_stub.SetProgramID(s.StubProgramID)
	ixStub, err := external_program_cpi_stub.NewU8InstructionDataInstruction(uint8(10)).ValidateAndBuild()
	s.Require().NoError(err, "Failed to create instruction data instruction")
	ixData, err := ixStub.Data()
	s.Require().NoError(err)
	accounts := make([]timelock2.InstructionAccount, len(ixStub.Accounts()))
	for i, account := range ixStub.Accounts() {
		accounts[i] = timelock2.InstructionAccount{
			Pubkey:     account.PublicKey,
			IsSigner:   account.IsSigner,
			IsWritable: account.IsWritable,
		}
	}
	opInstructions := []timelock2.InstructionData{{Data: ixData, ProgramId: ixStub.ProgramID(), Accounts: accounts}}
	operationID, err := solanasdk.HashOperation(opInstructions, predecessor, salt)
	s.Require().NoError(err)
	operationPDA, err := solanasdk.FindTimelockOperationPDA(s.TimelockProgramID, timelockID, operationID)
	s.Require().NoError(err)
	configPDA, err := solanasdk.FindTimelockConfigPDA(s.TimelockProgramID, timelockID)
	s.Require().NoError(err)
	proposerAC := s.RoleMap[timelock2.Proposer_Role].AccessController.PublicKey()
	ixs := []solana.Instruction{}
	initOpIx, err := timelock2.NewInitializeOperationInstruction(
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
		initializeIx, errAppend := timelock2.NewInitializeInstructionInstruction(
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

			appendIx, appendErr := timelock2.NewAppendInstructionDataInstruction(
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
	finOpIx, err := timelock2.NewFinalizeOperationInstruction(
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
	scheduleIx, err := timelock2.NewScheduleBatchInstruction(
		timelockID,
		operationID,
		1,
		operationPDA,
		configPDA,
		s.RoleMap[timelock2.Proposer_Role].AccessController.PublicKey(),
		s.TestPrivateKey.PublicKey(),
	).ValidateAndBuild()
	s.Require().NoError(err)
	testutils.SendAndConfirm(s.Ctx, s.T(), s.solanaClient, []solana.Instruction{scheduleIx}, s.TestPrivateKey, rpc.CommitmentConfirmed)

	return ixStub, operationID
}

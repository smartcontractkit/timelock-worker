package timelock

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/samber/lo"
	mcmssolanasdk "github.com/smartcontractkit/mcms/sdk/solana"
	mcmstypes "github.com/smartcontractkit/mcms/types"

	timelockbindings "github.com/smartcontractkit/chainlink-ccip/chains/solana/gobindings/timelock"
	solanautils "github.com/smartcontractkit/chainlink-ccip/chains/solana/utils/common"
)

func (tw *WorkerSolana) execute(ctx context.Context, op []TimelockCallScheduled) {
	if len(op) == 0 {
		tw.logger.Warn("no calls given")
		return
	}
	opId := op[0].Id()

	isReady, err := tw.inspector.IsOperationReady(ctx, tw.timelockFullAddress, opId)
	if err != nil {
		tw.logger.Errorf("unable to read operation %x \"ready\" status: %s", opId, err.Error())
		return
	}
	if !isReady {
		tw.logger.Infof("skipping operation %x: not ready", opId)
		return
	}

	batchOps, err := tw.mapCallScheduledEventsToMcmsBatchOperations(ctx, op)
	if err != nil {
		tw.logger.Errorf("unable to convert call scheduled events to mcms transactions: %v", err)
		return
	}

	timelockExecutor := mcmssolanasdk.NewTimelockExecutor(tw.solanaClient, tw.privateKey)

	for i, batchOp := range batchOps {
		tw.logger.Debugf("execute operation %x", op[i].Id())
		result, err := timelockExecutor.Execute(ctx, batchOp, tw.timelockFullAddress, op[i].Predecessor(), op[i].Salt())
		if err != nil {
			tw.logger.Errorf("execute operation %x error: %s", opId, err.Error())
		} else {
			tw.logger.Infof("execute operation %x success: %s", op[i].Id(), result.Hash)
		}
	}
}

func (tw *WorkerSolana) mapCallScheduledEventsToMcmsBatchOperations(
	ctx context.Context, events []TimelockCallScheduled,
) ([]mcmstypes.BatchOperation, error) {
	// transactions := make([]mcmstypes.Transaction, len(events))
	batchOps := make([]mcmstypes.BatchOperation, len(events))
	for i, event := range events {
		solanaEvent := event.(*solanaTimelockCallScheduled).callScheduledEvent

		transactions, err := tw.getTransactionsInBatchOperation(ctx, solanaEvent.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to get transactions in solana event 0x%x: %w", solanaEvent.ID, err)
		}
		batchOps[i] = mcmstypes.BatchOperation{Transactions: transactions}
	}

	return batchOps, nil
}

func (tw *WorkerSolana) getTransactionsInBatchOperation(
	ctx context.Context, operationID operationKey,
) ([]mcmstypes.Transaction, error) {
	operationPDA, err := mcmssolanasdk.FindTimelockOperationPDA(tw.timelockProgramKey, tw.instanceSeed, operationID)
	if err != nil {
		return nil, fmt.Errorf("failed to find operation PDA: %w", err)
	}

	var scheduledOperation timelockbindings.Operation
	err = solanautils.GetAccountDataBorshInto(ctx, tw.solanaClient, operationPDA, rpc.CommitmentConfirmed, &scheduledOperation)
	if err != nil {
		return nil, fmt.Errorf("failed to get operation data from PDA account: %w", err)
	}
	tw.logger.Debugf("scheduled operation pda: %#v", scheduledOperation)

	mcmsTxs := make([]mcmstypes.Transaction, len(scheduledOperation.Instructions))

	for i, instruction := range scheduledOperation.Instructions {
		additionalFields := mcmssolanasdk.AdditionalFields{Value: nil /* REVIEW */}
		for _, account := range instruction.Accounts {
			additionalFields.Accounts = append(additionalFields.Accounts, &solana.AccountMeta{
				PublicKey:  account.Pubkey,
				IsWritable: account.IsWritable,
				IsSigner:   account.IsSigner,
			})
		}
		accountsStr := strings.Join(lo.Map(instruction.Accounts, func(account timelockbindings.InstructionAccount, _ int) string {
			return fmt.Sprintf("%s W:%t S:%t", account.Pubkey, account.IsWritable, account.IsSigner)
		}), "\n  ")
		tw.logger.Infof("OPERATIONS SOLANA - EXECUTE - ADDING TRANSACTION %d - ACCOUNTS:\n  %s\n", i, accountsStr)

		marshaledAdditionalFields, err := json.Marshal(additionalFields)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal additional fields: %w", err)
		}

		mcmsTxs[i] = mcmstypes.Transaction{
			To:               instruction.ProgramId.String(),
			Data:             instruction.Data,
			AdditionalFields: marshaledAdditionalFields,
		}
	}

	return mcmsTxs, nil
}

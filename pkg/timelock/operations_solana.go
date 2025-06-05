package timelock

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	mcmssolanasdk "github.com/smartcontractkit/mcms/sdk/solana"
	mcmstypes "github.com/smartcontractkit/mcms/types"

	timelockbindings "github.com/smartcontractkit/chainlink-ccip/chains/solana/gobindings/timelock"
	solanautils "github.com/smartcontractkit/chainlink-ccip/chains/solana/utils/common"
)

func (tw *WorkerSolana) execute(ctx context.Context, ops []TimelockCallScheduled) {
	if len(ops) == 0 {
		tw.logger.Warn("no calls given")
		return
	}
	if len(ops) > 1 {
		tw.logger.Warn("too many calls given")
		return
	}
	op := ops[0]

	isReady, err := tw.inspector.IsOperationReady(ctx, tw.timelockFullAddress, op.Id())
	if err != nil {
		tw.logger.Errorf("unable to read operation %x \"ready\" status: %s", op.Id(), err.Error())
		return
	}
	if !isReady {
		tw.logger.Infof("skipping operation %x: not ready", op.Id())
		return
	}

	batchOp, err := tw.mapCallScheduledEventsToMcmsBatchOperation(ctx, op)
	if err != nil {
		tw.logger.Errorf("unable to convert call scheduled events to mcms transactions: %v", err)
		return
	}

	tw.logger.Debugf("executing operation %x", op.Id())
	timelockExecutor := mcmssolanasdk.NewTimelockExecutor(tw.solanaClient, tw.privateKey)
	result, err := timelockExecutor.Execute(ctx, batchOp, tw.timelockFullAddress, op.Predecessor(), op.Salt())
	if err != nil {
		tw.logger.Errorf("execute operation %x error: %s", op.Id(), err.Error())
	} else {
		tw.logger.Infof("execute operation %x success: %s", op.Id(), result.Hash)
	}
}

func (tw *WorkerSolana) mapCallScheduledEventsToMcmsBatchOperation(
	ctx context.Context, event TimelockCallScheduled,
) (mcmstypes.BatchOperation, error) {
	tw.logger.Debugf("mapping event %x to mcms batch operation", event.Id())
	solanaEvent := event.(*solanaTimelockCallScheduled).callScheduledEvent

	transactions, err := tw.getTransactionsInBatchOperation(ctx, solanaEvent.ID)
	if err != nil {
		return mcmstypes.BatchOperation{}, fmt.Errorf("failed to get transactions in solana event 0x%x: %w", solanaEvent.ID, err)
	}

	return mcmstypes.BatchOperation{Transactions: transactions}, nil
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
	tw.logger.Debugf("retrieved operation %x from pda, containing %d instructions", operationID,
		len(scheduledOperation.Instructions))

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

		marshaledAdditionalFields, err := json.Marshal(additionalFields)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal additional fields: %w", err)
		}

		mcmsTxs[i] = mcmstypes.Transaction{
			To:               instruction.ProgramId.String(),
			Data:             instruction.Data,
			AdditionalFields: marshaledAdditionalFields,
		}
		tw.logger.Debugf("added transaction %d to mcms batch operation %x", i, operationID)
	}

	return mcmsTxs, nil
}

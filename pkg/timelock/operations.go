package timelock

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	chainselectors "github.com/smartcontractkit/chain-selectors"
	"github.com/smartcontractkit/timelock-worker/pkg/timelock/contract"
)

// execute runs the CallScheduled operation if:
// - The predecessor operation is finished
// - The operation is ready to be executed
// Otherwise the operation will throw an info log and wait for a future tick.
func (tw *Worker) execute(ctx context.Context, op []*contract.TimelockCallScheduled) {
	if !isReady(ctx, tw.contract, op[0].Id) {
		tw.logger.Info().Msgf("skipping operation %x: not ready", op[0].Id)
	}

	tw.logger.Debug().Msgf("execute operation %x", op[0].Id)
	tx, err := tw.executeCallSchedule(ctx, &tw.executeContract.TimelockTransactor, op, tw.privateKey)
	if err != nil || tx == nil {
		tw.logger.Error().Msgf("execute operation %x error: %s", op[0].Id, err.Error())
	} else {
		tw.logger.Info().Msgf("execute operation %x success: %s", op[0].Id, tx.Hash())

		if _, err = bind.WaitMined(ctx, tw.ethClient, tx); err != nil {
			tw.logger.Error().Msgf("execute operation %x error: %s", op[0].Id, err.Error())
		}
	}
}

// executeCallScheduleOperation is the handler to execute a CallScheduled operation.
func (tw *Worker) executeCallSchedule(ctx context.Context, c *contract.TimelockTransactor, cs []*contract.TimelockCallScheduled, privateKey *ecdsa.PrivateKey) (*types.Transaction, error) {
	fromAddress, err := privateKeyToAddress(privateKey)
	if err != nil {
		return nil, err
	}

	// Compute all the different calls from each specific CallSchedule.
	var calls []contract.RBACTimelockCall
	for _, op := range cs {
		calls = append(calls, contract.RBACTimelockCall{
			Target: op.Target,
			Value:  op.Value,
			Data:   op.Data,
		})
	}

	chainID, err := tw.ethClient.NetworkID(context.Background())
	if err != nil {
		return nil, err
	}

	txOpts := &bind.TransactOpts{
		From:    fromAddress,
		Signer:  tw.signTx,
		Context: ctx,
	}

	// if chainId is zksync-testnet or mainnet use custom gasPrice to enforce legacy tx
	if chainID.Cmp(new(big.Int).SetUint64(chainselectors.ETHEREUM_MAINNET_ZKSYNC_1.EvmChainID)) == 0 || chainID.Cmp(new(big.Int).SetUint64(chainselectors.ETHEREUM_TESTNET_SEPOLIA_ZKSYNC_1.EvmChainID)) == 0 {
		txOpts.GasPrice = big.NewInt(1000000000) // gasPrice set to 1 gwei
	}

	// Execute the tx's with all the computed calls.
	// Predecessor and salt are the same for all the tx's.
	tx, err := c.ExecuteBatch(
		txOpts,
		calls,
		cs[0].Predecessor,
		cs[0].Salt)
	if err != nil {
		return nil, err
	}

	return tx, nil
}

// isOperation returns a boolean determining if this is a valid operation.
// It's mostly to be used for sanity checks.
func isOperation(ctx context.Context, c *contract.Timelock, id [32]byte) bool {
	isOp, err := c.IsOperation(&bind.CallOpts{Context: ctx}, id)
	if err != nil {
		return false
	}

	return isOp
}

// isReady returns if the schedule operation is ready.
// Not applicable to other operation types.
func isReady(ctx context.Context, c *contract.Timelock, id [32]byte) bool {
	isReady, err := c.IsOperationReady(&bind.CallOpts{Context: ctx}, id)
	if err != nil {
		return false
	}

	return isReady
}

// isDone returns true when the operation has been completed.
func isDone(ctx context.Context, c *contract.Timelock, id [32]byte) bool {
	isDone, err := c.IsOperationDone(&bind.CallOpts{Context: ctx}, id)
	if err != nil {
		return false
	}

	return isDone
}

// isReady returns if the schedule operation is pending.
// Not applicable to other operation types.
func isPending(ctx context.Context, c *contract.Timelock, id [32]byte) bool {
	isPending, err := c.IsOperationPending(&bind.CallOpts{Context: ctx}, id)
	if err != nil {
		return false
	}

	return isPending
}

// signTx is a function that implements the type SignerFn, so can be passed as a Signer method.
func (tw *Worker) signTx(address common.Address, tx *types.Transaction) (*types.Transaction, error) {
	chainID, err := tw.ethClient.NetworkID(context.Background())
	if err != nil {
		return nil, err
	}

	signedTx, err := types.SignTx(tx, types.LatestSignerForChainID(chainID), tw.privateKey)
	if err != nil {
		return nil, err
	}

	return signedTx, nil
}

// privateKeyToAddress is an util function to calculate the address of a given private key.
// From a private key the public key can be deducted, and with the pubkey is
// trivial to calculate the address.
func privateKeyToAddress(privateKey *ecdsa.PrivateKey) (common.Address, error) {
	publicKeyECDSA, ok := privateKey.Public().(*ecdsa.PublicKey)
	if !ok {
		return common.Address{}, fmt.Errorf("publicKey is not of type *ecdsa.PublicKey")
	}

	return crypto.PubkeyToAddress(*publicKeyECDSA), nil
}

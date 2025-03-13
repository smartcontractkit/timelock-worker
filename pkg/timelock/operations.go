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

	contracts "github.com/smartcontractkit/ccip-owner-contracts/gethwrappers"
)

// execute runs the CallScheduled operation if:
// - The predecessor operation is finished
// - The operation is ready to be executed
// Otherwise the operation will throw an info log and wait for a future tick.
func (tw *Worker) execute(ctx context.Context, op []*contracts.RBACTimelockCallScheduled) {
	isReady, err := isReady(ctx, tw.contract, op[0].Id)
	if err != nil {
		tw.logger.Errorw("unable to read operation %x \"ready\" status: %s", op[0].Id, err.Error())
		return
	}
	if !isReady {
		tw.logger.Infof("skipping operation %x: not ready", op[0].Id)
		return
	}

	tw.logger.Debugf("execute operation %x", op[0].Id)

	tx, err := tw.executeCallSchedule(ctx, &tw.executeContract.RBACTimelockTransactor, op, tw.privateKey)
	if err != nil || tx == nil {
		tw.logger.Errorf("execute operation %x error: %s", op[0].Id, err.Error())
	} else {
		tw.logger.Infof("execute operation %x success: %s", op[0].Id, tx.Hash())

		_, err := Retry(ctx, func(rctx context.Context) (*types.Receipt, error) {
			return bind.WaitMined(rctx, tw.ethClient, tx)
		})
		if err != nil {
			tw.logger.Errorf("execute operation %x error: %s", op[0].Id, err.Error())
		}
	}
}

// executeCallScheduleOperation is the handler to execute a CallScheduled operation.
func (tw *Worker) executeCallSchedule(ctx context.Context, c *contracts.RBACTimelockTransactor, cs []*contracts.RBACTimelockCallScheduled, privateKey *ecdsa.PrivateKey) (*types.Transaction, error) {
	fromAddress, err := privateKeyToAddress(privateKey)
	if err != nil {
		return nil, err
	}

	// Compute all the different calls from each specific CallSchedule.
	calls := make([]contracts.RBACTimelockCall, 0, len(cs))
	for _, op := range cs {
		calls = append(calls, contracts.RBACTimelockCall{
			Target: op.Target,
			Value:  op.Value,
			Data:   op.Data,
		})
	}

	chainID, err := Retry(ctx, func(rctx context.Context) (*big.Int, error) {
		return tw.ethClient.NetworkID(rctx)
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get network id: %w", err)
	}

	txOpts := &bind.TransactOpts{
		From:   fromAddress,
		Signer: tw.signTx(chainID),
	}

	// if chainId is zksync-testnet or mainnet use custom gasPrice to enforce legacy tx
	if chainID.Cmp(new(big.Int).SetUint64(chainselectors.ETHEREUM_MAINNET_ZKSYNC_1.EvmChainID)) == 0 || chainID.Cmp(new(big.Int).SetUint64(chainselectors.ETHEREUM_TESTNET_SEPOLIA_ZKSYNC_1.EvmChainID)) == 0 {
		tw.logger.Infof("zkSync chain detected, using legacy tx")
		txOpts.GasPrice = big.NewInt(1000000000) // gasPrice set to 1 gwei
	}

	tw.logger.Infof("Calling execute Batch...")
	// Execute the tx's with all the computed calls.
	// Predecessor and salt are the same for all the tx's.
	return Retry(ctx, func(rctx context.Context) (*types.Transaction, error) {
		txOpts.Context = rctx //nolint:fatcontext
		return c.ExecuteBatch(txOpts, calls, cs[0].Predecessor, cs[0].Salt)
	})
}

// isOperation returns a boolean determining if this is a valid operation.
// It's mostly to be used for sanity checks.
func isOperation(ctx context.Context, c *contracts.RBACTimelock, id [32]byte) (bool, error) {
	return Retry(ctx, func(rctx context.Context) (bool, error) {
		return c.IsOperation(&bind.CallOpts{Context: rctx}, id)
	})
}

// isReady returns if the schedule operation is ready.
// Not applicable to other operation types.
func isReady(ctx context.Context, c *contracts.RBACTimelock, id [32]byte) (bool, error) {
	return Retry(ctx, func(rctx context.Context) (bool, error) {
		return c.IsOperationReady(&bind.CallOpts{Context: rctx}, id)
	})
}

// isDone returns true when the operation has been completed.
func isDone(ctx context.Context, c *contracts.RBACTimelock, id [32]byte) (bool, error) {
	return Retry(ctx, func(rctx context.Context) (bool, error) {
		return c.IsOperationDone(&bind.CallOpts{Context: rctx}, id)
	})
}

// isReady returns if the schedule operation is pending.
// Not applicable to other operation types.
func isPending(ctx context.Context, c *contracts.RBACTimelock, id [32]byte) (bool, error) {
	return Retry(ctx, func(rctx context.Context) (bool, error) {
		return c.IsOperationPending(&bind.CallOpts{Context: rctx}, id)
	})
}

// signTx is a function that implements the type SignerFn, so can be passed as a Signer method.
func (tw *Worker) signTx(chainID *big.Int) bind.SignerFn {
	return func(address common.Address, tx *types.Transaction) (*types.Transaction, error) {
		signedTx, err := types.SignTx(tx, types.LatestSignerForChainID(chainID), tw.privateKey)
		if err != nil {
			return nil, err
		}

		return signedTx, nil
	}
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

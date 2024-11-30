package integration

import (
	"context"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	contracts "github.com/smartcontractkit/ccip-owner-contracts/gethwrappers"
	"github.com/stretchr/testify/require"

	test_contracts "github.com/smartcontractkit/timelock-worker/tests/contracts"
)

func DeployTimelock(
	t *testing.T, ctx context.Context, transactor *bind.TransactOpts, backend Backend,
	adminAccount common.Address, minDelay *big.Int,
) (
	common.Address, *types.Transaction, *types.Receipt, *contracts.RBACTimelock,
) {
	t.Helper()

	proposers := []common.Address{adminAccount}
	executors := []common.Address{adminAccount}
	cancellers := []common.Address{adminAccount}
	bypassers := []common.Address{}

	address, transaction, contract, err := contracts.DeployRBACTimelock(
		transactor, backend, minDelay, adminAccount, proposers, executors, cancellers, bypassers)
	require.NoError(t, err)

	backend.Commit()
	receipt, err := bind.WaitMined(ctx, backend, transaction)
	require.NoError(t, err)
	require.Equal(t, types.ReceiptStatusSuccessful, receipt.Status)
	t.Logf("timelock address: %v; deploy transaction: %v", address, transaction.Hash())

	return address, transaction, receipt, contract
}

func DeployCallProxy(
	t *testing.T, ctx context.Context, transactor *bind.TransactOpts, backend Backend, timelockAddress common.Address,
) (
	common.Address, *types.Transaction, *types.Receipt, *contracts.CallProxy,
) {
	t.Helper()

	address, transaction, contract, err := contracts.DeployCallProxy(transactor, backend, timelockAddress)
	require.NoError(t, err)

	backend.Commit()
	receipt, err := bind.WaitMined(ctx, backend, transaction)
	require.NoError(t, err)
	require.Equal(t, types.ReceiptStatusSuccessful, receipt.Status)
	t.Logf("call proxy address: %v; deploy transaction: %v", address, transaction.Hash())

	return address, transaction, receipt, contract
}

func DeployStorage(
	t *testing.T, ctx context.Context, transactor *bind.TransactOpts, backend Backend,
) (
	common.Address, *types.Transaction, *types.Receipt, *test_contracts.StorageContract,
) {
	t.Helper()

	address, transaction, contract, err := test_contracts.DeployStorageContract(transactor, backend)
	require.NoError(t, err)

	backend.Commit()
	receipt, err := bind.WaitMined(ctx, backend, transaction)
	require.NoError(t, err)
	require.Equal(t, types.ReceiptStatusSuccessful, receipt.Status)
	t.Logf("timelock address: %v; deploy transaction: %v", address, transaction.Hash())

	return address, transaction, receipt, contract
}

func UpdateDelay(
	t *testing.T, ctx context.Context, transactor *bind.TransactOpts, backend Backend,
	timelockContract *contracts.RBACTimelock, delay *big.Int,
) (
	*types.Transaction, *types.Receipt,
) {
	t.Helper()

	transaction, err := timelockContract.UpdateDelay(transactor, delay)
	require.NoError(t, err)

	backend.Commit()
	receipt, err := bind.WaitMined(ctx, backend, transaction)
	require.NoError(t, err)
	require.Equal(t, types.ReceiptStatusSuccessful, receipt.Status)
	t.Logf("update delay transaction: %v", transaction.Hash())

	return transaction, receipt
}

func ScheduleBatch(
	t *testing.T,
	ctx context.Context, transactor *bind.TransactOpts, backend Backend,
	timelockContract *contracts.RBACTimelock, calls []contracts.RBACTimelockCall,
	predecessor [32]byte, salt [32]byte, delay *big.Int,
) (
	*types.Transaction, *types.Receipt,
) {
	t.Helper()

	transaction, err := timelockContract.ScheduleBatch(transactor, calls, predecessor, salt, delay)
	require.NoError(t, err)

	backend.Commit()
	receipt, err := bind.WaitMined(ctx, backend, transaction)
	require.NoError(t, err)
	require.Equal(t, types.ReceiptStatusSuccessful, receipt.Status)
	t.Logf("schedule batch transaction: %v", transaction.Hash())

	return transaction, receipt
}

func ExecuteBatch(
	t *testing.T, ctx context.Context, transactor *bind.TransactOpts, backend Backend,
	timelockContract *contracts.RBACTimelock, calls []contracts.RBACTimelockCall,
	predecessor [32]byte, salt [32]byte,
) (
	*types.Transaction, *types.Receipt,
) {
	t.Helper()

	transaction, err := timelockContract.ExecuteBatch(transactor, calls, predecessor, salt)
	require.NoError(t, err)

	backend.Commit()
	receipt, err := bind.WaitMined(ctx, backend, transaction)
	require.NoError(t, err)
	require.Equal(t, types.ReceiptStatusSuccessful, receipt.Status)
	t.Logf("execute batch transaction: %v", transaction.Hash())

	return transaction, receipt
}

func CancelBatch(
	t *testing.T,
	ctx context.Context, transactor *bind.TransactOpts, backend Backend,
	timelockContract *contracts.RBACTimelock, operationID [32]byte,
) (
	*types.Transaction, *types.Receipt,
) {
	t.Helper()

	transaction, err := timelockContract.Cancel(transactor, operationID)
	require.NoError(t, err)

	backend.Commit()
	receipt, err := bind.WaitMined(ctx, backend, transaction)
	require.NoError(t, err)
	require.Equal(t, types.ReceiptStatusSuccessful, receipt.Status)
	t.Logf("cancel batch transaction: %v", transaction.Hash())

	return transaction, receipt
}

func SendTransaction(
	t *testing.T, ctx context.Context, transactor *bind.TransactOpts, backend Backend,
	account TestAccount, chainID *big.Int, value *big.Int, to common.Address, data []byte,
) {
	t.Helper()

	nonce, err := backend.PendingNonceAt(ctx, account.Address)
	require.NoError(t, err)

	gasLimit := uint64(21000)
	gasPrice, err := backend.SuggestGasPrice(ctx)
	require.NoError(t, err)

	transaction := types.NewTransaction(nonce, to, value, gasLimit, gasPrice, data)

	signedTx, err := types.SignTx(transaction, types.NewEIP155Signer(chainID), account.PrivateKey)
	require.NoError(t, err)

	err = backend.SendTransaction(ctx, signedTx)
	require.NoError(t, err)

	backend.Commit()
}

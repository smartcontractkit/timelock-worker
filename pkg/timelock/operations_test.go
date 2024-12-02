package timelock

import (
	"context"
	"crypto/ecdsa"
	"math/big"
	"reflect"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	contracts "github.com/smartcontractkit/ccip-owner-contracts/gethwrappers"
	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/timelock-worker/tests/integration"
	test_contracts "github.com/smartcontractkit/timelock-worker/tests/contracts"
)

func Test_is_methods(t *testing.T) {
	// --- arrange ---
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	account := integration.NewTestAccount(t)
	backend := integration.NewSimulatedBackend(t, types.GenesisAlloc{
		account.Address: types.Account{Balance: big.NewInt(1e18)},
	})

	chainID, err := backend.ChainID(ctx)
	require.NoError(t, err)

	transactor, err := bind.NewKeyedTransactorWithChainID(account.PrivateKey, chainID)
	require.NoError(t, err)

	_, _, _, timelockContract := integration.DeployTimelock(t, ctx, transactor,
		backend, account.Address, big.NewInt(1))

	storageAddress, _, _, _ := integration.DeployStorage(t, ctx, transactor, backend)

	predecessor := [32]byte{}
	salt := [32]byte{}
	calls := []contracts.RBACTimelockCall{
		{Target: storageAddress, Value: big.NewInt(0), Data: abiEncodedStoreCall(t, 123)},
	}

	operationId, err := timelockContract.HashOperationBatch(&bind.CallOpts{}, calls, predecessor, salt)
	require.NoError(t, err)
	t.Logf("operation id: %v", hexutil.Encode(operationId[:]))

	// --- act ---
	integration.ScheduleBatch(t, ctx, transactor, backend, timelockContract, calls,
		predecessor, salt, big.NewInt(1))

	// --- assert ---
	require.True(t, isOperation(ctx, timelockContract, operationId))
	require.True(t, isPending(ctx, timelockContract, operationId))
	require.False(t, isReady(ctx, timelockContract, operationId))
	require.False(t, isDone(ctx, timelockContract, operationId))

	// generate a new block then check isReady again
	backend.Commit()
	require.True(t, isReady(ctx, timelockContract, operationId))
	require.True(t, isPending(ctx, timelockContract, operationId))
	require.False(t, isDone(ctx, timelockContract, operationId))

	// execute then check isDone again
	integration.ExecuteBatch(t, ctx, transactor, backend, timelockContract, calls,
		predecessor, salt)
	require.True(t, isDone(ctx, timelockContract, operationId))
	require.False(t, isPending(ctx, timelockContract, operationId))
	require.False(t, isReady(ctx, timelockContract, operationId))
}

func Test_privateKeyToAddress(t *testing.T) {
	type args struct {
		privateKey *ecdsa.PrivateKey
	}
	tests := []struct {
		name    string
		args    args
		want    common.Address
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := privateKeyToAddress(tt.args.privateKey)
			if (err != nil) != tt.wantErr {
				t.Errorf("privateKeyToAddress() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("privateKeyToAddress() = %v, want %v", got, tt.want)
			}
		})
	}
}

// ----- helpers -----

func abiEncodedStoreCall(t *testing.T, value int64) []byte {
	t.Helper()

	abi, err := test_contracts.StorageContractMetaData.GetAbi()
	require.NoError(t, err)
	encoded, err := abi.Pack("store", big.NewInt(value))
	require.NoError(t, err)

	return encoded
}

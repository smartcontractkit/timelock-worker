package timelock

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"math/big"
	"reflect"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	contracts "github.com/smartcontractkit/ccip-owner-contracts/gethwrappers"
	"github.com/smartcontractkit/timelock-worker/pkg/timelock/mocks"
	test_contracts "github.com/smartcontractkit/timelock-worker/tests/contracts"
	"github.com/smartcontractkit/timelock-worker/tests/integration"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func Test_is_methods(t *testing.T) {
	// --- arrange ---
	ctx, cancel := context.WithTimeout(context.Background(), 1000*time.Second)
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
	requireEqual(t, true)(isOperation(ctx, timelockContract, operationId))
	requireEqual(t, true)(isPending(ctx, timelockContract, operationId))
	requireEqual(t, false)(isReady(ctx, timelockContract, operationId))
	requireEqual(t, false)(isDone(ctx, timelockContract, operationId))

	// generate a new block then check isReady again
	backend.Commit()
	requireEqual(t, true)(isReady(ctx, timelockContract, operationId))
	requireEqual(t, true)(isPending(ctx, timelockContract, operationId))
	requireEqual(t, false)(isDone(ctx, timelockContract, operationId))

	// execute then check isDone again
	integration.ExecuteBatch(t, ctx, transactor, backend, timelockContract, calls,
		predecessor, salt)
	requireEqual(t, true)(isDone(ctx, timelockContract, operationId))
	requireEqual(t, false)(isPending(ctx, timelockContract, operationId))
	requireEqual(t, false)(isReady(ctx, timelockContract, operationId))
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

type TestRBACTimelock struct {
	*contracts.RBACTimelock
	*mocks.ContractBackend
}

var (
	encodedTrue  = common.LeftPadBytes([]byte{1}, 32)
	encodedFalse = common.LeftPadBytes([]byte{0}, 32)
)

func Test_retries(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	account := integration.NewTestAccount(t)
	auth, err := bind.NewKeyedTransactorWithChainID(account.PrivateKey, big.NewInt(1337))
	require.NoError(t, err)

	type IsFn func(context.Context, *contracts.RBACTimelock, [32]byte) (bool, error)

	// mock setup functions
	mockReturnTrue := func(backend *mocks.ContractBackend) {
		backend.EXPECT().CallContract(mock.Anything, mock.Anything, mock.Anything).Return(encodedTrue, nil).Once()
	}
	mockReturnFalse := func(backend *mocks.ContractBackend) {
		backend.EXPECT().CallContract(mock.Anything, mock.Anything, mock.Anything).Return(encodedFalse, nil).Once()
	}
	mockReturnTrueAfterRetries := func(backend *mocks.ContractBackend) {
		setRetryParamsForTests(t)
		err := fmt.Errorf("is-operation-transient-error")
		backend.EXPECT().CallContract(mock.Anything, mock.Anything, mock.Anything).Return(encodedFalse, err).Times(4)
		backend.EXPECT().CallContract(mock.Anything, mock.Anything, mock.Anything).Return(encodedTrue, nil).Once()
	}
	mockFailureEvenAfterRetries := func(backend *mocks.ContractBackend) {
		setRetryParamsForTests(t)
		err := fmt.Errorf("is-operation-error")
		backend.EXPECT().CallContract(mock.Anything, mock.Anything, mock.Anything).Return(encodedFalse, err).Times(5)
	}

	tests := []struct {
		name        string
		isFunction  IsFn
		setup       func(*mocks.ContractBackend)
		operationID [32]byte
		want        bool
		wantErr     string
	}{
		// IsOperation
		{
			name:       "success - isOperation - true",
			isFunction: isOperation,
			setup:      mockReturnTrue,
			want:       true,
		},
		{
			name:       "success - isOperation - false",
			isFunction: isOperation,
			setup:      mockReturnFalse,
			want:       false,
		},
		{
			name:       "success - isOperation - true after retries",
			isFunction: isOperation,
			setup:      mockReturnTrueAfterRetries,
			want:       true,
		},
		{
			name:       "failure - isOperation - error after retries",
			isFunction: isOperation,
			setup:      mockFailureEvenAfterRetries,
			wantErr:    "is-operation-error",
		},

		// IsPending
		{
			name:       "success - isPending - true",
			isFunction: isPending,
			setup:      mockReturnTrue,
			want:       true,
		},
		{
			name:       "success - isPending - false",
			isFunction: isPending,
			setup:      mockReturnFalse,
			want:       false,
		},
		{
			name:       "success - isPending - true after retries",
			isFunction: isPending,
			setup:      mockReturnTrueAfterRetries,
			want:       true,
		},
		{
			name:       "failure - isPending - error after retries",
			isFunction: isPending,
			setup:      mockFailureEvenAfterRetries,
			wantErr:    "is-operation-error",
		},

		// IsReady
		{
			name:       "success - isReady - true",
			isFunction: isReady,
			setup:      mockReturnTrue,
			want:       true,
		},
		{
			name:       "success - isReady - false",
			isFunction: isReady,
			setup:      mockReturnFalse,
			want:       false,
		},
		{
			name:       "success - isReady - true after retries",
			isFunction: isReady,
			setup:      mockReturnTrueAfterRetries,
			want:       true,
		},
		{
			name:       "failure - isReady - error after retries",
			isFunction: isReady,
			setup:      mockFailureEvenAfterRetries,
			wantErr:    "is-operation-error",
		},

		// IsDone
		{
			name:       "success - isDone - true",
			isFunction: isDone,
			setup:      mockReturnTrue,
			want:       true,
		},
		{
			name:       "success - isDone - false",
			isFunction: isDone,
			setup:      mockReturnFalse,
			want:       false,
		},
		{
			name:       "success - isDone - true after retries",
			isFunction: isDone,
			setup:      mockReturnTrueAfterRetries,
			want:       true,
		},
		{
			name:       "failure - isDone - error after retries",
			isFunction: isDone,
			setup:      mockFailureEvenAfterRetries,
			wantErr:    "is-operation-error",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			contract, backend := newTestContract(t, auth)
			tt.setup(backend)

			got, err := tt.isFunction(ctx, contract, tt.operationID)

			if tt.wantErr == "" {
				require.NoError(t, err)
				require.Equal(t, tt.want, got)
			} else {
				require.ErrorContains(t, err, tt.wantErr)
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

func requireEqual[T any](t *testing.T, want T) func(T, error) {
	t.Helper()
	return func(value T, err error) {
		require.NoError(t, err)
		require.Equal(t, want, value)
	}
}

func newTestContract(t *testing.T, auth *bind.TransactOpts) (*contracts.RBACTimelock, *mocks.ContractBackend) {
	backend := mocks.NewContractBackend(t)
	mockDeployContract(backend)
	_, _, contract, err := contracts.DeployRBACTimelock(auth, backend, big.NewInt(1),
		common.Address{}, nil, nil, nil, nil)
	require.NoError(t, err)

	return contract, backend
}

func mockDeployContract(backend *mocks.ContractBackend) {
	backend.EXPECT().HeaderByNumber(mock.Anything, mock.Anything).Return(&types.Header{}, nil)
	backend.EXPECT().SuggestGasPrice(mock.Anything).Return(big.NewInt(50000), nil)
	backend.EXPECT().EstimateGas(mock.Anything, mock.Anything).Return(uint64(21000), nil)
	backend.EXPECT().PendingNonceAt(mock.Anything, mock.Anything).Return(uint64(10), nil)
	backend.EXPECT().SendTransaction(mock.Anything, mock.Anything).Return(nil)
}

func setRetryParamsForTests(t *testing.T) {
	savedDelays := retryIncrementalDelays
	retryIncrementalDelays = [4]int{5, 10, 15, 20}
	t.Cleanup(func() {
		retryIncrementalDelays = savedDelays
	})
}

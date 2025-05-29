package timelock

import (
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/core/types"
	bindings "github.com/smartcontractkit/ccip-owner-contracts/gethwrappers"
	"github.com/stretchr/testify/assert"
)

func TestNewEVMTimelockCallScheduled(t *testing.T) {
	callScheduled := makeTestCallScheduled()
	evmCallScheduled := NewEVMTimelockCallScheduled(callScheduled)
	assert.NotNil(t, evmCallScheduled)
}

func TestEVMTimelockCallScheduled_Id(t *testing.T) {
	callScheduled := makeTestCallScheduled()
	evmCallScheduled := NewEVMTimelockCallScheduled(callScheduled)
	assert.Equal(t, operationKey(callScheduled.Id), evmCallScheduled.Id())
}

func TestEVMTimelockCallScheduled_Index(t *testing.T) {
	callScheduled := makeTestCallScheduled()
	evmCallScheduled := NewEVMTimelockCallScheduled(callScheduled)
	assert.Equal(t, 42, evmCallScheduled.Index())
}

func TestEVMTimelockCallScheduled_BlockNumber(t *testing.T) {
	callScheduled := makeTestCallScheduled()
	evmCallScheduled := NewEVMTimelockCallScheduled(callScheduled)
	assert.Equal(t, big.NewInt(123456), evmCallScheduled.BlockNumber())
}

func TestEVMTimelockCallScheduled_TxHash(t *testing.T) {
	callScheduled := makeTestCallScheduled()
	evmCallScheduled := NewEVMTimelockCallScheduled(callScheduled)
	expected := "aabbcc0000000000000000000000000000000000000000000000000000000000"
	assert.Equal(t, expected, evmCallScheduled.TxHash())
}

// ----- helpers ------

func makeTestCallScheduled() *bindings.RBACTimelockCallScheduled {
	return &bindings.RBACTimelockCallScheduled{
		Id:    [32]byte{1, 2, 3},
		Index: big.NewInt(42),
		Raw: types.Log{
			BlockNumber: 123456,
			TxHash:      [32]byte{0xaa, 0xbb, 0xcc},
		},
	}
}

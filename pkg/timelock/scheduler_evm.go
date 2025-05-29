package timelock

import (
	"fmt"
	"math/big"

	bindings "github.com/smartcontractkit/ccip-owner-contracts/gethwrappers"
)

var _ TimelockCallScheduled = NewEVMTimelockCallScheduled(nil)

type evmTimelockCallScheduled struct {
	callScheduled *bindings.RBACTimelockCallScheduled
}

func NewEVMTimelockCallScheduled(callScheduled *bindings.RBACTimelockCallScheduled) TimelockCallScheduled {
	return &evmTimelockCallScheduled{
		callScheduled: callScheduled,
	}
}

func (cs *evmTimelockCallScheduled) Id() operationKey {
	return cs.callScheduled.Id
}

func (cs *evmTimelockCallScheduled) Index() int {
	return int(cs.callScheduled.Index.Int64())
}

func (cs *evmTimelockCallScheduled) BlockNumber() *big.Int {
	return new(big.Int).SetUint64(cs.callScheduled.Raw.BlockNumber)
}

func (cs *evmTimelockCallScheduled) TxHash() string {
	return fmt.Sprintf("%x", cs.callScheduled.Raw.TxHash[:])
}

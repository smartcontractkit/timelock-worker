package timelock

import (
	"fmt"
	"math/big"

	eth "github.com/ethereum/go-ethereum/common"
)

var _ TimelockCallScheduled = NewEVMTimelockCallScheduled(nil)

type solanaTimelockCallScheduled struct {
	callScheduledEvent SolanaTimelockCallScheduledEvent
}

func NewSolanaTimelockCallScheduled(callScheduledEvent SolanaTimelockCallScheduledEvent) TimelockCallScheduled {
	return &solanaTimelockCallScheduled{
		callScheduledEvent: callScheduledEvent,
	}
}

func (cs *solanaTimelockCallScheduled) Id() operationKey {
	return cs.callScheduledEvent.ID
}

func (cs *solanaTimelockCallScheduled) Index() int {
	return int(cs.callScheduledEvent.Index) //nolint:gosec
}

func (cs *solanaTimelockCallScheduled) BlockNumber() *big.Int {
	return cs.callScheduledEvent.BlockNumber
}

func (cs *solanaTimelockCallScheduled) TxHash() string {
	return cs.callScheduledEvent.TxHash
}

func (cs *solanaTimelockCallScheduled) Predecessor() eth.Hash {
	return cs.callScheduledEvent.Predecessor
}

func (cs *solanaTimelockCallScheduled) Salt() eth.Hash {
	return cs.callScheduledEvent.Salt
}

func (cs *solanaTimelockCallScheduled) String() string {
	return fmt.Sprintf("{solanaTimelockCallScheduled | id:0x%x, index:%v, blockNumber:%v, txHash:%v, predecessor:%v, salt:%v}",
		cs.Id(), cs.Index(), cs.BlockNumber(), cs.TxHash(), cs.Predecessor(), cs.Salt())
}

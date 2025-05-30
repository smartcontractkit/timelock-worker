package timelock

import (
	"math/big"
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

package timelock

import (
	"math/big"
	"testing"

	"github.com/gagliardetto/solana-go"
	"github.com/stretchr/testify/require"
)

func TestNewSolanaTimelockCallScheduled(t *testing.T) {
	solanaCallScheduled := NewSolanaTimelockCallScheduled(makeSolanaTimelockCallScheduledEvent())
	require.NotNil(t, solanaCallScheduled)
}

func TestSolanaTimelockCallScheduled_Id(t *testing.T) {
	event := makeSolanaTimelockCallScheduledEvent()
	solanaCallScheduled := NewSolanaTimelockCallScheduled(event)
	require.Equal(t, operationKey{1, 2, 3}, solanaCallScheduled.Id())
}

func TestSolanaTimelockCallScheduled_Index(t *testing.T) {
	event := makeSolanaTimelockCallScheduledEvent()
	solanaCallScheduled := NewSolanaTimelockCallScheduled(event)
	require.Equal(t, 42, solanaCallScheduled.Index())
}

func TestSolanaTimelockCallScheduled_BlockNumber(t *testing.T) {
	event := makeSolanaTimelockCallScheduledEvent()
	solanaCallScheduled := NewSolanaTimelockCallScheduled(event)
	require.Equal(t, big.NewInt(123456), solanaCallScheduled.BlockNumber())
}

func TestSolanaTimelockCallScheduled_TxHash(t *testing.T) {
	event := makeSolanaTimelockCallScheduledEvent()
	solanaCallScheduled := NewSolanaTimelockCallScheduled(event)
	require.Equal(t, "0xabc123", solanaCallScheduled.TxHash())
}

// ----- helpers ------

func makeSolanaTimelockCallScheduledEvent() SolanaTimelockCallScheduledEvent {
	return SolanaTimelockCallScheduledEvent{
		ID:          [32]byte{1, 2, 3},
		Index:       42,
		Target:      solana.PublicKey{4, 5, 6},
		Predecessor: [32]byte{7, 8, 9},
		Salt:        [32]byte{10, 11, 12},
		Delay:       100,
		Data:        []byte{13, 14, 15},
		BlockNumber: big.NewInt(123456),
		TxHash:      "0xabc123",
	}
}

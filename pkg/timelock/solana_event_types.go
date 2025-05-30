package timelock

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"math/big"
	"strings"

	bin "github.com/gagliardetto/binary"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
)

// TxTimelockEvents groups all decoded events from one transaction.
type SolanaTimelockEvents struct {
	Scheduled []SolanaTimelockCallScheduledEvent
	Executed  []SolanaTimelockCallExecutedEvent
	Cancelled []SolanaTimelockCallCancelledEvent
}

// Discriminators for each event type.
var (
	SolanaTimelockCallScheduledDiscriminator = solanaEventDiscriminator("CallScheduled")
	SolanaTimelockCallExecutedDiscriminator  = solanaEventDiscriminator("CallExecuted")
	SolanaTimelockCancelledDiscriminator     = solanaEventDiscriminator("Cancelled")
)

// --- Event structs ---

// SolanaTimelockCallScheduledEvent corresponds to #[event] CallScheduled { id, index, target, predecessor, salt, delay, data }.
type SolanaTimelockCallScheduledEvent struct {
	ID          operationKey
	Index       uint64
	Target      solana.PublicKey
	Predecessor operationKey
	Salt        [32]byte
	Delay       uint64
	Data        []byte
	BlockNumber *big.Int `borsh_skip:"true"`
	TxHash      string   `borsh_skip:"true"`
}

// SolanaTimelockCallExecutedEvent corresponds to #[event] CallExecuted { id, index, target, data }.
type SolanaTimelockCallExecutedEvent struct {
	ID     operationKey
	Index  uint64
	Target solana.PublicKey
	Data   []byte
}

// SolanaTimelockCallCancelledEvent corresponds to #[event] Cancelled { id }.
type SolanaTimelockCallCancelledEvent struct {
	ID operationKey
}

// ParseTimelockEvents extracts and decodes Anchor events from tx.Meta.LogMessages.
func ParseTimelockEvents(tx *rpc.TransactionWithMeta) (*SolanaTimelockEvents, error) {
	solanaTx, err := tx.GetParsedTransaction()
	if err != nil {
		return nil, fmt.Errorf("unable to get solana transaction: %w", err)
	}
	if len(solanaTx.Signatures) == 0 {
		return nil, fmt.Errorf("solana transaction does not have signatures")
	}

	out := &SolanaTimelockEvents{}

	for _, log := range tx.Meta.LogMessages {
		if !strings.HasPrefix(log, "Program data: ") {
			continue
		}
		b64 := strings.TrimPrefix(log, "Program data: ")
		blob, err := base64.StdEncoding.DecodeString(b64)
		if err != nil || len(blob) < 8 {
			continue
		}

		disc := blob[:8]
		payload := blob[8:]
		borshDecoder := bin.NewBorshDecoder(payload)

		switch {
		case bytes.Equal(disc, SolanaTimelockCallScheduledDiscriminator[:]):
			var event SolanaTimelockCallScheduledEvent
			if err := borshDecoder.Decode(&event); err == nil {
				event.BlockNumber = new(big.Int).SetUint64(tx.Slot)
				event.TxHash = solanaTx.Signatures[0].String()
				out.Scheduled = append(out.Scheduled, event)
			}

		case bytes.Equal(disc, SolanaTimelockCallExecutedDiscriminator[:]):
			var event SolanaTimelockCallExecutedEvent
			if err := borshDecoder.Decode(&event); err == nil {
				out.Executed = append(out.Executed, event)
			}

		case bytes.Equal(disc, SolanaTimelockCancelledDiscriminator[:]):
			var event SolanaTimelockCallCancelledEvent
			if err := borshDecoder.Decode(&event); err == nil {
				out.Cancelled = append(out.Cancelled, event)
			}
		}
	}

	return out, nil
}

// solanaEventDiscriminator computes the first 8 bytes of sha256("event:<EventName>").
func solanaEventDiscriminator(name string) [8]byte {
	h := sha256.Sum256([]byte("event:" + name))
	var disc [8]byte
	copy(disc[:], h[:8])

	return disc
}

package timelock

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"strings"

	bin "github.com/gagliardetto/binary"
	solana "github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	"go.uber.org/zap"
)

// eventDiscriminator computes the first 8 bytes of sha256("event:<EventName>").
func eventDiscriminator(name string) [8]byte {
	h := sha256.Sum256([]byte("event:" + name))
	var disc [8]byte
	copy(disc[:], h[:8])

	return disc
}

// --- Event structs ---

// CallScheduled corresponds to #[event] CallScheduled { id, index, target, predecessor, salt, delay, data }.
type CallScheduled struct {
	ID          [32]byte
	Index       uint64
	Target      solana.PublicKey
	Predecessor [32]byte
	Salt        [32]byte
	Delay       uint64
	Data        []byte
}

// CallExecuted corresponds to #[event] CallExecuted { id, index, target, data }.
type CallExecuted struct {
	ID     [32]byte
	Index  uint64
	Target solana.PublicKey
	Data   []byte
}

// Cancelled corresponds to #[event] Cancelled { id }.
type Cancelled struct {
	ID [32]byte
}

// Discriminators for each event type.
var (
	CallScheduledDiscriminator = eventDiscriminator("CallScheduled")
	CallExecutedDiscriminator  = eventDiscriminator("CallExecuted")
	CancelledDiscriminator     = eventDiscriminator("Cancelled")
)

// TimelockEvents groups all decoded events from one Solana transaction.
type TimelockEvents struct {
	Scheduled []CallScheduled
	Executed  []CallExecuted
	Cancelled []Cancelled
}

// ParseTimelockEvents extracts and decodes Anchor events from tx.Meta.LogMessages.
func ParseTimelockEvents(lggr *zap.SugaredLogger, tx *rpc.TransactionWithMeta) (*TimelockEvents, error) {
	out := &TimelockEvents{}

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

		switch {
		case bytes.Equal(disc, CallScheduledDiscriminator[:]):
			var e CallScheduled
			dec := bin.NewBorshDecoder(payload)
			if err := dec.Decode(&e); err == nil {
				out.Scheduled = append(out.Scheduled, e)
			} else {
				lggr.Warnf("Failed to decode CallScheduled event: %v", err)
			}

		case bytes.Equal(disc, CallExecutedDiscriminator[:]):
			var e CallExecuted
			dec := bin.NewBorshDecoder(payload)
			if err := dec.Decode(&e); err == nil {
				out.Executed = append(out.Executed, e)
			} else {
				lggr.Warnf("Failed to decode CallExecuted event: %v", err)
			}

		case bytes.Equal(disc, CancelledDiscriminator[:]):
			var e Cancelled
			dec := bin.NewBorshDecoder(payload)
			if err := dec.Decode(&e); err == nil {
				out.Cancelled = append(out.Cancelled, e)
			} else {
				lggr.Warnf("Failed to decode Cancelled event: %v", err)
			}

		default:
			continue
		}
	}

	return out, nil
}

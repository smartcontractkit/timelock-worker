package timelock

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"strings"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
)

// TxTimelockEvents groups all decoded events from one transaction.
type TxTimelockEvents struct {
	Scheduled []CallScheduled
	Executed  []CallExecuted
	Bypasser  []BypasserCallExecuted
	Cancelled []Cancelled
}

// eventDiscriminator computes the first 8 bytes of sha256("event:<EventName>")
func eventDiscriminator(name string) [8]byte {
	h := sha256.Sum256([]byte("event:" + name))
	var disc [8]byte
	copy(disc[:], h[:8])
	return disc
}

// CallScheduled matches the CallScheduled Anchor event:
// https://github.com/smartcontractkit/chainlink-ccip/blob/bca6b9d1cd7ebaf0487bf14eba2675d6ef1d0673/chains/solana/contracts/programs/timelock/src/event.rs#L5-L13
type CallScheduled struct {
	ID          [32]byte
	Index       uint64
	Target      solana.PublicKey
	Predecessor [32]byte
	Salt        [32]byte
	Delay       uint64
	Data        []byte
}

// CallScheduledDiscriminator 8 bytes discriminator for CallScheduled
var CallScheduledDiscriminator = eventDiscriminator("CallScheduled")

func (e *CallScheduled) UnmarshalWithDecoder(d *borsh.Decoder) error {
	return d.Decode(e)
}

// CallExecuted matches the Anchor event:
// #[event] pub struct CallExecuted { id: [u8;32], index: u64, target: Pubkey, data: Vec<u8> }
type CallExecuted struct {
	ID     [32]byte
	Index  uint64
	Target solana.PublicKey
	Data   []byte
}

// discriminator for CallExecuted
var CallExecutedDiscriminator = eventDiscriminator("CallExecuted")

func (e *CallExecuted) UnmarshalWithDecoder(d *borsh.Decoder) error {
	return d.Decode(e)
}

// BypasserCallExecuted matches the Anchor event:
// #[event] pub struct BypasserCallExecuted { index: u64, target: Pubkey, data: Vec<u8> }
type BypasserCallExecuted struct {
	Index  uint64
	Target solana.PublicKey
	Data   []byte
}

// discriminator for BypasserCallExecuted
var BypasserCallExecutedDiscriminator = eventDiscriminator("BypasserCallExecuted")

func (e *BypasserCallExecuted) UnmarshalWithDecoder(d *borsh.Decoder) error {
	return d.Decode(e)
}

// Cancelled matches the Anchor event:
// #[event] pub struct Cancelled { id: [u8;32] }
type Cancelled struct {
	ID [32]byte
}

// discriminator for Cancelled
var CancelledDiscriminator = eventDiscriminator("Cancelled")

func (e *Cancelled) UnmarshalWithDecoder(d *borsh.Decoder) error {
	return d.Decode(e)
}

// ParseTimelockEvents extracts and decodes Anchor events from tx.Meta.LogMessages.
func ParseTimelockEvents(tx *rpc.TransactionWithMeta) (*TxTimelockEvents, error) {
	out := &TxTimelockEvents{}

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

		d := borsh.NewDecoder(bytes.NewReader(payload))

		switch {
		case bytes.Equal(disc, CallScheduledDiscriminator[:]):
			var e CallScheduled
			if err := e.UnmarshalWithDecoder(d); err == nil {
				out.Scheduled = append(out.Scheduled, e)
			}

		case bytes.Equal(disc, CallExecutedDiscriminator[:]):
			var e CallExecuted
			if err := e.UnmarshalWithDecoder(d); err == nil {
				out.Executed = append(out.Executed, e)
			}

		case bytes.Equal(disc, BypasserCallExecutedDiscriminator[:]):
			var e BypasserCallExecuted
			if err := e.UnmarshalWithDecoder(d); err == nil {
				out.Bypasser = append(out.Bypasser, e)
			}

		case bytes.Equal(disc, CancelledDiscriminator[:]):
			var e Cancelled
			if err := e.UnmarshalWithDecoder(d); err == nil {
				out.Cancelled = append(out.Cancelled, e)
			}
		}
	}

	return out, nil
}

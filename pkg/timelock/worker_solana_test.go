// timelock/poll_signatures_test.go
package timelock

import (
	"context"
	"encoding/json"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/stretchr/testify/require"
)

// fakeSig is just enough of TransactionSignature to satisfy getSignaturesForAddress.
type fakeSig struct {
	Signature solana.Signature `json:"signature"`
	Slot      uint64           `json:"slot"`
}

func TestPollNewSignatures_Table(t *testing.T) {

	cases := []struct {
		name               string
		signaturesResponse []fakeSig
		txResponses        []map[string]interface{}
		wantTxCount        int
	}{
		{
			name:               "no new signatures",
			signaturesResponse: nil,
			txResponses:        nil,
			wantTxCount:        0,
		},
		{
			name: "single signature => single tx",
			signaturesResponse: []fakeSig{
				{Signature: solana.SignatureFromBytes([]byte{}), Slot: 10},
			},
			txResponses: []map[string]interface{}{{
				"slot": 10,
				"transaction": map[string]interface{}{
					"signatures": []string{"SigAAAA1111111111111111111111111111111111"},
					"message": map[string]interface{}{
						"accountKeys":  []string{"A"},
						"header":       map[string]interface{}{"numRequiredSignatures": 1, "numReadonlySignedAccounts": 0, "numReadonlyUnsignedAccounts": 0},
						"instructions": []interface{}{},
					},
				},
				"meta": map[string]interface{}{},
			}},
			wantTxCount: 1,
		},
		{
			name: "multiple signatures => multiple txs",
			signaturesResponse: []fakeSig{
				{Signature: solana.SignatureFromBytes([]byte{'1'}), Slot: 5},
				{Signature: solana.SignatureFromBytes([]byte{'2'}), Slot: 6},
			},
			txResponses: []map[string]interface{}{
				{
					"slot": 5,
					"transaction": map[string]interface{}{
						"signatures": []string{"Sig11111111111111111111111111111111111111"},
						"message":    map[string]interface{}{"accountKeys": []string{"A"}, "header": map[string]interface{}{"numRequiredSignatures": 1, "numReadonlySignedAccounts": 0, "numReadonlyUnsignedAccounts": 0}, "instructions": []interface{}{}},
					},
					"meta": map[string]interface{}{},
				},
				{
					"slot": 6,
					"transaction": map[string]interface{}{
						"signatures": []string{"Sig22222222222222222222222222222222222222"},
						"message":    map[string]interface{}{"accountKeys": []string{"B"}, "header": map[string]interface{}{"numRequiredSignatures": 1, "numReadonlySignedAccounts": 0, "numReadonlyUnsignedAccounts": 0}, "instructions": []interface{}{}},
					},
					"meta": map[string]interface{}{},
				},
			},
			wantTxCount: 2,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			var sigCalls, txCalls atomic.Int32

			mock := NewMockSolanaRPC(t, func(req rpcRequestSolana) (interface{}, error) {
				switch req.Method {
				case "getSignaturesForAddress":
					sigCalls.Add(1)
					return tc.signaturesResponse, nil

				case "getTransaction":
					txCalls.Add(1)
					// req.Params[0] is the signature string
					var sigStr string
					require.NoError(t, json.Unmarshal(req.Params[0], &sigStr))
					// find matching index
					idx := -1
					for i, fs := range tc.signaturesResponse {
						if fs.Signature.String() == sigStr {
							idx = i
							break
						}
					}
					if idx >= 0 {
						return tc.txResponses[idx], nil
					}
					return nil, nil

				default:
					t.Fatalf("unexpected method %s", req.Method)
					return nil, nil
				}
			})
			defer mock.Close()

			client := rpc.New(mock.URL)
			timelockKey, err := solana.NewRandomPrivateKey()
			require.NoError(t, err)
			w := &WorkerSolana{
				logger:             testLogger,
				solanaClient:       client,
				timelockProgramKey: timelockKey.PublicKey(),
				pollPeriod:         1,
				pollSize:           uint64(len(tc.signaturesResponse) + 1),
			}

			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()

			done, ch := w.pollNewSignatures(ctx)
			got := 0

		Loop:
			for {
				select {
				case tx, ok := <-ch:
					if !ok {
						break Loop
					}
					got++
					// re-extract the signature from tx.Transaction JSON
					raw, err := json.Marshal(tx.Transaction)
					require.NoError(t, err)
					var parsed struct {
						Signatures []string `json:"signatures"`
					}
					require.NoError(t, json.Unmarshal(raw, &parsed))
					exp := tc.txResponses[0]["transaction"].(map[string]interface{})["signatures"].([]string)
					require.Equal(t, exp, parsed.Signatures)
				case <-done:
					break Loop
				case <-ctx.Done():
					break Loop
				}
			}

			require.Equal(t, int32(2), sigCalls.Load(), "should call getSignaturesForAddress exactly once")
			require.Equal(t, int32(len(tc.signaturesResponse)), txCalls.Load(), "should call getTransaction once per signature")
			require.Equal(t, tc.wantTxCount, got, "number of transactions received")
		})
	}
}

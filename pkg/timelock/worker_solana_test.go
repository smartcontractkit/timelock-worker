// timelock/poll_signatures_test.go
package timelock

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/smartcontractkit/mcms/sdk/mocks"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// fakeSig is just enough of TransactionSignature to satisfy getSignaturesForAddress.
type fakeSig struct {
	Signature solana.Signature `json:"signature"`
	Slot      uint64           `json:"slot"`
}

func mustRandomSignature(t *testing.T) (solana.Signature, string) {
	var b [64]byte
	_, err := rand.Read(b[:])
	require.NoError(t, err)
	sig := solana.SignatureFromBytes(b[:])
	return sig, sig.String()
}
func TestPollNewSignatures(t *testing.T) {
	const (
		sigStrA = "3n8uFwJjTyBR3UqTGUjMncmzMJkjp7sk6uMvGCazgGNsCJpKaDxnKnUR3XNG2Exz4MyfpNHCEWGu2gZiSGGZVK3c"
		sigStr1 = "5eJiBS2dDCLdVZSLvXCLCeu3LL8eb9StAFmsDCpMTZZo8QAVDqAxowLqa5Yf2CtuwsAXeodDaUBgb63HGrj8Cxd6"
		sigStr2 = "2oCj3DD8iZ5YBJ7UbYiY2EM2kJrd1uPTftDbLbZbLmx1ybHJn2dcWxF9PPCfjVkTh2vYpGNP7dXZG74w4jHTHGSE"
	)
	var (
		sigA = solana.MustSignatureFromBase58(sigStrA)
		sig1 = solana.MustSignatureFromBase58(sigStr1)
		sig2 = solana.MustSignatureFromBase58(sigStr2)
	)
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
				{Signature: sigA, Slot: 10},
			},
			txResponses: []map[string]interface{}{{
				"slot": 10,
				"transaction": map[string]interface{}{
					"signatures": []string{sigStrA},
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
				{Signature: sig1, Slot: 5},
				{Signature: sig2, Slot: 6},
			},
			txResponses: []map[string]interface{}{
				{
					"slot": 5,
					"transaction": map[string]interface{}{
						"signatures": []string{sigStr1},
						"message": map[string]interface{}{
							"accountKeys":  []string{"A"},
							"header":       map[string]interface{}{"numRequiredSignatures": 1, "numReadonlySignedAccounts": 0, "numReadonlyUnsignedAccounts": 0},
							"instructions": []interface{}{},
						},
					},
					"meta": map[string]interface{}{},
				},
				{
					"slot": 6,
					"transaction": map[string]interface{}{
						"signatures": []string{sigStr2},
						"message": map[string]interface{}{
							"accountKeys":  []string{"B"},
							"header":       map[string]interface{}{"numRequiredSignatures": 1, "numReadonlySignedAccounts": 0, "numReadonlyUnsignedAccounts": 0},
							"instructions": []interface{}{},
						},
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
					var sigStr string
					require.NoError(t, json.Unmarshal(req.Params[0], &sigStr))

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
				pollSize:           len(tc.signaturesResponse) + 1,
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
					raw, err := json.Marshal(tx.Transaction)
					require.NoError(t, err)
					var parsed struct {
						Signatures []string `json:"signatures"`
					}
					require.NoError(t, json.Unmarshal(raw, &parsed))
					exp := tc.txResponses[len(tc.txResponses)-got]["transaction"].(map[string]interface{})["signatures"].([]string)
					require.Equal(t, exp, parsed.Signatures)

				case <-done:
					break Loop
				case <-ctx.Done():
					break Loop
				}
			}

			require.GreaterOrEqual(t, sigCalls.Load(), int32(1), "should call getSignaturesForAddress at least once per signature")
			require.GreaterOrEqual(t, txCalls.Load(), int32(len(tc.signaturesResponse)), "should call getTransaction at least once per signature")
			require.Equal(t, tc.wantTxCount, got, "number of transactions received")
		})
	}
}

func TestHandleEventCancelled(t *testing.T) {
	id := [32]byte{1, 2, 3}

	worker := &WorkerSolana{
		logger: testLogger,
	}

	worker.handleEventCancelled(context.Background(), Cancelled{ID: id})
	// TODO: once scheduler is added we can assert expectation from a mock scheduler here.
}

func TestHandleEventExecuted_Done(t *testing.T) {
	id := [32]byte{4, 5, 6}
	mockInsp := new(mocks.TimelockInspector)
	mockInsp.On("IsOperationDone", mock.Anything, "some.addr", id).Return(true, nil)

	worker := &WorkerSolana{
		timelockFullAddress: "some.addr",
		inspector:           mockInsp,
		logger:              testLogger,
	}
	event := CallExecuted{ID: id, Target: solana.PublicKey{}}
	err := worker.handleEventExecuted(context.Background(), event)
	require.NoError(t, err)
	mockInsp.AssertExpectations(t)
	// TODO: once scheduler is added we can assert expectation from a mock scheduler here.
}

func TestHandleEventScheduled_IsOp(t *testing.T) {
	id := [32]byte{7, 8, 9}
	mockInsp := new(mocks.TimelockInspector)
	mockInsp.On("IsOperationDone", mock.Anything, "some.addr", id).Return(false, nil)
	mockInsp.On("IsOperation", mock.Anything, "some.addr", id).Return(true, nil)

	worker := &WorkerSolana{
		timelockFullAddress: "some.addr",
		inspector:           mockInsp,
		logger:              testLogger,
	}
	event := CallScheduled{ID: id, Target: solana.PublicKey{}}
	err := worker.handleEventScheduled(context.Background(), event)
	require.NoError(t, err)
	mockInsp.AssertExpectations(t)
}

func TestHandleEventScheduled_OperationDone(t *testing.T) {
	id := [32]byte{10, 11, 12}
	mockInsp := new(mocks.TimelockInspector)
	mockInsp.On("IsOperationDone", mock.Anything, "some.addr", id).Return(true, nil)

	worker := &WorkerSolana{
		timelockFullAddress: "some.addr",
		inspector:           mockInsp,
		logger:              testLogger,
	}
	event := CallScheduled{ID: id, Target: solana.PublicKey{}}
	err := worker.handleEventScheduled(context.Background(), event)
	require.NoError(t, err)
	mockInsp.AssertExpectations(t)
}

func TestHandleEventExecuted_NotDone(t *testing.T) {
	id := [32]byte{13, 14, 15}
	mockInsp := new(mocks.TimelockInspector)
	mockInsp.On("IsOperationDone", mock.Anything, "some.addr", id).Return(false, nil)

	worker := &WorkerSolana{
		timelockFullAddress: "some.addr",
		inspector:           mockInsp,
		logger:              testLogger,
	}
	event := CallExecuted{ID: id, Target: solana.PublicKey{}}
	err := worker.handleEventExecuted(context.Background(), event)
	require.NoError(t, err)
	mockInsp.AssertExpectations(t)
	// TODO: once scheduler is added we can assert expectation from a mock scheduler here.
}

func TestHandleEventScheduled_IsOp_Error(t *testing.T) {
	id := [32]byte{16, 17, 18}
	mockInsp := new(mocks.TimelockInspector)
	mockInsp.On("IsOperationDone", mock.Anything, "some.addr", id).Return(false, nil)
	mockInsp.On("IsOperation", mock.Anything, "some.addr", id).Return(false, errors.New("boom"))

	worker := &WorkerSolana{
		timelockFullAddress: "some.addr",
		inspector:           mockInsp,
		logger:              testLogger,
	}
	event := CallScheduled{ID: id, Target: solana.PublicKey{}}
	require.ErrorContains(t, worker.handleEventScheduled(context.Background(), event), "timelock.isOperation call failed")
}

func TestHandleEventExecuted_Error(t *testing.T) {
	id := [32]byte{19, 20, 21}
	mockInsp := new(mocks.TimelockInspector)
	mockInsp.On("IsOperationDone", mock.Anything, "some.addr", id).Return(false, errors.New("bad state"))

	worker := &WorkerSolana{
		timelockFullAddress: "some.addr",
		inspector:           mockInsp,
		logger:              testLogger,
	}
	event := CallExecuted{ID: id, Target: solana.PublicKey{}}
	require.ErrorContains(t, worker.handleEventExecuted(context.Background(), event), "timelock.isOperationDone call failed")
	// TODO: once scheduler is added we can assert expectation from a mock scheduler here.
}

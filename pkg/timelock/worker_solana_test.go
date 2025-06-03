// timelock/poll_signatures_test.go
package timelock

import (
	"context"
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

func assertExpectedTransactions(t *testing.T, ch <-chan *rpc.TransactionWithMeta, done <-chan struct{}, ctx context.Context, want []map[string]any) int {
	t.Helper()
	got := 0

	for {
		select {
		case tx, ok := <-ch:
			if !ok {
				return got
			}
			got++
			raw, err := json.Marshal(tx.Transaction)
			require.NoError(t, err)

			var parsed struct {
				Signatures []string `json:"signatures"`
			}
			require.NoError(t, json.Unmarshal(raw, &parsed))

			exp := want[len(want)-got]["transaction"].(map[string]any)["signatures"].([]string)
			require.Equal(t, exp, parsed.Signatures)

		case <-done:
			return got
		case <-ctx.Done():
			return got
		}
	}
}

func TestStartPolling(t *testing.T) {
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
		txResponses        []map[string]any
		wantTxCount        int
	}{
		{
			name:               "no new signatures",
			signaturesResponse: nil,
			txResponses:        nil,
			wantTxCount:        0,
		},
		{
			name:               "single signature => single tx",
			signaturesResponse: []fakeSig{{Signature: sigA, Slot: 10}},
			txResponses: []map[string]any{{
				"slot": 10,
				"transaction": map[string]any{
					"signatures": []string{sigStrA},
					"message": map[string]any{
						"accountKeys":  []string{"A"},
						"header":       map[string]any{"numRequiredSignatures": 1, "numReadonlySignedAccounts": 0, "numReadonlyUnsignedAccounts": 0},
						"instructions": []any{},
					},
				},
				"meta": map[string]any{},
			}},
			wantTxCount: 1,
		},
		{
			name:               "multiple signatures => multiple txs",
			signaturesResponse: []fakeSig{{Signature: sig1, Slot: 5}, {Signature: sig2, Slot: 6}},
			txResponses: []map[string]any{
				{
					"slot": 5,
					"transaction": map[string]any{
						"signatures": []string{sigStr1},
						"message": map[string]any{
							"accountKeys":  []string{"A"},
							"header":       map[string]any{"numRequiredSignatures": 1, "numReadonlySignedAccounts": 0, "numReadonlyUnsignedAccounts": 0},
							"instructions": []any{},
						},
					},
					"meta": map[string]any{},
				},
				{
					"slot": 6,
					"transaction": map[string]any{
						"signatures": []string{sigStr2},
						"message": map[string]any{
							"accountKeys":  []string{"B"},
							"header":       map[string]any{"numRequiredSignatures": 1, "numReadonlySignedAccounts": 0, "numReadonlyUnsignedAccounts": 0},
							"instructions": []any{},
						},
					},
					"meta": map[string]any{},
				},
			},
			wantTxCount: 2,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var sigCalls, txCalls atomic.Int32

			mockRPC := NewMockSolanaRPC(t, func(req rpcRequestSolana) (any, error) {
				switch req.Method {
				case "getSignaturesForAddress":
					sigCalls.Add(1)
					return tc.signaturesResponse, nil
				case "getTransaction":
					txCalls.Add(1)
					var sigStr string
					require.NoError(t, json.Unmarshal(req.Params[0], &sigStr))
					for i, fs := range tc.signaturesResponse {
						if fs.Signature.String() == sigStr {
							return tc.txResponses[i], nil
						}
					}
					return nil, nil
				default:
					t.Fatalf("unexpected method %s", req.Method)
					return nil, nil
				}
			})
			defer mockRPC.Close()

			client := rpc.New(mockRPC.URL)
			timelockKey, err := solana.NewRandomPrivateKey()
			require.NoError(t, err)

			w := &WorkerSolana{
				logger:             testLogger,
				solanaClient:       client,
				timelockProgramKey: timelockKey.PublicKey(),
				pollPeriod:         1,
				pollSize:           len(tc.signaturesResponse) + 1,
			}

			ctx, cancel := context.WithTimeout(t.Context(), 2*time.Second)
			defer cancel()

			done, ch := w.startPolling(ctx)

			got := assertExpectedTransactions(t, ch, done, ctx, tc.txResponses)

			require.GreaterOrEqual(t, sigCalls.Load(), int32(1), "should call getSignaturesForAddress")
			require.GreaterOrEqual(t, txCalls.Load(), int32(len(tc.signaturesResponse)), "should call getTransaction")
			require.Equal(t, tc.wantTxCount, got, "received tx count")
		})
	}
}

func TestHandleEventCancelled(t *testing.T) {
	id := [32]byte{1, 2, 3}
	s := newMockScheduler(t)
	worker := &WorkerSolana{
		logger:    testLogger,
		scheduler: s,
	}
	s.On("delFromScheduler", mock.Anything).Return(nil)
	worker.handleEventCancelled(t.Context(), SolanaTimelockCallCancelledEvent{ID: id})
}

func TestHandleEventExecuted_Done(t *testing.T) {
	id := [32]byte{4, 5, 6}
	mockInsp := new(mocks.TimelockInspector)
	mockInsp.On("IsOperationDone", mock.Anything, "some.addr", id).Return(true, nil)
	s := newMockScheduler(t)
	worker := &WorkerSolana{
		timelockFullAddress: "some.addr",
		inspector:           mockInsp,
		scheduler:           s,
		logger:              testLogger,
	}
	s.On("delFromScheduler", mock.Anything).Return(nil)
	event := SolanaTimelockCallExecutedEvent{ID: id, Target: solana.PublicKey{}}
	err := worker.handleEventExecuted(t.Context(), event)
	require.NoError(t, err)
}

func TestHandleEventScheduled_IsOp(t *testing.T) {
	id := [32]byte{7, 8, 9}
	mockInsp := new(mocks.TimelockInspector)
	mockInsp.On("IsOperationDone", mock.Anything, "some.addr", id).Return(false, nil)
	mockInsp.On("IsOperation", mock.Anything, "some.addr", id).Return(true, nil)
	s := newMockScheduler(t)
	worker := &WorkerSolana{
		timelockFullAddress: "some.addr",
		inspector:           mockInsp,
		scheduler:           s,
		logger:              testLogger,
	}
	s.On("addToScheduler", mock.Anything).Return(nil)
	event := SolanaTimelockCallScheduledEvent{ID: id, Target: solana.PublicKey{}}
	err := worker.handleEventScheduled(t.Context(), event)
	require.NoError(t, err)
}

func TestHandleEventScheduled_OperationDone(t *testing.T) {
	id := [32]byte{10, 11, 12}
	mockInsp := new(mocks.TimelockInspector)
	mockInsp.On("IsOperationDone", mock.Anything, "some.addr", id).Return(true, nil)

	worker := &WorkerSolana{
		timelockFullAddress: "some.addr",
		inspector:           mockInsp,
		scheduler:           newMockScheduler(t),
		logger:              testLogger,
	}
	event := SolanaTimelockCallScheduledEvent{ID: id, Target: solana.PublicKey{}}
	err := worker.handleEventScheduled(t.Context(), event)
	require.NoError(t, err)
}

func TestHandleEventExecuted_NotDone(t *testing.T) {
	id := [32]byte{13, 14, 15}
	mockInsp := new(mocks.TimelockInspector)
	mockInsp.On("IsOperationDone", mock.Anything, "some.addr", id).Return(false, nil)

	worker := &WorkerSolana{
		timelockFullAddress: "some.addr",
		inspector:           mockInsp,
		scheduler:           newMockScheduler(t),
		logger:              testLogger,
	}
	event := SolanaTimelockCallExecutedEvent{ID: id, Target: solana.PublicKey{}}
	err := worker.handleEventExecuted(t.Context(), event)
	require.NoError(t, err)
}

func TestHandleEventScheduled_IsOp_Error(t *testing.T) {
	id := [32]byte{16, 17, 18}
	mockInsp := new(mocks.TimelockInspector)
	mockInsp.On("IsOperationDone", mock.Anything, "some.addr", id).Return(false, nil)
	mockInsp.On("IsOperation", mock.Anything, "some.addr", id).Return(false, errors.New("boom"))

	worker := &WorkerSolana{
		timelockFullAddress: "some.addr",
		inspector:           mockInsp,
		scheduler:           newMockScheduler(t),
		logger:              testLogger,
	}
	event := SolanaTimelockCallScheduledEvent{ID: id, Target: solana.PublicKey{}}
	require.ErrorContains(t, worker.handleEventScheduled(t.Context(), event), "timelock.isOperation call failed")
}

func TestHandleEventExecuted_Error(t *testing.T) {
	id := [32]byte{19, 20, 21}
	mockInsp := new(mocks.TimelockInspector)
	mockInsp.On("IsOperationDone", mock.Anything, "some.addr", id).Return(false, errors.New("bad state"))

	worker := &WorkerSolana{
		timelockFullAddress: "some.addr",
		inspector:           mockInsp,
		scheduler:           newMockScheduler(t),
		logger:              testLogger,
	}
	event := SolanaTimelockCallExecutedEvent{ID: id, Target: solana.PublicKey{}}
	require.ErrorContains(t, worker.handleEventExecuted(t.Context(), event), "timelock.isOperationDone call failed")
}

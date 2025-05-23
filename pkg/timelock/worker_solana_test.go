// timelock_solana_test.go
package timelock

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	"go.uber.org/zap"
)

func TestPollNewSignatures_Basic(t *testing.T) {
	step := 0

	srv := NewMockSolanaRPC(t, func(req rpcRequestSolana) (interface{}, error) {
		switch req.Method {
		case "getSignaturesForAddress":
			// inspect if client sent an "until" filter
			var opts map[string]interface{}
			if len(req.Params) > 1 {
				_ = json.Unmarshal(req.Params[1], &opts)
			}
			if step == 0 {
				step = 1
				return []map[string]interface{}{
					{"signature": "sig0"},
				}, nil
			}
			return []map[string]interface{}{
				{"signature": "sig1"},
			}, nil

		case "getConfirmedTransaction":
			return map[string]interface{}{
				"meta": map[string]interface{}{
					"logMessages": []string{
						`Program log: EVENT:CallScheduled:{"id":"id1","index":0,"target":"11111111111111111111111111111111","predecessor":"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA","salt":"BBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBB","delay":5,"data":"Zm9v"}`,
					},
				},
			}, nil

		default:
			t.Fatalf("unexpected RPC method: %s", req.Method)
			return nil, nil
		}
	})
	defer srv.Close()

	client := rpc.New(srv.URL)
	programKey, _ := solana.PublicKeyFromBase58("11111111111111111111111111111111")

	w := &WorkerSolana{
		solanaClient:       client,
		timelockProgramKey: programKey,
		pollPeriod:         1, // 1 second
		pollSize:           1, // fetch 1 sig per poll
		logger:             zap.NewExample().Sugar(),
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	done, txCh := w.pollNewSignatures(ctx)

	select {
	case tx := <-txCh:
		if tx == nil || tx.Meta == nil {
			t.Fatal("expected non-nil TransactionWithMeta.Meta")
		}
		if len(tx.Meta.LogMessages) != 1 {
			t.Fatalf("expected 1 log message, got %d", len(tx.Meta.LogMessages))
		}
		if !strings.Contains(tx.Meta.LogMessages[0], "EVENT:CallScheduled") {
			t.Errorf("unexpected log content: %q", tx.Meta.LogMessages[0])
		}
	case <-ctx.Done():
		t.Fatal("timed out waiting for a transaction")
	}

	// cancel and ensure the goroutine exits
	cancel()
	<-done
}

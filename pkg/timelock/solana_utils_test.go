package timelock

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

// rpcRequestSolana is what Solana JSON‐RPC looks like under the hood.
type rpcRequestSolana struct {
	JSONRPC string            `json:"jsonrpc"`
	ID      interface{}       `json:"id"`
	Method  string            `json:"method"`
	Params  []json.RawMessage `json:"params"`
}

// rpcHandler is your test’s little “router” for each incoming request.
type rpcHandler func(req rpcRequestSolana) (result interface{}, err error)

// NewMockSolanaRPC spins up a httptest.Server that will:
//
//   - if it sees a single JSON‐RPC object, decode into rpcRequestSolana
//   - if it sees a JSON array, decode into []rpcRequestSolana
//   - call once per rpcRequestSolana
//   - produce either a single JSON‐RPC response or a JSON‐RPC batch
func NewMockSolanaRPC(t *testing.T, handler rpcHandler) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("read body: %v", err)
		}

		// are we a batch (starts with “[”)?
		isBatch := len(body) > 0 && body[0] == '['

		if isBatch {
			// decode into slice
			var reqs []rpcRequestSolana
			if err := json.Unmarshal(body, &reqs); err != nil {
				t.Fatalf("unmarshal batch: %v\n--> %s", err, string(body))
			}
			// build a parallel slice of responses
			var batchResp []map[string]interface{}
			for _, req := range reqs {
				rpcObj := map[string]interface{}{
					"jsonrpc": "2.0",
					"id":      req.ID,
				}
				if res, err := handler(req); err != nil {
					rpcObj["error"] = map[string]interface{}{
						"code":    -32000,
						"message": err.Error(),
					}
				} else {
					rpcObj["result"] = res
				}
				batchResp = append(batchResp, rpcObj)
			}
			w.Header().Set("Content-Type", "application/json")
			if err := json.NewEncoder(w).Encode(batchResp); err != nil {
				t.Fatalf("encode batch resp: %v", err)
			}

			return
		}

		// single‐call
		var req rpcRequestSolana
		if err := json.Unmarshal(body, &req); err != nil {
			t.Fatalf("unmarshal single: %v\n--> %s", err, string(body))
		}
		resp := map[string]interface{}{
			"jsonrpc": "2.0",
			"id":      req.ID,
		}
		if res, err := handler(req); err != nil {
			resp["error"] = map[string]interface{}{
				"code":    -32000,
				"message": err.Error(),
			}
		} else {
			resp["result"] = res
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			t.Fatalf("encode resp: %v", err)
		}
	}))
}

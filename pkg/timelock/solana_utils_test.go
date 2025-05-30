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

		w.Header().Set("Content-Type", "application/json")

		var (
			reqs []rpcRequestSolana
			resp any
		)

		if isBatchRequest(body) {
			if err := json.Unmarshal(body, &reqs); err != nil {
				t.Fatalf("unmarshal batch: %v\n--> %s", err, string(body))
			}
			resp = buildBatchResponses(reqs, handler)
		} else {
			var req rpcRequestSolana
			if err := json.Unmarshal(body, &req); err != nil {
				t.Fatalf("unmarshal single request: %v\n--> %s", err, string(body))
			}
			resp = buildRPCResponse(req, handler)
		}

		if err := json.NewEncoder(w).Encode(resp); err != nil {
			t.Fatalf("encode response: %v", err)
		}
	}))
}

func isBatchRequest(body []byte) bool {
	return len(body) > 0 && body[0] == '['
}

func buildBatchResponses(reqs []rpcRequestSolana, handler rpcHandler) []map[string]any {
	responses := make([]map[string]any, 0, len(reqs))
	for _, req := range reqs {
		responses = append(responses, buildRPCResponse(req, handler))
	}
	return responses
}

func buildRPCResponse(req rpcRequestSolana, handler rpcHandler) map[string]any {
	resp := map[string]any{
		"jsonrpc": "2.0",
		"id":      req.ID,
	}
	if result, err := handler(req); err != nil {
		resp["error"] = map[string]any{
			"code":    -32000,
			"message": err.Error(),
		}
	} else {
		resp["result"] = result
	}
	return resp
}

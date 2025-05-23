// server_mock_test.go
package timelock

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// rpcRequestSolana is the minimal shape of a JSON-RPC request solana cares about.
type rpcRequestSolana struct {
	Method string            `json:"method"`
	Params []json.RawMessage `json:"params"`
	ID     interface{}       `json:"id"`
}

// rpcHandler is called for each incoming request; return either
// a successful result (under "result") or an error string (under "error").
type rpcHandler func(req rpcRequestSolana) (result interface{}, err error)

// NewMockSolanaRPC spins up an httptest.Server that implements
// Solana JSON-RPC by delegating to handler function.
func NewMockSolanaRPC(t *testing.T, handler rpcHandler) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req rpcRequestSolana
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("invalid JSON-RPC request: %v", err)
		}
		// build the base response envelope
		resp := map[string]interface{}{
			"jsonrpc": "2.0",
			"id":      req.ID,
		}
		// call user handler
		if result, err := handler(req); err != nil {
			resp["error"] = err.Error()
		} else {
			resp["result"] = result
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			t.Fatalf("failed to encode JSON-RPC response: %v", err)
		}
	}))
}

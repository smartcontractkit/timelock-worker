package timelock

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

// rpcRequestSolana is what Solana JSON‐RPC looks like under the hood.
type rpcRequestSolana struct {
	JSONRPC string            `json:"jsonrpc"`
	ID      any               `json:"id"`
	Method  string            `json:"method"`
	Params  []json.RawMessage `json:"params"`
}

// rpcHandler is your test’s little “router” for each incoming request.
type rpcHandler func(req rpcRequestSolana) (result any, err error)

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
		require.NoError(t, err, "read request body")

		w.Header().Set("Content-Type", "application/json")

		var (
			reqs []rpcRequestSolana
			resp any
		)

		if isBatchRequest(body) {
			err = json.Unmarshal(body, &reqs)
			require.NoError(t, err, "unmarshal batch request")
			resp = buildBatchResponses(reqs, handler)
		} else {
			var req rpcRequestSolana
			err := json.Unmarshal(body, &req)
			require.NoError(t, err, "unmarshal single request")
			resp = buildRPCResponse(req, handler)
		}

		err = json.NewEncoder(w).Encode(resp)
		require.NoError(t, err, "encode response")

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

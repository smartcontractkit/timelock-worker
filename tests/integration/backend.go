package integration

import (
	"context"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/ethclient/simulated"
	"github.com/stretchr/testify/require"
)

type Backend interface {
	simulated.Client
	Commit() common.Hash
}

// --- geth simulated backend ---

type simulatedBackendWrapper struct {
	backend *simulated.Backend
	simulated.Client
}

func (w *simulatedBackendWrapper) Commit() common.Hash {
	return w.backend.Commit()
}

func NewSimulatedBackend(t *testing.T, genesisAlloc types.GenesisAlloc) Backend {
	t.Helper()

	gethBackend := simulated.NewBackend(genesisAlloc)

	return &simulatedBackendWrapper{
		backend: gethBackend,
		Client:  gethBackend.Client(),
	}
}

// --- rpc backend ---

type rpcBackend struct {
	simulated.Client
}

func (b *rpcBackend) Commit() common.Hash {
	return common.Hash{} // nop
}

func NewRPCBackend(t *testing.T, ctx context.Context, url string) Backend {
	t.Helper()

	client, err := ethclient.DialContext(ctx, url)
	require.NoError(t, err)

	t.Cleanup(client.Close)

	return &rpcBackend{Client: client}
}

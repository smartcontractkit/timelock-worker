package integration

import (
	"crypto/ecdsa"
	"fmt"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/require"
)

// TestAccount is data type wrapping attributes typically needed when managing
// ethereum accounts.
type TestAccount struct {
	address       common.Address
	hexAddress    string
	privateKey    *ecdsa.PrivateKey
	hexPrivateKey string
}

func (ta TestAccount) String() string {
	return fmt.Sprintf("TestAccount{address: %s, privateKey: %s}", ta.hexAddress, ta.hexPrivateKey)
}

// NewTestAccount generates a new ecdsa key and returns a TestAccount structure
// with the associated attributes.
func NewTestAccount(t *testing.T) TestAccount {
	t.Helper()

	privateKey, err := crypto.GenerateKey()
	require.NoError(t, err)

	publicKeyECDSA, ok := privateKey.Public().(*ecdsa.PublicKey)
	require.True(t, ok)

	address := crypto.PubkeyToAddress(*publicKeyECDSA)

	return TestAccount{
		address:       address,
		hexAddress:    address.Hex(),
		privateKey:    privateKey,
		hexPrivateKey: hexutil.Encode(crypto.FromECDSA(privateKey))[2:],
	}
}

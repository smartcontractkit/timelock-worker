package evm

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
	Address       common.Address
	HexAddress    string
	PrivateKey    *ecdsa.PrivateKey
	HexPrivateKey string
}

func (ta TestAccount) String() string {
	return fmt.Sprintf("TestAccount{address: %s, privateKey: %s}", ta.HexAddress, ta.HexPrivateKey)
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
		Address:       address,
		HexAddress:    address.Hex(),
		PrivateKey:    privateKey,
		HexPrivateKey: hexutil.Encode(crypto.FromECDSA(privateKey))[2:],
	}
}

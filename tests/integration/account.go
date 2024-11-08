package integration

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

// Account #1.
var (
	account1         = common.HexToAddress("0x8aAfF1544c195Aa8719Fabd67253Fc366d38A563")
	privateKey1      = "1921763610b80b2b147ca676c775c172ba6c037f6ba135792bed6bc458c660f0"
	ecdsaPrivateKey1 = Must(crypto.HexToECDSA(privateKey1))
	keyFileJson1     = `{
		"version": 3,
		"id": "a3ece80c-e259-4d3b-9125-b3fd9d7f0f57",
		"address": "8aaff1544c195aa8719fabd67253fc366d38a563",
		"crypto": {
			"cipher": "aes-128-ctr",
			"ciphertext": "6b3d12ab4e6b9e9c2df69d358d1e30d1dc51268fc8e72d20b866073e6d945e38",
			"cipherparams": { "iv": "662529e8442f0c2ca108341b76e2e879" },
			"mac": "565585b4870acd67b6bb8a52db30a47316bc7f216fb98eb40e630f9801e878a8",
			"kdf": "scrypt",
			"kdfparams": {
				"dklen": 32,
				"n": 262144,
				"p": 1,
				"r": 8,
				"salt": "248621f31c264689c45956840ea18c8b9f109088dafbf99dfaaa491bbd4f1e8b"
			}
		}
	}`
)

func Must[T any](value T, err error) T {
	if err != nil {
		panic(err)
	}
	return value
}

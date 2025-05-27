package solana

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

//go:generate ./compile-timelock-programs.sh
func TestSolanaIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(solanaIntegrationTestSuite))
}

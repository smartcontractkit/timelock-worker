package evm

import (
	"context"
	"crypto/ecdsa"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/stretchr/testify/suite"

	"github.com/smartcontractkit/timelock-worker/tests/containers"
)

type integrationTestSuite struct {
	suite.Suite

	GethContainer *containers.GethContainer
	Ctx           context.Context //nolint:containedctx
}

func (s *integrationTestSuite) SetupSuite() {
	var err error
	s.Ctx = context.Background()
	s.GethContainer, err = containers.NewGethContainer(s.Ctx)
	s.Require().NoError(err)
}

func (s *integrationTestSuite) TearDownSuite() {
	err := s.GethContainer.Teardown(context.Background())
	s.Require().NoError(err)
}

func (s *integrationTestSuite) Log(format string) {
	s.T().Log(format)
}

func (s *integrationTestSuite) Logf(format string, args ...any) {
	s.T().Logf(format, args...)
}

func (s *integrationTestSuite) KeyedTransactor(privateKey *ecdsa.PrivateKey, chainID *big.Int) *bind.TransactOpts {
	if chainID == nil {
		chainID = new(big.Int).SetUint64(s.GethContainer.ChainID)
	}

	transactor, err := bind.NewKeyedTransactorWithChainID(privateKey, chainID)
	s.Require().NoError(err)

	// reset gas parameters and let go-ethereum estimate and set them automatically
	transactor.Value = nil
	transactor.GasPrice = nil
	transactor.GasFeeCap = nil
	transactor.GasTipCap = nil
	transactor.GasLimit = 0

	return transactor
}

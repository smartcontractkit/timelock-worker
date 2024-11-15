package integration

import (
	"context"
	"crypto/ecdsa"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	contracts "github.com/smartcontractkit/ccip-owner-contracts/gethwrappers"
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

func (s *integrationTestSuite) DeployTimelock(
	ctx context.Context, transactor *bind.TransactOpts, client *ethclient.Client,
	adminAccount common.Address, minDelay *big.Int,
) (
	common.Address, *types.Transaction, *types.Receipt, *contracts.RBACTimelock,
) {
	proposers := []common.Address{adminAccount}
	executors := []common.Address{adminAccount}
	cancellers := []common.Address{adminAccount}
	bypassers := []common.Address{}

	address, transaction, contract, err := contracts.DeployRBACTimelock(
		transactor, client, minDelay, adminAccount, proposers, executors, cancellers, bypassers)
	s.Require().NoError(err)

	receipt, err := bind.WaitMined(ctx, client, transaction)
	s.Require().NoError(err)
	s.Require().Equal(types.ReceiptStatusSuccessful, receipt.Status)

	s.Logf("timelock address: %v; deploy transaction: %v", address, transaction.Hash())

	return address, transaction, receipt, contract
}

func (s *integrationTestSuite) DeployCallProxy(
	ctx context.Context, transactor *bind.TransactOpts, client *ethclient.Client, timelockAddress common.Address,
) (
	common.Address, *types.Transaction, *types.Receipt, *contracts.CallProxy,
) {
	address, transaction, contract, err := contracts.DeployCallProxy(
		transactor, client, timelockAddress)
	s.Require().NoError(err)

	receipt, err := bind.WaitMined(ctx, client, transaction)
	s.Require().NoError(err)
	s.Require().Equal(types.ReceiptStatusSuccessful, receipt.Status)

	s.Logf("call proxy address: %v; deploy transaction: %v", address, transaction.Hash())

	return address, transaction, receipt, contract
}

func (s *integrationTestSuite) UpdateDelay(
	ctx context.Context, transactor *bind.TransactOpts, client *ethclient.Client,
	timelockContract *contracts.RBACTimelock, delay *big.Int,
) (
	*types.Transaction, *types.Receipt,
) {
	transaction, err := timelockContract.UpdateDelay(transactor, delay)
	s.Require().NoError(err)

	receipt, err := bind.WaitMined(ctx, client, transaction)
	s.Require().NoError(err)
	s.Require().Equal(types.ReceiptStatusSuccessful, receipt.Status)

	s.Logf("update delay transaction: %v", transaction.Hash())

	return transaction, receipt
}

func (s *integrationTestSuite) ScheduleBatch(
	ctx context.Context, transactor *bind.TransactOpts, client *ethclient.Client,
	timelockContract *contracts.RBACTimelock, calls []contracts.RBACTimelockCall,
	predecessor [32]byte, salt [32]byte, delay *big.Int,
) (
	*types.Transaction, *types.Receipt,
) {
	transaction, err := timelockContract.ScheduleBatch(transactor, calls, predecessor, salt, delay)
	s.Require().NoError(err)

	receipt, err := bind.WaitMined(ctx, client, transaction)
	s.Require().NoError(err)
	s.Require().Equal(receipt.Status, types.ReceiptStatusSuccessful)

	s.Logf("schedule batch transaction: %v", transaction.Hash())
	return transaction, receipt
}

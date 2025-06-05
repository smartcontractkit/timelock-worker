package solana

import (
	"context"
	"path/filepath"
	"runtime"
	"time"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/stretchr/testify/suite"

	timelockutils "github.com/smartcontractkit/chainlink-ccip/chains/solana/utils/timelock"
	"github.com/smartcontractkit/chainlink-testing-framework/framework"
	"github.com/smartcontractkit/chainlink-testing-framework/framework/components/blockchain"
)

// privateKey this key matches the public key in the config.toml so it gets funded by the genesis block.
const privateKey = "DmPfeHBC8Brf8s5qQXi25bmJ996v6BHRtaLc6AH51yFGSqQpUMy1oHkbbXobPNBdgGH2F29PAmoq9ZZua4K9vCc"

var (
	_, fileName, _, _ = runtime.Caller(0)
	ProjectRoot       = filepath.Dir(filepath.Dir(filepath.Dir(fileName)))
)

// Config defines the blockchain configuration.
type Config struct {
	SolanaChain *blockchain.Input `toml:"solana_config"`
}
type solanaIntegrationTestSuite struct {
	suite.Suite
	Ctx                       context.Context //nolint:containedctx
	solanaBlockchain          *blockchain.Output
	solanaClient              *rpc.Client
	TestPrivateKey            solana.PrivateKey
	McmProgramID              solana.PublicKey
	TimelockProgramID         solana.PublicKey
	StubProgramID             solana.PublicKey
	RmnRemoteProgramID        solana.PublicKey
	AccessControllerProgramID solana.PublicKey
	RoleMap                   timelockutils.RoleMap
	ProposerAccessController  solana.PublicKey
	ExecutorAccessController  solana.PublicKey
	CancellerAccessController solana.PublicKey
	BypasserAccessController  solana.PublicKey
}

func (s *solanaIntegrationTestSuite) SetupSuite() {
	var err error
	s.Ctx = s.T().Context()
	in, err := framework.Load[Config](s.T())
	s.Require().NoError(err, "Failed to load Solana configuration")
	if in.SolanaChain.ContractsDir == "" {
		in.SolanaChain.ContractsDir = filepath.Join(ProjectRoot, "integration/solana/artifacts")
	}

	// Initialize Solana client
	s.solanaBlockchain, err = blockchain.NewBlockchainNetwork(in.SolanaChain)
	s.Require().NoError(err, "Failed to initialize Solana blockchain network")
	s.solanaClient = rpc.New(s.solanaBlockchain.Nodes[0].HostHTTPUrl)

	// Test the connection by checking the health of the RPC node
	health, err := s.solanaClient.GetHealth(s.Ctx)
	s.Require().NoError(err)
	if health == rpc.HealthOk {
		s.Log("Connection to Solana RPC is successful!")
	} else {
		s.T().Fatal("Connection established, but node health is not OK.")
	}

	s.McmProgramID = solana.MustPublicKeyFromBase58(in.SolanaChain.SolanaPrograms["mcm"])
	s.TimelockProgramID = solana.MustPublicKeyFromBase58(in.SolanaChain.SolanaPrograms["timelock"])
	s.StubProgramID = solana.MustPublicKeyFromBase58(in.SolanaChain.SolanaPrograms["external_program_cpi_stub"])
	s.RmnRemoteProgramID = solana.MustPublicKeyFromBase58(in.SolanaChain.SolanaPrograms["rmn_remote"])
	s.AccessControllerProgramID = solana.MustPublicKeyFromBase58(in.SolanaChain.SolanaPrograms["access_controller"])
	s.TestPrivateKey = solana.MustPrivateKeyFromBase58(privateKey)

}

func (s *solanaIntegrationTestSuite) TearDownSuite() {
	err := s.solanaClient.Close()
	s.Require().NoError(err)
}

func (s *solanaIntegrationTestSuite) Log(format string) {
	s.T().Log(format)
}

func (s *solanaIntegrationTestSuite) Logf(format string, args ...any) {
	s.T().Logf(format, args...)
}

func (s *solanaIntegrationTestSuite) withTimeout(timeout time.Duration) {
	savedContext := s.T().Context()
	var cancel context.CancelFunc
	s.Ctx, cancel = context.WithTimeout(savedContext, timeout)

	s.T().Cleanup(func() {
		cancel()
		s.Ctx = savedContext
	})
}

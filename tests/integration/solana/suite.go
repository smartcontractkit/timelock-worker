package solana

import (
	"context"
	"path/filepath"
	"runtime"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/smartcontractkit/chainlink-ccip/chains/solana/utils/timelock"
	timelockutils "github.com/smartcontractkit/chainlink-ccip/chains/solana/utils/timelock"
	"github.com/smartcontractkit/chainlink-testing-framework/framework"
	"github.com/smartcontractkit/chainlink-testing-framework/framework/components/blockchain"
	"github.com/stretchr/testify/suite"
)

// privateKey this key matches the public key in the config.toml so it gets funded by the genesis block.
const privateKey = "DmPfeHBC8Brf8s5qQXi25bmJ996v6BHRtaLc6AH51yFGSqQpUMy1oHkbbXobPNBdgGH2F29PAmoq9ZZua4K9vCc"

var _, fileName, _, _ = runtime.Caller(0)
var ProjectRoot = filepath.Dir(filepath.Dir(filepath.Dir(fileName)))

// Config defines the blockchain configuration.
type Config struct {
	SolanaChain *blockchain.Input `toml:"solana_config"`
}
type solanaIntegrationTestSuite struct {
	suite.Suite
	solanaBlockchain          *blockchain.Output
	solanaClient              *rpc.Client
	TimelockProgramID         solana.PublicKey
	StubProgramID             solana.PublicKey
	AccessControllerProgramID solana.PublicKey
	TestPrivateKey            solana.PrivateKey
	RoleMap                   timelock.RoleMap
	Ctx                       context.Context //nolint:containedctx
}

func (s *solanaIntegrationTestSuite) SetupSuite() {
	var err error
	t := s.T()
	s.Ctx = s.T().Context()
	in, err := framework.Load[Config](s.T())
	s.Require().NoError(err, "Failed to load Solana configuration")
	if in.SolanaChain.ContractsDir == "" {
		in.SolanaChain.ContractsDir = filepath.Join(ProjectRoot, "integration/solana/artifacts")
	}
	// Initialize Solana client
	solanaBlockChainOutput, err := blockchain.NewBlockchainNetwork(in.SolanaChain)
	s.solanaBlockchain = solanaBlockChainOutput
	s.Require().NoError(err, "Failed to initialize Solana blockchain network")
	solanaClient := rpc.New(solanaBlockChainOutput.Nodes[0].HostHTTPUrl)
	s.solanaClient = solanaClient
	// Test the connection by checking the health of the RPC node
	var health string
	health, err = solanaClient.GetHealth(s.Ctx)
	s.Require().NoError(err)

	if health == rpc.HealthOk {
		t.Log("Connection to Solana RPC is successful!")
	} else {
		t.Fatal("Connection established, but node health is not OK.")
	}
	s.TimelockProgramID = solana.MustPublicKeyFromBase58(in.SolanaChain.SolanaPrograms["timelock"])
	s.StubProgramID = solana.MustPublicKeyFromBase58(in.SolanaChain.SolanaPrograms["external_program_cpi_stub"])
	s.AccessControllerProgramID = solana.MustPublicKeyFromBase58(in.SolanaChain.SolanaPrograms["access_controller"])
	// Initialize the test private key
	s.TestPrivateKey = solana.MustPrivateKeyFromBase58(privateKey)
	_, roleMap := timelockutils.TestRoleAccounts(2)
	s.RoleMap = roleMap
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

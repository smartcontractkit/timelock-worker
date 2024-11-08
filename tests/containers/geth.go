package containers

import (
	"context"
	"fmt"
	"math/big"
	"strings"
	"testing"
	"time"

	"github.com/docker/go-connections/nat"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

// GethContainer represents a go-ethereum container.
type GethContainer struct {
	Container testcontainers.Container
	HttpURL   string
	WsURL     string
	ChainID   uint64
}

const (
	dataDir         = "/tmp"
	keystoreDir     = dataDir + "/keystore"
	timestampFormat = "2006-01-02T15-04-05.000000000Z"
	chainID         = 1337
	httpPort        = "8545"
	wsPort          = "8546"
)

// NewGethContainer creates a new geth container.
func NewGethContainer(ctx context.Context) (*GethContainer, error) {
	request := testcontainers.ContainerRequest{
		Name:         "geth",
		Hostname:     "geth",
		Image:        "ethereum/client-go:stable",
		WaitingFor:   wait.ForLog("Starting work on payload").WithStartupTimeout(15 * time.Second),
		ExposedPorts: []string{httpPort, wsPort},
		Cmd: []string{
			"--dev", "--dev.period", "1",
			"--http", "--http.addr", "0.0.0.0", "--http.vhosts", "*",
			"--ws", "--ws.addr", "0.0.0.0", "--ws.origins", "*",
			"--cache.blocklogs", "1024",
			"--networkid", fmt.Sprintf("%d", chainID),
			"--datadir", dataDir,
		},
		// uncomment to print container logs to stdout
		// LogConsumerCfg: &testcontainers.LogConsumerConfig{
		// 	Opts: []testcontainers.LogProductionOption{
		// 		testcontainers.WithLogProductionTimeout(10 * time.Second),
		// 	},
		// 	Consumers: []testcontainers.LogConsumer{&StdoutLogConsumer{Prefix: "|| "}},
		// },
	}
	gethContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: request,
		Started:          true,
	})
	if err != nil {
		return nil, fmt.Errorf("could not start geth container: %w", err)
	}

	return &GethContainer{
		Container: gethContainer,
		HttpURL:   "http://geth:8545",
		WsURL:     "ws://geth:8546",
		ChainID:   chainID,
	}, nil
}

// HTTPConnStr returns the http connection string for the geth container (from the container host).
func (g *GethContainer) HTTPConnStr(t *testing.T, ctx context.Context) string {
	t.Helper()
	connStr, err := g.Container.PortEndpoint(ctx, nat.Port(httpPort), "http")
	require.NoError(t, err)
	return connStr
}

// WSConnStr returns the websocket connection string for the geth container (from the container host).
func (g *GethContainer) WSConnStr(t *testing.T, ctx context.Context) string {
	t.Helper()
	connStr, err := g.Container.PortEndpoint(ctx, nat.Port(wsPort), "ws")
	require.NoError(t, err)
	return connStr
}

func (g *GethContainer) Teardown(ctx context.Context) error {
	return g.Container.Terminate(ctx)
}

// CreateAccount creates and funds a new account in the geth container.
func (g *GethContainer) CreateAccount(
	ctx context.Context, address, privateKey string, initialBalance uint64,
) (big.Int, error) {
	keystorePath := fmt.Sprintf("%s/UTC--%s--%s",
		keystoreDir, time.Now().UTC().Format(timestampFormat), strings.TrimLeft(address, "0x"))

	err := g.Container.CopyToContainer(ctx, []byte(privateKey), keystorePath, 0o600)
	if err != nil {
		return big.Int{}, fmt.Errorf("error copying private key to container: %w", err)
	}

	// get list of accounts and ensure the given address was added
	listAccountsCommand := []string{"geth", "attach", "--datadir", dataDir, "--exec", "eth.accounts"}
	checkAccountsOutput := func(statusCode int, output string) error {
		if strings.Contains(strings.ToLower(output), strings.ToLower(address)) {
			return nil
		}
		return fmt.Errorf("new account (%s) is not found in geth instance: (%s)", address, output)
	}
	_, _, err = execUntil(ctx, g.Container, listAccountsCommand, checkAccountsOutput,
		200*time.Millisecond, 5000*time.Millisecond)
	if err != nil {
		return big.Int{}, err
	}

	// send `initialBalance` eth from the genesis account to the new account
	sendTransactionCommand := []string{
		"geth", "attach", "--datadir", dataDir, "--exec",
		fmt.Sprintf(`eth.sendTransaction({
			from:  eth.accounts[0],
			to:    %q,
			value: web3.toWei(%d, "ether")
		})`, address, initialBalance),
	}
	statusCode, _, err := exec(ctx, g.Container, sendTransactionCommand)
	if err != nil || statusCode != 0 {
		return big.Int{}, fmt.Errorf("failed to send funds to new account (code: %d): %w", statusCode, err)
	}

	// get the balance of the new account and ensure it matches what we expect
	expectedBalance := fmt.Sprintf("%d000000000000000000", initialBalance)
	getBalanceCommand := []string{
		"geth", "attach", "--datadir", dataDir, "--exec", fmt.Sprintf(`eth.getBalance("%s")`, strings.ToLower(address)),
	}
	checkBalanceOutput := func(statusCode int, output string) error {
		if strings.HasSuffix(output, expectedBalance+"\n") {
			return nil
		}
		return fmt.Errorf("command output does not contain expected balance: %v", output)
	}
	_, _, err = execUntil(ctx, g.Container, getBalanceCommand, checkBalanceOutput,
		200*time.Millisecond, 5000*time.Millisecond)
	if err != nil {
		return big.Int{}, err
	}

	expectedBalanceBigInt, ok := new(big.Int).SetString(expectedBalance, 10)
	if !ok {
		return big.Int{}, fmt.Errorf("unable to convert balance to a number: %v", expectedBalance)
	}
	return *expectedBalanceBigInt, nil
}

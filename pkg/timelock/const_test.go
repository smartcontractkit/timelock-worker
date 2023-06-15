package timelock

import (
	"math/big"
	"os"

	"github.com/smartcontractkit/timelock-worker/pkg/logger"
)

var (
	testNodeURL          = os.Getenv("TEST_NODE_URL")
	testTimelockAddress  = os.Getenv("TEST_TIMELOCK_ADDRESS")
	testCallProxyAddress = os.Getenv("TEST_PROXY_ADDRESS")
	testPrivateKey       = os.Getenv("TEST_PRIVATE_KEY")
	testFromBlock        = big.NewInt(0)
	testPollPeriod       = 5
	testLogger           = logger.Logger("info", "human")
)

package timelock

import (
	"math/big"

	"github.com/samber/lo"

	"github.com/smartcontractkit/timelock-worker/pkg/logger"
)

var (
	testNodeURL                 = "ws://node.url"
	testTimelockAddress         = "0x0000000000000000000000000000000000000000"
	testCallProxyAddress        = "0x0000000000000000000000000000000000000000"
	testPrivateKey              = "8064bf62c044d2654705b9d0cfbd666c2649fabb76ed8f4b9d8d3eb28267e3cf"
	testFromBlock               = big.NewInt(0)
	testMaxGasLimit             = uint64(0)
	testPollPeriod              = 5
	testEventListenerPollPeriod = 1
	testEventListenerPollSize   = uint64(10)
	testDryRun                  = false
	testLogger                  = lo.Must(logger.NewLogger("info", "human")).Sugar()
)

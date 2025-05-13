package cli_test

import (
	"os"
	"testing"

	chain_selectors "github.com/smartcontractkit/chain-selectors"
	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/timelock-worker/pkg/cli"
)

func Test_NewTimelockCLI(t *testing.T) {
	const configFileName = "timelock.env"

	tests := []struct {
		name    string
		setup   func(t *testing.T)
		want    *cli.Config
		wantErr string
	}{
		{
			name: "load from file",
			setup: func(t *testing.T) {
				unsetenvs(t, "NODE_URL", "TIMELOCK_ADDRESS", "CALL_PROXY_ADDRESS", "PRIVATE_KEY", "FROM_BLOCK",
					"POLL_PERIOD", "EVENT_LISTENER_POLL_PERIOD", "DRY_RUN")

				err := os.WriteFile(configFileName, []byte(string(
					"NODE_URL=wss://goerli/test\n"+
						"TIMELOCK_ADDRESS=0x12345\n"+
						"CALL_PROXY_ADDRESS=0x67890\n"+
						"PRIVATE_KEY=9876543210\n"+
						"FROM_BLOCK=1\n"+
						"POLL_PERIOD=2\n"+
						"EVENT_LISTENER_POLL_PERIOD=3\n"+
						"DRY_RUN=true\n",
				)), os.FileMode(0644))
				require.NoError(t, err)

				t.Cleanup(func() { os.Remove(configFileName) })
			},
			want: &cli.Config{
				NodeURL:                 "wss://goerli/test",
				ChainFamily:             chain_selectors.FamilyEVM,
				TimelockAddress:         "0x12345",
				CallProxyAddress:        "0x67890",
				PrivateKey:              "9876543210",
				FromBlock:               1,
				PollPeriod:              2,
				EventListenerPollPeriod: 3,
				DryRun:                  true,
			},
		},
		{
			name: "load from environment (overrides values in file)",
			setup: func(t *testing.T) {
				err := os.WriteFile(configFileName, []byte(string(
					"NODE_URL=wss://goerli/test\n"+
						"TIMELOCK_ADDRESS=0x12345\n"+
						"CALL_PROXY_ADDRESS=0x67890\n"+
						"PRIVATE_KEY=9876543210\n"+
						"FROM_BLOCK=1\n"+
						"POLL_PERIOD=2\n"+
						"EVENT_LISTENER_POLL_PERIOD=3\n"+
						"DRY_RUN=true\n",
				)), os.FileMode(0644))
				require.NoError(t, err)

				t.Setenv("NODE_URL", "http://node.url/path")
				t.Setenv("TIMELOCK_ADDRESS", "0x111111")
				t.Setenv("CALL_PROXY_ADDRESS", "0x222222")
				t.Setenv("PRIVATE_KEY", "333333")
				t.Setenv("FROM_BLOCK", "4")
				t.Setenv("POLL_PERIOD", "5")
				t.Setenv("EVENT_LISTENER_POLL_PERIOD", "6")
				t.Setenv("DRY_RUN", "false")

				t.Cleanup(func() { os.Remove(configFileName) })
			},
			want: &cli.Config{
				NodeURL:                 "http://node.url/path",
				ChainFamily:             chain_selectors.FamilyEVM,
				TimelockAddress:         "0x111111",
				CallProxyAddress:        "0x222222",
				PrivateKey:              "333333",
				FromBlock:               4,
				PollPeriod:              5,
				EventListenerPollPeriod: 6,
			},
		},
		{
			name: "load from environment and file",
			setup: func(t *testing.T) {
				err := os.WriteFile(configFileName, []byte(string(
					"NODE_URL=wss://goerli/test\n"+
						"TIMELOCK_ADDRESS=0x12345\n"+
						"CALL_PROXY_ADDRESS=0x67890\n",
				)), os.FileMode(0644))
				require.NoError(t, err)

				unsetenvs(t, "NODE_URL", "TIMELOCK_ADDRESS", "CALL_PROXY_ADDRESS")
				t.Setenv("PRIVATE_KEY", "333333")
				t.Setenv("FROM_BLOCK", "4")
				t.Setenv("POLL_PERIOD", "5")
				t.Setenv("EVENT_LISTENER_POLL_PERIOD", "6")
				t.Setenv("DRY_RUN", "yes")

				t.Cleanup(func() { os.Remove(configFileName) })
			},
			want: &cli.Config{
				NodeURL:                 "wss://goerli/test",
				ChainFamily:             chain_selectors.FamilyEVM,
				TimelockAddress:         "0x12345",
				CallProxyAddress:        "0x67890",
				PrivateKey:              "333333",
				FromBlock:               4,
				PollPeriod:              5,
				EventListenerPollPeriod: 6,
				DryRun:                  true,
			},
		},
		{
			name: "fails due to non-numeric FROM_BLOCK",
			setup: func(t *testing.T) {
				t.Setenv("NODE_URL", "http://node.url/path")
				t.Setenv("TIMELOCK_ADDRESS", "0x111111")
				t.Setenv("CALL_PROXY_ADDRESS", "0x222222")
				t.Setenv("PRIVATE_KEY", "333333")
				t.Setenv("POLL_PERIOD", "5")
				t.Setenv("EVENT_LISTENER_POLL_PERIOD", "6")
				t.Setenv("FROM_BLOCK", "invalid")
			},
			wantErr: "unable to parse FROM_BLOCK value:",
		},
		{
			name: "fails due to non-numeric POLL_PERIOD",
			setup: func(t *testing.T) {
				t.Setenv("NODE_URL", "http://node.url/path")
				t.Setenv("TIMELOCK_ADDRESS", "0x111111")
				t.Setenv("CALL_PROXY_ADDRESS", "0x222222")
				t.Setenv("PRIVATE_KEY", "333333")
				t.Setenv("FROM_BLOCK", "4")
				t.Setenv("EVENT_LISTENER_POLL_PERIOD", "6")
				t.Setenv("POLL_PERIOD", "invalid")
			},
			wantErr: "unable to parse POLL_PERIOD value:",
		},
		{
			name: "fails due to non-numeric EVENT_LISTSENER_POLL_PERIOD",
			setup: func(t *testing.T) {
				t.Setenv("NODE_URL", "http://node.url/path")
				t.Setenv("TIMELOCK_ADDRESS", "0x111111")
				t.Setenv("CALL_PROXY_ADDRESS", "0x222222")
				t.Setenv("PRIVATE_KEY", "333333")
				t.Setenv("FROM_BLOCK", "4")
				t.Setenv("POLL_PERIOD", "5")
				t.Setenv("EVENT_LISTENER_POLL_PERIOD", "invalid")
			},
			wantErr: "unable to parse EVENT_LISTENER_POLL_PERIOD value:",
		},
		{
			name: "defaults to evm if CHAIN_FAMILY is not set",
			setup: func(t *testing.T) {
				unsetenvs(t, "CHAIN_FAMILY")
				t.Setenv("NODE_URL", "http://localhost")
				t.Setenv("TIMELOCK_ADDRESS", "0xabc")
				t.Setenv("CALL_PROXY_ADDRESS", "0xdef")
				t.Setenv("PRIVATE_KEY", "123")
				t.Setenv("FROM_BLOCK", "10")
				t.Setenv("POLL_PERIOD", "15")
				t.Setenv("EVENT_LISTENER_POLL_PERIOD", "20")
				t.Setenv("EVENT_LISTENER_POLL_SIZE", "100")
				t.Setenv("DRY_RUN", "true")
			},
			want: &cli.Config{
				NodeURL:                 "http://localhost",
				ChainFamily:             "evm",
				TimelockAddress:         "0xabc",
				CallProxyAddress:        "0xdef",
				PrivateKey:              "123",
				FromBlock:               10,
				PollPeriod:              15,
				EventListenerPollPeriod: 20,
				EventListenerPollSize:   100,
				DryRun:                  true,
			},
		},
		{
			name: "overrides ChainFamily from environment",
			setup: func(t *testing.T) {
				t.Setenv("CHAIN_FAMILY", "solana")
				t.Setenv("NODE_URL", "http://solana.node")
				t.Setenv("TIMELOCK_ADDRESS", "0xsolana")
				t.Setenv("CALL_PROXY_ADDRESS", "0xproxy")
				t.Setenv("PRIVATE_KEY", "solana_key")
				t.Setenv("FROM_BLOCK", "1")
				t.Setenv("POLL_PERIOD", "2")
				t.Setenv("EVENT_LISTENER_POLL_PERIOD", "3")
				t.Setenv("EVENT_LISTENER_POLL_SIZE", "4")
				t.Setenv("DRY_RUN", "yes")
			},
			want: &cli.Config{
				NodeURL:                 "http://solana.node",
				ChainFamily:             "solana",
				TimelockAddress:         "0xsolana",
				CallProxyAddress:        "0xproxy",
				PrivateKey:              "solana_key",
				FromBlock:               1,
				PollPeriod:              2,
				EventListenerPollPeriod: 3,
				EventListenerPollSize:   4,
				DryRun:                  true,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup(t)

			got, err := cli.NewTimelockCLI()

			if tt.wantErr == "" {
				require.Equal(t, tt.want, got)
			} else {
				require.ErrorContains(t, err, tt.wantErr)
			}
		})
	}
}

func unsetenvs(t *testing.T, keys ...string) {
	for _, key := range keys {
		prevValue, ok := os.LookupEnv(key)

		if err := os.Unsetenv(key); err != nil {
			t.Fatalf("cannot unset environment variable: %v", err)
		}

		if ok {
			t.Cleanup(func() {
				os.Setenv(key, prevValue)
			})
		} else {
			t.Cleanup(func() {
				os.Unsetenv(key)
			})
		}
	}
}

package cli_test

import (
	"bufio"
	"fmt"
	"os"
	"reflect"
	"testing"

	"github.com/smartcontractkit/timelock-worker/pkg/cli"
	"github.com/stretchr/testify/assert"
)

func TestNewConfigRaw(t *testing.T) {
	var newConfig = &cli.Config{
		NodeURL:         "foo:test",
		TimelockAddress: "0x12345",
		PrivateKey:      "0123456789",
		FromBlock:       0,
	}

	if assert.NotNil(t, newConfig) {
		assert.Equal(t, "foo:test", newConfig.NodeURL)
		assert.Equal(t, "0x12345", newConfig.TimelockAddress)
		assert.Equal(t, "0123456789", newConfig.PrivateKey)
		assert.Equal(t, int64(0), newConfig.FromBlock)
	}
}

func TestNewTimelockCLIFromEnvVar(t *testing.T) {
	t.Setenv("NODE_URL", "wss://goerli/test")
	t.Setenv("TIMELOCK_ADDRESS", "0x2135C499f82d091323E5098Ef8EEb851C17BDf4b")
	t.Setenv("PRIVATE_KEY", "1234567890")
	t.Setenv("FROM_BLOCK", "1234567890")

	var wantedConfig = cli.Config{
		NodeURL:         "wss://goerli/test",
		TimelockAddress: "0x2135C499f82d091323E5098Ef8EEb851C17BDf4b",
		PrivateKey:      "1234567890",
		FromBlock:       1234567890,
	}

	tests := []struct {
		name    string
		want    *cli.Config
		wantErr bool
	}{
		{
			"Environment Vars - should succeed",
			&wantedConfig,
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := cli.NewTimelockCLI()
			if (err != nil) != tt.wantErr {
				t.Errorf("NewTimelockCLI() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewTimelockCLI() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewTimelockCLIFromEnvVarAndFile(t *testing.T) {
	f, err := os.Create("timelock.env")
	if err != nil {
		return
	}
	defer f.Close()

	w := bufio.NewWriter(f)
	_, err = fmt.Fprintf(w, "NODE_URL=wss://goerli.infura.io/ws/v3/123...\n")
	if err != nil {
		return
	}

	_, err = fmt.Fprintf(w, "TIMELOCK_ADDRESS=0x2135C499f82d091323E5098Ef8EEb851C17BDf4b\n")
	if err != nil {
		return
	}
	_, err = fmt.Fprintf(w, "PRIVATE_KEY=MyPrivateKey\n")
	if err != nil {
		return
	}
	_, err = fmt.Fprintf(w, "FROM_BLOCK=0\n")
	if err != nil {
		return
	}

	w.Flush()

	var config = cli.Config{
		NodeURL:         "wss://goerli.infura.io/ws/v3/123...",
		TimelockAddress: "0x2135C499f82d091323E5098Ef8EEb851C17BDf4b",
		PrivateKey:      "1234567890",
		FromBlock:       1234567890,
	}

	// envvars should prevail over timelock.env
	t.Setenv("PRIVATE_KEY", "1234567890")
	t.Setenv("FROM_BLOCK", "1234567890")

	tests := []struct {
		name    string
		want    *cli.Config
		wantErr bool
	}{
		{
			"timelock.env and env vars - should succeed",
			&config,
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := cli.NewTimelockCLI()
			if (err != nil) != tt.wantErr {
				t.Errorf("NewTimelockCLI() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewTimelockCLI() = %v, want %v", got, tt.want)
			}
		})
	}

	if err := os.Remove("timelock.env"); err != nil {
		return
	}
}

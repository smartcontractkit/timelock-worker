package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStartCommand_HasExpectedFlags(t *testing.T) {
	command := startCommand()

	flags := command.Flags()

	tests := []struct {
		name string
	}{
		{"node-url"},
		{"chain-family"},
		{"timelock-address"},
		{"call-proxy-address"},
		{"private-key"},
		{"from-block"},
		{"poll-period"},
		{"event-listener-poll-period"},
		{"event-listener-poll-size"},
		{"dry-run"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			assert.NotNil(t, flags.Lookup(tt.name), "flag '%s' should be defined", tt.name)
		})
	}
}

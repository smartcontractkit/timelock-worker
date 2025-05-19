package logger_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/smartcontractkit/timelock-worker/pkg/logger"
)

func Test_NewLogger(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		level   string
		output  string
		wantErr string
	}{
		{
			name:   "success: info human",
			level:  "info",
			output: "human",
		},
		{
			name:   "success: debug json",
			level:  "info",
			output: "json",
		},
		{
			name:    "invalid log level",
			level:   "invalid",
			output:  "human",
			wantErr: "unrecognized level: \"invalid\"",
		},
		{
			name:    "invalid output",
			level:   "info",
			output:  "invalid",
			wantErr: "invalid logger output: \"invalid\"",
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			logger, err := logger.NewLogger(tt.level, tt.output)

			if tt.wantErr == "" {
				require.NoError(t, err)
				require.IsType(t, logger, new(zap.Logger))
			} else {
				require.ErrorContains(t, err, tt.wantErr)
			}
		})
	}
}

package timelock

import (
	"context"
	"crypto/ecdsa"
	"math/big"
	"reflect"
	"testing"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/rs/zerolog"
	"github.com/smartcontractkit/timelock-worker/pkg/timelock/contract"
)

func TestWorker_execute(t *testing.T) {
	type fields struct {
		ethClient       *ethclient.Client
		contract        *contract.Timelock
		executeContract *contract.Timelock
		ABI             *abi.ABI
		address         []common.Address
		fromBlock       *big.Int
		pollPeriod      int64
		logger          *zerolog.Logger
		privateKey      *ecdsa.PrivateKey
		scheduler       scheduler
	}
	type args struct {
		ctx context.Context
		op  []*contract.TimelockCallScheduled
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tw := &Worker{
				ethClient:       tt.fields.ethClient,
				contract:        tt.fields.contract,
				executeContract: tt.fields.executeContract,
				ABI:             tt.fields.ABI,
				address:         tt.fields.address,
				fromBlock:       tt.fields.fromBlock,
				pollPeriod:      tt.fields.pollPeriod,
				logger:          tt.fields.logger,
				privateKey:      tt.fields.privateKey,
				scheduler:       tt.fields.scheduler,
			}
			tw.execute(tt.args.ctx, tt.args.op)
		})
	}
}

func Test_executeCallSchedule(t *testing.T) {
	type args struct {
		ctx        context.Context
		c          *contract.TimelockTransactor
		cs         []*contract.TimelockCallScheduled
		privateKey *ecdsa.PrivateKey
	}
	tests := []struct {
		name    string
		args    args
		want    *types.Transaction
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := executeCallSchedule(tt.args.ctx, tt.args.c, tt.args.cs, tt.args.privateKey)
			if (err != nil) != tt.wantErr {
				t.Errorf("executeCallSchedule() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("executeCallSchedule() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_isOperation(t *testing.T) {
	type args struct {
		ctx context.Context
		c   *contract.Timelock
		id  [32]byte
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isOperation(tt.args.ctx, tt.args.c, tt.args.id); got != tt.want {
				t.Errorf("isOperation() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_isReady(t *testing.T) {
	type args struct {
		ctx context.Context
		c   *contract.Timelock
		id  [32]byte
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isReady(tt.args.ctx, tt.args.c, tt.args.id); got != tt.want {
				t.Errorf("isReady() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_isDone(t *testing.T) {
	type args struct {
		ctx context.Context
		c   *contract.Timelock
		id  [32]byte
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isDone(tt.args.ctx, tt.args.c, tt.args.id); got != tt.want {
				t.Errorf("isDone() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_isPending(t *testing.T) {
	type args struct {
		ctx context.Context
		c   *contract.Timelock
		id  [32]byte
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isPending(tt.args.ctx, tt.args.c, tt.args.id); got != tt.want {
				t.Errorf("isPending() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_signTx(t *testing.T) {
	type args struct {
		address common.Address
		tx      *types.Transaction
	}
	tests := []struct {
		name    string
		args    args
		want    *types.Transaction
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := signTx(tt.args.address, tt.args.tx)
			if (err != nil) != tt.wantErr {
				t.Errorf("signTx() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("signTx() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_privateKeyToAddress(t *testing.T) {
	type args struct {
		privateKey *ecdsa.PrivateKey
	}
	tests := []struct {
		name    string
		args    args
		want    common.Address
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := privateKeyToAddress(tt.args.privateKey)
			if (err != nil) != tt.wantErr {
				t.Errorf("privateKeyToAddress() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("privateKeyToAddress() = %v, want %v", got, tt.want)
			}
		})
	}
}

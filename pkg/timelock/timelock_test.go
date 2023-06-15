package timelock

import (
	"context"
	"crypto/ecdsa"
	"math/big"
	"reflect"
	"syscall"
	"testing"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/rs/zerolog"
	"github.com/smartcontractkit/timelock-worker/pkg/timelock/contract"
	"github.com/stretchr/testify/assert"
)

func TestNewTimelockWorker(t *testing.T) {
	type args struct {
		nodeURL          string
		timelockAddress  string
		callProxyAddress string
		privateKey       string
		fromBlock        *big.Int
		pollPeriod       int64
		logger           *zerolog.Logger
	}
	tests := []struct {
		name    string
		args    args
		want    *Worker
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewTimelockWorker(tt.args.nodeURL, tt.args.timelockAddress, tt.args.callProxyAddress, tt.args.privateKey, tt.args.fromBlock, tt.args.pollPeriod, tt.args.logger)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewTimelockWorker() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewTimelockWorker() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestWorker_Listen(t *testing.T) {
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
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
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
			if err := tw.Listen(tt.args.ctx); (err != nil) != tt.wantErr {
				t.Errorf("Worker.Listen() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_handleOSSignal(t *testing.T) {
	outCh := make(chan string)

	go handleOSSignal(syscall.SIGINT, outCh)
	signal := <-outCh
	assert.Equal(t, syscall.SIGINT.String(), signal, "they should be equal")

	go handleOSSignal(syscall.SIGTERM, outCh)
	signal = <-outCh
	assert.Equal(t, syscall.SIGTERM.String(), signal, "they should be equal")

	close(outCh)
}

func TestWorker_startLog(t *testing.T) {
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
	tests := []struct {
		name   string
		fields fields
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
			tw.startLog()
		})
	}
}

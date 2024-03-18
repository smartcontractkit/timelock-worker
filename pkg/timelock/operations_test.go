package timelock

import (
	"context"
	"crypto/ecdsa"
	"reflect"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/smartcontractkit/timelock-worker/pkg/timelock/contract"
)

func Test_isOperation(t *testing.T) {
	testWorker := newTestTimelockWorker(t, testNodeURL, testTimelockAddress, testCallProxyAddress, testPrivateKey, testFromBlock, int64(testPollPeriod), testLogger)

	var ctx context.Context

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
		{
			name: "isOperation: empty, should fail",
			args: args{
				ctx: ctx,
				c:   testWorker.contract,
				id:  [32]byte{},
			},
			want: false,
		},
		{
			name: "isOperation: real operation, should succeed",
			args: args{
				ctx: ctx,
				c:   testWorker.contract,
				id:  [32]byte{79, 143, 96, 15, 56, 41, 187, 178, 167, 120, 61, 145, 140, 241, 3, 220, 155, 151, 111, 184, 96, 128, 73, 10, 146, 173, 93, 211, 80, 225, 69, 183},
			},
			want: true,
		},
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
	testWorker := newTestTimelockWorker(t, testNodeURL, testTimelockAddress, testCallProxyAddress, testPrivateKey, testFromBlock, int64(testPollPeriod), testLogger)

	var ctx context.Context

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
		{
			name: "isReady: empty, should fail",
			args: args{
				ctx: ctx,
				c:   testWorker.contract,
				id:  [32]byte{},
			},
			want: false,
		},
		{
			name: "isReady: real operation not ready, should fail",
			args: args{
				ctx: ctx,
				c:   testWorker.contract,
				id:  [32]byte{1, 6, 204, 130, 40, 127, 196, 24, 117, 166, 94, 233, 151, 222, 146, 45, 238, 98, 11, 85, 157, 173, 136, 226, 220, 74, 141, 29, 9, 125, 249, 119},
			},
			want: false,
		},
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
	testWorker := newTestTimelockWorker(t, testNodeURL, testTimelockAddress, testCallProxyAddress, testPrivateKey, testFromBlock, int64(testPollPeriod), testLogger)

	var ctx context.Context

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
		{
			name: "isOperation: empty, should fail",
			args: args{
				ctx: ctx,
				c:   testWorker.contract,
				id:  [32]byte{},
			},
			want: false,
		},
		{
			name: "isOperation: real operation, should succeed",
			args: args{
				ctx: ctx,
				c:   testWorker.contract,
				id:  [32]byte{79, 143, 96, 15, 56, 41, 187, 178, 167, 120, 61, 145, 140, 241, 3, 220, 155, 151, 111, 184, 96, 128, 73, 10, 146, 173, 93, 211, 80, 225, 69, 183},
			},
			want: true,
		},
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
	testWorker := newTestTimelockWorker(t, testNodeURL, testTimelockAddress, testCallProxyAddress, testPrivateKey, testFromBlock, int64(testPollPeriod), testLogger)

	var ctx context.Context

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
		{
			name: "isOperation: empty, should fail",
			args: args{
				ctx: ctx,
				c:   testWorker.contract,
				id:  [32]byte{},
			},
			want: false,
		},
		{
			name: "isOperation: real operation, should succeed",
			args: args{
				ctx: ctx,
				c:   testWorker.contract,
				id:  [32]byte{79, 143, 96, 15, 56, 41, 187, 178, 167, 120, 61, 145, 140, 241, 3, 220, 155, 151, 111, 184, 96, 128, 73, 10, 146, 173, 93, 211, 80, 225, 69, 183},
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isPending(tt.args.ctx, tt.args.c, tt.args.id); got != tt.want {
				t.Errorf("isPending() = %v, want %v", got, tt.want)
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

// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package contract

import (
	"errors"
	"math/big"
	"strings"

	ethereum "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/event"
)

// Reference imports to suppress errors if they are not otherwise used.
var (
	_ = errors.New
	_ = big.NewInt
	_ = strings.NewReader
	_ = ethereum.NotFound
	_ = bind.Bind
	_ = common.Big1
	_ = types.BloomLookup
	_ = event.NewSubscription
	_ = abi.ConvertType
)

// RBACTimelockCall is an auto generated low-level Go binding around an user-defined struct.
type RBACTimelockCall struct {
	Target common.Address
	Value  *big.Int
	Data   []byte
}

// TimelockMetaData contains all meta data concerning the Timelock contract.
var TimelockMetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"minDelay\",\"type\":\"uint256\"},{\"internalType\":\"address\",\"name\":\"admin\",\"type\":\"address\"},{\"internalType\":\"address[]\",\"name\":\"proposers\",\"type\":\"address[]\"},{\"internalType\":\"address[]\",\"name\":\"executors\",\"type\":\"address[]\"},{\"internalType\":\"address[]\",\"name\":\"cancellers\",\"type\":\"address[]\"},{\"internalType\":\"address[]\",\"name\":\"bypassers\",\"type\":\"address[]\"}],\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"index\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"address\",\"name\":\"target\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"value\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"bytes\",\"name\":\"data\",\"type\":\"bytes\"}],\"name\":\"BypasserCallExecuted\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"id\",\"type\":\"bytes32\"},{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"index\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"address\",\"name\":\"target\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"value\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"bytes\",\"name\":\"data\",\"type\":\"bytes\"}],\"name\":\"CallExecuted\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"id\",\"type\":\"bytes32\"},{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"index\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"address\",\"name\":\"target\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"value\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"bytes\",\"name\":\"data\",\"type\":\"bytes\"},{\"indexed\":false,\"internalType\":\"bytes32\",\"name\":\"predecessor\",\"type\":\"bytes32\"},{\"indexed\":false,\"internalType\":\"bytes32\",\"name\":\"salt\",\"type\":\"bytes32\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"delay\",\"type\":\"uint256\"}],\"name\":\"CallScheduled\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"id\",\"type\":\"bytes32\"}],\"name\":\"Cancelled\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"bytes4\",\"name\":\"selector\",\"type\":\"bytes4\"}],\"name\":\"FunctionSelectorBlocked\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"bytes4\",\"name\":\"selector\",\"type\":\"bytes4\"}],\"name\":\"FunctionSelectorUnblocked\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"oldDuration\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"newDuration\",\"type\":\"uint256\"}],\"name\":\"MinDelayChange\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"role\",\"type\":\"bytes32\"},{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"previousAdminRole\",\"type\":\"bytes32\"},{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"newAdminRole\",\"type\":\"bytes32\"}],\"name\":\"RoleAdminChanged\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"role\",\"type\":\"bytes32\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"sender\",\"type\":\"address\"}],\"name\":\"RoleGranted\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"role\",\"type\":\"bytes32\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"sender\",\"type\":\"address\"}],\"name\":\"RoleRevoked\",\"type\":\"event\"},{\"inputs\":[],\"name\":\"ADMIN_ROLE\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"BYPASSER_ROLE\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"CANCELLER_ROLE\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"DEFAULT_ADMIN_ROLE\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"EXECUTOR_ROLE\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"PROPOSER_ROLE\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes4\",\"name\":\"selector\",\"type\":\"bytes4\"}],\"name\":\"blockFunctionSelector\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"components\":[{\"internalType\":\"address\",\"name\":\"target\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"value\",\"type\":\"uint256\"},{\"internalType\":\"bytes\",\"name\":\"data\",\"type\":\"bytes\"}],\"internalType\":\"structRBACTimelock.Call[]\",\"name\":\"calls\",\"type\":\"tuple[]\"}],\"name\":\"bypasserExecuteBatch\",\"outputs\":[],\"stateMutability\":\"payable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"id\",\"type\":\"bytes32\"}],\"name\":\"cancel\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"components\":[{\"internalType\":\"address\",\"name\":\"target\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"value\",\"type\":\"uint256\"},{\"internalType\":\"bytes\",\"name\":\"data\",\"type\":\"bytes\"}],\"internalType\":\"structRBACTimelock.Call[]\",\"name\":\"calls\",\"type\":\"tuple[]\"},{\"internalType\":\"bytes32\",\"name\":\"predecessor\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"salt\",\"type\":\"bytes32\"}],\"name\":\"executeBatch\",\"outputs\":[],\"stateMutability\":\"payable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"index\",\"type\":\"uint256\"}],\"name\":\"getBlockedFunctionSelectorAt\",\"outputs\":[{\"internalType\":\"bytes4\",\"name\":\"\",\"type\":\"bytes4\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getBlockedFunctionSelectorCount\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getMinDelay\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"duration\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"role\",\"type\":\"bytes32\"}],\"name\":\"getRoleAdmin\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"role\",\"type\":\"bytes32\"},{\"internalType\":\"uint256\",\"name\":\"index\",\"type\":\"uint256\"}],\"name\":\"getRoleMember\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"role\",\"type\":\"bytes32\"}],\"name\":\"getRoleMemberCount\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"id\",\"type\":\"bytes32\"}],\"name\":\"getTimestamp\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"timestamp\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"role\",\"type\":\"bytes32\"},{\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"}],\"name\":\"grantRole\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"role\",\"type\":\"bytes32\"},{\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"}],\"name\":\"hasRole\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"components\":[{\"internalType\":\"address\",\"name\":\"target\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"value\",\"type\":\"uint256\"},{\"internalType\":\"bytes\",\"name\":\"data\",\"type\":\"bytes\"}],\"internalType\":\"structRBACTimelock.Call[]\",\"name\":\"calls\",\"type\":\"tuple[]\"},{\"internalType\":\"bytes32\",\"name\":\"predecessor\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"salt\",\"type\":\"bytes32\"}],\"name\":\"hashOperationBatch\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"hash\",\"type\":\"bytes32\"}],\"stateMutability\":\"pure\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"id\",\"type\":\"bytes32\"}],\"name\":\"isOperation\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"registered\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"id\",\"type\":\"bytes32\"}],\"name\":\"isOperationDone\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"done\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"id\",\"type\":\"bytes32\"}],\"name\":\"isOperationPending\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"pending\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"id\",\"type\":\"bytes32\"}],\"name\":\"isOperationReady\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"ready\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"},{\"internalType\":\"uint256[]\",\"name\":\"\",\"type\":\"uint256[]\"},{\"internalType\":\"uint256[]\",\"name\":\"\",\"type\":\"uint256[]\"},{\"internalType\":\"bytes\",\"name\":\"\",\"type\":\"bytes\"}],\"name\":\"onERC1155BatchReceived\",\"outputs\":[{\"internalType\":\"bytes4\",\"name\":\"\",\"type\":\"bytes4\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"},{\"internalType\":\"bytes\",\"name\":\"\",\"type\":\"bytes\"}],\"name\":\"onERC1155Received\",\"outputs\":[{\"internalType\":\"bytes4\",\"name\":\"\",\"type\":\"bytes4\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"},{\"internalType\":\"bytes\",\"name\":\"\",\"type\":\"bytes\"}],\"name\":\"onERC721Received\",\"outputs\":[{\"internalType\":\"bytes4\",\"name\":\"\",\"type\":\"bytes4\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"role\",\"type\":\"bytes32\"},{\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"}],\"name\":\"renounceRole\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"role\",\"type\":\"bytes32\"},{\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"}],\"name\":\"revokeRole\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"components\":[{\"internalType\":\"address\",\"name\":\"target\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"value\",\"type\":\"uint256\"},{\"internalType\":\"bytes\",\"name\":\"data\",\"type\":\"bytes\"}],\"internalType\":\"structRBACTimelock.Call[]\",\"name\":\"calls\",\"type\":\"tuple[]\"},{\"internalType\":\"bytes32\",\"name\":\"predecessor\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"salt\",\"type\":\"bytes32\"},{\"internalType\":\"uint256\",\"name\":\"delay\",\"type\":\"uint256\"}],\"name\":\"scheduleBatch\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes4\",\"name\":\"interfaceId\",\"type\":\"bytes4\"}],\"name\":\"supportsInterface\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes4\",\"name\":\"selector\",\"type\":\"bytes4\"}],\"name\":\"unblockFunctionSelector\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"newDelay\",\"type\":\"uint256\"}],\"name\":\"updateDelay\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"stateMutability\":\"payable\",\"type\":\"receive\"}]",
}

// TimelockABI is the input ABI used to generate the binding from.
// Deprecated: Use TimelockMetaData.ABI instead.
var TimelockABI = TimelockMetaData.ABI

// Timelock is an auto generated Go binding around an Ethereum contract.
type Timelock struct {
	TimelockCaller     // Read-only binding to the contract
	TimelockTransactor // Write-only binding to the contract
	TimelockFilterer   // Log filterer for contract events
}

// TimelockCaller is an auto generated read-only Go binding around an Ethereum contract.
type TimelockCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// TimelockTransactor is an auto generated write-only Go binding around an Ethereum contract.
type TimelockTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// TimelockFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type TimelockFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// TimelockSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type TimelockSession struct {
	Contract     *Timelock         // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// TimelockCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type TimelockCallerSession struct {
	Contract *TimelockCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts   // Call options to use throughout this session
}

// TimelockTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type TimelockTransactorSession struct {
	Contract     *TimelockTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts   // Transaction auth options to use throughout this session
}

// TimelockRaw is an auto generated low-level Go binding around an Ethereum contract.
type TimelockRaw struct {
	Contract *Timelock // Generic contract binding to access the raw methods on
}

// TimelockCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type TimelockCallerRaw struct {
	Contract *TimelockCaller // Generic read-only contract binding to access the raw methods on
}

// TimelockTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type TimelockTransactorRaw struct {
	Contract *TimelockTransactor // Generic write-only contract binding to access the raw methods on
}

// NewTimelock creates a new instance of Timelock, bound to a specific deployed contract.
func NewTimelock(address common.Address, backend bind.ContractBackend) (*Timelock, error) {
	contract, err := bindTimelock(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &Timelock{TimelockCaller: TimelockCaller{contract: contract}, TimelockTransactor: TimelockTransactor{contract: contract}, TimelockFilterer: TimelockFilterer{contract: contract}}, nil
}

// NewTimelockCaller creates a new read-only instance of Timelock, bound to a specific deployed contract.
func NewTimelockCaller(address common.Address, caller bind.ContractCaller) (*TimelockCaller, error) {
	contract, err := bindTimelock(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &TimelockCaller{contract: contract}, nil
}

// NewTimelockTransactor creates a new write-only instance of Timelock, bound to a specific deployed contract.
func NewTimelockTransactor(address common.Address, transactor bind.ContractTransactor) (*TimelockTransactor, error) {
	contract, err := bindTimelock(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &TimelockTransactor{contract: contract}, nil
}

// NewTimelockFilterer creates a new log filterer instance of Timelock, bound to a specific deployed contract.
func NewTimelockFilterer(address common.Address, filterer bind.ContractFilterer) (*TimelockFilterer, error) {
	contract, err := bindTimelock(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &TimelockFilterer{contract: contract}, nil
}

// bindTimelock binds a generic wrapper to an already deployed contract.
func bindTimelock(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := TimelockMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Timelock *TimelockRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Timelock.Contract.TimelockCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Timelock *TimelockRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Timelock.Contract.TimelockTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Timelock *TimelockRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Timelock.Contract.TimelockTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Timelock *TimelockCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Timelock.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Timelock *TimelockTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Timelock.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Timelock *TimelockTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Timelock.Contract.contract.Transact(opts, method, params...)
}

// ADMINROLE is a free data retrieval call binding the contract method 0x75b238fc.
//
// Solidity: function ADMIN_ROLE() view returns(bytes32)
func (_Timelock *TimelockCaller) ADMINROLE(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _Timelock.contract.Call(opts, &out, "ADMIN_ROLE")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// ADMINROLE is a free data retrieval call binding the contract method 0x75b238fc.
//
// Solidity: function ADMIN_ROLE() view returns(bytes32)
func (_Timelock *TimelockSession) ADMINROLE() ([32]byte, error) {
	return _Timelock.Contract.ADMINROLE(&_Timelock.CallOpts)
}

// ADMINROLE is a free data retrieval call binding the contract method 0x75b238fc.
//
// Solidity: function ADMIN_ROLE() view returns(bytes32)
func (_Timelock *TimelockCallerSession) ADMINROLE() ([32]byte, error) {
	return _Timelock.Contract.ADMINROLE(&_Timelock.CallOpts)
}

// BYPASSERROLE is a free data retrieval call binding the contract method 0x191cb7b3.
//
// Solidity: function BYPASSER_ROLE() view returns(bytes32)
func (_Timelock *TimelockCaller) BYPASSERROLE(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _Timelock.contract.Call(opts, &out, "BYPASSER_ROLE")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// BYPASSERROLE is a free data retrieval call binding the contract method 0x191cb7b3.
//
// Solidity: function BYPASSER_ROLE() view returns(bytes32)
func (_Timelock *TimelockSession) BYPASSERROLE() ([32]byte, error) {
	return _Timelock.Contract.BYPASSERROLE(&_Timelock.CallOpts)
}

// BYPASSERROLE is a free data retrieval call binding the contract method 0x191cb7b3.
//
// Solidity: function BYPASSER_ROLE() view returns(bytes32)
func (_Timelock *TimelockCallerSession) BYPASSERROLE() ([32]byte, error) {
	return _Timelock.Contract.BYPASSERROLE(&_Timelock.CallOpts)
}

// CANCELLERROLE is a free data retrieval call binding the contract method 0xb08e51c0.
//
// Solidity: function CANCELLER_ROLE() view returns(bytes32)
func (_Timelock *TimelockCaller) CANCELLERROLE(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _Timelock.contract.Call(opts, &out, "CANCELLER_ROLE")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// CANCELLERROLE is a free data retrieval call binding the contract method 0xb08e51c0.
//
// Solidity: function CANCELLER_ROLE() view returns(bytes32)
func (_Timelock *TimelockSession) CANCELLERROLE() ([32]byte, error) {
	return _Timelock.Contract.CANCELLERROLE(&_Timelock.CallOpts)
}

// CANCELLERROLE is a free data retrieval call binding the contract method 0xb08e51c0.
//
// Solidity: function CANCELLER_ROLE() view returns(bytes32)
func (_Timelock *TimelockCallerSession) CANCELLERROLE() ([32]byte, error) {
	return _Timelock.Contract.CANCELLERROLE(&_Timelock.CallOpts)
}

// DEFAULTADMINROLE is a free data retrieval call binding the contract method 0xa217fddf.
//
// Solidity: function DEFAULT_ADMIN_ROLE() view returns(bytes32)
func (_Timelock *TimelockCaller) DEFAULTADMINROLE(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _Timelock.contract.Call(opts, &out, "DEFAULT_ADMIN_ROLE")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// DEFAULTADMINROLE is a free data retrieval call binding the contract method 0xa217fddf.
//
// Solidity: function DEFAULT_ADMIN_ROLE() view returns(bytes32)
func (_Timelock *TimelockSession) DEFAULTADMINROLE() ([32]byte, error) {
	return _Timelock.Contract.DEFAULTADMINROLE(&_Timelock.CallOpts)
}

// DEFAULTADMINROLE is a free data retrieval call binding the contract method 0xa217fddf.
//
// Solidity: function DEFAULT_ADMIN_ROLE() view returns(bytes32)
func (_Timelock *TimelockCallerSession) DEFAULTADMINROLE() ([32]byte, error) {
	return _Timelock.Contract.DEFAULTADMINROLE(&_Timelock.CallOpts)
}

// EXECUTORROLE is a free data retrieval call binding the contract method 0x07bd0265.
//
// Solidity: function EXECUTOR_ROLE() view returns(bytes32)
func (_Timelock *TimelockCaller) EXECUTORROLE(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _Timelock.contract.Call(opts, &out, "EXECUTOR_ROLE")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// EXECUTORROLE is a free data retrieval call binding the contract method 0x07bd0265.
//
// Solidity: function EXECUTOR_ROLE() view returns(bytes32)
func (_Timelock *TimelockSession) EXECUTORROLE() ([32]byte, error) {
	return _Timelock.Contract.EXECUTORROLE(&_Timelock.CallOpts)
}

// EXECUTORROLE is a free data retrieval call binding the contract method 0x07bd0265.
//
// Solidity: function EXECUTOR_ROLE() view returns(bytes32)
func (_Timelock *TimelockCallerSession) EXECUTORROLE() ([32]byte, error) {
	return _Timelock.Contract.EXECUTORROLE(&_Timelock.CallOpts)
}

// PROPOSERROLE is a free data retrieval call binding the contract method 0x8f61f4f5.
//
// Solidity: function PROPOSER_ROLE() view returns(bytes32)
func (_Timelock *TimelockCaller) PROPOSERROLE(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _Timelock.contract.Call(opts, &out, "PROPOSER_ROLE")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// PROPOSERROLE is a free data retrieval call binding the contract method 0x8f61f4f5.
//
// Solidity: function PROPOSER_ROLE() view returns(bytes32)
func (_Timelock *TimelockSession) PROPOSERROLE() ([32]byte, error) {
	return _Timelock.Contract.PROPOSERROLE(&_Timelock.CallOpts)
}

// PROPOSERROLE is a free data retrieval call binding the contract method 0x8f61f4f5.
//
// Solidity: function PROPOSER_ROLE() view returns(bytes32)
func (_Timelock *TimelockCallerSession) PROPOSERROLE() ([32]byte, error) {
	return _Timelock.Contract.PROPOSERROLE(&_Timelock.CallOpts)
}

// GetBlockedFunctionSelectorAt is a free data retrieval call binding the contract method 0x03e56155.
//
// Solidity: function getBlockedFunctionSelectorAt(uint256 index) view returns(bytes4)
func (_Timelock *TimelockCaller) GetBlockedFunctionSelectorAt(opts *bind.CallOpts, index *big.Int) ([4]byte, error) {
	var out []interface{}
	err := _Timelock.contract.Call(opts, &out, "getBlockedFunctionSelectorAt", index)

	if err != nil {
		return *new([4]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([4]byte)).(*[4]byte)

	return out0, err

}

// GetBlockedFunctionSelectorAt is a free data retrieval call binding the contract method 0x03e56155.
//
// Solidity: function getBlockedFunctionSelectorAt(uint256 index) view returns(bytes4)
func (_Timelock *TimelockSession) GetBlockedFunctionSelectorAt(index *big.Int) ([4]byte, error) {
	return _Timelock.Contract.GetBlockedFunctionSelectorAt(&_Timelock.CallOpts, index)
}

// GetBlockedFunctionSelectorAt is a free data retrieval call binding the contract method 0x03e56155.
//
// Solidity: function getBlockedFunctionSelectorAt(uint256 index) view returns(bytes4)
func (_Timelock *TimelockCallerSession) GetBlockedFunctionSelectorAt(index *big.Int) ([4]byte, error) {
	return _Timelock.Contract.GetBlockedFunctionSelectorAt(&_Timelock.CallOpts, index)
}

// GetBlockedFunctionSelectorCount is a free data retrieval call binding the contract method 0x26bb2ec5.
//
// Solidity: function getBlockedFunctionSelectorCount() view returns(uint256)
func (_Timelock *TimelockCaller) GetBlockedFunctionSelectorCount(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _Timelock.contract.Call(opts, &out, "getBlockedFunctionSelectorCount")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetBlockedFunctionSelectorCount is a free data retrieval call binding the contract method 0x26bb2ec5.
//
// Solidity: function getBlockedFunctionSelectorCount() view returns(uint256)
func (_Timelock *TimelockSession) GetBlockedFunctionSelectorCount() (*big.Int, error) {
	return _Timelock.Contract.GetBlockedFunctionSelectorCount(&_Timelock.CallOpts)
}

// GetBlockedFunctionSelectorCount is a free data retrieval call binding the contract method 0x26bb2ec5.
//
// Solidity: function getBlockedFunctionSelectorCount() view returns(uint256)
func (_Timelock *TimelockCallerSession) GetBlockedFunctionSelectorCount() (*big.Int, error) {
	return _Timelock.Contract.GetBlockedFunctionSelectorCount(&_Timelock.CallOpts)
}

// GetMinDelay is a free data retrieval call binding the contract method 0xf27a0c92.
//
// Solidity: function getMinDelay() view returns(uint256 duration)
func (_Timelock *TimelockCaller) GetMinDelay(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _Timelock.contract.Call(opts, &out, "getMinDelay")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetMinDelay is a free data retrieval call binding the contract method 0xf27a0c92.
//
// Solidity: function getMinDelay() view returns(uint256 duration)
func (_Timelock *TimelockSession) GetMinDelay() (*big.Int, error) {
	return _Timelock.Contract.GetMinDelay(&_Timelock.CallOpts)
}

// GetMinDelay is a free data retrieval call binding the contract method 0xf27a0c92.
//
// Solidity: function getMinDelay() view returns(uint256 duration)
func (_Timelock *TimelockCallerSession) GetMinDelay() (*big.Int, error) {
	return _Timelock.Contract.GetMinDelay(&_Timelock.CallOpts)
}

// GetRoleAdmin is a free data retrieval call binding the contract method 0x248a9ca3.
//
// Solidity: function getRoleAdmin(bytes32 role) view returns(bytes32)
func (_Timelock *TimelockCaller) GetRoleAdmin(opts *bind.CallOpts, role [32]byte) ([32]byte, error) {
	var out []interface{}
	err := _Timelock.contract.Call(opts, &out, "getRoleAdmin", role)

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// GetRoleAdmin is a free data retrieval call binding the contract method 0x248a9ca3.
//
// Solidity: function getRoleAdmin(bytes32 role) view returns(bytes32)
func (_Timelock *TimelockSession) GetRoleAdmin(role [32]byte) ([32]byte, error) {
	return _Timelock.Contract.GetRoleAdmin(&_Timelock.CallOpts, role)
}

// GetRoleAdmin is a free data retrieval call binding the contract method 0x248a9ca3.
//
// Solidity: function getRoleAdmin(bytes32 role) view returns(bytes32)
func (_Timelock *TimelockCallerSession) GetRoleAdmin(role [32]byte) ([32]byte, error) {
	return _Timelock.Contract.GetRoleAdmin(&_Timelock.CallOpts, role)
}

// GetRoleMember is a free data retrieval call binding the contract method 0x9010d07c.
//
// Solidity: function getRoleMember(bytes32 role, uint256 index) view returns(address)
func (_Timelock *TimelockCaller) GetRoleMember(opts *bind.CallOpts, role [32]byte, index *big.Int) (common.Address, error) {
	var out []interface{}
	err := _Timelock.contract.Call(opts, &out, "getRoleMember", role, index)

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// GetRoleMember is a free data retrieval call binding the contract method 0x9010d07c.
//
// Solidity: function getRoleMember(bytes32 role, uint256 index) view returns(address)
func (_Timelock *TimelockSession) GetRoleMember(role [32]byte, index *big.Int) (common.Address, error) {
	return _Timelock.Contract.GetRoleMember(&_Timelock.CallOpts, role, index)
}

// GetRoleMember is a free data retrieval call binding the contract method 0x9010d07c.
//
// Solidity: function getRoleMember(bytes32 role, uint256 index) view returns(address)
func (_Timelock *TimelockCallerSession) GetRoleMember(role [32]byte, index *big.Int) (common.Address, error) {
	return _Timelock.Contract.GetRoleMember(&_Timelock.CallOpts, role, index)
}

// GetRoleMemberCount is a free data retrieval call binding the contract method 0xca15c873.
//
// Solidity: function getRoleMemberCount(bytes32 role) view returns(uint256)
func (_Timelock *TimelockCaller) GetRoleMemberCount(opts *bind.CallOpts, role [32]byte) (*big.Int, error) {
	var out []interface{}
	err := _Timelock.contract.Call(opts, &out, "getRoleMemberCount", role)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetRoleMemberCount is a free data retrieval call binding the contract method 0xca15c873.
//
// Solidity: function getRoleMemberCount(bytes32 role) view returns(uint256)
func (_Timelock *TimelockSession) GetRoleMemberCount(role [32]byte) (*big.Int, error) {
	return _Timelock.Contract.GetRoleMemberCount(&_Timelock.CallOpts, role)
}

// GetRoleMemberCount is a free data retrieval call binding the contract method 0xca15c873.
//
// Solidity: function getRoleMemberCount(bytes32 role) view returns(uint256)
func (_Timelock *TimelockCallerSession) GetRoleMemberCount(role [32]byte) (*big.Int, error) {
	return _Timelock.Contract.GetRoleMemberCount(&_Timelock.CallOpts, role)
}

// GetTimestamp is a free data retrieval call binding the contract method 0xd45c4435.
//
// Solidity: function getTimestamp(bytes32 id) view returns(uint256 timestamp)
func (_Timelock *TimelockCaller) GetTimestamp(opts *bind.CallOpts, id [32]byte) (*big.Int, error) {
	var out []interface{}
	err := _Timelock.contract.Call(opts, &out, "getTimestamp", id)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetTimestamp is a free data retrieval call binding the contract method 0xd45c4435.
//
// Solidity: function getTimestamp(bytes32 id) view returns(uint256 timestamp)
func (_Timelock *TimelockSession) GetTimestamp(id [32]byte) (*big.Int, error) {
	return _Timelock.Contract.GetTimestamp(&_Timelock.CallOpts, id)
}

// GetTimestamp is a free data retrieval call binding the contract method 0xd45c4435.
//
// Solidity: function getTimestamp(bytes32 id) view returns(uint256 timestamp)
func (_Timelock *TimelockCallerSession) GetTimestamp(id [32]byte) (*big.Int, error) {
	return _Timelock.Contract.GetTimestamp(&_Timelock.CallOpts, id)
}

// HasRole is a free data retrieval call binding the contract method 0x91d14854.
//
// Solidity: function hasRole(bytes32 role, address account) view returns(bool)
func (_Timelock *TimelockCaller) HasRole(opts *bind.CallOpts, role [32]byte, account common.Address) (bool, error) {
	var out []interface{}
	err := _Timelock.contract.Call(opts, &out, "hasRole", role, account)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// HasRole is a free data retrieval call binding the contract method 0x91d14854.
//
// Solidity: function hasRole(bytes32 role, address account) view returns(bool)
func (_Timelock *TimelockSession) HasRole(role [32]byte, account common.Address) (bool, error) {
	return _Timelock.Contract.HasRole(&_Timelock.CallOpts, role, account)
}

// HasRole is a free data retrieval call binding the contract method 0x91d14854.
//
// Solidity: function hasRole(bytes32 role, address account) view returns(bool)
func (_Timelock *TimelockCallerSession) HasRole(role [32]byte, account common.Address) (bool, error) {
	return _Timelock.Contract.HasRole(&_Timelock.CallOpts, role, account)
}

// HashOperationBatch is a free data retrieval call binding the contract method 0x515a3db3.
//
// Solidity: function hashOperationBatch((address,uint256,bytes)[] calls, bytes32 predecessor, bytes32 salt) pure returns(bytes32 hash)
func (_Timelock *TimelockCaller) HashOperationBatch(opts *bind.CallOpts, calls []RBACTimelockCall, predecessor [32]byte, salt [32]byte) ([32]byte, error) {
	var out []interface{}
	err := _Timelock.contract.Call(opts, &out, "hashOperationBatch", calls, predecessor, salt)

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// HashOperationBatch is a free data retrieval call binding the contract method 0x515a3db3.
//
// Solidity: function hashOperationBatch((address,uint256,bytes)[] calls, bytes32 predecessor, bytes32 salt) pure returns(bytes32 hash)
func (_Timelock *TimelockSession) HashOperationBatch(calls []RBACTimelockCall, predecessor [32]byte, salt [32]byte) ([32]byte, error) {
	return _Timelock.Contract.HashOperationBatch(&_Timelock.CallOpts, calls, predecessor, salt)
}

// HashOperationBatch is a free data retrieval call binding the contract method 0x515a3db3.
//
// Solidity: function hashOperationBatch((address,uint256,bytes)[] calls, bytes32 predecessor, bytes32 salt) pure returns(bytes32 hash)
func (_Timelock *TimelockCallerSession) HashOperationBatch(calls []RBACTimelockCall, predecessor [32]byte, salt [32]byte) ([32]byte, error) {
	return _Timelock.Contract.HashOperationBatch(&_Timelock.CallOpts, calls, predecessor, salt)
}

// IsOperation is a free data retrieval call binding the contract method 0x31d50750.
//
// Solidity: function isOperation(bytes32 id) view returns(bool registered)
func (_Timelock *TimelockCaller) IsOperation(opts *bind.CallOpts, id [32]byte) (bool, error) {
	var out []interface{}
	err := _Timelock.contract.Call(opts, &out, "isOperation", id)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// IsOperation is a free data retrieval call binding the contract method 0x31d50750.
//
// Solidity: function isOperation(bytes32 id) view returns(bool registered)
func (_Timelock *TimelockSession) IsOperation(id [32]byte) (bool, error) {
	return _Timelock.Contract.IsOperation(&_Timelock.CallOpts, id)
}

// IsOperation is a free data retrieval call binding the contract method 0x31d50750.
//
// Solidity: function isOperation(bytes32 id) view returns(bool registered)
func (_Timelock *TimelockCallerSession) IsOperation(id [32]byte) (bool, error) {
	return _Timelock.Contract.IsOperation(&_Timelock.CallOpts, id)
}

// IsOperationDone is a free data retrieval call binding the contract method 0x2ab0f529.
//
// Solidity: function isOperationDone(bytes32 id) view returns(bool done)
func (_Timelock *TimelockCaller) IsOperationDone(opts *bind.CallOpts, id [32]byte) (bool, error) {
	var out []interface{}
	err := _Timelock.contract.Call(opts, &out, "isOperationDone", id)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// IsOperationDone is a free data retrieval call binding the contract method 0x2ab0f529.
//
// Solidity: function isOperationDone(bytes32 id) view returns(bool done)
func (_Timelock *TimelockSession) IsOperationDone(id [32]byte) (bool, error) {
	return _Timelock.Contract.IsOperationDone(&_Timelock.CallOpts, id)
}

// IsOperationDone is a free data retrieval call binding the contract method 0x2ab0f529.
//
// Solidity: function isOperationDone(bytes32 id) view returns(bool done)
func (_Timelock *TimelockCallerSession) IsOperationDone(id [32]byte) (bool, error) {
	return _Timelock.Contract.IsOperationDone(&_Timelock.CallOpts, id)
}

// IsOperationPending is a free data retrieval call binding the contract method 0x584b153e.
//
// Solidity: function isOperationPending(bytes32 id) view returns(bool pending)
func (_Timelock *TimelockCaller) IsOperationPending(opts *bind.CallOpts, id [32]byte) (bool, error) {
	var out []interface{}
	err := _Timelock.contract.Call(opts, &out, "isOperationPending", id)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// IsOperationPending is a free data retrieval call binding the contract method 0x584b153e.
//
// Solidity: function isOperationPending(bytes32 id) view returns(bool pending)
func (_Timelock *TimelockSession) IsOperationPending(id [32]byte) (bool, error) {
	return _Timelock.Contract.IsOperationPending(&_Timelock.CallOpts, id)
}

// IsOperationPending is a free data retrieval call binding the contract method 0x584b153e.
//
// Solidity: function isOperationPending(bytes32 id) view returns(bool pending)
func (_Timelock *TimelockCallerSession) IsOperationPending(id [32]byte) (bool, error) {
	return _Timelock.Contract.IsOperationPending(&_Timelock.CallOpts, id)
}

// IsOperationReady is a free data retrieval call binding the contract method 0x13bc9f20.
//
// Solidity: function isOperationReady(bytes32 id) view returns(bool ready)
func (_Timelock *TimelockCaller) IsOperationReady(opts *bind.CallOpts, id [32]byte) (bool, error) {
	var out []interface{}
	err := _Timelock.contract.Call(opts, &out, "isOperationReady", id)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// IsOperationReady is a free data retrieval call binding the contract method 0x13bc9f20.
//
// Solidity: function isOperationReady(bytes32 id) view returns(bool ready)
func (_Timelock *TimelockSession) IsOperationReady(id [32]byte) (bool, error) {
	return _Timelock.Contract.IsOperationReady(&_Timelock.CallOpts, id)
}

// IsOperationReady is a free data retrieval call binding the contract method 0x13bc9f20.
//
// Solidity: function isOperationReady(bytes32 id) view returns(bool ready)
func (_Timelock *TimelockCallerSession) IsOperationReady(id [32]byte) (bool, error) {
	return _Timelock.Contract.IsOperationReady(&_Timelock.CallOpts, id)
}

// SupportsInterface is a free data retrieval call binding the contract method 0x01ffc9a7.
//
// Solidity: function supportsInterface(bytes4 interfaceId) view returns(bool)
func (_Timelock *TimelockCaller) SupportsInterface(opts *bind.CallOpts, interfaceId [4]byte) (bool, error) {
	var out []interface{}
	err := _Timelock.contract.Call(opts, &out, "supportsInterface", interfaceId)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// SupportsInterface is a free data retrieval call binding the contract method 0x01ffc9a7.
//
// Solidity: function supportsInterface(bytes4 interfaceId) view returns(bool)
func (_Timelock *TimelockSession) SupportsInterface(interfaceId [4]byte) (bool, error) {
	return _Timelock.Contract.SupportsInterface(&_Timelock.CallOpts, interfaceId)
}

// SupportsInterface is a free data retrieval call binding the contract method 0x01ffc9a7.
//
// Solidity: function supportsInterface(bytes4 interfaceId) view returns(bool)
func (_Timelock *TimelockCallerSession) SupportsInterface(interfaceId [4]byte) (bool, error) {
	return _Timelock.Contract.SupportsInterface(&_Timelock.CallOpts, interfaceId)
}

// BlockFunctionSelector is a paid mutator transaction binding the contract method 0x9f5a23f7.
//
// Solidity: function blockFunctionSelector(bytes4 selector) returns()
func (_Timelock *TimelockTransactor) BlockFunctionSelector(opts *bind.TransactOpts, selector [4]byte) (*types.Transaction, error) {
	return _Timelock.contract.Transact(opts, "blockFunctionSelector", selector)
}

// BlockFunctionSelector is a paid mutator transaction binding the contract method 0x9f5a23f7.
//
// Solidity: function blockFunctionSelector(bytes4 selector) returns()
func (_Timelock *TimelockSession) BlockFunctionSelector(selector [4]byte) (*types.Transaction, error) {
	return _Timelock.Contract.BlockFunctionSelector(&_Timelock.TransactOpts, selector)
}

// BlockFunctionSelector is a paid mutator transaction binding the contract method 0x9f5a23f7.
//
// Solidity: function blockFunctionSelector(bytes4 selector) returns()
func (_Timelock *TimelockTransactorSession) BlockFunctionSelector(selector [4]byte) (*types.Transaction, error) {
	return _Timelock.Contract.BlockFunctionSelector(&_Timelock.TransactOpts, selector)
}

// BypasserExecuteBatch is a paid mutator transaction binding the contract method 0x0db866b1.
//
// Solidity: function bypasserExecuteBatch((address,uint256,bytes)[] calls) payable returns()
func (_Timelock *TimelockTransactor) BypasserExecuteBatch(opts *bind.TransactOpts, calls []RBACTimelockCall) (*types.Transaction, error) {
	return _Timelock.contract.Transact(opts, "bypasserExecuteBatch", calls)
}

// BypasserExecuteBatch is a paid mutator transaction binding the contract method 0x0db866b1.
//
// Solidity: function bypasserExecuteBatch((address,uint256,bytes)[] calls) payable returns()
func (_Timelock *TimelockSession) BypasserExecuteBatch(calls []RBACTimelockCall) (*types.Transaction, error) {
	return _Timelock.Contract.BypasserExecuteBatch(&_Timelock.TransactOpts, calls)
}

// BypasserExecuteBatch is a paid mutator transaction binding the contract method 0x0db866b1.
//
// Solidity: function bypasserExecuteBatch((address,uint256,bytes)[] calls) payable returns()
func (_Timelock *TimelockTransactorSession) BypasserExecuteBatch(calls []RBACTimelockCall) (*types.Transaction, error) {
	return _Timelock.Contract.BypasserExecuteBatch(&_Timelock.TransactOpts, calls)
}

// Cancel is a paid mutator transaction binding the contract method 0xc4d252f5.
//
// Solidity: function cancel(bytes32 id) returns()
func (_Timelock *TimelockTransactor) Cancel(opts *bind.TransactOpts, id [32]byte) (*types.Transaction, error) {
	return _Timelock.contract.Transact(opts, "cancel", id)
}

// Cancel is a paid mutator transaction binding the contract method 0xc4d252f5.
//
// Solidity: function cancel(bytes32 id) returns()
func (_Timelock *TimelockSession) Cancel(id [32]byte) (*types.Transaction, error) {
	return _Timelock.Contract.Cancel(&_Timelock.TransactOpts, id)
}

// Cancel is a paid mutator transaction binding the contract method 0xc4d252f5.
//
// Solidity: function cancel(bytes32 id) returns()
func (_Timelock *TimelockTransactorSession) Cancel(id [32]byte) (*types.Transaction, error) {
	return _Timelock.Contract.Cancel(&_Timelock.TransactOpts, id)
}

// ExecuteBatch is a paid mutator transaction binding the contract method 0x6ceef480.
//
// Solidity: function executeBatch((address,uint256,bytes)[] calls, bytes32 predecessor, bytes32 salt) payable returns()
func (_Timelock *TimelockTransactor) ExecuteBatch(opts *bind.TransactOpts, calls []RBACTimelockCall, predecessor [32]byte, salt [32]byte) (*types.Transaction, error) {
	return _Timelock.contract.Transact(opts, "executeBatch", calls, predecessor, salt)
}

// ExecuteBatch is a paid mutator transaction binding the contract method 0x6ceef480.
//
// Solidity: function executeBatch((address,uint256,bytes)[] calls, bytes32 predecessor, bytes32 salt) payable returns()
func (_Timelock *TimelockSession) ExecuteBatch(calls []RBACTimelockCall, predecessor [32]byte, salt [32]byte) (*types.Transaction, error) {
	return _Timelock.Contract.ExecuteBatch(&_Timelock.TransactOpts, calls, predecessor, salt)
}

// ExecuteBatch is a paid mutator transaction binding the contract method 0x6ceef480.
//
// Solidity: function executeBatch((address,uint256,bytes)[] calls, bytes32 predecessor, bytes32 salt) payable returns()
func (_Timelock *TimelockTransactorSession) ExecuteBatch(calls []RBACTimelockCall, predecessor [32]byte, salt [32]byte) (*types.Transaction, error) {
	return _Timelock.Contract.ExecuteBatch(&_Timelock.TransactOpts, calls, predecessor, salt)
}

// GrantRole is a paid mutator transaction binding the contract method 0x2f2ff15d.
//
// Solidity: function grantRole(bytes32 role, address account) returns()
func (_Timelock *TimelockTransactor) GrantRole(opts *bind.TransactOpts, role [32]byte, account common.Address) (*types.Transaction, error) {
	return _Timelock.contract.Transact(opts, "grantRole", role, account)
}

// GrantRole is a paid mutator transaction binding the contract method 0x2f2ff15d.
//
// Solidity: function grantRole(bytes32 role, address account) returns()
func (_Timelock *TimelockSession) GrantRole(role [32]byte, account common.Address) (*types.Transaction, error) {
	return _Timelock.Contract.GrantRole(&_Timelock.TransactOpts, role, account)
}

// GrantRole is a paid mutator transaction binding the contract method 0x2f2ff15d.
//
// Solidity: function grantRole(bytes32 role, address account) returns()
func (_Timelock *TimelockTransactorSession) GrantRole(role [32]byte, account common.Address) (*types.Transaction, error) {
	return _Timelock.Contract.GrantRole(&_Timelock.TransactOpts, role, account)
}

// OnERC1155BatchReceived is a paid mutator transaction binding the contract method 0xbc197c81.
//
// Solidity: function onERC1155BatchReceived(address , address , uint256[] , uint256[] , bytes ) returns(bytes4)
func (_Timelock *TimelockTransactor) OnERC1155BatchReceived(opts *bind.TransactOpts, arg0 common.Address, arg1 common.Address, arg2 []*big.Int, arg3 []*big.Int, arg4 []byte) (*types.Transaction, error) {
	return _Timelock.contract.Transact(opts, "onERC1155BatchReceived", arg0, arg1, arg2, arg3, arg4)
}

// OnERC1155BatchReceived is a paid mutator transaction binding the contract method 0xbc197c81.
//
// Solidity: function onERC1155BatchReceived(address , address , uint256[] , uint256[] , bytes ) returns(bytes4)
func (_Timelock *TimelockSession) OnERC1155BatchReceived(arg0 common.Address, arg1 common.Address, arg2 []*big.Int, arg3 []*big.Int, arg4 []byte) (*types.Transaction, error) {
	return _Timelock.Contract.OnERC1155BatchReceived(&_Timelock.TransactOpts, arg0, arg1, arg2, arg3, arg4)
}

// OnERC1155BatchReceived is a paid mutator transaction binding the contract method 0xbc197c81.
//
// Solidity: function onERC1155BatchReceived(address , address , uint256[] , uint256[] , bytes ) returns(bytes4)
func (_Timelock *TimelockTransactorSession) OnERC1155BatchReceived(arg0 common.Address, arg1 common.Address, arg2 []*big.Int, arg3 []*big.Int, arg4 []byte) (*types.Transaction, error) {
	return _Timelock.Contract.OnERC1155BatchReceived(&_Timelock.TransactOpts, arg0, arg1, arg2, arg3, arg4)
}

// OnERC1155Received is a paid mutator transaction binding the contract method 0xf23a6e61.
//
// Solidity: function onERC1155Received(address , address , uint256 , uint256 , bytes ) returns(bytes4)
func (_Timelock *TimelockTransactor) OnERC1155Received(opts *bind.TransactOpts, arg0 common.Address, arg1 common.Address, arg2 *big.Int, arg3 *big.Int, arg4 []byte) (*types.Transaction, error) {
	return _Timelock.contract.Transact(opts, "onERC1155Received", arg0, arg1, arg2, arg3, arg4)
}

// OnERC1155Received is a paid mutator transaction binding the contract method 0xf23a6e61.
//
// Solidity: function onERC1155Received(address , address , uint256 , uint256 , bytes ) returns(bytes4)
func (_Timelock *TimelockSession) OnERC1155Received(arg0 common.Address, arg1 common.Address, arg2 *big.Int, arg3 *big.Int, arg4 []byte) (*types.Transaction, error) {
	return _Timelock.Contract.OnERC1155Received(&_Timelock.TransactOpts, arg0, arg1, arg2, arg3, arg4)
}

// OnERC1155Received is a paid mutator transaction binding the contract method 0xf23a6e61.
//
// Solidity: function onERC1155Received(address , address , uint256 , uint256 , bytes ) returns(bytes4)
func (_Timelock *TimelockTransactorSession) OnERC1155Received(arg0 common.Address, arg1 common.Address, arg2 *big.Int, arg3 *big.Int, arg4 []byte) (*types.Transaction, error) {
	return _Timelock.Contract.OnERC1155Received(&_Timelock.TransactOpts, arg0, arg1, arg2, arg3, arg4)
}

// OnERC721Received is a paid mutator transaction binding the contract method 0x150b7a02.
//
// Solidity: function onERC721Received(address , address , uint256 , bytes ) returns(bytes4)
func (_Timelock *TimelockTransactor) OnERC721Received(opts *bind.TransactOpts, arg0 common.Address, arg1 common.Address, arg2 *big.Int, arg3 []byte) (*types.Transaction, error) {
	return _Timelock.contract.Transact(opts, "onERC721Received", arg0, arg1, arg2, arg3)
}

// OnERC721Received is a paid mutator transaction binding the contract method 0x150b7a02.
//
// Solidity: function onERC721Received(address , address , uint256 , bytes ) returns(bytes4)
func (_Timelock *TimelockSession) OnERC721Received(arg0 common.Address, arg1 common.Address, arg2 *big.Int, arg3 []byte) (*types.Transaction, error) {
	return _Timelock.Contract.OnERC721Received(&_Timelock.TransactOpts, arg0, arg1, arg2, arg3)
}

// OnERC721Received is a paid mutator transaction binding the contract method 0x150b7a02.
//
// Solidity: function onERC721Received(address , address , uint256 , bytes ) returns(bytes4)
func (_Timelock *TimelockTransactorSession) OnERC721Received(arg0 common.Address, arg1 common.Address, arg2 *big.Int, arg3 []byte) (*types.Transaction, error) {
	return _Timelock.Contract.OnERC721Received(&_Timelock.TransactOpts, arg0, arg1, arg2, arg3)
}

// RenounceRole is a paid mutator transaction binding the contract method 0x36568abe.
//
// Solidity: function renounceRole(bytes32 role, address account) returns()
func (_Timelock *TimelockTransactor) RenounceRole(opts *bind.TransactOpts, role [32]byte, account common.Address) (*types.Transaction, error) {
	return _Timelock.contract.Transact(opts, "renounceRole", role, account)
}

// RenounceRole is a paid mutator transaction binding the contract method 0x36568abe.
//
// Solidity: function renounceRole(bytes32 role, address account) returns()
func (_Timelock *TimelockSession) RenounceRole(role [32]byte, account common.Address) (*types.Transaction, error) {
	return _Timelock.Contract.RenounceRole(&_Timelock.TransactOpts, role, account)
}

// RenounceRole is a paid mutator transaction binding the contract method 0x36568abe.
//
// Solidity: function renounceRole(bytes32 role, address account) returns()
func (_Timelock *TimelockTransactorSession) RenounceRole(role [32]byte, account common.Address) (*types.Transaction, error) {
	return _Timelock.Contract.RenounceRole(&_Timelock.TransactOpts, role, account)
}

// RevokeRole is a paid mutator transaction binding the contract method 0xd547741f.
//
// Solidity: function revokeRole(bytes32 role, address account) returns()
func (_Timelock *TimelockTransactor) RevokeRole(opts *bind.TransactOpts, role [32]byte, account common.Address) (*types.Transaction, error) {
	return _Timelock.contract.Transact(opts, "revokeRole", role, account)
}

// RevokeRole is a paid mutator transaction binding the contract method 0xd547741f.
//
// Solidity: function revokeRole(bytes32 role, address account) returns()
func (_Timelock *TimelockSession) RevokeRole(role [32]byte, account common.Address) (*types.Transaction, error) {
	return _Timelock.Contract.RevokeRole(&_Timelock.TransactOpts, role, account)
}

// RevokeRole is a paid mutator transaction binding the contract method 0xd547741f.
//
// Solidity: function revokeRole(bytes32 role, address account) returns()
func (_Timelock *TimelockTransactorSession) RevokeRole(role [32]byte, account common.Address) (*types.Transaction, error) {
	return _Timelock.Contract.RevokeRole(&_Timelock.TransactOpts, role, account)
}

// ScheduleBatch is a paid mutator transaction binding the contract method 0xa944142d.
//
// Solidity: function scheduleBatch((address,uint256,bytes)[] calls, bytes32 predecessor, bytes32 salt, uint256 delay) returns()
func (_Timelock *TimelockTransactor) ScheduleBatch(opts *bind.TransactOpts, calls []RBACTimelockCall, predecessor [32]byte, salt [32]byte, delay *big.Int) (*types.Transaction, error) {
	return _Timelock.contract.Transact(opts, "scheduleBatch", calls, predecessor, salt, delay)
}

// ScheduleBatch is a paid mutator transaction binding the contract method 0xa944142d.
//
// Solidity: function scheduleBatch((address,uint256,bytes)[] calls, bytes32 predecessor, bytes32 salt, uint256 delay) returns()
func (_Timelock *TimelockSession) ScheduleBatch(calls []RBACTimelockCall, predecessor [32]byte, salt [32]byte, delay *big.Int) (*types.Transaction, error) {
	return _Timelock.Contract.ScheduleBatch(&_Timelock.TransactOpts, calls, predecessor, salt, delay)
}

// ScheduleBatch is a paid mutator transaction binding the contract method 0xa944142d.
//
// Solidity: function scheduleBatch((address,uint256,bytes)[] calls, bytes32 predecessor, bytes32 salt, uint256 delay) returns()
func (_Timelock *TimelockTransactorSession) ScheduleBatch(calls []RBACTimelockCall, predecessor [32]byte, salt [32]byte, delay *big.Int) (*types.Transaction, error) {
	return _Timelock.Contract.ScheduleBatch(&_Timelock.TransactOpts, calls, predecessor, salt, delay)
}

// UnblockFunctionSelector is a paid mutator transaction binding the contract method 0x3a98b4e4.
//
// Solidity: function unblockFunctionSelector(bytes4 selector) returns()
func (_Timelock *TimelockTransactor) UnblockFunctionSelector(opts *bind.TransactOpts, selector [4]byte) (*types.Transaction, error) {
	return _Timelock.contract.Transact(opts, "unblockFunctionSelector", selector)
}

// UnblockFunctionSelector is a paid mutator transaction binding the contract method 0x3a98b4e4.
//
// Solidity: function unblockFunctionSelector(bytes4 selector) returns()
func (_Timelock *TimelockSession) UnblockFunctionSelector(selector [4]byte) (*types.Transaction, error) {
	return _Timelock.Contract.UnblockFunctionSelector(&_Timelock.TransactOpts, selector)
}

// UnblockFunctionSelector is a paid mutator transaction binding the contract method 0x3a98b4e4.
//
// Solidity: function unblockFunctionSelector(bytes4 selector) returns()
func (_Timelock *TimelockTransactorSession) UnblockFunctionSelector(selector [4]byte) (*types.Transaction, error) {
	return _Timelock.Contract.UnblockFunctionSelector(&_Timelock.TransactOpts, selector)
}

// UpdateDelay is a paid mutator transaction binding the contract method 0x64d62353.
//
// Solidity: function updateDelay(uint256 newDelay) returns()
func (_Timelock *TimelockTransactor) UpdateDelay(opts *bind.TransactOpts, newDelay *big.Int) (*types.Transaction, error) {
	return _Timelock.contract.Transact(opts, "updateDelay", newDelay)
}

// UpdateDelay is a paid mutator transaction binding the contract method 0x64d62353.
//
// Solidity: function updateDelay(uint256 newDelay) returns()
func (_Timelock *TimelockSession) UpdateDelay(newDelay *big.Int) (*types.Transaction, error) {
	return _Timelock.Contract.UpdateDelay(&_Timelock.TransactOpts, newDelay)
}

// UpdateDelay is a paid mutator transaction binding the contract method 0x64d62353.
//
// Solidity: function updateDelay(uint256 newDelay) returns()
func (_Timelock *TimelockTransactorSession) UpdateDelay(newDelay *big.Int) (*types.Transaction, error) {
	return _Timelock.Contract.UpdateDelay(&_Timelock.TransactOpts, newDelay)
}

// Receive is a paid mutator transaction binding the contract receive function.
//
// Solidity: receive() payable returns()
func (_Timelock *TimelockTransactor) Receive(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Timelock.contract.RawTransact(opts, nil) // calldata is disallowed for receive function
}

// Receive is a paid mutator transaction binding the contract receive function.
//
// Solidity: receive() payable returns()
func (_Timelock *TimelockSession) Receive() (*types.Transaction, error) {
	return _Timelock.Contract.Receive(&_Timelock.TransactOpts)
}

// Receive is a paid mutator transaction binding the contract receive function.
//
// Solidity: receive() payable returns()
func (_Timelock *TimelockTransactorSession) Receive() (*types.Transaction, error) {
	return _Timelock.Contract.Receive(&_Timelock.TransactOpts)
}

// TimelockBypasserCallExecutedIterator is returned from FilterBypasserCallExecuted and is used to iterate over the raw logs and unpacked data for BypasserCallExecuted events raised by the Timelock contract.
type TimelockBypasserCallExecutedIterator struct {
	Event *TimelockBypasserCallExecuted // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *TimelockBypasserCallExecutedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(TimelockBypasserCallExecuted)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(TimelockBypasserCallExecuted)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *TimelockBypasserCallExecutedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *TimelockBypasserCallExecutedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// TimelockBypasserCallExecuted represents a BypasserCallExecuted event raised by the Timelock contract.
type TimelockBypasserCallExecuted struct {
	Index  *big.Int
	Target common.Address
	Value  *big.Int
	Data   []byte
	Raw    types.Log // Blockchain specific contextual infos
}

// FilterBypasserCallExecuted is a free log retrieval operation binding the contract event 0x6b983f337bab73dfe37faca733adf3ea35b45b8b144ec8ee2de3a1b224564b0c.
//
// Solidity: event BypasserCallExecuted(uint256 indexed index, address target, uint256 value, bytes data)
func (_Timelock *TimelockFilterer) FilterBypasserCallExecuted(opts *bind.FilterOpts, index []*big.Int) (*TimelockBypasserCallExecutedIterator, error) {

	var indexRule []interface{}
	for _, indexItem := range index {
		indexRule = append(indexRule, indexItem)
	}

	logs, sub, err := _Timelock.contract.FilterLogs(opts, "BypasserCallExecuted", indexRule)
	if err != nil {
		return nil, err
	}
	return &TimelockBypasserCallExecutedIterator{contract: _Timelock.contract, event: "BypasserCallExecuted", logs: logs, sub: sub}, nil
}

// WatchBypasserCallExecuted is a free log subscription operation binding the contract event 0x6b983f337bab73dfe37faca733adf3ea35b45b8b144ec8ee2de3a1b224564b0c.
//
// Solidity: event BypasserCallExecuted(uint256 indexed index, address target, uint256 value, bytes data)
func (_Timelock *TimelockFilterer) WatchBypasserCallExecuted(opts *bind.WatchOpts, sink chan<- *TimelockBypasserCallExecuted, index []*big.Int) (event.Subscription, error) {

	var indexRule []interface{}
	for _, indexItem := range index {
		indexRule = append(indexRule, indexItem)
	}

	logs, sub, err := _Timelock.contract.WatchLogs(opts, "BypasserCallExecuted", indexRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(TimelockBypasserCallExecuted)
				if err := _Timelock.contract.UnpackLog(event, "BypasserCallExecuted", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseBypasserCallExecuted is a log parse operation binding the contract event 0x6b983f337bab73dfe37faca733adf3ea35b45b8b144ec8ee2de3a1b224564b0c.
//
// Solidity: event BypasserCallExecuted(uint256 indexed index, address target, uint256 value, bytes data)
func (_Timelock *TimelockFilterer) ParseBypasserCallExecuted(log types.Log) (*TimelockBypasserCallExecuted, error) {
	event := new(TimelockBypasserCallExecuted)
	if err := _Timelock.contract.UnpackLog(event, "BypasserCallExecuted", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// TimelockCallExecutedIterator is returned from FilterCallExecuted and is used to iterate over the raw logs and unpacked data for CallExecuted events raised by the Timelock contract.
type TimelockCallExecutedIterator struct {
	Event *TimelockCallExecuted // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *TimelockCallExecutedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(TimelockCallExecuted)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(TimelockCallExecuted)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *TimelockCallExecutedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *TimelockCallExecutedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// TimelockCallExecuted represents a CallExecuted event raised by the Timelock contract.
type TimelockCallExecuted struct {
	Id     [32]byte
	Index  *big.Int
	Target common.Address
	Value  *big.Int
	Data   []byte
	Raw    types.Log // Blockchain specific contextual infos
}

// FilterCallExecuted is a free log retrieval operation binding the contract event 0xc2617efa69bab66782fa219543714338489c4e9e178271560a91b82c3f612b58.
//
// Solidity: event CallExecuted(bytes32 indexed id, uint256 indexed index, address target, uint256 value, bytes data)
func (_Timelock *TimelockFilterer) FilterCallExecuted(opts *bind.FilterOpts, id [][32]byte, index []*big.Int) (*TimelockCallExecutedIterator, error) {

	var idRule []interface{}
	for _, idItem := range id {
		idRule = append(idRule, idItem)
	}
	var indexRule []interface{}
	for _, indexItem := range index {
		indexRule = append(indexRule, indexItem)
	}

	logs, sub, err := _Timelock.contract.FilterLogs(opts, "CallExecuted", idRule, indexRule)
	if err != nil {
		return nil, err
	}
	return &TimelockCallExecutedIterator{contract: _Timelock.contract, event: "CallExecuted", logs: logs, sub: sub}, nil
}

// WatchCallExecuted is a free log subscription operation binding the contract event 0xc2617efa69bab66782fa219543714338489c4e9e178271560a91b82c3f612b58.
//
// Solidity: event CallExecuted(bytes32 indexed id, uint256 indexed index, address target, uint256 value, bytes data)
func (_Timelock *TimelockFilterer) WatchCallExecuted(opts *bind.WatchOpts, sink chan<- *TimelockCallExecuted, id [][32]byte, index []*big.Int) (event.Subscription, error) {

	var idRule []interface{}
	for _, idItem := range id {
		idRule = append(idRule, idItem)
	}
	var indexRule []interface{}
	for _, indexItem := range index {
		indexRule = append(indexRule, indexItem)
	}

	logs, sub, err := _Timelock.contract.WatchLogs(opts, "CallExecuted", idRule, indexRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(TimelockCallExecuted)
				if err := _Timelock.contract.UnpackLog(event, "CallExecuted", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseCallExecuted is a log parse operation binding the contract event 0xc2617efa69bab66782fa219543714338489c4e9e178271560a91b82c3f612b58.
//
// Solidity: event CallExecuted(bytes32 indexed id, uint256 indexed index, address target, uint256 value, bytes data)
func (_Timelock *TimelockFilterer) ParseCallExecuted(log types.Log) (*TimelockCallExecuted, error) {
	event := new(TimelockCallExecuted)
	if err := _Timelock.contract.UnpackLog(event, "CallExecuted", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// TimelockCallScheduledIterator is returned from FilterCallScheduled and is used to iterate over the raw logs and unpacked data for CallScheduled events raised by the Timelock contract.
type TimelockCallScheduledIterator struct {
	Event *TimelockCallScheduled // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *TimelockCallScheduledIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(TimelockCallScheduled)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(TimelockCallScheduled)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *TimelockCallScheduledIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *TimelockCallScheduledIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// TimelockCallScheduled represents a CallScheduled event raised by the Timelock contract.
type TimelockCallScheduled struct {
	Id          [32]byte
	Index       *big.Int
	Target      common.Address
	Value       *big.Int
	Data        []byte
	Predecessor [32]byte
	Salt        [32]byte
	Delay       *big.Int
	Raw         types.Log // Blockchain specific contextual infos
}

// FilterCallScheduled is a free log retrieval operation binding the contract event 0x4f4da6666f52e3b6dbc3638d8eae4017722678fe58bca79cd8320817807a65be.
//
// Solidity: event CallScheduled(bytes32 indexed id, uint256 indexed index, address target, uint256 value, bytes data, bytes32 predecessor, bytes32 salt, uint256 delay)
func (_Timelock *TimelockFilterer) FilterCallScheduled(opts *bind.FilterOpts, id [][32]byte, index []*big.Int) (*TimelockCallScheduledIterator, error) {

	var idRule []interface{}
	for _, idItem := range id {
		idRule = append(idRule, idItem)
	}
	var indexRule []interface{}
	for _, indexItem := range index {
		indexRule = append(indexRule, indexItem)
	}

	logs, sub, err := _Timelock.contract.FilterLogs(opts, "CallScheduled", idRule, indexRule)
	if err != nil {
		return nil, err
	}
	return &TimelockCallScheduledIterator{contract: _Timelock.contract, event: "CallScheduled", logs: logs, sub: sub}, nil
}

// WatchCallScheduled is a free log subscription operation binding the contract event 0x4f4da6666f52e3b6dbc3638d8eae4017722678fe58bca79cd8320817807a65be.
//
// Solidity: event CallScheduled(bytes32 indexed id, uint256 indexed index, address target, uint256 value, bytes data, bytes32 predecessor, bytes32 salt, uint256 delay)
func (_Timelock *TimelockFilterer) WatchCallScheduled(opts *bind.WatchOpts, sink chan<- *TimelockCallScheduled, id [][32]byte, index []*big.Int) (event.Subscription, error) {

	var idRule []interface{}
	for _, idItem := range id {
		idRule = append(idRule, idItem)
	}
	var indexRule []interface{}
	for _, indexItem := range index {
		indexRule = append(indexRule, indexItem)
	}

	logs, sub, err := _Timelock.contract.WatchLogs(opts, "CallScheduled", idRule, indexRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(TimelockCallScheduled)
				if err := _Timelock.contract.UnpackLog(event, "CallScheduled", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseCallScheduled is a log parse operation binding the contract event 0x4f4da6666f52e3b6dbc3638d8eae4017722678fe58bca79cd8320817807a65be.
//
// Solidity: event CallScheduled(bytes32 indexed id, uint256 indexed index, address target, uint256 value, bytes data, bytes32 predecessor, bytes32 salt, uint256 delay)
func (_Timelock *TimelockFilterer) ParseCallScheduled(log types.Log) (*TimelockCallScheduled, error) {
	event := new(TimelockCallScheduled)
	if err := _Timelock.contract.UnpackLog(event, "CallScheduled", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// TimelockCancelledIterator is returned from FilterCancelled and is used to iterate over the raw logs and unpacked data for Cancelled events raised by the Timelock contract.
type TimelockCancelledIterator struct {
	Event *TimelockCancelled // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *TimelockCancelledIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(TimelockCancelled)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(TimelockCancelled)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *TimelockCancelledIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *TimelockCancelledIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// TimelockCancelled represents a Cancelled event raised by the Timelock contract.
type TimelockCancelled struct {
	Id  [32]byte
	Raw types.Log // Blockchain specific contextual infos
}

// FilterCancelled is a free log retrieval operation binding the contract event 0xbaa1eb22f2a492ba1a5fea61b8df4d27c6c8b5f3971e63bb58fa14ff72eedb70.
//
// Solidity: event Cancelled(bytes32 indexed id)
func (_Timelock *TimelockFilterer) FilterCancelled(opts *bind.FilterOpts, id [][32]byte) (*TimelockCancelledIterator, error) {

	var idRule []interface{}
	for _, idItem := range id {
		idRule = append(idRule, idItem)
	}

	logs, sub, err := _Timelock.contract.FilterLogs(opts, "Cancelled", idRule)
	if err != nil {
		return nil, err
	}
	return &TimelockCancelledIterator{contract: _Timelock.contract, event: "Cancelled", logs: logs, sub: sub}, nil
}

// WatchCancelled is a free log subscription operation binding the contract event 0xbaa1eb22f2a492ba1a5fea61b8df4d27c6c8b5f3971e63bb58fa14ff72eedb70.
//
// Solidity: event Cancelled(bytes32 indexed id)
func (_Timelock *TimelockFilterer) WatchCancelled(opts *bind.WatchOpts, sink chan<- *TimelockCancelled, id [][32]byte) (event.Subscription, error) {

	var idRule []interface{}
	for _, idItem := range id {
		idRule = append(idRule, idItem)
	}

	logs, sub, err := _Timelock.contract.WatchLogs(opts, "Cancelled", idRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(TimelockCancelled)
				if err := _Timelock.contract.UnpackLog(event, "Cancelled", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseCancelled is a log parse operation binding the contract event 0xbaa1eb22f2a492ba1a5fea61b8df4d27c6c8b5f3971e63bb58fa14ff72eedb70.
//
// Solidity: event Cancelled(bytes32 indexed id)
func (_Timelock *TimelockFilterer) ParseCancelled(log types.Log) (*TimelockCancelled, error) {
	event := new(TimelockCancelled)
	if err := _Timelock.contract.UnpackLog(event, "Cancelled", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// TimelockFunctionSelectorBlockedIterator is returned from FilterFunctionSelectorBlocked and is used to iterate over the raw logs and unpacked data for FunctionSelectorBlocked events raised by the Timelock contract.
type TimelockFunctionSelectorBlockedIterator struct {
	Event *TimelockFunctionSelectorBlocked // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *TimelockFunctionSelectorBlockedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(TimelockFunctionSelectorBlocked)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(TimelockFunctionSelectorBlocked)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *TimelockFunctionSelectorBlockedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *TimelockFunctionSelectorBlockedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// TimelockFunctionSelectorBlocked represents a FunctionSelectorBlocked event raised by the Timelock contract.
type TimelockFunctionSelectorBlocked struct {
	Selector [4]byte
	Raw      types.Log // Blockchain specific contextual infos
}

// FilterFunctionSelectorBlocked is a free log retrieval operation binding the contract event 0x15b40cf8ed4c95cd3c0e1dedfdb3987c3f9bf3d3770d13ddf6dc4daa5ffae9ef.
//
// Solidity: event FunctionSelectorBlocked(bytes4 indexed selector)
func (_Timelock *TimelockFilterer) FilterFunctionSelectorBlocked(opts *bind.FilterOpts, selector [][4]byte) (*TimelockFunctionSelectorBlockedIterator, error) {

	var selectorRule []interface{}
	for _, selectorItem := range selector {
		selectorRule = append(selectorRule, selectorItem)
	}

	logs, sub, err := _Timelock.contract.FilterLogs(opts, "FunctionSelectorBlocked", selectorRule)
	if err != nil {
		return nil, err
	}
	return &TimelockFunctionSelectorBlockedIterator{contract: _Timelock.contract, event: "FunctionSelectorBlocked", logs: logs, sub: sub}, nil
}

// WatchFunctionSelectorBlocked is a free log subscription operation binding the contract event 0x15b40cf8ed4c95cd3c0e1dedfdb3987c3f9bf3d3770d13ddf6dc4daa5ffae9ef.
//
// Solidity: event FunctionSelectorBlocked(bytes4 indexed selector)
func (_Timelock *TimelockFilterer) WatchFunctionSelectorBlocked(opts *bind.WatchOpts, sink chan<- *TimelockFunctionSelectorBlocked, selector [][4]byte) (event.Subscription, error) {

	var selectorRule []interface{}
	for _, selectorItem := range selector {
		selectorRule = append(selectorRule, selectorItem)
	}

	logs, sub, err := _Timelock.contract.WatchLogs(opts, "FunctionSelectorBlocked", selectorRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(TimelockFunctionSelectorBlocked)
				if err := _Timelock.contract.UnpackLog(event, "FunctionSelectorBlocked", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseFunctionSelectorBlocked is a log parse operation binding the contract event 0x15b40cf8ed4c95cd3c0e1dedfdb3987c3f9bf3d3770d13ddf6dc4daa5ffae9ef.
//
// Solidity: event FunctionSelectorBlocked(bytes4 indexed selector)
func (_Timelock *TimelockFilterer) ParseFunctionSelectorBlocked(log types.Log) (*TimelockFunctionSelectorBlocked, error) {
	event := new(TimelockFunctionSelectorBlocked)
	if err := _Timelock.contract.UnpackLog(event, "FunctionSelectorBlocked", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// TimelockFunctionSelectorUnblockedIterator is returned from FilterFunctionSelectorUnblocked and is used to iterate over the raw logs and unpacked data for FunctionSelectorUnblocked events raised by the Timelock contract.
type TimelockFunctionSelectorUnblockedIterator struct {
	Event *TimelockFunctionSelectorUnblocked // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *TimelockFunctionSelectorUnblockedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(TimelockFunctionSelectorUnblocked)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(TimelockFunctionSelectorUnblocked)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *TimelockFunctionSelectorUnblockedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *TimelockFunctionSelectorUnblockedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// TimelockFunctionSelectorUnblocked represents a FunctionSelectorUnblocked event raised by the Timelock contract.
type TimelockFunctionSelectorUnblocked struct {
	Selector [4]byte
	Raw      types.Log // Blockchain specific contextual infos
}

// FilterFunctionSelectorUnblocked is a free log retrieval operation binding the contract event 0xd91859a8d88193a56a2983deb65a5253985141c49c70bf016880b5243bd432e1.
//
// Solidity: event FunctionSelectorUnblocked(bytes4 indexed selector)
func (_Timelock *TimelockFilterer) FilterFunctionSelectorUnblocked(opts *bind.FilterOpts, selector [][4]byte) (*TimelockFunctionSelectorUnblockedIterator, error) {

	var selectorRule []interface{}
	for _, selectorItem := range selector {
		selectorRule = append(selectorRule, selectorItem)
	}

	logs, sub, err := _Timelock.contract.FilterLogs(opts, "FunctionSelectorUnblocked", selectorRule)
	if err != nil {
		return nil, err
	}
	return &TimelockFunctionSelectorUnblockedIterator{contract: _Timelock.contract, event: "FunctionSelectorUnblocked", logs: logs, sub: sub}, nil
}

// WatchFunctionSelectorUnblocked is a free log subscription operation binding the contract event 0xd91859a8d88193a56a2983deb65a5253985141c49c70bf016880b5243bd432e1.
//
// Solidity: event FunctionSelectorUnblocked(bytes4 indexed selector)
func (_Timelock *TimelockFilterer) WatchFunctionSelectorUnblocked(opts *bind.WatchOpts, sink chan<- *TimelockFunctionSelectorUnblocked, selector [][4]byte) (event.Subscription, error) {

	var selectorRule []interface{}
	for _, selectorItem := range selector {
		selectorRule = append(selectorRule, selectorItem)
	}

	logs, sub, err := _Timelock.contract.WatchLogs(opts, "FunctionSelectorUnblocked", selectorRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(TimelockFunctionSelectorUnblocked)
				if err := _Timelock.contract.UnpackLog(event, "FunctionSelectorUnblocked", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseFunctionSelectorUnblocked is a log parse operation binding the contract event 0xd91859a8d88193a56a2983deb65a5253985141c49c70bf016880b5243bd432e1.
//
// Solidity: event FunctionSelectorUnblocked(bytes4 indexed selector)
func (_Timelock *TimelockFilterer) ParseFunctionSelectorUnblocked(log types.Log) (*TimelockFunctionSelectorUnblocked, error) {
	event := new(TimelockFunctionSelectorUnblocked)
	if err := _Timelock.contract.UnpackLog(event, "FunctionSelectorUnblocked", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// TimelockMinDelayChangeIterator is returned from FilterMinDelayChange and is used to iterate over the raw logs and unpacked data for MinDelayChange events raised by the Timelock contract.
type TimelockMinDelayChangeIterator struct {
	Event *TimelockMinDelayChange // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *TimelockMinDelayChangeIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(TimelockMinDelayChange)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(TimelockMinDelayChange)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *TimelockMinDelayChangeIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *TimelockMinDelayChangeIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// TimelockMinDelayChange represents a MinDelayChange event raised by the Timelock contract.
type TimelockMinDelayChange struct {
	OldDuration *big.Int
	NewDuration *big.Int
	Raw         types.Log // Blockchain specific contextual infos
}

// FilterMinDelayChange is a free log retrieval operation binding the contract event 0x11c24f4ead16507c69ac467fbd5e4eed5fb5c699626d2cc6d66421df253886d5.
//
// Solidity: event MinDelayChange(uint256 oldDuration, uint256 newDuration)
func (_Timelock *TimelockFilterer) FilterMinDelayChange(opts *bind.FilterOpts) (*TimelockMinDelayChangeIterator, error) {

	logs, sub, err := _Timelock.contract.FilterLogs(opts, "MinDelayChange")
	if err != nil {
		return nil, err
	}
	return &TimelockMinDelayChangeIterator{contract: _Timelock.contract, event: "MinDelayChange", logs: logs, sub: sub}, nil
}

// WatchMinDelayChange is a free log subscription operation binding the contract event 0x11c24f4ead16507c69ac467fbd5e4eed5fb5c699626d2cc6d66421df253886d5.
//
// Solidity: event MinDelayChange(uint256 oldDuration, uint256 newDuration)
func (_Timelock *TimelockFilterer) WatchMinDelayChange(opts *bind.WatchOpts, sink chan<- *TimelockMinDelayChange) (event.Subscription, error) {

	logs, sub, err := _Timelock.contract.WatchLogs(opts, "MinDelayChange")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(TimelockMinDelayChange)
				if err := _Timelock.contract.UnpackLog(event, "MinDelayChange", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseMinDelayChange is a log parse operation binding the contract event 0x11c24f4ead16507c69ac467fbd5e4eed5fb5c699626d2cc6d66421df253886d5.
//
// Solidity: event MinDelayChange(uint256 oldDuration, uint256 newDuration)
func (_Timelock *TimelockFilterer) ParseMinDelayChange(log types.Log) (*TimelockMinDelayChange, error) {
	event := new(TimelockMinDelayChange)
	if err := _Timelock.contract.UnpackLog(event, "MinDelayChange", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// TimelockRoleAdminChangedIterator is returned from FilterRoleAdminChanged and is used to iterate over the raw logs and unpacked data for RoleAdminChanged events raised by the Timelock contract.
type TimelockRoleAdminChangedIterator struct {
	Event *TimelockRoleAdminChanged // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *TimelockRoleAdminChangedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(TimelockRoleAdminChanged)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(TimelockRoleAdminChanged)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *TimelockRoleAdminChangedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *TimelockRoleAdminChangedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// TimelockRoleAdminChanged represents a RoleAdminChanged event raised by the Timelock contract.
type TimelockRoleAdminChanged struct {
	Role              [32]byte
	PreviousAdminRole [32]byte
	NewAdminRole      [32]byte
	Raw               types.Log // Blockchain specific contextual infos
}

// FilterRoleAdminChanged is a free log retrieval operation binding the contract event 0xbd79b86ffe0ab8e8776151514217cd7cacd52c909f66475c3af44e129f0b00ff.
//
// Solidity: event RoleAdminChanged(bytes32 indexed role, bytes32 indexed previousAdminRole, bytes32 indexed newAdminRole)
func (_Timelock *TimelockFilterer) FilterRoleAdminChanged(opts *bind.FilterOpts, role [][32]byte, previousAdminRole [][32]byte, newAdminRole [][32]byte) (*TimelockRoleAdminChangedIterator, error) {

	var roleRule []interface{}
	for _, roleItem := range role {
		roleRule = append(roleRule, roleItem)
	}
	var previousAdminRoleRule []interface{}
	for _, previousAdminRoleItem := range previousAdminRole {
		previousAdminRoleRule = append(previousAdminRoleRule, previousAdminRoleItem)
	}
	var newAdminRoleRule []interface{}
	for _, newAdminRoleItem := range newAdminRole {
		newAdminRoleRule = append(newAdminRoleRule, newAdminRoleItem)
	}

	logs, sub, err := _Timelock.contract.FilterLogs(opts, "RoleAdminChanged", roleRule, previousAdminRoleRule, newAdminRoleRule)
	if err != nil {
		return nil, err
	}
	return &TimelockRoleAdminChangedIterator{contract: _Timelock.contract, event: "RoleAdminChanged", logs: logs, sub: sub}, nil
}

// WatchRoleAdminChanged is a free log subscription operation binding the contract event 0xbd79b86ffe0ab8e8776151514217cd7cacd52c909f66475c3af44e129f0b00ff.
//
// Solidity: event RoleAdminChanged(bytes32 indexed role, bytes32 indexed previousAdminRole, bytes32 indexed newAdminRole)
func (_Timelock *TimelockFilterer) WatchRoleAdminChanged(opts *bind.WatchOpts, sink chan<- *TimelockRoleAdminChanged, role [][32]byte, previousAdminRole [][32]byte, newAdminRole [][32]byte) (event.Subscription, error) {

	var roleRule []interface{}
	for _, roleItem := range role {
		roleRule = append(roleRule, roleItem)
	}
	var previousAdminRoleRule []interface{}
	for _, previousAdminRoleItem := range previousAdminRole {
		previousAdminRoleRule = append(previousAdminRoleRule, previousAdminRoleItem)
	}
	var newAdminRoleRule []interface{}
	for _, newAdminRoleItem := range newAdminRole {
		newAdminRoleRule = append(newAdminRoleRule, newAdminRoleItem)
	}

	logs, sub, err := _Timelock.contract.WatchLogs(opts, "RoleAdminChanged", roleRule, previousAdminRoleRule, newAdminRoleRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(TimelockRoleAdminChanged)
				if err := _Timelock.contract.UnpackLog(event, "RoleAdminChanged", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseRoleAdminChanged is a log parse operation binding the contract event 0xbd79b86ffe0ab8e8776151514217cd7cacd52c909f66475c3af44e129f0b00ff.
//
// Solidity: event RoleAdminChanged(bytes32 indexed role, bytes32 indexed previousAdminRole, bytes32 indexed newAdminRole)
func (_Timelock *TimelockFilterer) ParseRoleAdminChanged(log types.Log) (*TimelockRoleAdminChanged, error) {
	event := new(TimelockRoleAdminChanged)
	if err := _Timelock.contract.UnpackLog(event, "RoleAdminChanged", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// TimelockRoleGrantedIterator is returned from FilterRoleGranted and is used to iterate over the raw logs and unpacked data for RoleGranted events raised by the Timelock contract.
type TimelockRoleGrantedIterator struct {
	Event *TimelockRoleGranted // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *TimelockRoleGrantedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(TimelockRoleGranted)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(TimelockRoleGranted)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *TimelockRoleGrantedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *TimelockRoleGrantedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// TimelockRoleGranted represents a RoleGranted event raised by the Timelock contract.
type TimelockRoleGranted struct {
	Role    [32]byte
	Account common.Address
	Sender  common.Address
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterRoleGranted is a free log retrieval operation binding the contract event 0x2f8788117e7eff1d82e926ec794901d17c78024a50270940304540a733656f0d.
//
// Solidity: event RoleGranted(bytes32 indexed role, address indexed account, address indexed sender)
func (_Timelock *TimelockFilterer) FilterRoleGranted(opts *bind.FilterOpts, role [][32]byte, account []common.Address, sender []common.Address) (*TimelockRoleGrantedIterator, error) {

	var roleRule []interface{}
	for _, roleItem := range role {
		roleRule = append(roleRule, roleItem)
	}
	var accountRule []interface{}
	for _, accountItem := range account {
		accountRule = append(accountRule, accountItem)
	}
	var senderRule []interface{}
	for _, senderItem := range sender {
		senderRule = append(senderRule, senderItem)
	}

	logs, sub, err := _Timelock.contract.FilterLogs(opts, "RoleGranted", roleRule, accountRule, senderRule)
	if err != nil {
		return nil, err
	}
	return &TimelockRoleGrantedIterator{contract: _Timelock.contract, event: "RoleGranted", logs: logs, sub: sub}, nil
}

// WatchRoleGranted is a free log subscription operation binding the contract event 0x2f8788117e7eff1d82e926ec794901d17c78024a50270940304540a733656f0d.
//
// Solidity: event RoleGranted(bytes32 indexed role, address indexed account, address indexed sender)
func (_Timelock *TimelockFilterer) WatchRoleGranted(opts *bind.WatchOpts, sink chan<- *TimelockRoleGranted, role [][32]byte, account []common.Address, sender []common.Address) (event.Subscription, error) {

	var roleRule []interface{}
	for _, roleItem := range role {
		roleRule = append(roleRule, roleItem)
	}
	var accountRule []interface{}
	for _, accountItem := range account {
		accountRule = append(accountRule, accountItem)
	}
	var senderRule []interface{}
	for _, senderItem := range sender {
		senderRule = append(senderRule, senderItem)
	}

	logs, sub, err := _Timelock.contract.WatchLogs(opts, "RoleGranted", roleRule, accountRule, senderRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(TimelockRoleGranted)
				if err := _Timelock.contract.UnpackLog(event, "RoleGranted", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseRoleGranted is a log parse operation binding the contract event 0x2f8788117e7eff1d82e926ec794901d17c78024a50270940304540a733656f0d.
//
// Solidity: event RoleGranted(bytes32 indexed role, address indexed account, address indexed sender)
func (_Timelock *TimelockFilterer) ParseRoleGranted(log types.Log) (*TimelockRoleGranted, error) {
	event := new(TimelockRoleGranted)
	if err := _Timelock.contract.UnpackLog(event, "RoleGranted", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// TimelockRoleRevokedIterator is returned from FilterRoleRevoked and is used to iterate over the raw logs and unpacked data for RoleRevoked events raised by the Timelock contract.
type TimelockRoleRevokedIterator struct {
	Event *TimelockRoleRevoked // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *TimelockRoleRevokedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(TimelockRoleRevoked)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(TimelockRoleRevoked)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *TimelockRoleRevokedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *TimelockRoleRevokedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// TimelockRoleRevoked represents a RoleRevoked event raised by the Timelock contract.
type TimelockRoleRevoked struct {
	Role    [32]byte
	Account common.Address
	Sender  common.Address
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterRoleRevoked is a free log retrieval operation binding the contract event 0xf6391f5c32d9c69d2a47ea670b442974b53935d1edc7fd64eb21e047a839171b.
//
// Solidity: event RoleRevoked(bytes32 indexed role, address indexed account, address indexed sender)
func (_Timelock *TimelockFilterer) FilterRoleRevoked(opts *bind.FilterOpts, role [][32]byte, account []common.Address, sender []common.Address) (*TimelockRoleRevokedIterator, error) {

	var roleRule []interface{}
	for _, roleItem := range role {
		roleRule = append(roleRule, roleItem)
	}
	var accountRule []interface{}
	for _, accountItem := range account {
		accountRule = append(accountRule, accountItem)
	}
	var senderRule []interface{}
	for _, senderItem := range sender {
		senderRule = append(senderRule, senderItem)
	}

	logs, sub, err := _Timelock.contract.FilterLogs(opts, "RoleRevoked", roleRule, accountRule, senderRule)
	if err != nil {
		return nil, err
	}
	return &TimelockRoleRevokedIterator{contract: _Timelock.contract, event: "RoleRevoked", logs: logs, sub: sub}, nil
}

// WatchRoleRevoked is a free log subscription operation binding the contract event 0xf6391f5c32d9c69d2a47ea670b442974b53935d1edc7fd64eb21e047a839171b.
//
// Solidity: event RoleRevoked(bytes32 indexed role, address indexed account, address indexed sender)
func (_Timelock *TimelockFilterer) WatchRoleRevoked(opts *bind.WatchOpts, sink chan<- *TimelockRoleRevoked, role [][32]byte, account []common.Address, sender []common.Address) (event.Subscription, error) {

	var roleRule []interface{}
	for _, roleItem := range role {
		roleRule = append(roleRule, roleItem)
	}
	var accountRule []interface{}
	for _, accountItem := range account {
		accountRule = append(accountRule, accountItem)
	}
	var senderRule []interface{}
	for _, senderItem := range sender {
		senderRule = append(senderRule, senderItem)
	}

	logs, sub, err := _Timelock.contract.WatchLogs(opts, "RoleRevoked", roleRule, accountRule, senderRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(TimelockRoleRevoked)
				if err := _Timelock.contract.UnpackLog(event, "RoleRevoked", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseRoleRevoked is a log parse operation binding the contract event 0xf6391f5c32d9c69d2a47ea670b442974b53935d1edc7fd64eb21e047a839171b.
//
// Solidity: event RoleRevoked(bytes32 indexed role, address indexed account, address indexed sender)
func (_Timelock *TimelockFilterer) ParseRoleRevoked(log types.Log) (*TimelockRoleRevoked, error) {
	event := new(TimelockRoleRevoked)
	if err := _Timelock.contract.UnpackLog(event, "RoleRevoked", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

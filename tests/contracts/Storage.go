// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package contracts

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

// StorageContractMetaData contains all meta data concerning the StorageContract contract.
var StorageContractMetaData = &bind.MetaData{
	ABI: "[{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"_from\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"_oldNumber\",\"type\":\"uint256\"},{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"_number\",\"type\":\"uint256\"}],\"name\":\"storedNumber\",\"type\":\"event\"},{\"inputs\":[],\"name\":\"retrieve\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"num\",\"type\":\"uint256\"}],\"name\":\"store\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]",
	Bin: "0x608060405234801561001057600080fd5b5060bc8061001f6000396000f3fe6080604052348015600f57600080fd5b506004361060325760003560e01c80632e64cec11460375780636057361d14604c575b600080fd5b60005460405190815260200160405180910390f35b605b60573660046097565b605d565b005b6000805482825560405190918391839133917f87f16aa184eca14ea45e132328a5effbb79b9f921657bd03d83608f26d76f3cf9190a45050565b60006020828403121560a857600080fd5b503591905056fea164736f6c634300080f000a",
}

// StorageContractABI is the input ABI used to generate the binding from.
// Deprecated: Use StorageContractMetaData.ABI instead.
var StorageContractABI = StorageContractMetaData.ABI

// StorageContractBin is the compiled bytecode used for deploying new contracts.
// Deprecated: Use StorageContractMetaData.Bin instead.
var StorageContractBin = StorageContractMetaData.Bin

// DeployStorageContract deploys a new Ethereum contract, binding an instance of StorageContract to it.
func DeployStorageContract(auth *bind.TransactOpts, backend bind.ContractBackend) (common.Address, *types.Transaction, *StorageContract, error) {
	parsed, err := StorageContractMetaData.GetAbi()
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	if parsed == nil {
		return common.Address{}, nil, nil, errors.New("GetABI returned nil")
	}

	address, tx, contract, err := bind.DeployContract(auth, *parsed, common.FromHex(StorageContractBin), backend)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &StorageContract{StorageContractCaller: StorageContractCaller{contract: contract}, StorageContractTransactor: StorageContractTransactor{contract: contract}, StorageContractFilterer: StorageContractFilterer{contract: contract}}, nil
}

// StorageContract is an auto generated Go binding around an Ethereum contract.
type StorageContract struct {
	StorageContractCaller     // Read-only binding to the contract
	StorageContractTransactor // Write-only binding to the contract
	StorageContractFilterer   // Log filterer for contract events
}

// StorageContractCaller is an auto generated read-only Go binding around an Ethereum contract.
type StorageContractCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// StorageContractTransactor is an auto generated write-only Go binding around an Ethereum contract.
type StorageContractTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// StorageContractFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type StorageContractFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// StorageContractSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type StorageContractSession struct {
	Contract     *StorageContract  // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// StorageContractCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type StorageContractCallerSession struct {
	Contract *StorageContractCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts          // Call options to use throughout this session
}

// StorageContractTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type StorageContractTransactorSession struct {
	Contract     *StorageContractTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts          // Transaction auth options to use throughout this session
}

// StorageContractRaw is an auto generated low-level Go binding around an Ethereum contract.
type StorageContractRaw struct {
	Contract *StorageContract // Generic contract binding to access the raw methods on
}

// StorageContractCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type StorageContractCallerRaw struct {
	Contract *StorageContractCaller // Generic read-only contract binding to access the raw methods on
}

// StorageContractTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type StorageContractTransactorRaw struct {
	Contract *StorageContractTransactor // Generic write-only contract binding to access the raw methods on
}

// NewStorageContract creates a new instance of StorageContract, bound to a specific deployed contract.
func NewStorageContract(address common.Address, backend bind.ContractBackend) (*StorageContract, error) {
	contract, err := bindStorageContract(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &StorageContract{StorageContractCaller: StorageContractCaller{contract: contract}, StorageContractTransactor: StorageContractTransactor{contract: contract}, StorageContractFilterer: StorageContractFilterer{contract: contract}}, nil
}

// NewStorageContractCaller creates a new read-only instance of StorageContract, bound to a specific deployed contract.
func NewStorageContractCaller(address common.Address, caller bind.ContractCaller) (*StorageContractCaller, error) {
	contract, err := bindStorageContract(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &StorageContractCaller{contract: contract}, nil
}

// NewStorageContractTransactor creates a new write-only instance of StorageContract, bound to a specific deployed contract.
func NewStorageContractTransactor(address common.Address, transactor bind.ContractTransactor) (*StorageContractTransactor, error) {
	contract, err := bindStorageContract(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &StorageContractTransactor{contract: contract}, nil
}

// NewStorageContractFilterer creates a new log filterer instance of StorageContract, bound to a specific deployed contract.
func NewStorageContractFilterer(address common.Address, filterer bind.ContractFilterer) (*StorageContractFilterer, error) {
	contract, err := bindStorageContract(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &StorageContractFilterer{contract: contract}, nil
}

// bindStorageContract binds a generic wrapper to an already deployed contract.
func bindStorageContract(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := StorageContractMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_StorageContract *StorageContractRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _StorageContract.Contract.StorageContractCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_StorageContract *StorageContractRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _StorageContract.Contract.StorageContractTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_StorageContract *StorageContractRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _StorageContract.Contract.StorageContractTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_StorageContract *StorageContractCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _StorageContract.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_StorageContract *StorageContractTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _StorageContract.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_StorageContract *StorageContractTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _StorageContract.Contract.contract.Transact(opts, method, params...)
}

// Retrieve is a free data retrieval call binding the contract method 0x2e64cec1.
//
// Solidity: function retrieve() view returns(uint256)
func (_StorageContract *StorageContractCaller) Retrieve(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _StorageContract.contract.Call(opts, &out, "retrieve")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// Retrieve is a free data retrieval call binding the contract method 0x2e64cec1.
//
// Solidity: function retrieve() view returns(uint256)
func (_StorageContract *StorageContractSession) Retrieve() (*big.Int, error) {
	return _StorageContract.Contract.Retrieve(&_StorageContract.CallOpts)
}

// Retrieve is a free data retrieval call binding the contract method 0x2e64cec1.
//
// Solidity: function retrieve() view returns(uint256)
func (_StorageContract *StorageContractCallerSession) Retrieve() (*big.Int, error) {
	return _StorageContract.Contract.Retrieve(&_StorageContract.CallOpts)
}

// Store is a paid mutator transaction binding the contract method 0x6057361d.
//
// Solidity: function store(uint256 num) returns()
func (_StorageContract *StorageContractTransactor) Store(opts *bind.TransactOpts, num *big.Int) (*types.Transaction, error) {
	return _StorageContract.contract.Transact(opts, "store", num)
}

// Store is a paid mutator transaction binding the contract method 0x6057361d.
//
// Solidity: function store(uint256 num) returns()
func (_StorageContract *StorageContractSession) Store(num *big.Int) (*types.Transaction, error) {
	return _StorageContract.Contract.Store(&_StorageContract.TransactOpts, num)
}

// Store is a paid mutator transaction binding the contract method 0x6057361d.
//
// Solidity: function store(uint256 num) returns()
func (_StorageContract *StorageContractTransactorSession) Store(num *big.Int) (*types.Transaction, error) {
	return _StorageContract.Contract.Store(&_StorageContract.TransactOpts, num)
}

// StorageContractStoredNumberIterator is returned from FilterStoredNumber and is used to iterate over the raw logs and unpacked data for StoredNumber events raised by the StorageContract contract.
type StorageContractStoredNumberIterator struct {
	Event *StorageContractStoredNumber // Event containing the contract specifics and raw log

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
func (it *StorageContractStoredNumberIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(StorageContractStoredNumber)
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
		it.Event = new(StorageContractStoredNumber)
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
func (it *StorageContractStoredNumberIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *StorageContractStoredNumberIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// StorageContractStoredNumber represents a StoredNumber event raised by the StorageContract contract.
type StorageContractStoredNumber struct {
	From      common.Address
	OldNumber *big.Int
	Number    *big.Int
	Raw       types.Log // Blockchain specific contextual infos
}

// FilterStoredNumber is a free log retrieval operation binding the contract event 0x87f16aa184eca14ea45e132328a5effbb79b9f921657bd03d83608f26d76f3cf.
//
// Solidity: event storedNumber(address indexed _from, uint256 indexed _oldNumber, uint256 indexed _number)
func (_StorageContract *StorageContractFilterer) FilterStoredNumber(opts *bind.FilterOpts, _from []common.Address, _oldNumber []*big.Int, _number []*big.Int) (*StorageContractStoredNumberIterator, error) {

	var _fromRule []interface{}
	for _, _fromItem := range _from {
		_fromRule = append(_fromRule, _fromItem)
	}
	var _oldNumberRule []interface{}
	for _, _oldNumberItem := range _oldNumber {
		_oldNumberRule = append(_oldNumberRule, _oldNumberItem)
	}
	var _numberRule []interface{}
	for _, _numberItem := range _number {
		_numberRule = append(_numberRule, _numberItem)
	}

	logs, sub, err := _StorageContract.contract.FilterLogs(opts, "storedNumber", _fromRule, _oldNumberRule, _numberRule)
	if err != nil {
		return nil, err
	}
	return &StorageContractStoredNumberIterator{contract: _StorageContract.contract, event: "storedNumber", logs: logs, sub: sub}, nil
}

// WatchStoredNumber is a free log subscription operation binding the contract event 0x87f16aa184eca14ea45e132328a5effbb79b9f921657bd03d83608f26d76f3cf.
//
// Solidity: event storedNumber(address indexed _from, uint256 indexed _oldNumber, uint256 indexed _number)
func (_StorageContract *StorageContractFilterer) WatchStoredNumber(opts *bind.WatchOpts, sink chan<- *StorageContractStoredNumber, _from []common.Address, _oldNumber []*big.Int, _number []*big.Int) (event.Subscription, error) {

	var _fromRule []interface{}
	for _, _fromItem := range _from {
		_fromRule = append(_fromRule, _fromItem)
	}
	var _oldNumberRule []interface{}
	for _, _oldNumberItem := range _oldNumber {
		_oldNumberRule = append(_oldNumberRule, _oldNumberItem)
	}
	var _numberRule []interface{}
	for _, _numberItem := range _number {
		_numberRule = append(_numberRule, _numberItem)
	}

	logs, sub, err := _StorageContract.contract.WatchLogs(opts, "storedNumber", _fromRule, _oldNumberRule, _numberRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(StorageContractStoredNumber)
				if err := _StorageContract.contract.UnpackLog(event, "storedNumber", log); err != nil {
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

// ParseStoredNumber is a log parse operation binding the contract event 0x87f16aa184eca14ea45e132328a5effbb79b9f921657bd03d83608f26d76f3cf.
//
// Solidity: event storedNumber(address indexed _from, uint256 indexed _oldNumber, uint256 indexed _number)
func (_StorageContract *StorageContractFilterer) ParseStoredNumber(log types.Log) (*StorageContractStoredNumber, error) {
	event := new(StorageContractStoredNumber)
	if err := _StorageContract.contract.UnpackLog(event, "storedNumber", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

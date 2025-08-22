// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package helloworld

import (
	"math/big"
	"strings"

	"github.com/FISCO-BCOS/go-sdk/abi"
	"github.com/FISCO-BCOS/go-sdk/abi/bind"
	"github.com/FISCO-BCOS/go-sdk/core/types"
	ethereum "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
)

// Reference imports to suppress errors if they are not otherwise used.
var (
	_ = big.NewInt
	_ = strings.NewReader
	_ = ethereum.NotFound
	_ = abi.U256
	_ = bind.Bind
	_ = common.Big1
	_ = types.BloomLookup
)

// CertOperKVABI is the input ABI used to generate the binding from.
const CertOperKVABI = "[{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"oper_id\",\"type\":\"bytes32\"},{\"indexed\":false,\"internalType\":\"string\",\"name\":\"value\",\"type\":\"string\"}],\"name\":\"Set\",\"type\":\"event\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"oper_id\",\"type\":\"bytes32\"}],\"name\":\"get\",\"outputs\":[{\"internalType\":\"string\",\"name\":\"\",\"type\":\"string\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"oper_id\",\"type\":\"bytes32\"},{\"internalType\":\"string\",\"name\":\"value\",\"type\":\"string\"}],\"name\":\"set\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]"

// CertOperKVBin is the compiled bytecode used for deploying new contracts.
var CertOperKVBin = "0x608060405234801561001057600080fd5b506103fe806100206000396000f3fe608060405234801561001057600080fd5b50600436106100365760003560e01c80638eaa6ac01461003b578063b4800033146100e2575b600080fd5b6100676004803603602081101561005157600080fd5b81019080803590602001909291905050506101a7565b6040518080602001828103825283818151815260200191508051906020019080838360005b838110156100a757808201518184015260208101905061008c565b50505050905090810190601f1680156100d45780820380516001836020036101000a031916815260200191505b509250505060405180910390f35b6101a5600480360360408110156100f857600080fd5b81019080803590602001909291908035906020019064010000000081111561011f57600080fd5b82018360208201111561013157600080fd5b8035906020019184600183028401116401000000008311171561015357600080fd5b91908080601f016020809104026020016040519081016040528093929190818152602001838380828437600081840152601f19601f82011690508083019250505050505050919291929050505061025b565b005b60606000808381526020019081526020016000208054600181600116156101000203166002900480601f01602080910402602001604051908101604052809291908181526020018280546001816001161561010002031660029004801561024f5780601f106102245761010080835404028352916020019161024f565b820191906000526020600020905b81548152906001019060200180831161023257829003601f168201915b50505050509050919050565b806000808481526020019081526020016000209080519060200190610281929190610323565b50817fc664efa026b4ec93a75df843ec1e56838c5932a79b631b0b759b345f431a441a826040518080602001828103825283818151815260200191508051906020019080838360005b838110156102e55780820151818401526020810190506102ca565b50505050905090810190601f1680156103125780820380516001836020036101000a031916815260200191505b509250505060405180910390a25050565b828054600181600116156101000203166002900490600052602060002090601f016020900481019282601f1061036457805160ff1916838001178555610392565b82800160010185558215610392579182015b82811115610391578251825591602001919060010190610376565b5b50905061039f91906103a3565b5090565b6103c591905b808211156103c15760008160009055506001016103a9565b5090565b9056fea264697066735822122024e95c9244c33ca6249d6e5c354d22b4f19549b4e8e572d1f49be0c8b3fd488464736f6c634300060a0033"

// DeployCertOperKV deploys a new kvtabletest, binding an instance of CertOperKV to it.
func DeployCertOperKV(auth *bind.TransactOpts, backend bind.ContractBackend) (common.Address, *types.Transaction, *CertOperKV, error) {
	parsed, err := abi.JSON(strings.NewReader(CertOperKVABI))
	if err != nil {
		return common.Address{}, nil, nil, err
	}

	address, tx, kvtabletest, err := bind.DeployContract(auth, parsed, common.FromHex(CertOperKVBin), backend)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &CertOperKV{CertOperKVCaller: CertOperKVCaller{kvtabletest: kvtabletest}, CertOperKVTransactor: CertOperKVTransactor{kvtabletest: kvtabletest}, CertOperKVFilterer: CertOperKVFilterer{kvtabletest: kvtabletest}}, nil
}

func AsyncDeployCertOperKV(auth *bind.TransactOpts, handler func(*types.Receipt, error), backend bind.ContractBackend) (*types.Transaction, error) {
	parsed, err := abi.JSON(strings.NewReader(CertOperKVABI))
	if err != nil {
		return nil, err
	}

	tx, err := bind.AsyncDeployContract(auth, handler, parsed, common.FromHex(CertOperKVBin), backend)
	if err != nil {
		return nil, err
	}
	return tx, nil
}

// CertOperKV is an auto generated Go binding around a Solidity kvtabletest.
type CertOperKV struct {
	CertOperKVCaller     // Read-only binding to the kvtabletest
	CertOperKVTransactor // Write-only binding to the kvtabletest
	CertOperKVFilterer   // Log filterer for kvtabletest events
}

// CertOperKVCaller is an auto generated read-only Go binding around a Solidity kvtabletest.
type CertOperKVCaller struct {
	kvtabletest *bind.BoundContract // Generic kvtabletest wrapper for the low level calls
}

// CertOperKVTransactor is an auto generated write-only Go binding around a Solidity kvtabletest.
type CertOperKVTransactor struct {
	kvtabletest *bind.BoundContract // Generic kvtabletest wrapper for the low level calls
}

// CertOperKVFilterer is an auto generated log filtering Go binding around a Solidity kvtabletest events.
type CertOperKVFilterer struct {
	kvtabletest *bind.BoundContract // Generic kvtabletest wrapper for the low level calls
}

// CertOperKVSession is an auto generated Go binding around a Solidity kvtabletest,
// with pre-set call and transact options.
type CertOperKVSession struct {
	Contract     *CertOperKV       // Generic kvtabletest binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// CertOperKVCallerSession is an auto generated read-only Go binding around a Solidity kvtabletest,
// with pre-set call options.
type CertOperKVCallerSession struct {
	Contract *CertOperKVCaller // Generic kvtabletest caller binding to set the session for
	CallOpts bind.CallOpts     // Call options to use throughout this session
}

// CertOperKVTransactorSession is an auto generated write-only Go binding around a Solidity kvtabletest,
// with pre-set transact options.
type CertOperKVTransactorSession struct {
	Contract     *CertOperKVTransactor // Generic kvtabletest transactor binding to set the session for
	TransactOpts bind.TransactOpts     // Transaction auth options to use throughout this session
}

// CertOperKVRaw is an auto generated low-level Go binding around a Solidity kvtabletest.
type CertOperKVRaw struct {
	Contract *CertOperKV // Generic kvtabletest binding to access the raw methods on
}

// CertOperKVCallerRaw is an auto generated low-level read-only Go binding around a Solidity kvtabletest.
type CertOperKVCallerRaw struct {
	Contract *CertOperKVCaller // Generic read-only kvtabletest binding to access the raw methods on
}

// CertOperKVTransactorRaw is an auto generated low-level write-only Go binding around a Solidity kvtabletest.
type CertOperKVTransactorRaw struct {
	Contract *CertOperKVTransactor // Generic write-only kvtabletest binding to access the raw methods on
}

// NewCertOperKV creates a new instance of CertOperKV, bound to a specific deployed kvtabletest.
func NewCertOperKV(address common.Address, backend bind.ContractBackend) (*CertOperKV, error) {
	kvtabletest, err := bindCertOperKV(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &CertOperKV{CertOperKVCaller: CertOperKVCaller{kvtabletest: kvtabletest}, CertOperKVTransactor: CertOperKVTransactor{kvtabletest: kvtabletest}, CertOperKVFilterer: CertOperKVFilterer{kvtabletest: kvtabletest}}, nil
}

// NewCertOperKVCaller creates a new read-only instance of CertOperKV, bound to a specific deployed kvtabletest.
func NewCertOperKVCaller(address common.Address, caller bind.ContractCaller) (*CertOperKVCaller, error) {
	kvtabletest, err := bindCertOperKV(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &CertOperKVCaller{kvtabletest: kvtabletest}, nil
}

// NewCertOperKVTransactor creates a new write-only instance of CertOperKV, bound to a specific deployed kvtabletest.
func NewCertOperKVTransactor(address common.Address, transactor bind.ContractTransactor) (*CertOperKVTransactor, error) {
	kvtabletest, err := bindCertOperKV(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &CertOperKVTransactor{kvtabletest: kvtabletest}, nil
}

// NewCertOperKVFilterer creates a new log filterer instance of CertOperKV, bound to a specific deployed kvtabletest.
func NewCertOperKVFilterer(address common.Address, filterer bind.ContractFilterer) (*CertOperKVFilterer, error) {
	kvtabletest, err := bindCertOperKV(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &CertOperKVFilterer{kvtabletest: kvtabletest}, nil
}

// bindCertOperKV binds a generic wrapper to an already deployed kvtabletest.
func bindCertOperKV(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(CertOperKVABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) kvtabletest method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_CertOperKV *CertOperKVRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _CertOperKV.Contract.CertOperKVCaller.kvtabletest.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the kvtabletest, calling
// its default method if one is available.
func (_CertOperKV *CertOperKVRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, *types.Receipt, error) {
	return _CertOperKV.Contract.CertOperKVTransactor.kvtabletest.Transfer(opts)
}

// Transact invokes the (paid) kvtabletest method with params as input values.
func (_CertOperKV *CertOperKVRaw) TransactWithResult(opts *bind.TransactOpts, result interface{}, method string, params ...interface{}) (*types.Transaction, *types.Receipt, error) {
	return _CertOperKV.Contract.CertOperKVTransactor.kvtabletest.TransactWithResult(opts, result, method, params...)
}

// Call invokes the (constant) kvtabletest method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_CertOperKV *CertOperKVCallerRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _CertOperKV.Contract.kvtabletest.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the kvtabletest, calling
// its default method if one is available.
func (_CertOperKV *CertOperKVTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, *types.Receipt, error) {
	return _CertOperKV.Contract.kvtabletest.Transfer(opts)
}

// Transact invokes the (paid) kvtabletest method with params as input values.
func (_CertOperKV *CertOperKVTransactorRaw) TransactWithResult(opts *bind.TransactOpts, result interface{}, method string, params ...interface{}) (*types.Transaction, *types.Receipt, error) {
	return _CertOperKV.Contract.kvtabletest.TransactWithResult(opts, result, method, params...)
}

// Get is a free data retrieval call binding the kvtabletest method 0x8eaa6ac0.
//
// Solidity: function get(bytes32 oper_id) constant returns(string)
func (_CertOperKV *CertOperKVCaller) Get(opts *bind.CallOpts, oper_id [32]byte) (string, error) {
	var (
		ret0 = new(string)
	)
	out := ret0
	err := _CertOperKV.kvtabletest.Call(opts, out, "get", oper_id)
	return *ret0, err
}

// Get is a free data retrieval call binding the kvtabletest method 0x8eaa6ac0.
//
// Solidity: function get(bytes32 oper_id) constant returns(string)
func (_CertOperKV *CertOperKVSession) Get(oper_id [32]byte) (string, error) {
	return _CertOperKV.Contract.Get(&_CertOperKV.CallOpts, oper_id)
}

// Get is a free data retrieval call binding the kvtabletest method 0x8eaa6ac0.
//
// Solidity: function get(bytes32 oper_id) constant returns(string)
func (_CertOperKV *CertOperKVCallerSession) Get(oper_id [32]byte) (string, error) {
	return _CertOperKV.Contract.Get(&_CertOperKV.CallOpts, oper_id)
}

// Set is a paid mutator transaction binding the kvtabletest method 0xb4800033.
//
// Solidity: function set(bytes32 oper_id, string value) returns()
func (_CertOperKV *CertOperKVTransactor) Set(opts *bind.TransactOpts, oper_id [32]byte, value string) (*types.Transaction, *types.Receipt, error) {
	var ()
	out := &[]interface{}{}
	transaction, receipt, err := _CertOperKV.kvtabletest.TransactWithResult(opts, out, "set", oper_id, value)
	return transaction, receipt, err
}

func (_CertOperKV *CertOperKVTransactor) AsyncSet(handler func(*types.Receipt, error), opts *bind.TransactOpts, oper_id [32]byte, value string) (*types.Transaction, error) {
	return _CertOperKV.kvtabletest.AsyncTransact(opts, handler, "set", oper_id, value)
}

// Set is a paid mutator transaction binding the kvtabletest method 0xb4800033.
//
// Solidity: function set(bytes32 oper_id, string value) returns()
func (_CertOperKV *CertOperKVSession) Set(oper_id [32]byte, value string) (*types.Transaction, *types.Receipt, error) {
	return _CertOperKV.Contract.Set(&_CertOperKV.TransactOpts, oper_id, value)
}

func (_CertOperKV *CertOperKVSession) AsyncSet(handler func(*types.Receipt, error), oper_id [32]byte, value string) (*types.Transaction, error) {
	return _CertOperKV.Contract.AsyncSet(handler, &_CertOperKV.TransactOpts, oper_id, value)
}

// Set is a paid mutator transaction binding the kvtabletest method 0xb4800033.
//
// Solidity: function set(bytes32 oper_id, string value) returns()
func (_CertOperKV *CertOperKVTransactorSession) Set(oper_id [32]byte, value string) (*types.Transaction, *types.Receipt, error) {
	return _CertOperKV.Contract.Set(&_CertOperKV.TransactOpts, oper_id, value)
}

func (_CertOperKV *CertOperKVTransactorSession) AsyncSet(handler func(*types.Receipt, error), oper_id [32]byte, value string) (*types.Transaction, error) {
	return _CertOperKV.Contract.AsyncSet(handler, &_CertOperKV.TransactOpts, oper_id, value)
}

// CertOperKVSet represents a Set event raised by the CertOperKV kvtabletest.
type CertOperKVSet struct {
	OperId [32]byte
	Value  string
	Raw    types.Log // Blockchain specific contextual infos
}

// WatchSet is a free log subscription operation binding the kvtabletest event 0xc664efa026b4ec93a75df843ec1e56838c5932a79b631b0b759b345f431a441a.
//
// Solidity: event Set(bytes32 indexed oper_id, string value)
func (_CertOperKV *CertOperKVFilterer) WatchSet(fromBlock *uint64, handler func(int, []types.Log), oper_id [32]byte) (string, error) {
	return _CertOperKV.kvtabletest.WatchLogs(fromBlock, handler, "Set", oper_id)
}

func (_CertOperKV *CertOperKVFilterer) WatchAllSet(fromBlock *uint64, handler func(int, []types.Log)) (string, error) {
	return _CertOperKV.kvtabletest.WatchLogs(fromBlock, handler, "Set")
}

// ParseSet is a log parse operation binding the kvtabletest event 0xc664efa026b4ec93a75df843ec1e56838c5932a79b631b0b759b345f431a441a.
//
// Solidity: event Set(bytes32 indexed oper_id, string value)
func (_CertOperKV *CertOperKVFilterer) ParseSet(log types.Log) (*CertOperKVSet, error) {
	event := new(CertOperKVSet)
	if err := _CertOperKV.kvtabletest.UnpackLog(event, "Set", log); err != nil {
		return nil, err
	}
	return event, nil
}

// WatchSet is a free log subscription operation binding the kvtabletest event 0xc664efa026b4ec93a75df843ec1e56838c5932a79b631b0b759b345f431a441a.
//
// Solidity: event Set(bytes32 indexed oper_id, string value)
func (_CertOperKV *CertOperKVSession) WatchSet(fromBlock *uint64, handler func(int, []types.Log), oper_id [32]byte) (string, error) {
	return _CertOperKV.Contract.WatchSet(fromBlock, handler, oper_id)
}

func (_CertOperKV *CertOperKVSession) WatchAllSet(fromBlock *uint64, handler func(int, []types.Log)) (string, error) {
	return _CertOperKV.Contract.WatchAllSet(fromBlock, handler)
}

// ParseSet is a log parse operation binding the kvtabletest event 0xc664efa026b4ec93a75df843ec1e56838c5932a79b631b0b759b345f431a441a.
//
// Solidity: event Set(bytes32 indexed oper_id, string value)
func (_CertOperKV *CertOperKVSession) ParseSet(log types.Log) (*CertOperKVSet, error) {
	return _CertOperKV.Contract.ParseSet(log)
}

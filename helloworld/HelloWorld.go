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

// HelloWorldABI is the input ABI used to generate the binding from.
const HelloWorldABI = "[{\"inputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"inputs\":[],\"name\":\"get\",\"outputs\":[{\"internalType\":\"string\",\"name\":\"\",\"type\":\"string\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"string\",\"name\":\"v\",\"type\":\"string\"}],\"name\":\"set\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]"

// HelloWorldBin is the compiled bytecode used for deploying new contracts.
var HelloWorldBin = "0x608060405234801561001057600080fd5b506040518060400160405280600d81526020017f48656c6c6f2c20576f726c6421000000000000000000000000000000000000008152506000908051906020019061005c929190610062565b50610107565b828054600181600116156101000203166002900490600052602060002090601f016020900481019282601f106100a357805160ff19168380011785556100d1565b828001600101855582156100d1579182015b828111156100d05782518255916020019190600101906100b5565b5b5090506100de91906100e2565b5090565b61010491905b808211156101005760008160009055506001016100e8565b5090565b90565b610310806101166000396000f3fe608060405234801561001057600080fd5b50600436106100365760003560e01c80634ed3885e1461003b5780636d4ce63c146100f6575b600080fd5b6100f46004803603602081101561005157600080fd5b810190808035906020019064010000000081111561006e57600080fd5b82018360208201111561008057600080fd5b803590602001918460018302840111640100000000831117156100a257600080fd5b91908080601f016020809104026020016040519081016040528093929190818152602001838380828437600081840152601f19601f820116905080830192505050505050509192919290505050610179565b005b6100fe610193565b6040518080602001828103825283818151815260200191508051906020019080838360005b8381101561013e578082015181840152602081019050610123565b50505050905090810190601f16801561016b5780820380516001836020036101000a031916815260200191505b509250505060405180910390f35b806000908051906020019061018f929190610235565b5050565b606060008054600181600116156101000203166002900480601f01602080910402602001604051908101604052809291908181526020018280546001816001161561010002031660029004801561022b5780601f106102005761010080835404028352916020019161022b565b820191906000526020600020905b81548152906001019060200180831161020e57829003601f168201915b5050505050905090565b828054600181600116156101000203166002900490600052602060002090601f016020900481019282601f1061027657805160ff19168380011785556102a4565b828001600101855582156102a4579182015b828111156102a3578251825591602001919060010190610288565b5b5090506102b191906102b5565b5090565b6102d791905b808211156102d35760008160009055506001016102bb565b5090565b9056fea2646970667358221220ae145a10176f42f6981acd9071f3ce342e500560ed04a1ef7f04e34c08930eb064736f6c634300060a0033"

// DeployHelloWorld deploys a new kvtabletest, binding an instance of HelloWorld to it.
func DeployHelloWorld(auth *bind.TransactOpts, backend bind.ContractBackend) (common.Address, *types.Transaction, *HelloWorld, error) {
	parsed, err := abi.JSON(strings.NewReader(HelloWorldABI))
	if err != nil {
		return common.Address{}, nil, nil, err
	}

	address, tx, contract, err := bind.DeployContract(auth, parsed, common.FromHex(HelloWorldBin), backend)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &HelloWorld{HelloWorldCaller: HelloWorldCaller{contract: contract}, HelloWorldTransactor: HelloWorldTransactor{contract: contract}, HelloWorldFilterer: HelloWorldFilterer{contract: contract}}, nil
}

func AsyncDeployHelloWorld(auth *bind.TransactOpts, handler func(*types.Receipt, error), backend bind.ContractBackend) (*types.Transaction, error) {
	parsed, err := abi.JSON(strings.NewReader(HelloWorldABI))
	if err != nil {
		return nil, err
	}

	tx, err := bind.AsyncDeployContract(auth, handler, parsed, common.FromHex(HelloWorldBin), backend)
	if err != nil {
		return nil, err
	}
	return tx, nil
}

// HelloWorld is an auto generated Go binding around a Solidity kvtabletest.
type HelloWorld struct {
	HelloWorldCaller     // Read-only binding to the kvtabletest
	HelloWorldTransactor // Write-only binding to the kvtabletest
	HelloWorldFilterer   // Log filterer for kvtabletest events
}

// HelloWorldCaller is an auto generated read-only Go binding around a Solidity kvtabletest.
type HelloWorldCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// HelloWorldTransactor is an auto generated write-only Go binding around a Solidity kvtabletest.
type HelloWorldTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// HelloWorldFilterer is an auto generated log filtering Go binding around a Solidity kvtabletest events.
type HelloWorldFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// HelloWorldSession is an auto generated Go binding around a Solidity kvtabletest,
// with pre-set call and transact options.
type HelloWorldSession struct {
	Contract     *HelloWorld       // Generic kvtabletest binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// HelloWorldCallerSession is an auto generated read-only Go binding around a Solidity kvtabletest,
// with pre-set call options.
type HelloWorldCallerSession struct {
	Contract *HelloWorldCaller // Generic kvtabletest caller binding to set the session for
	CallOpts bind.CallOpts     // Call options to use throughout this session
}

// HelloWorldTransactorSession is an auto generated write-only Go binding around a Solidity kvtabletest,
// with pre-set transact options.
type HelloWorldTransactorSession struct {
	Contract     *HelloWorldTransactor // Generic kvtabletest transactor binding to set the session for
	TransactOpts bind.TransactOpts     // Transaction auth options to use throughout this session
}

// HelloWorldRaw is an auto generated low-level Go binding around a Solidity kvtabletest.
type HelloWorldRaw struct {
	Contract *HelloWorld // Generic kvtabletest binding to access the raw methods on
}

// HelloWorldCallerRaw is an auto generated low-level read-only Go binding around a Solidity kvtabletest.
type HelloWorldCallerRaw struct {
	Contract *HelloWorldCaller // Generic read-only kvtabletest binding to access the raw methods on
}

// HelloWorldTransactorRaw is an auto generated low-level write-only Go binding around a Solidity kvtabletest.
type HelloWorldTransactorRaw struct {
	Contract *HelloWorldTransactor // Generic write-only kvtabletest binding to access the raw methods on
}

// NewHelloWorld creates a new instance of HelloWorld, bound to a specific deployed kvtabletest.
func NewHelloWorld(address common.Address, backend bind.ContractBackend) (*HelloWorld, error) {
	contract, err := bindHelloWorld(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &HelloWorld{HelloWorldCaller: HelloWorldCaller{contract: contract}, HelloWorldTransactor: HelloWorldTransactor{contract: contract}, HelloWorldFilterer: HelloWorldFilterer{contract: contract}}, nil
}

// NewHelloWorldCaller creates a new read-only instance of HelloWorld, bound to a specific deployed kvtabletest.
func NewHelloWorldCaller(address common.Address, caller bind.ContractCaller) (*HelloWorldCaller, error) {
	contract, err := bindHelloWorld(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &HelloWorldCaller{contract: contract}, nil
}

// NewHelloWorldTransactor creates a new write-only instance of HelloWorld, bound to a specific deployed kvtabletest.
func NewHelloWorldTransactor(address common.Address, transactor bind.ContractTransactor) (*HelloWorldTransactor, error) {
	contract, err := bindHelloWorld(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &HelloWorldTransactor{contract: contract}, nil
}

// NewHelloWorldFilterer creates a new log filterer instance of HelloWorld, bound to a specific deployed kvtabletest.
func NewHelloWorldFilterer(address common.Address, filterer bind.ContractFilterer) (*HelloWorldFilterer, error) {
	contract, err := bindHelloWorld(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &HelloWorldFilterer{contract: contract}, nil
}

// bindHelloWorld binds a generic wrapper to an already deployed kvtabletest.
func bindHelloWorld(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(HelloWorldABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) kvtabletest method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_HelloWorld *HelloWorldRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _HelloWorld.Contract.HelloWorldCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the kvtabletest, calling
// its default method if one is available.
func (_HelloWorld *HelloWorldRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, *types.Receipt, error) {
	return _HelloWorld.Contract.HelloWorldTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) kvtabletest method with params as input values.
func (_HelloWorld *HelloWorldRaw) TransactWithResult(opts *bind.TransactOpts, result interface{}, method string, params ...interface{}) (*types.Transaction, *types.Receipt, error) {
	return _HelloWorld.Contract.HelloWorldTransactor.contract.TransactWithResult(opts, result, method, params...)
}

// Call invokes the (constant) kvtabletest method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_HelloWorld *HelloWorldCallerRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _HelloWorld.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the kvtabletest, calling
// its default method if one is available.
func (_HelloWorld *HelloWorldTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, *types.Receipt, error) {
	return _HelloWorld.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) kvtabletest method with params as input values.
func (_HelloWorld *HelloWorldTransactorRaw) TransactWithResult(opts *bind.TransactOpts, result interface{}, method string, params ...interface{}) (*types.Transaction, *types.Receipt, error) {
	return _HelloWorld.Contract.contract.TransactWithResult(opts, result, method, params...)
}

// Get is a free data retrieval call binding the kvtabletest method 0x6d4ce63c.
//
// Solidity: function get() constant returns(string)
func (_HelloWorld *HelloWorldCaller) Get(opts *bind.CallOpts) (string, error) {
	var (
		ret0 = new(string)
	)
	out := ret0
	err := _HelloWorld.contract.Call(opts, out, "get")
	return *ret0, err
}

// Get is a free data retrieval call binding the kvtabletest method 0x6d4ce63c.
//
// Solidity: function get() constant returns(string)
func (_HelloWorld *HelloWorldSession) Get() (string, error) {
	return _HelloWorld.Contract.Get(&_HelloWorld.CallOpts)
}

// Get is a free data retrieval call binding the kvtabletest method 0x6d4ce63c.
//
// Solidity: function get() constant returns(string)
func (_HelloWorld *HelloWorldCallerSession) Get() (string, error) {
	return _HelloWorld.Contract.Get(&_HelloWorld.CallOpts)
}

// Set is a paid mutator transaction binding the kvtabletest method 0x4ed3885e.
//
// Solidity: function set(string v) returns()
func (_HelloWorld *HelloWorldTransactor) Set(opts *bind.TransactOpts, v string) (*types.Transaction, *types.Receipt, error) {
	var ()
	out := &[]interface{}{}
	transaction, receipt, err := _HelloWorld.contract.TransactWithResult(opts, out, "set", v)
	return transaction, receipt, err
}

func (_HelloWorld *HelloWorldTransactor) AsyncSet(handler func(*types.Receipt, error), opts *bind.TransactOpts, v string) (*types.Transaction, error) {
	return _HelloWorld.contract.AsyncTransact(opts, handler, "set", v)
}

// Set is a paid mutator transaction binding the kvtabletest method 0x4ed3885e.
//
// Solidity: function set(string v) returns()
func (_HelloWorld *HelloWorldSession) Set(v string) (*types.Transaction, *types.Receipt, error) {
	return _HelloWorld.Contract.Set(&_HelloWorld.TransactOpts, v)
}

func (_HelloWorld *HelloWorldSession) AsyncSet(handler func(*types.Receipt, error), v string) (*types.Transaction, error) {
	return _HelloWorld.Contract.AsyncSet(handler, &_HelloWorld.TransactOpts, v)
}

// Set is a paid mutator transaction binding the kvtabletest method 0x4ed3885e.
//
// Solidity: function set(string v) returns()
func (_HelloWorld *HelloWorldTransactorSession) Set(v string) (*types.Transaction, *types.Receipt, error) {
	return _HelloWorld.Contract.Set(&_HelloWorld.TransactOpts, v)
}

func (_HelloWorld *HelloWorldTransactorSession) AsyncSet(handler func(*types.Receipt, error), v string) (*types.Transaction, error) {
	return _HelloWorld.Contract.AsyncSet(handler, &_HelloWorld.TransactOpts, v)
}

package replay

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

type Tracer interface {

	// Init
	BlockInit()
	TxInit(seq int)

	// Transfer
	AddTransfer(from string, to string, balance *big.Int, reason string)
	SetTransferSeq()
	ValidateTransfer(seq int)

	// Refund
	SetRefundGas(refund *big.Int)

	// Account Management
	AddBlockCreatedAccount(addr common.Address)
	AddTxCreatedAccount(addr common.Address)
	GetTxCreatedAccounts() []common.Address
	GetCreatedAccounts() map[string]int

	SetCreatedAddress(addr common.Address)
	SetSuicidedAddress(seq int, addr common.Address)
	GetSuicidedAccounts() map[int]string
	IsCreated(addr common.Address) bool

	// Code Management
	AddCode(addr common.Address, code []byte)

	// To string/Dump
	//StrTxCreatedAccounts() string
	StrBlockCreatedAccounts() string
	StrTransaction() string
	StrTransfer(seq int) string
	StrTxBasic(nonce, gasPrice, startGas, gasUsed uint64, data, init, to, value string, v int64, r, s, hash string)

	// Failure
	FailCreate(startPos int, err error)
	FailTx(err error)
	Fail(seq int, err error, rollback bool)
	GetTxFailReason() string

	// State
	SetInputAccount(addr common.Address, codeHash common.Hash, balance *big.Int, nonce uint64)
	SetOutputAccount(addr common.Address, codeHash common.Hash, balance *big.Int, nonce uint64)
	GetInputAccounts() []string

	// Simulate environment
	GetDepth() int
	EnterEnv(addr common.Address)
	ExitEnv()
	RevertStore()

	ResizeMemory(size int)
	MemLen() int
	SetMemory(seq int)
	Pop() Content
	Top() Content
	Push(c Content)
	Swap(n int)
	Dup(n int)
	Loc(id int) Content
	StackLen() int

	// Precompiled
	IsPrecompiled(addr common.Address) bool
	SetPrecompiledMemIO(seq int)
	SetPrecompiled(addr common.Address, gas, err int, input, output []byte)

	// Trace
	ScanTrace()
	GetSeq() int
	InitTrace(name string, pc int, gas int, calldepth int) int
	AddMemSize(seq int, before, after int)
	AddGas(seq int, gas int64)
	SetStackInput(seq int, num int)
	SetStackOutput(seq int, val *big.Int)
	SetMemoryInput(seq int, offset *big.Int, size *big.Int)
	SetMemoryOutput(seq int, offset *big.Int, size *big.Int, val []byte)
	AddStackInput(idx int, argN int)
	SetCreated()
}

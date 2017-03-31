package replay

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

type DefaultTracer struct{}

// Init
func (r *DefaultTracer) BlockInit()     {}
func (r *DefaultTracer) TxInit(seq int) {}

// Transfer
func (r *DefaultTracer) AddTransfer(from common.Address, to common.Address, balance *big.Int, reason string) {
}
func (r *DefaultTracer) SetTransferSeq()          {}
func (r *DefaultTracer) ValidateTransfer(seq int) {}

// Refund
func (r *DefaultTracer) SetRefundGas(refund *big.Int) {}

// Account Management
func (r *DefaultTracer) AddBlockCreatedAccount(addr common.Address) {}
func (r *DefaultTracer) AddTxCreatedAccount(addr common.Address)    {}
func (r *DefaultTracer) GetTxCreatedAccounts() []common.Address {
	return []common.Address{}
}
func (r *DefaultTracer) GetCreatedAccounts() map[string]int {
	return map[string]int{}
}

func (r *DefaultTracer) SetCreatedAddress(addr common.Address) {

}
func (r *DefaultTracer) SetSuicidedAddress(seq int, addr common.Address) {

}

func (r *DefaultTracer) GetSuicidedAccounts() map[int]string {
	return map[int]string{}
}

func (r *DefaultTracer) IsCreated(addr common.Address) bool {
	return false
}

// Code Management
func (r *DefaultTracer) AddCode(addr common.Address, code []byte) {}

// To string/Dump
func (r *DefaultTracer) StrBlockCreatedAccounts() string {
	return ""
}

func (r *DefaultTracer) StrTransaction() string {
	return ""
}

func (r *DefaultTracer) StrTransfer(seq int) string {
	return ""
}

func (r *DefaultTracer) StrTxBasic(nonce, gasPrice, startGas, gasUsed uint64, data, init, to, value string, v int64, r_, s, hash string) {

}

// Failure
func (r *DefaultTracer) FailCreate(startPos int, err error)     {}
func (r *DefaultTracer) FailTx(err error)                       {}
func (r *DefaultTracer) Fail(seq int, err error, rollback bool) {}
func (r *DefaultTracer) GetTxFailReason() string {
	return ""
}

// State
func (r *DefaultTracer) SetInputAccount(addr common.Address, codeHash common.Hash, balance *big.Int, nonce uint64) {
}
func (r *DefaultTracer) SetOutputAccount(addr common.Address, codeHash common.Hash, balance *big.Int, nonce uint64) {
}
func (r *DefaultTracer) GetInputAccounts() []string {
	return []string{}
}

// Simulate environment
func (r *DefaultTracer) GetDepth() int {
	return 0
}
func (r *DefaultTracer) EnterEnv(addr common.Address) {}
func (r *DefaultTracer) ExitEnv()                     {}
func (r *DefaultTracer) RevertStore()                 {}

func (r *DefaultTracer) ResizeMemory(size int) {}
func (r *DefaultTracer) MemLen() int {
	return 0
}
func (r *DefaultTracer) SetMemory(seq int) {}
func (r *DefaultTracer) Pop() Content {
	return Content{}
}
func (r *DefaultTracer) Top() Content {
	return Content{}
}
func (r *DefaultTracer) Push(c Content) {}
func (r *DefaultTracer) Swap(n int)     {}
func (r *DefaultTracer) Dup(n int)      {}
func (r *DefaultTracer) Loc(id int) Content {
	return Content{}
}
func (r *DefaultTracer) StackLen() int {
	return 0
}

// Precompiled

func (r *DefaultTracer) IsPrecompiled(addr common.Address) bool {
	return false
}
func (r *DefaultTracer) SetPrecompiledMemIO(seq int)                                            {}
func (r *DefaultTracer) SetPrecompiled(addr common.Address, gas, err int, input, output []byte) {}

// Trace

func (r *DefaultTracer) ScanTrace() {}
func (r *DefaultTracer) GetSeq() int {
	return 0
}
func (r *DefaultTracer) InitTrace(name string, pc int, gas int, calldepth int) int {
	return 0
}
func (r *DefaultTracer) AddMemSize(seq int, before, after int)                               {}
func (r *DefaultTracer) AddGas(seq int, gas *big.Int)                                        {}
func (r *DefaultTracer) SetStackInput(seq int, num int)                                      {}
func (r *DefaultTracer) SetStackOutput(seq int, val *big.Int)                                {}
func (r *DefaultTracer) SetMemoryInput(seq int, offset *big.Int, size *big.Int)              {}
func (r *DefaultTracer) SetMemoryOutput(seq int, offset *big.Int, size *big.Int, val []byte) {}
func (r *DefaultTracer) AddStackInput(idx int, argN int)                                     {}
func (r *DefaultTracer) SetCreated()                                                         {}

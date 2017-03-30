package replay

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

type tracer interface {

	// Transfer
	AddTransfer(from common.Address, to common.Address, balance *big.Int, reason string)
	SetTransferSeq()

	// Refund
	SetRefundGas(refund *big.Int)

	// Gas
	AddGas(seq int, gas *big.Int)

	// Account Management
	AddBlockCreatedAccount(addr common.Address)
	AddTxCreatedAccount(addr common.Address)
	GetTxCreatedAccounts() map[string]int
	SetSuicidedAddress(seq int, addr common.Address)

	// Code Management
	AddCode(addr common.Address, code []byte)

	// To string/Dump
	StrTxCreatedAccounts() string

	// Failure
	FailCreate(startPos int, err error)
	FailTx(err error)
	Fail(seq int, err error, rollback bool)

	// InputState
	SetInputAccount(addr common.Address, codeHash common.Hash, balance *big.Int, nonce uint64)

	// Simulate environment
	EnterEnv(addr common.Address)
	ExitEnv()
	RevertStore()

	SetStackInput(seq int, num int)
	SetStackOutput(seq int, val *big.Int)
	SetMemoryInput(seq int, offset *big.Int, size *big.Int)
	SetMemoryOutput(seq int, offset *big.Int, size *big.Int, val []byte)

	// Precompiled
	IsPrecompiled(addr common.Address)
	SetPrecompiledMemIO(seq int)
}

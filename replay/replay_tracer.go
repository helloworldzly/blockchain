package replay

import (
	"encoding/json"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

type context map[string]Content
type storage map[string]context
type stack []Content
type memory []Content

type accountState struct {
	Address        string
	beforeCodeHash string
	afterCodeHash  string
	beforeNonce    int
	afterNonce     int
	beforeBalance  string
	afterBalance   string
	isSmart        int
}

type (
	stringStore  map[string]map[string]string
	contentStore map[string]map[string]Content
)

type Trans struct {
	From  string
	To    string
	Value string
	Type  string
	TxID  int
	TxSeq int
	Valid int
}

type ReplayTracer struct {
	txMem                []memory
	txStack              []stack
	txStore              []storage
	txAddr               []string
	pAddr                string
	pStore               storage
	pStack               stack
	pMem                 memory
	depth                int
	txFailReason         string
	txCreatedAccounts    []common.Address
	codeStore            map[string]string
	ioAccounts           map[string]accountState
	refundGas            int
	blockCreatedAccounts []string
	traceLog             []Trace
	storeIn              stringStore
	storeOut             contentStore
	createdAccounts      map[string]int
	suicidedTrace        map[int]string
	TxID                 int
	Miner                string
	externalCallFlag     int
	transList            []Trans
	TxBasic              string
	strTx                []string
}

// Init
func (r *ReplayTracer) BlockInit() {
	r.transList = []Trans{}
	r.TxID = -1
	r.strTx = []string{}
	r.blockCreatedAccounts = []string{}
}

func (r *ReplayTracer) TxInit(seq int) {
	r.TxBasic = ""
	r.TxID = seq
	r.traceLog = make([]Trace, 0)
	r.storeIn = make(stringStore)
	r.storeOut = make(contentStore)
	r.createdAccounts = make(map[string]int)
	r.suicidedTrace = make(map[int]string)

	r.txFailReason = ""
	r.txCreatedAccounts = make([]common.Address, 0)

	r.ioAccounts = make(map[string]accountState)
	r.codeStore = make(map[string]string)

	r.refundGas = 0

	r.externalCallFlag = 0
}

// Transfer
func (r *ReplayTracer) AddTransfer(from common.Address, to common.Address, balance *big.Int, reason string) {
	obj := Trans{
		From:  from.Hex(),
		To:    to.Hex(),
		Value: balance.String(),
		Type:  reason,
		TxID:  r.TxID,
		TxSeq: -1,
		Valid: 0,
	}

	if reason == "MinerReward" || reason == "UncleReward" {
		obj.TxID = -1
	}

	if reason == "CallTransfer" && r.externalCallFlag == 0 {
		obj.Type = "ExternalCallTransfer"
		r.externalCallFlag = 1
	}
	r.transList = append(r.transList, obj)

}

func (r *ReplayTracer) SetTransferSeq() {
	r.transList[len(r.transList)-1].TxSeq = len(r.traceLog) - 1
}

func (r *ReplayTracer) ValidateTransfer(seq int) {
	for i, v := range r.transList {
		if v.TxID != seq {
			continue
		}

		if v.Type == "ExternalCallTransfer" {
			if r.txFailReason != "" {
				r.transList[i].Valid = -1
			}
		}

		if v.TxSeq != -1 && r.traceLog[v.TxSeq].Basic.Reverted == 1 {
			r.transList[i].Valid = -1
		}
	}
}

// Refund
func (r *ReplayTracer) SetRefundGas(refund *big.Int) {
	r.refundGas = int(refund.Uint64())
}

// Account Management
func (r *ReplayTracer) AddBlockCreatedAccount(addr common.Address) {
	r.blockCreatedAccounts = append(r.blockCreatedAccounts, "\""+addr.Hex()+"\"")
}

func (r *ReplayTracer) AddTxCreatedAccount(addr common.Address) {
	r.txCreatedAccounts = append(r.txCreatedAccounts, addr)
}

func (r *ReplayTracer) GetTxCreatedAccounts() []common.Address {
	return r.txCreatedAccounts
}

func (r *ReplayTracer) GetCreatedAccounts() map[string]int {
	return r.createdAccounts
}

func (r *ReplayTracer) SetCreatedAddress(addr common.Address) {
	r.createdAccounts[addr.Hex()] = 1
}

func (r *ReplayTracer) SetSuicidedAddress(seq int, addr common.Address) {
	r.suicidedTrace[seq] = addr.Hex()
}

func (r *ReplayTracer) GetSuicidedAccounts() map[int]string {
	return r.suicidedTrace
}

func (r *ReplayTracer) IsCreated(addr common.Address) bool {
	if _, ok := r.createdAccounts[addr.Hex()]; ok {
		return true
	}
	return false
}

// Code Management
func (r *ReplayTracer) AddCode(addr common.Address, code []byte) {
	if string(code) != "" && string(code) != "0x" {
		r.codeStore[addr.Hex()] = string(code)
	}
}

// To string/Dump
/*
func (r *ReplayTracer) StrTxCreatedAccounts() string {
	return ""
}
*/

func (r *ReplayTracer) StrBlockCreatedAccounts() string {
	var (
		fstr string
	)

	for _, v := range r.blockCreatedAccounts {
		if fstr != "" {
			fstr += ","
		}
		fstr += v
	}

	return fstr
}

func (r *ReplayTracer) StrTransaction() string {
	fstr := "{" + r.TxBasic
	fstr = fstr + "," + fmt.Sprintf("\"instructionCount\":%d", len(r.traceLog))
	if r.refundGas != 0 {
		fstr = fstr + "," + fmt.Sprintf("\"totalRefund\":%d", r.refundGas)
	}
	fstr = fstr + "," + fmt.Sprintf("\"inputStates\":%s", r.strInputState())
	fstr = fstr + "," + fmt.Sprintf("\"outputStates\":%s", r.strOutputState())
	fstr = fstr + "," + fmt.Sprintf("\"failReason\":\"%s\"", r.txFailReason)
	fstr = fstr + "," + r.strCreateSuicide()

	return fstr + ",\"codeStorage\":" + r.strCode() + ",\"codeTrace\":" + r.strTrace() + "}"
}

func (r *ReplayTracer) StrTransfer(seq int) string {
	fstr := ""
	for _, v := range r.transList {
		//fmt.Printf("[DEBUGING] %s %d %d %d\n", v.Type, v.TxID, v.TxSeq, v.Valid)
		if v.Type == "MinerReward" && seq == 0 {
			continue
		}
		if v.Valid == -1 {
			continue
		}
		if fstr != "" {
			fstr += ","
		}
		tmp, _ := json.Marshal(v)
		fstr = fstr + fmt.Sprintf("%s", tmp)
	}
	return "[" + fstr + "]"
}

func (r *ReplayTracer) StrTxBasic(nonce, gasPrice, startGas, gasUsed uint64, data, init, to, value string, v int64, r_, s, hash string) {
	if r.TxBasic != "" {
		r.TxBasic += ","
	}
	r.TxBasic += fmt.Sprintf("\"nonce\":%d,\"gasPrice\":%d,\"startGas\":%d,\"gasUsed\":%d,\"data\":\"%s\",\"init\":\"%s\",\"to\":\"%s\",\"value\":\"%s\",\"v\":%d,\"r\":\"%s\",\"s\":\"%s\",\"hash\":\"%s\"",
		nonce, gasPrice, startGas, gasUsed, data, init, to, value, v, r_, s, hash)
}

// Failure
func (r *ReplayTracer) FailCreate(startPoint int, err error) {
	var (
		endPoint = r.GetSeq()
		failInfo = errStr(err)
	)

	for i := startPoint; i <= endPoint; i++ {
		r.traceLog[i].Basic.Reverted = 1
		r.traceLog[i].Basic.FailInfo = failInfo
	}
}

func (r *ReplayTracer) FailTx(err error) {
	var (
		failInfo = errStr(err)
	)
	r.txFailReason = failInfo
}

func (r *ReplayTracer) Fail(seq int, err error, rollback bool) {
	var (
		failDepth = r.traceLog[seq].Basic.CallDepth
		failInfo  = errStr(err)
	)
	if rollback {
		for i := seq; i >= 0; i-- {
			if r.traceLog[i].Basic.CallDepth < failDepth {
				break
			}
			r.traceLog[i].Basic.Reverted = 1
			r.traceLog[i].Basic.FailInfo = failInfo
		}
		r.traceLog = r.traceLog[:len(r.traceLog)-1]
	} else {
		r.traceLog[seq].Basic.Reverted = 1
		r.traceLog[seq].Basic.FailInfo = failInfo
	}
	if len(r.traceLog) > 0 {
		r.traceLog[len(r.traceLog)-1].Basic.ExceptionTag = 1
	}

}

func (r *ReplayTracer) GetTxFailReason() string {
	return r.txFailReason
}

// State
func (r *ReplayTracer) SetInputAccount(addr common.Address, codeHash common.Hash, balance *big.Int, nonce uint64) {
	if _, ok := r.ioAccounts[addr.Hex()]; !ok {
		r.ioAccounts[addr.Hex()] = accountState{
			Address:        addr.Hex(),
			beforeCodeHash: codeHash.Hex(),
			beforeBalance:  balance.String(),
			beforeNonce:    int(nonce),
		}
	}

}

func (r *ReplayTracer) SetOutputAccount(addr common.Address, codeHash common.Hash, balance *big.Int, nonce uint64) {
	if _, ok := r.ioAccounts[addr.Hex()]; !ok {
		r.ioAccounts[addr.Hex()] = accountState{
			Address:       addr.Hex(),
			beforeBalance: "invalid",
			afterCodeHash: codeHash.Hex(),
			afterBalance:  balance.String(),
			afterNonce:    int(nonce),
		}
	}

	tmp := r.ioAccounts[addr.Hex()]
	tmp.afterBalance = balance.String()
	tmp.afterNonce = int(nonce)
	tmp.afterCodeHash = codeHash.Hex()
	r.ioAccounts[addr.Hex()] = tmp
}

func (r *ReplayTracer) GetInputAccounts() []string {
	tmp := []string{}
	for k := range r.ioAccounts {
		tmp = append(tmp, k)
	}
	return tmp
}

// Simulate environment
func (r *ReplayTracer) GetDepth() int {
	return r.depth
}

func (r *ReplayTracer) EnterEnv(addr common.Address) {
	if r.depth != 0 {
		r.txStack = append(r.txStack, r.pStack)
		r.txMem = append(r.txMem, r.pMem)
		r.txAddr = append(r.txAddr, r.pAddr)
		r.txStore = append(r.txStore, r.pStore)
	}
	r.pAddr = addr.Hex()
	r.pStore = storage{}
	r.pStore[r.pAddr] = make(map[string]Content)
	r.pStack = stack{}
	r.pMem = memory{}
	r.depth++
}

func (r *ReplayTracer) ExitEnv() {
	if r.depth != 1 {
		r.pAddr = r.txAddr[len(r.txAddr)-1]
		r.txAddr = r.txAddr[:len(r.txAddr)-1]

		r.pMem = r.txMem[len(r.txMem)-1]
		r.txMem = r.txMem[:len(r.txMem)-1]

		r.pStack = r.txStack[len(r.txStack)-1]
		r.txStack = r.txStack[:len(r.txStack)-1]

		r.pStore = mergeStore(r.txStore[len(r.txStore)-1], r.pStore)
		r.txStore = r.txStore[:len(r.txStore)-1]
	}
	r.depth--
}

func (r *ReplayTracer) RevertStore() {
	r.pStore = storage{}
}

func (r *ReplayTracer) ResizeMemory(size int) {
	csize := len(r.pMem)
	if csize < size {
		r.pMem = append(r.pMem, make([]Content, size-csize)...)
		for i := csize; i < size; i++ {
			r.pMem[i].Size = 1
			r.pMem[i].Val = []byte{byte(0)}
		}
	}
}

func (r *ReplayTracer) MemLen() int {
	return len(r.pMem)
}

func (r *ReplayTracer) SetMemory(seq int) {
	switch r.traceLog[seq].Basic.OpName {
	case "MSTORE":
		addr := int(common.Bytes2Big(r.traceLog[seq].StackInput.Val[0].Val).Int64())
		val := r.traceLog[seq].StackInput.Val[1].Val
		// Must extend
		val = common.BigToBytes(common.Bytes2Big(val), 256)
		for i := addr; i < addr+32; i++ {
			r.pMem[i] = Content{
				Val:  []byte{val[i-addr]},
				Size: 1,
				Src:  []Source{r.traceLog[seq].StackInput.Val[1].Src[0]}, // Should has only 1 source indeed, should check
			}
			r.pMem[i].Src[0].Offset = i - addr
		}
	case "MSTORE8":
		addr := int(common.Bytes2Big(r.traceLog[seq].StackInput.Val[0].Val).Int64())
		val := r.traceLog[seq].StackInput.Val[1]
		r.pMem[addr] = val
	default:
	}
}

func (r *ReplayTracer) Pop() Content {
	c := r.pStack[len(r.pStack)-1]
	r.pStack = r.pStack[:len(r.pStack)-1]
	return c
}

func (r *ReplayTracer) Top() Content {
	c := r.pStack[len(r.pStack)-1]
	return c
}

func (r *ReplayTracer) Push(c Content) {
	r.pStack = append(r.pStack, c)
}

func (r *ReplayTracer) Swap(n int) {
	pLen := len(r.pStack)
	r.pStack[pLen-n], r.pStack[pLen-1] = r.pStack[pLen-1], r.pStack[pLen-n]
}

func (r *ReplayTracer) Dup(n int) {
	r.Push(r.pStack[len(r.pStack)-n])
}

func (r *ReplayTracer) Loc(id int) Content {
	return r.pStack[len(r.pStack)-id]
}

func (r *ReplayTracer) StackLen() int {
	return len(r.pStack)
}

// Precompiled
func (r *ReplayTracer) IsPrecompiled(addr common.Address) bool {
	if addr.Hex() == string(common.LeftPadBytes([]byte{1}, 20)) {
		return true
	}
	if addr.Hex() == string(common.LeftPadBytes([]byte{2}, 20)) {
		return true
	}
	if addr.Hex() == string(common.LeftPadBytes([]byte{3}, 20)) {
		return true
	}
	if addr.Hex() == string(common.LeftPadBytes([]byte{4}, 20)) {
		return true
	}
	return false
}

func (r *ReplayTracer) SetPrecompiledMemIO(seq int) {
	r.traceLog[seq+1].MemoryInput = r.traceLog[seq].MemoryInput
	r.traceLog[seq+1].MemoryOutput = r.traceLog[seq].MemoryOutput
	if r.traceLog[seq+1].Basic.OpName == "V-IDENTITY" {
		r.traceLog[seq+1].MemoryOutput = r.traceLog[seq].MemoryInput
		r.traceLog[seq].MemoryOutput = r.traceLog[seq].MemoryInput
	}
}

func (r *ReplayTracer) SetPrecompiled(addr common.Address, gas, err int, input, output []byte) {

	var (
		name string
		seq  int
	)

	switch addr.Str() {
	case string(common.LeftPadBytes([]byte{1}, 20)):
		name = "V-ECREC"
	case string(common.LeftPadBytes([]byte{2}, 20)):
		name = "V-SHA256"
	case string(common.LeftPadBytes([]byte{3}, 20)):
		name = "V-RIP160"
	case string(common.LeftPadBytes([]byte{4}, 20)):
		name = "V-IDENTITY"
	default:
		return
	}

	seq = r.GetSeq()
	basic := &OpStat{
		OpName:        name,
		PC:            0,
		GasUsed:       gas,
		AccountAddr:   addr.Hex(),
		Reverted:      err,
		BeforeMemSize: -1,
		AfterMemSize:  -1,
	}
	if basic.Reverted == 1 {
		basic.FailInfo = "OutOfGas"
		basic.ExceptionTag = 1
	}

	if seq == -1 {
		// We are directly called
		basic.CallDepth = 0
		basic.TraceSeq = 0
		temp := Trace{
			Basic: basic,
			MemoryInput: &MemArray{
				Offset: 0,
				Val: Content{
					Val:  input,
					Size: len(input),
					Src: []Source{
						Source{
							Type:   "environment",
							Opcode: name,
						},
					},
				},
			},
			MemoryOutput: &MemArray{
				Offset: 0,
				Val: Content{
					Val:  output,
					Size: len(output),
					Src: []Source{
						Source{
							Type:   "computed",
							Opcode: name,
						},
					},
				},
			},
		}
		if name == "V-IDENTITY" {
			temp.MemoryOutput = temp.MemoryInput
		}
		r.traceLog = append(r.traceLog, temp)
	} else {
		// Called in code
		basic.CallDepth = r.depth
		basic.TraceSeq = r.GetSeq() + 1
		temp := Trace{
			Basic: basic,
			// Left the Mem I/O in the opCall to set
		}
		r.traceLog = append(r.traceLog, temp)
	}

}

// Trace
func (r *ReplayTracer) ScanTrace() {
	if r.txFailReason != "" {
		for id := range r.traceLog {
			r.traceLog[id].Basic.FailInfo = r.txFailReason
			r.traceLog[id].Basic.Reverted = 1
		}
	} else {
		for _, addr := range r.GetTxCreatedAccounts() {
			r.createdAccounts[addr.Hex()] = 1
		}
	}
	for id, log := range r.traceLog {
		if log.Basic.Reverted == 1 && log.Basic.OpName != "SLOAD" {
			continue
		}
		opName := log.Basic.OpName
		if opName == "SLOAD" {
			if log.StackOutput.Val[0].Src[0].OriginInstructionSeq != id {
				continue
			}
			addr := log.Basic.AccountAddr
			key := common.Bytes2Hex(common.LeftPadBytes(log.StackInput.Val[0].Val, 32))
			val := common.Bytes2Hex(common.LeftPadBytes(log.StackOutput.Val[0].Val, 32))
			if !r.hasStore(addr, key) {
				r.setIn(addr, key, val)
			}
		} else if opName == "SSTORE" {
			addr := log.Basic.AccountAddr
			key := common.Bytes2Hex(common.LeftPadBytes(log.StackInput.Val[0].Val, 32))
			val := log.StackInput.Val[1]
			r.setOut(addr, key, val)
		} else if opName == "CALL" {
			if log.Basic.IsCreatedNewAddress == 1 {
				addr := "0x" + common.Bytes2Hex(common.LeftPadBytes(log.StackInput.Val[1].Val, 20))
				r.createdAccounts[addr] = 1
			}
		} else if opName == "CREATE" {
			addr := "0x" + common.Bytes2Hex(common.LeftPadBytes(log.StackOutput.Val[0].Val, 20))
			if addr != "0x0000000000000000000000000000000000000000" {
				r.createdAccounts[addr] = 1
			}
		} else if opName == "SUICIDE" {
			if log.Basic.IsCreatedNewAddress == 1 {
				addr := "0x" + common.Bytes2Hex(common.LeftPadBytes(log.StackInput.Val[0].Val, 20))
				r.createdAccounts[addr] = 1
			}
		} else {
			continue
		}
	}
}

func (r *ReplayTracer) GetSeq() int {
	return len(r.traceLog) - 1
}

func (r *ReplayTracer) InitTrace(name string, pc int, gas int, calldepth int) int {
	basic := &OpStat{
		OpName:        name,
		PC:            pc,
		GasUsed:       gas,
		TraceSeq:      r.GetSeq() + 1,
		AccountAddr:   r.pAddr,
		CallDepth:     calldepth,
		BeforeMemSize: -1,
		AfterMemSize:  -1,
	}
	temp := Trace{
		Basic: basic,
	}

	r.traceLog = append(r.traceLog, temp)
	return r.GetSeq()
}

func (r *ReplayTracer) AddMemSize(seq int, before, after int) {
	r.traceLog[seq].Basic.BeforeMemSize = before
	r.traceLog[seq].Basic.AfterMemSize = after
}

func (r *ReplayTracer) AddGas(seq int, gas *big.Int) {
	r.refundGas = int(gas.Int64())
}

func (r *ReplayTracer) SetStackInput(seq int, num int) {
	if num > len(r.pStack) {
		fmt.Printf("%d > %d, Error...not matched in stack\n", num, len(r.pStack))
		return
	}
	stkArr := StackArray{}
	for i := 0; i < num; i++ {
		if r.traceLog[seq].Basic.OpName == "MSTORE8" && i == 1 {
			c := r.Pop()
			c.Val = []byte{byte(common.Bytes2Big(c.Val).Int64() & 0xff)}
			c.Size = 1
			stkArr.Add(c)
			continue
		}
		stkArr.Add(r.Pop())
	}
	r.traceLog[seq].StackInput = &stkArr

	// SSTORE takes effect here
	if r.traceLog[seq].Basic.OpName == "SSTORE" {
		r.saveStorage(common.Bytes2Hex(r.traceLog[seq].StackInput.Val[0].Val), r.traceLog[seq].StackInput.Val[1])
	}
}

func (r *ReplayTracer) SetStackOutput(seq int, val *big.Int) {
	stkArr := StackArray{}
	bArr := val.Bytes()
	blen := len(bArr)
	if blen == 0 {
		bArr = common.BigToBytes(val, 8)
		blen = len(bArr)
	}
	blen = 32
	c := Content{
		Val:  bArr,
		Size: blen,
	}
	stkArr.Add(c)
	r.traceLog[seq].StackOutput = &stkArr
	// Set source of stack output
	switch r.traceLog[seq].Basic.OpName {
	case "PUSH1", "PUSH2", "PUSH3", "PUSH4", "PUSH5", "PUSH6", "PUSH7", "PUSH8":
		fallthrough
	case "PUSH9", "PUSH10", "PUSH11", "PUSH12", "PUSH13", "PUSH14", "PUSH15", "PUSH16":
		fallthrough
	case "PUSH17", "PUSH18", "PUSH19", "PUSH20", "PUSH21", "PUSH22", "PUSH23", "PUSH24":
		fallthrough
	case "PUSH25", "PUSH26", "PUSH27", "PUSH28", "PUSH29", "PUSH30", "PUSH31", "PUSH32":
		r.setSource("code", seq)
	case "MLOAD", "SLOAD":
		r.setSource("", seq)
	case "PC", "MSIZE", "GAS":
		r.setSource("environment", seq)
	case "BLOCKHASH", "COINBASE", "TIMESTAMP", "NUMBER", "DIFFICULTY", "GASLIMIT":
		r.setSource("block", seq)
	case "ADDRESS", "BALANCE", "ORIGIN", "CALLER", "CALLVALUE", "CALLDATALOAD", "CALLDATASIZE", "CODESIZE", "GASPRICE", "EXTCODESIZE":
		r.setSource("environment", seq)
	case "ADD", "SUB", "MUL", "DIV", "SDIV", "MOD", "SMOD", "ADDMOD", "MULMOD", "EXP", "SIGNEXTEND":
		fallthrough
	case "LT", "GT", "SLT", "SGT", "EQ", "ISZERO", "AND", "OR", "XOR", "NOT", "BYTE":
		fallthrough
	case "SHA3":
		fallthrough
	case "CREATE", "CALL", "CALLCODE", "DELEGATECALL":
		r.setSource("computed", seq)
	case "SUICIDE":
		r.setSource("environment", seq)
	default:
	}
	// Push to tracer's internal stack
	for _, c := range r.traceLog[seq].StackOutput.Val {
		r.Push(c)
	}
}

func (r *ReplayTracer) SetMemoryInput(seq int, offset *big.Int, size *big.Int) {
	c := MemArray{}
	c.Offset = int(offset.Uint64())
	c.Val = r.getMemory(int(offset.Uint64()), int(size.Uint64()))
	r.traceLog[seq].MemoryInput = &c
}

func (r *ReplayTracer) SetMemoryOutput(seq int, offset *big.Int, size *big.Int, val []byte) {
	mstart := int(offset.Uint64())
	msize := int(size.Uint64())
	minSize := msize
	if len(val) < minSize {
		minSize = len(val)
	}

	if minSize == 0 {
		// Set MemoryOutput
		r.traceLog[seq].MemoryOutput = &MemArray{
			Offset: mstart,
			Val: Content{
				Val:  val[:minSize],
				Size: minSize,
				Src:  []Source{},
			},
		}
		return
	}

	s := Source{
		Type:                 "computed",
		Opcode:               r.traceLog[seq].Basic.OpName,
		OriginInstructionSeq: seq,
	}

	name := r.traceLog[seq].Basic.OpName

	if (name == "CALL" || name == "CALLCODE" || name == "CREATE" || name == "DELEGATECALL") && val != nil {
		if r.traceLog[r.GetSeq()].Basic.OpName == "RETURN" {
			if len(r.traceLog[r.GetSeq()].MemoryInput.Val.Src) == 0 {
				s = Source{
					Type: "invalid",
				}
			} else {
				s = r.traceLog[r.GetSeq()].MemoryInput.Val.Src[0]
			}
		}
	} else if name != "CODECOPY" && name != "EXTCODECOPY" && name != "CALLDATACOPY" {
		if r.traceLog[seq].StackInput.Len() != 0 {
			for _, cnt := range r.traceLog[seq].StackInput.Val {
				for _, w := range cnt.Src {
					s.OperandIDs = append(s.OperandIDs, w.OriginInstructionSeq)
				}
			}
		}
	}

	if name == "CODECOPY" {
		s.Type = "code"
		s.OperandIDs = append(s.OperandIDs, r.traceLog[seq].StackInput.Val[1].Src[0].OriginInstructionSeq)
		s.OperandIDs = append(s.OperandIDs, r.traceLog[seq].StackInput.Val[2].Src[0].OriginInstructionSeq)
	}

	if name == "EXTCODECOPY" {
		s.Type = "code"
		s.OperandIDs = append(s.OperandIDs, r.traceLog[seq].StackInput.Val[0].Src[0].OriginInstructionSeq)
		s.OperandIDs = append(s.OperandIDs, r.traceLog[seq].StackInput.Val[2].Src[0].OriginInstructionSeq)
		s.OperandIDs = append(s.OperandIDs, r.traceLog[seq].StackInput.Val[3].Src[0].OriginInstructionSeq)
	}

	if name == "CALLDATACOPY" {
		s.Type = "calldata"
		s.OperandIDs = append(s.OperandIDs, r.traceLog[seq].StackInput.Val[1].Src[0].OriginInstructionSeq)
		s.OperandIDs = append(s.OperandIDs, r.traceLog[seq].StackInput.Val[2].Src[0].OriginInstructionSeq)
	}

	// Write memory in tracer's internal memory
	for i := mstart; i < mstart+minSize; i++ {
		ns := Source(s)
		ns.Offset = i - mstart
		ns.Length = 1
		r.pMem[i] = Content{
			Val:  VarVal{val[i-mstart]},
			Size: 1,
			Src:  []Source{ns},
		}
	}
	// Set MemoryOutput
	r.traceLog[seq].MemoryOutput = &MemArray{
		Offset: mstart,
		Val: Content{
			Val:  val[:minSize],
			Size: minSize,
			Src:  []Source{s},
		},
	}
	// Clear invalid source
	if r.traceLog[seq].MemoryOutput.Val.Src[0].Type == "invalid" {
		r.traceLog[seq].MemoryOutput.Val.Src = []Source{}
	}

}

func (r *ReplayTracer) AddStackInput(seq int, argN int) {
	if r.traceLog[seq].StackInput == nil {
		r.traceLog[seq].StackInput = &StackArray{}
	}
	r.traceLog[seq].StackInput.Add(r.Loc(argN))
}

func (r *ReplayTracer) SetCreated() {
	r.traceLog[r.GetSeq()].Basic.IsCreatedNewAddress = 1
}

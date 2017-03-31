package replay

import (
	"encoding/hex"
	"encoding/json"
	"fmt"

	"github.com/ethereum/go-ethereum/common"
)

// String return string
func (s VarVal) String() string {
	return common.Bytes2Big(s).String()
}

// MarshalJSON help encode []byte
func (s VarVal) MarshalJSON() ([]byte, error) {
	return s, nil
}

func array2JSON(name string, arr []int) string {
	var substr, ext string
	if len(arr) == 0 {
		return ext
	}
	for i, w := range arr {
		if i != 0 {
			substr = substr + ","
		}
		substr += fmt.Sprintf("%d", w)
	}
	ext = fmt.Sprintf(",\"%s\":[%s]", name, substr)
	return ext
}

// MarshalJSON helps to turn source into meaningful JSON representation
func (t Source) MarshalJSON() ([]byte, error) {
	var fstr, s, ext string
	s = fmt.Sprintf("\"offset\":%d,\"len\":%d,\"outputOffset\":%d,\"srcType\":\"%s\",\"opcode\":\"%s\",\"instrSeq\":%d",
		t.Offset, t.Length, t.OutputOffset, t.Type, t.Opcode, t.OriginInstructionSeq)
	ext = array2JSON("opIds", t.OperandIDs) + array2JSON("opOffsets", t.OperandOffsets) +
		array2JSON("opLens", t.OperandLens) + array2JSON("opOutputOffsets", t.OperandOutputOffsets) +
		array2JSON("implicitOpValues", t.OperandContents)
	fstr = "{" + s + ext + "}"
	return []byte(fstr), nil
}

func srcArray2JSON(arr []Source) string {
	var substr, fstr string
	for _, w := range arr {
		if substr != "" {
			substr = substr + ","
		}
		tmp, _ := json.Marshal(w)
		substr += fmt.Sprintf("%s", string(tmp))
	}
	fstr = fmt.Sprintf("[%s]", substr)
	return fstr
}

// MarshalJSON help encode []byte
func (s Content) MarshalJSON() ([]byte, error) {
	var fstr, str string
	str = fmt.Sprintf("\"value\":\"%s\",\"size\":%d,\"src\":%s", "0x"+hex.EncodeToString(s.Val), s.Size, srcArray2JSON(s.Src))
	fstr = "{" + str + "}"
	return []byte(fstr), nil
	//return json.Marshal(hex.EncodeToString(s))
}

func (s Content) String() string {
	tmp, _ := json.Marshal(s)
	return string(tmp)
}

func (s *OpStat) String() string {
	fstr := fmt.Sprintf("\"name\":\"%s\",\"gasUsed\":%d,\"pc\":%d,\"sequenceNo\":%d,\"accountAddr\":\"%s\",\"callDepth\":%d,\"reverted\":%d,\"failInfo\":\"%s\"",
		s.OpName, s.GasUsed, s.PC, s.TraceSeq, s.AccountAddr, s.CallDepth, s.Reverted, s.FailInfo)
	if s.BeforeMemSize != -1 && s.AfterMemSize != -1 {
		fstr += fmt.Sprintf(",\"beforeMemSize\":%d,\"afterMemSize\":%d",
			s.BeforeMemSize, s.AfterMemSize)
	}
	if s.ExceptionTag != 0 {
		fstr += fmt.Sprintf(",\"exceptionTag\":%d", s.ExceptionTag)
	}
	return fstr
}

// MarshalJSON custom the JSON format of Trace
func (t Trace) MarshalJSON() ([]byte, error) {
	var basic, stkin, stkout, memin, memout string

	basic = t.Basic.String()
	if t.Basic.OpName == "SUICIDE" || t.Basic.OpName == "CALL" {
		basic += fmt.Sprintf(",\"newAccountCreated\":%d", t.Basic.IsCreatedNewAddress)
	}

	if t.StackInput != nil {
		inputName := []string{}
		switch t.Basic.OpName {
		case "SHA3":
			inputName = []string{"memStart", "memSize"}
		case "CALLDATACOPY":
			inputName = []string{"memStart", "dataStart", "dataSize"}
		case "CODECOPY":
			inputName = []string{"memStart", "codeStart", "codeSize"}
		case "EXTCODECOPY":
			inputName = []string{"codeAddr", "memStart", "codeStart", "codeSize"}
		case "EXP":
			inputName = []string{"base", "exponent"}
		case "MLOAD":
			inputName = []string{"memStart"}
		case "SSTORE", "MSTORE8", "MSTORE":
			inputName = []string{"storeStart", "storeValue"}
		case "JUMP":
			inputName = []string{"opPc"}
		case "JUMPI":
			inputName = []string{"opPc", "jumpFlag"}
		case "DUP1", "DUP2", "DUP3", "DUP4", "DUP5", "DUP6", "DUP7", "DUP8":
			fallthrough
		case "DUP9", "DUP10", "DUP11", "DUP12", "DUP13", "DUP14", "DUP15", "DUP16":
			inputName = []string{"item"}
		case "SWAP1", "SWAP2", "SWAP3", "SWAP4", "SWAP5", "SWAP6", "SWAP7", "SWAP8":
			fallthrough
		case "SWAP9", "SWAP10", "SWAP11", "SWAP12", "SWAP13", "SWAP14", "SWAP15", "SWAP16":
			inputName = []string{"item0", "item1"}
		case "LOG0":
			inputName = []string{"memStart", "memSize"}
		case "LOG1":
			inputName = []string{"memStart", "memSize", "topic0"}
		case "LOG2":
			inputName = []string{"memStart", "memSize", "topic0", "topic1"}
		case "LOG3":
			inputName = []string{"memStart", "memSize", "topic0", "topic1", "topic2"}
		case "LOG4":
			inputName = []string{"memStart", "memSize", "topic0", "topic1", "topic2", "topic3"}
		case "CREATE":
			inputName = []string{"value", "memStart", "memSize"}
		case "CALL", "CALLCODE":
			inputName = []string{"gas", "to", "value", "memInStart", "memInSize", "memOutStart", "memOutSize"}
		case "DELEGATECALL":
			inputName = []string{"gas", "to", "memInStart", "memInSize", "memOutStart", "memOutSize"}
		case "RETURN":
			inputName = []string{"memStart", "memSize"}
		case "SUICIDE":
			inputName = []string{"to"}
		case "POP":
			inputName = []string{"itemValue"}

		default:
			for i := 0; i < len(t.StackInput.Val); i++ {
				inputName = append(inputName, fmt.Sprintf("input%d", i))
			}
		}
		if len(inputName) != 0 {
			for i, c := range t.StackInput.Val {
				stkin += fmt.Sprintf("\"%s\": %d", inputName[i], c.Src[0].OriginInstructionSeq)
				if i != len(t.StackInput.Val)-1 {
					stkin += ","
				}
			}
		}
		if stkin != "" {
			stkin = "," + stkin
		}
	}

	if t.StackOutput != nil {
		outputName := ""
		switch t.Basic.OpName {
		case "PUSH1", "PUSH2", "PUSH3", "PUSH4", "PUSH5", "PUSH6", "PUSH7", "PUSH8":
			fallthrough
		case "PUSH9", "PUSH10", "PUSH11", "PUSH12", "PUSH13", "PUSH14", "PUSH15", "PUSH16":
			fallthrough
		case "PUSH17", "PUSH18", "PUSH19", "PUSH20", "PUSH21", "PUSH22", "PUSH23", "PUSH24":
			fallthrough
		case "PUSH25", "PUSH26", "PUSH27", "PUSH28", "PUSH29", "PUSH30", "PUSH31", "PUSH32":
			fallthrough
		case "CALL", "CALLCODE", "DELEGATECALL":
			fallthrough
		case "CREATE":
			fallthrough
		default:
			outputName = "output"
		}
		if len(outputName) != 0 {
			tmp, _ := json.Marshal(t.StackOutput.Val[0])
			stkout = fmt.Sprintf("\"%s\":", outputName) + string(tmp)
		}
		if stkout != "" {
			stkout = "," + stkout
		}
	}

	memInName, memOutName := "memData", "memData"

	if t.MemoryInput != nil && t.MemoryOutput != nil {
		memInName, memOutName = "memInData", "memOutData"
	}

	if t.Basic.OpName[:2] == "V-" {
		memInName, memOutName = "input", "output"
	}

	if len(t.Basic.OpName) > 5 && t.Basic.OpName[len(t.Basic.OpName)-4:] == "COPY" {
		memInName, memOutName = "input", "output"
	}

	if t.MemoryInput != nil {
		subStr := "["
		for k, v := range t.MemoryInput.Val.Src {
			if k != 0 {
				subStr += ","
			}
			subStr += fmt.Sprintf("[%d,%d,%d,%d]", v.OriginInstructionSeq, v.Offset, v.Length, v.OutputOffset)
		}
		subStr += "]"
		val := fmt.Sprintf("{\"value\":\"0x%s\",\"size\":%d,\"src\":%s}", common.Bytes2Hex(t.MemoryInput.Val.Val), t.MemoryInput.Val.Size, subStr)
		memin = fmt.Sprintf(",\"%s\":%s", memInName, val)
	}

	if t.MemoryOutput != nil {
		tmp, _ := json.Marshal(t.MemoryOutput.Val)
		memout = fmt.Sprintf(",\"%s\":%s", memOutName, string(tmp))
	}

	fstr := "{" + basic + stkin + stkout + memin + memout + "}"
	return []byte(fstr), nil
}

package replay

import "github.com/ethereum/go-ethereum/common"

func (r *ReplayTracer) setSource(tp string, seq int) {
	switch r.traceLog[seq].Basic.OpName {
	case "SLOAD":
		addr := common.Bytes2Hex(r.traceLog[seq].StackInput.Val[0].Val)
		// We must make sure we load it before we set the trace
		val := r.getStoreValue(addr)
		if len(val.Src) == 0 {
			src := Source{
				Type:                 "storage",
				Opcode:               "SLOAD",
				OriginInstructionSeq: seq,
				OperandIDs:           []int{r.traceLog[seq].StackInput.Val[0].Src[0].OriginInstructionSeq},
				Offset:               0,
				Length:               r.traceLog[seq].StackOutput.Val[0].Size,
			}
			val.Src = append(val.Src, src)
			r.saveStorage(addr, val)
		}
		// Fetch the value
		r.traceLog[seq].StackOutput.Val[0].Src = []Source{val.Src[0]}

	case "MLOAD", "SHA3":
		src := Source{
			Type:                 "memory",
			Opcode:               "MLOAD",
			OriginInstructionSeq: seq,
			Offset:               0,
			Length:               r.traceLog[seq].StackOutput.Val[0].Size,
		}
		mLen := 32
		if r.traceLog[seq].Basic.OpName == "SHA3" {
			src.Opcode = "SHA3"
			src.Type = "computed"
			mLen = int(common.Bytes2Big(r.traceLog[seq].StackInput.Val[1].Val).Int64())
		}
		// Get the 32byte trace from memory
		startAddr := int(common.Bytes2Big(r.traceLog[seq].StackInput.Val[0].Val).Int64())
		pt := startAddr
		for addr := startAddr; addr < startAddr+mLen; addr++ {
			if addr == pt {
				if len(r.pMem[addr].Src) == 0 {
					pt = addr + 1
				}
				continue
			}

			// We meet initialize memory
			if len(r.pMem[addr].Src) == 0 {
				if len(r.pMem[addr-1].Src) != 0 {
					src.OperandIDs = append(src.OperandIDs, r.pMem[pt].Src[0].OriginInstructionSeq)
					src.OperandOffsets = append(src.OperandOffsets, r.pMem[pt].Src[0].Offset)
					src.OperandLens = append(src.OperandLens, addr-pt)
					src.OperandOutputOffsets = append(src.OperandOutputOffsets, pt-startAddr)
				}
				pt = addr + 1
				continue
			}

			if len(r.pMem[addr].Src) != 0 && len(r.pMem[addr-1].Src) == 0 {
				continue
			}

			if !(r.pMem[addr-1].Src[0].OriginInstructionSeq == r.pMem[addr].Src[0].OriginInstructionSeq) && (r.pMem[addr-1].Src[0].Offset+1 == r.pMem[addr].Src[0].Offset) {
				src.OperandIDs = append(src.OperandIDs, r.pMem[pt].Src[0].OriginInstructionSeq)
				src.OperandOffsets = append(src.OperandOffsets, r.pMem[pt].Src[0].Offset)
				src.OperandLens = append(src.OperandLens, addr-pt)
				src.OperandOutputOffsets = append(src.OperandOutputOffsets, pt-startAddr)
				pt = addr
			}

			// For last byte
			if addr == startAddr+mLen-1 {
				if pt == startAddr+mLen {
					break
				}
				src.OperandIDs = append(src.OperandIDs, r.pMem[pt].Src[0].OriginInstructionSeq)
				src.OperandOffsets = append(src.OperandOffsets, r.pMem[pt].Src[0].Offset)
				src.OperandLens = append(src.OperandLens, addr-pt+1)
				src.OperandOutputOffsets = append(src.OperandOutputOffsets, pt-startAddr)
				pt = addr + 1
			}
		}

		if r.traceLog[seq].Basic.OpName == "MLOAD" {
			if len(src.OperandIDs) == 1 && src.OperandOffsets[0] == 0 && src.OperandLens[0] == 32 {
				src = Source(r.pMem[startAddr].Src[0])
			}
		}

		r.traceLog[seq].StackOutput.Val[0].Src = []Source{src}

	case "PUSH1", "PUSH2", "PUSH3", "PUSH4", "PUSH5", "PUSH6", "PUSH7", "PUSH8":
		fallthrough
	case "PUSH9", "PUSH10", "PUSH11", "PUSH12", "PUSH13", "PUSH14", "PUSH15", "PUSH16":
		fallthrough
	case "PUSH17", "PUSH18", "PUSH19", "PUSH20", "PUSH21", "PUSH22", "PUSH23", "PUSH24":
		fallthrough
	case "PUSH25", "PUSH26", "PUSH27", "PUSH28", "PUSH29", "PUSH30", "PUSH31", "PUSH32":
		src := Source{
			Type:                 "code",
			OriginInstructionSeq: seq,
			Opcode:               r.traceLog[seq].Basic.OpName,
			OperandContents:      []int{r.traceLog[seq].Basic.PC + 1},
			Offset:               0,
			Length:               r.traceLog[seq].StackOutput.Val[0].Size,
		}

		r.traceLog[seq].StackOutput.Val[0].Src = []Source{src}
		// Additional operation on PUSH, mush push the Value before

	default:
		src := Source{
			Type:                 tp,
			Opcode:               r.traceLog[seq].Basic.OpName,
			OriginInstructionSeq: int(seq),
			Offset:               0,
			Length:               r.traceLog[seq].StackOutput.Val[0].Size,
		}

		if r.traceLog[seq].StackInput != nil {
			for _, cnt := range r.traceLog[seq].StackInput.Val {
				for _, w := range cnt.Src {
					src.OperandIDs = append(src.OperandIDs, w.OriginInstructionSeq)
				}
			}
		}
		r.traceLog[seq].StackOutput.Val[0].Src = []Source{src}
	}

}

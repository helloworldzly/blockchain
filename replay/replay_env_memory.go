package replay

func (r *ReplayTracer) getMemory(mstart, msize int) Content {
	pt := mstart
	c := Content{
		Size: msize,
	}

	for addr := mstart; addr < mstart+msize; addr++ {
		c.Val = append(c.Val, r.pMem[addr].Val[0])
		if addr == pt {
			if len(r.pMem[addr].Src) == 0 {
				pt = addr + 1
			}
			continue
		}

		// We meet initialize memory
		if len(r.pMem[addr].Src) == 0 {
			if len(r.pMem[addr-1].Src) != 0 {
				src := Source(r.pMem[pt].Src[0])
				src.Length = addr - pt
				src.OutputOffset = pt - mstart
				c.Src = append(c.Src, src)
			}
			pt = addr + 1
			continue
		}

		if len(r.pMem[addr].Src) != 0 && len(r.pMem[addr-1].Src) == 0 {
			continue
		}

		if !((r.pMem[addr-1].Src[0].OriginInstructionSeq == r.pMem[addr].Src[0].OriginInstructionSeq) && (r.pMem[addr-1].Src[0].Offset+1 == r.pMem[addr].Src[0].Offset)) {
			src := Source(r.pMem[pt].Src[0])
			src.Length = addr - pt
			src.OutputOffset = pt - mstart
			c.Src = append(c.Src, src)
			pt = addr
		}

		// For last byte
		if addr == mstart+msize-1 {
			if pt == mstart+msize {
				break
			}
			src := Source(r.pMem[pt].Src[0])
			src.Length = addr - pt + 1
			src.OutputOffset = pt - mstart
			c.Src = append(c.Src, src)
			pt = addr + 1
		}
	}
	return c
}

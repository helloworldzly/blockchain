package replay

func (r *ReplayTracer) setOut(addr, key string, val Content) {
	if _, ok := r.storeOut[addr]; !ok {
		r.storeOut[addr] = make(map[string]Content)
	}
	r.storeOut[addr][key] = val
}

func (r *ReplayTracer) setIn(addr, key, val string) {
	if _, ok := r.storeIn[addr]; !ok {
		r.storeIn[addr] = make(map[string]string)
	}
	if _, ok := r.storeIn[addr][key]; !ok {
		r.storeIn[addr][key] = val
	}
}

func (r *ReplayTracer) hasStore(addr, key string) bool {
	if _, ok := r.storeOut[addr]; !ok {
		return false
	}
	if _, ok := r.storeOut[addr][key]; !ok {
		return false
	}
	return true
}

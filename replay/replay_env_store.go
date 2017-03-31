package replay

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

func (r *ReplayTracer) getStoreValue(loc string) Content {
	if v, ok := r.pStore[r.pAddr][loc]; ok {
		return v
	}
	for i := len(r.txStore) - 1; i >= 0; i-- {
		if v, ok := r.txStore[i][r.pAddr][loc]; ok {
			return v
		}
	}
	return Content{
		Val:  common.BigToBytes(big.NewInt(0), 256),
		Size: 32,
	}
}

func (r *ReplayTracer) saveStorage(loc string, val Content) {
	r.pStore[r.pAddr][loc] = val
}

func mergeStore(p1, p2 storage) storage {
	for k, v := range p2 {
		if _, ok := p1[k]; !ok {
			p1[k] = context{}
		}
		for subk, subv := range v {
			p1[k][subk] = subv
		}
	}
	return p1
}

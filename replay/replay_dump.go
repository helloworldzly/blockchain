package replay

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sync"
)

func padding(s string, n int) string {
	zero := ""
	for i := 0; i < n-len(s); i++ {
		zero += "0"
	}
	return zero + s
}

func strStore(w map[string]string) string {
	fstr := ""
	for k, v := range w {
		if fstr != "" {
			fstr = fstr + ","
		}
		fstr += fmt.Sprintf("\"%s\":\"%s\"", padding(k, 64), padding(v, 64))
	}
	return "{" + fstr + "}"
}

func (r *ReplayTracer) strAccount(val accountState, t string) string {
	fstr := ""
	fstr += fmt.Sprintf("\"accountAddr\":\"%s\"", val.Address)
	if t == "input" {
		fstr += fmt.Sprintf(",\"codeHash\":\"%s\"", val.beforeCodeHash)
		fstr += fmt.Sprintf(",\"nonce\":%d", val.beforeNonce)
		fstr += fmt.Sprintf(",\"balance\":\"%s\"", val.beforeBalance)
		fstr += r.strInputStore(val.Address)
	} else {
		fstr += fmt.Sprintf(",\"codeHash\":\"%s\"", val.afterCodeHash)
		fstr += fmt.Sprintf(",\"nonce\":%d", val.afterNonce)
		fstr += fmt.Sprintf(",\"balance\":\"%s\"", val.afterBalance)
		if val.isSmart != -1 {
			fstr += fmt.Sprintf(",\"isSmart\":%d", val.isSmart)
		}
		fstr += r.strOutputStore(val.Address)
	}

	return "{" + fstr + "}"
}

func (r *ReplayTracer) strInputState() string {
	fstr := ""
	for _, v := range r.ioAccounts {
		if v.beforeBalance == "invalid" {
			continue
		}
		if fstr != "" {
			fstr = fstr + ","
		}
		fstr += fmt.Sprintf("%s", r.strAccount(v, "input"))
	}
	return "[" + fstr + "]"
}

func (r *ReplayTracer) strOutputState() string {
	fstr := ""
	for _, v := range r.ioAccounts {
		if v.afterBalance == "invalid" {
			continue
		}
		if fstr != "" {
			fstr = fstr + ","
		}
		fstr += fmt.Sprintf("%s", r.strAccount(v, "output"))
	}
	return "[" + fstr + "]"
}

func (r *ReplayTracer) getInputStore(addr string) map[string]string {
	if v, ok := r.storeIn[addr]; ok {
		return v
	}
	return make(map[string]string)
}

func (r *ReplayTracer) getOutputStore(addr string) map[string]Content {
	if v, ok := r.storeOut[addr]; ok {
		return v
	}
	return make(map[string]Content)
}

func (r *ReplayTracer) strInputStore(addr string) string {
	w := r.getInputStore(addr)
	fstr := ""
	for k, v := range w {
		if fstr != "" {
			fstr += ","
		}
		fstr += fmt.Sprintf("\"0x%s\":\"0x%s\"", k, v)
	}
	fstr = ",\"storageContents\":{" + fstr + "}"
	return fstr

}

func (r *ReplayTracer) strOutputStore(addr string) string {
	w := r.getOutputStore(addr)
	fstr := ""
	for k, v := range w {
		if fstr != "" {
			fstr += ","
		}
		fstr += fmt.Sprintf("\"0x%s\":%s", k, v)
	}
	fstr = ",\"storageContents\":{" + fstr + "}"
	return fstr

}

func (r *ReplayTracer) strCreateSuicide() string {
	fstr := ""

	s := ""
	for k := range r.GetCreatedAccounts() {
		if s != "" {
			s = s + ","
		}
		s = s + "\"" + k + "\""
	}
	s = "\"createdAccounts\":[" + s + "]"
	fstr = fstr + s

	s = ""
	for k, v := range r.GetSuicidedAccounts() {
		if r.traceLog[k].Basic.Reverted == 1 {
			continue
		}
		if s != "" {
			s = s + ","
		}
		s = s + "\"" + v + "\""
	}
	s = "\"deletedAccounts\":[" + s + "]"
	fstr = fstr + "," + s

	return fstr
}

func (r *ReplayTracer) strCode() string {
	fstr := ""
	cnt := 0
	fstr = fstr + "{"
	for k, v := range r.codeStore {
		if v == "" || v == "0x" {
			continue
		}
		if cnt != 0 {
			fstr = fstr + ","
		}
		cnt++
		fstr = fstr + fmt.Sprintf("\"%s\":\"%s\"", padding(k, 40), v)
	}
	fstr = fstr + "}"
	return fstr
}

func (r *ReplayTracer) threadStrTrace(start, end int, buf *bytes.Buffer, wg *sync.WaitGroup) {
	for i := start; i < end; i++ {
		if i != 0 {
			buf.WriteString(",")
		}
		tmp, _ := json.Marshal(r.traceLog[i])
		buf.Write(tmp)
	}
	wg.Done()
}

func (r *ReplayTracer) strTrace() string {
	var (
		wg        sync.WaitGroup
		subBuf    = [20]bytes.Buffer{}
		threadNum = 1
		batch     = 1000
		fstr      string
	)

	if len(r.traceLog) > 1000 {
		threadNum = 8
	}

	fstr = fstr + "["
	seq := 0
	for seq < len(r.traceLog) {
		subBuf = [20]bytes.Buffer{}
		for i := 0; i < threadNum; i++ {
			if seq == len(r.traceLog) {
				break
			}
			stPos := seq
			enPos := seq + batch
			if enPos >= len(r.traceLog) {
				enPos = len(r.traceLog) - 1
			}
			seq = enPos + 1
			wg.Add(1)
			go r.threadStrTrace(stPos, enPos+1, &subBuf[i], &wg)
		}
		wg.Wait()
		for i := 0; i < threadNum; i++ {
			//fp.Write(subBuf[i].Bytes())
			fstr = fstr + subBuf[i].String()
		}
	}
	fstr = fstr + "]"
	return fstr
}

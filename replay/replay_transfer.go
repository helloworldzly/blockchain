package replay

import "fmt"

// MarshalJSON helps to turn source into meaningful JSON representation
func (t Trans) MarshalJSON() ([]byte, error) {
	var fstr, s string

	s = fmt.Sprintf("\"from\":\"%s\",\"to\":\"%s\",\"value\":\"%s\",\"type\":\"%s\"",
		t.From, t.To, t.Value, t.Type)

	if t.TxID != -1 {
		s += fmt.Sprintf(",\"txSeqNo\":%d", t.TxID)
	}

	if t.TxSeq != -1 {
		s += fmt.Sprintf(",\"traceSeqNo\":%d", t.TxSeq)
	}

	fstr = "{" + s + "}"

	return []byte(fstr), nil
}

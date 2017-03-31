package replay

// OpStat is the basic common part of each trace
type OpStat struct {
	OpName              string
	TraceSeq            int
	GasUsed             int
	PC                  int
	AccountAddr         string
	CallDepth           int
	Reverted            int
	FailInfo            string
	ExceptionTag        int
	BeforeMemSize       int
	AfterMemSize        int
	IsCreatedNewAddress int
}

// Source is a structure which trace the value
type Source struct {
	// For Memory Input/Output ref.
	Offset               int
	Length               int
	OutputOffset         int
	Type                 string
	Opcode               string
	OriginInstructionSeq int
	OperandIDs           []int
	OperandContents      []int
	// For partial ref. Should only be not-nil in MLOAD op (create new Source by merging all bytes)
	OperandOffsets       []int
	OperandLens          []int
	OperandOutputOffsets []int
}

// VarVal is the variable size array in byte
type VarVal []byte

// Content is the representation of each input and output
type Content struct {
	Val  VarVal   `json:"value"` // Val is the actual Content in flexible type (byte array)
	Size int      `json:"size"`  // Length in byte
	Src  []Source `json:"src"`   // Source array, for most case it should be only one element
}

// MemArray is the representation of array in Memory
type MemArray struct {
	Offset int
	Val    Content
}

// StackArray is the representation of a number of stack parameters in stack
type StackArray struct {
	Val []Content
}

// Len return the number of stack elements
func (s *StackArray) Len() int {
	return len(s.Val)
}

// Add push one content into stackarray
func (s *StackArray) Add(c Content) {
	s.Val = append(s.Val, c)
}

// Trace is stack and memory operation Trace
type Trace struct {
	Basic        *OpStat
	StackInput   *StackArray
	StackOutput  *StackArray
	MemoryInput  *MemArray
	MemoryOutput *MemArray
}

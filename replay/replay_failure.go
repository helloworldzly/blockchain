package replay

func errStr(err error) string {
	var (
		failInfo = "MaxCodeSizeExceeded"
		errStr   = err.Error()
	)
	if errStr == "MaxCodeSizeExceeded" {
		failInfo = "MaxCodeSizeExceeded"
	} else if errStr == "out of gas" {
		failInfo = "OutOfGas"
	} else if errStr == "contract creation code storage out of gas" {
		failInfo = "CodeStorageOutOfGas"
	} else if errStr[:8] == "max call" {
		failInfo = "StackOverflow"
	} else if errStr[:11] == "stack under" {
		failInfo = "LackOfStackItems"
	} else if errStr[:12] == "insufficient" {
		failInfo = "InsufficientBalance"
	} else if errStr[:12] == "invalid jump" {
		failInfo = "InvalidDest"
	} else if errStr[:14] == "Invalid opcode" {
		failInfo = "InvalidInstruction"
	}
	return failInfo
}

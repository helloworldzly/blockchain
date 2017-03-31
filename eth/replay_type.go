package eth

import (
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/params"
	"github.com/ethereum/go-ethereum/replay"
)

var (
	// Block level
	blockInputStates  string
	blockOutputStates string
	//blockCreatedAccounts string
)

func blockInit() {
	blockInputStates = ""
	blockOutputStates = ""
	//blockCreatedAccounts = ""
}

func addDaoForkBeforeState(statedb *state.StateDB) {
	addInputStates(params.DAORefundContract, statedb)
	for _, addr := range params.DAODrainList {
		if account := statedb.GetStateObject(addr); account != nil {
			addInputStates(addr, statedb)
		}
	}
}

func addDaoForkAfterState(statedb *state.StateDB) {
	addOutputStates(params.DAORefundContract, statedb)
	for _, addr := range params.DAODrainList {
		if account := statedb.GetStateObject(addr); account != nil {
			addOutputStates(addr, statedb)
		}
	}
}

func setInputStates(blk *types.Block, db *state.StateDB) {
	if blockInputStates != "" {
		blockInputStates += ","
	}
	blockInputStates += strState(blk.Coinbase(), db)
	for _, uncle := range blk.Uncles() {
		if blockInputStates != "" {
			blockInputStates += ","
		}
		blockInputStates += strState(uncle.Coinbase, db)
	}
}

func setOutputStates(blk *types.Block, db *state.StateDB) {
	if blockOutputStates != "" {
		blockOutputStates += ","
	}
	blockOutputStates += strState(blk.Coinbase(), db)
	for _, uncle := range blk.Uncles() {
		blockOutputStates += "," + strState(uncle.Coinbase, db)
	}
}

func addInputStates(addr common.Address, db *state.StateDB) {
	if blockInputStates != "" {
		blockInputStates += ","
	}
	blockInputStates += strState(addr, db)
}

func addOutputStates(addr common.Address, db *state.StateDB) {
	if blockOutputStates != "" {
		blockOutputStates += ","
	}
	blockOutputStates += strState(addr, db)
}

func numField(name string, value uint64) string {
	return fmt.Sprintf("\"%s\":%d", name, value)
}

func strField(name string, value string) string {
	return fmt.Sprintf("\"%s\":\"%s\"", name, value)
}

func padding(origin string, leng int) string {
	final := origin
	for i := 0; i < leng-len(origin); i++ {
		final = "0" + origin
	}
	return final
}

func strJSONArray(name string, content string) string {
	return fmt.Sprintf("\"%s\":[%s]", name, content)
}

func strBlockHeader(blk *types.Block) string {
	str := strField("ownHash", blk.Hash().Hex())
	str += "," + strField("prevHash", blk.ParentHash().Hex())
	str += "," + strField("uncleHash", blk.UncleHash().Hex())
	str += "," + strField("coinBase", blk.Coinbase().Hex())
	str += "," + strField("stateRoot", blk.Root().Hex())
	str += "," + strField("transactionRoot", blk.TxHash().Hex())
	str += "," + strField("receiptsRoot", blk.ReceiptHash().Hex())
	str += "," + strField("logBloom", "0x"+common.Bytes2Hex(blk.Bloom().Bytes()))
	str += "," + strField("difficulty", blk.Difficulty().String())
	str += "," + numField("number", blk.NumberU64())
	str += "," + numField("gasLimit", blk.GasLimit().Uint64())
	str += "," + numField("gasUsed", blk.GasUsed().Uint64())
	str += "," + numField("timestamp", blk.Time().Uint64())
	str += "," + strField("extraData", "0x"+common.Bytes2Hex(blk.Extra()))
	str += "," + strField("mixHash", blk.MixDigest().Hex())
	str += "," + strField("nonce", fmt.Sprintf("0x%x", blk.Nonce()))
	str += "," + numField("transactionCount", uint64(blk.Transactions().Len()))
	hstr := "{" + str + "}"
	return hstr
}

func strBlockUncle(blk *types.Header, bc *core.BlockChain) string {
	str := strField("ownHash", blk.Hash().Hex())
	str += "," + strField("prevHash", blk.ParentHash.Hex())
	str += "," + strField("uncleHash", blk.UncleHash.Hex())
	str += "," + strField("coinBase", blk.Coinbase.Hex())
	str += "," + strField("stateRoot", blk.Root.Hex())
	str += "," + strField("transactionRoot", blk.TxHash.Hex())
	str += "," + strField("receiptsRoot", blk.ReceiptHash.Hex())
	str += "," + strField("logBloom", "0x"+common.Bytes2Hex(blk.Bloom.Bytes()))
	str += "," + strField("difficulty", blk.Difficulty.String())
	str += "," + numField("number", blk.Number.Uint64())
	str += "," + numField("gasLimit", blk.GasLimit.Uint64())
	str += "," + numField("gasUsed", blk.GasUsed.Uint64())
	str += "," + numField("timestamp", blk.Time.Uint64())
	str += "," + strField("extraData", "0x"+common.Bytes2Hex(blk.Extra))
	str += "," + strField("mixHash", blk.MixDigest.Hex())
	str += "," + strField("nonce", fmt.Sprintf("0x%x", blk.Nonce))
	str += "," + numField("transactionCount", uint64(bc.GetBlockByNumber(blk.Number.Uint64()).Transactions().Len()))
	hstr := "{" + str + "}"
	return hstr
}

func strBlockUncles(blk *types.Block, bc *core.BlockChain) string {
	str := ""
	for _, v := range blk.Uncles() {
		if str != "" {
			str += ", "
		}
		str += strBlockUncle(v, bc)
	}
	return "[" + str + "]"
}

func strState(address common.Address, db *state.StateDB) string {
	return fmt.Sprintf("{\"accountAddr\":\"%s\",\"codeHash\":\"%s\",\"nonce\":%d,\"balance\":\"%s\"}",
		address.Hex(), db.GetCodeHash(address).Hex(), db.GetNonce(address), db.GetBalance(address).String())
}

func blkToJSON(blk *types.Block, db *state.StateDB, tracer replay.Tracer, bc *core.BlockChain) string {
	var (
		fstr string
	)
	//ts := time.Now()
	fstr += "," + fmt.Sprintf("\"blockHeader\":%s", strBlockHeader(blk))

	fstr += "," + fmt.Sprintf("\"uncleBlockHeaderList\":%s", strBlockUncles(blk, bc))
	fstr += "," + fmt.Sprintf("\"transferTrace\":%s", tracer.StrTransfer(int(blk.Number().Int64())))
	fstr += "," + strJSONArray("blockInputStates", blockInputStates)
	fstr += "," + strJSONArray("blockOutputStates", blockOutputStates)
	if blk.Number().Int64() == 0 {
		fstr += "," + strJSONArray("blockCreatedAccounts", "")
	} else {
		fstr += "," + strJSONArray("blockCreatedAccounts", tracer.StrBlockCreatedAccounts())
	}
	fstr += "}"
	return fstr
}

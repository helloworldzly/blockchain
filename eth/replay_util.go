package eth

import (
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/logger"
	"github.com/ethereum/go-ethereum/logger/glog"
	"github.com/ethereum/go-ethereum/params"
	"github.com/ethereum/go-ethereum/replay"
)

// replayBlock simulate the Processor()
func replayBlock(block *types.Block, statedb *state.StateDB, cfg *params.ChainConfig, bc *core.BlockChain) string {
	var (
		//receipts     types.Receipts
		totalUsedGas = big.NewInt(0)
		header       = block.Header()
		//allLogs      []*types.Log
		gp        = new(core.GasPool).AddGas(block.GasLimit())
		tracer    = replay.ReplayTracer{}
		blkString string
	)
	// Init block level data
	blockInit()
	tracer.BlockInit()
	blkString += fmt.Sprintf("{\"transactionList\":[")
	setInputStates(block, statedb)
	// Mutate the the block and state according to any hard-fork specs
	if cfg.DAOForkSupport && cfg.DAOForkBlock != nil && cfg.DAOForkBlock.Cmp(block.Number()) == 0 {
		addDaoForkBeforeState(statedb)
		ApplyDAOHardFork(statedb, &tracer)
		addDaoForkAfterState(statedb)
	}

	// Iterate over and process the individual transactions
	for i, tx := range block.Transactions() {
		//fmt.Println("tx:", i)
		tracer.TxInit(i)
		statedb.StartRecord(tx.Hash(), block.Hash(), i)
		nonce, gasPrice, startGas := tx.Nonce(), tx.GasPrice().Uint64(), tx.Gas().Uint64()
		tracer.SetInputAccount(block.Coinbase(), statedb.GetCodeHash(block.Coinbase()),
			statedb.GetBalance(block.Coinbase()), statedb.GetNonce(block.Coinbase()))

		receipt, _, _ := replayApplyTransaction(cfg, bc, gp, statedb, header, tx, totalUsedGas, vm.Config{}, &tracer)
		/*
			if err != nil {
				return nil, nil, nil, err
			}
		*/
		strData := ""
		strInit := ""
		strToAddr := ""
		if tx.To() == nil {
			strInit = "0x" + common.Bytes2Hex(tx.Data())
			strToAddr = "NIL"
		} else {
			strData = "0x" + common.Bytes2Hex(tx.Data())
			strToAddr = tx.To().Hex()
		}

		v, r, s := tx.RawSignatureValues()

		tracer.StrTxBasic(nonce, gasPrice, startGas, receipt.GasUsed.Uint64(),
			strData, strInit, strToAddr, tx.Value().String(),
			v.Int64(), r.String(), s.String(), tx.Hash().Hex())

		tracer.ScanTrace()
		for _, key := range tracer.GetInputAccounts() {
			addr := common.HexToAddress(key)
			tracer.SetOutputAccount(addr, statedb.GetCodeHash(addr), statedb.GetBalance(addr), statedb.GetNonce(addr))
		}

		if i != 0 {
			blkString += fmt.Sprintf(",")
		}

		blkString += tracer.StrTransaction()

		tracer.ValidateTransfer(i)
	}

	AccumulateRewards(statedb, header, block.Uncles(), &tracer)

	setOutputStates(block, statedb)

	blkString += "]"
	blkString += blkToJSON(block, statedb, &tracer, bc)

	return blkString
}

// AccumulateRewards credits the coinbase of the given block with the
// mining reward. The total reward consists of the static block reward
// and rewards for included uncles. The coinbase of each uncle block is
// also rewarded.
func AccumulateRewards(statedb *state.StateDB, header *types.Header, uncles []*types.Header, tracer replay.Tracer) {
	big8 := big.NewInt(8)
	big32 := big.NewInt(32)

	reward := new(big.Int).Set(core.BlockReward)
	r := new(big.Int)
	for _, uncle := range uncles {
		r.Add(uncle.Number, big8)
		r.Sub(r, header.Number)
		r.Mul(r, core.BlockReward)
		r.Div(r, big8)
		if !statedb.Exist(uncle.Coinbase) {
			tracer.AddBlockCreatedAccount(uncle.Coinbase)
		}
		statedb.AddBalance(uncle.Coinbase, r)
		tracer.AddTransfer("NIL", uncle.Coinbase.Hex(), r, "UncleReward")
		r.Div(core.BlockReward, big32)
		reward.Add(reward, r)
	}
	if !statedb.Exist(header.Coinbase) {
		tracer.AddBlockCreatedAccount(header.Coinbase)
	}
	statedb.AddBalance(header.Coinbase, reward)
	tracer.AddTransfer("NIL", header.Coinbase.Hex(), reward, "MinerReward")

}

// replayApplyTransaction is a modified version of ApplyTransaction in core package
func replayApplyTransaction(config *params.ChainConfig, bc *core.BlockChain, gp *core.GasPool, statedb *state.StateDB, header *types.Header, tx *types.Transaction, usedGas *big.Int, cfg vm.Config, tracer *replay.ReplayTracer) (*types.Receipt, *big.Int, error) {
	msg, err := tx.AsMessage(types.MakeSigner(config, header.Number))
	if err != nil {
		return nil, nil, err
	}
	addr := msg.From()
	tracer.TxBasic += fmt.Sprintf("\"sender\":\"%s\"", addr.Hex())

	tracer.SetInputAccount(addr, statedb.GetCodeHash(addr), statedb.GetBalance(addr), statedb.GetNonce(addr))

	if msg.To() != nil {
		address := *msg.To()
		if !statedb.Exist(address) {
			tracer.AddTxCreatedAccount(address)
		}
		tracer.SetInputAccount(address, statedb.GetCodeHash(address),
			statedb.GetBalance(address), statedb.GetNonce(address))
	}

	// Create a new context to be used in the EVM environment
	context := core.NewEVMContext(msg, header, bc)
	// Create a new environment which holds all relevant information
	// about the transaction and calling mechanisms.
	vmenv := vm.NewReplayEVM(context, statedb, config, cfg, tracer)

	// Apply the transaction to the current state (included in the env)
	_, gas, err := core.ApplyMessage(vmenv, msg, gp)
	if err != nil {
		return nil, nil, err
	}

	// Update the state with pending changes
	usedGas.Add(usedGas, gas)
	// Create a new receipt for the transaction, storing the intermediate root and gas used by the tx
	// based on the eip phase, we're passing wether the root touch-delete accounts.
	receipt := types.NewReceipt(statedb.IntermediateRoot(config.IsEIP158(header.Number)).Bytes(), usedGas)
	receipt.TxHash = tx.Hash()
	receipt.GasUsed = new(big.Int).Set(gas)
	// if the transaction created a contract, store the creation address in the receipt.
	if msg.To() == nil {
		receipt.ContractAddress = crypto.CreateAddress(vmenv.Context.Origin, tx.Nonce())
		if tracer.GetTxFailReason() == "" {
			tracer.SetCreatedAddress(receipt.ContractAddress)
		}
	}

	// Set the receipt logs and create a bloom for filtering
	receipt.Logs = statedb.GetLogs(tx.Hash())
	receipt.Bloom = types.CreateBloom(types.Receipts{receipt})

	glog.V(logger.Debug).Infoln(receipt)

	return receipt, gas, err
}

// ApplyDAOHardFork modifies the state database according to the DAO hard-fork
// rules, transferring all balances of a set of DAO accounts to a single refund
// contract.
func ApplyDAOHardFork(statedb *state.StateDB, tracer replay.Tracer) {
	// Retrieve the contract to refund balances into
	refund := statedb.GetOrNewStateObject(params.DAORefundContract)

	// Move every DAO account and extra-balance account funds into the refund contract
	for _, addr := range params.DAODrainList {
		if account := statedb.GetStateObject(addr); account != nil {
			tracer.AddTransfer(account.Address().Hex(), refund.Address().Hex(), account.Balance(), "DAOForkTransfer")
			refund.AddBalance(account.Balance())
			account.SetBalance(new(big.Int))
		}
	}
}

// Copyright 2015 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package eth

import (
	"math/big"

	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/eth/downloader"
	"github.com/ethereum/go-ethereum/eth/gasprice"
	"github.com/ethereum/go-ethereum/ethdb"
	"github.com/ethereum/go-ethereum/event"
	"github.com/ethereum/go-ethereum/internal/ethapi"
	"github.com/ethereum/go-ethereum/params"
	"github.com/ethereum/go-ethereum/rpc"
	"golang.org/x/net/context"
)

// EthApiBackend implements ethapi.Backend for full nodes
type EthApiBackend struct {
	eth *Ethereum
	gpo *gasprice.GasPriceOracle
}

func (b *EthApiBackend) ChainConfig() *params.ChainConfig {
	return b.eth.chainConfig
}

func (b *EthApiBackend) CurrentBlock() *types.Block {
	return b.eth.blockchain.CurrentBlock()
}

func (b *EthApiBackend) SetHead(number uint64) {
	b.eth.blockchain.SetHead(number)
}

func (b *EthApiBackend) HeaderByNumber(ctx context.Context, blockNr rpc.BlockNumber) (*types.Header, error) {
	// Pending block is only known by the miner
	if blockNr == rpc.PendingBlockNumber {
		block := b.eth.miner.PendingBlock()
		return block.Header(), nil
	}
	// Otherwise resolve and return the block
	if blockNr == rpc.LatestBlockNumber {
		return b.eth.blockchain.CurrentBlock().Header(), nil
	}
	return b.eth.blockchain.GetHeaderByNumber(uint64(blockNr)), nil
}

func (b *EthApiBackend) BlockByNumber(ctx context.Context, blockNr rpc.BlockNumber) (*types.Block, error) {
	// Pending block is only known by the miner
	if blockNr == rpc.PendingBlockNumber {
		block := b.eth.miner.PendingBlock()
		return block, nil
	}
	// Otherwise resolve and return the block
	if blockNr == rpc.LatestBlockNumber {
		return b.eth.blockchain.CurrentBlock(), nil
	}
	return b.eth.blockchain.GetBlockByNumber(uint64(blockNr)), nil
}

func (b *EthApiBackend) ReplayBlockByNumber(ctx context.Context, blockNr rpc.BlockNumber) string {
	blk := b.eth.blockchain.GetBlockByNumber(uint64(blockNr))
	prvBlk := b.eth.blockchain.GetBlockByNumber(uint64(blockNr) - 1)
	config := b.eth.blockchain.Config()
	bc := b.eth.blockchain
	db, _ := bc.StateAt(prvBlk.Root())
	return replayBlock(blk, db, config, bc)
	//return b.eth.blockchain.GetBlockByNumber(uint64(blockNr)), nil

}

func (b *EthApiBackend) StateAndHeaderByNumber(ctx context.Context, blockNr rpc.BlockNumber) (ethapi.State, *types.Header, error) {
	// Pending state is only known by the miner
	if blockNr == rpc.PendingBlockNumber {
		block, state := b.eth.miner.Pending()
		return EthApiState{state}, block.Header(), nil
	}
	// Otherwise resolve the block number and return its state
	header, err := b.HeaderByNumber(ctx, blockNr)
	if header == nil || err != nil {
		return nil, nil, err
	}
	stateDb, err := b.eth.BlockChain().StateAt(header.Root)
	return EthApiState{stateDb}, header, err
}

func (b *EthApiBackend) GetBlock(ctx context.Context, blockHash common.Hash) (*types.Block, error) {
	return b.eth.blockchain.GetBlockByHash(blockHash), nil
}

func (b *EthApiBackend) GetReceipts(ctx context.Context, blockHash common.Hash) (types.Receipts, error) {
	return core.GetBlockReceipts(b.eth.chainDb, blockHash, core.GetBlockNumber(b.eth.chainDb, blockHash)), nil
}

func (b *EthApiBackend) GetTd(blockHash common.Hash) *big.Int {
	return b.eth.blockchain.GetTdByHash(blockHash)
}

func (b *EthApiBackend) GetVMEnv(ctx context.Context, msg core.Message, state ethapi.State, header *types.Header) (*vm.EVM, func() error, error) {
	statedb := state.(EthApiState).state
	from := statedb.GetOrNewStateObject(msg.From())
	from.SetBalance(common.MaxBig)
	vmError := func() error { return nil }

	context := core.NewEVMContext(msg, header, b.eth.BlockChain())
	return vm.NewEVM(context, statedb, b.eth.chainConfig, vm.Config{}), vmError, nil
}

func (b *EthApiBackend) SendTx(ctx context.Context, signedTx *types.Transaction) error {
	b.eth.txMu.Lock()
	defer b.eth.txMu.Unlock()

	b.eth.txPool.SetLocal(signedTx)
	return b.eth.txPool.Add(signedTx)
}

func (b *EthApiBackend) RemoveTx(txHash common.Hash) {
	b.eth.txMu.Lock()
	defer b.eth.txMu.Unlock()

	b.eth.txPool.Remove(txHash)
}

func (b *EthApiBackend) GetPoolTransactions() (types.Transactions, error) {
	b.eth.txMu.Lock()
	defer b.eth.txMu.Unlock()

	pending, err := b.eth.txPool.Pending()
	if err != nil {
		return nil, err
	}

	var txs types.Transactions
	for _, batch := range pending {
		txs = append(txs, batch...)
	}
	return txs, nil
}

func (b *EthApiBackend) GetPoolTransaction(hash common.Hash) *types.Transaction {
	b.eth.txMu.Lock()
	defer b.eth.txMu.Unlock()

	return b.eth.txPool.Get(hash)
}

func (b *EthApiBackend) GetPoolNonce(ctx context.Context, addr common.Address) (uint64, error) {
	b.eth.txMu.Lock()
	defer b.eth.txMu.Unlock()

	return b.eth.txPool.State().GetNonce(addr), nil
}

func (b *EthApiBackend) Stats() (pending int, queued int) {
	b.eth.txMu.Lock()
	defer b.eth.txMu.Unlock()

	return b.eth.txPool.Stats()
}

func (b *EthApiBackend) TxPoolContent() (map[common.Address]types.Transactions, map[common.Address]types.Transactions) {
	b.eth.txMu.Lock()
	defer b.eth.txMu.Unlock()

	return b.eth.TxPool().Content()
}

func (b *EthApiBackend) Downloader() *downloader.Downloader {
	return b.eth.Downloader()
}

func (b *EthApiBackend) ProtocolVersion() int {
	return b.eth.EthVersion()
}

func (b *EthApiBackend) SuggestPrice(ctx context.Context) (*big.Int, error) {
	return b.gpo.SuggestPrice(), nil
}

func (b *EthApiBackend) ChainDb() ethdb.Database {
	return b.eth.ChainDb()
}

func (b *EthApiBackend) EventMux() *event.TypeMux {
	return b.eth.EventMux()
}

func (b *EthApiBackend) AccountManager() *accounts.Manager {
	return b.eth.AccountManager()
}

type EthApiState struct {
	state *state.StateDB
}

func (s EthApiState) GetBalance(ctx context.Context, addr common.Address) (*big.Int, error) {
	return s.state.GetBalance(addr), nil
}

func (s EthApiState) GetCode(ctx context.Context, addr common.Address) ([]byte, error) {
	return s.state.GetCode(addr), nil
}

func (s EthApiState) GetState(ctx context.Context, a common.Address, b common.Hash) (common.Hash, error) {
	return s.state.GetState(a, b), nil
}

func (s EthApiState) GetNonce(ctx context.Context, addr common.Address) (uint64, error) {
	return s.state.GetNonce(addr), nil
}

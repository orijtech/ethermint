package ethereum

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/state"
	ethTypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/eth"
	"github.com/ethereum/go-ethereum/event"
	"github.com/ethereum/go-ethereum/node"
	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/rpc"

	abciTypes "github.com/tendermint/abci/types"

	emtTypes "github.com/tendermint/ethermint/types"

	rpcClient "github.com/tendermint/tendermint/rpc/lib/client"
)

/*
Backend interface
- describes a set of functionality that all backends need to be able to deliver
  and it doesn't care whether it is backed by parity or go-ethereum

EthereumBackend struct implements Backend
- starts only the blockchain, evm and rpc layer and nothing else
# interface level
- needs to listen to transactions created over geth-rpc
- needs to forward those transactions to the tendermint rpc
- needs to be able to communicate with the underlying ethereum blockchain and write transactions
- needs to forward tendermint queries to the ethereum client
- deliverTx takes an address to deposit the transaction fee too
- needs to create new coins from scratch when received over IBC and destroy them when sending over IBC
- ability to define the log level

# private
- needs to be configurable through a go-ethereum/parity config file
  - homestead or not
  - gas price
  - gas limit
- needs to disable PoW and only validate that transactions are correct state changes
- needs to be able to credit transactions fees to validator accounts
- shouldn't start the networking layer
- shouldn't start mining
- should only expose the correct RPC interfaces
  - syncronization message should be fired by tendermint core in a regular basis
*/

//----------------------------------------------------------------------
// Backend manages the underlying ethereum state for storage and processing,
// and maintains the connection to Tendermint for forwarding txs

// Backend handles the chain database and VM
type Backend struct {
	// backing ethereum structures
	config   *eth.Config
	ethereum *eth.Ethereum

	// txBroadcastLoop subscription
	txSub *event.TypeMuxSubscription

	// pending ...
	pending *pending

	// client for forwarding txs to tendermint
	client rpcClient.HTTPClient
}

// NewBackend creates a new Backend
func NewBackend(ctx *node.ServiceContext, config *eth.Config, client rpcClient.HTTPClient) (*Backend, error) {
	p := newPending()

	// eth.New takes a ServiceContext for the EventMux, the AccountManager,
	// and some basic functions around the DataDir.
	ethereum, err := eth.New(ctx, config, p)
	if err != nil {
		return nil, err
	}

	// send special event to go-ethereum to switch homestead=true
	currentBlock := ethereum.BlockChain().CurrentBlock()
	ethereum.EventMux().Post(core.ChainHeadEvent{currentBlock}) // nolint: vet, errcheck

	// We don't need PoW/Uncle validation
	ethereum.BlockChain().SetValidator(NullBlockProcessor{})

	ethBackend := &Backend{
		ethereum: ethereum,
		pending:  p,
		client:   client,
		config:   config,
	}
	return ethBackend, nil
}

// Ethereum returns the underlying the ethereum object
func (b *Backend) Ethereum() *eth.Ethereum {
	return b.ethereum
}

// Config returns the eth.Config
func (b *Backend) Config() *eth.Config {
	return b.config
}

//----------------------------------------------------------------------
// Handle block processing

// DeliverTx appends a transaction to the current block
func (b *Backend) DeliverTx(tx *ethTypes.Transaction) error {
	return b.pending.deliverTx(b.ethereum.BlockChain(), b.config, b.ethereum.ApiBackend.ChainConfig(), tx)
}

// AccumulateRewards accumulates the rewards based on the given strategy
func (b *Backend) AccumulateRewards(strategy *emtTypes.Strategy) {
	b.pending.accumulateRewards(strategy)
}

// Commit finalises the current block
func (b *Backend) Commit(receiver common.Address) (common.Hash, error) {
	return b.pending.commit(b.ethereum, receiver)
}

// ResetWork resets the current block to a fresh object
func (b *Backend) ResetWork(receiver common.Address) error {
	work, err := b.pending.resetWork(b.ethereum.BlockChain(), receiver)
	b.pending.work = work
	return err
}

// UpdateHeaderWithTimeInfo uses the tendermint header to update the ethereum header
func (b *Backend) UpdateHeaderWithTimeInfo(tmHeader *abciTypes.Header) {
	b.pending.updateHeaderWithTimeInfo(b.ethereum.ApiBackend.ChainConfig(), tmHeader.Time, tmHeader.GetNumTxs())
}

// GasLimit returns the maximum gas per block
func (b *Backend) GasLimit() big.Int {
	return b.pending.gasLimit()
}

//----------------------------------------------------------------------
// Implements: node.Service

// APIs returns the collection of RPC services the ethereum package offers.
func (b *Backend) APIs() []rpc.API {
	apis := b.Ethereum().APIs()
	retApis := []rpc.API{}
	for _, v := range apis {
		if v.Namespace == "net" {
			v.Service = NewNetRPCService(b.config.NetworkId)
		}
		if v.Namespace == "miner" {
			continue
		}
		if _, ok := v.Service.(*eth.PublicMinerAPI); ok {
			continue
		}
		retApis = append(retApis, v)
	}
	return retApis
}

// Start implements node.Service, starting all internal goroutines needed by the
// Ethereum protocol implementation.
func (b *Backend) Start(srvr *p2p.Server) error {
	go b.txBroadcastLoop()
	return nil
}

// Stop implements node.Service, terminating all internal goroutines used by the
// Ethereum protocol.
func (b *Backend) Stop() error {
	b.txSub.Unsubscribe()
	b.ethereum.Stop() // nolint: errcheck
	return nil
}

// Protocols implements node.Service, returning all the currently configured
// network protocols to start.
func (b *Backend) Protocols() []p2p.Protocol {
	return nil
}

//----------------------------------------------------------------------
// We need a block processor that just ignores PoW and uncles and so on

// NullBlockProcessor does not validate anything
type NullBlockProcessor struct{}

// ValidateBody does not validate anything
func (NullBlockProcessor) ValidateBody(*ethTypes.Block) error { return nil }

// ValidateState does not validate anything
func (NullBlockProcessor) ValidateState(block, parent *ethTypes.Block, state *state.StateDB, receipts ethTypes.Receipts, usedGas *big.Int) error {
	return nil
}

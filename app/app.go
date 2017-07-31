package app

import (
	abci "github.com/tendermint/abci/types"
	tmLog "github.com/tendermint/tmlibs/log"

	"github.com/tendermint/ethermint/ethereum"
	"github.com/tendermint/ethermint/strategies"
)

// EthermintApplication implements an ABCI application.
// It holds a reference to an ethereum backend, which can be implemented through
// various means, as long as it satisfies the ethereum.Backend interface.
// Furthermore, it also holds the strategy for this app. A strategy describes
// how to distribute rewards, such as block rewards and transaction fees, as
// well as to whom they should be given. A strategy also deals with validator
// set changes.
type EthermintApplication struct {
	// backend handles all ethereum related services, such as blockchain
	// and RPC.
	backend *ethereum.Backend

	// strategy for ABCI management. It handles block rewards, transaction
	// fees and validator set changes.
	strategy strategies.Strategy

	// logger is a configurable logger that can be nested.
	logger tmLog.logger
}

// NewEthermintApplication creates the abci application for ethermint
func NewEthermintApplication(backend *ethereum.Backend,
	strategy strategies.Strategy,
	logger tmLog.Logger) *EthermintApplication {
	app := &EthermintApplication{
		backend:  backend,
		strategy: strategy,
		logger:   logger,
	}
	return app
}

// Info returns information about the last height and app_hash to the tendermint
// engine.
func (a *EthermintApplication) Info() abci.ResponseInfo {
	return a.backend.Info()
}

// SetOption sets a configuration option.
// SetOption is used to set parameters such as gasprice that don't affect
// consensus.
func (a *EthermintApplication) SetOption(key string, value string) (log string) {
	return a.backend.SetOption(key, value)
}

// InitChain initialises the validator set.
func (a *EthermintApplication) InitChain(validators []*abci.Validator) {
	return a.backend.InitChain(validators)
}

// CheckTx checks a transaction is valid but does not mutate the state.
func (a *EthermintApplication) CheckTx(txBytes []byte) abci.Result {
	return a.backend.CheckTx(txBytes)
}

// DeliverTx executes a transaction against the latest state.
func (a *EthermintApplication) DeliverTx(txBytes []byte) abci.Result {
	return a.backend.DeliverTx(txBytes)
}

// BeginBlock starts a new ethereum block.
func (a *EthermintApplication) BeginBlock(hash []byte, tmHeader *abci.Header) {
	a.backend.BeginBlock(hash, tmHeader)
}

// EndBlock returns changes to the validator set.
func (a *EthermintApplication) EndBlock(height uint64) abci.ResponseEndBlock {
	return a.backend.EndBlock(height)
}

// Commit commits the block and returns a hash of the current state.
func (a *EthermintApplication) Commit() abci.Result {
	return a.backend.Commit()
}

// Query queries the state of the underlying ethereum instance.
func (a *EthermintApplication) Query(query abci.RequestQuery) abci.ResponseQuery {
	return a.backend.Query(query)
}

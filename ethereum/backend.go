package ethereum

import (
	abci "github.com/tendermint/abci/types"
)

// Backend is an interface that describes the functionality needed by
// different backend implementations. It doesn't care whether this
// backend is provided by go-ethereum or parity.
// It is an extended version of the ABCI protocol that is specific to Ethereum.
type Backend interface {
	// Info/Query connection
	// Info returns application information
	Info() abci.ResponseInfo
	// SetOptions on the application. These options must not affect consensus.
	SetOption(key string, value string) string
	// Query queries the underlying ethereum backend for its state.
	Query(query abci.RequestQuery) abci.ResponseQuery

	// Mempool connection
	// CheckTx validates a transaction for the mempool without running any
	// permanent state changes. This call should be very cheap.
	CheckTx(tx []byte) abci.Result

	// Consensus connection
	// InitChain sets the initial validator set for this ethermint instance.
	InitChain(validators []*abci.Validator)
	// BeginBlock signals the beginning of a new block.
	BeginBlock(hash []byte, header *abci.Header)
	// DeliverTx sends a transaction for full-processing. This call should
	// to update the state of the app.
	DeliverTx(tx []byte) abci.Result
	// EndBlock signals the end of a block. If during this block the
	// validator set has changed this should be returned here.
	EndBlock(height uint64) abci.ResponseEndBlock
	// Commit should fnialise the state changes and write them to the
	// underlying ethereum instance. It needs to return the root hash
	// of the current state.
	Commit() abci.Result
}

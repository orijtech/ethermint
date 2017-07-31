package app

import (
	abci "github.com/tendermint/abci/types"
	tmLog "github.com/tendermint/tmlibs/log"

	"github.com/tendermint/ethermint/ethereum"
	"github.com/tendermint/ethermint/strategies"
)

/*
Ethermint struct
- holds reference to an ethereum backend interface
- listens to incoming ABCI messages and forwards them to the backend
- has a strategy interface for what is a valid transaction in terms of price,
  and how validators should be compensated
*/

// EthermintApplication implements an ABCI application
type EthermintApplication struct {
	// backend handles all ethereum related services, such as blockchain and RPC
	backend *ethereum.Backend

	// strategy for validator compensation
	strategy strategies.Strategy

	logger tmLog.logger
}

// NewEthermintApplication creates the abci application for ethermint
func NewEthermintApplication(backend *ethereum.Backend,
	strategy strategies.Strategy, logger tmLog.Logger) *EthermintApplication {
	app := &EthermintApplication{
		backend:  backend,
		strategy: strategy,
		logger:   logger,
	}

	//err := app.backend.ResetWork(app.Receiver()) // init the block results
	return app
}

// Info returns information about the last height and app_hash to the tendermint engine
func (a *EthermintApplication) Info() abci.ResponseInfo {
	/*
		log.Info("Info")
		blockchain := app.backend.Ethereum().BlockChain()
		currentBlock := blockchain.CurrentBlock()
		height := currentBlock.Number()
		hash := currentBlock.Hash()

		// This check determines whether it is the first time ethermint gets started.
		// If it is the first time, then we have to respond with an empty hash, since
		// that is what tendermint expects.
		if height.Cmp(big.NewInt(0)) == 0 {
			return abciTypes.ResponseInfo{
				Data:             "ABCIEthereum",
				LastBlockHeight:  height.Uint64(),
				LastBlockAppHash: []byte{},
			}
		}

		return abciTypes.ResponseInfo{
			Data:             "ABCIEthereum",
			LastBlockHeight:  height.Uint64(),
			LastBlockAppHash: hash[:],
		}
	*/
}

// SetOption sets a configuration option
// SetOption is used to set parameters such as gasprice that don't affect consensus
func (a *EthermintApplication) SetOption(key string, value string) (log string) {
	//log.Info("SetOption")
	return ""
}

// InitChain initializes the validator set
func (a *EthermintApplication) InitChain(validators []*abci.Validator) {
	//log.Info("InitChain")
	//app.SetValidators(validators)
}

// CheckTx checks a transaction is valid but does not mutate the state
func (a *EthermintApplication) CheckTx(txBytes []byte) abci.Result {
	/*
		tx, err := decodeTx(txBytes)
		log.Info("Received CheckTx", "tx", tx)
		if err != nil {
			return abciTypes.ErrEncodingError.AppendLog(err.Error())
		}

		return app.validateTx(tx)
	*/
	return abci.OK
}

// DeliverTx executes a transaction against the latest state
func (a *EthermintApplication) DeliverTx(txBytes []byte) abci.Result {
	/*
		tx, err := decodeTx(txBytes)
		if err != nil {
			return abciTypes.ErrEncodingError.AppendLog(err.Error())
		}

		log.Info("Got DeliverTx", "tx", tx)
		err = app.backend.DeliverTx(tx)
		if err != nil {
			log.Warn("DeliverTx error", "err", err)

			return abciTypes.ErrInternalError.AppendLog(err.Error())
		}
		app.CollectTx(tx)
	*/
	return abci.OK
}

// BeginBlock starts a new Ethereum block
func (a *EthermintApplication) BeginBlock(hash []byte, tmHeader *abci.Header) {
	/*
		// update the eth header with the tendermint header
		app.backend.UpdateHeaderWithTimeInfo(tmHeader)
	*/
}

// EndBlock accumulates rewards for the validators and updates them
func (a *EthermintApplication) EndBlock(height uint64) abci.ResponseEndBlock {
	/*
		log.Info("EndBlock")
		app.backend.AccumulateRewards(app.strategy)
		return app.GetUpdatedValidators()
	*/
	return abci.ResponseEndBlock{}
}

// Commit commits the block and returns a hash of the current state
func (a *EthermintApplication) Commit() abci.Result {
	/*
		log.Info("Commit")
		blockHash, err := app.backend.Commit(app.Receiver())
		if err != nil {
			log.Warn("Error getting latest ethereum state", "err", err)
			return abciTypes.ErrInternalError.AppendLog(err.Error())
		}
		return abciTypes.NewResultOK(blockHash[:], "")
	*/
	return abci.OK
}

// Query queries the state of EthermintApplication
func (a *EthermintApplication) Query(query abci.RequestQuery) abci.ResponseQuery {
	/*
		log.Info("Query")
		var in jsonRequest
		if err := json.Unmarshal(query.Data, &in); err != nil {
			return abciTypes.ResponseQuery{Code: abciTypes.ErrEncodingError.Code, Log: err.Error()}
		}
		var result interface{}
		if err := app.rpcClient.Call(&result, in.Method, in.Params...); err != nil {
			return abciTypes.ResponseQuery{Code: abciTypes.ErrInternalError.Code, Log: err.Error()}
		}
		bytes, err := json.Marshal(result)
		if err != nil {
			return abciTypes.ResponseQuery{Code: abciTypes.ErrInternalError.Code, Log: err.Error()}
		}
		return abciTypes.ResponseQuery{Code: abciTypes.OK.Code, Value: bytes}
	*/
	return abci.ResponseQuery{}
}

/*
//-------------------------------------------------------

// validateTx checks the validity of a tx against the blockchain's current state.
// it duplicates the logic in ethereum's tx_pool
func (app *EthermintApplication) validateTx(tx *ethTypes.Transaction) abciTypes.Result {
	currentState, err := app.currentState()
	if err != nil {
		return abciTypes.ErrInternalError.AppendLog(err.Error())
	}

	var signer ethTypes.Signer = ethTypes.FrontierSigner{}
	if tx.Protected() {
		signer = ethTypes.NewEIP155Signer(tx.ChainId())
	}

	from, err := ethTypes.Sender(signer, tx)
	if err != nil {
		return abciTypes.ErrBaseInvalidSignature.
			AppendLog(core.ErrInvalidSender.Error())
	}

	// Make sure the account exist. Non existent accounts
	// haven't got funds and well therefor never pass.
	if !currentState.Exist(from) {
		return abciTypes.ErrBaseUnknownAddress.
			AppendLog(core.ErrInvalidSender.Error())
	}

	// Check for nonce errors
	currentNonce := currentState.GetNonce(from)
	if currentNonce > tx.Nonce() {
		return abciTypes.ErrBadNonce.
			AppendLog(fmt.Sprintf("Got: %d, Current: %d", tx.Nonce(), currentNonce))
	}

	// Check the transaction doesn't exceed the current block limit gas.
	gasLimit := app.backend.GasLimit()
	if gasLimit.Cmp(tx.Gas()) < 0 {
		return abciTypes.ErrInternalError.AppendLog(core.ErrGasLimitReached.Error())
	}

	// Transactions can't be negative. This may never happen
	// using RLP decoded transactions but may occur if you create
	// a transaction using the RPC for example.
	if tx.Value().Cmp(common.Big0) < 0 {
		return abciTypes.ErrBaseInvalidInput.
			SetLog(core.ErrNegativeValue.Error())
	}

	// Transactor should have enough funds to cover the costs
	// cost == V + GP * GL
	currentBalance := currentState.GetBalance(from)
	if currentBalance.Cmp(tx.Cost()) < 0 {
		return abciTypes.ErrInsufficientFunds.
			AppendLog(fmt.Sprintf("Current balance: %s, tx cost: %s", currentBalance, tx.Cost()))

	}

	intrGas := core.IntrinsicGas(tx.Data(), tx.To() == nil, true) // homestead == true
	if tx.Gas().Cmp(intrGas) < 0 {
		return abciTypes.ErrBaseInsufficientFees.
			SetLog(core.ErrIntrinsicGas.Error())
	}

	return abciTypes.OK
}
*/

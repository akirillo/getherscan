package api_server

import (
	"getherscan/pkg/models"
	"net/http"

	"github.com/ethereum/go-ethereum/common"
	"github.com/gorilla/mux"
	"github.com/jackc/pgtype"
)

func (apiServer *APIServer) HandleGetHead(writer http.ResponseWriter, request *http.Request) {
	head, err := apiServer.DB.GetHead()
	if err != nil {
		RespondWithError(
			request,
			writer,
			http.StatusInternalServerError,
			err.Error(),
		)
		return
	}

	RespondWithJSON(
		request,
		writer,
		http.StatusOK,
		*head,
	)
}

func (apiServer *APIServer) HandleGetBlockByHash(writer http.ResponseWriter, request *http.Request) {
	routeVars := mux.Vars(request)
	blockHash := routeVars["blockHash"]

	block, err := apiServer.DB.GetBlockByHash(blockHash)
	if err != nil {
		RespondWithError(
			request,
			writer,
			http.StatusBadRequest,
			err.Error(),
		)
		return
	}

	RespondWithJSON(
		request,
		writer,
		http.StatusOK,
		block,
	)
}

func (apiServer *APIServer) HandleGetBlockByNumber(writer http.ResponseWriter, request *http.Request) {
	routeVars := mux.Vars(request)
	blockNumber := new(pgtype.Numeric)
	err := blockNumber.Set(routeVars["blockNumber"])
	if err != nil {
		RespondWithError(
			request,
			writer,
			http.StatusBadRequest,
			err.Error(),
		)
		return
	}

	block, err := apiServer.DB.GetBlockByNumber(*blockNumber)
	if err != nil {
		RespondWithError(
			request,
			writer,
			http.StatusBadRequest,
			err.Error(),
		)
		return
	}

	RespondWithJSON(
		request,
		writer,
		http.StatusOK,
		block,
	)
}

type GetBlocksByTransactionHashPayload struct {
	CanonicalBlock models.Block
	OrphanedBlocks []models.OrphanedBlock
}

func (apiServer *APIServer) HandleGetBlocksByTransactionHash(writer http.ResponseWriter, request *http.Request) {
	routeVars := mux.Vars(request)
	transactionHash := routeVars["transactionHash"]

	payload := new(GetBlocksByTransactionHashPayload)

	transaction, err := apiServer.DB.GetTransactionByHash(transactionHash, true)
	if err != nil {
		RespondWithError(
			request,
			writer,
			http.StatusBadRequest,
			err.Error(),
		)
		return
	}

	payload.CanonicalBlock = transaction.Block

	orphanedTransactions, err := apiServer.DB.GetOrphanedTransactionsByHash(transactionHash)
	if err != nil {
		RespondWithError(
			request,
			writer,
			http.StatusBadRequest,
			err.Error(),
		)
		return
	}

	payload.OrphanedBlocks = make([]models.OrphanedBlock, len(orphanedTransactions))
	for i, orphanedTransaction := range orphanedTransactions {
		payload.OrphanedBlocks[i] = orphanedTransaction.OrphanedBlock
	}

	RespondWithJSON(
		request,
		writer,
		http.StatusOK,
		payload,
	)
}

func (apiServer *APIServer) HandleGetTransactionByHash(writer http.ResponseWriter, request *http.Request) {
	routeVars := mux.Vars(request)
	transactionHash := routeVars["transactionHash"]

	transaction, err := apiServer.DB.GetTransactionByHash(transactionHash, false)
	if err != nil {
		RespondWithError(
			request,
			writer,
			http.StatusBadRequest,
			err.Error(),
		)
		return
	}

	RespondWithJSON(
		request,
		writer,
		http.StatusOK,
		transaction,
	)
}

func (apiServer *APIServer) HandleGetAddressBalanceByBlockHash(writer http.ResponseWriter, request *http.Request) {
	routeVars := mux.Vars(request)
	if !common.IsHexAddress(routeVars["address"]) {
		RespondWithError(
			request,
			writer,
			http.StatusBadRequest,
			"Invalid address",
		)
		return
	}

	address := routeVars["address"]
	blockHash := routeVars["blockHash"]

	balance, err := apiServer.DB.GetAddressBalanceByBlockHash(address, blockHash)
	if err != nil {
		RespondWithError(
			request,
			writer,
			http.StatusBadRequest,
			err.Error(),
		)
	}

	RespondWithJSON(
		request,
		writer,
		http.StatusOK,
		balance,
	)
}

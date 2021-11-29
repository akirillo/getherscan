package api_server

import (
	"fmt"
	"getherscan/pkg/models"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

type APIServer struct {
	Server *http.Server
	Router *mux.Router
	DB     *models.DB
}

func (apiServer *APIServer) Initialize(dbConnectionString, port string) error {
	apiServer.DB = new(models.DB)
	err := apiServer.DB.Initialize(dbConnectionString)
	if err != nil {
		return err
	}

	apiServer.Router = mux.NewRouter()

	apiServer.Server = &http.Server{
		Addr:    fmt.Sprintf(":%s", port),
		Handler: apiServer.Router,
	}

	apiServer.Router.HandleFunc(
		"/getHead",
		apiServer.HandleGetHead,
	).Methods("GET")

	apiServer.Router.HandleFunc(
		"/getBlockByHash/{blockHash}",
		apiServer.HandleGetBlockByHash,
	).Methods("GET")

	apiServer.Router.HandleFunc(
		"/getBlockByNumber/{blockNumber}",
		apiServer.HandleGetBlockByNumber,
	).Methods("GET")

	apiServer.Router.HandleFunc(
		"/getBlocksByTransactionHash/{transactionHash}",
		apiServer.HandleGetBlocksByTransactionHash,
	).Methods("GET")

	apiServer.Router.HandleFunc(
		"/getTransactionByHash/{transactionHash}",
		apiServer.HandleGetTransactionByHash,
	).Methods("GET")

	apiServer.Router.HandleFunc(
		"/getAddressBalanceByBlockHash/{address}/{blockHash}",
		apiServer.HandleGetAddressBalanceByBlockHash,
	).Methods("GET")

	return nil
}

func (apiServer *APIServer) Serve() {
	log.Fatal(apiServer.Server.ListenAndServe())
}

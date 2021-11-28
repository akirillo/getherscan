package api_server

import (
	"fmt"
	"getherscan/pkg/models"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

type APIServer struct {
	Router *mux.Router
	DB     *models.DB
}

func (apiServer *APIServer) Initialize(dbConnectionString string) error {
	apiServer.DB = new(models.DB)
	err := apiServer.DB.Initialize(dbConnectionString)
	if err != nil {
		return err
	}

	apiServer.Router = mux.NewRouter()

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

func (apiServer *APIServer) Serve(port string) {
	log.Fatal(
		http.ListenAndServe(
			fmt.Sprintf(":%s", port),
			apiServer.Router,
		),
	)
}

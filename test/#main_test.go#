package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"getherscan/pkg/api_server"
	"getherscan/pkg/models"
	"getherscan/pkg/poller"
	"getherscan/pkg/test_utils"
	"log"
	"math/rand"
	"net/http"
	"os"
	"testing"

	"github.com/ethereum/go-ethereum/core/types"
)

// NOTE: ALL TESTS ASSUME THEY ARE BEING RUN FROM /test DIR (IMPORTANT
// FOR FETCHING TESTDATA FILES)

var testPoller *poller.Poller
var testAPIServer *api_server.APIServer

func TestMain(m *testing.M) {
	var err error

	wsRPCEndpoint := flag.String(
		"ws-rpc-endpoint",
		"wss://mainnet.infura.io/ws/v3/5b913333cf074541ac8566a9e91d807b",
		"websocket RPC endpoint",
	)

	dbConnectionString := flag.String(
		"db-connection-string",
		"host=localhost port=5432 user=postgres password=12345 dbname=postgres sslmode=disable",
		"database connection string",
	)

	// By default, we don't use any tracked addresses in testing,
	// because this requires on-the-fly querying of the RPC
	// endpoint to get address balances during indexing, and this
	// fails if we're testing blocks further than 128 spots from
	// the head of the current chain (our RPC endpoint doesn't
	// have access to archival state)
	trackedAddressesFilePath := flag.String(
		"tracked-addresses",
		"",
		"path to JSON file with array of addresses to track",
	)

	port := flag.String(
		"port",
		"8000",
		"port for API server",
	)

	flag.Parse()

	trackedAddresses := []string{}
	if *trackedAddressesFilePath != "" {
		trackedAddresses, err = poller.GetTrackedAddressesFromFile(*trackedAddressesFilePath)
		if err != nil {
			log.Fatal(err)
		}
	}

	testPoller = new(poller.Poller)
	err = testPoller.Initialize(*wsRPCEndpoint, *dbConnectionString, trackedAddresses)
	if err != nil {
		log.Fatal(err)
	}

	testAPIServer = new(api_server.APIServer)
	err = testAPIServer.Initialize(*dbConnectionString, *port)
	if err != nil {
		log.Fatal(err)
	}

	go testAPIServer.Serve()

	exitCode := m.Run()

	os.Exit(exitCode)
}

func testPrologue() error {
	err := testPoller.DB.ClearDB()
	if err != nil {
		return err
	}

	return nil
}

func TestBasicIndexing(t *testing.T) {
	err := testPrologue()
	if err != nil {
		t.Fatal(err)
	}

	blocks, err := test_utils.GetBlocksFromDir("testdata/basic_test/basic_blocks")
	if err != nil {
		t.Fatal(err)
	}

	err = test_utils.TestPoll(testPoller, blocks)
	if err != nil {
		t.Fatal(err)
	}

	err = test_utils.AssertCanonicalBlocks(testPoller, []types.Block{
		blocks[3],
		blocks[2],
		blocks[1],
		blocks[0],
	})
	if err != nil {
		t.Fatal(err)
	}

	err = test_utils.AssertOrphanedBlocks(testPoller, []types.Block{})
	if err != nil {
		t.Fatal(err)
	}
}

func TestReorgIndexing(t *testing.T) {
	err := testPrologue()
	if err != nil {
		t.Fatal(err)
	}

	// Indexes blocks in the following order:
	// 0. Canonical ancestor
	// 1. Soon-to-be-orphaned block (indexed canonically)
	// 2. Soon-to-be-canonical block (indexed as orphan)
	// 3. Child of 2 (triggers reorg)

	// Final state looks like:
	//  ---      ---      ---
	// | 0 |----| 2 |----| 3 |
	//  --- \    ---      ---
	//       \   ---
	//        \-| 1 |
	//           ---

	blocks, err := test_utils.GetBlocksFromDir("testdata/reorg_test/reorg_blocks")
	if err != nil {
		t.Fatal(err)
	}

	err = test_utils.TestPoll(testPoller, blocks)
	if err != nil {
		t.Fatal(err)
	}

	err = test_utils.AssertCanonicalBlocks(testPoller, []types.Block{
		blocks[3],
		blocks[2],
		blocks[0],
	})
	if err != nil {
		t.Fatal(err)
	}

	err = test_utils.AssertOrphanedBlocks(testPoller, []types.Block{blocks[1]})
	if err != nil {
		t.Fatal(err)
	}
}

// TODO: Test for multiple reorgs?

func TestGetHead(t *testing.T) {
	err := testPrologue()
	if err != nil {
		t.Fatal(err)
	}

	blocks, err := test_utils.GetBlocksFromDir("testdata/basic_test/basic_blocks")
	if err != nil {
		t.Fatal(err)
	}

	err = test_utils.TestPoll(testPoller, blocks)
	if err != nil {
		t.Fatal(err)
	}

	response, err := http.Get(fmt.Sprintf("http://localhost%s/getHead", testAPIServer.Server.Addr))
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()

	var head models.Block
	err = json.NewDecoder(response.Body).Decode(&head)
	if err != nil {
		t.Fatal(err)
	}

	if head.Hash != blocks[len(blocks)-1].Hash().Hex() {
		t.Fatal(errors.New("Incorrect head"))
	}
}

func TestGetBlockByHash(t *testing.T) {
	err := testPrologue()
	if err != nil {
		t.Fatal(err)
	}

	blocks, err := test_utils.GetBlocksFromDir("testdata/basic_test/basic_blocks")
	if err != nil {
		t.Fatal(err)
	}

	err = test_utils.TestPoll(testPoller, blocks)
	if err != nil {
		t.Fatal(err)
	}

	block := blocks[rand.Intn(len(blocks))]
	blockHash := block.Hash().Hex()

	response, err := http.Get(fmt.Sprintf(
		"http://localhost%s/getBlockByHash/%s",
		testAPIServer.Server.Addr,
		blockHash,
	))
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()

	var blockModel models.Block
	err = json.NewDecoder(response.Body).Decode(&blockModel)
	if err != nil {
		t.Fatal(err)
	}

	if blockModel.Hash != blockHash {
		t.Fatal(errors.New("Incorrect block"))
	}
}

func TestGetBlockByNumber(t *testing.T) {
	err := testPrologue()
	if err != nil {
		t.Fatal(err)
	}

	blocks, err := test_utils.GetBlocksFromDir("testdata/basic_test/basic_blocks")
	if err != nil {
		t.Fatal(err)
	}

	err = test_utils.TestPoll(testPoller, blocks)
	if err != nil {
		t.Fatal(err)
	}

	block := blocks[rand.Intn(len(blocks))]
	blockHash := block.Hash().Hex()
	blockNumberString := block.Number().String()

	response, err := http.Get(fmt.Sprintf(
		"http://localhost%s/getBlockByNumber/%s",
		testAPIServer.Server.Addr,
		blockNumberString,
	))
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()

	var blockModel models.Block
	err = json.NewDecoder(response.Body).Decode(&blockModel)
	if err != nil {
		t.Fatal(err)
	}

	if blockModel.Hash != blockHash {
		t.Fatal(errors.New("Incorrect block"))
	}
}

// TODO: TestGetBlocksByTransactionHash

func TestGetTransactionByHash(t *testing.T) {
	err := testPrologue()
	if err != nil {
		t.Fatal(err)
	}

	blocks, err := test_utils.GetBlocksFromDir("testdata/basic_test/basic_blocks")
	if err != nil {
		t.Fatal(err)
	}

	err = test_utils.TestPoll(testPoller, blocks)
	if err != nil {
		t.Fatal(err)
	}

	block := blocks[rand.Intn(len(blocks))]
	transactions := block.Transactions()
	transaction := transactions[rand.Intn(len(transactions))]
	transactionHash := transaction.Hash().Hex()

	response, err := http.Get(fmt.Sprintf(
		"http://localhost%s/getTransactionByHash/%s",
		testAPIServer.Server.Addr,
		transactionHash,
	))
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()

	var transactionModel models.Transaction
	err = json.NewDecoder(response.Body).Decode(&transactionModel)
	if err != nil {
		t.Fatal(err)
	}

	if transactionModel.Hash != transactionHash {
		t.Fatal(errors.New("Incorrect transaction"))
	}
}

// TODO: TestGetAddressBalanceByBlockHash

// General TODOs:
// - Test indexing of balances
// - Test edge cases
// - Fuzz testing

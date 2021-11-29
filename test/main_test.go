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

func testPrologue() (bool, error) {
	err := testPoller.DB.ClearDB()
	if err != nil {
		return false, err
	}

	trackedAddressesFlagIsSet := false
	flag.Visit(func(f *flag.Flag) {
		if f.Name == "tracked-addresses" {
			trackedAddressesFlagIsSet = true
		}
	})

	return trackedAddressesFlagIsSet, nil
}

func TestBasicIndexing(t *testing.T) {
	trackedAddressesFlagIsSet, err := testPrologue()
	if err != nil {
		t.Fatal(err)
	}

	var blocks []types.Block
	if trackedAddressesFlagIsSet {
		blocks, err = test_utils.GetBlocksFromDir("testdata/balance_test/recent_blocks")
		if err != nil {
			t.Fatal(err)
		}
	} else {
		blocks, err = test_utils.GetBlocksFromDir("testdata/basic_test/basic_blocks")
		if err != nil {
			t.Fatal(err)
		}
	}

	err = test_utils.TestPoll(testPoller, blocks)
	if err != nil {
		t.Fatal(err)
	}

	// If using recent blocks s.t. balances can be asserted, this
	// test assumes that the recent blocks do not contain a reorg

	canonicalBlocks := make([]types.Block, len(blocks))
	for i, _ := range blocks {
		canonicalBlocks[i] = blocks[len(blocks)-1-i]
	}

	orphanedBlocks := []types.Block{}

	err = test_utils.AssertCanonicalBlocks(testPoller, canonicalBlocks)
	if err != nil {
		t.Fatal(err)
	}

	err = test_utils.AssertOrphanedBlocks(testPoller, orphanedBlocks)
	if err != nil {
		t.Fatal(err)
	}

	if trackedAddressesFlagIsSet {
		err = test_utils.AssertBalances(testPoller, canonicalBlocks, orphanedBlocks)
		if err != nil {
			t.Fatal(err)
		}
	}
}

func TestReorgIndexing(t *testing.T) {
	trackedAddressesFlagIsSet, err := testPrologue()
	if err != nil {
		t.Fatal(err)
	}

	var blocks []types.Block
	if trackedAddressesFlagIsSet {
		blocks, err = test_utils.GetBlocksFromDir("testdata/balance_test/recent_blocks")
		if err != nil {
			t.Fatal(err)
		}
	} else {
		blocks, err = test_utils.GetBlocksFromDir("testdata/reorg_test/reorg_blocks")
		if err != nil {
			t.Fatal(err)
		}
	}

	err = test_utils.TestPoll(testPoller, blocks)
	if err != nil {
		t.Fatal(err)
	}

	// Regardless of whether or not recent blocks are being used,
	// assumes blocks are in the following order:
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

	canonicalBlocks := []types.Block{
		blocks[3],
		blocks[2],
		blocks[0],
	}

	orphanedBlocks := []types.Block{blocks[1]}

	err = test_utils.AssertCanonicalBlocks(testPoller, canonicalBlocks)
	if err != nil {
		t.Fatal(err)
	}

	err = test_utils.AssertOrphanedBlocks(testPoller, orphanedBlocks)
	if err != nil {
		t.Fatal(err)
	}

	if trackedAddressesFlagIsSet {
		err = test_utils.AssertBalances(testPoller, canonicalBlocks, orphanedBlocks)
		if err != nil {
			t.Fatal(err)
		}
	}
}

// TODO: Test for multiple reorgs?

func TestGetHead(t *testing.T) {
	trackedAddressesFlagIsSet, err := testPrologue()
	if err != nil {
		t.Fatal(err)
	}

	var blocks []types.Block
	if trackedAddressesFlagIsSet {
		blocks, err = test_utils.GetBlocksFromDir("testdata/balance_test/recent_blocks")
		if err != nil {
			t.Fatal(err)
		}
	} else {
		blocks, err = test_utils.GetBlocksFromDir("testdata/basic_test/basic_blocks")
		if err != nil {
			t.Fatal(err)
		}
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
	trackedAddressesFlagIsSet, err := testPrologue()
	if err != nil {
		t.Fatal(err)
	}

	var blocks []types.Block
	if trackedAddressesFlagIsSet {
		blocks, err = test_utils.GetBlocksFromDir("testdata/balance_test/recent_blocks")
		if err != nil {
			t.Fatal(err)
		}
	} else {
		blocks, err = test_utils.GetBlocksFromDir("testdata/basic_test/basic_blocks")
		if err != nil {
			t.Fatal(err)
		}
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
	trackedAddressesFlagIsSet, err := testPrologue()
	if err != nil {
		t.Fatal(err)
	}

	var blocks []types.Block
	if trackedAddressesFlagIsSet {
		blocks, err = test_utils.GetBlocksFromDir("testdata/balance_test/recent_blocks")
		if err != nil {
			t.Fatal(err)
		}
	} else {
		blocks, err = test_utils.GetBlocksFromDir("testdata/basic_test/basic_blocks")
		if err != nil {
			t.Fatal(err)
		}
	}

	err = test_utils.TestPoll(testPoller, blocks)
	if err != nil {
		t.Fatal(err)
	}

	block := blocks[rand.Intn(len(blocks))]
	blockHash := block.Hash().Hex()
	// NOTE: big.Int.String() will truncate trailing zeroes...
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

func TestGetBlocksByTransactionHash(t *testing.T) {
	trackedAddressesFlagIsSet, err := testPrologue()
	if err != nil {
		t.Fatal(err)
	}

	var blocks []types.Block
	if trackedAddressesFlagIsSet {
		blocks, err = test_utils.GetBlocksFromDir("testdata/balance_test/recent_blocks")
		if err != nil {
			t.Fatal(err)
		}
	} else {
		blocks, err = test_utils.GetBlocksFromDir("testdata/reorg_test/reorg_blocks")
		if err != nil {
			t.Fatal(err)
		}
	}

	err = test_utils.TestPoll(testPoller, blocks)
	if err != nil {
		t.Fatal(err)
	}

	// Assumes same ordering as in TestReorgIndexing

	// Find highest-fee transaction in canonical block, it should
	// be present in orphaned block
	block := blocks[2]
	blockHash := block.Hash().Hex()
	// Assumes proper indexing transactions (delegate the checking
	// of this to TestReorgIndexing)
	transaction, err := testPoller.DB.GetMostExpensiveTransactionForBlockHash(blockHash)
	if err != nil {
		t.Fatal(err)
	}

	response, err := http.Get(fmt.Sprintf(
		"http://localhost%s/getBlocksByTransactionHash/%s",
		testAPIServer.Server.Addr,
		transaction.Hash,
	))
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()

	var payload api_server.GetBlocksByTransactionHashPayload
	err = json.NewDecoder(response.Body).Decode(&payload)
	if err != nil {
		t.Fatal(err)
	}

	if payload.CanonicalBlock.Hash != blockHash {
		t.Fatal(errors.New("Incorrect canonical block"))
	}

	if len(payload.OrphanedBlocks) == 0 {
		t.Fatal(errors.New("No orphaned blocks"))
	}

	if len(payload.OrphanedBlocks) > 1 {
		t.Fatal(errors.New("Too many orphaned blocks"))
	}

	if payload.OrphanedBlocks[0].Hash != blocks[1].Hash().Hex() {
		t.Fatal(errors.New("Incorrect orphaned block"))
	}
}

func TestGetTransactionByHash(t *testing.T) {
	trackedAddressesFlagIsSet, err := testPrologue()
	if err != nil {
		t.Fatal(err)
	}

	var blocks []types.Block
	if trackedAddressesFlagIsSet {
		blocks, err = test_utils.GetBlocksFromDir("testdata/balance_test/recent_blocks")
		if err != nil {
			t.Fatal(err)
		}
	} else {
		blocks, err = test_utils.GetBlocksFromDir("testdata/basic_test/basic_blocks")
		if err != nil {
			t.Fatal(err)
		}
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

// Before running this test, make sure to use the save_blocks CLI
// command to save a set of recent blocks, so that we can fetch
// balances for them on-the-fly using the RPC endpoint
func TestGetAddressBalanceByBlockHash(t *testing.T) {
	trackedAddressesFlagIsSet, err := testPrologue()
	if err != nil {
		t.Fatal(err)
	}

	var blocks []types.Block
	if !trackedAddressesFlagIsSet {
		t.Log("tracked-addresses flag not set, skipping...")
		return
	} else {
		blocks, err = test_utils.GetBlocksFromDir("testdata/balance_test/recent_blocks")
		if err != nil {
			t.Fatal(err)
		}
	}

	err = test_utils.TestPoll(testPoller, blocks)
	if err != nil {
		t.Fatal(err)
	}

	address := testPoller.TrackedAddresses[rand.Intn(len(testPoller.TrackedAddresses))]

	block := blocks[rand.Intn(len(blocks))]
	blockHash := block.Hash().Hex()

	response, err := http.Get(fmt.Sprintf(
		"http://localhost%s/getAddressBalanceByBlockHash/%s/%s",
		testAPIServer.Server.Addr,
		address,
		blockHash,
	))
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()

	var balanceModel models.Balance
	err = json.NewDecoder(response.Body).Decode(&balanceModel)
	if err != nil {
		t.Fatal(err)
	}

	err = test_utils.AssertBalance(testPoller, balanceModel, address, block)
	if err != nil {
		t.Fatal(err)
	}
}

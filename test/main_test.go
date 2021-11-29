package test

import (
	"flag"
	"getherscan/pkg/models"
	"getherscan/pkg/poller"
	"getherscan/pkg/test_utils"
	"log"
	"testing"
)

// NOTE: ALL TESTS ASSUME THEY ARE BEING RUN FROM /test DIR (IMPORTANT
// FOR FETCHING TESTDATA FILES)

var wsRPCEndpoint string
var dbConnectionString string
var trackedAddresses []string

func init() {
	flag.StringVar(
		&wsRPCEndpoint,
		"ws-rpc-endpoint",
		"wss://mainnet.infura.io/ws/v3/5b913333cf074541ac8566a9e91d807b",
		"websocket RPC endpoint",
	)

	flag.StringVar(
		&dbConnectionString,
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

	log.Println(*trackedAddressesFilePath)

	if *trackedAddressesFilePath != "" {
		var err error
		trackedAddresses, err = poller.GetTrackedAddressesFromFile(*trackedAddressesFilePath)
		if err != nil {
			log.Fatal(err)
		}
	}
}

func testPrologue() (*poller.Poller, error) {
	testPoller := new(poller.Poller)
	err := testPoller.Initialize(wsRPCEndpoint, dbConnectionString, trackedAddresses)
	if err != nil {
		return nil, err
	}

	err = testPoller.DB.ClearDB()
	if err != nil {
		return nil, err
	}

	return testPoller, nil
}

func TestReorg(t *testing.T) {
	testPoller, err := testPrologue()
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

	// Assert that head is block 3
	head, err := testPoller.DB.GetHead()
	if err != nil {
		t.Fatal(err)
	}

	if head.Hash != blocks[3].Hash().Hex() {
		t.Fatal("Incorrect head")
	}

	// Assert that parent is block 2
	parent, err := testPoller.DB.GetBlockByHash(head.ParentHash)
	if err != nil {
		t.Fatal(err)
	}

	if parent.Hash != blocks[2].Hash().Hex() {
		t.Fatal("Incorrect parent")
	}

	// Assert that next parent is block 0
	grandParent, err := testPoller.DB.GetBlockByHash(parent.ParentHash)
	if err != nil {
		t.Fatal(err)
	}

	if grandParent.Hash != blocks[0].Hash().Hex() {
		t.Fatal("Incorrect grandparent")
	}

	// Assert that sole orphan is block 1
	var orphanedBlocks []models.OrphanedBlock
	err = testPoller.DB.Find(&orphanedBlocks).Error
	if err != nil {
		t.Fatal(err)
	}

	if len(orphanedBlocks) > 1 {
		t.Fatal("Too many orphans")
	} else if orphanedBlocks[0].Hash != blocks[1].Hash().Hex() {
		t.Fatal("Incorrect orphan")
	}
}

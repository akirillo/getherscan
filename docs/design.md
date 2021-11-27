# Golang Ethereum Indexer Design Doc

## Assignment Recap
API must enable:
- Query block by number
- Query block by hash
- Query transaction by hash
- Query blocks (plural) by transaction hash
- Query head of the chain
- Query address balance by block hash
  - From configurable set of tracked addresses

Indexer should appropriately handle reorgs / uncle blocks
- Need to keep uncle blocks to be able to respond to 4th query (I think)
  - Not sure what else could be meant by returning multiple blocks for a single transaction hash

Integration tests over real blocks, tests should include reorg

Should I include load balancing across many nodes/RPC endpoints?
- Load balancing in the sense that we load balance outbound polling requests to multiple nodes?
- Or in the sense that we load balance node responses across many poller instances?

## Components
1. Poller - Server that polls Ethereum RPC endpoints to download chain state, write it back to DB
2. DB - Database that indexes chain state
3. API - Server that exposes endpoints to the DB for querying of indexed data

## Poller
- As far as I can tell, this should handle all the functionality of what their design calls the "processor"
- If there is a reorg of depth > 1:
	- How are the uncle blocks expressed?
		- Are all the now-stale blocks considered uncles of the new head of the chain
		- NO - only the stale block that's the direct child of the canonical ancestor
	- Have to scan back to start of fork
### Functions
`poll`
- Runs as a listener for new blocks on websocket RPC endpoint
- ASSUMES THAT IT HEARS ABOUT ALL NEW BLOCKS
- For new block:
	- If new block parent hash != latest indexed block parent hash:
		- If new block total difficulty > latest indexed block total difficulty:
			- Reorg
				- Find common ancestor
				- Update all blocks after common ancestor up to and including latest indexed block to be orphan blocks
				- Update all blocks after common ancestor up to and inluding new block to be canonical blocks
		- Else if total difficulties equal and new block number < latest indexed block number:
			- Reorg
		- Else:
			- Write in as orphaned block
	- Else:
		- Create model for `block`
		- For each transaction in block:
			- Create model for `transaction`
		- For each address:
			- Create model for `balance`

## DB
- Proper structure to relate transactions to uncle blocks?
	- Could be present in any number of uncle blocks
	- Option 1: Keep only `block` and `transaction` models
		- `block` would have an `is_uncle` field
		- `transaction` would have primary key (`hash`, `block.hash`)
		- `transaction` would have `is_uncle` field
		- Getting most recent block / block by number would require `WHERE is_uncle = false`, uses scan
		- Getting transaction by hash would require `WHERE is_uncle = false`, uses scan
			- ASSUMING transaction hash stays the same between uncle and canonical blocks
		- Getting blocks by transaction hash would require scan across only one relation
		- Uncling a block would require `SET is_uncle = true`, lightweight
		- Uncling a transaction would require `SET is_uncle = true`, lightweight
	- Option 2: Have `uncle_block`, `uncle_transaction` models
		- `uncle_block` would have same fields as `block`
		- `uncle_transaction` would have primary key (`hash`, `uncle_block.hash`)
		- `transaction` would just have primary key `hash`
		- Getting most recent block / block by number doesn't require scan, just sorted index on `number` (true for option 1, too)
		- Getting transaction by hash doesn't require scan
			- Again, ASSUMING transaction hash stays the same between uncle and canonical blocks
		- Getting blocks by transaction hash requires scan across only one relation
		- Uncling a block would require `DELETE FROM blocks` and `INSERT INTO uncle_blocks`, small, constant # IOs (check this)
		- Uncling a transaction would require `DELETE FROM transactions` and `INSERT INTO uncle_transactions`, small, constant # IOs (check this)
- TLDR: Use `uncle_blocks`, `uncle_transactions` relations b/c (under current assumptions) the IO cost of deleting from `blocks`/`transactions` and inserting into `uncle_blocks`/`uncle_transactions` is less than the IO cost of performing scans on queries

### Models
`block`
- Primary key `hash`
- Sorted index on `number`

`uncle_block`
- Same fields as `block`
- No index on `number`

`transaction`
- Primary key `hash`
- Foreign key `block.hash`

`uncle_transaction`
- Primary key (`hash`, `uncle_block.hash`)
- Foreign key `uncle_block.hash`

`balance`
- Primary key `address`, `block.hash`
	- ASSUMING we don't care about balances on uncle blocks
		- Actually, in this case, we can keep the same primary key, just not have a foreign key on `hash`
- Foreign key `block.hash`

## API
Query block by number:
```
SELECT *
FROM blocks
WHERE number = ?;
```

Query block by hash:
```
SELECT *
FROM blocks
WHERE hash = ?;
```

Query transaction by hash:
```
SELECT *
FROM transactions
WHERE hash = ?;
```

Query blocks by transaction hash:
```
WITH tx_block_hash AS (
	SELECT block_hash
	FROM transactions
	WHERE hash = ?
);

SELECT *
FROM blocks
WHERE hash IN tx_block_hash;

UNION

WITH uncle_tx_block_hash AS (
	SELECT uncle_block_hash
	FROM uncle_transactions
	WHERE hash = ?
);

SELECT *
FROM uncle_blocks
WHERE hash IN uncle_tx_block_hash;
```

Query head of the chain:
```
SELECT *
FROM blocks
ORDER BY number DESC
LIMIT 1;
```

Query address balance by block hash:
```
SELECT *
FROM balances
WHERE address = ?_1
	AND block_hash = ?_2;
```

## Libraries
- [GETH](https://github.com/ethereum/go-ethereum)
	- [EthClient Medium Post](https://medium.com/coinmonks/web3-go-part-1-31c68c68e20e)
	- [GETH Book](https://goethereumbook.org/en/)
	- [EthClient Deep Dive](https://www.youtube.com/watch?v=R0Ia1U9Gxjg)
 - [Gorm](https://gorm.io/)

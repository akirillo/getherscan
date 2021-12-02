# Design Decisions

## Primary Components
While [the assignment](https://docs.google.com/document/d/1kuw1lCASYK8MS_8S1bn8EyjBwU7kALXC1rtzYJW2QBM/edit) called for an architecture that had a database, poller, API, and processor components, I found that all of the indexer-specific logic could be handled by the poller as it is reading in new blocks, and thus did not need a separate processor.

As such, the system consists of those 3 components: a database, a poller, and an API server.

## Database

The database runs PostgreSQL for no reason other than preference and prior experience - that, and the fact that I felt that a relational database model made more sense here than NoSQL.
The reason being that there are only very few models that needed to be tracked (laid out in detail below, but mostly just blocks and transactions), which already have rigorous schema according to the Ethereum database.
There are also well-defined relationships within these models, e.g. a transaction belongs to a block, etc.

It was certainly clear that there needed to be models for blocks, and for transactions.
One of the biggest decisions to make was figuring out how orphaned blocks and their transactions should be tracked. There were essentially 2 options:
1. Maintain an `is_orphaned` field on the block model
2. Have a separate model for orphaned blocks

There are a number of issues with the first option:
- It necessitates maintaining a many-to-many relationship between blocks and transactions, requiring a costly join  on (`block.hash`, `transaction.hash`).
	- For example, when we try to query blocks by transaction hash.
- Querying the most recent block / querying a block by number would require adding a `WHERE is_orphaned = false` clause, implying a costly table scan.
The benefit to the first option, though, is that orphaning a block simply requires setting `is_orphaned = true`, a relatively lightweight operation.

The second option, on the other hand, only keeps a transaction record for each block that it appears in (as opposed to a join table, which is as if each transaction was present in each block).

This means also maintaining a separate model for transactions that appear in orphaned blocks. If we didn't do this, transactions would need a composite primary key on (`block.hash`, `transaction.hash`), and querying a transaction by hash would imply a `WHERE` clause with a subquery checking if `block.hash` is found in the table of canonical blocks, meaning a costly scan.

By having separate models for orphaned blocks and their transactions, we can:
- Query blocks by transaction hash with one direct record access (for the canonical block containing the transaction), and a table scan (for the "orphaned" transactions with that hash, and then a direct access for each orphaned block they're associated with), instead of performing a join.
- Query the most recent block / block by number with a direct record access.
This does mean that to orphan a block, however, we would need to delete the block model, and all transaction models, and then create the orphaned block model, and the orphaned transaction models.

I reasoned that the second option is more optimal, as the cost of deleting/creating to orphan a block in the relatively rare case of reorgs is not as bad as the table scans and joins implied in the queries the indexer must handle.

Finally, in order to support querying of address balances, we could either store balances within the block model, or make a separate model for them. The issue with storing balances within a block model is that if we ever decide to change the set of addresses being tracked, we'd have to redefine the schema for blocks. Thus, a separate model with a composite primary key on (`address`, `block.hash`) made more sense.

## Poller

The poller was implemented using much of the code and logic from [geth](https://github.com/ethereum/go-ethereum). Using the geth's Ethereum client library, it listens for new blocks via a websocket to an RPC endpoint, and calls a single function, `Index`, for each one. It's worth noting that the websocket RPC endpoint I got access to through Infura did not provide access to "archival state," i.e. any blocks deeper than 128 from the current head.

The `Index` function basically has to handle 3 cases:
1. Normal operation: blocks come in sequentially, forming a continuous chain, with occasional uncles. Index incoming blocks and their transactions as canonical, and the uncle blocks as orphans.
2. Missing blocks: a new block comes in, doesn't point to the currently indexed canonical head, but has no indexed orphaned parent. Fetch all the ancestor blocks until an indexed canonical ancestor, index them as orphaned blocks, and fall through to the reorg logic.
3. Reorgs: blocks are mined on top of an orphan, or otherwise an orphaned fork has higher difficulty than the currently indexed canonical chain. Orphaned fork is canonicalized, canonical fork is orphaned.

One important thing to note is that (to spare my computer), the poller does not index back to the genesis block. What this means is that:
1. In the missing blocks case, the poller assumes that a canonical ancestor to the newly received block has been indexed.
2. If no blocks have been indexed yet (a special case of missing blocks), the newly received block is immediately indexed as canonical.

Because of this, we can't get the exact total difficulty of a chain at a certain block. So, to reconcile whether or not a reorg is necesary between two forks, we get their total difficulty since their shared ancestor. The total difficulty at the shared ancestor is the same, and we already operate under the assumption that the shared ancestor has been indexed.

## API Server

The implementation of the API server is fairly straightforward. It connects to the database, and exposes a REST API with endpoints for each of the queries listed in the assignment. It returns payloads in JSON format.

# Scaling Considerations

There are scaling approaches for each of the 3 components of the system.

## Database

Given that the database system is relational, scaling it horizontally typically comes with some challenges. Namely, while one can replicate horizontally, it's much harder to efficiently shard the database, since servicing queries may require collating rows from multiple servers.

However, only one of our queries requires returning multiple rows - querying blocks by transaction hash. Thus, this makes it feasible to distribute records across database instances such that a query can be always serviced by one instance. For example, we could ensure that all transactions records are always on the same server as their associated blocks, and partition transaction records across servers by hash. Thus, all orphaned transaction records with the same hash would be on the same server, and their associated block records would be, too, allowing us to service the "blocks by transaction hash" query from one instance.

The remaining queries don't require collating any rows. As such, despite being relational, this database is relatively shardable given the query specification.

Beyond that, classic database scaling techniques like read replicas, load balancing, and connection pooling can be employed.

## Poller



## API Server

- Load balancing
	- Connecting to multiple nodes/RPC endpoints (running multiple pollers)
	- Running multiple API servers
	- Distributing database
		- Consider NoSQL

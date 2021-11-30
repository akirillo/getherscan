# Design Decisions

- Don't need separate `processor` module, `poller` handles all the indexer-specific logic b/c it must be executed during polling
- Assumption that poller hears about all new blocks
- Model structure (separate models for canonical & orphaned blocks/transactions)

# Scaling Considerations

- Fault-tolerant operation
	- Removing assumption that indexer hears about all blocks
	- Requires access to archival state
- Load balancing
	- Connecting to multiple nodes/RPC endpoints (running multiple pollers)
	- Running multiple API servers
	- Distributing database
		- Consider NoSQL

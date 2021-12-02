# Design Decisions

- Don't need separate `processor` module, `poller` handles all the indexer-specific logic b/c it must be executed during polling
- Model structure (separate models for canonical & orphaned blocks/transactions)

## Primary Components
While [the assignment](https://docs.google.com/document/d/1kuw1lCASYK8MS_8S1bn8EyjBwU7kALXC1rtzYJW2QBM/edit) called for an architecture with the 

# Scaling Considerations

- Fault-tolerant operation
	- Requires access to archival state
- Load balancing
	- Connecting to multiple nodes/RPC endpoints (running multiple pollers)
	- Running multiple API servers
	- Distributing database
		- Consider NoSQL

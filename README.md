# getherscan

An Ethereum indexer written in Go, made possible by the open-source packages implemented in [geth](https://github.com/ethereum/go-ethereum).

The indexer consists of 3 primary components:
1. A PostgreSQL database which indexes blocks, transactions, orphaned blocks and their transactions, and address balances according these [models](pkg/models/).
2. The [poller](pkg/poller/), which listens for new blocks on a websocket RPC endpoint and indexes them into the database. Optionally takes in a list of addresses for which to track Ether balances on a per-block basis.
3. The [API server](pkg/api_server/), which serves responses to the following queries from the database:
    - GET `"/getHead"` - Fetches the currently indexed (canonical) head of the chain.
    - GET `"/getBlockByHash/{blockHash}"` - Fetches the (canonical) block with the given `blockHash`.
    - GET `"/getBlockByNumber/{blockNumber}"` - Fetches the (canonical) block with the given `blockNumber`.
    - GET `"/getBlocksByTransactionHash/{transactionHash}"` - Fetches the canonical block containing the transaction with the given `transactionHash`, along with any orphaned blocks that contain this transaction.
    - GET `"getTransactionByHash/{transactionHash}"` - Fetches the transaction with the given `transactionHash`.
    - GET `"getAddressBalanceByBlockHash/{address}/{blockHash}"` - Fetches the given `address`'s Ether balance at the block with the given `blockHash`, provided that this address was included in the list of addresses to track.

## Running `getherscan`

While the indexer can operate in a cloud environment, the following instructions spell out how to run it locally.

### Running the database

The most straightforward way to get a PostgreSQL instance running locally is to run a `postgres` docker container.

Make sure you have [Docker installed](https://docs.docker.com/get-docker/), then run a container with the following command:
```shell
docker run -d -p <HOST PORT>:5432 --name <CONTAINER NAME> -e POSTGRES_PASSWORD=<PASSWORD> postgres
```

As an example, here's what I used:
```shell
docker run -d -p 5432:5432 --name getherscan-postgres -e POSTGRES_PASSWORD=12345 postgres
```

### Running the poller

To run the poller, you need a websocket RPC endpoint, a connection string to the Postgres database instance defined above, and, optionally, a JSON file containing an array of hex addresses for which to track balances.

Once you have these, run the following command from the project root:
```shell
go run cmd/poller/main.go poll "<WEBSOCKET RPC ENDPOINT>" "<POSTGRES CONNECTION STRING>" <PATH TO TRACKED ADDRESSES JSON>
```

Here is an example using the database defined exactly as above, using an Infura websocket endpoint in my account, and using a list of addresses to track stored in this repo (run this from project root):
```shell
go run cmd/poller/main.go poll "wss://mainnet.infura.io/ws/v3/5b913333cf074541ac8566a9e91d807b" "host=localhost port=5432 user=postgres password=12345 dbname=postgres sslmode=disable" test/testdata/tracked_addresses.json
```

Once you see the `Listening for blocks...` log line, the poller is up and running! You should see it start printing `Indexed block <BLOCK NUMBER>` shortly.

### Running the API server

The API server also needs a connection string to the Postgres database instance, and a port number on which to run.

Run the following command in another shell (from the project root):
```shell
go run cmd/api_server/main.go serve "<POSTGRES CONNECTION STRING>" <PORT NUMBER>
```

Here is an example using the database defined exactly as above, using port `8000`:
```shell
go run cmd/api_server/main.go serve "host=localhost port=5432 user=postgres password=12345 dbname=postgres sslmode=disable" 8000
```

Once you see the `Listening on port <PORT NUMBER>` log line, the API server is up and running! You can now send the defined queries as GET requests to `"http://localhost:<PORT NUMBER>"` using `curl` or a tool like [Postman](https://www.postman.com/).

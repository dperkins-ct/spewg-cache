# spewg-cache

A distributed, in-memory cache with support for LRU and time-based eviction, designed for horizontal scaling and replication over HTTP.

## Features

- **Distributed caching**: Run multiple cache nodes with replication.
- **Eviction policies**: Supports both LRU (Least Recently Used) and time-based expiry.
- **HTTP API**: Simple, language-agnostic interface for cache operations and replication.

## Why HTTP?

HTTP is chosen as the communication protocol for its ubiquity, ease of integration, and compatibility with a wide range of tools and environments. It enables simple scaling, monitoring, and debugging, and allows clients in any language to interact with the cache cluster.

## Eviction Strategies

### LRU (Least Recently Used) Eviction

- Each cache node tracks access order for keys.
- When the cache reaches its maximum size, the least recently accessed item is evicted.
- Ensures frequently used items remain in cache.

### Time-based Expiry

- Each cache entry can have a TTL (time-to-live).
- Entries are automatically removed after their TTL expires, regardless of access frequency.
- Useful for caching data that becomes stale after a certain period.

## Example: Running Multiple Caches with Replication

Suppose you want to run three cache nodes on the same machine, each replicating to the others.

### Start three cache nodes

```sh
# Node 1
go run cache.go server.go  main.go -port=:8080 -peers=http://localhost:8081,http://localhost:8082

# Node 2
go run cache.go server.go  main.go -port=:8081 -peers=http://localhost:8080,http://localhost:8082
# Node 3
go run cache.go server.go  main.go -port=:8082 -peers=http://localhost:8081,http://localhost:8080
```

Each node will replicate cache updates to its peers.

## Example HTTP API Usage

### Set a value with TTL (in seconds)

```sh
curl -X POST -H "Content-Type: application/json" -d '{"key": "foo", "value": "bar"}' -i http://localhost:8080/set
```

### Get a value from peer node

```sh
curl -i "http://localhost:8081/get?key=foo"
```

### Confirm valid replication in logs
You should see log message in the requested server (in this case port 8080) that the input was successsfully replicated. Example:
`replication successful to http://localhost:8081`

### LRU Eviction Example

If you set a maximum cache size (e.g., `--max-size 1000`), the cache will evict the least recently used items when full.

## Rationale

- **Scalability**: Add or remove nodes as needed.
- **Simplicity**: HTTP API is easy to use and debug.
- **Flexibility**: Choose eviction strategy per use case.

---

_See source code and CLI help for more options and details._

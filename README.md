# spewg-cache

A distributed, in-memory cache with support for LRU and time-based eviction, designed for horizontal scaling and replication over HTTP.

## Features

- **Distributed caching**: Run multiple cache nodes with replication.
- **Eviction policies**: Supports both LRU (Least Recently Used) and time-based expiry.
- **Sharding**: Partitions data across multiple nodes
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

## Sharding
Sharding is a fundamental technic that is used to partition data across multiple nodeas, ensuring scalability and performance. Sharding offers the following benefits
    - Horizontal Scalaing: Allows you to scale horizontally by adding more nodes to your system. This enables the cache to handle larger datasets and higher request volumes without degrading performance
    - Load Distribution
    - Parallel Processing: multiple shards can process requests in parallel
    - Isolation of Failures: if one shard fails, others can continue to operate
    - Simplified management

### Hash-based Sharding
This cache uses hash-based sharding, where a hash based function is applied ot the shard key to determine the shard. This ensures a uniform distribution of data across shards. 

The hash ring provides a consistent way to map keys to nodes, even as the system scales. Consistent hashing minimizes disruptions caused by adding or removing nodes. The implementation in the patch focuses on simplicity.


## Example: Running Multiple Caches with Replication

Suppose you want to run three cache nodes on the same machine, each replicating to the others.

### Start three cache nodes

```sh
# Run the first instance
go run main.go -port=:8083 -peers=http://localhost:8080

# Run the second instance
go run main.go -port=:8080 -peers=http://localhost:8083
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
# OR
curl -i "http://localhost:8083/get?key=foo"
```

Depdending on how `foo` hashes, the value should be returned from either port 8080 or 8083. To test the hashing and key distribution, you can set multiple keys. 

### Confirm valid replication in logs
You should see log message in the requested server (in this case port 8080) that the input was successsfully replicated. Example:
`replication successful to http://localhost:8081`

### LRU Eviction Example

If you set a maximum cache size (e.g., `--max-size 1000`), the cache will evict the least recently used items when full.

## Rationale

- **Scalability**: Add or remove nodes as needed.
- **Simplicity**: HTTP API is easy to use and debug.
- **Flexibility**: Choose eviction strategy per use case.

## Further Enhancements
- **Optimization**
    - Cache replacement algorithms: these algorithms can offer improved hit rates and better adaptability to varying workloads compared to the traditional LRU algorithm. Examples include Low Inter-Reference Recency Set (LIRS) or Adaptive Replacement Cache (ARC)
    - Tuning eviction policies: fine tune the TTL values and LRU threshholds based on access patterns. This prevents premature eviction of valueable data.
    - Compression: Implement data compression techniques to reduce memory footprint of cached items.
    - Connection Pooling: optimize network communication by implementing connection pooling between cache clients and servers. This reduces overhead for establishing new connections for each request. Leading to faster response times
- **Metrics and Monitoring**
    - Key metrics: continuously monitor essential metrics such as cache hit rate, cache miss rate, eviction rate, latency, throughput and memory usage. Thses metrics provide valuable isights into the performance and may identify potential bottlenecks
    - Visualization: Use visualization tools such as Grafana to create dashboards that display metrics in real time. 
    - Alerting: set up alerts based on threshlds for critical metrics. For example, receive an alert if the hit rate drops below a certain percentage
- **Profiling**
    - CPU Profiling: Identify CPU intensive functions and pinpoint areas where optimization can yield performance gains
    - Memory Profiling: Analyze memory usage patterns to detect memory leaks or inefficient memory allocation. 
---

_See source code and CLI help for more options and details._

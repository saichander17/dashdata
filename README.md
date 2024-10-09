# Dashdata
Goal of this project is to develop a fully capable in memory database in Golang while also learning the language

## Features

- Fast in-memory storage
- Thread-safe operations
- Fine-grained locking mechanism for high concurrency
- Basic CRUD operations (Create, Read, Update, Delete)

## Quick Start

```go
store := NewStore()

// Set a value
store.Set("key", "value")

// Get a value
value, exists := store.Get("key")

// Delete a value
store.Delete("key")
```

## API Documentation
For detailed API documentation, see [API.md](API.md).

# Things to do
- [X] Write a key value pair with string data type for both key and value
- [X] Read the value using key
- [X] Update the value using key
- [X] Delete the value using key
- [ ] Persist the data to disk upon every write
- [ ] Configurable persistence level (High/Eventual)
- [ ] Add CRUD support for int, list, map and set datatypes
- [ ] Add support for disk fetching when memory is not enough to have all keys in memory
- [ ] Query language to query the data. Follow Redis mechanism
- [ ] Add support for TTL
- [ ] Write service which is listening on a port and accepts requests from clients
- [ ] All the requests must use multi threaded approach to utilise the compute to max efficiency
- [ ] Add support for horizontal scaling with configurable consistency levels. All nodes can accept read/write both at all times
- [ ] Add support for aggregation queries
- [ ] Data compression to optimise storage?
- [ ] Auth



### Blog for queuing requests
https://encore.dev/blog/queueing?ref=dailydev#processed-time-outs
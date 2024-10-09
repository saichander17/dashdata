# Dashdata API Documentation

## Store

The `Store` struct is the main data structure in Dashdata.

### NewStore()

Creates and returns a new `Store` instance.

```go
func NewStore() *Store
```

### Set(key, value string)

Sets a key-value pair in the store.

```go
func (s *Store) Set(key, value string)
```


### Get(key string) (string, bool)

Retrieves a value by its key. Returns the value and a boolean indicating whether the key exists.

```go
func (s *Store) Get(key string) (string, bool)
```

### Delete(key string)

Removes a key-value pair from the store.

```go
func (s *Store) Delete(key string)
```

### Internal Functions

#### lockIndex(key string) uint32
Determines the lock index for a given key using FNV hash.

```go
func (s *Store) lockIndex(key string) uint32
```
package store

// Store interface defines the methods that any store implementation must provide
type Store interface {
	Set(key, value string)
	Get(key string) (string, bool)
	Delete(key string)
}

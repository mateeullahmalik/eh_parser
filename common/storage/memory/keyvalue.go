package memory

import (
	"errors"
	"sync"

	"github.com/mateeullahmalik/eh_parser/common/storage"
)

type keyValue struct {
	store sync.Map
}

// Get retrieves a value by key. If the key does not exist, ErrKeyValueNotFound is returned.
func (db *keyValue) Get(key string) ([]byte, error) {
	if value, ok := db.store.Load(key); ok {
		if bytes, ok := value.([]byte); ok {
			return bytes, nil
		}
		return nil, errors.New("unable to get bytes from value")
	}
	return nil, storage.ErrKeyValueNotFound
}

// Delete removes a key and its value from the store.
func (db *keyValue) Delete(key string) error {
	db.store.Delete(key)
	return nil
}

// Set stores a key-value pair without expiration.
func (db *keyValue) Set(key string, value []byte) error {
	db.store.Store(key, value)
	return nil
}

// NewKeyValue returns a new instance of keyValue storage.
func NewKeyValue() storage.KeyValue {
	return &keyValue{}
}

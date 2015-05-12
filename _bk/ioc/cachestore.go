package ioc

import (
	"errors"
)

var (
	ErrCacheMiss = errors.New("cachestore: cache miss")
)

type CacheStoreReader interface {
	Get(string, interface{}) error
}

type CacheStoreWriter interface {
	Set(string, interface{}) error
	Delete(string) error
}

type CacheStoreReaderWriter interface {
	CacheStoreReader
	CacheStoreWriter
}

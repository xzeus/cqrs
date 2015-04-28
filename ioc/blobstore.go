package ioc

type BlobStoreReader interface {
	Get(string, int64, interface{}) error
}

type BlobStoreWriter interface {
	Put(string, int64, interface{}) error
	PutWithId(string, interface{}) (int64, error)
	Delete(string, int64) error
}

type BlobStoreReaderWriter interface {
	BlobStoreReader
	BlobStoreWriter
}

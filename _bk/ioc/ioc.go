package ioc

type DataDependencies interface {
	BlobStore() BlobStoreReaderWriter
	CacheStore() CacheStoreReaderWriter
	DataStore() DataStoreReaderWriter
	EventStore() EventStoreReaderWriter
}

type ServiceDependencies interface {
	Crypto() Crypto
	Exception() Exception
	HttpClient() HttpClient
	Logger() Logger
	Publisher() Publisher
	Time() Time
}

type Dependencies interface {
	DataDependencies
	ServiceDependencies
}

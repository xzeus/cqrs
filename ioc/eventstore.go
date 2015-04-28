package ioc

import (
	"errors"
	"github.com/xzeus/cqrs"
)

var (
	ErrInvalidEventKey       = errors.New("nil valued event key")
	ErrDeserializingEventKey = errors.New("unable to extract key from retrieved event")
	ErrStaleEventVersion     = errors.New("stale event version")
	ErrAggregateIdInUse      = errors.New("aggregate id was already in use")
	ErrAggregateKeyCollision = errors.New("aggregate key hash was in use by another aggreagate")
)

type EventStoreReader interface {
	GetSnapshot(domain int32, id int64) (cqrs.Aggregate, error)
	GetEvent(domain int32, id int64, version int32) (cqrs.Message, error)
	GetDomainEvents(domain int32, min_ts, max_ts int64) ([]cqrs.Message, error)
	GetAggregateEvents(domain int32, id int64, min_version int32) ([]cqrs.Message, error)
	GetAggregateEventsByPeriod(domain int32, id int64, min_ts, max_ts int64) ([]cqrs.Message, error)
	GetAggregateEventsWithSnapshot(domain int32, id int64) ([]cqrs.Message, cqrs.Aggregate, error)
	GetKeyedAggregateEvents(domain int32, key []byte, min_version int32) ([]cqrs.Message, error)
	// TODO: Keyed with snapshot?
}

type EventStoreWriter interface {
	StoreSnapshot(cqrs.Aggregate) error
	AppendEvent(id int64, version int32, origin []cqrs.AggregateHeader, payload cqrs.MessageDefiner) (cqrs.Message, error)
	AppendKeyedEvent(key []byte, origin []cqrs.AggregateHeader, payload cqrs.MessageDefiner) (cqrs.Message, error)
	DeleteEvent(domain int32, id int64, version int32) error
	DeleteAggregate(domain int32, id int64) error
}

type EventStoreReaderWriter interface {
	EventStoreReader
	EventStoreWriter
}

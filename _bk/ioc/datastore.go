package ioc

import (
	"errors"
)

var (
	ErrNoSuchData           = errors.New("datastore: data not stored")
	ErrKeyDataSliceMismatch = errors.New("datastore: keys not equal length to data types")
	ErrNestedTransaction    = errors.New("datastore: nested transactions")
)

type DataStoreReader interface {
	ExecQuery(query DataStoreQuerier, data interface{}) error
	Get(kind, key string, data interface{}) error
	GetInt(kind string, id int64, data interface{}) error
	GetMulti(kind string, keys []string, data ...interface{}) []error
	GetMultiInt(kind string, keys []int64, data ...interface{}) []error
	GetKinds() ([]string, error)
}

type DataStoreWriter interface {
	Put(kind, key string, data interface{}) error
	PutInt(kind string, id int64, data interface{}) error
	PutMulti(kind string, keys []string, data ...interface{}) []error
	PutMultiInt(kind string, keys []int64, data ...interface{}) []error
	Delete(kind, key string) error
	DeleteInt(kind string, id int64) error
	DeleteMulti(kind string, keys []string) []error
	DeleteMultiInt(kind string, keys []int64) []error
	DeleteKind(kind string) error
}

type DataStoreReaderWriter interface {
	RunInTransaction(trx_ds func(DataStoreReaderWriter) error) error
	DataStoreReader
	DataStoreWriter
}

type DataStoreQuerier interface {
	ToQuery() *DataStoreQuery
}

type DataStoreOffsetter interface {
	DataStoreQuerier
	Skip(skip int) interface{}
}

type DataStoreLimiter interface {
	DataStoreOffsetter
	Take(max_results int) DataStoreOffsetter
}

type DataStoreConstrainer interface {
	DataStoreLimiter
	Order(property string) DataStoreConstrainer
}

type DataStoreProjector interface {
	DataStoreConstrainer
	Project(properties ...string) DataStoreConstrainer
}

type DataStoreFilterer interface {
	DataStoreProjector
	Filter(property string, operand string, value interface{}) DataStoreFilterer
	Equals(property string, value interface{}) DataStoreFilterer
}

type DataStoreQueryFilter struct {
	Property string
	Operand  string
	Value    interface{}
}

type DataStoreQuery struct {
	Kind        string
	Filters     []DataStoreQueryFilter
	Projections []string
	OrderBy     []string
	Limit       int
	Offset      int
}

func NewDataStoreQuery(kind string) DataStoreFilterer {
	return &DataStoreQuery{
		Kind:        kind,
		Filters:     make([]DataStoreQueryFilter, 0),
		Projections: make([]string, 0),
		OrderBy:     make([]string, 0),
		Limit:       0,
		Offset:      0,
	}
}

func (query *DataStoreQuery) ToQuery() *DataStoreQuery {
	return query
}

func (query *DataStoreQuery) Filter(property string, operand string, value interface{}) DataStoreFilterer {
	query.Filters = append(query.Filters, DataStoreQueryFilter{
		Property: property,
		Operand:  operand,
		Value:    value,
	})
	return query
}

func (query *DataStoreQuery) Equals(property string, value interface{}) DataStoreFilterer {
	return query.Filter(property, "=", value)
}

func (query *DataStoreQuery) Project(projections ...string) DataStoreConstrainer {
	query.Projections = projections
	return query
}

func (query *DataStoreQuery) Order(property string) DataStoreConstrainer {
	query.OrderBy = append(query.OrderBy, property)
	return query
}

func (query *DataStoreQuery) Take(max_results int) DataStoreOffsetter {
	query.Limit = max_results
	return query
}

func (query *DataStoreQuery) Skip(offset int) interface{} {
	query.Offset = offset
	return query
}

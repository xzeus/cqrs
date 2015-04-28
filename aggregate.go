package cqrs

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

// NewAggregateHeaderData creates an aggregate instance with UUId derived from the provided values
func NewAggregateHeader(source_id int64, domain_id int32, id int64, version int32) AggregateHeaderData {
	return AggregateHeaderData{
		SourceId: source_id,
		DomainId: domain_id,
		Id:       id,
		Version:  version,
	}
}

func NewAggregate(source_id int64, domain_id int32, id int64, version int32, state Serializable) Aggregate {
	state_data, err := state.Serialize(state)
	if err != nil {
		state_data = make([]byte, 0)
	}
	buffer := new(bytes.Buffer)
	binary.Write(buffer, binary.BigEndian, source_id)
	binary.Write(buffer, binary.BigEndian, domain_id)
	binary.Write(buffer, binary.BigEndian, id)
	binary.Write(buffer, binary.BigEndian, version)
	binary.Write(buffer, binary.BigEndian, state_data)
	return AggregateBodyData{buffer.Bytes()}
}

// Aggregate is a structured header describing the UUId of an aggregate instance
type AggregateHeaderData struct {
	// source provides a location to store a 8 byte hash to identify the system of record
	SourceId int64 `datastore:",noindex" json:"_source"`
	// domain is the type of aggregate (type is semantically equivalent to doman)
	DomainId int32 `datastore:",noindex" json:"_domain"`
	// id is an [application / domain] unique identifier for the aggregate instance
	// and should never be duplicated within that partition
	Id int64 `json:"_id"`
	// version establishes a deterministic order for events
	Version int32 `datastore:",noindex" json:"_ver"`
}

// GetSourceId returns the source (or system of record) of this aggregate
func (a AggregateHeaderData) GetSourceId() int64 {
	return a.SourceId
}

// GetDomainId returns the domain (or aggregate type) of this aggregate
func (a AggregateHeaderData) GetDomainId() int32 {
	return a.DomainId
}

// GetId returns the id of the aggregate which is unique within the
// partition provided by the combination of application and domain
func (a AggregateHeaderData) GetId() int64 {
	return a.Id
}

// GetVersion returns the number of events in this aggregates state
func (a AggregateHeaderData) GetVersion() int32 {
	return a.Version
}

// GetUUID returns the unique identifier for this aggregate reference
func (a AggregateHeaderData) GetUUID() []byte {
	buffer := new(bytes.Buffer)
	binary.Write(buffer, binary.BigEndian, a.SourceId)
	binary.Write(buffer, binary.BigEndian, a.DomainId)
	binary.Write(buffer, binary.BigEndian, a.Id)
	binary.Write(buffer, binary.BigEndian, a.Version)
	return buffer.Bytes()
}

// Body masks the container to allow consistent interface with body data
func (a AggregateHeaderData) Body() AggregateHeaderData {
	return a
}

// String returns the string representation of the aggregate
func (a AggregateHeaderData) String() string {
	return fmt.Sprintf("%X|%X|%X|%X", uint64(a.SourceId), uint32(a.DomainId), uint64(a.Id), uint32(a.Version))
}

type AggregateBodyData struct {
	Data []byte `json:"_data"`
}

func (a AggregateBodyData) GetSourceId() int64 {
	r, _ := binary.Varint(a.Data[:8])
	return int64(r)
}

func (a AggregateBodyData) GetDomainId() int32 {
	r, _ := binary.Varint(a.Data[8:12])
	return int32(r)
}

func (a AggregateBodyData) GetId() int64 {
	r, _ := binary.Varint(a.Data[12:20])
	return r
}

func (a AggregateBodyData) GetVersion() int32 {
	r, _ := binary.Varint(a.Data[20:24])
	return int32(r)
}

// GetUUID returns the unique identifier for this aggregate reference
func (a AggregateBodyData) GetUUID() []byte {
	return a.Data[:HeaderBytes]
}

func (a AggregateBodyData) Body() AggregateHeaderData {
	buffer := bytes.NewBuffer(a.Data)
	source_id, _ := binary.Varint(buffer.Next(8))
	domain_id, _ := binary.Varint(buffer.Next(4))
	id, _ := binary.Varint(buffer.Next(8))
	version, _ := binary.Varint(buffer.Next(4))
	return AggregateHeaderData{source_id, int32(domain_id), id, int32(version)}
}

// String returns the string representation of the aggregate
func (a AggregateBodyData) String() string {
	return fmt.Sprintf("%X|%X|%X|%X", uint64(a.GetSourceId()), uint32(a.GetDomainId()), uint64(a.GetId()), uint32(a.GetVersion()))
}

func (a AggregateBodyData) GetData() []byte {
	return a.Data[HeaderBytes:]
}

func (a AggregateBodyData) GetBytes() []byte {
	return a.Data
}

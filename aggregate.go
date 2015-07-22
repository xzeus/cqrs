package cqrs

import (
//"bytes"
//"encoding/binary"
//"fmt"
)

/*
// HostDef is a structured header describing the location component of the an aggregate definition
type HostDef struct {
	// host represents a logical or physical boundary
	HostId int64 `datastore:",noindex" json:"hst"`
	// source provides a location to identify the system of record, data storage
	SourceId int64 `datastore:",noindex" json:"src"`

	// tenant represents an implicit partition within a domain to enforce data segregation
	TenantId int64 `datastore:",noindex" json:"ten"`
} // 24 bits / host
*/

// AggregateDef is a structured header describing the identity component of an aggregate instance
type AggregateDef struct {
	// bounded context is a logical container around one or more domain instances
	BoundedContextId int64 `json:"bc"`
	// domain is the type of aggregate (type is semantically equivalent to doman)
	DomainId int64 `json:"dom"`
	// id is an [application / domain] unique identifier for the aggregate instance
	// and should never be duplicated within that partition
	Id int64 `json:"id"`
	// version establishes a deterministic order for events
	Version int64 `json:"ver"`
}

type MessageDef struct {
	// clock is a bc scoped lamport incrementing counter to establish fuzzy causal ordering
	Clock int64 `json:"clk"`
	// MessageType is a domain unique identifier for the type of
	// message which captures the semantic intent of the command or event
	MessageTypeId MessageTypeId `json:"typ"`
	// Contains the message payload
	Payload []byte `json:"dat"`
}

// tiemstamp is a unix timestamp from the processing server (*unreliable for causation)
//Timestamp int64 `datastore:",noindex" json:"tim"`

/*
type MetadataDef struct {
	// Expresses a non integer identifier which is used with consistant hashing at commit
	// to establish a strictly identified location for the aggregate at rest
	//Key []byte `datastore:",noindex" json:"key, omitempty"`
	// Establishes causation for a given message
	Origin []AggregateDef `datastore:",noindex" json:"org, omitempty"`
	// Enables authorization delegation across domains and bounded contexts
	Tokens [][]byte `datastore:",noindex" json:"tok, omitempty"`

}
*/

/*
// Defines the identity of the message
	Header AggregateDef
*/

//MetadataDef
// [  ]
/*
func (m *MessageDef) String() string {
	var msg string
	if m.MessageType.IsCommand() {
		msg = fmt.Sprintf("CMD \033[0;106;90m %s/\033[1;30m%s \033[0;49;37m  [ %X v:%d ]\033[0;49;39m",
			h.domain.Name(),
			h.domain.MessageName(h.command_payload),
			uint64(id),
			ver)
	} else {

	}
	for i, o := range h.command.GetOrigin() {
		oid := o.GetId()
		ov := o.GetVersion()
		od := Meta().Domains[o.GetDomainId()].Domain
		oname := od.Name()
		h.deps.Logger().Infof("\t\033[90m ORIG [ %d ] [ %s - %X v:%d ]\033[0;49;39m", i, oname, uint64(oid), ov)
	}
}


const (
	KIND_EVENT   = "EVT"
	KIND_COMMAND = "CMD"
)

func (s *DomainCacheDef) PrintMessage(deps ioc.Dependencies, e cqrs.Message) {
	id := e.GetId()
	v := e.GetVersion()
	deps.Logger().Infof("%s \033[36m%s/\033[96m%s\033[37m  [ %X v:%d ]\033[0;49;39m", s.MessageKind, s.DomainName, s.MessageName, uint64(id), v)
	for i, o := range e.GetOrigin() {
		oid := o.GetId()
		ov := o.GetVersion()
		od := domains.Meta().Domains[o.GetDomainId()].Domain
		oname := od.Name()
		deps.Logger().Infof("\t\033[90m ORIG [ %d ] [ %s - %X v:%d ]\033[0;49;39m", i, oname, uint64(oid), ov)
	}
}
*/

/*
// NewAggregateHeaderData creates an aggregate instance with UUId derived from the provided values
func NewAggregate(source_id int64, domain_id int32, id int64, version int32) AggregateDef {
	return AggregateDef{
		SourceId: source_id,
		DomainId: domain_id,
		Id:       id,
		Version:  version,
	}
}
*/

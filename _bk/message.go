package cqrs

import (
	"fmt"
)

func NewMessage(id int64, version int32, timestamp int64, origins []AggregateHeader, payload MessageDefiner) Message {
	domain := payload.Domain()
	source_id := domain.SourceId()
	domain_id := domain.Id()
	message_type := domain.MessageType(payload)
	data := Must(payload).MustSerialize(payload)
	if origins == nil {
		origins = NoOrigin
	}
	l := len(origins)
	d := make([]AggregateHeaderData, l, l)
	for i, o := range origins {
		d[i] = o.Body()
	}
	return &MessageData{
		Aggregate:   NewAggregateHeader(source_id, domain_id, id, version),
		Origin:      d,
		Timestamp:   timestamp,
		MessageType: message_type,
		Data:        data,
	}
}

type MessageData struct {
	//
	Aggregate AggregateHeaderData `json:"_agg"`
	//
	Origin []AggregateHeaderData `datastore:",noindex" json:"_orig"`
	//
	Timestamp int64 `json:"_ts"`
	// MessageType is an [ application / domain ] unique identifier for the type of
	// message which captures the semantic intent of the command or event
	MessageType MessageType `datastore:",noindex" json:"_type"`
	//
	Data []byte `datastore:",noindex" json:"_data"`
}

func (msg MessageData) GetSourceId() int64 {
	return msg.Aggregate.SourceId
}

func (msg MessageData) GetDomainId() int32 {
	return msg.Aggregate.DomainId
}

func (msg MessageData) GetId() int64 {
	return msg.Aggregate.Id
}

func (msg MessageData) GetVersion() int32 {
	return msg.Aggregate.Version
}

func (msg MessageData) Body() AggregateHeaderData {
	return msg.Aggregate
}

func (msg MessageData) GetUUID() []byte {
	return msg.Body().GetUUID()
}

func (msg MessageData) String() string {
	return fmt.Sprintf("%X|%X|ID:%X|V:%X", uint64(msg.Aggregate.SourceId), uint32(msg.Aggregate.DomainId), uint64(msg.Aggregate.Id), uint32(msg.Aggregate.Version))
}

func (msg MessageData) GetOrigin() []AggregateHeader {
	l := len(msg.Origin)
	o := make([]AggregateHeader, l, l)
	for i, d := range msg.Origin {
		o[i] = d
	}
	return o
}

func (msg MessageData) GetTimestamp() int64 {
	return msg.Timestamp
}

func (msg MessageData) GetMessageType() MessageType {
	return msg.MessageType
}

func (msg MessageData) GetData() []byte {
	return msg.Data
}

func (msg MessageData) Reference() *MessageData {
	return &msg
}

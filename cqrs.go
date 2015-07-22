package cqrs

import (
	"fmt"
)

type MessageTypeId int64

func (t MessageTypeId) IsCommand() bool {
	return IsCommand(int64(t))
}

const (
	event_bit_mask   = 0x7FFFFFFFFFFFFFFF
	command_bit_mask = 0x8000000000000000
	version_bit_mask = 0x7F00000000000000
	id_bit_mask      = 0xFFFFFFFFFFFFFF
)

func MakeVersionedMessageType(type_id int64, version uint8) uint64 {
	return ((uint64(version) << 56) & version_bit_mask) | uint64(type_id&id_bit_mask)
}

// MakeVersionedCommandType provides a utility to union a command's version and
// type identifiers and masks off the leftmost bit as 1 to indicate a command
func MakeVersionedCommandType(type_id int64, version uint8) MessageTypeId {
	//log.Printf("\n\nCommand [ %d, %d ]", type_id, version)
	return MessageTypeId(int64(command_bit_mask | MakeVersionedMessageType(type_id, version)))
}

// MakeVersionedEventType provides a utility to union an event's version and
// type identifiers and masks off the leftmost bit as 0 to indicate an event
func MakeVersionedEventType(type_id int64, version uint8) MessageTypeId {
	//log.Printf("\n\nEvent [ %d, %d ]", type_id, version)
	return MessageTypeId(int64(event_bit_mask & MakeVersionedMessageType(type_id, version)))
}

func IsCommand(id int64) bool {
	return command_bit_mask&(uint64(id)) == command_bit_mask
}

type Domain interface {
	fmt.Stringer
	Uri() string
	Id() int64
	Name() string
	Version() int32
	NewAggregate() Aggregate
	MessageTypes() []MessageType
	MessageTypeDefs() []MessageTypeDef
	CommandType(MessageDefiner) (MessageType, error)
	EventType(MessageDefiner) (MessageType, error)
}

type MessageType interface {
	Domain() Domain
	New() MessageDefiner
	IsCommand() bool
	MessageTypeId() MessageTypeId
	DisplayName() string
	LowerName() string
	CanonicalName() string
	Version() uint8
	Id() int64
}

type DomainLinker interface {
	Domain() Domain
}

type MessageDefiner interface {
	DomainLinker
}

type Aggregate interface {
	Init() Aggregate
	Apply(Event) Aggregate
	Handle(Command) EventMessage
}

type Event interface {
	Message() MessageDefiner
}

type EventMessage interface{}

type Command interface {
	Message() MessageDefiner
}

type Header interface{}

type Payload interface{}

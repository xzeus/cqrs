package cqrs

import (
//"log"
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
	Uri() string
	Id() int64
	Name() string
	Version() int32
	NewAggregate() Aggregate
	MessageTypes() MessageTypeContainer
}

type MessageTypeSet interface {
	All() []MessageType
	ByInstance(interface{}) MessageType
	ByMessageTypeId(MessageTypeId) MessageType
	ByMessageTypeIds(...MessageTypeId) []MessageType
}

type MessageTypeContainer interface {
	MessageTypeSet
	Commands() MessageTypeSet
	Events() MessageTypeSet
}

type MessageType interface {
	Domain() Domain
	IsCommand() bool
	MessageTypeId() MessageTypeId
	DisplayName() string
	LowerName() string
	CanonicalName() string
	Version() uint8
	Id() int64
	NewPayload() MessageDefiner
}

type DomainLinker interface {
	Domain() Domain
}

type MessageDefiner interface {
	DomainLinker
}

type DomainMessage struct {
	MessageDefiner
}

func (__ DomainMessage) Clone() MessageDefiner { return __.Domain().MessageTypes().ByInstance(__) }

type Aggregate interface {
	Init() Aggregate
	Apply(Event) Aggregate
	Handle(Command) EventMessage
}

type Event interface {
	Message() EventMessage
}

type EventMessage interface {
}

type Command interface {
	Message() interface{}
}

type Header interface {
}

type Payload interface {
}

var BoundedContext = &BC{}

type BC struct{}

func (_ *BC) Register(t interface{}) interface{} {
	return nil
}

type CommandContext struct {
	Aggregate Aggregate
	Events    <-chan Event
	Command   Command
}

func (ctx *CommandContext) ExecAsync() <-chan EventMessage {
	respChan := make(chan EventMessage, 1)
	go func(_ctx *CommandContext) {
		for e := range ctx.Events {
			ctx.Aggregate = ctx.Aggregate.Apply(e)
		}
		respChan <- ctx.Aggregate.Handle(ctx.Command)
	}(ctx)
	return respChan
}

func (ctx *CommandContext) ExecSync() EventMessage {
	return <-ctx.ExecAsync()
}

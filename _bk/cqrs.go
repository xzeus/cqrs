package cqrs

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
)

var (
	// ErrNoSuchAggregate
	ErrNoSuchAggregate = errors.New("aggregate not persisted")
	// ErrInvalidAggregateId is used to inform a consumer when they've
	// provided an aggregate id that is not available due to either
	// overlap with an existing aggregate or domain specific command
	// handler rules
	ErrInvalidAggregateState = errors.New("invalid aggregate state")

	// ErrInvalidDomain is used to inform a consumer when they've
	// provided an aggregate that doesn't have a valid domain id that
	// the receiving service is able to process
	// * Domain is semantically equal to Aggregate Type
	ErrInvalidDomain = errors.New("invalid domain identifier")

	// ErrInvalidAggregateId is used to inform a consumer when they've
	// provided an aggregate id that is not available due to either
	// overlap with an existing aggregate or domain specific command
	// handler rules
	ErrInvalidAggregateId = errors.New("invalid aggregate identifier")

	// ErrInvalidVersion is used to inform a consumer when they've
	// provided an aggregate with a version that cannot be sync'd
	// with the current domain version
	ErrInvalidVersion = errors.New("invalid aggregate version")

	// ErrVersionOutOfBounds is used to inform a consumer when they've
	// attempted to use an aggregate version less than one
	ErrVersionOutOfBounds = errors.New("aggregate version out of bounds")

	// ErrInvalidCommandType is used to inform a consumer when they've
	// provided a command type that isn't valid for the application and
	// domain partition
	ErrInvalidCommandType = errors.New("invalid command type identifier")

	// ErrInvalidEventType is used to inform a consumer when they've
	// provided an event type that isn't valid for the application and
	// domain partition
	ErrInvalidEventType = errors.New("invalid event type identifier")

	// ErrSerializationError
	ErrSerializationError = errors.New("serialization error")

	// ErrNestedTransaction is fired by the datastore in the event of
	// an invalid attempt to nest transaction scope
	ErrNestedTransaction = errors.New("invalid use of nested transaction")
)

const (
	HeaderBytes int = 24

	// NoVersionControl is the default value to use for the version of commands
	// whose handler does not evaluate the version of the aggregate to determine
	// the validity of a command
	NoVersionControl int32 = 0

	NoVersion int32 = 0

	//
	NoAssignedTime int64 = 0
)

// NoOrigin is the default value to use for origin commands which were not
// a result of a previous event.  Used to specify no causation or initial action.
var NoOrigin = make([]AggregateHeader, 0) //NewAggregateHeader(0, 0, 0)

type MessageType int32

func (t MessageType) IsCommand() bool {
	return IsCommand(int32(t))
}

// MakeVersionedCommandType provides a utility to union a command's version and
// type identifiers and masks off the leftmost bit as 1 to indicate a command
func MakeVersionedCommandType(version uint8, type_id uint32) MessageType {
	return MessageType(int32(0x80000000 | (uint32(version) << 24 & 0x7F000000) | (type_id & 0xFFFFFF)))
}

// MakeVersionedEventType provides a utility to union an event's version and
// type identifiers and masks off the leftmost bit as 0 to indicate an event
func MakeVersionedEventType(version uint8, type_id uint32) MessageType {
	return MessageType(int32(0x7FFFFFFF&(uint32(version)<<24&0x7F000000) | (type_id & 0xFFFFFF)))
}

func IsCommand(message_type int32) bool {
	return 0x80000000&(uint32(message_type)) == 0x80000000
}

type Initable interface {
	Serializable
	Init() Initable
}

type DataContainer interface {
	GetData() []byte
}

type Keyed struct {
	JsonSerializedKey
	AggregateKey []byte `json:"aggkey" desc:"Aggregate key provides an alternative method of creating a UID within a domain"`
}

type DomainMessagesFunc func() (Domain, map[MessageType]func() MessageDefiner)

type Domain interface {
	Name() string
	Version() string
	Uri() string
	SourceId() int64
	Id() int32
	Aggregate() AggregateState
	Message(message_type MessageType) MessageDefiner
	Messages(message_type ...MessageType) map[MessageType]func() MessageDefiner
	Commands(message_type ...MessageType) map[MessageType]func() MessageDefiner
	Events(message_type ...MessageType) map[MessageType]func() MessageDefiner
	MessageType(message MessageDefiner) MessageType
	MessageName(message MessageDefiner) string
	Handler(deps interface{}, command Message) Message
	Services(MessageType) map[string]EventHandler
	// Config Functions
	DefCommandHandler(factory func(CommandHandlerDef) CommandHandlerFunc) CommandHandler
	DefEventHandler(message_type MessageType, service_key string, handler EventHandler) EventHandler
	DefService(factory func(EventHandlerDef) EventHandlerFunc, subs ...map[MessageType]func() MessageDefiner) EventHandler
	DefCommand(version uint8, id uint32, m MessageDefiner) MessageType
	DefEvent(version uint8, id uint32, m MessageDefiner) MessageType
}

type DomainDefiner interface {
	Domain() Domain
}

type CommandHandler func(interface{}, Message) Message
type CommandHandlerFactory func(interface{}, Domain, Message) CommandHandler
type CommandHandlerFunc func(AggregateHeader, AggregateState, Message, MessageDefiner)

type EventHandler func(interface{}, Message)
type EventHandlerFactory func(interface{}, Domain, Message) EventHandler
type EventHandlerFunc func(Message, MessageDefiner)

type CommandHandlerDef interface {
	// Server actions
	Exec(handler CommandHandlerFunc) Message
	// Handler Actions
	Publish(MessageDefiner, ...func(*MessageOptionsDef))
	Error(string, ...interface{})
	Assert(bool, string, ...interface{})
	// Data access
	State() AggregateState
	Command() Message
	CommandPayload() MessageDefiner
	EventOptions() MessageOptions
	EventPayload() MessageDefiner
}

type EventHandlerDef interface {
	Deps() interface{}
	Publish(MessageDefiner, ...func(*MessageOptionsDef)) Message
	Error(string, ...interface{}) Message
	Exec(EventHandlerFunc)
}

type Service interface {
	EventHandlerDef
	Handle(Message, MessageDefiner)
}

type ServiceFactory func(interface{}, Message) Service

type ServiceDefiner interface {
	DomainDefiner
	Subscriptions() map[MessageType]func() MessageDefiner
	Factory(deps interface{}, domain Domain, event Message) Service
}

// Aggregate provides a base interface for things that contain
// aggregate header information
type AggregateHeader interface {
	GetSourceId() int64
	GetDomainId() int32
	GetId() int64
	GetVersion() int32
	GetUUID() []byte // [ 16 bytes ]
	String() string
	Body() AggregateHeaderData
}

type AggregateState interface {
	Init() AggregateState
	Handle(MessageDefiner)
}

type Aggregate interface {
	AggregateHeader
	GetData() []byte
	GetBytes() []byte
}

// Message provides a base interface for all commands and events in the
// system which includes the event's aggregate header information and the
// originating event, if applicable
type Message interface {
	AggregateHeader
	GetTimestamp() int64
	GetOrigin() []AggregateHeader
	GetMessageType() MessageType
	GetData() []byte
	Reference() *MessageData
}

type MessageOptions interface {
	Id() int64
	Key() string
	Version() int32
	Timestamp() int64
}

type MessageDefiner interface {
	DomainDefiner
	Serializable
}

const DEFAULT_KEY = ""

var keyed_type = reflect.TypeOf(Keyed{})
var keyed_name = keyed_type.Field(1).Name

func ExtractKey(payload MessageDefiner) (key string) {
	payload_type := reflect.TypeOf(payload)
	ptr := reflect.Ptr == payload_type.Kind()
	if ptr {
		payload_type = payload_type.Elem()
	}
	if _, keyed := payload_type.FieldByName(keyed_name); keyed {
		payload_value := reflect.ValueOf(payload)
		if ptr {
			payload_value = payload_value.Elem()
		}
		field := payload_value.FieldByName(keyed_name)
		key = fmt.Sprintf("%s", field.Interface().([]byte))
	} else {
		key = DEFAULT_KEY
	}
	return
}

// Serializtion methods

type Serializable interface {
	Serialize(interface{}) ([]byte, error)
	Deserialize([]byte, interface{}) error
}

type SerializableKey interface {
	SerializeKey(interface{}) ([]byte, error)
	DeserializeKey([]byte, interface{}) error
}

type MustSerializable interface {
	MustSerialize(interface{}) []byte
	MustDeserialize([]byte, interface{})
}

type JsonSerialized struct{}

func (o JsonSerialized) Serialize(message interface{}) ([]byte, error) {
	return json.MarshalIndent(message, "", "")
}

func (o JsonSerialized) Deserialize(data []byte, message interface{}) error {
	return json.Unmarshal(data, message)
}

type JsonSerializedKey struct{}

func (o JsonSerializedKey) SerializeKey(message interface{}) ([]byte, error) {
	return json.MarshalIndent(message, "", "")
}

func (o JsonSerializedKey) DeserializeKey(data []byte, message interface{}) error {
	return json.Unmarshal(data, message)
}

type mustSerializable struct {
	serializer   func(interface{}) ([]byte, error)
	deserializer func([]byte, interface{}) error
}

func Must(serializable Serializable) MustSerializable {
	return mustSerializable{
		serializer:   serializable.Serialize,
		deserializer: serializable.Deserialize,
	}
}

func (must mustSerializable) MustSerialize(message interface{}) []byte {
	data, err := must.serializer(message)
	if err != nil {
		panic(err)
	}
	return data
}

func (must mustSerializable) MustDeserialize(data []byte, message interface{}) {
	err := must.deserializer(data, message)
	if err != nil {
		panic(err)
	}
}

func Extract(dest Serializable, src DataContainer) (err error) {
	defer func() {
		if recoverErr := recover(); recoverErr != nil {
			err = ErrSerializationError
		}
	}()
	if dest == nil || src == nil {
		err = ErrSerializationError
		return
	}
	data := src.GetData()
	if data == nil {
		err = ErrSerializationError
		return
	}
	if err = dest.Deserialize(data, dest); err != nil {
		err = ErrSerializationError
		return
	}
	return
}

func KeyExtract(dest SerializableKey, src DataContainer) (err error) {
	defer func() {
		if recoverErr := recover(); recoverErr != nil {
			err = ErrSerializationError
		}
	}()
	if dest == nil || src == nil {
		err = ErrSerializationError
		return
	}
	data := src.GetData()
	if data == nil {
		err = ErrSerializationError
		return
	}
	if err = dest.DeserializeKey(data, dest); err != nil {
		err = ErrSerializationError
		return
	}
	return
}

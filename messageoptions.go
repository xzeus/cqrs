package cqrs

import ()

type MessageOptionsDef struct {
	// RetryCreate int -- number of times to try a new id before failing
	id        int64
	key       string
	version   int32
	timestamp int64
}

func (def *MessageOptionsDef) Id() int64 {
	return def.id
}

func (def *MessageOptionsDef) Key() string {
	return def.key
}

func (def *MessageOptionsDef) Version() int32 {
	if def.version < 1 {
		return 1
	}
	return def.version
}

func (def *MessageOptionsDef) Timestamp() int64 {
	return def.timestamp
}

func NewMessageOptions(id int64, version int32, timestamp int64) *MessageOptionsDef {
	return &MessageOptionsDef{
		id:        id,
		key:       "",
		version:   version,
		timestamp: timestamp,
	}
}

func NewKeyedMessageOptions(key string, version int32, timestamp int64) *MessageOptionsDef {
	return &MessageOptionsDef{
		id:        0,
		key:       key,
		version:   version,
		timestamp: timestamp,
	}
}

func WithOptions(new_options MessageOptions) func(*MessageOptionsDef) {
	return func(options *MessageOptionsDef) {
		options.id = new_options.Id()
		options.key = new_options.Key()
		options.version = new_options.Version()
		options.timestamp = new_options.Timestamp()
	}
}

func OriginId(message Message) func(*MessageOptionsDef) {
	return OriginStack(message, 0)
}

func OriginStack(message Message, i int) func(*MessageOptionsDef) {
	return func(options *MessageOptionsDef) {
		if len(message.GetOrigin()) > i {
			options.id = message.GetOrigin()[i].GetId()
		}
	}
}

func OriginQueue(message Message, i int) func(*MessageOptionsDef) {
	return func(options *MessageOptionsDef) {
		l := len(message.GetOrigin())
		if l > i {
			options.id = message.GetOrigin()[l-i-1].GetId()
		}
	}
}

func Id(value int64) func(*MessageOptionsDef) {
	return func(options *MessageOptionsDef) {
		options.id = value
	}
}

func Key(value string) func(*MessageOptionsDef) {
	return func(options *MessageOptionsDef) {
		options.id = 0
		options.key = value
	}
}

func Version(value int32) func(*MessageOptionsDef) {
	return func(options *MessageOptionsDef) {
		options.version = value
	}
}

func Timestamp(value int64) func(*MessageOptionsDef) {
	return func(options *MessageOptionsDef) {
		options.timestamp = value
	}
}

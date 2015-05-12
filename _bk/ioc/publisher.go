package ioc

import (
	"github.com/xzeus/cqrs"
)

type AsyncPublishCallback func(deps Dependencies, service_key string, event_message *cqrs.MessageData)

type Publisher interface {
	Publish(cqrs.Message)
}

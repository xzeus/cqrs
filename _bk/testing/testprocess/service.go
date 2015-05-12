package testprocess

import (
	"github.com/xzeus/cqrs"
	. "github.com/xzeus/providers/mockprovider"
)

var Mock_Handle = func(header cqrs.AggregateHeader, aggregate cqrs.AggregateState, message cqrs.Message, payload cqrs.MessageDefiner) {
	panic("mockprovider.testdomain.service: Mock handle not implemented")
}

var Service = handlers.NewEventHandlerFactory(Domain,
	func(deps interface{},
		command cqrs.Message) cqrs.Service {
		return &EventHandler{handlers.NewEventHandler(deps.(handlers.Dependencies), command)}
	},
	Domain.Messages(
		E_TestEvent, // 1, 1
	),
)

type EventHandler struct {
	cqrs.EventHandlerDef
}

func (e *EventHandler) Handle(header cqrs.AggregateHeader, aggregate cqrs.AggregateState, message cqrs.Message, payload cqrs.MessageDefiner) {
	Mock_Handle(header, aggregate, message, payload)
}

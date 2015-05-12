package domains

import (
	"github.com/xzeus/cqrs"
	"github.com/xzeus/cqrs/ioc"
)

type eventHandlerDef struct {
	deps          ioc.Dependencies
	domain        cqrs.Domain
	event         cqrs.Message
	event_payload cqrs.MessageDefiner
}

var default_command_options = cqrs.NewMessageOptions(0, 1, int64(0))

func NewEventHandler(deps ioc.Dependencies, domain cqrs.Domain, event cqrs.Message) cqrs.EventHandlerDef {
	event_domain := Meta().Domains[event.GetDomainId()].Domain
	handler := &eventHandlerDef{
		deps:          deps,
		domain:        domain,
		event:         event,
		event_payload: event_domain.Message(event.GetMessageType()),
	}
	if err := cqrs.Extract(handler.event_payload, event); err != nil {
		panic("Shouldn't ever receive an event that can't be unserialized")
	}
	return handler
}

func (h *eventHandlerDef) Deps() interface{} {
	return h.deps
}

func (h *eventHandlerDef) Exec(handler cqrs.EventHandlerFunc) {
	handler(h.event, h.event_payload)
}

func (h *eventHandlerDef) Publish(command_payload cqrs.MessageDefiner, options ...func(*cqrs.MessageOptionsDef)) cqrs.Message {
	var domain = command_payload.Domain()
	var command_options = cqrs.NewMessageOptions(0, 0, int64(0))
	for _, modifier := range options {
		modifier(command_options)
	}
	l := len(h.event.GetOrigin()) + 1
	o := make([]cqrs.AggregateHeader, l, l)
	o[0] = h.event.Body()
	for i, d := range h.event.GetOrigin() {
		o[i+1] = d
	}
	command := cqrs.NewMessage(
		command_options.Id(),
		command_options.Version(),
		command_options.Timestamp(),
		o,
		command_payload)
	return domain.Handler(h.deps, command)
}

func (h *eventHandlerDef) Error(message string, args ...interface{}) cqrs.Message {
	command, options := h.deps.Exception().Error(message, args...)
	return h.Publish(command, cqrs.WithOptions(options))
}

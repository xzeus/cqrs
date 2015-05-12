package domains

import (
	"fmt"
	"github.com/xzeus/cqrs"
	. "github.com/xzeus/cqrs/ioc"
)

type commandHandlerDef struct {
	deps            Dependencies
	domain          cqrs.Domain
	header          cqrs.AggregateHeader
	state           cqrs.AggregateState
	command         cqrs.Message
	command_payload cqrs.MessageDefiner
	event_options   *cqrs.MessageOptionsDef //cqrs.MessageOptions
	event_payload   cqrs.MessageDefiner
	event_loader    func() ([]cqrs.Message, error)
	event_append    func() (cqrs.Message, error)
}

var default_event_options = cqrs.NewMessageOptions(0, 1, int64(0))

func NewCommandHandler(deps Dependencies, domain cqrs.Domain, command cqrs.Message) cqrs.CommandHandlerDef {
	h := &commandHandlerDef{
		deps:            deps,
		domain:          domain,
		command:         command,
		command_payload: domain.Message(command.GetMessageType()),
		event_options:   default_event_options,
	}
	if err := cqrs.Extract(h.command_payload, command); err != nil {
		h.event_payload, h.event_options = deps.Exception().Error("Unable to extract command payload: [ %s ]", err)
		return h // If extract fails return an error event when run
	}
	var key string
	if key = cqrs.ExtractKey(h.command_payload); key == cqrs.DEFAULT_KEY {
		h.event_loader = func() ([]cqrs.Message, error) {
			return deps.EventStore().GetAggregateEvents(h.domain.Id(), command.GetId(), 0)
		}
	} else { // Key based message
		k := []byte(key)
		h.event_loader = func() ([]cqrs.Message, error) {
			return deps.EventStore().GetKeyedAggregateEvents(h.domain.Id(), k, 0)
		}
	}
	var events []cqrs.Message
	var err error
	if events, err = h.event_loader(); err != nil {
		if events, err = h.event_loader(); err != nil { // Single retry
			h.Error("Error loading aggregate events [ %s ]", err)
			return h
		}
	} // Event load success, hydrate the aggregate
	version := int32(len(events)) + 1
	timestamp := int64(0)
	if key == cqrs.DEFAULT_KEY {
		h.event_options = cqrs.NewMessageOptions(command.GetId(), version, timestamp)
	} else {
		id := deps.Crypto().Hash64([]byte(key))
		h.event_options = cqrs.NewMessageOptions(id, version, timestamp)
	}
	state := h.domain.Aggregate().Init()
	var payload cqrs.MessageDefiner
	for i, event := range events {
		payload = h.domain.Message(event.GetMessageType())
		if err := cqrs.Extract(payload, event); err != nil {
			h.Error("Error extracting event[ %d ] [ %#v ]", i, event)
			return h
		} // Event payload extracted, apply it to the state
		state.Handle(payload)
	}
	h.state = state
	h.header = cqrs.NewAggregateHeader(h.domain.SourceId(), h.domain.Id(), h.event_options.Id(), int32(len(events)))
	return h
}

func (h *commandHandlerDef) Exec(handler cqrs.CommandHandlerFunc) (result cqrs.Message) {
	defer func() { // Best effort to commit result
		var err error
		if h.event_payload == nil {
			h.Error("Error event not published in handler", err)
		}
		var key string
		if key = cqrs.ExtractKey(h.event_payload); key == cqrs.DEFAULT_KEY {
			h.event_append = func() (cqrs.Message, error) {
				id := h.event_options.Id()
				ver := h.event_options.Version()

				h.deps.Logger().Infof("CMD \033[0;106;90m %s/\033[1;30m%s \033[0;49;37m  [ %X v:%d ]\033[0;49;39m", h.domain.Name(), h.domain.MessageName(h.command_payload), uint64(id), ver)
				for i, o := range h.command.GetOrigin() {
					oid := o.GetId()
					ov := o.GetVersion()
					od := Meta().Domains[o.GetDomainId()].Domain
					oname := od.Name()
					h.deps.Logger().Infof("\t\033[90m ORIG [ %d ] [ %s - %X v:%d ]\033[0;49;39m", i, oname, uint64(oid), ov)
				}
				return h.deps.EventStore().AppendEvent(id, ver, h.command.GetOrigin(), h.event_payload)
			}
		} else { // Key based message
			k := []byte(key)
			h.event_append = func() (cqrs.Message, error) {
				return h.deps.EventStore().AppendKeyedEvent(k, h.command.GetOrigin(), h.event_payload)
			}
		}
		if result, err = h.event_append(); err != nil {
			h.deps.Logger().Infof("Trigger error for [\n%s\n]", h.EventPayload())
			h.ForceError("Error appending event [ %s ]", err)
			if result, err = h.event_append(); err != nil { // Try to append the failure message
				panic(err) // Multiple append errors
			}
		} // Exec publisher if defined
		h.deps.Publisher().Publish(result)
	}() // Check for errors from constructor
	if h.event_payload != nil { // Enforce single publish maxim
		return
	}
	handler(h.header, h.state, h.command, h.command_payload)
	return
}

func (h *commandHandlerDef) Publish(event_payload cqrs.MessageDefiner, options ...func(*cqrs.MessageOptionsDef)) {
	if h.event_payload != nil { // Enforce single publish maxim
		return
	}
	for _, modifier := range options {
		modifier(h.event_options)
	}
	h.event_payload = event_payload
}

func (h *commandHandlerDef) Error(message string, args ...interface{}) {
	if h.event_payload != nil { // Enforce single publish maxim
		return
	}
	h.event_payload, h.event_options = h.deps.Exception().Error(message, args...)
	h.deps.Logger().Infof("\n\n***\tCalled error: [\n%#v\n]", h.event_payload, h.event_options)
}

func (h *commandHandlerDef) ForceError(message string, args ...interface{}) {
	h.event_payload, h.event_options = h.deps.Exception().Error(message, args...)
	h.deps.Logger().Infof("\n\n***\tCalled force error: [\n%#v\n]", h.event_payload, h.event_options)
}

func (h *commandHandlerDef) Assert(predicate bool, message string, args ...interface{}) {
	if predicate || h.event_payload != nil { // Enforce single publish maxim
		return
	} // Assertion failed and no prior event
	message = fmt.Sprintf(message, args)
	h.deps.Logger().Infof("Assert failed [ %s ]", message)
	h.Error("Assert failed: " + message)
}

func (h *commandHandlerDef) State() cqrs.AggregateState {
	return h.state
}

func (h *commandHandlerDef) Command() cqrs.Message {
	return h.command
}

func (h *commandHandlerDef) CommandPayload() cqrs.MessageDefiner {
	return h.command_payload
}

func (h *commandHandlerDef) EventOptions() cqrs.MessageOptions {
	return h.event_options
}

func (h *commandHandlerDef) EventPayload() cqrs.MessageDefiner {
	return h.event_payload
}

package domains

import (
	"fmt"
	"github.com/vizidrix/crypto"
	"github.com/xzeus/cqrs"
	"github.com/xzeus/cqrs/ioc"
	"log"
	"reflect"
	"runtime/debug"
)

var meta = SourceMetadata{
	SourceUri: "",
	Domains:   make(map[int32]*DomainMetadata),
}

func Meta() SourceMetadata {
	return meta
}

type DomainImpl struct {
	uri             string
	id              int32
	domain          cqrs.Domain
	command_handler cqrs.CommandHandler
	event_handlers  map[cqrs.MessageType]map[string]cqrs.EventHandler
	factory         func() cqrs.AggregateState
	factory_map     map[cqrs.MessageType]func() cqrs.MessageDefiner
	type_map        map[string]cqrs.MessageType
}

type SourceMetadata struct {
	SourceUri string
	SourceId  int64
	Domains   map[int32]*DomainMetadata
}

func (m SourceMetadata) SetSourceUri(uri string) {
	if meta.SourceUri != "" {
		panic("Source already established")
	}
	meta.SourceUri = uri
	meta.SourceId = crypto.New64a([]byte(uri))
}

func (m SourceMetadata) GetSourceUri() string {
	return meta.SourceUri
}

func (m SourceMetadata) GetSourceId() int64 {
	return meta.SourceId
}

// TODO: Add info about what services handle this domains's events in the current context
type DomainMetadata struct {
	Domain   cqrs.Domain
	Uri      string
	Commands map[cqrs.MessageType]*MessageMetadata
	Events   map[cqrs.MessageType]*MessageMetadata
}

type MessageMetadata struct {
	Name    string
	Factory func() cqrs.MessageDefiner
}

var noop_command_handler = func(interface{}, cqrs.Message) cqrs.Message {
	panic("command handler not defined")
}

func NewDomain(domain cqrs.DomainDefiner, uri string, a cqrs.AggregateState, configs ...func(cqrs.Domain)) cqrs.Domain {
	id := crypto.New32a([]byte(uri))
	v := reflect.ValueOf(a).Elem().Type()
	f := func() cqrs.AggregateState {
		return reflect.New(v).Interface().(cqrs.AggregateState)
	}

	domain_impl := &DomainImpl{
		uri:             uri,
		id:              id,
		domain:          domain.Domain(),
		command_handler: noop_command_handler,
		event_handlers:  make(map[cqrs.MessageType]map[string]cqrs.EventHandler),
		factory:         f,
		factory_map:     make(map[cqrs.MessageType]func() cqrs.MessageDefiner),
		type_map:        make(map[string]cqrs.MessageType),
	}

	meta.Domains[id] = &DomainMetadata{
		Domain:   domain_impl,
		Uri:      uri,
		Commands: map[cqrs.MessageType]*MessageMetadata{},
		Events:   map[cqrs.MessageType]*MessageMetadata{},
	}

	return domain_impl
}

func (s *DomainImpl) DefCommandHandler(w func(cqrs.CommandHandlerDef) cqrs.CommandHandlerFunc) cqrs.CommandHandler {
	s.command_handler = func(deps interface{}, c cqrs.Message) cqrs.Message {
		h := NewCommandHandler(deps.(ioc.Dependencies), s, c)
		return h.Exec(w(h))
	}
	return s.command_handler
}

func (s *DomainImpl) Services(message_type cqrs.MessageType) map[string]cqrs.EventHandler {
	return s.event_handlers[message_type]
}

func (s *DomainImpl) DefEventHandler(message_type cqrs.MessageType, service_key string, handler cqrs.EventHandler) cqrs.EventHandler {
	if existing, found := s.event_handlers[message_type]; found { // Append to exiting
		existing[service_key] = handler
	} else { // No existing handlers for this message type
		services := make(map[string]cqrs.EventHandler)
		services[service_key] = handler
		s.event_handlers[message_type] = services
	}
	return handler
}

func (s *DomainImpl) DefService(f func(cqrs.EventHandlerDef) cqrs.EventHandlerFunc, subs ...map[cqrs.MessageType]func() cqrs.MessageDefiner) cqrs.EventHandler {
	defer func() {
		if recoverErr := recover(); recoverErr != nil {
			log.Printf("\n\n\n***\n***\n***\n\nRecovered in [ %s ]\n\n\n\n", "Domain/DefService")
			log.Printf("\n\n\n[ %v ]\n\n\n\n", recoverErr)
			log.Printf("\n\n\n%s\n\n\n***\n***\n***\n\n\n", debug.Stack())

		} // Attempt to write server fault failed
	}()
	var eh cqrs.EventHandler
	eh = func(deps interface{}, event cqrs.Message) {
		handler := NewEventHandler(deps.(ioc.Dependencies), s, event)
		handler.Exec(f(handler))
	}
	for _, sub := range subs { // For each sub in the list
		for t, f_msg := range sub { // Link to each message type
			m := f_msg()                      // Make MessageDefiner to access it's domain
			d := m.Domain()                   // Use refernece to bind handler to target domain
			d.DefEventHandler(t, s.Uri(), eh) // Make the pointer back to this handler
		} // Should be able to locate target domain by message type
	} // With target domain, should be able to get map of url to event handler
	return eh
}

func (s *DomainImpl) def(t cqrs.MessageType, m cqrs.MessageDefiner) {
	v := reflect.ValueOf(m).Elem().Type()
	f := func() cqrs.MessageDefiner {
		m := reflect.New(v).Interface().(cqrs.MessageDefiner)
		return m
	}
	mm := &MessageMetadata{
		Name:    v.Name(),
		Factory: f,
	}
	l := meta.Domains[m.Domain().Id()]
	if t.IsCommand() {
		l.Commands[t] = mm
	} else {
		l.Events[t] = mm
	}
	s.factory_map[t] = f
	s.type_map[mm.Name] = t
}

func (s *DomainImpl) DefCommand(v uint8, id uint32, m cqrs.MessageDefiner) cqrs.MessageType {
	t := cqrs.MakeVersionedCommandType(v, id)
	s.def(t, m)
	return t
}

func (s *DomainImpl) DefEvent(v uint8, id uint32, m cqrs.MessageDefiner) cqrs.MessageType {
	t := cqrs.MakeVersionedEventType(v, id)
	s.def(t, m)
	return t
}

func (s *DomainImpl) Name() string {
	return s.uri // TODO: hack out second to last url param
}

func (s *DomainImpl) Version() string {
	return s.uri // TODO: hack out last url param
}

func (s *DomainImpl) SourceUri() string {
	return Meta().GetSourceUri()
}

func (s *DomainImpl) SourceId() int64 {
	return Meta().GetSourceId()
}

func (s *DomainImpl) Uri() string {
	return s.uri
}

func (s *DomainImpl) Id() int32 {
	return s.id
}

func (s *DomainImpl) Aggregate() cqrs.AggregateState {
	return s.factory()
}

func (s *DomainImpl) Message(message_type cqrs.MessageType) cqrs.MessageDefiner {
	if f, ok := s.factory_map[message_type]; ok {
		return f()
	}
	panic(fmt.Sprintf("requested message type [ %X ] from domain [ %s ]", uint32(message_type), s.Uri()))
}

func (s *DomainImpl) Messages(message_types ...cqrs.MessageType) map[cqrs.MessageType]func() cqrs.MessageDefiner {
	var result = make(map[cqrs.MessageType]func() cqrs.MessageDefiner)
	if len(message_types) == 0 { // If empty then copy all entries
		for message_type, factory := range s.factory_map {
			result[message_type] = factory
		}
	} else { // Otherwise provide just the specified list
		for _, message_type := range message_types {
			result[message_type] = s.factory_map[message_type]
		}
	}
	return result
}

func (s *DomainImpl) Commands(message_types ...cqrs.MessageType) map[cqrs.MessageType]func() cqrs.MessageDefiner {
	m := s.Messages(message_types...)
	for t, _ := range m {
		if !t.IsCommand() {
			delete(m, t)
		}
	}
	return m
}

func (s *DomainImpl) Events(message_types ...cqrs.MessageType) map[cqrs.MessageType]func() cqrs.MessageDefiner {
	m := s.Messages(message_types...)
	for t, _ := range m {
		if t.IsCommand() {
			delete(m, t)
		}
	}
	return m
}

func (s *DomainImpl) MessageType(m cqrs.MessageDefiner) cqrs.MessageType {
	return s.type_map[s.MessageName(m)]
}

func (s *DomainImpl) MessageName(m cqrs.MessageDefiner) string {
	t := reflect.TypeOf(m)
	n := t.Name()
	if n == "" {
		n = t.Elem().Name()
	}
	return n
}

func (s *DomainImpl) Handler(deps interface{}, command cqrs.Message) cqrs.Message {
	return s.command_handler(deps, command)
}

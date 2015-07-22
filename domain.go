package cqrs

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/vizidrix/crypto"
	"log"
	"reflect"
	"strconv"
	"strings"
)

var (
	ErrInvalidDomainUri      = errors.New("invalid domain uri")
	ErrAggregateNotProvided  = errors.New("invalid definition aggregate not provided")
	ErrCommandsNotProvided   = errors.New("invalid definition commands not provided")
	ErrEventsNotProvided     = errors.New("invalid definition events not provided")
	ErrInvalidAggregate      = errors.New("invalid aggregate in definition")
	ErrInvalidMessage        = errors.New("invalid message in definition")
	ErrInvalidMessageVersion = errors.New("invalid message version in defintion")
	ErrCommandTypeNotFound   = errors.New("command type not found in domain")
	ErrEventTypeNotFound     = errors.New("event type not found in domain")
)

const slash = "/"

func DomainFromMeta(meta interface{}) Domain {
	d := MustCompile(meta)
	log.Printf("%s", d)
	return d
}

type DomainDef struct {
	uri              string
	id               int64
	name             string
	version          int32
	proto            func() Aggregate
	numcommand       int
	messagetypes     []MessageType
	messagetype_defs []MessageTypeDef
	command_lookup   map[reflect.Type]*MessageTypeDef
	event_lookup     map[reflect.Type]*MessageTypeDef
}

func (d *DomainDef) String() string {
	b := &bytes.Buffer{}
	b.WriteString(fmt.Sprintf("\n\nDOMAIN [ %s v%d ] @ [ %s ] id[ %X ]", d.name, d.version, d.uri, uint64(d.id)))
	for i, m := range d.MessageTypes() {
		b.WriteString(fmt.Sprintf("\n\t[%d]: %s", i, m))
	}
	return b.String()
}

func (d *DomainDef) Uri() string {
	return d.uri
}

func (d *DomainDef) Id() int64 {
	return d.id
}

func (d *DomainDef) Name() string {
	return d.name
}

func (d *DomainDef) Version() int32 {
	return d.version
}

func (d *DomainDef) NewAggregate() Aggregate {
	return d.proto()
}

func (d *DomainDef) MessageTypes() []MessageType {
	return d.messagetypes
}

func (d *DomainDef) MessageTypeDefs() []MessageTypeDef {
	return d.messagetype_defs
}

func (d *DomainDef) CommandType(m MessageDefiner) (t MessageType, err error) {
	var ok bool
	if t, ok = d.command_lookup[reflect.TypeOf(m)]; !ok {
		err = ErrCommandTypeNotFound
	}
	return
}

func (d *DomainDef) EventType(m MessageDefiner) (t MessageType, err error) {
	var ok bool
	if t, ok = d.event_lookup[reflect.TypeOf(m)]; !ok {
		err = ErrEventTypeNotFound
	}
	return
}

var (
	i_aggregate      = reflect.TypeOf((*Aggregate)(nil)).Elem()
	i_messagedefiner = reflect.TypeOf((*MessageDefiner)(nil)).Elem()
)

func MustCompile(meta interface{}) Domain {
	d, err := Compile(meta)
	if err != nil {
		panic(err)
	}
	return d
}

func Compile(meta interface{}) (Domain, error) {
	var err error
	t_meta := reflect.TypeOf(meta)
	if t_meta.Kind() == reflect.Ptr {
		t_meta = t_meta.Elem()
	}
	f_aggregate, found := t_meta.FieldByName("Aggregate")
	if !found {
		return nil, ErrAggregateNotProvided
	} // Parse aggregate info
	t_aggregate := f_aggregate.Type
	if t_aggregate.Kind() == reflect.Ptr {
		t_aggregate = t_aggregate.Elem()
	}
	if !t_aggregate.Implements(i_aggregate) {
		return nil, ErrInvalidAggregate
	}
	uri := t_meta.PkgPath()
	if uri_tag := f_aggregate.Tag.Get("uri"); uri_tag != "" {
		uri = uri_tag
	}
	tokens := strings.Split(uri, "/")
	tokens_l := len(tokens)
	if tokens_l < 3 || strings.Index(tokens[tokens_l-1], "v") != 0 {
		return nil, ErrInvalidDomainUri
	}
	var uri_v int64
	uri_v, err = strconv.ParseInt(tokens[tokens_l-1][1:], 10, 32)
	if err != nil {
		return nil, ErrInvalidDomainUri
	} // Setup message storage
	type_by_name := func(n string) (reflect.Type, error) {
		if s, f := t_meta.FieldByName(n); !f {
			return nil, ErrInvalidMessage
		} else {
			if s.Type.Kind() != reflect.Struct || s.Type.NumField() == 0 {
				return nil, ErrInvalidMessage
			}
			return s.Type, nil
		}
	}
	var t_commands, t_events reflect.Type
	if t_commands, err = type_by_name("Commands"); err != nil {
		return nil, ErrCommandsNotProvided
	}
	if t_events, err = type_by_name("Events"); err != nil {
		return nil, ErrEventsNotProvided
	}
	c_l := t_commands.NumField()
	e_l := t_events.NumField()
	m_l := c_l + e_l
	defs := make([]MessageTypeDef, m_l, m_l)
	types := make([]MessageType, m_l, m_l)
	command_lookup := make(map[reflect.Type]*MessageTypeDef)
	event_lookup := make(map[reflect.Type]*MessageTypeDef)
	id := crypto.CrcHash64([]byte(uri))
	name := tokens[tokens_l-2]
	proto := func() Aggregate {
		return reflect.New(t_aggregate).Interface().(Aggregate)
	}
	d := &DomainDef{
		uri:     uri,
		id:      id,
		name:    name,
		version: int32(uri_v),
		proto:   proto,
	}
	var j int = 0
	parser := func(t reflect.Type, m func(id int64, version uint8) MessageTypeId) error {
		for i := 0; i < t.NumField(); i++ { // Fields under header should all be message defs
			sf := t.FieldByIndex([]int{i})
			t_sf := sf.Type
			if t_sf.Kind() == reflect.Ptr { // Convention is to use a pointer reference
				t_sf = t_sf.Elem()
			}
			name := t_sf.Name()
			lower := strings.ToLower(name)
			var v uint8 = 1
			if v_tag := sf.Tag.Get("v"); v_tag != "" {
				if v_parsed, err := strconv.ParseUint(v_tag, 10, 8); err != nil {
					return ErrInvalidMessageVersion
				} else {
					v = uint8(v_parsed)
				}
			} else { // Version tag overrides all else
				if i := strings.LastIndex(lower, "_v"); i > 0 {
					if v_parsed, err := strconv.ParseUint(lower[i+2:], 10, 8); err != nil {
						return ErrInvalidMessageVersion
					} else {
						v = uint8(v_parsed)
					}
				}
			}
			canonical := fmt.Sprintf("%s_v%d", lower, v)
			id := crypto.CrcHash64([]byte(canonical))
			if t_sf == nil || !reflect.PtrTo(t_sf).Implements(i_messagedefiner) {
				return ErrInvalidMessage
			}
			p := func() MessageDefiner {
				return reflect.New(t_sf).Interface().(MessageDefiner)
			}
			m_id := m(id, v)
			iscommand := m_id.IsCommand()
			def := MessageTypeDef{
				domain:        d,
				iscommand:     iscommand,
				messagetypeid: m_id,
				displayname:   name,
				lowername:     lower,
				canonicalname: canonical,
				version:       v,
				id:            id,
				proto:         p,
			}
			defs[j] = def
			types[j] = def
			if iscommand {
				command_lookup[t_sf] = &def
			} else {
				event_lookup[t_sf] = &def
			}
			j++
		}
		return nil
	}
	if err = parser(t_commands, MakeVersionedCommandType); err != nil {
		return nil, err
	}
	if err = parser(t_events, MakeVersionedEventType); err != nil {
		return nil, err
	}
	d.messagetype_defs = defs
	d.messagetypes = types
	d.command_lookup = command_lookup
	d.event_lookup = event_lookup
	return d, nil
}

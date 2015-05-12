package cqrs

import (
	"errors"
	"fmt"
	"github.com/vizidrix/crypto"
	//"log"
	"reflect"
	"sort"
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
)

func NewMessageTypeContainer(defs ...*MessageTypeDef) MessageTypeContainer {
	sort.Sort(ByPurposeAndName(defs))
	l := len(defs)
	r := &MessageTypeContainerDef{
		defs:          defs,
		command_count: 0,
		cache_all:     make([]MessageType, l, l),
		cache_type:    make(map[reflect.Type]MessageType),
		cache_typeid:  make(map[MessageTypeId]MessageType),
	}
	for i, m := range defs {
		r.cache_all[i] = m
		r.cache_type[reflect.TypeOf(m.NewPayload())] = m
		r.cache_typeid[m.MessageTypeId()] = m
		if r.command_count == 0 {
			if !m.iscommand { // Swap to event parsing
				r.command_count = i
			}
		}
	}
	return r
}

type DomainDef struct {
	uri                  string
	id                   int64
	name                 string
	version              int32
	proto                func() Aggregate
	messagetypecontainer MessageTypeContainer
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

func (d *DomainDef) MessageTypes() MessageTypeContainer {
	return d.messagetypecontainer
}

type MessageTypeSetDef struct {
	iscommands bool
	parent     *MessageTypeContainerDef
}

func (set *MessageTypeSetDef) All() []MessageType {
	if set.iscommands {
		return set.parent.cache_all[:set.parent.command_count]
	} else {
		return set.parent.cache_all[set.parent.command_count:]
	}
}

func (set *MessageTypeSetDef) ByInstance(t interface{}) (r MessageType) {
	r = set.parent.ByInstance(t)
	if set.iscommands != r.IsCommand() {
		return nil
	}
	return
}

func (set *MessageTypeSetDef) ByMessageTypeId(id MessageTypeId) (r MessageType) {
	r = set.parent.ByMessageTypeId(id)
	if set.iscommands != r.IsCommand() {
		return nil
	}
	return
}

func (set *MessageTypeSetDef) ByMessageTypeIds(ids ...MessageTypeId) (r []MessageType) {
	l := len(ids)
	r = make([]MessageType, l, l)
	for i, id := range ids {
		r[i] = set.parent.ByMessageTypeId(id)
		if set.iscommands != r[i].IsCommand() {
			return nil
		}
	}
	return
}

type MessageTypeContainerDef struct {
	defs          []*MessageTypeDef
	command_count int
	cache_all     []MessageType
	cache_type    map[reflect.Type]MessageType
	cache_typeid  map[MessageTypeId]MessageType
}

func (set *MessageTypeContainerDef) All() []MessageType {
	return set.cache_all
}

func (set *MessageTypeContainerDef) ByInstance(t interface{}) MessageType {
	return set.cache_type[reflect.TypeOf(t)]
}

func (set *MessageTypeContainerDef) ByMessageTypeId(id MessageTypeId) MessageType {
	return set.cache_typeid[id]
}

func (set *MessageTypeContainerDef) ByMessageTypeIds(ids ...MessageTypeId) []MessageType {
	l := len(ids)
	r := make([]MessageType, l, l)
	for i := 0; i < l; i++ {
		r[i] = set.cache_typeid[ids[i]]
	}
	return r
}

func (set *MessageTypeContainerDef) Commands() MessageTypeSet {
	return &MessageTypeSetDef{
		iscommands: true,
		parent:     set,
	}
}

func (set *MessageTypeContainerDef) Events() MessageTypeSet {
	return &MessageTypeSetDef{
		iscommands: false,
		parent:     set,
	}
}

type MessageTypeDef struct {
	domain        *DomainDef
	iscommand     bool
	messagetypeid MessageTypeId
	displayname   string
	lowername     string
	canonicalname string
	version       uint8
	id            int64
	proto         func() MessageDefiner
}

type ByPurposeAndName []*MessageTypeDef

func (set ByPurposeAndName) Len() int      { return len(set) }
func (set ByPurposeAndName) Swap(i, j int) { set[i], set[j] = set[j], set[i] }
func (set ByPurposeAndName) Less(i, j int) bool {
	return (set[i].IsCommand() && !set[j].IsCommand()) ||
		set[i].CanonicalName() < set[j].CanonicalName()
}

func (d *MessageTypeDef) Domain() Domain {
	return d.domain
}

func (d *MessageTypeDef) MessageTypeId() MessageTypeId {
	return d.messagetypeid
}

func (d *MessageTypeDef) IsCommand() bool {
	return d.iscommand
}

func (d *MessageTypeDef) DisplayName() string {
	return d.displayname
}

func (d *MessageTypeDef) LowerName() string {
	return d.lowername
}

func (d *MessageTypeDef) CanonicalName() string {
	return d.canonicalname
}

func (d *MessageTypeDef) Version() uint8 {
	return d.version
}

func (d *MessageTypeDef) Id() int64 {
	return d.id
}

func (d *MessageTypeDef) NewPayload() MessageDefiner {
	return d.proto()
}

var (
	i_aggregate      = reflect.TypeOf((*Aggregate)(nil)).Elem()
	i_messagedefiner = reflect.TypeOf((*MessageDefiner)(nil)).Elem()
)

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
	m_l := t_commands.NumField() + t_events.NumField()
	types := make([]*MessageTypeDef, m_l, m_l)
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
			def := &MessageTypeDef{
				domain:        d,
				iscommand:     m_id.IsCommand(),
				messagetypeid: m_id,
				displayname:   name,
				lowername:     lower,
				canonicalname: canonical,
				version:       v,
				id:            id,
				proto:         p,
			}
			types[j] = def
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
	d.messagetypecontainer = NewMessageTypeContainer(types...)

	return d, nil
}

func MustCompile(meta interface{}) Domain {
	d, err := Compile(meta)
	if err != nil {
		panic(err)
	}
	return d
}

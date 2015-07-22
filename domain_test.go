package cqrs_test

import (
	"fmt"
	. "github.com/stretchr/testify/assert"
	"github.com/vizidrix/crypto"
	"github.com/xzeus/cqrs"
	"github.com/xzeus/cqrs/mocks/valuechanger/v1"
	"reflect"
	"testing"
)

func Test_Domain_ValidMeta(t *testing.T) {
	d, err := cqrs.Compile(valuechanger.Meta{})
	Nil(t, err)
	NotNil(t, d)
	Equal(t, d.Uri(), "github.com/xzeus/cqrs/mocks/valuechanger/v1")
	Equal(t, d.Id(), int64(5116695313427259383))
	Equal(t, d.Name(), "valuechanger")
	Equal(t, d.Version(), int32(1))
}

var messages = []struct {
	command     bool
	displayname string
	lowername   string
	id          int64
	proto       cqrs.MessageDefiner
}{
	{true, "SetEmpty", "setempty", -372956480745328814, &valuechanger.SetEmpty{}},
	{true, "SetValue", "setvalue", 6184284646522414091, &valuechanger.SetValue{}},
	{false, "ValueChanged", "valuechanged", 6177095821187969785, &valuechanger.ValueChanged{}},
	{false, "ValueChangeFailed", "valuechangefailed", -410083745455383610, &valuechanger.ValueChangeFailed{}},
}

func Test_Messages_parsed_correctly(t *testing.T) {
	m := &valuechanger.Meta{}
	d, err := cqrs.Compile(m)
	Nil(t, err)
	NotPanics(t, func() { cqrs.MustCompile(m) })
	d_str := d.String()
	Equal(t, "valuechanger", d.Name())
	Equal(t, int32(1), d.Version())
	Contains(t, d_str, d.Name())
	Contains(t, d_str, fmt.Sprintf("%d", d.Version()))
	for i, m := range d.MessageTypes() {
		e := messages[i]
		id := crypto.CrcHash64([]byte(m.CanonicalName()))
		if e.command {
			Equal(t, m.MessageTypeId(), cqrs.MakeVersionedCommandType(id, 1))
		} else {
			Equal(t, m.MessageTypeId(), cqrs.MakeVersionedEventType(id, 1))
		}
		Equal(t, d, m.Domain())
		Equal(t, m.DisplayName(), e.displayname)
		Equal(t, m.LowerName(), e.lowername)
		Equal(t, m.Id(), e.id)
		Equal(t, reflect.TypeOf(m.New()), reflect.TypeOf(e.proto))
	}
}

func Test_Override_uri(t *testing.T) {
	m := &struct {
		Aggregate *valuechanger.ValueChanger `uri:"github.com/other/domainname/v1"`
		Commands  struct {
			_ valuechanger.SetValue
		}
		Events struct {
			_ valuechanger.ValueChanged
		}
	}{}
	d, err := cqrs.Compile(m)
	Nil(t, err)
	NotNil(t, d)
	Equal(t, "github.com/other/domainname/v1", d.Uri())
	Equal(t, int64(-3032497499499725252), d.Id())
}

func Test_Override_version_tag(t *testing.T) {
	m := &struct {
		Aggregate *valuechanger.ValueChanger `uri:"github.com/other/domainname/v1"`
		Commands  struct {
			_ valuechanger.SetValue `v:"2"`
		}
		Events struct {
			_ valuechanger.ValueChanged
		}
	}{}
	d, err := cqrs.Compile(m)
	Nil(t, err)
	NotNil(t, d)
	Equal(t, uint8(2), d.MessageTypes()[0].Version())
}

func Test_Override_version_by_name(t *testing.T) {
	type SetValue_v3 struct {
		valuechanger.SetValue
	}
	m := &struct {
		Aggregate *valuechanger.ValueChanger `uri:"github.com/other/domainname/v1"`
		Commands  struct {
			_ SetValue_v3
		}
		Events struct {
			_ valuechanger.ValueChanged
		}
	}{}
	d, err := cqrs.Compile(m)
	Nil(t, err)
	NotNil(t, d)
	Equal(t, uint8(3), d.MessageTypes()[0].Version())
}

func Test_Out_of_order_messages(t *testing.T) {
	type SetValue_v3 struct {
		valuechanger.SetValue
	}
	m := &struct {
		Aggregate valuechanger.ValueChanger `uri:"github.com/other/domainname/v1"`
		Commands  struct {
			_ valuechanger.SetValue
			_ valuechanger.SetEmpty
			_ SetValue_v3
		}
		Events struct {
			_ valuechanger.ValueChanged
		}
	}{}
	d, err := cqrs.Compile(m)
	d.MessageTypes()
	Nil(t, err)
	NotNil(t, d)
}

var invalid_metas = []struct {
	name         string
	expected_err error
	meta         interface{}
}{
	{"missing aggregate", cqrs.ErrAggregateNotProvided, &struct{}{}},
	{"invalid aggregate", cqrs.ErrInvalidAggregate, &struct {
		Aggregate struct{}
	}{}},
	{"missing commands", cqrs.ErrCommandsNotProvided, &struct {
		Aggregate valuechanger.ValueChanger `uri:"a/a/v1"`
	}{}},
	{"invalid commands type", cqrs.ErrCommandsNotProvided, &struct {
		Aggregate valuechanger.ValueChanger `uri:"a/a/v1"`
		Commands  interface{}
	}{}},
	{"empty commands", cqrs.ErrCommandsNotProvided, &struct {
		Aggregate valuechanger.ValueChanger `uri:"a/a/v1"`
		Commands  struct{}
	}{}},
	{"invalid commands", cqrs.ErrInvalidMessage, &struct {
		Aggregate valuechanger.ValueChanger `uri:"a/a/v1"`
		Commands  struct {
			command interface{}
		}
		Events struct {
			_ valuechanger.ValueChanged
		}
	}{}},
	{"invalid command types", cqrs.ErrInvalidMessage, &struct {
		Aggregate valuechanger.ValueChanger `uri:"a/a/v1"`
		Commands  struct {
			command struct{}
		}
		Events struct {
			_ valuechanger.ValueChanged
		}
	}{}},
	{"missing events", cqrs.ErrEventsNotProvided, &struct {
		Aggregate valuechanger.ValueChanger `uri:"a/a/v1"`
		Commands  struct {
			_ valuechanger.SetValue
		}
	}{}},
	{"invalid events type", cqrs.ErrEventsNotProvided, &struct {
		Aggregate valuechanger.ValueChanger `uri:"a/a/v1"`
		Commands  struct {
			_ valuechanger.SetValue
		}
		Events interface{}
	}{}},
	{"empty events", cqrs.ErrEventsNotProvided, &struct {
		Aggregate valuechanger.ValueChanger `uri:"a/a/v1"`
		Commands  struct {
			_ valuechanger.SetValue
		}
		Events struct{}
	}{}},
	{"invalid events", cqrs.ErrInvalidMessage, &struct {
		Aggregate valuechanger.ValueChanger `uri:"a/a/v1"`
		Commands  struct {
			_ valuechanger.SetValue
		}
		Events struct {
			_ interface{}
		}
	}{}},
	{"invalid event types", cqrs.ErrInvalidMessage, &struct {
		Aggregate valuechanger.ValueChanger `uri:"a/a/v1"`
		Commands  struct {
			_ valuechanger.SetValue
		}
		Events struct {
			_ struct{}
		}
	}{}},
	{"invalid version tag", cqrs.ErrInvalidMessageVersion, &struct {
		Aggregate valuechanger.ValueChanger `uri:"a/a/v1"`
		Commands  struct {
			_ valuechanger.SetValue
		}
		Events struct {
			_ valuechanger.ValueChanged `v:"ver"`
		}
	}{}},
	{"invalid version tag", cqrs.ErrInvalidMessageVersion, &struct {
		Aggregate valuechanger.ValueChanger `uri:"a/a/v1"`
		Commands  struct {
			_ valuechanger.SetValue
		}
		Events struct {
			_ ValueChanged_vA
		}
	}{}},
}

type ValueChanged_vA struct {
	valuechanger.ValueChanged
}

func Test_Domain_InvalidMeta(t *testing.T) {
	msg := "invalid meta [ %#v ]"
	for _, invalid_meta := range invalid_metas {
		d, err := cqrs.Compile(invalid_meta.meta)
		Nil(t, d, msg, invalid_meta)
		NotNil(t, err, msg, invalid_meta)
		Panics(t, func() { cqrs.MustCompile(invalid_meta.meta) }, msg, invalid_meta)
	}
}

var invalid_uri_defs = []interface{}{
	&struct {
		Aggregate *valuechanger.ValueChanger `uri:"github.com/other/domainname"`
		Commands  struct{}
		Events    struct{}
	}{},
	&struct {
		Aggregate *valuechanger.ValueChanger `uri:"github.com/other"`
		Commands  struct{}
		Events    struct{}
	}{},
	&struct {
		Aggregate *valuechanger.ValueChanger `uri:"github.com/other/domainname/vA"`
		Commands  struct{}
		Events    struct{}
	}{},
}

func Test_Override_invalid_uri(t *testing.T) {
	msg := "Test index [ %d ]"
	for i, m := range invalid_uri_defs {
		_, err := cqrs.Compile(m)
		NotNil(t, err, msg, i)
		Equal(t, err, cqrs.ErrInvalidDomainUri)
		Panics(t, func() { cqrs.MustCompile(m) }, msg, i)
	}
}

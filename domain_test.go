package cqrs_test

import (
	. "github.com/stretchr/testify/assert"
	"github.com/vizidrix/crypto"
	"github.com/xzeus/cqrs"
	"github.com/xzeus/cqrs/mocks/testdomain/v1"
	"reflect"
	//"log"
	"fmt"
	"testing"
)

func Test_Domain_ValidMeta(t *testing.T) {
	d, err := cqrs.Compile(testdomain.Meta{})

	Nil(t, err)
	NotNil(t, d)

	Equal(t, d.Uri(), "github.com/xzeus/cqrs/mocks/testdomain/v1")
	Equal(t, d.Id(), int64(5116695313427259383))
	Equal(t, d.Name(), "testdomain")
	Equal(t, d.Version(), int32(1))
	Len(t, d.MessageTypes().Commands().All(), 2)
	Len(t, d.MessageTypes().Events().All(), 2)
	Len(t, d.MessageTypes().All(), 4)
}

var messages = []struct {
	command     bool
	displayname string
	lowername   string
	id          int64
	proto       cqrs.MessageDefiner
}{
	{true, "SetEmpty", "setempty", -372956480745328814, &testdomain.SetEmpty{}},
	{true, "SetValue", "setvalue", 6184284646522414091, &testdomain.SetValue{}},
	{false, "ValueChanged", "valuechanged", 6177095821187969785, &testdomain.ValueChanged{}},
	{false, "ValueChangeFailed", "valuechangefailed", -410083745455383610, &testdomain.ValueChangeFailed{}},
}

func Test_Messages_parsed_correctly(t *testing.T) {
	m := &testdomain.Meta{}
	d, err := cqrs.Compile(m)
	Nil(t, err)
	NotPanics(t, func() { cqrs.MustCompile(m) })
	d_str := d.String()
	Equal(t, "testdomain", d.Name())
	Equal(t, int32(1), d.Version())

	Contains(t, d_str, d.Name())
	Contains(t, d_str, fmt.Sprintf("%d", d.Version()))
	for i, m := range d.MessageTypes().All() {
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
		m1 := d.MessageTypes().ByInstance(e.proto)
		msg := "Message by instance [ %s ]"
		if NotNil(t, m1, msg, m.CanonicalName()) {
			Equal(t, m1.CanonicalName(), m.CanonicalName(), msg, m.CanonicalName())
			Equal(t, m1.LowerName(), e.lowername, msg, m.CanonicalName())
		}
		m2 := d.MessageTypes().ByMessageTypeId(m.MessageTypeId())
		msg = "Message by typeid [ %s ]"
		if NotNil(t, m2, msg, m.CanonicalName()) {
			Equal(t, m2.CanonicalName(), m.CanonicalName(), msg, m.CanonicalName())
			Equal(t, m2.LowerName(), e.lowername, msg, m.CanonicalName())
		}
	}
}

func Test_Messages_return_by_lookup(t *testing.T) {
	d, err := cqrs.Compile(&testdomain.Meta{})
	Nil(t, err)
	NotNil(t, d)
	all := d.MessageTypes().All()
	mult := d.MessageTypes().ByMessageTypeIds(all[1].MessageTypeId(), all[3].MessageTypeId())
	Equal(t, mult[0].CanonicalName(), all[1].CanonicalName())
	Equal(t, mult[1].CanonicalName(), all[3].CanonicalName())
	a := d.NewAggregate()
	Implements(t, (*cqrs.Aggregate)(nil), a)
	IsType(t, &testdomain.TestDomain{}, a)
}

func Test_Message_projections_return_by_lookup(t *testing.T) {
	d, err := cqrs.Compile(&testdomain.Meta{})
	Nil(t, err)
	for i, m := range d.MessageTypes().All() {
		e := messages[i]
		if e.command {
			c_set := d.MessageTypes().Commands()
			Equal(t, m.CanonicalName(), c_set.ByInstance(m.New()).CanonicalName())
			Equal(t, m.CanonicalName(), c_set.ByMessageTypeId(m.MessageTypeId()).CanonicalName())
			cs := c_set.All()
			mult := c_set.ByMessageTypeIds(cs[0].MessageTypeId(), cs[1].MessageTypeId())
			Equal(t, mult[0].CanonicalName(), cs[0].CanonicalName())
			Equal(t, mult[1].CanonicalName(), cs[1].CanonicalName())
		} else {
			e_set := d.MessageTypes().Events()
			Equal(t, m.CanonicalName(), e_set.ByInstance(m.New()).CanonicalName())
			Equal(t, m.CanonicalName(), e_set.ByMessageTypeId(m.MessageTypeId()).CanonicalName())
			es := e_set.All()
			mult := e_set.ByMessageTypeIds(es[0].MessageTypeId(), es[1].MessageTypeId())
			Equal(t, mult[0].CanonicalName(), es[0].CanonicalName())
			Equal(t, mult[1].CanonicalName(), es[1].CanonicalName())
		}
	}
}

func Test_Message_projections_invalid_lookup(t *testing.T) {
	d, err := cqrs.Compile(&testdomain.Meta{})
	Nil(t, err)
	all := d.MessageTypes().All()
	for i, _ := range all {
		e := messages[i]
		if e.command {
			inv := all[2:]
			c_set := d.MessageTypes().Commands()
			Nil(t, c_set.ByInstance(inv[0].New()))
			Nil(t, c_set.ByMessageTypeId(inv[0].MessageTypeId()))
			Nil(t, c_set.ByMessageTypeIds(inv[0].MessageTypeId(), inv[1].MessageTypeId()))
		} else {
			inv := all[:2]
			e_set := d.MessageTypes().Events()
			Nil(t, e_set.ByInstance(inv[0].New()))
			Nil(t, e_set.ByMessageTypeId(inv[0].MessageTypeId()))
			Nil(t, e_set.ByMessageTypeIds(inv[0].MessageTypeId(), inv[1].MessageTypeId()))
		}
	}
}

func Test_Override_uri(t *testing.T) {
	m := &struct {
		Aggregate *testdomain.TestDomain `uri:"github.com/other/domainname/v1"`
		Commands  struct {
			_ *testdomain.SetValue
		}
		Events struct {
			_ *testdomain.ValueChanged
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
		Aggregate *testdomain.TestDomain `uri:"github.com/other/domainname/v1"`
		Commands  struct {
			_ *testdomain.SetValue `v:"2"`
		}
		Events struct {
			_ *testdomain.ValueChanged
		}
	}{}
	d, err := cqrs.Compile(m)
	Nil(t, err)
	NotNil(t, d)
	Equal(t, uint8(2), d.MessageTypes().Commands().All()[0].Version())
}

func Test_Override_version_by_name(t *testing.T) {
	type SetValue_v3 struct {
		testdomain.SetValue
	}
	m := &struct {
		Aggregate *testdomain.TestDomain `uri:"github.com/other/domainname/v1"`
		Commands  struct {
			_ *SetValue_v3
		}
		Events struct {
			_ *testdomain.ValueChanged
		}
	}{}
	d, err := cqrs.Compile(m)
	Nil(t, err)
	NotNil(t, d)
	Equal(t, uint8(3), d.MessageTypes().Commands().All()[0].Version())
}

func Test_Out_of_order_messages(t *testing.T) {
	type SetValue_v3 struct {
		testdomain.SetValue
	}
	m := &struct {
		Aggregate testdomain.TestDomain `uri:"github.com/other/domainname/v1"`
		Commands  struct {
			_ testdomain.SetValue
			_ testdomain.SetEmpty
			_ SetValue_v3
		}
		Events struct {
			_ testdomain.ValueChanged
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
		Aggregate testdomain.TestDomain `uri:"a/a/v1"`
	}{}},
	{"invalid commands type", cqrs.ErrCommandsNotProvided, &struct {
		Aggregate testdomain.TestDomain `uri:"a/a/v1"`
		Commands  interface{}
	}{}},
	{"empty commands", cqrs.ErrCommandsNotProvided, &struct {
		Aggregate testdomain.TestDomain `uri:"a/a/v1"`
		Commands  struct{}
	}{}},
	{"invalid commands", cqrs.ErrInvalidMessage, &struct {
		Aggregate testdomain.TestDomain `uri:"a/a/v1"`
		Commands  struct {
			command interface{}
		}
		Events struct {
			_ testdomain.ValueChanged
		}
	}{}},
	{"invalid command types", cqrs.ErrInvalidMessage, &struct {
		Aggregate testdomain.TestDomain `uri:"a/a/v1"`
		Commands  struct {
			command struct{}
		}
		Events struct {
			_ testdomain.ValueChanged
		}
	}{}},
	{"missing events", cqrs.ErrEventsNotProvided, &struct {
		Aggregate testdomain.TestDomain `uri:"a/a/v1"`
		Commands  struct {
			_ testdomain.SetValue
		}
	}{}},
	{"invalid events type", cqrs.ErrEventsNotProvided, &struct {
		Aggregate testdomain.TestDomain `uri:"a/a/v1"`
		Commands  struct {
			_ testdomain.SetValue
		}
		Events interface{}
	}{}},
	{"empty events", cqrs.ErrEventsNotProvided, &struct {
		Aggregate testdomain.TestDomain `uri:"a/a/v1"`
		Commands  struct {
			_ testdomain.SetValue
		}
		Events struct{}
	}{}},
	{"invalid events", cqrs.ErrInvalidMessage, &struct {
		Aggregate testdomain.TestDomain `uri:"a/a/v1"`
		Commands  struct {
			_ testdomain.SetValue
		}
		Events struct {
			_ interface{}
		}
	}{}},
	{"invalid event types", cqrs.ErrInvalidMessage, &struct {
		Aggregate testdomain.TestDomain `uri:"a/a/v1"`
		Commands  struct {
			_ testdomain.SetValue
		}
		Events struct {
			_ struct{}
		}
	}{}},
	{"invalid version tag", cqrs.ErrInvalidMessageVersion, &struct {
		Aggregate testdomain.TestDomain `uri:"a/a/v1"`
		Commands  struct {
			_ testdomain.SetValue
		}
		Events struct {
			_ testdomain.ValueChanged `v:"ver"`
		}
	}{}},
	{"invalid version tag", cqrs.ErrInvalidMessageVersion, &struct {
		Aggregate testdomain.TestDomain `uri:"a/a/v1"`
		Commands  struct {
			_ testdomain.SetValue
		}
		Events struct {
			_ ValueChanged_vA
		}
	}{}},
}

type ValueChanged_vA struct {
	testdomain.ValueChanged
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
		Aggregate *testdomain.TestDomain `uri:"github.com/other/domainname"`
		Commands  struct{}
		Events    struct{}
	}{},
	&struct {
		Aggregate *testdomain.TestDomain `uri:"github.com/other"`
		Commands  struct{}
		Events    struct{}
	}{},
	&struct {
		Aggregate *testdomain.TestDomain `uri:"github.com/other/domainname/vA"`
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

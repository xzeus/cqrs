//go:generate ffjson $GOFILE
package testdomain

import (
	"errors"
	. "github.com/xzeus/cqrs"
)

var (
	ErrValueEmpty = errors.New("cannot set empty value")
)

var DomainRef = DomainFromMeta(Meta{})

type __ struct{}

func (_ __) Domain() Domain       { return DomainRef }
func (__ __) New() MessageDefiner { return DomainRef.MessageTypes().ByInstance(__) }

type Meta struct {
	Aggregate TestDomain
	Commands  struct {
		_ SetEmpty
		_ SetValue
	}
	Events struct {
		_ ValueChanged
		_ ValueChangeFailed
	}
}

type TestDomain struct {
	Value string
}

func (s TestDomain) Init() Aggregate {
	return TestDomain{
		Value: "",
	}
}

func (s TestDomain) Apply(e Event) Aggregate {
	switch m := e.Message().(type) {
	case ValueChanged:
		s.Value = m.NewValue
	}
	return s
}

func (s TestDomain) Handle(c Command) EventMessage {
	switch m := c.Message().(type) {
	case SetEmpty:
		if s.Value == "" {
			return ValueChangeFailed{
				Message: "Value was already empty",
			}
		}
		return ValueChanged{
			PreviousValue: s.Value,
			NewValue:      "",
		}
	case SetValue:
		return ValueChanged{
			PreviousValue: s.Value,
			NewValue:      m.Value,
		}
	default:
		panic("invalid command")
	}

}

type SetEmpty struct {
	__
}

func (t SetEmpty) Validate(c Command) error {
	return nil // TODO: Check header for 'clear value' claim?
}

type SetValue struct {
	__
	Value string
}

func (t SetValue) Validate(c Command) error {
	if t.Value == "" {
		return ErrValueEmpty
	}
	return nil
}

type ValueChanged struct {
	__
	PreviousValue string
	NewValue      string
}

type ValueChangeFailed struct {
	__
	Message string
}

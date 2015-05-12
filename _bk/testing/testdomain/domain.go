package testdomain

import (
	"github.com/xzeus/cqrs"
	"github.com/xzeus/cqrs/domains"
)

const (
	Uri = "github.com/vizidrix/zeus/providers/mockprovider/testdomain"
)

type __ struct{}

func (_ __) Domain() cqrs.Domain { return Domain }

var (
	Domain = domains.NewDomain(&__{}, Uri, &TestAggregate{})

	Handler = Domain.DefCommandHandler(func(h cqrs.CommandHandlerDef) cqrs.CommandHandlerFunc {
		return (&CommandHandler{h}).Handle
	})

	C_TestCommand      = Domain.DefCommand(1, 1, &TestCommand{})
	C_AltTestCommand   = Domain.DefCommand(1, 2, &AltTestCommand{})
	C_TestKeyedCommand = Domain.DefCommand(1, 3, &TestKeyedCommand{})

	E_TestEvent      = Domain.DefEvent(1, 1, &TestEvent{})
	E_AltTestEvent   = Domain.DefEvent(1, 2, &AltTestEvent{})
	E_TestKeyedEvent = Domain.DefEvent(1, 3, &TestKeyedEvent{})
)

var Mock_Handle = func(cqrs.AggregateHeader, cqrs.AggregateState, cqrs.Message, cqrs.MessageDefiner) {
	panic("mockprovider.testdomain: Mock handle not implemented")
}

type CommandHandler struct {
	cqrs.CommandHandlerDef
}

func (c *CommandHandler) Handle(
	header cqrs.AggregateHeader,
	aggregate cqrs.AggregateState,
	command cqrs.Message,
	payload cqrs.MessageDefiner) {
	Mock_Handle(header, aggregate, command, payload)
}

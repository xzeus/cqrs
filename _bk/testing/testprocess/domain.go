package testprocess

import (
	"github.com/xzeus/cqrs"
	"github.com/xzeus/cqrs/domains"
)

const (
	Name    = "TestDomain2"
	Version = "v1_0"
	Url     = "github.com/vizidrix/zeus/providers/mockdependencies/testdomain2"
)

type __ struct{}

func (_ __) Domain() cqrs.Domain { return Domain }

var (
	Domain = domains.NewDomain(&__{}, Name, Version, Url, CommandHandlerFactory, &TestAggregate{})

	Handler = Domain.DefHandler(func(h cqrs.CommandHandlerDef) cqrs.CommandHandlerFunc {
		return (&CommandHandler{h}).Handle
	})

	C_TestCommand = Domain.DefCommand(1, 1, &TestCommand{})

	E_TestEvent = Domain.DefEvent(1, 1, &TestEvent{})
)

var Mock_Handle = func(cqrs.AggregateHeader, cqrs.AggregateState, cqrs.Message, cqrs.MessageDefiner) {
	panic("testdependencies.testdomain: Mock handle not implementd")
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

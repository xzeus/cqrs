package ioc

import (
	"github.com/xzeus/cqrs"
)

type Exception interface {
	Error(message string, args ...interface{}) (cqrs.MessageDefiner, *cqrs.MessageOptionsDef)
	Panic(message string, args ...interface{}) (cqrs.MessageDefiner, *cqrs.MessageOptionsDef)
}

package testdomain

import "github.com/xzeus/cqrs"

type TestCommand struct {
	cqrs.JsonSerialized
	__
	Value string `json:"value"`
}

type AltTestCommand struct {
	cqrs.JsonSerialized
	__
	Value string `json:"value"`
}

type TestKeyedCommand struct {
	cqrs.JsonSerialized
	cqrs.Keyed
	__
	Value string `json:"value"`
}

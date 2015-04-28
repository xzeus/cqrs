package testdomain

import "github.com/xzeus/cqrs"

type TestEvent struct {
	cqrs.JsonSerialized
	__
	Value string `json:"value"`
}

type AltTestEvent struct {
	cqrs.JsonSerialized
	__
	Value string `json:"value"`
}

type TestKeyedEvent struct {
	cqrs.JsonSerialized
	cqrs.Keyed
	__
	Value string `json:"value"`
}

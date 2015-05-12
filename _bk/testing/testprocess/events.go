package testprocess

import "github.com/xzeus/cqrs"

type TestEvent struct {
	cqrs.JsonSerialized
	__
	Value string `json:"value"`
}

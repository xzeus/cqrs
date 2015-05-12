package testprocess

import "github.com/xzeus/cqrs"

type TestCommand struct {
	cqrs.JsonSerialized
	__
	Value string `json:"value"`
}

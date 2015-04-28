package testdomain

import "github.com/xzeus/cqrs"

type TestAggregate struct {
	cqrs.AggregateState
	cqrs.JsonSerialized
	__
	Value string `json:"value"`
}

func (a *TestAggregate) Init() cqrs.AggregateState {
	return &TestAggregate{
		Value: "",
	}
}

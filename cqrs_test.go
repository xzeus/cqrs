package cqrs_test

import (
	. "github.com/stretchr/testify/assert"
	"github.com/xzeus/cqrs"
	"github.com/xzeus/cqrs/mocks/testdomain/v1"
	"testing"
)

func Test_MessageType_GenerateValidIds(t *testing.T) {
	ms := []struct {
		version uint8
		type_id int64
	}{
		{version: 1, type_id: 1},
		{version: 100, type_id: 1},
		{version: 1, type_id: 10000},
	}
	for i, m := range ms {
		msg := "Typedef index [ %d ]"
		mt_command := cqrs.MakeVersionedCommandType(m.type_id, m.version)
		mt_event := cqrs.MakeVersionedEventType(m.type_id, m.version)
		True(t, mt_command.IsCommand(), msg, i)
		False(t, mt_event.IsCommand(), msg, i)
		NotEqual(t, mt_event, mt_command, msg, i)
	}
}

func Test_MessageInstance_New(t *testing.T) {
	m := &testdomain.SetEmpty{}
	c := m.Domain().MessageTypes().ByInstance(m).New()
	NotNil(t, c)
	IsType(t, m, c)
	m2 := testdomain.SetEmpty{}
	c = m2.Domain().MessageTypes().ByInstance(m2).New()
	NotNil(t, c)
	IsType(t, &m2, c)
}

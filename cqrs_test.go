package cqrs_test

import (
	"fmt"
	. "github.com/smartystreets/goconvey/convey"
	"github.com/xzeus/cqrs"
	"testing"
)

func TestCqrs(t *testing.T) {
	Convey("Given message type defs", t, func() {
		ms := []struct {
			version uint8
			type_id int64
		}{
			{version: 1, type_id: 1},
		}
		for i, m := range ms {
			Convey(fmt.Sprintf("[ %d ] Should make and verify message types", i), func() {
				mt_command := cqrs.MakeVersionedCommandType(m.type_id, m.version)
				So(mt_command.IsCommand(), ShouldBeTrue)
				mt_event := cqrs.MakeVersionedEventType(m.type_id, m.version)
				So(mt_event.IsCommand(), ShouldBeFalse)
				So(mt_command, ShouldNotEqual, mt_event)
			})
		}
	})

	//Convey("Given a valid command context", t, func() {

	//})
}

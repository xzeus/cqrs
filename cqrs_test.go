package cqrs_test

import (
	"github.com/vizidrix/zeus/cqrs"
	. "github.com/vizidrix/zeus/testing"
	. "github.com/vizidrix/zeus/testing/testdomain"
	"testing"
)

var int64_to_hex_values = []struct {
	value    uint64
	expected string
}{
	{0, "0"},
	{10974667226560855258, "984DD9B23F47FCDA"},
}

func Test_Should_Hex(t *testing.T) {
	for _, e := range int64_to_hex_values {
		hex := cqrs.Hex(int64(e.value))
		Equals(t, e.expected, hex, "Expected same value", hex)
	}
}

var hex_to_int64_values = []struct {
	value    string
	expected uint64
}{
	{"984DD9B23F47FCDA", 10974667226560855258},
}

func Test_Should_UnHex(t *testing.T) {
	for _, e := range hex_to_int64_values {
		unhex, err := cqrs.UnHex64(e.value)
		Ok(t, err)
		Equals(t, e.expected, uint64(unhex), "Expected same value", unhex)
	}
}

func Test_Should_define_valid_domain(t *testing.T) {
	Equals(t, Url, Domain.Url(), "")
	Equals(t, cqrs.Hash(Url), Domain.Id(), "")
}

func Test_Should_return_valid_aggregate_fields(t *testing.T) {
	var exp_domain int32 = 1
	var exp_id int64 = 1
	var exp_version int32 = 1
	test_aggregate := cqrs.NewAggregate(exp_domain, exp_id, exp_version, cqrs.JsonSerialized{})
	Assert(t, test_aggregate.GetId() == exp_id, "Expected id [ %d ] but received [ %d ]", exp_id, test_aggregate.GetId())
	Assert(t, test_aggregate.GetVersion() == exp_version, "Expected version [ %d ] but received [ %d ]", exp_version, test_aggregate.GetVersion())
	test_aggregate.GetUUID()
	test_aggregate.GetData()
	test_aggregate.GetBytes()
	test_aggregate.String()
}

func Test_Should_return_valid_command(t *testing.T) {
	id := int64(1)
	version := int32(2)
	time := int64(3)
	origin := cqrs.NoOrigin
	payload := &TestCommand{}

	e := cqrs.NewMessage(id, version, time, origin, payload)

	c_id := uint32(0x81000001)

	Equals(t, id, e.GetId(), "")
	Equals(t, version, e.GetVersion(), "")
	Equals(t, time, e.GetTimestamp(), "")
	Equals(t, cqrs.NoOrigin, e.GetOrigin(), "")
	Equals(t, int32(c_id), int32(e.GetMessageType()), "")
}

func Test_Should_return_valid_event(t *testing.T) {
	id := int64(1)
	version := int32(2)
	time := int64(3)
	origin := cqrs.NoOrigin
	payload := &TestEvent{}

	e := cqrs.NewMessage(id, version, time, origin, payload)

	Equals(t, id, e.GetId(), "")
	Equals(t, version, e.GetVersion(), "")
	Equals(t, time, e.GetTimestamp(), "")
	Equals(t, cqrs.NoOrigin, e.GetOrigin(), "")
	Equals(t, int32(0x1000001), int32(e.GetMessageType()), "")
}

func Test_Should_deserialize_correctly(t *testing.T) {
	test_val := "test"
	expected_struct := struct {
		cqrs.JsonSerialized
		Test string `json:"test"`
	}{Test: test_val}
	data, err := expected_struct.Serialize(expected_struct)
	Ok(t, err)
	test_struct := struct {
		cqrs.JsonSerialized
		Test string `json:"test"`
		test string
	}{}
	err = test_struct.Deserialize(data, &test_struct)
	Ok(t, err)
	Assert(t, test_struct.Test == test_val, "Expected Test value [ %s ] but received [ %s ]", test_val, test_struct.Test)
	Assert(t, test_struct.test == "", "Expected empty test value but received [ %s ]", test_struct.test)
}

func Test_Should_extract_keyed_payload_correctly(t *testing.T) {
	expected_key := "test"
	test_payload := &TestKeyedEvent{
		Keyed: cqrs.Keyed{LookupKey: []byte(expected_key)},
		Value: expected_key,
	}
	result := cqrs.ExtractKey(test_payload)
	Assert(t, expected_key == result, "Expected Keyed look up key [ %s ] but received [ %s ]", expected_key, result)
}

func Test_Should_return_all_messages(t *testing.T) {
	event_found := false
	command_found := false
	for message_type, _ := range Domain.Messages() {
		if message_type == E_TestEvent {
			event_found = true
		}
		if message_type == C_TestCommand {
			command_found = true
		}
	}
	Assert(t, event_found, "Expected event to appear")
	Assert(t, command_found, "Expected command to appear")
}

func Test_Should_return_filtered_messages(t *testing.T) {
	event_found := false
	command_found := false
	for message_type, _ := range Domain.Messages(
		E_TestEvent) {
		if message_type == E_TestEvent {
			event_found = true
		}
		if message_type == C_TestCommand {
			command_found = true
		}
	}
	Assert(t, event_found, "Expected event to appear")
	Assert(t, !command_found, "Expected command to appear")
}

/*
func Test_Should_include_command_by_default(t *testing.T) {
	event_found := false
	command_found := false

	service := Domain.DefService(func(deps interface{}, message cqrs.Message) cqrs.Service {
		return &service{handlers.NewCommandHandler(deps.(handlers.Dependencies), message)}
	})
	for message_type, _ := range service.Subscriptions() {
		if message_type == E_TestEvent {
			event_found = true
		}
		if message_type == C_TestCommand {
			command_found = true
		}
	}
	Assert(t, !event_found, "Expected event to appear")
	Assert(t, command_found, "Expected command to appear")
}

func Test_Should_include_provided_events_also(t *testing.T) {
	event_found := false
	command_found := false
	service := Domain.DefService(func(deps interface{}, message cqrs.Message) cqrs.Service {
		return &service{handlers.NewCommandHandler(deps.(handlers.Dependencies), message)}
	}, Domain.Messages(E_TestEvent))
	for message_type, _ := range service.Subscriptions() {
		if message_type == E_TestEvent {
			event_found = true
		}
		if message_type == C_TestCommand {
			command_found = true
		}
	}
	Assert(t, event_found, "Expected event to appear")
	Assert(t, command_found, "Expected command to appear")
}
*/

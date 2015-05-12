package apiserver

import (
	"testing"
)

func Test_Should_parse_valid_int64(t *testing.T) {
	var value = "10"
	var expected int64 = 16

	actual, err := Int64(value)

	if err != nil {
		t.Errorf("Unable to parse value [ %s ] because [ %s ]", value, err)
		return
	}
	if actual != expected {
		t.Errorf("Expected [ %d ] but was [ %d ]", expected, actual)
	}
}

func Test_Should_parse_valid_int32(t *testing.T) {
	var value = "10"
	var expected int32 = 16

	actual, err := Int32(value)

	if err != nil {
		t.Errorf("Unable to parse value [ %s ] because [ %s ]", value, err)
		return
	}
	if actual != expected {
		t.Errorf("Expected [ %d ] but was [ %d ]", expected, actual)
	}
}

func Test_Should_parse_positive_int64_from_hex(t *testing.T) {
	value := "77DEF6A07BDAAFE"
	var expected int64 = 539850769029769982

	actual, err := Int64(value)

	if err != nil {
		t.Errorf("Unable to parse value [ %s ] because [ %s ]", value, err)
		return
	}
	if actual != expected {
		t.Errorf("Expected [ %d ] but was [ %d ]", expected, actual)
	}
}

func Test_Should_parse_negative_int64_from_hex(t *testing.T) {
	value := "9C125700644973E4"
	var expected int64 = -7200597195017849884

	actual, err := Int64(value)

	if err != nil {
		t.Errorf("Unable to parse value [ %s ] because [ %s ]", value, err)
		return
	}
	if actual != expected {
		t.Errorf("Expected [ %d ] but was [ %d ]", expected, actual)
	}
}

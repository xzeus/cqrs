package cqrs_test

import (
	"github.com/vizidrix/zeus/cqrs"
	"testing"
)

var EqualSets = []struct {
	Set      []int64
	Expected []int64
}{
	{[]int64{}, []int64{}},
	{[]int64{10, 20}, []int64{10, 20}},
	{[]int64{10, 20}, []int64{10, 20}},
}

func Test_Should_match_equal_sets(t *testing.T) {
	for i, tt := range EqualSets {
		s := cqrs.NewInt64Set(tt.Set...)
		if !s.Equals(tt.Expected...) {
			t.Errorf("[ %d ] Expected [ %#v ] but was [ %#v ]", i, tt.Expected, s.ToSlice())
		}
	}
}

var AppendTests = []struct {
	Set            []int64
	Value          int64
	ExpectedValues []int64
	ExpectedOut    bool
}{
	{[]int64{}, 10, []int64{10}, true},
	{[]int64{}, 20, []int64{20}, true},
	{[]int64{20, 30}, 40, []int64{20, 30, 40}, true},
	{[]int64{20, 30}, 30, []int64{20, 30}, false},
}

func Test_Should_add_with_proper_result(t *testing.T) {
	for i, tt := range AppendTests {
		s := cqrs.NewInt64Set(tt.Set...)
		if r := s.Add(tt.Value); r != tt.ExpectedOut {
			t.Errorf("[ %d ] Expected out [ %#v ] but was [ %#v ]", i, tt.ExpectedOut, r)
		}
		if !s.Equals(tt.ExpectedValues...) {
			t.Errorf("[ %d ] Expected [ %#v ] but was [ %#v ]", i, tt.ExpectedValues, s.ToSlice())
		}
	}
}

var RemoveTests = []struct {
	Set            []int64
	Value          int64
	ExpectedValues []int64
	ExpectedOut    bool
}{
	{[]int64{10}, 10, []int64{}, true},
	{[]int64{10}, 20, []int64{10}, false},
	{[]int64{10}, 20, []int64{10}, false},
	{[]int64{10, 20}, 20, []int64{10}, true},
	{[]int64{10, 20, 30}, 20, []int64{10}, true},
}

func Test_Should_remove_with_proper_result(t *testing.T) {
	for i, tt := range RemoveTests {
		s := cqrs.NewInt64Set(tt.Set...)
		if r := s.Remove(tt.Value); r != tt.ExpectedOut {
			t.Errorf("[ %d ] Expected out [ %#v ] but was [ %#v ]", i, tt.ExpectedOut, r)
		}
		if !s.Equals(tt.ExpectedValues...) {
			t.Errorf("[ %d ] Expected [ %#v ] but was [ %#v ]", i, tt.ExpectedValues, s.ToSlice())
		}
	}
}

var DiffTests = []struct {
	Set               []int64
	DiffSet           []int64
	ExpectedLeftSet   []int64
	ExpectedCommonSet []int64
	ExpectedRightSet  []int64
}{
	{[]int64{}, []int64{}, []int64{}, []int64{}, []int64{}},
	{[]int64{10}, []int64{10}, []int64{}, []int64{10}, []int64{}},
	{[]int64{10, 20}, []int64{10, 30}, []int64{20}, []int64{10}, []int64{30}},
	{[]int64{10, 20, 30}, []int64{}, []int64{10, 20, 30}, []int64{}, []int64{}},
}

func Test_Should_(t *testing.T) {
	for i, tt := range DiffTests {
		s := cqrs.NewInt64Set(tt.Set...)
		l, c, r := s.DiffSet(tt.DiffSet...)
		if !l.Equals(tt.ExpectedLeftSet...) {
			t.Errorf("[ %d ] Expected [ %#v ] but was [ %#v ]", i, tt.ExpectedLeftSet, l.ToSlice())
		}
		if !c.Equals(tt.ExpectedCommonSet...) {
			t.Errorf("[ %d ] Expected [ %#v ] but was [ %#v ]", i, tt.ExpectedCommonSet, c.ToSlice())
		}
		if !r.Equals(tt.ExpectedRightSet...) {
			t.Errorf("[ %d ] Expected [ %#v ] but was [ %#v ]", i, tt.ExpectedRightSet, r.ToSlice())
		}
	}
}

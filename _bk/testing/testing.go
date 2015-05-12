package testing

import (
	"fmt"
	"path/filepath"
	"reflect"
	"runtime"
	"testing"
)

// fails the test if the condition is false.
func Assert(tb testing.TB, condition bool, msg string, v ...interface{}) {
	if !condition {
		_, file, line, _ := runtime.Caller(1)
		fmt.Printf("\033[31m%s:%d: "+msg+"\033[39m\n\n", append([]interface{}{filepath.Base(file), line}, v...)...)
		//tb.FailNow()
		tb.Fail()
	}
}

// fails the test if an err is nil.
func Ok(tb testing.TB, err error) {
	if err != nil {
		_, file, line, _ := runtime.Caller(1)
		fmt.Printf("\033[31m%s:%d: unexpected error: %s\033[39m\n\n", filepath.Base(file), line, err.Error())
		//tb.FailNow()
		tb.Fail()
	}
}

// fails the test if an err is not nil.
func NotOk(tb testing.TB, err error) {
	if err == nil {
		_, file, line, _ := runtime.Caller(1)
		fmt.Printf("\033[31m%s:%d: unexpected error: %s\033[39m\n\n", filepath.Base(file), line, err.Error())
		//tb.FailNow()
		tb.Fail()
	}
}

// fails the test if exp is not equal to act.
func Equals(tb testing.TB, exp, act interface{}, msg string, v ...interface{}) {
	if !reflect.DeepEqual(exp, act) {
		_, file, line, _ := runtime.Caller(1)
		fmt.Printf("\033[31m%s:%d:\n[\n%s\n]\n\n\n\texp: %#v\n\n\tgot: %#v\033[39m\n\n", filepath.Base(file), line, fmt.Sprintf(msg, v...), exp, act)
		//tb.FailNow()
		tb.Fail()
	}
}

// fails the test if exp is equal to act.
func NotEquals(tb testing.TB, exp, act interface{}, msg string, v ...interface{}) {
	if reflect.DeepEqual(exp, act) {
		_, file, line, _ := runtime.Caller(1)
		fmt.Printf("\033[31m%s:%d:\n[\n%s\n]\n\n\n\texp: %#v\n\n\tgot: %#v\033[39m\n\n", filepath.Base(file), line, fmt.Sprintf(msg, v...), exp, act)
		//tb.FailNow()
		tb.Fail()
	}
}

func MultiOk(tb testing.TB, merr []error) {
	for _, err := range merr {
		_, file, line, _ := runtime.Caller(1)
		fmt.Printf("\033[31m%s:%d: unexpected error: [ %s ]\033[39m\n\n", filepath.Base(file), line, err.Error())
		tb.Fail()
	}
}

func MultiOk_err(tb testing.TB, expected_err, actual_err []error) {
	success := true
	defer func() {
		if success {
			return
		}
		for i, err := range expected_err {
			_, file, line, _ := runtime.Caller(1)
			fmt.Printf("\033[31m%s:%d: expected error: [ %s ] but was [ %s ]\033[39m\n\n", filepath.Base(file), line, err, actual_err[i])
			tb.Fail()
		}
	}()
	for i, err := range expected_err {
		if actual_err[i] != err {
			success = false
			return
		}
	}
}

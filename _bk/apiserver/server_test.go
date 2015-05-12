package apiserver_test

import (
	j "github.com/vizidrix/jose"
	. "github.com/xzeus/cqrs/apiserver"
	"github.com/xzeus/cqrs/ioc"
	. "github.com/xzeus/cqrs/testing"
	mock "github.com/xzeus/cqrs/testing/mockprovider"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func na() {
	log.Printf("", ioutil.WriteFile, http.Redirect, httptest.NewRecorder)
}

func MockProvider(p ioc.Dependencies) ProviderFunc {
	if p == nil {
		p = mock.NewDependencies()
	}
	return func(r *http.Request) func() ioc.Dependencies {
		return func() ioc.Dependencies {
			return p
		}
	}
}

func Test_Should_not_allow_nil_provider(t *testing.T) {
	_, err := NewServer(nil)
	NotOk(t, err)
}

func Test_Should_compose_simple_endpoint_structure(t *testing.T) {
	apiFunc := func(Request, Response) {}
	s, err := NewServer(MockProvider(nil))
	Ok(t, err)

	s.Define(
		View("foo", apiFunc),
	)

	handlers := s.Compose()

	Assert(t, len(handlers) == 1, "unexpected endpoint map size")

	ep_string := "/foo"
	_, exists := handlers[ep_string]
	Assert(t, exists, ep_string+" has no entry in endpoints map")
}

func Test_Should_compose_complex_endpoint_structure(t *testing.T) {
	apiFunc := func(Request, Response) {}
	s, err := NewServer(MockProvider(nil))
	Ok(t, err)

	s.Define(
		Sub("v1",
			Sub("foo",
				View("glue", apiFunc),
				Sub("chew",
					View("stew", apiFunc),
					Sub("two",
						View("blue", apiFunc),
					),
				),
			),
			Sub("bar",
				View("star", apiFunc),
			),
		),
	)
	handlers := s.Compose()
	Assert(t, len(handlers) == 4, "unexpected endpoint map size")

	ep_string := "/v1/foo/glue"
	_, exists := handlers[ep_string]
	Assert(t, exists, ep_string+" has no entry in endpoints map")

	ep_string = "/v1/foo/chew/stew"
	_, exists = handlers[ep_string]
	Assert(t, exists, ep_string+" has no entry in endpoints map")

	ep_string = "/v1/foo/chew/two/blue"
	_, exists = handlers[ep_string]
	Assert(t, exists, ep_string+" has no entry in endpoints map")

	ep_string = "/v1/bar/star"
	_, exists = handlers[ep_string]
	Assert(t, exists, ep_string+" has no entry in endpoints map")
}

func Test_Should_deny_request_without_session(t *testing.T) {
	spy_called := false
	apiFuncSpy := func(req Request, resp Response) {
		spy_called = true
	}
	s, err := NewServer(MockProvider(nil))
	Ok(t, err)
	s.Define(
		View("test2", apiFuncSpy),
	)
	r := s.BuildRouter()
	test_server := httptest.NewServer(r)
	res, err := http.Get(test_server.URL + "/test2")
	Ok(t, err)
	body, err := ioutil.ReadAll(res.Body)
	Ok(t, err)
	sbody := string(body)
	NotEquals(t, "", sbody, "spy should return invalid session")
	Assert(t, strings.Contains(sbody, "invalid session"), "message should indicate invalid session")
	res.Body.Close()
	defer test_server.Close()
	Assert(t, !spy_called, "apiFuncSpy was called after Compose and Bind but should not have been")
}

func Test_Should_pick_up_session_from_state_param(t *testing.T) {
	spy_called := false
	apiFuncSpy := func(req Request, resp Response) {
		spy_called = true
	}
	md := mock.NewDependencies()
	md.Mock_Crypto.Mock_DecodeToken = func(m *mock.Mock_Crypto, token []byte, mods ...j.TokenModifier) (*j.TokenDef, error) {
		return j.Decode(token, j.RemoveConstraints(j.None_Algo))
	}
	s, err := NewServer(MockProvider(md))
	Ok(t, err)
	s.Define(
		View("test", apiFuncSpy),
	)
	r := s.BuildRouter()
	test_server := httptest.NewServer(r)
	uri := test_server.URL + "/test?state=eyJ0eXAiOiJKV1QiLCJhbGciOiJub25lIn0.eyJqaWQiOiJCQjgifQ."
	res, err := http.Get(uri)
	Ok(t, err)
	body, err := ioutil.ReadAll(res.Body)
	Ok(t, err)
	Equals(t, "", string(body), "spy should return empty payload")
	res.Body.Close()
	defer test_server.Close()
	Assert(t, spy_called, "apiFuncSpy was not called after Compose and Bind")
}

func Test_Should_pick_up_session_from_request_headers(t *testing.T) {
	spy_called := false
	apiFuncSpy := func(req Request, resp Response) {
		spy_called = true
	}
	md := mock.NewDependencies()
	md.Mock_Crypto.Mock_DecodeToken = func(m *mock.Mock_Crypto, token []byte, mods ...j.TokenModifier) (*j.TokenDef, error) {
		return j.Decode(token, j.RemoveConstraints(j.None_Algo))
	}
	s, err := NewServer(MockProvider(md))
	Ok(t, err)
	s.Define(
		View("test", apiFuncSpy),
	)
	r := s.BuildRouter()
	test_server := httptest.NewServer(r)
	c := &http.Client{}
	uri := test_server.URL + "/test"
	req, _ := http.NewRequest("GET", uri, nil)
	req.Header.Set("x-session-token", "eyJ0eXAiOiJKV1QiLCJhbGciOiJub25lIn0.eyJqaWQiOiJCQjgifQ.")
	res, err := c.Do(req)
	Ok(t, err)
	body, err := ioutil.ReadAll(res.Body)
	Ok(t, err)
	Equals(t, "", string(body), "spy should return empty payload")
	res.Body.Close()
	defer test_server.Close()
	Assert(t, spy_called, "apiFuncSpy was not called after Compose and Bind")
}

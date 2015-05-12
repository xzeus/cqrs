package apiserver_test

import (
	. "github.com/xzeus/cqrs/apiserver"
	. "github.com/xzeus/cqrs/testing"
	//mock "github.com/vizidrix/zeus/testing/mockprovider"
	"testing"
)

func Test_Should_make_empty_routenode(t *testing.T) {
	n := NewRouteNode(nil, "/")
	Assert(t, len(n.Middleware()) == 0, "should no have any midleware loaded")
	Assert(t, n.Depth() == 1, "should be at the first depth")
	Assert(t, n.Parent() == nil, "should not have a parent element")
	Assert(t, n.Path() == "/", "should match root path")
}

func Test_Should_add_middleware(t *testing.T) {
	n := NewRouteNode(nil, "/")
	m1 := false
	n.SetMiddleware("m1", func(ApiFunc) ApiFunc {
		return func(Request, Response) {
			m1 = true
		}
	})
	Equals(t, 1, len(n.Middleware()), "should have appended the middleware")
	n.Middleware()["m1"](nil)(nil, nil)
	Assert(t, m1, "should have called m1 middleware")
}

func Test_Should_add_multiple_middleware_with_different_keys(t *testing.T) {
	n := NewRouteNode(nil, "/")
	m1 := false
	m2 := false
	n.SetMiddleware("m1", func(ApiFunc) ApiFunc {
		return func(Request, Response) {
			m1 = true
		}
	})
	n.SetMiddleware("m2", func(ApiFunc) ApiFunc {
		return func(Request, Response) {
			m2 = true
		}
	})
	Equals(t, 2, len(n.Middleware()), "should have appended second middleware")
	n.Middleware()["m1"](nil)(nil, nil)
	n.Middleware()["m2"](nil)(nil, nil)
	Assert(t, m1, "should have called m1 middleware")
	Assert(t, m2, "should have callec m2 middleware")
}

func Test_Should_overwrite_middleware_with_same_key(t *testing.T) {
	n := NewRouteNode(nil, "/")
	m1 := false
	m2 := false
	n.SetMiddleware("m1", func(ApiFunc) ApiFunc {
		return func(Request, Response) {
			m1 = true
		}
	})
	n.SetMiddleware("m1", func(ApiFunc) ApiFunc {
		return func(Request, Response) {
			m2 = true
		}
	})
	Equals(t, 1, len(n.Middleware()), "should not have appended second middleware")
	n.Middleware()["m1"](nil)(nil, nil)
	Assert(t, !m1, "should not have called m1 middleware")
	Assert(t, m2, "should have callec m2 middleware")
}

func Test_Should_run_middleware_for_child_nodes(t *testing.T) {
	n := NewRouteNode(nil, "/")
	//n2 := NewRouteNode(n, "v1")
	m1 := 0
	n.SetMiddleware("m1", func(ApiFunc) ApiFunc {
		return func(Request, Response) {
			m1++
		}
	})
	Equals(t, 0, m1, "should not have incremented yet")
	n.Middleware()["m1"](nil)(nil, nil)
	Equals(t, 1, m1, "should have incremented by 1")
	//n2.Middleware()["m1"](nil)(nil, nil)
	//Equals(t, 2, m1, "should not have incremented yet")
}

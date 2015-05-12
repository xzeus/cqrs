package apiserver

import (
	"fmt"
	"github.com/xzeus/cqrs"
	"github.com/xzeus/cqrs/ioc"
	"log"
	"strings"
)

type ViewFunc func(deps ioc.Dependencies) (interface{}, error)
type ViewByStringFunc func(deps ioc.Dependencies, v string) (interface{}, error)
type ViewByStringsFunc func(deps ioc.Dependencies, q []string) (interface{}, error)
type ViewByStringMapFunc func(deps ioc.Dependencies, q map[string][]string) (interface{}, error)
type ViewByPageLoaderFunc func(deps ioc.Dependencies, page, offset int64) (interface{}, error)
type ViewByIdLoaderFunc func(deps ioc.Dependencies, id int64) (interface{}, error)

func Publish(req Request, id int64, p cqrs.MessageDefiner, mods ...func(*cqrs.MessageOptionsDef)) {
	deps := req.Deps()
	d := p.Domain()
	o := cqrs.NewMessageOptions(id, 0, 0)
	for _, mod := range mods {
		mod(o)
	}
	c := cqrs.NewMessage(o.Id(), o.Version(), o.Timestamp(), cqrs.NoOrigin, p)
	d.Handler(deps, c)
}

func CommandHandler(m cqrs.MessageDefiner, requireBody bool) ApiFunc {
	return func(req Request, resp Response) {
		var id int64
		if t, err := req.GetToken("session"); err != nil {
			log.Printf("No session token found so one was generated")
			id = req.Deps().Crypto().RandInt64()
		} else { // Should have been verified by middleware if required
			if err != nil {
				log.Printf("Token err [ %s ]", err)
				resp.Error("invalid session token", 40110, 500, err)
				return
			}
			session_id := t.GetId()
			log.Printf("Session [ %s ]", session_id)
			if id, err = ParseInt(session_id, 16, 64); err != nil {
				resp.Error(fmt.Sprintf("invalid session id [ %s ]", t.GetId()), 40040, 500, err)
				return
			}
		}
		log.Printf("Session Id [ %X ]", uint64(id))
		if err := req.Json(m); requireBody && err != nil {
			resp.Error(fmt.Sprintf("invalid request: [ %s ]", err), 40030, 500, err)
			return
		}
		Publish(req, id, m)
		resp.Empty(202)
	}
}

// TODO:
//func Domain(d cqrs.Domain, mods ...CommandModifier) func(RouteNode) {
// Domain(..., WhitelistCommand)
func Domain(d cqrs.Domain, mods ...func(RouteNodeHandler)) func(RouteNode) {
	return func(r RouteNode) {
		for _, f := range d.Commands() {
			c := f()
			Command(c, true, mods...)(r)
		}
	}
}

func Command(c cqrs.MessageDefiner, requireBody bool, mods ...func(RouteNodeHandler)) func(RouteNode) {
	d := c.Domain()
	n := strings.ToLower(d.MessageName(c))
	return func(r RouteNode) {
		h := NewHandler(n, []string{"POST", "OPTIONS"}, CommandHandler(c, requireBody), mods...)
		r.AppendHandler(h)
	}
}

func View(path string, handler ApiFunc, mods ...func(RouteNodeHandler)) func(RouteNode) {
	return func(r RouteNode) {
		h := NewHandler(path, []string{"GET"}, handler, mods...)
		r.AppendHandler(h)
	}
}

func NoParams(f ViewFunc) ApiFunc {
	return func(req Request, resp Response) {
		v, err := f(req.Deps())
		resp.View(v, err)
	}
}
func ByPage(f ViewByPageLoaderFunc) ApiFunc {
	return func(req Request, resp Response) {
		var page int64 = 0
		var offset int64 = 0
		v, err := f(req.Deps(), page, offset)
		resp.View(v, err)
	}
}

func ByIdFromPath(f ViewByIdLoaderFunc) ApiFunc {
	return func(req Request, resp Response) {
		if id, found := req.ExtractInt64ElementId(); !found {
			resp.Error("id not found in request", 40030, 404, nil)
		} else {
			v, err := f(req.Deps(), id)
			resp.View(v, err)
		}
	}
}

func ByStringFromPath(f ViewByStringFunc) ApiFunc {
	return func(req Request, resp Response) {
		if v, found := req.ExtractStringElementId(); !found {
			resp.Error("value not found in request", 40040, 404, nil)
		} else {
			v, err := f(req.Deps(), v)
			resp.View(v, err)
		}
	}
}

func ByIdFromSession(f ViewByIdLoaderFunc) ApiFunc {
	return func(req Request, resp Response) {
		if t, err := req.GetToken("session"); err != nil {
			resp.Error("invalid session", 40080, 500, err)
		} else { // Session token found
			if session_id, err := Int64(t.GetId()); err != nil {
				resp.Error("invalid session id", 40100, 500, err)
			} else { // Session Id was a valid value
				v, err := f(req.Deps(), session_id)
				resp.View(v, err)
			}
		}
	}
}

func ByStringFromQuery(key string, f ViewByStringsFunc) ApiFunc {
	return func(req Request, resp Response) {
		if params, ok := req.QueryParams()[key]; !ok {
			resp.Error(fmt.Sprintf("query [ %s ] not found in request", key), 40020, 404, nil)
		} else {
			v, err := f(req.Deps(), params)
			resp.View(v, err)
		}
	}
}

func ByLowerCaseStringFromQuery(key string, f ViewByStringsFunc) ApiFunc {
	return func(req Request, resp Response) {
		if params, ok := req.QueryParams()[key]; !ok {
			resp.Error(fmt.Sprintf("query [ %s ] not found in request", key), 40020, 404, nil)
		} else {
			for i, _ := range params {
				params[i] = strings.ToLower(params[i])
			}
			v, err := f(req.Deps(), params)
			resp.View(v, err)
		}
	}
}

func ByStringsFromQuery(keys []string, f ViewByStringMapFunc) ApiFunc {
	return func(req Request, resp Response) {
		r := make(map[string][]string)
		for _, key := range keys {
			if params, ok := req.QueryParams()[key]; ok {
				r[key] = params
			}
		}
		v, err := f(req.Deps(), r)
		resp.View(v, err)
	}
}

func ByLowerCaseStringsFromQuery(keys []string, f ViewByStringMapFunc) ApiFunc {
	return func(req Request, resp Response) {
		r := make(map[string][]string)
		for _, key := range keys {
			if params, ok := req.QueryParams()[key]; ok {
				for i, _ := range params {
					params[i] = strings.ToLower(params[i])
				}
				r[key] = params
			}
		}
		v, err := f(req.Deps(), r)
		resp.View(v, err)
	}
}

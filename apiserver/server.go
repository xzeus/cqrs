package apiserver

import (
	"errors"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/xzeus/cqrs/ioc"
	//"log"
	"net/http"
	"net/url"
	"runtime/debug"
	"strconv"
	"strings"
	"time"
)

var (
	ErrInvalidProvider     = errors.New("invalid dependency provider")
	ErrResponseFlushed     = errors.New("Response already flushed")
	ErrInternalServerError = errors.New("Internal Server Error")
	ErrInvalidValidation   = errors.New("Validation Error")
	ErrInvalidTask         = errors.New("invalid task id or token")
	ErrInvalidToken        = errors.New("invalid token")
	ErrInvalidId           = errors.New("invalid id")
)

const (
	URL_HEX_ID = "{id:-?[a-fA-F0-9]+}"
	URL_DEC_ID = "{id:-?[0-9]}+"
	URL_EMPTY  = ""
)

const (
	OPTIONS                    = "OPTIONS"
	XFrameOptions              = "X-Frame-Options"
	DefaultXFrameOptions       = "deny"
	XContentTypeOptions        = "X-Content-Type-Options"
	DefaultXContentTypeOptions = "nosniff"
	XRequestLatency            = "X-Request-Latency"

	CORS_AccessControlAllowOrigin  = "Access-Control-Allow-Origin"
	CORS_AccessControlAllowMethods = "Access-Control-Allow-Methods"
	CORS_DefaultMethods            = "GET, POST, OPTIONS"
	CORS_AccessControlAllowHeaders = "Access-Control-Allow-Headers"
	CORS_DefaultHeaders            = "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, X-Session-Token"
)

type ProviderFunc func(*http.Request) func() ioc.Dependencies
type MiddlewareFunc func(ApiFunc) ApiFunc
type ApiFunc func(Request, Response)

type RouteModifier func(RouteNode)
type ServerModifier func(Server)

type Handler struct {
	Methods    []string
	HandleFunc http.HandlerFunc
}

func DependencyProvider(p ProviderFunc) ServerModifier {
	return func(s Server) {
		s.SetDepsProvider(p)
	}
}

type Server interface {
	RouteNode
	Config(...ServerModifier) Server
	//
	SetDepsProvider(ProviderFunc)
	AppendEndpoint(RouteNode)
	DepsProvider() ProviderFunc

	Compose() map[string]Handler
	BuildRouter() http.Handler
}

type ServerDef struct {
	RouteNode
	deps_provider ProviderFunc
	running       bool
	route_nodes   []RouteNode
}

func NewServer(p ProviderFunc) (Server, error) {
	if p == nil {
		return nil, ErrInvalidProvider
	}
	return &ServerDef{
		RouteNode:     NewRouteNode(nil, ""),
		deps_provider: p,
		route_nodes:   make([]RouteNode, 0),
	}, nil
}

func (s *ServerDef) Config(configs ...ServerModifier) Server {
	for _, config := range configs {
		config(s)
	}
	return s
}

//

func (s *ServerDef) SetDepsProvider(provider ProviderFunc) {
	s.deps_provider = provider
}

func (s *ServerDef) AppendEndpoint(route_node RouteNode) {
	s.AppendNode(route_node)
}

//

func (s *ServerDef) DepsProvider() ProviderFunc {
	return s.deps_provider
}

func makePath(path_stack []string, leaf string) string {
	if len(path_stack) == 0 {
		return "/"
	}
	path := ""
	for _, component := range path_stack[1:len(path_stack)] {
		path += "/" + component
	}
	return path + "/" + leaf
}

func flattenMiddleware(middleware map[string]MiddlewareFunc) []MiddlewareFunc {
	r := make([]MiddlewareFunc, 0, len(middleware))
	for _, m := range middleware {
		r = append(r, m)
	}
	return r
}

func flattenMiddlewareStack(middleware_stack [][]MiddlewareFunc) []MiddlewareFunc {
	r := make([]MiddlewareFunc, 0)
	for _, set := range middleware_stack {
		r = append(r, set...)
	}
	return r
}

func (s *ServerDef) Compose() map[string]Handler {
	var _compose func(RouteNode, Server)
	r := make(map[string]Handler)
	path_stack := []string{}
	middleware_stack := make([][]MiddlewareFunc, 0)
	_compose = func(node RouteNode, s Server) {
		path_stack = append(path_stack, node.Path())
		middleware_stack = append(middleware_stack, flattenMiddleware(node.Middleware()))
		for _, handler := range node.RouteNodeHandlers() {
			h := handler
			path := makePath(path_stack, h.Path())
			middleware := flattenMiddlewareStack(middleware_stack)
			r[path] = Handler{
				Methods:    h.Methods(),
				HandleFunc: makeHandlerFunc(node, h, s, path, middleware),
			}
		}
		for _, child_node := range node.RouteNodes() {
			_compose(child_node, s)
		}
		path_stack = path_stack[0 : len(path_stack)-1]
		middleware_stack = middleware_stack[0 : len(middleware_stack)-1]
	}
	_compose(s, s)
	return r
}

func (s *ServerDef) BuildRouter() http.Handler {
	r := mux.NewRouter()
	c := s.Compose()
	for p, h := range c {
		r.HandleFunc(p, h.HandleFunc).Methods(h.Methods...)
	}
	return r
}

func makeHandlerFunc(n RouteNode, handler RouteNodeHandler, s Server, path string, middleware []MiddlewareFunc) http.HandlerFunc {
	p := s.DepsProvider()
	h := func(w http.ResponseWriter, r *http.Request) {
		start_time := time.Now()
		deps := p(r)()
		log := deps.Logger()
		log.Infof("Got request at [ %s ]", path)

		req := NewRequest(deps, r)
		resp := NewResponse(w)
		defer func() {
			if recoverErr := recover(); recoverErr != nil {
				log.Infof("\n\n\n***\n***\n***\n\nRecovered in [ %s ]\n\n\n\n", path)
				log.Infof("\n\n\n[ %v ]\n\n\n\n", recoverErr)
				log.Infof("\n\n\n%s\n\n\n***\n***\n***\n\n\n", debug.Stack())
				if err := resp.Fail(errors.New(fmt.Sprintf("%s", recoverErr))); err != nil {
				} // Attempt to write server fault failed
			} // Push the response to the client
			resp.ResponseWriter().Header().Set(XFrameOptions, DefaultXFrameOptions)             // Turn off frmes for older browsers
			resp.ResponseWriter().Header().Set(XContentTypeOptions, DefaultXContentTypeOptions) // Use explicit content types
			duration := time.Since(start_time)
			resp.ResponseWriter().Header().Set(XRequestLatency, fmt.Sprintf("%s", duration))

			resp.Flush()
		}()
		t := handler.Handler()

		for _, m := range middleware {
			t = m(t)
		}
		t(req, resp)
	}
	return h
}

func ExtractValue(url *url.URL, key string, default_value int) (result int) {
	var err error
	result_string := url.Query().Get(key)
	if result, err = strconv.Atoi(result_string); err != nil {
		result = default_value
	}
	return
}

// ExtractPage extracts album page info from request
func ExtractPage(url *url.URL, default_count, default_offset int) (count int, offset int) {
	page_str := url.Query().Get("p")
	if page_str != "" {
		count = default_count
		offset = default_offset
		page_sub_str := strings.Split(page_str, "_")
		if len(page_sub_str) == 2 {
			var err error
			if page_sub_str[0] != "" {
				count, err = strconv.Atoi(page_sub_str[0])
				if err != nil {
					count = ExtractValue(url, "count", default_count)
				}
			}
			if page_sub_str[1] != "" {
				offset, err = strconv.Atoi(page_sub_str[1])
				if err != nil {
					offset = ExtractValue(url, "offset", default_offset)
				}
			}
		}
	} else {
		count = ExtractValue(url, "count", default_count)
		offset = ExtractValue(url, "offset", default_offset)
	}
	if count < 1 {
		count = default_count
	}
	if offset < 0 {
		offset = 0
	}
	return
}

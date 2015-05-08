package apiserver

type DepthCount int

func (depth DepthCount) Tab() string {
	count := depth * 1
	result := make([]byte, count, count)
	for i := 0; i < int(count); i++ {
		result[i] = 9
	}
	return string(result)
}

type RouteNode interface {
	Define(...RouteModifier) RouteNode
	Parent() RouteNode
	Depth() DepthCount
	Path() string
	SetMiddleware(string, MiddlewareFunc)
	AppendNode(RouteNode)
	AppendHandler(RouteNodeHandler)
	Middleware() map[string]MiddlewareFunc
	RouteNodes() []RouteNode
	RouteNodeHandlers() []RouteNodeHandler
}

type RouteNodeDef struct {
	parent      RouteNode
	depth       DepthCount
	path        string
	middleware  map[string]MiddlewareFunc
	route_nodes []RouteNode
	handlers    []RouteNodeHandler
}

func NewRouteNode(parent RouteNode, path string) RouteNode {
	var depth DepthCount = 1
	if parent != nil {
		depth = parent.Depth() + 1
	}
	return &RouteNodeDef{
		parent:      parent,
		depth:       depth,
		path:        path,
		middleware:  make(map[string]MiddlewareFunc),
		route_nodes: make([]RouteNode, 0),
		handlers:    make([]RouteNodeHandler, 0),
	}
}

func (r *RouteNodeDef) Define(defs ...RouteModifier) RouteNode {
	for _, def := range defs {
		def(r)
	}
	return r
}

func (r *RouteNodeDef) Parent() RouteNode {
	return r.parent
}

func (r *RouteNodeDef) Depth() DepthCount {
	return r.depth
}

func (r *RouteNodeDef) Path() string {
	return r.path
}

func (r *RouteNodeDef) SetMiddleware(key string, handler MiddlewareFunc) {
	r.middleware[key] = handler
}

func (r *RouteNodeDef) AppendNode(route_node RouteNode) {
	r.route_nodes = append(r.route_nodes, route_node)
}

func (r *RouteNodeDef) AppendHandler(handler RouteNodeHandler) {
	r.handlers = append(r.handlers, handler)
}

func (r *RouteNodeDef) Middleware() map[string]MiddlewareFunc {
	return r.middleware
}

func (r *RouteNodeDef) RouteNodes() []RouteNode {
	return r.route_nodes
}

func (r *RouteNodeDef) RouteNodeHandlers() []RouteNodeHandler {
	return r.handlers
}

func Middleware(key string, handler MiddlewareFunc) RouteModifier {
	return func(r RouteNode) {
		r.SetMiddleware(key, handler)
	}
}

func Sub(path string, routes ...RouteModifier) RouteModifier {
	return func(r RouteNode) {
		route_node := NewRouteNode(r, path)
		r.AppendNode(route_node)
		for _, route := range routes {
			route(route_node)
		}
	}
}

type RouteNodeHandler interface {
	Methods() []string
	Path() string
	Handler() ApiFunc
	SetMethods(...string)
}

type RouteNodeHandlerDef struct {
	methods []string
	path    string
	handler ApiFunc
}

func NewHandler(path string, methods []string, handler ApiFunc, mods ...func(RouteNodeHandler)) RouteNodeHandler {
	h := &RouteNodeHandlerDef{
		methods: methods,
		path:    path,
		handler: handler,
	}
	for _, mod := range mods {
		mod(h)
	}
	return h
}

func (r *RouteNodeHandlerDef) Methods() []string {
	return r.methods
}

func (r *RouteNodeHandlerDef) Path() string {
	return r.path
}

func (r *RouteNodeHandlerDef) Handler() ApiFunc {
	return r.handler
}

func (r *RouteNodeHandlerDef) SetMethods(methods ...string) {
	r.methods = methods
}

func SetMethods(methods ...string) func(RouteNodeHandler) {
	return func(r RouteNodeHandler) {
		r.SetMethods(methods...)
	}
}

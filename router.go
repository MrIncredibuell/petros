package petros

import (
	"net/http"
	"strings"
	"unicode"
)

type Router interface {
	AddRoute(Route)
	http.Handler
}

type router struct {
	routes          []Route
	NotFoundHandler http.HandlerFunc
}

func (r *router) AddRoute(route Route) {
	r.routes = append(r.routes, route)
}

func (r *router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	path := req.URL.Path
	for _, route := range r.routes {
		if route.Match(path) {
			req = ParsedRequest(req)
			if route, ok := route.(*paramRoute); ok {
				args, err := Args(req)
				if err == nil {
					args.URLParams = route.getParams(path)
				}
			}

			route.HandlerFunc()(w, req)
			return
		}
	}

	if r.NotFoundHandler == nil {
		http.NotFound(w, req)
	}
}

func NewRouter() Router {
	return &router{}
}

type Route interface {
	Match(string) bool
	HandlerFunc() http.HandlerFunc
}
type exactMatchRoute struct {
	pattern string
	handler http.Handler
}

func (r *exactMatchRoute) Match(url string) bool {
	if len(url) > 0 && url[len(url)-1] != '/' {
		url += "/"
	}
	return url == r.pattern
}

func (r *exactMatchRoute) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.handler.ServeHTTP(w, req)
}

func (r *exactMatchRoute) HandlerFunc() http.HandlerFunc {
	return r.ServeHTTP
}

func NewExactMatchRoute(pattern string, handler http.HandlerFunc) Route {
	if len(pattern) == 0 || pattern[len(pattern)-1] != '/' {
		pattern += "/"
	}
	return &exactMatchRoute{
		pattern: pattern,
		handler: handler,
	}
}

type prefixRoute struct {
	prefix  string
	handler http.Handler
}

func (r *prefixRoute) Match(url string) bool {
	return strings.HasPrefix(url, r.prefix)
}

func (r *prefixRoute) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.handler.ServeHTTP(w, req)
}

func (r *prefixRoute) HandlerFunc() http.HandlerFunc {
	return r.ServeHTTP
}

func NewPrefixRoute(prefix string, handler http.HandlerFunc) Route {
	return &prefixRoute{
		prefix:  prefix,
		handler: handler,
	}
}

type paramRoute struct {
	// pattern  string
	segments   []string
	validators []func(string) bool
	handler    http.HandlerFunc
}

func (r *paramRoute) Match(url string) bool {
	urlSegments := removeTrailingBlanks(strings.Split(url, "/"))
	if len(urlSegments) != len(r.segments) {
		return false
	}

	args := make(map[string]string)

	validatorIndex := 0

	for i, segment := range urlSegments {
		if len(r.segments[i]) > 1 && r.segments[i][0] == ':' && len(segment) > 0 {
			if validatorIndex < len(r.validators) {
				valid := r.validators[validatorIndex](segment)
				if !valid {
					return false
				}
				validatorIndex += 1
			}
			args[r.segments[i][1:]] = segment
		} else if segment != r.segments[i] {
			return false
		}
	}

	return true
}

func (r *paramRoute) getParams(url string) map[string]string {
	urlSegments := removeTrailingBlanks(strings.Split(url, "/"))

	args := make(map[string]string)

	for i, segment := range urlSegments {
		if len(r.segments[i]) > 1 && r.segments[i][0] == ':' {
			args[r.segments[i][1:]] = segment
		}
	}

	return args
}

func removeTrailingBlanks(segments []string) []string {
	for len(segments) > 0 && segments[len(segments)-1] == "" {
		segments = segments[:len(segments)-1]
	}

	return segments
}

func (r *paramRoute) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.handler(w, req)
}

func (r *paramRoute) HandlerFunc() http.HandlerFunc {
	return r.ServeHTTP
}

func NewParamRoute(pattern string, handler http.HandlerFunc, validators ...func(string) bool) Route {
	segments := removeTrailingBlanks(strings.Split(pattern, "/"))
	return &paramRoute{
		segments:   segments,
		validators: validators,
		handler:    handler,
	}
}

func UUIDValidator(s string) bool {
	if len(s) != 36 {
		return false
	}
	for i, r := range s {
		if i == 8 || i == 13 || i == 18 || i == 23 {
			if r != '-' {
				return false
			}
		} else if !(unicode.IsDigit(r) || r == 'a' || r == 'b' || r == 'c' || r == 'd' || r == 'e' || r == 'f') {
			return false
		}
	}
	return true
}

func NewRouterWithStatic(staticPrefix string, staticDir string) Router {
	r := NewRouter()
	fs := http.FileServer(http.Dir(staticDir))
	r.AddRoute(NewPrefixRoute(staticPrefix, http.StripPrefix(staticPrefix, fs).ServeHTTP))
	return r
}

// URL router provides a simple clean way of setting up multiplexed url routes.
package urlrouter

import (
	"fmt"
	"net/http"
	"strings"
	"unicode/utf8"
)

type routes map[string]*route

type route struct {
	namedChildren *route
	urlChildren   routes
	handlerFunc   http.HandlerFunc
}

func (r *route) addUrl(path string, handlerFunc http.HandlerFunc) {
	splitPaths := []string{}

	for index, char := range path {
		if len(splitPaths) == 0 {
			splitPaths = append(splitPaths, string(char))
		} else {
			splitPaths[0] += string(char)
		}

		if char == '/' {
			endIndex := index + utf8.RuneCountInString(string(char))
			if endIndex != len(path) {
				splitPaths = append(splitPaths, path[endIndex:])
			}

			break
		}
	}

	fmt.Printf("path: %#v\n", path)
	fmt.Printf("split path: %#v\n", splitPaths)

	switch len(splitPaths) {
	case 1:
		// must be at the end

		// this is a named parameter
		if strings.HasPrefix(splitPaths[0], ":") {
			if r.namedChildren == nil {
				*r.namedChildren = route{handlerFunc: handlerFunc}
			} else {
				r.namedChildren.handlerFunc = handlerFunc
			}

			return
		}

		// this is a url path
		if childRoutes, ok := r.urlChildren[splitPaths[0]]; ok {
			childRoutes.handlerFunc = handlerFunc
		} else {
			r.urlChildren = routes{}
			r.urlChildren[splitPaths[0]] = &route{handlerFunc: handlerFunc}
		}
	default: // always will be 2
		// must be able to recurse

		// this is a named parameter
		if strings.HasPrefix(splitPaths[0], ":") {
			if r.namedChildren == nil {
				*r.namedChildren = route{}
			}

			r.namedChildren.addUrl(splitPaths[1], handlerFunc)
			return
		}

		// this is a url path
		if childRoutes, ok := r.urlChildren[splitPaths[0]]; ok {
			childRoutes.addUrl(splitPaths[1], handlerFunc)
		} else {
			r.urlChildren[splitPaths[0]] = &route{}
			r.urlChildren[splitPaths[0]].addUrl(splitPaths[1], handlerFunc)
		}
	}
}

func (route *route) serveHTTP(path string, w http.ResponseWriter, r *http.Request) {
	splitPaths := strings.SplitN(path, "/", 2)

	switch len(splitPaths) {
	case 1:
		// this is a proper url found
		if urlChild, ok := route.urlChildren[splitPaths[0]]; ok {
			urlChild.namedChildren.handlerFunc(w, r)
			return
		}

		// check to see if it is a named parameter
		if route.namedChildren != nil {
			route.namedChildren.handlerFunc(w, r)
			return
		}

		// must not be found
		http.NotFoundHandler().ServeHTTP(w, r)
	default:
	}
}

type Router struct {
	routes routes
}

func New() *Router {
	return &Router{
		routes: routes{},
	}
}

// Add a new url handler to the router. If a route already exists with the same url
// path, then this will overwrite the previous handler.
func (router *Router) HandleFunc(method string, path string, handlerFunc http.HandlerFunc) {
	var foundRoute *route

	if knownRoute, ok := router.routes[method]; ok {
		foundRoute = knownRoute
	} else {
		foundRoute = &route{}
		router.routes[method] = foundRoute
	}

	foundRoute.addUrl(path, handlerFunc)
}

func (router *Router) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	method := r.Method

	if route, ok := router.routes[method]; ok {
		route.serveHTTP(r.URL.Path, w, r)
	} else {
		http.NotFoundHandler().ServeHTTP(w, r)
	}
}

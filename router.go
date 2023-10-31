// URL router provides a simple clean way of setting up multiplexed url routes.
package urlrouter

import (
	"net/http"
)

type routes map[string]*route

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
//
//		PARAMS:
//		- method - API method to match against. Commonly one of: POST, PUT, PATCH, GET, DELETE
//		- path - The path of a URL. This will panic if path is the empty string
//	 - handlerFunc - handler callback to used when a pathi is found. This will panic if the handlerFunc is nil
func (router *Router) HandleFunc(method string, path string, handlerFunc http.HandlerFunc) {
	if path == "" {
		panic("recieved an empty path")
	}

	if handlerFunc == nil {
		panic("received and empty handler function")
	}

	var foundRoute *route

	if knownRoute, ok := router.routes[method]; ok {
		foundRoute = knownRoute
	} else {
		foundRoute = &route{childName: method}
		router.routes[method] = foundRoute
	}

	foundRoute.addUrl(path, handlerFunc)
}

func (router *Router) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	method := r.Method

	if route, ok := router.routes[method]; ok {
		if !route.serveHTTP(r.URL.Path, w, r) {
			http.NotFoundHandler().ServeHTTP(w, r)
		}
	} else {
		http.NotFoundHandler().ServeHTTP(w, r)
	}
}

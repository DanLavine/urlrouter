package urlrouter

import (
	"net/http"
	"strings"
	"unicode/utf8"
)

type route struct {
	namedChildren *route
	urlChildren   routes
	handlerFunc   http.HandlerFunc
}

func splitPahts(path string) []string {
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

	return splitPaths
}

func (r *route) addUrl(path string, handlerFunc http.HandlerFunc) {
	splitPaths := splitPahts(path)

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
			r.urlChildren = routes{}
			r.urlChildren[splitPaths[0]] = &route{}
			r.urlChildren[splitPaths[0]].addUrl(splitPaths[1], handlerFunc)
		}
	}
}

func (route *route) serveHTTP(path string, w http.ResponseWriter, r *http.Request) {
	splitPaths := splitPahts(path)

	switch len(splitPaths) {
	case 1:
		// this is a proper url found
		if urlChild, ok := route.urlChildren[splitPaths[0]]; ok {
			urlChild.handlerFunc(w, r)
			return
		}

		// check to see if it is a named parameter
		if route.namedChildren != nil {
			route.namedChildren.handlerFunc(w, r)
			return
		}

		// must not be found
		http.NotFoundHandler().ServeHTTP(w, r)
	case 2:
		// this is a proper url found
		if urlChild, ok := route.urlChildren[splitPaths[0]]; ok {
			urlChild.serveHTTP(splitPaths[1], w, r)
			return
		}

		// check to see if it is a named parameter
		if route.namedChildren != nil {
			route.namedChildren.serveHTTP(splitPaths[1], w, r)
			return
		}

		// see if there is a handler at this level to capture all unkown paths
		if route.handlerFunc != nil {
			route.handlerFunc(w, r)
			return
		}

		// must not be found
		http.NotFoundHandler().ServeHTTP(w, r)
	}
}

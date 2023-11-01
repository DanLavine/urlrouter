package urlrouter

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"unicode/utf8"
)

type urlNamedParameter string

const (
	NAMED_PAAMTERS urlNamedParameter = "urlrouter_named_parameters"
)

func GetNamedParamters(ctx context.Context) map[string]string {
	if value := ctx.Value(NAMED_PAAMTERS); value != nil {
		return value.(map[string]string)
	}

	return nil
}

func setNamedParameter(name, value string, r *http.Request) *http.Request {
	values := r.Context().Value(NAMED_PAAMTERS)

	// adding the first named parameter
	if values == nil {
		return r.WithContext(context.WithValue(r.Context(), NAMED_PAAMTERS, map[string]string{name: value}))
	}

	// insert the new parameter
	values.(map[string]string)[name] = value
	return r
}

func trimPaths(path string) string {
	path = strings.TrimPrefix(path, ":")
	return strings.TrimSuffix(path, "/")
}

type route struct {
	name          string
	namedChildren *route

	urlChildren routes
	handlerFunc http.HandlerFunc
}

// Splits strings on the "/" index each string will not start with a '/'
// If a string that is split is "", that indicates it was a "/" character and
// should be treated as a wildcard
func splitPahts(path string) []string {
	var splitPaths []string

	endIndex := 0
	for index, char := range path {
		if char == '/' {
			fmt.Println("dsl index:", index)

			// this was the case that the string started with a '/'
			if index == 0 {
				endIndex = utf8.RuneCountInString(string(char))
				splitPaths = append(splitPaths, "")
				fmt.Println("dsl splitPaths:", splitPaths)
				break
			}

			// there was a string before the '/'. I.E: 'abc/'
			endIndex = index                                 //+ utf8.RuneCountInString(string(char))
			splitPaths = append(splitPaths, path[:endIndex]) //-utf8.RuneCountInString(string(char))])

			break
		}
	}

	switch len(splitPaths) {
	case 0:
		// must just be a single value

		// return nil if it is the empty string
		if path == "" {
			return nil
		}

		// return the path that was passed in
		return []string{path}
	case 1:
		// try and append all the ending values to be parsed on the next iteration
		fmt.Println("end index:", endIndex)
		fmt.Println("path len:", len(path))
		if endIndex <= len(path)-1 {
			splitPaths = append(splitPaths, path[endIndex:])
		}
	}

	return splitPaths
}

// used to construct the url paths
func (r *route) addUrl(path string, handlerFunc http.HandlerFunc) {
	splitPaths := splitPahts(path)

	//fmt.Printf("add url splitPaths: %#v\n", splitPaths)

	switch len(splitPaths) {
	case 1:
		// must be at the end

		// this is a named parameter
		if strings.HasPrefix(splitPaths[0], ":") {
			if r.namedChildren == nil {
				r.namedChildren = &route{name: trimPaths(splitPaths[0]), handlerFunc: handlerFunc}
			} else {
				r.namedChildren.name = trimPaths(splitPaths[0])
				r.namedChildren.handlerFunc = handlerFunc
			}

			return
		}

		// this is a url path
		if r.urlChildren == nil {
			r.urlChildren = routes{}
		}

		if childRoute, ok := r.urlChildren[splitPaths[0]]; ok {
			childRoute.name = trimPaths(splitPaths[0])
			childRoute.handlerFunc = handlerFunc
		} else {
			r.urlChildren[splitPaths[0]] = &route{name: trimPaths(splitPaths[0]), handlerFunc: handlerFunc}
		}
	default: // always will be 2
		// must be able to recurse

		// this is a named parameter
		if strings.HasPrefix(splitPaths[0], ":") {
			if r.namedChildren == nil {
				r.namedChildren = &route{}
			}

			r.namedChildren.name = trimPaths(splitPaths[0])
			r.namedChildren.addUrl(splitPaths[1], handlerFunc)
			return
		}

		// this is a url path
		if r.urlChildren == nil {
			r.urlChildren = routes{}
		}

		if childRoutes, ok := r.urlChildren[splitPaths[0]]; ok {
			childRoutes.addUrl(splitPaths[1], handlerFunc)
		} else {
			r.urlChildren[splitPaths[0]] = &route{name: trimPaths(splitPaths[0])}
			r.urlChildren[splitPaths[0]].addUrl(splitPaths[1], handlerFunc)
		}
	}

	//fmt.Printf("created route: %#v\n", r)
}

// used when parsing server requests to determine which handler to use
func (route *route) serveHTTP(path string, w http.ResponseWriter, r *http.Request) bool {
	splitPaths := splitPahts(path)

	//fmt.Printf("splitPaths: %#v\n", splitPaths)
	//fmt.Printf("route: %#v\n", route)

	switch len(splitPaths) {
	case 1:
		// this is a proper url found
		if urlChild, ok := route.urlChildren[splitPaths[0]]; ok {
			urlChild.handlerFunc(w, r)
			return true
		}

		// check to see if it is a named parameter
		if route.namedChildren != nil {
			route.namedChildren.handlerFunc(w, setNamedParameter(route.namedChildren.name, trimPaths(splitPaths[0]), r))
			return true
		}
	default: // split case 2
		// this is a proper url found
		if urlChild, ok := route.urlChildren[splitPaths[0]]; ok {
			if urlChild.serveHTTP(splitPaths[1], w, r) {
				return true
			}
		}

		// check to see if it is a named parameter
		if route.namedChildren != nil {
			if route.namedChildren.serveHTTP(splitPaths[1], w, setNamedParameter(route.namedChildren.name, trimPaths(splitPaths[0]), r)) {
				return true
			}
		}
	}

	// see if there is a handler at this level to capture all unkown paths
	if route.handlerFunc != nil {
		route.handlerFunc(w, r)
		return true
	}

	// must not be found
	return false
}

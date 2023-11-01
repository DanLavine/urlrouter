package urlrouter

import (
	"context"
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
	name string

	namedChildren *route
	urlChildren   routes

	handlerFunc  http.HandlerFunc
	wildcardFunc http.HandlerFunc
}

// Splits strings on the "/" index each string will not start with a '/'
// If a string that is split is "", that indicates it was a "/" character and
// should be treated as a wildcard
func splitPaths(path string) ([]string, bool) {
	var splitPaths []string
	startIndex := 0

	if path == "" {
		return nil, false
	}

	for index, char := range path {
		if char == '/' {
			if index == 0 {
				// add the new start path of '/'
				splitPaths = append(splitPaths, string(char))
			} else {
				if startIndex == index {
					// path was a single '/'
					splitPaths = append(splitPaths, string(char))
				} else {
					// add the new paths, 1 for the string before '/' and one for the '/' char
					splitPaths = append(splitPaths, path[startIndex:index], string(char))
				}
			}

			// update the start index
			startIndex = index + utf8.RuneCountInString(string(char))
		}

	}

	// always add the final path portion if there is one
	if startIndex < len(path) {
		splitPaths = append(splitPaths, path[startIndex:])
	}

	return splitPaths, splitPaths[len(splitPaths)-1] == "/"
}

// used to construct the url paths
func (r *route) addUrl(path string, handlerFunc http.HandlerFunc) {
	splitPaths, wildcard := splitPaths(path)

	currentRoute := r
	for _, path := range splitPaths {
		// this is a named parameters
		if strings.HasPrefix(path, ":") {
			if currentRoute.namedChildren == nil {
				currentRoute.namedChildren = &route{name: trimPaths(path), handlerFunc: handlerFunc}
			} else {
				currentRoute.namedChildren.name = trimPaths(path)
			}

			// update the new route
			currentRoute = currentRoute.namedChildren
			continue
		}

		// this is url route path
		if currentRoute.urlChildren == nil {
			currentRoute.urlChildren = routes{}
		}

		if childRoute, ok := currentRoute.urlChildren[path]; ok {
			currentRoute = childRoute
		} else {
			currentRoute.urlChildren[path] = &route{name: trimPaths(path)}
			currentRoute = currentRoute.urlChildren[path]
		}
	}

	// add the handler or wildcard if it is true
	if wildcard {
		currentRoute.wildcardFunc = handlerFunc
	} else {
		currentRoute.handlerFunc = handlerFunc
	}
}

// used to parse server requests, determining which handler to use
func (r *route) serveHTTP(path string, w http.ResponseWriter, req *http.Request) bool {
	splitPaths, _ := splitPaths(path)

	if callabck := r.parseWithNamedParameters(splitPaths, req); callabck != nil {
		callabck(w, req)
		return true
	}

	return false
}

func (r *route) parseWithNamedParameters(paths []string, req *http.Request) http.HandlerFunc {
	// this is a proper url found
	if len(paths) == 0 {
		return nil
	}

	if urlChild, ok := r.urlChildren[paths[0]]; ok {
		switch len(paths) {
		case 1:
			if urlChild.handlerFunc != nil {
				return urlChild.handlerFunc
			}

			return urlChild.wildcardFunc
		default:
			callback := urlChild.parseWithNamedParameters(paths[1:], req)

			if callback != nil {
				return callback
			}

			// try to return the wild card on the chid if there is one
			return urlChild.wildcardFunc
		}
	}

	// this is a named parameter
	if r.namedChildren != nil {
		switch len(paths) {
		case 1:
			// have an exact match for a named child.
			if r.namedChildren.handlerFunc != nil {
				*req = *setNamedParameter(r.namedChildren.name, paths[0], req)
				return r.namedChildren.handlerFunc
			}

			// named children will never have wildcards
		default:
			callback := r.namedChildren.parseWithNamedParameters(paths[1:], req)

			// update the context to include the named parameter
			if callback != nil {
				*req = *setNamedParameter(r.namedChildren.name, paths[0], req)
				return callback
			}
		}
	}

	// at this point, there is nothing to return, hit a bad index
	return nil
}

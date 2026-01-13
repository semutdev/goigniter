package core

import (
	"goigniter/system/core/internal/radix"
)

// Router manages HTTP routes using a radix tree for each HTTP method.
type Router struct {
	trees map[string]*radix.Tree
}

func newRouter() *Router {
	return &Router{
		trees: make(map[string]*radix.Tree),
	}
}

func (r *Router) Add(method, pattern string, handler HandlerFunc) {
	tree, ok := r.trees[method]
	if !ok {
		tree = radix.New()
		r.trees[method] = tree
	}
	tree.Insert(pattern, handler)
}

func (r *Router) Find(method, path string) (HandlerFunc, map[string]string, bool) {
	tree, ok := r.trees[method]
	if !ok {
		return nil, nil, false
	}

	handler, params, found := tree.Search(path)
	if !found {
		return nil, nil, false
	}

	return handler.(HandlerFunc), params, true
}

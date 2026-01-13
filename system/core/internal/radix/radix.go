package radix

// Node represents a node in the radix tree.
type Node struct {
	path     string
	children []*Node
	handler  any
	param    string // parameter name (e.g., "id" for ":id")
	wildcard bool   // true if this is a wildcard node (*)
}

// Tree represents a radix tree for route matching.
type Tree struct {
	root *Node
}

// New creates a new radix tree.
func New() *Tree {
	return &Tree{
		root: &Node{
			children: make([]*Node, 0),
		},
	}
}

// Insert adds a new route pattern with its handler.
func (t *Tree) Insert(pattern string, handler any) {
	t.root.insert(pattern, handler)
}

// Search finds a handler for the given path and returns path parameters.
func (t *Tree) Search(path string) (handler any, params map[string]string, found bool) {
	params = make(map[string]string)
	handler, found = t.root.search(path, params)
	return handler, params, found
}

func (n *Node) insert(path string, handler any) {
	// Handle root path
	if path == "" || path == "/" {
		n.handler = handler
		return
	}

	// Remove leading slash for processing
	if path[0] == '/' {
		path = path[1:]
	}

	n.insertPath(path, handler)
}

func (n *Node) insertPath(path string, handler any) {
	// Find the next segment
	segment, rest := splitPath(path)

	// Check if this is a parameter segment
	if len(segment) > 0 && segment[0] == ':' {
		paramName := segment[1:]
		child := n.findParamChild()
		if child == nil {
			child = &Node{
				path:     ":",
				param:    paramName,
				children: make([]*Node, 0),
			}
			n.children = append(n.children, child)
		}
		if rest == "" {
			child.handler = handler
		} else {
			child.insertPath(rest, handler)
		}
		return
	}

	// Check if this is a wildcard segment
	if len(segment) > 0 && segment[0] == '*' {
		paramName := segment[1:]
		child := &Node{
			path:     "*",
			param:    paramName,
			wildcard: true,
			children: make([]*Node, 0),
			handler:  handler, // Wildcard always terminates
		}
		n.children = append(n.children, child)
		return
	}

	// Static segment
	child := n.findChild(segment)
	if child == nil {
		child = &Node{
			path:     segment,
			children: make([]*Node, 0),
		}
		n.children = append(n.children, child)
	}

	if rest == "" {
		child.handler = handler
	} else {
		child.insertPath(rest, handler)
	}
}

func (n *Node) search(path string, params map[string]string) (handler any, found bool) {
	// Handle root path
	if path == "" || path == "/" {
		if n.handler != nil {
			return n.handler, true
		}
		return nil, false
	}

	// Remove leading slash
	if path[0] == '/' {
		path = path[1:]
	}

	return n.searchPath(path, params)
}

func (n *Node) searchPath(path string, params map[string]string) (handler any, found bool) {
	segment, rest := splitPath(path)

	// Try static match first (highest priority)
	for _, child := range n.children {
		if child.path == segment && !child.wildcard {
			if rest == "" {
				if child.handler != nil {
					return child.handler, true
				}
			} else {
				if h, f := child.searchPath(rest, params); f {
					return h, true
				}
			}
		}
	}

	// Try parameter match
	for _, child := range n.children {
		if child.path == ":" {
			params[child.param] = segment
			if rest == "" {
				if child.handler != nil {
					return child.handler, true
				}
			} else {
				if h, f := child.searchPath(rest, params); f {
					return h, true
				}
			}
			// Backtrack if not found
			delete(params, child.param)
		}
	}

	// Try wildcard match (lowest priority, catches all remaining path)
	for _, child := range n.children {
		if child.wildcard {
			// Wildcard captures the entire remaining path
			remaining := segment
			if rest != "" {
				remaining = segment + "/" + rest
			}
			params[child.param] = remaining
			return child.handler, true
		}
	}

	return nil, false
}

func (n *Node) findChild(path string) *Node {
	for _, child := range n.children {
		if child.path == path {
			return child
		}
	}
	return nil
}

func (n *Node) findParamChild() *Node {
	for _, child := range n.children {
		if child.path == ":" {
			return child
		}
	}
	return nil
}

// splitPath splits a path into the first segment and the rest.
func splitPath(path string) (segment, rest string) {
	for i := 0; i < len(path); i++ {
		if path[i] == '/' {
			return path[:i], path[i+1:]
		}
	}
	return path, ""
}

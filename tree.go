package liquiddb

import (
	"github.com/go-errors/errors"
)

const (
	//TreeRoot is the key of the root node in the tree
	TreeRoot string = "root"
)

var (
	//ErrNotFound is returned when the requested path by Get is not found
	ErrNotFound = errors.New("Not found")
)

type normalizedData struct {
	key   []string
	value interface{}
}

//Node is a node in the tree
type Node struct {
	Key      string
	Value    interface{}
	Parent   *Node
	Path     []string
	Children map[string]*Node

	pristine bool
}

//newNode creates a new node in the tree
func newNode(key string, parent *Node) *Node {
	var parentPath []string
	if parent == nil || parent.Key == TreeRoot {
		parentPath = []string{}
	} else {
		parentPath = parent.Path
	}

	node := &Node{
		Key:      key,
		Value:    nil,
		Parent:   parent,
		Children: map[string]*Node{},
		Path:     append(parentPath, key),

		pristine: true,
	}

	if parent != nil {
		parent.Children[node.Key] = node
		parent.Value = nil
	}

	return node
}

type tree struct {
	root *Node
}

func newTree() *tree {
	return &tree{
		root: newNode(TreeRoot, nil),
	}
}

func (t tree) normalize(data map[string]interface{}, relative []string) ([]normalizedData, error) {
	res := make([]normalizedData, 0)

	var n func(interface{}, []string) error
	n = func(value interface{}, path []string) error {
		switch v := value.(type) {
		case map[string]interface{}:
			for k, v := range v {
				n(v, append(path, k))
			}
		default:
			//[]byte, string, int, int16, int32, int64, int8, float32, float64, bool, []interface{}
			res = append(res, normalizedData{path, v})
		}

		return nil
	}

	for k, v := range data {
		if err := n(v, append(relative, k)); err != nil {
			return nil, err
		}
	}

	return res, nil
}

func (t tree) findNode(path []string, autoCreate bool) *Node {
	if len(path) == 1 && path[0] == TreeRoot {
		return t.root
	}

	node := t.root
	if len(path) > 0 && path[0] == TreeRoot {
		path = path[1:]
	}

	for {
		if len(path) == 0 {
			break
		}

		key := path[0]
		path = path[1:]
		if _, ok := node.Children[key]; !ok {
			if autoCreate {
				newNode(key, node)
			} else {
				return nil
			}
		}

		node = node.Children[key]
	}

	return node
}

func (t tree) performOnNodes(data []normalizedData) []EventData {
	ops := make([]EventData, 0) //TODO: optimize

	for _, d := range data {
		for i := range d.key {
			node := t.findNode(d.key[:i+1], true)
			if node.Key == d.key[len(d.key)-1] {
				node.Value = d.value
			}

			var op EventOperation
			if node.pristine && node.Key != TreeRoot {
				op = EventOperationInsert
			} else {
				op = EventOperationUpdate
			}

			info := EventData{
				Key:       node.Key,
				Operation: op,
				Path:      node.Path,
				Value:     node.Value,
			}

			ops = append(ops, info)

			node.pristine = false
		}
	}

	return ops
}

func (t tree) do(data map[string]interface{}, relative []string) ([]EventData, error) {
	normalizedData, err := t.normalize(data, relative)
	if err != nil {
		return nil, err
	}

	return t.performOnNodes(normalizedData), nil
}

func (t tree) Set(data map[string]interface{}) ([]EventData, error) {
	ops, err := t.do(data, []string{})
	if err != nil {
		return nil, err
	}

	return ops, nil
}

func (t tree) setTreePathData(path []string, data interface{}) (EventData, error) {
	node := t.findNode(path, true)
	var op EventOperation
	if node.pristine {
		op = EventOperationInsert
	} else {
		op = EventOperationUpdate
	}

	node.Value = data
	node.pristine = false

	return EventData{
		Key:       node.Key,
		Operation: op,
		Path:      path,
		Value:     data,
	}, nil
}

func (t tree) SetPath(path []string, data interface{}) ([]EventData, error) {
	var ops []EventData

	switch d := data.(type) {
	case map[string]interface{}:
		node := t.findNode(path, true)
		diff := []string{}
		for k := range node.Children {
			if d[k] == nil {
				diff = append(diff, k)
			}
		}

		o, err := t.do(d, path)
		if err != nil {
			return nil, err
		}

		for _, k := range diff {
			deletedOps, deleted := t.Delete(append(path, k))
			if deleted {
				o = append(o, deletedOps...)
			}
		}

		ops = o
	default:
		op, err := t.setTreePathData(path, data)
		if err != nil {
			return nil, err
		}

		ops = []EventData{op}
	}

	return ops, nil
}

func (t tree) iterateDescendants(node *Node, f func(node *Node), includeSelf bool) {
	if includeSelf {
		f(node)
	}

	for _, n := range node.Children {
		t.iterateDescendants(n, f, true)
	}
}

func (t tree) Delete(path []string) ([]EventData, bool) {
	node := t.findNode(path, false)
	if node == nil {
		return nil, false
	}

	eventData := make([]EventData, 0) //TODO: optimize size

	t.iterateDescendants(node, func(node *Node) {
		eventData = append(eventData, EventData{
			Key:       node.Key,
			Operation: EventOperationDelete,
			Path:      node.Path,
			Value:     node.Value,
		})

		if node.Parent != nil { //probably root
			delete(node.Parent.Children, node.Key)
		}

		node.Value = nil
		node.Parent = nil
	}, true)

	return eventData, true
}

func (t tree) getJSON(node *Node, level int) interface{} {
	res := make(map[string]interface{})

	if node == nil {
		return nil
	}

	if len(node.Children) == 0 {
		if node.Value == nil {
			return res
		}

		return node.Value
	}

	setJSONValue := func(json map[string]interface{}, path []string, value interface{}) {
		currentJSON := json
		for i, p := range path {
			if i == len(path)-1 {
				currentJSON[p] = value
			} else {
				if currentJSON[p] == nil {
					currentJSON[p] = make(map[string]interface{})
				}

				currentJSON = currentJSON[p].(map[string]interface{})
			}
		}

	}

	t.iterateDescendants(node, func(childNode *Node) {
		//we need to iterate only the leafs
		if len(childNode.Children) == 0 {
			setJSONValue(res, childNode.Path[level:], childNode.Value)
		}
	}, false)

	return res
}

func (t tree) Get(path []string) (EventData, error) {
	node := t.findNode(path, false)

	var eventPath []string
	if node != nil {
		eventPath = node.Path
	} else {
		eventPath = path
	}

	eventValue := t.getJSON(node, len(path))

	var eventKey string
	if node != nil {
		eventKey = node.Key
	} else {
		eventKey = path[len(path)-1]
	}

	return EventData{
		Key:       eventKey,
		Operation: EventOperationGet,
		Path:      eventPath,
		Value:     eventValue,
	}, nil //TODO: not returning not found, i think its fine
}

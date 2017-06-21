package liquiddb

import (
	"fmt"

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

	return &Node{
		Key:      key,
		Value:    nil,
		Parent:   parent,
		Children: map[string]*Node{},
		Path:     append(parentPath, key),

		pristine: true,
	}
}

type tree struct {
	Root *Node
}

func newTree() *tree {
	return &tree{
		Root: newNode(TreeRoot, nil),
	}
}

func (t tree) normalize(data map[string]interface{}, relative []string) ([]normalizedData, error) {
	res := make([]normalizedData, 0)

	var n func(interface{}, []string) error
	n = func(value interface{}, path []string) error {
		//TODO: handle more types, e.g. arrays
		switch v := value.(type) {
		case []byte, string, int, int16, int32, int64, int8, float32, float64, bool:
			res = append(res, normalizedData{path, v})
		case map[string]interface{}:
			for k, v := range v {
				n(v, append(path, k))
			}
		default:
			return fmt.Errorf("Invalid datatype %s", v)
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
	node := t.Root
	for {
		if len(path) == 0 {
			break
		}

		key := path[0]
		path = path[1:]
		if _, ok := node.Children[key]; !ok {
			if autoCreate {
				node.Children[key] = newNode(key, node)
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
		node := t.findNode(d.key[:], true)
		node.Value = d.value

		var op EventOperation
		if node.pristine {
			op = EventOperationInsert
		} else {
			op = EventOperationUpdate
		}

		info := EventData{
			Key:       node.Key,
			Operation: op,
			Path:      d.key,
			Value:     d.value,
		}

		ops = append(ops, info)
		node.pristine = false
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
	return t.do(data, []string{})
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
	switch d := data.(type) {
	case map[string]interface{}:
		return t.do(d, path)
	default:
		op, err := t.setTreePathData(path, data)
		if err != nil {
			return nil, err
		}

		return []EventData{op}, nil
	}
}

func (t tree) Delete(path []string) ([]EventData, bool) {
	node := t.findNode(path, false)
	if node == nil {
		return nil, false
	}

	key := node.Key
	val := node.Value

	eventData := make([]EventData, 0) //TODO: optimize size

	var deleteNodeDescendants func(node *Node)
	deleteNodeDescendants = func(node *Node) {
		for _, n := range node.Children {
			deleteNodeDescendants(n)
		}

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
	}
	deleteNodeDescendants(node)

	return append(eventData, EventData{
		Key:       key,
		Operation: EventOperationDelete,
		Path:      path,
		Value:     val,
	}), true
}

func (t tree) Get(path []string) (EventData, error) {
	//TODO: if path is empty, return the whole json data,
	//this will come when a json model is kept in parallel with the tree
	//and probably around the time persistance is done

	//TODO: GetInt, GetString etc? - not so sure we need them now
	node := t.findNode(path, false)

	var eventPath []string
	if node != nil {
		eventPath = node.Path
	} else {
		eventPath = path
	}

	var eventValue interface{}
	if node != nil {
		eventValue = node.Value
	}

	var eventKey string
	if node != nil {
		eventKey = node.Key
	} else {
		eventKey = path[len(path) - 1]
	}

	return EventData{
		Key:       eventKey,
		Operation: EventOperationGet,
		Path:      eventPath,
		Value:     eventValue,
	}, nil //TODO: not returning not found, i think its fine
}

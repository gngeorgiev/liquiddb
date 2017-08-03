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
		if _, ok := node.Children.Get(key); !ok {
			if autoCreate {
				newNode(key, node)
			} else {
				return nil
			}
		}

		n, _ := node.Children.Get(key)
		node = n.(*Node)
	}

	return node
}

func (t tree) performOnNodes(data []normalizedData) []EventData {
	ops := make([]EventData, 0) //TODO: optimize

	for _, d := range data {
		for i := range d.key {
			node := t.findNode(d.key[:i+1], true)

			if node.Key == d.key[len(d.key)-1] {
				node.SetValue(d.value)
			}

			var op EventOperation
			if node.GetPristine() && node.Key != TreeRoot {
				op = EventOperationInsert
			} else {
				op = EventOperationUpdate
			}

			info := EventData{
				Key:       node.Key,
				Operation: op,
				Path:      node.Path,
				Value:     node.GetValue(),
			}

			ops = append(ops, info)

			node.SetPristine(false)
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
	if node.GetPristine() {
		op = EventOperationInsert
	} else {
		op = EventOperationUpdate
	}

	node.SetValue(data)
	node.SetPristine(false)

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

		for item := range node.Children.IterBuffered() {
			if d[item.Key] == nil {
				diff = append(diff, item.Key)
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
	//TODO: can we lock on a less general places without introducing data races?
	if includeSelf {
		f(node)
	}

	for item := range node.Children.IterBuffered() {
		t.iterateDescendants(item.Val.(*Node), f, true)
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
			Value:     node.value,
		})

		parent := node.GetParent()
		if parent != nil { //probably root
			parent.Children.Remove(node.Key)
		}

		node.SetValue(nil)
		node.SetParent(nil)
	}, true)

	return eventData, true
}

func (t tree) getJSON(node *Node, level int) interface{} {
	res := make(map[string]interface{})

	if node == nil {
		return nil
	}

	if node.Children.Count() == 0 {
		val := node.GetValue()
		if val == nil {
			return res
		}

		return val
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
		if childNode.Children.Count() == 0 {
			setJSONValue(res, childNode.Path[level:], childNode.GetValue())
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
	}, nil //TODO: not returning not found, i think it's fine
}

package liquiddb

import (
	"github.com/sasha-s/go-deadlock"

	"github.com/orcaman/concurrent-map"
)

//Node is a node in the tree
type Node struct {
	Key      string
	Path     []string
	Children cmap.ConcurrentMap

	valueMutex deadlock.Mutex
	value      interface{}

	parentMutex deadlock.Mutex
	parent      *Node

	pristineMutex deadlock.Mutex
	pristine      bool
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
		value:    nil,
		parent:   parent,
		Children: cmap.New(),
		Path:     append(parentPath, key),

		pristine: true,
	}

	if parent != nil {
		parent.Children.Set(node.Key, node)
		parent.SetValue(nil)
	}

	return node
}

func (n *Node) GetValue() interface{} {
	n.valueMutex.Lock()
	defer n.valueMutex.Unlock()

	return n.value
}

func (n *Node) SetValue(v interface{}) {
	n.valueMutex.Lock()
	defer n.valueMutex.Unlock()

	n.value = v
}

func (n *Node) SetPristine(p bool) {
	n.pristineMutex.Lock()
	defer n.pristineMutex.Unlock()

	n.pristine = p
}

func (n *Node) GetPristine() bool {
	n.pristineMutex.Lock()
	defer n.pristineMutex.Unlock()

	return n.pristine
}

func (n *Node) SetParent(newParent *Node) {
	n.parentMutex.Lock()
	defer n.parentMutex.Unlock()

	n.parent = newParent
}

func (n *Node) GetParent() *Node {
	n.parentMutex.Lock()
	defer n.parentMutex.Unlock()

	return n.parent
}

package rgbtree

import "fmt"

type color bool

const (
	black, red color = true, false
)

/*Node - ветвь*/
type Node struct {
	Key    uint64
	Value  int
	color  color
	parent *Node // родительская
	left   *Node
	right  *Node
}

/*UTree - дерево*/
type UTree struct {
	root  *Node
	nodes []Node
	size  int
	count int
}

func (tree *UTree) newNode(key uint64, val int, color color) *Node {
	if tree.count >= len(tree.nodes) {
		fmt.Println("Add++")
		tree.nodes = append(tree.nodes, Node{})
	}
	tree.nodes[tree.count].left = nil
	tree.nodes[tree.count].right = nil
	tree.nodes[tree.count].parent = nil
	tree.nodes[tree.count].Key = key
	tree.nodes[tree.count].Value = val
	tree.nodes[tree.count].color = color
	//}
	defer func() { tree.count++ }()
	return &tree.nodes[tree.count]
}

/*NewUTree - новое дерево*/
func NewUTree(nodes []Node) UTree {
	return UTree{nil, nodes, 0, 0}
}

/*Put - новое значение*/
// Put inserts node into the tree.
// Key should adhere to the comparator's type assertion, otherwise method panics.
func (tree *UTree) Put(key uint64, value int) {
	var insertedNode *Node
	if tree.root == nil {
		// Assert key is of comparator's type for initial tree

		tree.root = tree.newNode(key, value, red)
		insertedNode = tree.root
	} else {
		node := tree.root
		loop := true
		for loop {

			switch {
			case key == node.Key:
				node.Key = key
				node.Value = value
				return
			case key < node.Key:
				if node.left == nil {
					node.left = tree.newNode(key, value, red) //&Node{Key: key, Value: value, color: red}
					insertedNode = node.left
					loop = false
				} else {
					node = node.left
				}
			case key > node.Key:
				if node.right == nil {
					node.right = tree.newNode(key, value, red) //&Node{Key: key, Value: value, color: red}
					insertedNode = node.right
					loop = false
				} else {
					node = node.right
				}
			}
		}
		insertedNode.parent = node
	}
	tree.insertCase1(insertedNode)
	tree.size++
}

/*lookup - поиск*/
func (tree *UTree) lookup(key uint64) *Node {
	node := tree.root
	for node != nil {

		switch {
		case node.Key == key:
			return node
		case key < node.Key:
			node = node.left
		case key > node.Key:
			node = node.right
		}
	}
	return nil
}

func (node *Node) grandparent() *Node {
	if node != nil && node.parent != nil {
		return node.parent.parent
	}
	return nil
}

func (node *Node) uncle() *Node {
	if node == nil || node.parent == nil || node.parent.parent == nil {
		return nil
	}
	return node.parent.sibling()
}

/*Get - получение*/
func (tree *UTree) Get(key uint64) (value int, found bool) {
	node := tree.lookup(key)
	if node != nil {
		return node.Value, true
	}
	return 0, false
}

/*Remove - удаление*/
func (tree *UTree) Remove(key uint64) {
	var child *Node
	node := tree.lookup(key)
	if node == nil { // не найдено
		return
	}
	if node.left != nil && node.right != nil {
		pred := node.left.maximumNode()
		node.Key = pred.Key
		node.Value = pred.Value
		node = pred
	}
	if node.left == nil || node.right == nil {
		if node.right == nil {
			child = node.left
		} else {
			child = node.right
		}
		if node.color == black {
			node.color = nodeColor(child)
			tree.deleteCase1(node)
		}
		tree.replaceNode(node, child)
		if node.parent == nil && child != nil {
			child.color = black
		}
	}
	tree.size--
}

func (node *Node) maximumNode() *Node {
	if node == nil {
		return nil
	}
	for node.right != nil {
		node = node.right
	}
	return node
}

func nodeColor(node *Node) color {
	if node == nil {
		return black
	}
	return node.color
}

func (tree *UTree) deleteCase1(node *Node) {
	if node.parent == nil {
		return
	}
	tree.deleteCase2(node)
}

func (tree *UTree) deleteCase2(node *Node) {
	sibling := node.sibling()
	if nodeColor(sibling) == red {
		node.parent.color = red
		sibling.color = black
		if node == node.parent.left {
			tree.rotateLeft(node.parent)
		} else {
			tree.rotateRight(node.parent)
		}
	}
	tree.deleteCase3(node)
}

func (tree *UTree) deleteCase3(node *Node) {
	sibling := node.sibling()
	if nodeColor(node.parent) == black &&
		nodeColor(sibling) == black &&
		nodeColor(sibling.left) == black &&
		nodeColor(sibling.right) == black {
		sibling.color = red
		tree.deleteCase1(node.parent)
	} else {
		tree.deleteCase4(node)
	}
}

func (tree *UTree) deleteCase4(node *Node) {
	sibling := node.sibling()
	if nodeColor(node.parent) == red &&
		nodeColor(sibling) == black &&
		nodeColor(sibling.left) == black &&
		nodeColor(sibling.right) == black {
		sibling.color = red
		node.parent.color = black
	} else {
		tree.deleteCase5(node)
	}
}

func (tree *UTree) deleteCase5(node *Node) {
	sibling := node.sibling()
	if node == node.parent.left &&
		nodeColor(sibling) == black &&
		nodeColor(sibling.left) == red &&
		nodeColor(sibling.right) == black {
		sibling.color = red
		sibling.left.color = black
		tree.rotateRight(sibling)
	} else if node == node.parent.right &&
		nodeColor(sibling) == black &&
		nodeColor(sibling.right) == red &&
		nodeColor(sibling.left) == black {
		sibling.color = red
		sibling.right.color = black
		tree.rotateLeft(sibling)
	}
	tree.deleteCase6(node)
}

func (tree *UTree) deleteCase6(node *Node) {
	sibling := node.sibling()
	sibling.color = nodeColor(node.parent)
	node.parent.color = black
	if node == node.parent.left && nodeColor(sibling.right) == red {
		sibling.right.color = black
		tree.rotateLeft(node.parent)
	} else if nodeColor(sibling.left) == red {
		sibling.left.color = black
		tree.rotateRight(node.parent)
	}
}

func (tree *UTree) rotateLeft(node *Node) {
	right := node.right
	tree.replaceNode(node, right)
	node.right = right.left
	if right.left != nil {
		right.left.parent = node
	}
	right.left = node
	node.parent = right
}

func (tree *UTree) rotateRight(node *Node) {
	left := node.left
	tree.replaceNode(node, left)
	node.left = left.right
	if left.right != nil {
		left.right.parent = node
	}
	left.right = node
	node.parent = left
}
func (tree *UTree) replaceNode(old *Node, new *Node) {
	if old.parent == nil {
		tree.root = new
	} else {
		if old == old.parent.left {
			old.parent.left = new
		} else {
			old.parent.right = new
		}
	}
	if new != nil {
		new.parent = old.parent
	}
}

func (node *Node) sibling() *Node {
	if node == nil || node.parent == nil {
		return nil
	}
	if node == node.parent.left {
		return node.parent.right
	}
	return node.parent.left
}

func (tree *UTree) insertCase1(node *Node) {
	if node.parent == nil {
		node.color = black
	} else {
		tree.insertCase2(node)
	}
}

func (tree *UTree) insertCase2(node *Node) {
	if nodeColor(node.parent) == black {
		return
	}
	tree.insertCase3(node)
}

func (tree *UTree) insertCase3(node *Node) {
	uncle := node.uncle()
	if nodeColor(uncle) == red {
		node.parent.color = black
		uncle.color = black
		node.grandparent().color = red
		tree.insertCase1(node.grandparent())
	} else {
		tree.insertCase4(node)
	}
}

func (tree *UTree) insertCase4(node *Node) {
	grandparent := node.grandparent()
	if node == node.parent.right && node.parent == grandparent.left {
		tree.rotateLeft(node.parent)
		node = node.left
	} else if node == node.parent.left && node.parent == grandparent.right {
		tree.rotateRight(node.parent)
		node = node.right
	}
	tree.insertCase5(node)
}

func (tree *UTree) insertCase5(node *Node) {
	node.parent.color = black
	grandparent := node.grandparent()
	grandparent.color = red
	if node == node.parent.left && node.parent == grandparent.left {
		tree.rotateRight(grandparent)
	} else if node == node.parent.right && node.parent == grandparent.right {
		tree.rotateLeft(grandparent)
	}
}

/*Clear - Удаление всех элементов*/
func (tree *UTree) Clear() {
	tree.root = nil
	tree.size = 0
	tree.count = 0
}

// Empty returns true if tree does not contain any nodes
func (tree *UTree) Empty() bool {
	return tree.size == 0
}

// Size returns number of nodes in the tree.
func (tree *UTree) Size() int {
	return tree.size
}

// String returns a string representation of container
func (tree *UTree) String() string {
	str := "RedBlackTree\n"
	if !tree.Empty() {
		output(tree.root, "", true, &str)
	}
	return str
}

func (node *Node) String() string {
	return fmt.Sprintf("%v", node.Key)
}

func output(node *Node, prefix string, isTail bool, str *string) {
	if node.right != nil {
		newPrefix := prefix
		if isTail {
			newPrefix += "│   "
		} else {
			newPrefix += "    "
		}
		output(node.right, newPrefix, false, str)
	}
	*str += prefix
	if isTail {
		*str += "└── "
	} else {
		*str += "┌── "
	}
	*str += node.String() + "\n"
	if node.left != nil {
		newPrefix := prefix
		if isTail {
			newPrefix += "    "
		} else {
			newPrefix += "│   "
		}
		output(node.left, newPrefix, true, str)
	}
}

// Left returns the left-most (min) node or nil if tree is empty.
func (tree *UTree) Left() *Node {
	var parent *Node
	current := tree.root
	for current != nil {
		parent = current
		current = current.left
	}
	return parent
}

// Right returns the right-most (max) node or nil if tree is empty.
func (tree *UTree) Right() *Node {
	var parent *Node
	current := tree.root
	for current != nil {
		parent = current
		current = current.right
	}
	return parent
}

func (tree *UTree) Keys(buff []uint64) []uint64 {
	keys := buff
	it := tree.Iterator()
	for i := 0; it.Next(); i++ {
		keys[i] = it.Key()
	}
	return keys[:tree.size]
}

// Values returns all values in-order based on the key.
func (tree *UTree) Values(buff []int) []int {
	values := buff
	it := tree.Iterator()
	for i := 0; it.Next(); i++ {
		values[i] = it.Value()
	}
	return values[:tree.size]
}

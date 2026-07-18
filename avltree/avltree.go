package avltree

type Node struct {
	key      int
	value    int
	parent   *Node
	Children [2]*Node
	size     int
}

type AvlTree struct {
	root *Node
	size int
}

func NewAvlTree() *AvlTree {
	return new(AvlTree)
}

func (tree *AvlTree) Put(key, value int) {

}

func (tree *AvlTree) Get(key int) (value int, found bool) {
	res := tree.GetNode(key)
	if res != nil {
		return res.value, true
	}
	return value, false
}

func (tree *AvlTree) GetNode(key int) *Node {
	res := tree.root
	for res != nil {
		if key == res.key {
			return res
		} else if key < res.key {
			res = res.Children[0]
		} else {
			res = res.Children[1]
		}

	}
	return res
}

func (tree *AvlTree) Remove(key int) {

}

func (tree *AvlTree) IndexOf(idx int) *Node {
	return nil
}

func (tree *AvlTree) put(key, value int, parent *Node, cur *Node) bool {
	if cur == nil {
		tree.size++
		cur = &Node{key: key, value: value, parent: parent}
	}

	var next *Node
	if key == cur.key {
		cur.key, cur.value = key, value
		return false
	} else if key < cur.key {
		next = cur.Children[0]
	} else {
		next = cur.Children[1]
	}
	// TODO:fixup
	return tree.put(key, value, cur, next)
}

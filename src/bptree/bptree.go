// single-writer in-memory B+ tree implementation
// uncomment for testing
//package main
package bptree

import (
	"errors"
	"fmt"
	//"math/rand"	//Uncomment for test
	"strconv"
	//"os"		//Uncomment for test
)

type Item uint64	// key

type children []*node
type items []Item	// slice of keys
type values []string	// slice of values

// tree
type Bptree struct {
	degree	int
	length	int
	root	*node
}

// tree node
type node struct {
	leaf bool
	level int
	keys	items
	children children
	vals	values
	parent	*node
	prev	*node
	next	*node
}

// find idx in key slice where key should insert
func (k *items) find (key Item) (index int, found bool) {
	var i int
	// sequential search through key slice
	// TBD: binary-search for large slices 
	for i=0; i < len(*k); i++ {
		if key < (*k)[i] {
			return i, false
		} else if key == (*k)[i] {
			return i, true
		}
	}
	return i, false
}

// recursive search for Leaf where key ought to reside
func (n *node) findLeaf(key Item) *node {
	if !n.leaf {
		var i int
		for i=0; i<len(n.keys); i++ {
			switch {
			case key < n.keys[i]:
			// search child i+1
				n = n.children[i]
				n = n.findLeaf(key)
				return n
			case key > n.keys[i]:
			// skip to next entry	
			default:
			// search child i+1
				n = n.children[i+1]
				return n.findLeaf(key)
			}
		}
		n = n.children[i]
		return n.findLeaf(key)
	}
	return n
}

// find N next larger keys and return their values
func (n *node) getNextN(key Item, N int) []string {
	// first find the leaf node
	n = n.findLeaf(key)
	idx, found := n.keys.find(key)
	if found {
		idx+=1
	}
	var nextN []string

	// traverse leaf link list until N copied
	for i:=0; i<N;  {
		if idx < len(n.keys) {
			nextN = append(nextN, n.vals[idx])
			idx++
			i++
		} else {
			n = n.next
			idx = 0
		}
	}
	return nextN
}

// find key,value starting at leaf
func (n *node) get(key Item) string {
	n = n.findLeaf(key)
	idx, found := n.keys.find(key)
	if found {
		return n.vals[idx]
	}
	return ""
}

// return sibling node ptr
func (n *node) sibling() *node {
	p := n.prev
	q := n.next
	if n.parent == q.parent {
		return q
	}
	return p
}

// remove key,value pair from leaf
func (n *node) removeKey(idx int) {
	nks := make(items, len(n.keys)-1)
	nvs := make([]string, len(n.vals)-1)
	if idx == 0 {
		copy(nks, n.keys[idx+1:])
		n.keys = nks
		copy(nvs, n.vals[idx+1:])
		n.vals = nvs
	} else {
		copy(nks, n.keys[:idx])
		copy(nks[idx:], n.keys[idx+1:])
		n.keys = nks
		copy(nvs, n.vals[:idx])
		copy(nvs[idx:], n.vals[idx+1:])
		n.vals = nvs
	}
}

// redistribute key,vals
// sib has more keys
func (n *node) redistribLeaf(idx int, sib *node, maxk int) {
	if n.keys[0] > sib.keys[0] {
		// redistribute len(sib)-maxk/2 keys from sib into n
		sks := make(items, maxk/2)
		copy(sks, sib.keys[:maxk/2])
		nks := make(items, len(sib.keys)-maxk/2)
		copy(nks, sib.keys[maxk/2:])
		nks = append(nks, n.keys...)
		n.keys = nks
		sib.keys = sks
		svs := make([]string, maxk/2)
		// redistribute len(sib)-maxk/2 vals from sib into n
		copy(svs, sib.vals[:maxk/2])
		nvs := make([]string, len(sib.vals)-maxk/2)
		copy(nvs, sib.vals[maxk/2:])
		nvs = append(nvs, n.vals...)
		n.vals = nvs
		sib.vals = svs
	} else {
		// redistribute up to max/2 k,v from sib into n
		nks := make(items, len(n.keys))
		copy(nks, n.keys)
		nks = append(nks, sib.keys[:maxk/2]...)
		n.keys = nks
		nvs := make([]string, maxk/2)
		copy(nvs, n.vals)
		nvs = append(nvs, sib.vals[:maxk/2]...)
		n.vals = nvs
		// shift remaining k,v in the sib node
		nks = make(items, len(sib.keys)-maxk/2)
		copy(nks, sib.keys[maxk/2:])
		sib.keys = nks
		nvs = make([]string, len(sib.vals)-maxk/2)
		copy(nvs, sib.vals[maxk/2:])
		sib.vals = nvs
	}
}

// fix key in parent internal node after children keys are redistrib'd
func (n *node) fixup(key Item) {
	for i:=0; i < len(n.keys); i++ {
		if n.keys[i] != n.children[i+1].keys[0] {
			n.keys[i] = n.children[i+1].keys[0]
			return
		}
	}
}

// delete key from tree starting at leaf
func (n* node) del(key Item, maxk int) (bool, string) {
	n = n.findLeaf(key)
	idx, found := n.keys.find(key)
	if !found {
		return false, ""
	}
	// found key,value pair in leaf
	retval := n.vals[idx]

	// check if deletion will lead to too few keys in node
	if (len(n.keys)-1) <= maxk/2 {
		sib := n.sibling()
		if len(sib.keys) > maxk/2 {
			// remove key from n
			n.removeKey(idx)
			// redistribute with sibling
			n.redistribLeaf(idx, sib, maxk)
			// fixup parent key
			n.parent.fixup(key)
		} else {
		//remove & merge with sibling
		}
		return true, retval
	}
	// remove k,v pair
	n.removeKey(idx)
	return true, retval
}

// return min key in the tree
func (n *node) minKey() Item {
	if n.leaf {
		return n.keys[0]
	}
	n = n.children[0]
	return n.minKey()
}

// return max key in the tree
func (n *node) maxKey() Item {
	if n.leaf {
		return n.keys[len(n.keys)-1]
	}
	n = n.children[len(n.keys)]
	return n.maxKey()
}

// insert into Leaf node
// may grow bigger than max degree
// split will happen in caller
func (n *node) insertInLeaf(key Item, value string) {
	idx, found := n.keys.find(key)
	if !found {
		if  idx < len(n.keys) {
		// insert in beginning or middle
			nks := make(items, len(n.keys)+1)
			copy(nks[idx+1:], n.keys[idx:])
			copy(nks[:idx], n.keys[:idx])
			nks[idx] = key
			n.keys = nks
			n.vals = append(n.vals, "")
			copy(n.vals[idx+1:], n.vals[idx:])
			n.vals[idx] = value
		} else {
			n.vals = append(n.vals, value)
			n.keys = append(n.keys, key)
		}
	} else {
	// replace pair
		n.keys[idx] = key
		n.vals[idx] = value
	}
}

// insert into internal node
// handles split on exceeding max degree
func (n *node) insertDir(lchld *node, rchld *node, maxk int) *node {
	newn := n
	if len(n.keys) == 0 {
		// insert in new parent
		nchn := make([]*node, 2)
		n.children = nchn
		n.children[0] = lchld
		n.children[1] = rchld
		n.keys = append(n.keys, rchld.minKey())
	} else {
		idx, _ := n.keys.find(rchld.keys[0])
		switch idx {
		case len(n.keys):
			// Append rchld key
			n.keys = append(n.keys, rchld.minKey())
			// Append rchld @idx
			n.children = append(n.children, rchld)
		default:
			// Insert rchld key @idx
			nks := make(items, len(n.keys)+1)
			copy(nks[idx+1:], n.keys[idx:])
			copy(nks[:idx], n.keys[:idx])
			nks[idx] = rchld.minKey()
			n.keys = nks
			// Shift children right
			n.children = append(n.children, nil)
			copy(n.children[idx+1:], n.children[idx:])
			// Replace rchld @idx+1
			n.children[idx+1] = rchld
		}
	}
	if len(n.keys) > maxk {
		// split internal node
		newnd := n.split()

		// link siblings
		n.linkSiblings(newnd)

		// insert into parent of internal node
		if n.parent == nil {
			n.parent = new(node)
			newnd.parent = n.parent
			n.parent.level = (n.level + 1)
			newn = n.parent.insertDir(n, newnd, maxk)
		} else {
			newnd.parent = n.parent
			newn = n.parent.insertDir(n, newnd, maxk)
		}
	}
	return newn
}

// link siblings in split
func (n *node) linkSiblings(newnd *node) {
	if n.prev == n {
		n.next = newnd
		n.prev = newnd
		newnd.prev = n
		newnd.next = n
	} else if n.prev != n {
		n.next.prev = newnd
		newnd.prev = n
		newnd.next = n.next
		n.next = newnd
	}
}

// insert into tree starting at leaf node
func (n *node) insert(key Item, value string, lchld *node, rchld *node, maxk int) *node {
	var root, newroot *node
	root = n

	n = n.findLeaf(key)
	n.insertInLeaf(key, value)

	if len(n.keys) > maxk {
		// split leaf
		newnd := n.split()

		// link siblings
		n.linkSiblings(newnd)

		// insert siblings into parent of leaf
		if n.parent == nil {
			// new parent for both siblings
			n.parent = new(node)
			newnd.parent = n.parent
			// set parent level to 1
			n.parent.level = (n.level + 1)
			// set next,prev of parent to itself
			n.parent.next = n.parent
			n.parent.prev= n.parent
			// insert siblings into parent
			newroot = n.parent.insertDir(n, newnd, maxk)
		} else {
			newnd.parent = n.parent
			// insert siblings into parent
			newroot = n.parent.insertDir(n, newnd, maxk)
		}
	}
	if newroot != nil {
		if newroot.parent == nil {
			return newroot
		}
	}
	return root
}

// split a node (leaf or internal)
func (n *node) split() *node {
	nn := new(node)
	p := len(n.keys)/2
	nn.level = n.level
	if n.leaf {
		// split keys
		nn.keys = make([]Item, len(n.keys[p:]))
		copy(nn.keys, n.keys[p:])
		nks := make([]Item, p)
		copy(nks, n.keys[:p])
		n.keys = nks
		// split values
		nn.vals = make([]string, len(n.vals[p:]))
		copy(nn.vals, n.vals[p:])
		nvs := make([]string, p)
		copy(nvs, n.vals[:p])
		n.vals = nvs
		nn.leaf = true
	} else {
		// distribute children
		nn.children = make([]*node, p)
		copy(nn.children, n.children[p+1:])
		ncs := make([]*node, p+1)
		copy(ncs, n.children[:p+1])
		n.children = ncs
		// distribute keys
		nn.keys = make([]Item, p-1)
		copy(nn.keys, n.keys[p+1:])
		nks := make([]Item, p)
		copy(nks, n.keys[:p])
		n.keys = nks
		// update parent links
		for i:=0; i < len(nn.children); i++ {
			nn.children[i].parent = nn
		}
		for i:=0; i < len(n.children); i++ {
			n.children[i].parent = n
		}
	}
	return nn
}

// print Item
func (a Item) Print() {
	fmt.Printf("%s", strconv.AppendUint(make([]byte, 0), uint64(a), 10))
}

// print the tree BFS nodes
func (n *node) printnode() {
	if !n.leaf {
		if n.parent == nil {
			fmt.Printf("\nl%d:", n.level)
			fmt.Println(n.keys)
			fmt.Printf("\nl%d:", n.level-1)
		} else {
			fmt.Printf("\nl%d:", n.level-1)
		}

		if !n.children[0].leaf {
			for i:=0; i < len(n.children); i++ {
				if i < len(n.children)-1 {
					fmt.Printf("%d, ", n.children[i].keys)
				} else {
					fmt.Printf("%d\n", n.children[i].keys)
				}
			}
		}
		for i:=0; i < len(n.children); i++ {
			n.children[i].printnode()
		}
		return
	} else {
		fmt.Printf("[ ")
		for i:=0; i < len(n.keys); i++ {
			n.keys[i].Print()
			fmt.Printf(":%s ", n.vals[i])
		}
		fmt.Printf(" ] ")
	}
}

//
// *** Main APIs below here ***
//

// create a new tree
func New(degree int) (*Bptree, error) {
	if degree < 3 {
		return nil, errors.New("Minimum degree 3")
	}
	return &Bptree{degree: degree}, nil
}

// insert into tree
func (tree *Bptree) Insert(key Item, value string) Item {
	fmt.Println("Inserting", key, value)
	if tree.root == nil {
		tree.root = new(node)
		tree.root.leaf = true
		tree.root.keys = append(tree.root.keys, key)
		tree.root.vals = append(tree.root.vals, value)
		tree.root.next = tree.root
		tree.root.prev = tree.root
		tree.length++
	} else {
		tree.root = tree.root.insert(key, value, nil, nil, tree.degree)
	}
	//tree.Print()
	return key
}

// delete key from tree
func (tree *Bptree) Del(key Item) (bool, string) {
	if tree.root == nil {
		return false, ""
	}
	return tree.root.del(key, tree.degree)
}

// get value at key
func (tree *Bptree) Get(key Item) string {
	if tree.root == nil {
		return ""
	}
	return tree.root.get(key)
}

// get N node values at nodes greater than key
func (tree *Bptree) GetNextN (key Item, N int) []string {
	if tree.root == nil {
		return nil
	}
	return tree.root.getNextN(key, N)
}

// print the whole tree
func (tree *Bptree) Print() {
	if tree.root == nil {
		return
	}
	fmt.Println("Min:", tree.root.minKey())
	fmt.Println("Max:", tree.root.maxKey())
	tree.root.printnode()
	fmt.Println()
}

/*/** Test Driver: Uncomment to test
func main() {
	bt, err := New(3)
	if  err != nil {
		fmt.Println(err)
		return
	}

	k := rand.Uint32()
	for i:=0; i < 20; i++ {
		bt.Insert(Item(k), "val"+strconv.Itoa(i))
		k = rand.Uint32()
	}

	bt.Print()
	if len(os.Args) < 2 {
		fmt.Println("Usage:./bptree <some key>")
		return
	}
	ii, err := strconv.Atoi(os.Args[1])
	if err != nil {
		fmt.Println("Arg", os.Args[1], " invalid")
	}
	fmt.Println("Get:key=", ii, "val=", bt.Get(Item(ii)))
	fmt.Println("GetNext", 3, "from", ii, ":", bt.GetNextN(Item(ii), 3))
	_, retval := bt.Del(Item(ii))
	fmt.Println("Del:key=", ii, "val=", retval)
	bt.Print()
}
*/

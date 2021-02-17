/*
   Implementation of nest tree

   NOTE: nests that don't have base graph nodes inside ARE allowed. Even if a nest doesn't
   		 have child nests, it is still allowed to have no base graph nodes inside
*/

package graph

import (
	"errors"
)

const NT_ROOT_NEST_LEVEL = 0

// Variables of the below type map printable nest properties to actual nest attributes.
// For example, if the "LabelAttr" field is not "nil" - i.e. equal to a pointer to some
// nest string attribute - then it means that "label" property is represented by this
// attribute
type NestEmitSpec struct {
	// Label of a nest
	LabelAttr *NestStrAttr
}

// Type representing string attribute of nests and nest tree as a whole
type nestTreeStrAttr struct {
	// Number of an attribute in an array of string attributes
	attr_num int
	// Whether the attribute is valid
	is_valid bool
	// Reference to a nest tree
	nestTree *NestTree
}

// Type representing nest string attribute
type NestStrAttr nestTreeStrAttr

// Representation of the invalid nest string attribute
var nest_str_attr_invalid = NestStrAttr{-1, false, nil}

// Nest representation
type Nest struct {
	// Unique ID of a nest
	id int
	// Nest tree to which a nest belongs
	nestTree *NestTree
	// Level of a nest relative to the root of the tree. The root nest has level "0", its
	// childs have level "1" and so on
	level int
	// Parent (or outer) nest of a nest
	parentNest *Nest
	// First child (or inner) nest of a nest
	firstChildNest *Nest
	// Last child nest (needed to make traversal of a nest tree in the reverse
	// direction more efficient)
	lastChildNest *Nest
	// Next nest that belongs to the same parent nest
	nextSiblingNest *Nest
	// Previous nest that belongs to the same parent nest
	prevSiblingNest *Nest
	// First graph node belonging to a nest
	firstNode *Node
	// Last graph node belonging to a nest (needed to make traversal of graph nodes in the
	// reverse direction more efficient)
	lastNode *Node
	// First graph edge belonging to a nest
	// In contrast to nodes, there is no link to "last edge" belonging to a nest. That's
	// because no out-of-the-box iterator for graph edges is provided now (the "last node"
	// link exists to aid in iterating over graph nodes). To iterate over graph edges, one
	// is expected to iterate over graph nodes and then iterate over their adjacent edges
	firstEdge *Edge
	// Array of string attributes
	strAttrs []strAttrVal
}

// Nest tree representation
type NestTree struct {
	// Graph for which a nest tree is built
	baseGraph *Graph
	// Nest count (must always increase and never decrease. Even if nests get deleted).
	// Creation of a new node increments the counter
	nestCount int
	// Root nest of a tree
	rootNest *Nest
	// Allocation map for nest string attributes
	// An element holds TRUE if corresponding attribute is allocated and FALSE
	// in the opposite case
	nestStrAttrAllocMap []bool
}

// Get unique ID of a nest
func (nest *Nest) GetID() int {
	return nest.id
}

// Get nest tree to which a nest belongs
func (nest *Nest) GetNestTree() *NestTree {
	return nest.nestTree
}

// Get parent (or outer) nest of a nest
func (nest *Nest) GetParentNest() *Nest {
	return nest.parentNest
}

// Get first child (or inner) nest of a nest
func (nest *Nest) GetFirstChildNest() *Nest {
	return nest.firstChildNest
}

// Get last child (or inner) nest of a nest
//
// (Needed to make traversal of a nest tree in the reverse direction more efficient)
func (nest *Nest) GetLastChildNest() *Nest {
	return nest.lastChildNest
}

// Get next nest that belongs to the same parent nest
func (nest *Nest) GetNextSiblingNest() *Nest {
	return nest.nextSiblingNest
}

// Get previous nest that belongs to the same parent nest
func (nest *Nest) GetPrevSiblingNest() *Nest {
	return nest.prevSiblingNest
}

// Get next nest in the entire tree
func (nest *Nest) GetNextNest() *Nest {
	if next_nest := nest.firstChildNest; next_nest != nil {
		return next_nest
	}

	if next_nest := nest.nextSiblingNest; next_nest != nil {
		return next_nest
	}

	// Here we redefine the function argument
	nest = nest.parentNest

	for nest != nil && nest.nextSiblingNest == nil {
		nest = nest.parentNest
	}

	if nest != nil {
		return nest.nextSiblingNest
	} else {
		return nil
	}
}

// Get previous nest in the entire tree
func (nest *Nest) GetPrevNest() *Nest {
	if nest.prevSiblingNest == nil {
		return nest.parentNest
	}

	prev_nest := nest.prevSiblingNest

	for prev_nest.lastChildNest != nil {
		prev_nest = prev_nest.lastChildNest
	}

	return prev_nest
}

// Get first graph node belonging to a nest
func (nest *Nest) GetFirstNode() *Node {
	return nest.firstNode
}

// Get last graph node belonging to a nest
func (nest *Nest) GetLastNode() *Node {
	return nest.lastNode
}

// Get first graph edge belonging to a nest
func (nest *Nest) GetFirstEdge() *Edge {
	return nest.firstEdge
}

// Get value of a nest string attribute
func (nest *Nest) GetStrAttrVal(attr *NestStrAttr) (string, error) {
	if !attr.is_valid {
		return "", errors.New("The attribute is invalid")
	}

	if attr.nestTree != nest.nestTree {
		return "", errors.New("The attribute and the nest belong to different nest trees")
	}

	if !nest.strAttrs[attr.attr_num].isSet {
		return "", errors.New("The attribute is not set for the nest")
	}

	return nest.strAttrs[attr.attr_num].data, nil
}

// Set value of a nest string attribute
func (nest *Nest) SetStrAttrVal(attr *NestStrAttr, val string) error {
	if attr.is_valid == false {
		return errors.New("The attribute is invalid")
	}

	if attr.nestTree != nest.nestTree {
		return errors.New("The attribute and the nest belong to different nest trees")
	}

	nest.strAttrs[attr.attr_num].isSet = true
	nest.strAttrs[attr.attr_num].data = val

	return nil
}

// Remove string attribute from a specific nest
func (nest *Nest) RemoveStrAttr(attr *NestStrAttr) error {
	if !attr.is_valid {
		errors.New("The attribute is invalid")
	}

	if attr.nestTree != nest.nestTree {
		errors.New("The attribute and the nest belong to different nest trees")
	}

	nest.strAttrs[attr.attr_num].isSet = false

	return nil
}

// Check wheter a string attribute is set for a nest
func (nest *Nest) IsStrAttrSet(attr *NestStrAttr) (bool, error) {
	if !attr.is_valid {
		return false, errors.New("The attribute is invalid")
	}

	if attr.nestTree != nest.nestTree {
		return false, errors.New("The attribute and the nest belong to different nest " +
			"trees")
	}

	return nest.strAttrs[attr.attr_num].isSet, nil
}

// Add a graph node to a nest
//
// This method has an auxiliary purpose. It must be available inside the Graph package
// only and stay invisible from outside. The package clients should use "MoveToNest()"
// method of "Node" to assign a graph node to a specific nest
func (nest *Nest) addNode(node *Node) {
	panic_msg_prefix := "Panic while adding a graph node to a nest: "

	// Some sanity checks first
	if nest.nestTree == nil {
		panic(panic_msg_prefix + "the nest is not linked to any nest tree")
	}

	if nest.nestTree.baseGraph != node.graph {
		panic(panic_msg_prefix + "the nest was possibly created for a different graph")
	}

	first_node := nest.firstNode

	if first_node != nil {
		first_node.prevNodeInNest = node
	} else {
		nest.lastNode = node
	}

	node.nextNodeInNest = first_node
	nest.firstNode = node

	return
}

// Remove a graph node from a nest
//
// This method has an auxiliary purpose. It must be available inside the Graph package
// only and stay invisible from outside. From the perspective of the package clients, a
// graph node cannot be just removed from a nest. A node always belongs to some nest, but
// can be moved from one nest to another. This is just an internal implementation detail
// that when a node is moved across nests, it is first deleted from the source nest and
// then added to the destination one (this order is not significant and can change)
func (nest *Nest) removeNode(node *Node) {
	panic_msg_prefix := "Panic while removing a graph node from a nest: "

	// Some sanity checks first
	if nest.nestTree == nil {
		panic(panic_msg_prefix + "the nest is not linked to any nest tree")
	}

	if nest.nestTree.baseGraph != node.graph {
		panic(panic_msg_prefix + "the nest was possibly created for a different graph")
	}

	next_node := node.nextNodeInNest
	prev_node := node.prevNodeInNest

	if next_node != nil {
		next_node.prevNodeInNest = prev_node
	} else {
		nest.lastNode = prev_node
	}

	if prev_node != nil {
		prev_node.nextNodeInNest = next_node
	} else {
		nest.firstNode = next_node
	}

	node.nextNodeInNest = nil
	node.prevNodeInNest = nil

	return
}

// Add a graph edge to a nest
//
// This method has an auxiliary purpose. It must be available inside the Graph package
// only and stay invisible from outside. The edges are never assigned to nests explicitly,
// they just "follow" the nodes. The nodes can be added to nests explicitly; nests for
// edges are automatically calculated based on nests to which nodes adjacent to the edges
// belong
func (nest *Nest) addEdge(edge *Edge) {
	// Some sanity checks first
	panic_msg_prefix := "Panic while adding a graph edge to a nest: "

	if nest.nestTree == nil {
		panic(panic_msg_prefix + "the nest is not linked to any nest tree")
	}

	if nest.nestTree.baseGraph != edge.graph {
		panic(panic_msg_prefix + "the nest was possibly created for a different graph")
	}

	first_edge := nest.firstEdge

	if first_edge != nil {
		first_edge.prevEdgeInNest = edge
	}

	edge.nextEdgeInNest = first_edge
	nest.firstEdge = edge

	return

}

// Remove a graph edge from a nest
//
// This method has an auxiliary purpose. It must be available inside the Graph package
// only and stay invisible from outside. The edges are never get deleted from nests by the
// package clients explicitly. An edge can be deleted from the corresponding nest -
// transparently to a client - in two cases:
//     1) it is deleted from the graph
//     2) it needs to be moved to a different nest because nest attribution of (at least)
//        one of the edge's adjacent nodes has changed
// In the second case the edge will be deleted from the source nest and then added to the
// target nest, but both operations will be transparent to a Graph package client
func (nest *Nest) removeEdge(edge *Edge) {
	// Some sanity checks first
	panic_msg_prefix := "Panic while removing a graph edge from a nest: "

	if nest.nestTree == nil {
		panic(panic_msg_prefix + "the nest is not linked to any nest tree")
	}

	if nest.nestTree.baseGraph != edge.graph {
		panic(panic_msg_prefix + "the nest was possibly created for a different graph")
	}

	next_edge := edge.nextEdgeInNest
	prev_edge := edge.prevEdgeInNest

	if next_edge != nil {
		next_edge.prevEdgeInNest = prev_edge
	}

	if prev_edge != nil {
		prev_edge.nextEdgeInNest = next_edge
	} else {
		nest.firstEdge = next_edge
	}

	edge.nextEdgeInNest = nil
	edge.prevEdgeInNest = nil

	return

}

// Create a nest tree
//
// NOTE: it's expected below that all base graph fields - except "nestTree" - were
//       properly initialized before calling "newNestTree()"
func newNestTree(base_graph *Graph) *NestTree {
	// A nest tree can be created from inside the Graph package only. It's expected that
	// use of nest tree from inside the package is correct. "base_graph" cannot be "nil"
	if base_graph == nil {
		panic("Base graph is \"nil\" when creating a nest tree")
	}

	// NOTE: it's expected below that graph attribute specification was properly
	//       initialized before calling "newNestTree()"
	nt_p := &NestTree{
		baseGraph:           base_graph,
		nestCount:           0,
		rootNest:            nil,
		nestStrAttrAllocMap: make([]bool, base_graph.attrSpec.NestStrAttrNum),
	}

	root_nest_p := &Nest{
		id:              0,
		nestTree:        nt_p,
		level:           NT_ROOT_NEST_LEVEL,
		parentNest:      nil,
		firstChildNest:  nil,
		lastChildNest:   nil,
		nextSiblingNest: nil,
		prevSiblingNest: nil,
		firstNode:       nil,
		lastNode:        nil,
		firstEdge:       nil,
	}

	nt_p.nestCount++
	nt_p.rootNest = root_nest_p

	return nt_p
}

// Create a new nest in a nest tree
//
// Newly created nests have the root nest as a parent. A different parent can be assigned
// later by calling "SetParentNest()" method of the nest
func (nt *NestTree) NewNest() *Nest {
	// Some sanity checks first
	panic_msg_prefix := "Panic while creating a new nest: "

	if nt.rootNest == nil {
		panic(panic_msg_prefix + "the tree has zero reference to the root nest")
	}

	if nt.baseGraph == nil {
		panic(panic_msg_prefix + "the tree has zero reference to the base graph")
	}

	nest_p := &Nest{
		id:              nt.nestCount,
		nestTree:        nt,
		level:           nt.rootNest.level + 1,
		parentNest:      nt.rootNest,
		firstChildNest:  nil,
		lastChildNest:   nil,
		nextSiblingNest: nt.rootNest.firstChildNest,
		prevSiblingNest: nil,
		firstNode:       nil,
		lastNode:        nil,
		firstEdge:       nil,
		strAttrs:        make([]strAttrVal, nt.baseGraph.attrSpec.NestStrAttrNum),
	}

	if sibling := nt.rootNest.firstChildNest; sibling != nil {
		sibling.prevSiblingNest = nest_p
	} else {
		nt.rootNest.lastChildNest = nest_p
	}

	nt.rootNest.firstChildNest = nest_p
	nt.nestCount++

	return nest_p
}

// Allocate new nest string attribute for a nest tree
func (nt *NestTree) NewNestStrAttr() (*NestStrAttr, error) {
	// Find non-allocated attribute
	for i := 0; i < len(nt.nestStrAttrAllocMap); i++ {
		if nt.nestStrAttrAllocMap[i] == false {
			nt.nestStrAttrAllocMap[i] = true
			new_attr := NestStrAttr{i, true, nt}

			return &new_attr, nil
		}
	}

	return &nest_str_attr_invalid, errors.New("No available nest string attributes")
}

// Release nest string attribute for a nest tree
func (nt *NestTree) ReleaseNestStrAttr(attr *NestStrAttr) error {
	if !attr.is_valid {
		return errors.New("The attribute cannot be released. It's invalid")
	}

	if attr.nestTree != nt {
		return errors.New("The attribute doesn't belong to the nest tree")
	}

	attr_num := attr.attr_num

	// Remove the attribute from all existing nests
	for nest := nt.GetRootNest(); nest != nil; nest = nest.GetNextNest() {
		// Explicitly ingnore error that may be returned by the below call
		// (since no error is expected)
		nest.RemoveStrAttr(attr)
	}

	// Finally, deallocate the attribute (remove it from the attribute allocation map)
	nt.nestStrAttrAllocMap[attr_num] = false
	*attr = nest_str_attr_invalid

	return nil
}

// Get graph for which a nest tree is built
func (nt *NestTree) GetBaseGraph() *Graph {
	return nt.baseGraph
}

// Get root nest of a tree
func (nt *NestTree) GetRootNest() *Nest {
	return nt.rootNest
}

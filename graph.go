/*
  Structures for representing arbitrary graphs

  The package is NOT thread-safe. It was designed to be used in a single-threaded
  code. Additional efforts are required to make the code below thread-safe

  NOTE: whole-graph attributes are implemented in the same way as node and edge
        attributes. They could be implemented differently because for graph attributes
		one doesn't need to keep attribute allocation map separately. Since each graph
		attribute is singular (in contrast to node attributes, for example, where
		it's required to be able to store the attribute value for each node), it's
		possible to make allocation information a part of an attribute itself. Both
		options - the used one and the alternative described here - seem to not have
		significant benefits one over the other. But the implemented option has a
		benefit of unification with node and edge attributes
  NOTE: a separate iterator for all graph edges is NOT provided. Since every edge is
  		connected to two nodes, it's possible to iterate over all graph edges by
		iterating over all graph nodes and for each node iterating over all its
		successors (or predecessors). The iterator for all graph nodes IS provided
*/

package graph

import (
	"errors"
)

/**
 * Generic graph interfaces, structures and functions
 */

// String attribute value representation
type strAttrVal struct {
	isSet bool
	data  string
}

// Type representing string attribute of graph nodes, edges and graph as a whole
type graphStrAttr struct {
	// Number of an attribute in an array of string attributes
	attrNum int
	// Whether the attribute is valid
	isValid bool
	// Reference to a parent graph
	graph *Graph
}

// String attribute of graph as a whole
type GraphStrAttr graphStrAttr

// String attribute of graph node
type NodeStrAttr graphStrAttr

// Representation of the invalid graph string attribute
var graph_str_attr_invalid = GraphStrAttr{-1, false, nil}

// Representation of the invalid node string attribute
var node_str_attr_invalid = NodeStrAttr{-1, false, nil}

// Type describing which and how many attributes a graph should have
// A variable of this type must be provided when creating a new graph
type AttrSpec struct {
	// Number of string attributes a graph can have
	GraphStrAttrNum int
	// Number of string attributes a node can have
	NodeStrAttrNum int
	// Number of string attributes a nest can have
	NestStrAttrNum int
}

// Return default attribute specification
func DefaultAttrSpec() AttrSpec {
	return AttrSpec{
		NodeStrAttrNum: 0,
	}
}

/**
 * End: Generic graph interfaces and structures
 */

/**
 * Implementation of Graph
 */

// Graph node representation
type Node struct {
	// First incoming edge
	firstIncomingEdge *Edge
	// First outcoming edge
	firstOutcomingEdge *Edge
	// Node ID
	id int
	// Nest to which a node belongs
	nest *Nest
	// Next node belonging to the same nest
	nextNodeInNest *Node
	// Previous node belonging to the same nest
	prevNodeInNest *Node
	// Reference to the parent graph
	graph *Graph
	// Array of string attributes
	strAttrs []strAttrVal
}

// Graph edge representation
type Edge struct {
	// Edge ID
	id int
	// Reference to the parent graph
	graph *Graph
	// Source node of edge
	srcNode *Node
	// Destination node of edge
	dstNode *Node
	// Next outcoming edge
	nextOutcomingEdge *Edge
	// Previous outcoming edge
	prevOutcomingEdge *Edge
	// Next incoming edge
	nextIncomingEdge *Edge
	// Previous incoming edge
	prevIncomingEdge *Edge
	// Next edge in a list of edges belonging to the same nest
	nextEdgeInNest *Edge
	// Previous edge in a list of edges belonging to the same nest
	prevEdgeInNest *Edge
	// Nest to which an edge belongs
	nest *Nest
}

// Graph representation
type Graph struct {
	// Nest tree of a graph
	nestTree *NestTree
	// Node count (must always increase and never decrease. Even if nodes get deleted).
	// Creation of a new node increments the counter
	nodeCount int
	// Edge count (must always increase and never decrease. Even if edges get deleted).
	// Creation of a new edge increments the counter
	edgeCount int
	// Specification of graph's attributes
	attrSpec AttrSpec
	// Allocation map for graph string attributes
	// An element holds TRUE if corresponding attribute is allocated and FALSE
	// in the opposite case
	graphStrAttrAllocMap []bool
	// Allocation map for node string attributes
	// An element holds TRUE if corresponding attribute is allocated and FALSE
	// in the opposite case
	nodeStrAttrAllocMap []bool
	// Array of graph string attributes
	strAttrs []strAttrVal
}

// Create new Graph
func NewGraph(attr_spec AttrSpec) *Graph {
	graph_p := &Graph{
		nestTree:             nil,
		nodeCount:            0,
		edgeCount:            0,
		attrSpec:             attr_spec,
		graphStrAttrAllocMap: make([]bool, attr_spec.GraphStrAttrNum),
		nodeStrAttrAllocMap:  make([]bool, attr_spec.NodeStrAttrNum),
		strAttrs:             make([]strAttrVal, attr_spec.GraphStrAttrNum),
	}

	graph_p.nestTree = newNestTree(graph_p)

	return graph_p
}

// Get nest tree of a graph
func (graph *Graph) GetNestTree() *NestTree {
	return graph.nestTree
}

// Get first Graph node in a list of all nodes
func (graph *Graph) GetFirstNode() *Node {
	// Some sanity checks first
	if graph.nestTree == nil || graph.nestTree.rootNest == nil {
		panic("Graph without a nest tree or the nest tree without the root nest")
	}

	nest := graph.nestTree.rootNest

	for nest != nil && nest.firstNode == nil {
		nest = nest.GetNextNest()
	}

	if nest == nil {
		return nil
	}

	return nest.firstNode
}

// Allocate new Graph string attribute
func (graph *Graph) NewGraphStrAttr() (*GraphStrAttr, error) {
	// Find non-allocated attribute
	for i := 0; i < len(graph.graphStrAttrAllocMap); i++ {
		if graph.graphStrAttrAllocMap[i] == false {
			graph.graphStrAttrAllocMap[i] = true
			new_attr := GraphStrAttr{i, true, graph}

			return &new_attr, nil
		}
	}

	return &graph_str_attr_invalid, errors.New("No available graph string attributes")
}

// Remove string attribute from a Graph
func (graph *Graph) RemoveStrAttr(attr *GraphStrAttr) error {
	if !attr.isValid {
		errors.New("The attribute is invalid")
	}

	if attr.graph != graph {
		errors.New("The attribute doesn't belong to the graph")
	}

	graph.strAttrs[attr.attrNum].isSet = false

	return nil
}

// Check wheter a string attribute is set for a Graph
func (graph *Graph) IsStrAttrSet(attr *GraphStrAttr) (bool, error) {
	if !attr.isValid {
		return false, errors.New("The attribute is invalid")
	}

	if attr.graph != graph {
		return false, errors.New("The attribute doesn't belong to the graph")
	}

	return graph.strAttrs[attr.attrNum].isSet, nil
}

// Set value of a Graph string attribute
func (graph *Graph) SetStrAttrVal(attr *GraphStrAttr, val string) error {
	if attr.isValid == false {
		return errors.New("The attribute is invalid")
	}

	if attr.graph != graph {
		return errors.New("The attribute doesn't belong to the graph")
	}

	graph.strAttrs[attr.attrNum].isSet = true
	graph.strAttrs[attr.attrNum].data = val

	return nil
}

// Get value of a Graph string attribute
func (graph *Graph) GetStrAttrVal(attr *GraphStrAttr) (string, error) {
	if !attr.isValid {
		return "", errors.New("The attribute is invalid")
	}

	if attr.graph != graph {
		return "", errors.New("The attribute doesn't belong to the graph")
	}

	if !graph.strAttrs[attr.attrNum].isSet {
		return "", errors.New("The attribute is not set for the graph")
	}

	return graph.strAttrs[attr.attrNum].data, nil
}

// Release Graph string attribute
func (graph *Graph) ReleaseGraphStrAttr(attr *GraphStrAttr) error {
	if !attr.isValid {
		return errors.New("The attribute cannot be released. It's invalid")
	}

	if attr.graph != graph {
		return errors.New("The attribute doesn't belong to the graph")
	}

	attr_num := attr.attrNum

	// Remove the attribute from the graph
	graph.RemoveStrAttr(attr)

	// Finally, deallocate the attribute (remove it from the attribute allocation map)
	graph.graphStrAttrAllocMap[attr_num] = false
	*attr = graph_str_attr_invalid

	return nil
}

// Allocate new node string attribute for a Graph
func (graph *Graph) NewNodeStrAttr() (*NodeStrAttr, error) {
	// Find non-allocated attribute
	for i := 0; i < len(graph.nodeStrAttrAllocMap); i++ {
		if graph.nodeStrAttrAllocMap[i] == false {
			graph.nodeStrAttrAllocMap[i] = true
			new_attr := NodeStrAttr{i, true, graph}

			return &new_attr, nil
		}
	}

	return &node_str_attr_invalid, errors.New("No available node string attributes")
}

// Release node string attribute for a Graph
func (graph *Graph) ReleaseNodeStrAttr(attr *NodeStrAttr) error {
	if !attr.isValid {
		return errors.New("The attribute cannot be released. It's invalid")
	}

	if attr.graph != graph {
		return errors.New("The attribute doesn't belong to the graph")
	}

	attr_num := attr.attrNum

	// Remove the attribute from all existing nodes
	for node := graph.GetFirstNode(); node != nil; node = node.GetNextNode() {
		// Explicitly ingnore error that may be returned by the below call
		// (since no error is expected)
		node.RemoveStrAttr(attr)
	}

	// Finally, deallocate the attribute (remove it from the attribute allocation map)
	graph.nodeStrAttrAllocMap[attr_num] = false
	*attr = node_str_attr_invalid

	return nil
}

// Create new Graph node
//
// A newly created Graph node is assigned to the root nest. Later it can be assigned to
// a different nest by calling "SetNest()" method of the node
func (graph *Graph) NewNode() *Node {
	// Some sanity checks first
	if graph.nestTree == nil || graph.nestTree.rootNest == nil {
		panic("Graph without a nest tree or the nest tree without the root nest")
	}

	node_p := &Node{
		firstIncomingEdge:  nil,
		firstOutcomingEdge: nil,
		id:                 graph.nodeCount,
		nest:               graph.nestTree.rootNest,
		nextNodeInNest:     nil,
		prevNodeInNest:     nil,
		graph:              graph,
		strAttrs:           make([]strAttrVal, graph.attrSpec.NodeStrAttrNum),
	}

	graph.nestTree.rootNest.addNode(node_p)
	graph.nodeCount++

	return node_p
}

// Create new edge between pre-existing nodes in a Graph
//
// Multiple edges in the same direction between two given nodes ARE allowed
func (graph *Graph) NewEdge(src_node *Node, dst_node *Node) (*Edge, error) {

	if src_node == nil {
		return nil, errors.New("Pointer to the source node cannot be \"nil\"")
	}

	if dst_node == nil {
		return nil, errors.New("Pointer to the destination node cannot be \"nil\"")
	}

	if src_node.graph != graph {
		return nil, errors.New("Source node doesn't belong to the graph for " +
			"which the method is called")
	}

	if dst_node.graph != graph {
		return nil, errors.New("Destination node doesn't belong to the graph " +
			"for which the method is called")
	}

	src_first_out_edge := src_node.firstOutcomingEdge
	dst_first_in_edge := dst_node.firstIncomingEdge
	edge_p := &Edge{
		id:                graph.edgeCount,
		srcNode:           src_node,
		dstNode:           dst_node,
		nextOutcomingEdge: src_first_out_edge,
		prevOutcomingEdge: nil,
		nextIncomingEdge:  dst_first_in_edge,
		prevIncomingEdge:  nil,
		graph:             graph,
	}

	if src_first_out_edge != nil {
		src_first_out_edge.prevOutcomingEdge = edge_p
	}

	src_node.firstOutcomingEdge = edge_p

	if dst_first_in_edge != nil {
		dst_first_in_edge.prevIncomingEdge = edge_p
	}

	dst_node.firstIncomingEdge = edge_p
	edge_p.calcNestAndMoveToIt()
	graph.edgeCount++

	return edge_p, nil
}

// Get attribute specification of a Graph
func (graph *Graph) GetAttrSpec() AttrSpec {
	return graph.attrSpec
}

// Get node ID
func (node *Node) GetID() int {
	return node.id
}

// Get nest to which a node belongs
func (node *Node) GetNest() *Nest {
	return node.nest
}

// Get next node belonging to the same nest
func (node *Node) GetNextNodeInNest() *Node {
	return node.nextNodeInNest
}

// Get previous node belonging to the same nest
func (node *Node) GetPrevNodeInNest() *Node {
	return node.prevNodeInNest
}

// Get graph to which a node belongs
func (node *Node) GetGraph() *Graph {
	return node.graph
}

// Get next node in a list of all Graph nodes
func (node *Node) GetNextNode() *Node {
	if node.nextNodeInNest != nil {
		return node.nextNodeInNest
	}

	if node.nest == nil {
		panic("Graph node not assigned to any nest")
	}

	nest := node.nest.GetNextNest()

	for nest != nil && nest.firstNode == nil {
		nest = nest.GetNextNest()
	}

	if nest == nil {
		return nil
	}

	return nest.firstNode
}

// Get previous node in a list of all Graph nodes
func (node *Node) GetPrevNode() *Node {
	if node.prevNodeInNest != nil {
		return node.prevNodeInNest
	}

	if node.nest == nil {
		panic("Graph node not assigned to any nest")
	}

	nest := node.nest.GetPrevNest()

	for nest != nil && nest.lastNode == nil {
		nest = nest.GetPrevNest()
	}

	if nest == nil {
		return nil
	}

	return nest.lastNode
}

// Get first outcoming edge of a Basic Node
func (node *Node) GetFirstOutcomingEdge() *Edge {
	return node.firstOutcomingEdge
}

// Get first incoming edge of a Basic Node
func (node *Node) GetFirstIncomingEdge() *Edge {
	return node.firstIncomingEdge
}

// Set value of a Basic Node string attribute
func (node *Node) SetStrAttrVal(attr *NodeStrAttr, val string) error {
	if attr.isValid == false {
		return errors.New("The attribute is invalid")
	}

	if attr.graph != node.graph {
		return errors.New("The attribute and the node belong to different graphs")
	}

	node.strAttrs[attr.attrNum].isSet = true
	node.strAttrs[attr.attrNum].data = val

	return nil
}

// Get value of a Basic Node string attribute
func (node *Node) GetStrAttrVal(attr *NodeStrAttr) (string, error) {
	if !attr.isValid {
		return "", errors.New("The attribute is invalid")
	}

	if attr.graph != node.graph {
		return "", errors.New("The attribute and the node belong to different graphs")
	}

	if !node.strAttrs[attr.attrNum].isSet {
		return "", errors.New("The attribute is not set for the node")
	}

	return node.strAttrs[attr.attrNum].data, nil
}

// Remove string attribute from a specific Basic Node
func (node *Node) RemoveStrAttr(attr *NodeStrAttr) error {
	if !attr.isValid {
		errors.New("The attribute is invalid")
	}

	if attr.graph != node.graph {
		errors.New("The attribute and the node belong to different graphs")
	}

	node.strAttrs[attr.attrNum].isSet = false

	return nil
}

// Check wheter a string attribute is set for a Basic Node
func (node *Node) IsStrAttrSet(attr *NodeStrAttr) (bool, error) {
	if !attr.isValid {
		return false, errors.New("The attribute is invalid")
	}

	if attr.graph != node.graph {
		return false, errors.New("The attribute and the node belong to different graphs")
	}

	return node.strAttrs[attr.attrNum].isSet, nil
}

// Move graph node to a specific nest
//
// Nests get automatically recalculated for edges incoming to and outcoming from the node
func (node *Node) MoveToNest(nest *Nest) error {
	panic_msg_prefix := "Panic while moving a graph node to a different nest: "

	if nest.nestTree.baseGraph != node.graph {
		return errors.New("Attempt to move a graph node to a nest that belongs to a " +
			"different graph")
	}

	if node.nest == nil {
		panic(panic_msg_prefix + "the node is not assigned to any \"source\" nest")
	}

	node.nest.removeNode(node)
	node.nest = nest
	nest.addNode(node)

	// Fix nest attribution for edges incoming to the node
	for edge := node.GetFirstIncomingEdge(); edge != nil; edge.GetNextIncomingEdge() {
		if edge.nest == nil {
			panic(panic_msg_prefix + "the node has in incoming edge that is not " +
				"assigned to any nest")
		}

		edge.calcNestAndMoveToIt()
	}

	// Fix nest attribution for edges outcoming from the node
	for edge := node.GetFirstOutcomingEdge(); edge != nil; edge.GetNextOutcomingEdge() {
		if edge.nest == nil {
			panic(panic_msg_prefix + "the node has in outcoming edge that is not " +
				"assigned to any nest")
		}

		edge.calcNestAndMoveToIt()
	}

	return nil
}

// Get Basic Edge ID
func (edge *Edge) GetID() int {
	return edge.id
}

// Get graph to which an edge belongs
func (edge *Edge) GetGraph() *Graph {
	return edge.graph
}

// Get source node of Basic Edge
func (edge *Edge) GetSrcNode() *Node {
	return edge.srcNode
}

// Get destination node of Basic Edge
func (edge *Edge) GetDstNode() *Node {
	return edge.dstNode
}

// Get next outcoming Basic Edge
func (edge *Edge) GetNextOutcomingEdge() *Edge {
	return edge.nextOutcomingEdge
}

// Get previous outcoming Basic Edge
func (edge *Edge) GetPrevOutcomingEdge() *Edge {
	return edge.prevOutcomingEdge
}

// Get next incoming Basic Edge
func (edge *Edge) GetNextIncomingEdge() *Edge {
	return edge.nextIncomingEdge
}

// Get previous incoming Basic Edge
func (edge *Edge) GetPrevIncomingEdge() *Edge {
	return edge.prevIncomingEdge
}

// Get next edge belonging to the same nest
func (edge *Edge) GetNextEdgeInNest() *Edge {
	return edge.nextEdgeInNest
}

// Get previous edge belonging to the same nest
func (edge *Edge) GetPrevEdgeInNest() *Edge {
	return edge.prevEdgeInNest
}

// Calculate nest to which an edge should belong. Add the edge to this nest
//
// This method must not be visible outside the Graph package. Only graph nodes can be
// added to any nest explicitly. Edge attribution to some nest is calculated automatically
// each time an edge is created or some node gets moved to a specific nest
func (edge *Edge) calcNestAndMoveToIt() {
	panic_msg_prefix := "Panic while calculating nest for an edge"
	src_node := edge.srcNode
	dst_node := edge.dstNode

	// Some sanity checks. Since the method is not visible outside the package, it must be
	// called on consistent data sets. If an error is detected by the below checks, it
	// will not be returned by the method. Such an error might occur only due to incorrect
	// implementation of the Graph package itself, which must be very unlikely :)
	if src_node == nil || dst_node == nil {
		panic(panic_msg_prefix + " which is not adjecent to a node at least at one end")
	}

	if src_node.graph != dst_node.graph {
		panic(panic_msg_prefix + " that connects nodes belonging to different graphs")
	}

	if edge.graph != src_node.graph {
		panic(panic_msg_prefix + " that connects nodes belonging to a different graph " +
			"than the edge itself")
	}

	if src_node.nest == nil || dst_node.nest == nil {
		panic(panic_msg_prefix + " for which at least one adjacent node is not " +
			"assigned to any nest")
	}

	src_nest := src_node.nest
	dst_nest := dst_node.nest

	if src_nest.nestTree != dst_nest.nestTree {
		panic(panic_msg_prefix + " that connects nodes assigned to nests from " +
			"different nest trees")
	}

	if src_nest.nestTree.baseGraph != edge.graph {
		panic("Nest tree to which an edge is to be assigned relates to a different " +
			"graph than the edge itself")
	}

	// Find the closest nest that contains both the source and destination nest. The edge
	// will be added to the found nest
	panic_msg_inconsistent_nt := ": either the nest tree has disconnected components " +
		"or the nests have inconsistent levels"

	// If the source nest is "deeper" than the destination nest in the nest hierarchy,
	// then find a nest which contains the source nest but belongs to the same
	// level in the nest hierarchy as the destination nest
	for src_nest.level > dst_nest.level {
		src_nest = src_nest.parentNest

		if src_nest == nil {
			panic(panic_msg_prefix + panic_msg_inconsistent_nt)
		}
	}

	// If the destination nest is "deeper" than the source nest in the nest hierarchy,
	// then find a nest which contains the destination nest but belongs to the same level
	// in the nest hierarchy as the source nest
	for dst_nest.level > src_nest.level {
		dst_nest = dst_nest.parentNest

		if dst_nest == nil {
			panic(panic_msg_prefix + panic_msg_inconsistent_nt)
		}
	}

	for dst_nest != src_nest {
		dst_nest = dst_nest.parentNest
		src_nest = src_nest.parentNest

		if dst_nest == nil || src_nest == nil {
			panic(panic_msg_prefix + panic_msg_inconsistent_nt)
		}
	}

	// Newly created edges may not be assigned to any nest
	if edge.nest != nil {
		edge.nest.removeEdge(edge)
	}

	edge.nest = src_nest
	src_nest.addEdge(edge)

	return
}

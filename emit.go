/*
  Emit graph in various formats
*/

package graph

import (
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"os"
	"strings"
)

const EMIT_WRITE_ERR_MSG_PREFIX = "Error writing to the output file: "

const EMIT_INDENT = "  "

// Variables of the below type map printable properties of a graph as a whole to actual
// graph attributes. For example, if the "label" field is not "nil" - i.e. equal to a
// pointer to some graph string attribute - then it means that "label" property is
// represented by this attribute
type GlobalEmitSpec struct {
	// Label of a graph
	LabelAttr *GraphStrAttr
}

// Variables of the below type map printable node properties to actual node attributes.
// For example, if the "label" field is not "nil" - i.e. equal to a pointer to some node
// string attribute - then it means that "label" property is represented by this attribute
type NodeEmitSpec struct {
	LabelAttr *NodeStrAttr
}

// Variables of the below type map printable properties of a graph and its elements into
// actual attributes of the graph, its nodes, its edges
type GraphEmitSpec struct {
	// Printable properties of a graph as a whole
	Graph GlobalEmitSpec
	// Per-node printable properties mapped into node attributes
	Node NodeEmitSpec
	// Per-nest printable properties mapped into nest attributes
	Nest NestEmitSpec
}

// Emit nodes and edges of a nest in Graphviz format
func emitGVSubgraphNodesAndEdges(nest *Nest,
	graph_emit_spec *GraphEmitSpec,
	out_file *os.File,
	indent string) error {

	if nest.GetNestTree() == nil {
		return errors.New("The nest is not linked to any nest tree")
	}

	graph := nest.GetNestTree().GetBaseGraph()

	if graph == nil {
		return errors.New("The nest tree to which the nest belongs is not linked to " +
			"any graph")
	}

	// Emit graph nodes belonging to the nest
	for node := nest.GetFirstNode(); node != nil; node = node.GetNextNodeInNest() {
		var node_label string

		node_label_attr := graph_emit_spec.Node.LabelAttr
		node_desc_line := fmt.Sprintf(indent+"%d", node.GetID())

		if node_label_attr != nil {
			if is_set, err := node.IsStrAttrSet(node_label_attr); err != nil {
				err_msg := fmt.Sprintf("Error checking whether node label attribute is "+
					"set [node ID = %d]: ", node.GetID())

				return errors.New(err_msg + err.Error())
			} else if is_set {
				node_label, err = node.GetStrAttrVal(node_label_attr)

				if err != nil {
					err_msg := fmt.Sprintf("Error retrieving node label attribute "+
						"[node ID = %d]: ", node.GetID())

					return errors.New(err_msg + err.Error())
				}

				node_desc_line += " [label=\"" + node_label + "\"]"
			}
		}

		node_desc_line += ";\n"

		if _, err := out_file.WriteString(node_desc_line); err != nil {
			return errors.New(EMIT_WRITE_ERR_MSG_PREFIX + err.Error())
		}
	}

	// Emit graph edges belonging to the nest
	for edge := nest.GetFirstEdge(); edge != nil; edge = edge.GetNextEdgeInNest() {
		if edge.GetSrcNode() == nil || edge.GetDstNode() == nil {
			return errors.New("At least one end of an edge belonging to the nest is " +
				"not connected to any graph node")
		}

		src_node := edge.GetSrcNode()
		dst_node := edge.GetDstNode()

		if edge.GetGraph() != graph {
			return errors.New("An edge belonging to the nest is attributed to a " +
				"different graph than the nest itself")
		}

		if src_node.GetGraph() != graph || dst_node.GetGraph() != graph {
			return errors.New("At least one of the nodes connected by an edge " +
				"belonging to the nest is attributed to a different graph (than the " +
				"edge itself)")
		}

		edge_desc_line := fmt.Sprintf(indent+"%d -> %d;\n", src_node.GetID(),
			dst_node.GetID())

		if _, err := out_file.WriteString(edge_desc_line); err != nil {
			return errors.New(EMIT_WRITE_ERR_MSG_PREFIX + err.Error())
		}
	}

	return nil
}

// Emit a nested sub-graph in Graphviz format
func emitGVSubgraph(nest *Nest,
	graph_emit_spec *GraphEmitSpec,
	out_file *os.File,
	indent string) error {

	panic_msg_prefix := "Panic while emitting a nest in Graphviz format: "

	if graph_emit_spec == nil {
		panic(panic_msg_prefix + "zero reference to graph emit specification")
	}

	if out_file == nil {
		panic(panic_msg_prefix + "zero reference to output file")
	}

	// Emit subgraph opening clause
	nest_id_as_str := fmt.Sprintf("%d", nest.GetID())
	_, err := out_file.WriteString(indent + "subgraph cluster_" + nest_id_as_str + " {\n")

	if err != nil {
		return errors.New(EMIT_WRITE_ERR_MSG_PREFIX + err.Error())
	}

	// Emit subgraph label (if exists)
	nest_label_attr := graph_emit_spec.Nest.LabelAttr

	if nest_label_attr != nil {
		is_set, err := nest.IsStrAttrSet(nest_label_attr)

		if err != nil {
			return errors.New("Error while checking whether a value of the nest string " +
				"attribute is set: " + err.Error())
		}

		if is_set {
			nest_label, _ := nest.GetStrAttrVal(nest_label_attr)
			_, err := out_file.WriteString(indent + EMIT_INDENT + "label=\"" +
				nest_label + "\";\n")

			if err != nil {
				return errors.New(EMIT_WRITE_ERR_MSG_PREFIX + err.Error())
			}
		}
	}

	if err != nil {
		return errors.New(EMIT_WRITE_ERR_MSG_PREFIX + err.Error())
	}

	// Emit nested subgraphs. Nodes and edges of the current nest will be emitted after
	// that
	child_nest := nest.GetFirstChildNest()

	for ; child_nest != nil; child_nest = child_nest.GetNextSiblingNest() {
		if nest.GetNestTree() != child_nest.GetNestTree() {
			return errors.New("A child nest belongs to a different nest tree or is not " +
				"linked to any nest tree at all")
		}

		err = emitGVSubgraph(child_nest, graph_emit_spec, out_file, indent+EMIT_INDENT)

		// Because of the recursive call in this loop, the prefix of the below error
		// message may be repeated multiple times. It's considered ok for now. Because
		// later, for example, an ID of each intermediate nest could be added to the
		// message (hence, the chain of the exact nests would be reported)
		if err != nil {
			return errors.New("Couldn't emit a child nest: " + err.Error())
		}
	}

	err = emitGVSubgraphNodesAndEdges(nest, graph_emit_spec, out_file, indent+EMIT_INDENT)

	if err != nil {
		return errors.New("Couldn't emit nodes and edges belonging to a nest: " +
			err.Error())
	}

	// Emit sub-graph closing bracket
	if _, err := out_file.WriteString(indent + "}\n"); err != nil {
		return errors.New(EMIT_WRITE_ERR_MSG_PREFIX + err.Error())
	}

	return nil
}

// Print text description of a Graph in Graphviz DOT language.
// The description can be further compiled by Graphviz "dot" tool
// into Postscript file, PNG image, etc. For example, the following
// command will produce a PNG drawing of the graph (assuming the text
// description of the graph is stored in "graph.gv"):
//    dot -Tpng graph.gv -o graph.png
//
// Input: full path to the output file (all parent directories should
//        exist; the file itself must NOT exist)
func EmitInGVFormat(graph *Graph, graph_emit_spec *GraphEmitSpec, out_path string) error {
	out_file, err := os.OpenFile(out_path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)

	if err != nil {
		return errors.New("Cannot create output file: " + err.Error())
	}

	defer out_file.Close()

	// If no emit specification is provided, we create the default one. We do that to
	// simplify the code, so that we don't need to check whether graph_emit_spec is
	// "nil" every time we're going to use it
	// NOTE: here the function parameter "graph_emit_spec" is intentionally re-assigned
	if graph_emit_spec == nil {
		graph_emit_spec = &GraphEmitSpec{}
	}

	EMIT_WRITE_ERR_MSG_PREFIX := "Cannot write to the output file: "

	// Get graph label (if any). It will be used as a header and as a label
	var has_graph_label bool
	var graph_label string

	graph_label_attr := graph_emit_spec.Graph.LabelAttr

	if graph_label_attr != nil {
		if is_set, err := graph.IsStrAttrSet(graph_label_attr); err != nil {
			return errors.New("Error checking whether graph label attribute is set")
		} else if is_set {
			has_graph_label = true

			if graph_label, err = graph.GetStrAttrVal(graph_label_attr); err != nil {
				return errors.New("Error getting value of an attribute that keeps the " +
					"graph label")
			}
		}
	}

	// Emit Graph header
	graph_name := "NO NAME"

	if has_graph_label {
		graph_name = graph_label
	}

	_, err = out_file.WriteString("digraph \"" + graph_name + "\" {\n")

	if err != nil {
		return errors.New(EMIT_WRITE_ERR_MSG_PREFIX + err.Error())
	}

	// Emit Graph global properties
	// Drawing orientation property: left to right
	if _, err := out_file.WriteString("\trankdir = LR\n"); err != nil {
		return errors.New(EMIT_WRITE_ERR_MSG_PREFIX + err.Error())
	}

	// Graph label propery
	if has_graph_label {
		_, err = out_file.WriteString("\tlabel = \"" + graph_label + "\"\n")
	}

	if err != nil {
		return errors.New(EMIT_WRITE_ERR_MSG_PREFIX + err.Error())
	}

	// Emit nested subgraphs. Nodes and edges of the root nest will be emitted after that
	if graph.GetNestTree() == nil {
		return errors.New("The graph doesn't have a nest tree")
	}

	root_nest := graph.GetNestTree().GetRootNest()

	if root_nest == nil {
		return errors.New("The graph doesn't have a root nest")
	}

	child_nest := root_nest.GetFirstChildNest()

	for ; child_nest != nil; child_nest = child_nest.GetNextSiblingNest() {
		if root_nest.GetNestTree() != child_nest.GetNestTree() {
			return errors.New("A child nest belongs to a different nest tree or is not " +
				"linked to any nest tree at all")
		}

		err = emitGVSubgraph(child_nest, graph_emit_spec, out_file, EMIT_INDENT)

		// Because of the recursive call in this loop, the prefix of the below error
		// message may be repeated multiple times. It's considered ok for now. Because
		// later, for example, an ID of each intermediate nest could be added to the
		// message (hence, the chain of the exact nests would be reported)
		if err != nil {
			return errors.New("Couldn't emit a child nest: " + err.Error())
		}
	}

	// Emit Graph nodes
	// Set shape for all the nodes
	if _, err := out_file.WriteString("\tnode [shape=box];\n"); err != nil {
		return errors.New(EMIT_WRITE_ERR_MSG_PREFIX + err.Error())
	}

	err = emitGVSubgraphNodesAndEdges(root_nest, graph_emit_spec, out_file, EMIT_INDENT)

	if err != nil {
		return errors.New("Couldn't emit nodes and edges belonging to the root nest: " +
			err.Error())
	}

	// Emit Graph description closing bracket
	if _, err := out_file.WriteString("}"); err != nil {
		return errors.New(EMIT_WRITE_ERR_MSG_PREFIX + err.Error())
	}

	return nil
}

// GraphML extension families
const (
	gML_EXT_FAMILY_STANDARD = iota
	gML_EXT_FAMILY_YFILES   = iota
)

// yFiles GraphML attributes
const (
	// Node attribute "nodegraphics"
	yFILES_NATTR_NODEGRAPHICS = iota
	yFILES_ATTR_NUM           = iota
)

// Enumeration of yFiles attribute types (native types as well as extension types)
const (
	yFILES_ATTR_TYPE_NODEGRAPHICS = iota
	yFILES_ATTR_TYPE_NUM          = iota
)

// Graph element types supported by yFiles
const (
	yFILES_ELEM_NODE = iota
	yFILES_ELEM_NUM  = iota
)

type gMLAttr struct {
	// Unique attribute identifier
	id int
	// Extension family to which an attribute belongs. All standard GraphML attributes
	// belong to gml_EXT_FAMILY_STANDARD
	extFamily int
	// Type of an attribute. For example, it can be "string" or "double" which are
	// GraphML attribute types. But also it can any other type coming from some GraphML
	// extension (like "nodegraphics" which is a part of yFiles GraphML extension)
	attrType int
	// Type of graph element that an attribute is applicable to. NOTE: native GraphML
	// allows defining a single attribute that is applicable to several element types
	// simultaneously. This capability is not supported by this package. Instead of
	// defining an attribute that would be applicable to "all" graph elements, one would
	// need to define a separate attribute for each graph element type
	elemType int
}

var yFilesGMLAttrs = []gMLAttr{
	{yFILES_NATTR_NODEGRAPHICS, gML_EXT_FAMILY_YFILES, yFILES_ATTR_TYPE_NODEGRAPHICS,
		yFILES_ELEM_NODE},
}

func checkYFilesAttrArrayConsistency() error {
	if len(yFilesGMLAttrs) != yFILES_ATTR_NUM {
		return errors.New("The number of defined yFiles GraphML attributes differs " +
			"from the length of an array describing those attributes")
	}

	for i := 0; i < yFILES_ATTR_NUM; i++ {
		if yFilesGMLAttrs[i].id != i {
			err_str := fmt.Sprintf("%d-th element of an array describing yFiles "+
				"GraphML attributes has an unexpected id", i)

			return errors.New(err_str)
		}
	}

	return nil
}

// For a given yFiles attribute return an ID that should be used to identify the attribute
// inside a yFiles GraphML document
func getYFilesAttrDocumentId(attr_id int) int {
	if attr_id >= yFILES_ATTR_NUM || attr_id < 0 {
		panic("Panic while obtaining a document ID for a yFiles GraphML attribute: " +
			"the provided logical attribute ID is out of range")
	}

	// Return just the logical attribute ID itself. Later this function could be optimized
	// for cases when not all the available yFiles attributes will be used in a document.
	// For example, if only attributes with logical IDs 0, 5 and 7 will be used in a
	// document, then it would still be acceptable to assign document IDs 0, 5 and 7 to
	// those attributes. But there would be "gaps" in the document attribute IDs in this
	// case, because, for example, there will be no attribute with the document ID 1. So,
	// to avoid such gaps, it might be reasonable to assign document IDs 0, 1, 2 to
	// logical attributes 0, 5 and 7
	return attr_id
}

// For a given attribute type return a string representation of that type (i.e. a string
// that can be used in a yFiles GraphML document)
func getYFilesAttrDocumentType(attr_type int) string {
	panic_msg_prefix := "Panic while obtaining a document type for a yFiles GraphML " +
		"attribute: "

	var document_type string

	switch attr_type {
	case yFILES_NATTR_NODEGRAPHICS:
		document_type = "nodegraphics"
	default:
		panic(panic_msg_prefix + "the provided logical attribute type is unexpected " +
			"for yFiles documents")
	}

	return document_type
}

// For a given type of yFiles graph element return a printable name of that element as it
// is used in yFiles GraphML documents
func getYFilesAttrDocumentElem(elem_type int) string {
	panic_msg_prefix := "Panic while obtaining a printable name of a yFiles graph " +
		"element: "

	var document_elem string

	switch elem_type {
	case yFILES_ELEM_NODE:
		document_elem = "node"
	default:
		panic(panic_msg_prefix + "the provided logical element type is unexpected " +
			"for yFiles documents")
	}

	return document_elem
}

func emitYFilesAttrDecls(out_file *os.File, indent string) error {
	// Emit all known attribute declarations. Later this function can be optimized to emit
	// only those attributes that will actually be used
	for i := 0; i < len(yFilesGMLAttrs); i++ {
		attr := yFilesGMLAttrs[i]
		attr_document_id := getYFilesAttrDocumentId(attr.id)

		var attr_type_family string

		switch attr.extFamily {
		case gML_EXT_FAMILY_YFILES:
			attr_type_family = "yfiles.type"
		default:
			panic("Panic while emitting yFiles GraphML attribute to a document: the " +
				"attribute belongs to extension family that is not expected in yFiles " +
				"documents")
		}

		attr_document_type := getYFilesAttrDocumentType(attr.attrType)
		attr_document_elem := getYFilesAttrDocumentElem(attr.elemType)
		str_to_emit := fmt.Sprintf(indent+"<key id=\"d%d\" %s=\"%s\" for=\"%s\"/>\n",
			attr_document_id, attr_type_family, attr_document_type, attr_document_elem)

		if _, err := out_file.WriteString(str_to_emit); err != nil {
			return errors.New(EMIT_WRITE_ERR_MSG_PREFIX + err.Error())
		}
	}

	return nil
}

// Emit yFiles group node and all the graph elements that are transitively contained
// inside this group node
func emitYFilesGroup(nest *Nest,
	graph_emit_spec *GraphEmitSpec,
	out_file *os.File,
	id_prefix *string,
	indent string) error {

	panic_msg_str := "Panic while emitting an yFiles group node: "

	if nest == nil {
		panic(panic_msg_str + "zero reference to a nest representing the group")
	}

	// Emit group node open tag
	node_id := fmt.Sprintf("%snest%d", *id_prefix, nest.GetID())
	// Presence of "yfiles.foldertype" attribute means that this is a group node (i.e. it
	// has some other nodes inside. The value "folder" of this attribute means that the
	// node must be drawn in a folded state (i.e. a user will need to "unfold" the node to
	// see graph elements contained inside it). To draw a node in an unfolded state one
	// would need to use a "group" value of the attribute. This package assigns "folder"
	// state to all group nodes
	node_open_tag := fmt.Sprintf("<node id=\"%s\" yfiles.foldertype=\"folder\">", node_id)

	if _, err := out_file.WriteString(indent + node_open_tag + "\n"); err != nil {
		return errors.New(EMIT_WRITE_ERR_MSG_PREFIX + err.Error())
	}

	// Emit attribute that defines graphical representation of the group node
	// Emit open tag for "nodegraphics" attribute
	ng_attr := yFilesGMLAttrs[yFILES_NATTR_NODEGRAPHICS]
	ng_attr_doc_id := getYFilesAttrDocumentId(ng_attr.id)
	ng_open_tag := fmt.Sprintf("<data key=\"d%d\">", ng_attr_doc_id)
	emit_str := indent + EMIT_INDENT + ng_open_tag + "\n"

	if _, err := out_file.WriteString(emit_str); err != nil {
		return errors.New(EMIT_WRITE_ERR_MSG_PREFIX + err.Error())
	}

	// Emit "y:ProxyAutoBoundsNode" open tag
	emit_str = indent + strings.Repeat(EMIT_INDENT, 2) + "<y:ProxyAutoBoundsNode>\n"

	if _, err := out_file.WriteString(emit_str); err != nil {
		return errors.New(EMIT_WRITE_ERR_MSG_PREFIX + err.Error())
	}

	// Emit "realizers" of the group node. At the moment the code of this function was
	// first written, no documentation existed for yFiles GraphML format. So, I had to
	// figure out on my own what the yFiles GraphML syntax really means. As far as I
	// understood, "realizer" is a concept used to describe behavior and appearance of
	// graph elements that can exist in several states. For example, a group node can
	// exist in two states: it can be folded and unfolded (or to be a "folder" or "group"
	// in yFiles terminology). Hence, there exist two "realizers" for a group node: one
	// that describes its apperance and behavior in an unfolded state and one serving the
	// same purpose for a folded state. The "active" attribute of the "y:Realizers" tag
	// contains a sequence number of a realizer that should be used to describe the
	// corresponding graph element when the yFiles GraphML file is first opened. For
	// example, assume we want a group node to be first shown in a folded state. If a
	// realizer for this state is located second in the file (i.e. after the realizer for
	// an unfolded state), we must assign value "1" to the "active" attribute. A value of
	// the "active" attribute, an order of the realizers in the file and a value of
	// "yfiles.foldertype" attribute of "node" tag must correlate. For example, if
	// "yfiles.foldertype" equals "folder", then the "active" attribute must point to a
	// realizer that describes a folded state of the group node. Empirically I've
	// discovered that the following emitting schemas should be used:
	//   1) to emit a group node that first should be drawn in a folded state:
	//      - yfiles.foldertype="folder"
	//      - active="1"
	//      - and a realizer of the folded state must be emitted second
	//   2) to emit a group node that first should be drawn in an unfolded state:
	//      - yfiles.foldertype="group"
	//      - active="0"
	//      - and a realizer of the unfolded state must be emitted first
	// Other schemas don't work well for some reason

	// Emit "y:Realizers" open tag. The "active" attribute must point to a relizer that
	// describes a folded state of the group node. This realizer will be emitted before
	// the realizer for an unfolded state
	emit_str = indent + strings.Repeat(EMIT_INDENT, 3) + "<y:Realizers active=\"1\">\n"

	if _, err := out_file.WriteString(emit_str); err != nil {
		return errors.New(EMIT_WRITE_ERR_MSG_PREFIX + err.Error())
	}

	// Get group node label
	var nest_label string

	is_emit_label := false
	nest_label_attr := graph_emit_spec.Nest.LabelAttr

	if nest_label_attr != nil {
		if is_set, err := nest.IsStrAttrSet(nest_label_attr); err != nil {
			err_msg := fmt.Sprintf("Error checking whether nest label attribute is "+
				"set [nest ID = %d]: ", nest.GetID())

			return errors.New(err_msg + err.Error())
		} else if is_set {
			nest_label, err = nest.GetStrAttrVal(nest_label_attr)

			if err != nil {
				err_msg := fmt.Sprintf("Error retrieving nest label attribute "+
					"[nest ID = %d]: ", nest.GetID())

				return errors.New(err_msg + err.Error())
			}

			is_emit_label = true
		}
	}

	// Emit realizer of an unfolded state
	// Emit open tag for a realizer of an unfolded state
	emit_str = indent + strings.Repeat(EMIT_INDENT, 4) + "<y:GroupNode>\n"

	if _, err := out_file.WriteString(emit_str); err != nil {
		return errors.New(EMIT_WRITE_ERR_MSG_PREFIX + err.Error())
	}

	// Emit group node label (if any)
	if is_emit_label {
		// It's requested inside the "y:NodeLabel" tag that for unfolded group nodes
		// labels are shown inside the group bounds and at the very top of the group area
		emit_str = indent + strings.Repeat(EMIT_INDENT, 5) +
			"<y:NodeLabel modelName=\"internal\" modelPosition=\"t\">" +
			nest_label + "</y:NodeLabel>\n"

		if _, err := out_file.WriteString(emit_str); err != nil {
			return errors.New(EMIT_WRITE_ERR_MSG_PREFIX + err.Error())
		}
	}

	// Emit "state" tag for a folded node. The "state" must NOT be "closed"
	emit_str = indent + strings.Repeat(EMIT_INDENT, 5) + "<y:State closed=\"false\"/>\n"

	if _, err := out_file.WriteString(emit_str); err != nil {
		return errors.New(EMIT_WRITE_ERR_MSG_PREFIX + err.Error())
	}

	// Force node to adjust its size to accomodate the label
	emit_str = indent + strings.Repeat(EMIT_INDENT, 5) +
		"<y:NodeBounds considerNodeLabelSize=\"true\"/>\n"

	if _, err := out_file.WriteString(emit_str); err != nil {
		return errors.New(EMIT_WRITE_ERR_MSG_PREFIX + err.Error())
	}

	// Emit close tag for a realizer of an unfolded state
	emit_str = indent + strings.Repeat(EMIT_INDENT, 4) + "</y:GroupNode>\n"

	if _, err := out_file.WriteString(emit_str); err != nil {
		return errors.New(EMIT_WRITE_ERR_MSG_PREFIX + err.Error())
	}

	// Emit realizer of a folded state
	// Emit open tag for a realizer of a folded state
	emit_str = indent + strings.Repeat(EMIT_INDENT, 4) + "<y:GroupNode>\n"

	if _, err := out_file.WriteString(emit_str); err != nil {
		return errors.New(EMIT_WRITE_ERR_MSG_PREFIX + err.Error())
	}

	// Emit group node label (if any)
	if is_emit_label {
		emit_str = indent + strings.Repeat(EMIT_INDENT, 5) + "<y:NodeLabel>" +
			nest_label + "</y:NodeLabel>\n"

		if _, err := out_file.WriteString(emit_str); err != nil {
			return errors.New(EMIT_WRITE_ERR_MSG_PREFIX + err.Error())
		}
	}

	// Emit "state" tag for a folded node. The "state" must be "closed"
	emit_str = indent + strings.Repeat(EMIT_INDENT, 5) + "<y:State closed=\"true\"/>\n"

	if _, err := out_file.WriteString(emit_str); err != nil {
		return errors.New(EMIT_WRITE_ERR_MSG_PREFIX + err.Error())
	}

	// NOTE: "considerNodeLabelSize" property of "y:NodeBounds" tag is not emitted for
	// a folded state realizer of a group node. That's because this property is not
	// currently  taken into account by the visualization software. Unfortunately, folded
	// group nodes don't automatically accomodate to the size of their labels

	// Emit close tag for a realizer of a folded state
	emit_str = indent + strings.Repeat(EMIT_INDENT, 4) + "</y:GroupNode>\n"

	if _, err := out_file.WriteString(emit_str); err != nil {
		return errors.New(EMIT_WRITE_ERR_MSG_PREFIX + err.Error())
	}

	// Emit close tag for "y:Realizers"
	emit_str = indent + strings.Repeat(EMIT_INDENT, 3) + "</y:Realizers>\n"

	if _, err := out_file.WriteString(emit_str); err != nil {
		return errors.New(EMIT_WRITE_ERR_MSG_PREFIX + err.Error())
	}

	// Emit close tag for "y:ProxyAutoBoundsNode"
	emit_str = indent + strings.Repeat(EMIT_INDENT, 2) + "</y:ProxyAutoBoundsNode>\n"

	if _, err := out_file.WriteString(emit_str); err != nil {
		return errors.New(EMIT_WRITE_ERR_MSG_PREFIX + err.Error())
	}

	// Emit close tag for "nodegraphics" attribute
	if _, err := out_file.WriteString(indent + EMIT_INDENT + "</data>\n"); err != nil {
		return errors.New(EMIT_WRITE_ERR_MSG_PREFIX + err.Error())
	}

	// Emit subgraph contained inside the node. This is a - potentially - recursive
	// operation. That's because the subgraph may contain other subgraphs that require
	// their own group node wrapper
	err := emitYFilesSubgraph(nest, graph_emit_spec, out_file, id_prefix,
		indent+EMIT_INDENT)

	// Because the above function call is recursive, the prefix of the below error
	// message may be repeated multiple times. It's considered ok for now. Because
	// later, for example, an ID of each intermediate nest could be added to the
	// message (hence, the chain of the exact nests would be reported)
	if err != nil {
		return errors.New("Couldn't emit a nested subgraph")
	}

	// Emit group node close tag
	if _, err := out_file.WriteString(indent + "</node>\n"); err != nil {
		return errors.New(EMIT_WRITE_ERR_MSG_PREFIX + err.Error())
	}

	return nil
}

// Emit an yFiles regular node. "Regular" means that a node is not of a special type
// (like group node, for example)
func emitYFilesRegularNode(node *Node,
	id_prefix string,
	graph_emit_spec *GraphEmitSpec,
	out_file *os.File,
	indent string) error {

	panic_msg_str := "Panic while emitting a yFiles regular node: "

	if node == nil {
		panic(panic_msg_str + "zero reference to the graph node")
	}

	if graph_emit_spec == nil {
		panic(panic_msg_str + "zero reference to a graph emit specification")
	}

	// Emit node open tag
	emit_str := fmt.Sprintf(indent+"<node id=\"%sn%d\">\n", id_prefix, node.GetID())

	if _, err := out_file.WriteString(emit_str); err != nil {
		return errors.New(EMIT_WRITE_ERR_MSG_PREFIX + err.Error())
	}

	// Emit attribute that defines graphical representation of a regular node
	// Emit open tag for "nodegraphics" attribute
	ng_attr := yFilesGMLAttrs[yFILES_NATTR_NODEGRAPHICS]
	ng_attr_doc_id := getYFilesAttrDocumentId(ng_attr.id)
	ng_open_tag := fmt.Sprintf("<data key=\"d%d\">", ng_attr_doc_id)
	emit_str = indent + EMIT_INDENT + ng_open_tag + "\n"

	if _, err := out_file.WriteString(emit_str); err != nil {
		return errors.New(EMIT_WRITE_ERR_MSG_PREFIX + err.Error())
	}

	// Emit "y:ShapeNode" open tag
	emit_str = indent + strings.Repeat(EMIT_INDENT, 2) + "<y:ShapeNode>\n"

	if _, err := out_file.WriteString(emit_str); err != nil {
		return errors.New(EMIT_WRITE_ERR_MSG_PREFIX + err.Error())
	}

	// Emit label of a regular node
	// Get node label
	var node_label string

	is_emit_label := false
	node_label_attr := graph_emit_spec.Node.LabelAttr

	if node_label_attr != nil {
		if is_set, err := node.IsStrAttrSet(node_label_attr); err != nil {
			err_msg := fmt.Sprintf("Error checking whether a node label attribute is "+
				"set [node ID = %d]: ", node.GetID())

			return errors.New(err_msg + err.Error())
		} else if is_set {
			node_label, err = node.GetStrAttrVal(node_label_attr)

			if err != nil {
				err_msg := fmt.Sprintf("Error retrieving a node label attribute "+
					"[node ID = %d]: ", node.GetID())

				return errors.New(err_msg + err.Error())
			}

			is_emit_label = true
		}
	}

	// Emit the label (if any)
	if is_emit_label {
		// Escape the label for XML
		var buf bytes.Buffer

		xml.Escape(&buf, []byte(node_label))
		emit_str = indent + strings.Repeat(EMIT_INDENT, 3) + "<y:NodeLabel>" +
			buf.String() + "</y:NodeLabel>\n"

		if _, err := out_file.WriteString(emit_str); err != nil {
			return errors.New(EMIT_WRITE_ERR_MSG_PREFIX + err.Error())
		}
	}

	// Emit close tag for "y:ShapeNode"
	emit_str = indent + strings.Repeat(EMIT_INDENT, 2) + "</y:ShapeNode>\n"

	if _, err := out_file.WriteString(emit_str); err != nil {
		return errors.New(EMIT_WRITE_ERR_MSG_PREFIX + err.Error())
	}

	// Emit close tag for "nodegraphics" attribute
	if _, err := out_file.WriteString(indent + EMIT_INDENT + "</data>\n"); err != nil {
		return errors.New(EMIT_WRITE_ERR_MSG_PREFIX + err.Error())
	}

	// Emit node close tag
	if _, err := out_file.WriteString(indent + "</node>\n"); err != nil {
		return errors.New(EMIT_WRITE_ERR_MSG_PREFIX + err.Error())
	}

	return nil
}

// Calculate yFiles document id of a graph node. The document id of a node is used to
// identify the node in a yFiles GraphML document
//
// NOTE: currently the document id of the same graph node may be caclulated several times:
//       1) one time when the node is emitted. In this scenario the id is calculated in a
//          very efficient way, without a call the below function
//       2) zero or many times when emitting edges adjacent to this node. In this scenario
//          the id will be calculated by means of a call to the below function.
//       The second option is inefficient (especially, taking into account that it can be
//       used several times for the same node). For now, it is left as is. But in future,
//       if it becomes a bottleneck, node ids required to emit edges may be obtained in a
//       different way
func emitCalcYFilesNodeDocumentId(node *Node) (string, error) {
	panic_msg_str := "Panic while calculating yFiles GraphML document id of a graph " +
		"node: "

	graph := node.GetGraph()

	if graph == nil {
		panic(panic_msg_str + "the node is not attributed to any graph")
	}

	nest_tree := graph.GetNestTree()

	if nest_tree == nil {
		panic(panic_msg_str + "the graph to which the node belongs has a zero " +
			"reference to the nest tree")
	}

	node_doc_id := fmt.Sprintf("n%d", node.GetID())
	nest := node.GetNest()

	if nest == nil {
		panic(panic_msg_str + "the node has zero reference to the containing nest")
	}

	for ; nest.GetParentNest() != nil; nest = nest.GetParentNest() {
		if nest.GetNestTree() != nest_tree {
			panic(panic_msg_str + "one of the ancestor nests refers to a different " +
				"nest tree than the graph to which the node belongs")
		}

		node_doc_id = fmt.Sprintf("nest%d::%s", nest.GetID(), node_doc_id)
	}

	return node_doc_id, nil
}

// Emit an yFiles edge
func emitYFilesEdge(edge *Edge,
	id_prefix string,
	out_file *os.File,
	indent string) error {

	panic_msg_str := "Panic while emitting an yFiles edge: "

	if edge.GetSrcNode() == nil || edge.GetDstNode() == nil {
		return errors.New("At least one end of the edge is not connected to any " +
			"graph node")
	}

	src_node := edge.GetSrcNode()
	dst_node := edge.GetDstNode()
	graph := edge.GetGraph()

	if src_node.GetGraph() != graph || dst_node.GetGraph() != graph {
		panic(panic_msg_str + "at least one of the nodes connected by the edge " +
			"is attributed to a different graph than the edge itself")
	}

	var src_node_doc_id, dst_node_doc_id string
	var err error

	if src_node_doc_id, err = emitCalcYFilesNodeDocumentId(src_node); err != nil {
		return errors.New("Couldn't calculate the source node's yFiles GraphML " +
			"document id")
	}

	if dst_node_doc_id, err = emitCalcYFilesNodeDocumentId(dst_node); err != nil {
		return errors.New("Couldn't calculate the destination node's yFiles GraphML " +
			"document id")
	}

	// Emit edge open tag
	emit_str := fmt.Sprintf(indent+"<edge id=\"%se%d\" source=\"%s\" target=\"%s\">\n",
		id_prefix, edge.GetID(), src_node_doc_id, dst_node_doc_id)

	if _, err := out_file.WriteString(emit_str); err != nil {
		return errors.New(EMIT_WRITE_ERR_MSG_PREFIX + err.Error())
	}

	// Emit edge close tag
	if _, err := out_file.WriteString(indent + "</edge>\n"); err != nil {
		return errors.New(EMIT_WRITE_ERR_MSG_PREFIX + err.Error())
	}

	return nil
}

// Emit nodes and edges of an yFiles subgraph represented by a nest
func emitYFilesSubgraphNodesAndEdges(nest *Nest,
	id_prefix string,
	graph_emit_spec *GraphEmitSpec,
	out_file *os.File,
	indent string) error {

	panic_msg_str := "Panic while emitting edges and nodes of an yFiles subgraph: "

	if nest.GetNestTree() == nil {
		panic(panic_msg_str + "a nest representing the subgraph is not linked to any " +
			"nest tree")
	}

	graph := nest.GetNestTree().GetBaseGraph()

	if graph == nil {
		panic(panic_msg_str + "a nest tree to which a nest representing the subgraph " +
			"belongs is not linked to any graph")
	}

	// Emit graph nodes belonging to the nest
	for node := nest.GetFirstNode(); node != nil; node = node.GetNextNodeInNest() {
		err := emitYFilesRegularNode(node, id_prefix, graph_emit_spec, out_file, indent)

		if err != nil {
			return errors.New("Error emitting an yFiles regular node: " + err.Error())
		}
	}

	// Emit graph edges belonging to the nest
	for edge := nest.GetFirstEdge(); edge != nil; edge = edge.GetNextEdgeInNest() {
		if edge.GetGraph() != graph {
			panic(panic_msg_str + "an edge belonging to a nest representing the " +
				"subgraph is attributed to a different graph than the nest itself")
		}

		err := emitYFilesEdge(edge, id_prefix, out_file, indent)

		if err != nil {
			return errors.New("Error emitting an yFiles edge: " + err.Error())
		}
	}

	return nil
}

// Emit subgraph associated with a specific nest (including all nests transitively
// contained inside this nest). The entire graph will be emitted, if this function is
// called for the root nest
func emitYFilesSubgraph(nest *Nest,
	graph_emit_spec *GraphEmitSpec,
	out_file *os.File,
	id_prefix *string,
	indent string) error {

	panic_msg_str := "Panic while emitting an yFiles subgraph: "

	if nest == nil {
		panic(panic_msg_str + "zero reference to a nest containing the subgraph")
	}

	var graph_id string
	var new_id_prefix string

	// Zero ID prefix means that the function was called for the root nest
	if id_prefix == nil {
		// For some reason in all GraphML examples that I've seen the entire graph has
		// id="G". Seems, there is no such requirement in the basic GraphML specification,
		// but this convention (whether it's formal or not) is followed by lots of people.
		// This package also follows this "convention". Let it be "G" :)
		graph_id = "G"
		new_id_prefix = ""
	} else {
		graph_id = fmt.Sprintf("%snest%d:", *id_prefix, nest.GetID())
		new_id_prefix = graph_id + ":"
	}

	// Emit "graph" open tag
	graph_open_tag := fmt.Sprintf("<graph id=\"%s\" edgedefault=\"directed\">", graph_id)

	if _, err := out_file.WriteString(indent + graph_open_tag + "\n"); err != nil {
		return errors.New(EMIT_WRITE_ERR_MSG_PREFIX + err.Error())
	}

	// Emit contained node groups first. After that nodes and edges of the current
	// subgraph will be emitted
	child_nest := nest.GetFirstChildNest()

	for ; child_nest != nil; child_nest = child_nest.GetNextSiblingNest() {
		if nest.GetNestTree() != child_nest.GetNestTree() {
			panic(panic_msg_str + "a child nest belongs to a different nest tree or is " +
				"not linked to any nest tree at all")
		}

		err := emitYFilesGroup(child_nest, graph_emit_spec, out_file, &new_id_prefix,
			indent+EMIT_INDENT)

		// Because of the recursive call in this loop, the prefix of the below error
		// message may be repeated multiple times. It's considered ok for now. Because
		// later, for example, an ID of each intermediate nest could be added to the
		// message (hence, the chain of the exact nests would be reported)
		if err != nil {
			return errors.New("Couldn't emit a nested node group: " + err.Error())
		}
	}

	err := emitYFilesSubgraphNodesAndEdges(nest, new_id_prefix, graph_emit_spec, out_file,
		indent+EMIT_INDENT)

	if err != nil {
		return errors.New("Error while emitting nodes and edges of an yFiles subgraph: " +
			err.Error())
	}

	// Emit "graph" close tag
	if _, err := out_file.WriteString(indent + "</graph>\n"); err != nil {
		return errors.New(EMIT_WRITE_ERR_MSG_PREFIX + err.Error())
	}

	return nil
}

func EmitInYFilesFormat(graph *Graph,
	graph_emit_spec *GraphEmitSpec,
	out_path string) error {

	panic_msg_prefix := "Panic while emitting a graph in yFiles format: "

	if err := checkYFilesAttrArrayConsistency(); err != nil {
		panic(panic_msg_prefix + "consistency check on an array describing yFiles " +
			"GraphML attributes has failed: " + err.Error())
	}

	out_file, err := os.OpenFile(out_path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)

	if err != nil {
		return errors.New("Cannot create output file: " + err.Error())
	}

	defer out_file.Close()

	// If no emit specification is provided, we create the default one. We do that to
	// simplify the code, so that we don't need to check whether graph_emit_spec is
	// "nil" every time we're going to use it
	// NOTE: here the function parameter "graph_emit_spec" is intentionally re-assigned
	if graph_emit_spec == nil {
		graph_emit_spec = &GraphEmitSpec{}
	}

	EMIT_WRITE_ERR_MSG_PREFIX := "Cannot write to the output file: "
	// Emit "xml" clause
	_, err = out_file.WriteString("<?xml version=\"1.0\" encoding=\"UTF-8\" " +
		"standalone=\"no\"?>\n")

	if err != nil {
		return errors.New(EMIT_WRITE_ERR_MSG_PREFIX + err.Error())
	}

	// Emit "graphml" open tag
	_, err = out_file.WriteString("<graphml " +
		"xmlns=\"http://graphml.graphdrawing.org/xmlns\" " +
		"xmlns:sys=\"http://www.yworks.com/xml/yfiles-common/markup/primitives/2.0\" " +
		"xmlns:x=\"http://www.yworks.com/xml/yfiles-common/markup/2.0\" " +
		"xmlns:xsi=\"http://www.w3.org/2001/XMLSchema-instance\" " +
		"xmlns:y=\"http://www.yworks.com/xml/graphml\" " +
		"xsi:schemaLocation=\"http://graphml.graphdrawing.org/xmlns " +
		"http://www.yworks.com/xml/schema/graphml/1.1/ygraphml.xsd\">\n")

	if err != nil {
		return errors.New(EMIT_WRITE_ERR_MSG_PREFIX + err.Error())
	}

	// Emit declarations of YFiles GraphML attributes
	if err := emitYFilesAttrDecls(out_file, EMIT_INDENT); err != nil {
		return errors.New("Error while emitting yFiles GraphML attribute declarations: " +
			err.Error())
	}

	// Obtain a reference to the root nest
	if graph.GetNestTree() == nil {
		panic(panic_msg_prefix + "the graph doesn't have a nest tree")
	}

	root_nest := graph.GetNestTree().GetRootNest()

	if root_nest == nil {
		panic(panic_msg_prefix + "the graph doesn't have a root nest")
	}

	// Emit the entire graph
	err = emitYFilesSubgraph(root_nest, graph_emit_spec, out_file, nil, EMIT_INDENT)

	// Emit "graphml" close tag
	if _, err := out_file.WriteString("</graphml>"); err != nil {
		return errors.New(EMIT_WRITE_ERR_MSG_PREFIX + err.Error())
	}

	/*
		// Emit Graph global properties
		// Drawing orientation property: left to right
		if _, err := out_file.WriteString("\trankdir = LR\n"); err != nil {
			return errors.New(EMIT_WRITE_ERR_MSG_PREFIX + err.Error())
		}

		err = emitGVSubgraphNodesAndEdges(root_nest, graph_emit_spec, out_file, EMIT_INDENT)

		if err != nil {
			return errors.New("Couldn't emit nodes and edges belonging to the root nest: " +
				err.Error())
		}

	*/
	return nil
}

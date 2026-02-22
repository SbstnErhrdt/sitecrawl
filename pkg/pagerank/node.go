package pagerank

// NodeID identifies a graph node.
type NodeID string

// String returns the string representation of a node ID.
func (n *NodeID) String() string {
	if n == nil {
		return ""
	}
	return string(*n)
}

// Node is a directed-graph vertex with rank and adjacency maps.
type Node struct {
	Id       NodeID
	Rank     float64
	Outgoing map[EdgeID]*Edge
	Incoming map[EdgeID]*Edge
}

// NewNode constructs a node with empty adjacency maps.
func NewNode(id string) *Node {
	node := Node{
		Id:       NodeID(id),
		Rank:     0,
		Outgoing: map[EdgeID]*Edge{},
		Incoming: map[EdgeID]*Edge{},
	}
	return &node
}

// OutDegree is the number of neighbors of a node based on outgoing edges
func (n *Node) OutDegree() uint {
	return uint(len(n.Outgoing))
}

// InDegree is the number of neighbors of a node based on incoming edges
func (n *Node) InDegree() uint {
	return uint(len(n.Incoming))
}

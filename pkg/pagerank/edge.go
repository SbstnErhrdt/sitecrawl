package pagerank

// EdgeID is the unique identifier of a directed edge.
type EdgeID string

// Edge represents a weighted directed edge.
type Edge struct {
	Id     EdgeID
	From   *Node
	To     *Node
	Weight float64
}

// GenerateEdgeID builds a stable edge identifier from source and destination nodes.
func GenerateEdgeID(from, to *Node) EdgeID {
	return EdgeID(from.Id + ":" + to.Id)
}

// NewEdge creates a directed edge and attaches it to source/target node maps.
func NewEdge(from, to *Node, weight float64) *Edge {
	edgeID := GenerateEdgeID(from, to)
	edge := Edge{
		Id:     edgeID,
		From:   from,
		To:     to,
		Weight: weight,
	}
	// create links
	from.Outgoing[edgeID] = &edge
	to.Incoming[edgeID] = &edge
	return &edge
}

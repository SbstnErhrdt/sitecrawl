package crawler

import (
	"sort"

	"github.com/sbstn/sitecrawl/pkg/pagerank"
)

type LinkGraph struct {
	nodes map[string]struct{}
	edges map[string]map[string]struct{}
}

// NewLinkGraph creates an in-memory adjacency graph for rank computation.
func NewLinkGraph() *LinkGraph {
	return &LinkGraph{
		nodes: map[string]struct{}{},
		edges: map[string]map[string]struct{}{},
	}
}

// AddNode ensures a node is present in the graph.
func (g *LinkGraph) AddNode(node string) {
	if node == "" {
		return
	}
	g.nodes[node] = struct{}{}
}

// AddEdge adds a directed edge between two normalized URLs.
func (g *LinkGraph) AddEdge(from, to string) {
	if from == "" || to == "" {
		return
	}
	g.AddNode(from)
	g.AddNode(to)
	if _, ok := g.edges[from]; !ok {
		g.edges[from] = map[string]struct{}{}
	}
	g.edges[from][to] = struct{}{}
}

// ComputePageRankScores adapts LinkGraph into pkg/pagerank and returns score + order.
func ComputePageRankScores(g *LinkGraph) (map[string]float64, []string) {
	scores := map[string]float64{}
	if g == nil || len(g.nodes) == 0 {
		return scores, nil
	}

	prGraph := pagerank.NewGraph()
	for node := range g.nodes {
		prGraph.AddNode(node)
	}
	for from, targets := range g.edges {
		for to := range targets {
			prGraph.AddEdge(from, to)
		}
	}

	pr := pagerank.NewPageRank(prGraph)
	pr.CalcPageRank()

	if len(prGraph.Edges) == 0 {
		order := make([]string, 0, len(g.nodes))
		uniform := 1.0 / float64(len(g.nodes))
		for node := range g.nodes {
			scores[node] = uniform
			order = append(order, node)
		}
		sort.Strings(order)
		return scores, order
	}

	for nodeID, node := range prGraph.Nodes {
		scores[string(nodeID)] = node.Rank
	}

	pr.OrderResults()
	ids := pr.GetMaxToMinOrder()
	order := make([]string, 0, len(ids))
	for _, nodeID := range ids {
		order = append(order, string(nodeID))
	}
	return scores, order
}

// ApplyPageRankScores annotates crawl pages with scores and score-desc ordering.
func ApplyPageRankScores(result *CrawlResult, g *LinkGraph) {
	if result == nil {
		return
	}

	result.PageRankImplementation = PageRankImplementation
	scores, order := ComputePageRankScores(g)
	orderPos := map[string]int{}
	for idx, url := range order {
		orderPos[url] = idx
	}

	for _, page := range result.Pages {
		if page.Status != StatusOK {
			continue
		}
		key := page.FinalURL
		if key == "" {
			key = page.URL
		}
		score := scores[key]
		page.Score = &score
	}

	sort.SliceStable(result.Pages, func(i, j int) bool {
		iScore := -1.0
		jScore := -1.0
		if result.Pages[i].Score != nil {
			iScore = *result.Pages[i].Score
		}
		if result.Pages[j].Score != nil {
			jScore = *result.Pages[j].Score
		}
		if iScore != jScore {
			return iScore > jScore
		}
		iURL := result.Pages[i].FinalURL
		if iURL == "" {
			iURL = result.Pages[i].URL
		}
		jURL := result.Pages[j].FinalURL
		if jURL == "" {
			jURL = result.Pages[j].URL
		}
		iPos, iOK := orderPos[iURL]
		jPos, jOK := orderPos[jURL]
		if iOK && jOK && iPos != jPos {
			return iPos < jPos
		}
		return iURL < jURL
	})
}

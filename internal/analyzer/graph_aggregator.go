package analyzer

import (
	"github.com/halleck45/ast-metrics/internal/engine"
	pb "github.com/halleck45/ast-metrics/pb"
	"sort"
)

type GraphAggregator struct{}

func NewGraphAggregator() *GraphAggregator { return &GraphAggregator{} }

func (ga *GraphAggregator) Calculate(aggregate *Aggregated) {
	if aggregate == nil {
		return
	}
	if aggregate.Graph == nil {
		aggregate.Graph = &pb.Graph{Nodes: make(map[string]*pb.Node)}
	}

	ensureNode := func(id string, name *pb.Name) *pb.Node {
		if id == "" {
			return nil
		}
		if aggregate.Graph.Nodes[id] == nil {
			n := &pb.Node{Id: id}
			if name != nil {
				n.Name = name
			} else {
				n.Name = &pb.Name{Qualified: id, Short: id}
			}
			aggregate.Graph.Nodes[id] = n
		}
		return aggregate.Graph.Nodes[id]
	}

	// keep a per-run set of edges to avoid O(N^2) duplicate checks
	edgesSeen := make(map[string]map[string]struct{})

	addEdge := func(from, to string) {
		if from == "" || to == "" || from == to {
			return
		}
		n := ensureNode(from, nil)
		ensureNode(to, nil)
		if n == nil {
			return
		}
		if edgesSeen[from] == nil {
			edgesSeen[from] = make(map[string]struct{})
		}
		if _, exists := edgesSeen[from][to]; exists {
			return
		}
		edgesSeen[from][to] = struct{}{}
		n.Edges = append(n.Edges, to)
	}

	// Prepare weighted package edges (package-only projection)
	edgesCount := make(map[string]map[string]int)

	for _, file := range aggregate.ConcernedFiles {
		// Skip empty files
		if file == nil || file.Stmts == nil {
			continue
		}
		// Gather dependencies at file level
		deps := engine.GetDependenciesInFile(file)
		for _, dep := range deps {
			if dep == nil {
				continue
			}
			// Use a moderate namespace depth to aggregate weights and keep meaningful edges
			fromNs := engine.ReduceDepthOfNamespace(dep.From, 3)
			toNs := engine.ReduceDepthOfNamespace(dep.Namespace, 3)
			if fromNs == "" || toNs == "" || fromNs == toNs {
				continue
			}
			if edgesCount[fromNs] == nil {
				edgesCount[fromNs] = make(map[string]int)
			}
			edgesCount[fromNs][toNs]++
		}
	}

	// Apply combined filtering: abs + relative threshold and top-K per source, with fallback top-1
	const (
		absThreshold = 1
		relThreshold = 0.10 // 10% of total outgoing from source
		topK         = 5
	)
	// compute total outgoing per source
	totalOut := make(map[string]int, len(edgesCount))
	for fromNs, tos := range edgesCount {
		sum := 0
		for _, w := range tos {
			sum += w
		}
		totalOut[fromNs] = sum
	}
	for fromNs, tos := range edgesCount {
		// ensure source node exists and is marked as package
		ensureNode(fromNs, &pb.Name{Qualified: fromNs, Short: fromNs, Package: fromNs})
		// collect and sort targets by weight desc
		type pair struct {
			to string
			w  int
		}
		arr := make([]pair, 0, len(tos))
		for toNs, w := range tos {
			arr = append(arr, pair{to: toNs, w: w})
		}
		sort.Slice(arr, func(i, j int) bool { return arr[i].w > arr[j].w })
		kept := map[string]bool{}
		for i, p := range arr {
			if i < topK || p.w >= absThreshold || float64(p.w) >= relThreshold*float64(totalOut[fromNs]) {
				kept[p.to] = true
			}
			// always ensure target node exists for visibility even if edge filtered
			ensureNode(p.to, &pb.Name{Qualified: p.to, Short: p.to, Package: p.to})
		}
		// fallback: if nothing kept but at least one candidate, keep the strongest edge
		if len(kept) == 0 && len(arr) > 0 {
			kept[arr[0].to] = true
		}
		for toNs := range kept {
			addEdge(fromNs, toNs)
		}
	}
}

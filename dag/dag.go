// Package dag implements DAG construction, topological sorting, and cycle detection.
package dag

import (
	"fmt"
	"sort"
)

// Node represents a job in the DAG.
type Node struct {
	ID    string
	Needs []string
	Depth int // computed execution depth (stage)
}

// Graph is the dependency graph for a Pipeline.
type Graph struct {
	Nodes map[string]*Node
}

// NewGraph builds a DAG from job definitions.
func NewGraph(jobs map[string]struct{ Needs []string }) (*Graph, error) {
	nodes := make(map[string]*Node, len(jobs))
	for id := range jobs {
		nodes[id] = &Node{ID: id, Needs: nil}
	}

	// validate edges and build adjacency
	for id, job := range jobs {
		for _, dep := range job.Needs {
			if _, ok := nodes[dep]; !ok {
				return nil, fmt.Errorf("job %q needs unknown job %q", id, dep)
			}
		}
		nodes[id].Needs = append(nodes[id].Needs, job.Needs...)
	}

	// cycle detection via DFS
	visited := make(map[string]bool)
	recStack := make(map[string]bool)
	var detect func(string) error
	detect = func(id string) error {
		visited[id] = true
		recStack[id] = true
		for _, dep := range nodes[id].Needs {
			if !visited[dep] {
				if err := detect(dep); err != nil {
					return err
				}
			} else if recStack[dep] {
				return fmt.Errorf("cycle detected involving %q", id)
			}
		}
		recStack[id] = false
		return nil
	}
	for id := range nodes {
		if !visited[id] {
			if err := detect(id); err != nil {
				return nil, err
			}
		}
	}

	// compute depths for implicit staging
	memo := make(map[string]int)
	var depth func(string) int
	depth = func(id string) int {
		if d, ok := memo[id]; ok {
			return d
		}
		max := 0
		for _, dep := range nodes[id].Needs {
			if d := depth(dep) + 1; d > max {
				max = d
			}
		}
		memo[id] = max
		nodes[id].Depth = max
		return max
	}
	for id := range nodes {
		depth(id)
	}

	return &Graph{Nodes: nodes}, nil
}

// TopoSort returns a topologically ordered slice of node IDs.
// All roots (no needs) appear first.
func (g *Graph) TopoSort() []string {
	// Build in-degree map and dependents map
	inDegree := make(map[string]int, len(g.Nodes))
	dependents := make(map[string][]string)
	for id := range g.Nodes {
		inDegree[id] = 0
	}
	for id, n := range g.Nodes {
		for _, dep := range n.Needs {
			dependents[dep] = append(dependents[dep], id)
			inDegree[id]++
		}
	}

	queue := []string{}
	for id, d := range inDegree {
		if d == 0 {
			queue = append(queue, id)
		}
	}
	sort.Strings(queue)

	result := []string{}
	for len(queue) > 0 {
		id := queue[0]
		queue = queue[1:]
		result = append(result, id)

		for _, dep := range dependents[id] {
			inDegree[dep]--
			if inDegree[dep] == 0 {
				queue = append(queue, dep)
			}
		}
		sort.Strings(queue)
	}

	return result
}

// Roots returns all nodes with no dependencies, sorted alphabetically.
func (g *Graph) Roots() []string {
	var roots []string
	for id, n := range g.Nodes {
		if len(n.Needs) == 0 {
			roots = append(roots, id)
		}
	}
	sort.Strings(roots)
	return roots
}

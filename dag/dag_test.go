package dag

import (
	"slices"
	"testing"
)

func TestNewGraphSuccess(t *testing.T) {
	jobs := map[string]struct{ Needs []string }{
		"a": {},
		"b": {Needs: []string{"a"}},
		"c": {Needs: []string{"a"}},
		"d": {Needs: []string{"b", "c"}},
	}
	g, err := NewGraph(jobs)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(g.Nodes) != 4 {
		t.Fatalf("expected 4 nodes, got %d", len(g.Nodes))
	}
	if g.Nodes["d"].Depth != 2 {
		t.Errorf("expected depth of d = 2, got %d", g.Nodes["d"].Depth)
	}
}

func TestCycleDetection(t *testing.T) {
	jobs := map[string]struct{ Needs []string }{
		"a": {Needs: []string{"c"}},
		"b": {Needs: []string{"a"}},
		"c": {Needs: []string{"b"}},
	}
	if _, err := NewGraph(jobs); err == nil {
		t.Fatal("expected cycle detection error")
	}
}

func TestUnknownDependency(t *testing.T) {
	jobs := map[string]struct{ Needs []string }{
		"a": {Needs: []string{"x"}},
	}
	if _, err := NewGraph(jobs); err == nil {
		t.Fatal("expected unknown dependency error")
	}
}

func TestTopoSort(t *testing.T) {
	jobs := map[string]struct{ Needs []string }{
		"build":  {},
		"test":   {Needs: []string{"build"}},
		"lint":   {},
		"deploy": {Needs: []string{"test"}},
	}
	g, err := NewGraph(jobs)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	order := g.TopoSort()
	if len(order) != 4 {
		t.Fatalf("expected 4 items, got %d", len(order))
	}
	if idx := slices.Index(order, "test"); idx < 0 || idx < slices.Index(order, "build") {
		t.Errorf("test should come after build")
	}
	if idx := slices.Index(order, "deploy"); idx < 0 || idx < slices.Index(order, "test") {
		t.Errorf("deploy should come after test")
	}
}

func TestRoots(t *testing.T) {
	jobs := map[string]struct{ Needs []string }{
		"x": {},
		"y": {},
		"z": {Needs: []string{"x"}},
	}
	g, err := NewGraph(jobs)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	roots := g.Roots()
	if len(roots) != 2 {
		t.Fatalf("expected 2 roots, got %d", len(roots))
	}
	if roots[0] != "x" || roots[1] != "y" {
		t.Errorf("expected roots x,y got %v", roots)
	}
}

func TestDepthOnly(t *testing.T) {
	jobs := map[string]struct{ Needs []string }{
		"a": {},
		"b": {Needs: []string{"a"}},
		"c": {Needs: []string{"a"}},
		"d": {Needs: []string{"b", "c"}},
	}
	g, _ := NewGraph(jobs)
	expected := map[string]int{"a": 0, "b": 1, "c": 1, "d": 2}
	for id, want := range expected {
		if got := g.Nodes[id].Depth; got != want {
			t.Errorf("depth(%s) = %d, want %d", id, got, want)
		}
	}
}

func TestParallelRoots(t *testing.T) {
	jobs := map[string]struct{ Needs []string }{
		"a": {}, "b": {}, "c": {}, "d": {},
	}
	g, err := NewGraph(jobs)
	if err != nil {
		t.Fatal(err)
	}
	sort := g.TopoSort()
	if len(sort) != 4 {
		t.Fatalf("expected 4, got %d", len(sort))
	}
	if slices.Index(sort, "a") == -1 || slices.Index(sort, "b") == -1 || slices.Index(sort, "c") == -1 || slices.Index(sort, "d") == -1 {
		t.Errorf("expected all nodes in topo sort")
	}
}

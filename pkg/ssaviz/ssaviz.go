// Package ssaviz helps user to visualize SSA function.
package ssaviz

// This file contains public APIs of go-ssaviz.

import (
	"fmt"

	"golang.org/x/tools/go/ssa"
)

// Package meta informations.
const (
	Prog    = "ssaviz"
	Version = "0.1.0"
	Author  = "Shengyu Zhang <silverrainz.me>"
	Desc    = "Visualize Go SSA function using Graphviz"
)

var (
	// Debug decides whether to output debug logs.
	Debug bool
)

// Build builds a graph of specific [Kind] from SSA function.
func Build(kind Kind, f *ssa.Function) (*Graph, error) {
	switch kind {
	case CFG:
		return buildCFG(f), nil
	default:
		return nil, fmt.Errorf("unknown graph kind: %s", kind)
	}
}

// Report aggregates multiple graphs to a single HTML report.
func Report(graphs []*Graph) (*HTML, error) {
	return buildHTML(graphs)
}

package ssaviz

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"

	"github.com/emicklei/dot"
)

// Kind is kind of graph.
type Kind string

const (
	CFG Kind = "Control Flow Graph"
)

// Format is file format of rendered graph.
type Format string

const (
	DOT Format = "dot"
	SVG Format = "svg"
)

func (s Format) Ext() string {
	switch s {
	case DOT:
		return ".gv"
	case SVG:
		return ".svg"
	default:
		return ".unknown"
	}
}

// Attr is attribute of vertex and edge of Graphviz.
type Attr string

var (
	Fontame Attr = "Courier New"
)

type Graph struct {
	Name string
	Kind Kind
	g    *dot.Graph
}

// Render renders to specific format. See [Format] for available formats.
func (g Graph) Render(format Format) ([]byte, error) {
	dot := g.g.String()
	switch format {
	case DOT:
		return []byte(dot), nil
	case SVG:
		var svg bytes.Buffer
		cmd := exec.Command("dot", "-Tsvg")
		cmd.Stdin = strings.NewReader(dot)
		cmd.Stdout = &svg
		if err := cmd.Run(); err != nil {
			exitErr := err.(*exec.ExitError)
			return nil, fmt.Errorf("failed to run %q: %w, stdout: %s", cmd, exitErr, string(exitErr.Stderr))
		}
		return svg.Bytes(), nil
	default:
		return nil, fmt.Errorf("unknown render format: %s", format)
	}
}

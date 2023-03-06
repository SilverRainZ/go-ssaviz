package ssaviz

import (
	"bytes"
	"fmt"
	"go/types"
	"log"
	"strconv"
	"strings"

	"github.com/beevik/etree"
	"github.com/emicklei/dot"
	"golang.org/x/tools/go/ssa"
)

// buildCFG visualize an SSA function as a directed graph.
func buildCFG(f *ssa.Function) *Graph {
	g := &Graph{
		Name: f.Name(),
		Kind: CFG,
		g:    dot.NewGraph(dot.Directed),
	}

	buildCFGInfo(g.g, f)
	buildCFGNode(g.g, f.Blocks[0]) // Blocks[0] is entry of function

	if Debug {
		log.Printf("%s for %s: %s", g.Kind, g.Name, g.g.String())
	}

	return g
}

func genCFGNodeID(blk *ssa.BasicBlock) string {
	return strconv.Itoa(blk.Index)
}

// genCFGNodeLabel generates basic block label in Graphviz HTML.
//
// This function borrows a lot from [golang.org/x/tools/go/ssa.WriteFunction].
func genCFGNodeLabel(blk *ssa.BasicBlock) string {
	const punchcard = 50

	doc := etree.NewDocument()

	table := doc.CreateElement("table")
	table.CreateAttr("border", "0")
	table.CreateAttr("cellborder", "0")
	table.CreateAttr("cellspacing", "0")
	table.CreateAttr("cellpadding", "4")

	header := table.CreateElement("tr").CreateElement("td")
	header.CreateAttr("bgcolor", "2")
	header.CreateAttr("balign", "left")
	header = header.CreateElement("font")
	header.CreateAttr("color", "white")
	bidx := fmt.Sprintf("%d:", blk.Index)
	header.CreateText(bidx)
	bmsg := fmt.Sprintf("%s P:%d S:%d", blk.Comment, len(blk.Preds), len(blk.Succs))
	header.CreateText(fmt.Sprintf("%*s%s", punchcard-1-len(bidx)-len(bmsg), "", bmsg))
	header.CreateElement("br")

	body := table.CreateElement("tr").CreateElement("td")
	body.CreateAttr("bgcolor", "1")
	body.CreateAttr("balign", "left")
	body = body.CreateElement("font")
	for _, instr := range blk.Instrs {
		switch v := instr.(type) {
		case ssa.Value:
			l := punchcard
			// Left-align the instruction.
			if name := v.Name(); name != "" {
				s := fmt.Sprintf("%s = ", name)
				body.CreateText(s)
				l -= len(s)
			}
			s := instr.String()
			body.CreateText(s)
			l -= len(s)
			// Right-align the type if there's space.
			if t := v.Type(); t != nil {
				body.CreateText(" ")
				ts := relType(t, blk.Parent().Pkg.Pkg)
				l -= len(ts) + len("  ") // (spaces before and after type)
				if l > 0 {
					body.CreateText(fmt.Sprintf("%*s", l, ""))
				}
				body.CreateText(ts)
			}
		case nil:
			// Be robust against bad transforms.
			s := "<deleted>"
			body.CreateText(fmt.Sprintf("%s%*s", s, punchcard-len(s), ""))
		default:
			s := instr.String()
			body.CreateText(fmt.Sprintf("%s%*s", s, punchcard-len(s), ""))
		}
		body.CreateElement("br")
	}

	var buf bytes.Buffer
	doc.WriteTo(&buf)
	return "<" + buf.String() + ">"
}

// genCFGInfo generates graph info in Graphviz HTML.
//
// This function borrows a lot from [golang.org/x/tools/go/ssa.WriteFunction].
func genCFGInfo(f *ssa.Function) string {
	doc := etree.NewDocument()

	doc.CreateText("Kind: " + string(CFG))
	doc.CreateElement("br").CreateAttr("align", "left")
	doc.CreateText("Name: " + f.String())
	doc.CreateElement("br").CreateAttr("align", "left")

	if f.Pkg != nil {
		doc.CreateText("Package: " + f.Pkg.Pkg.Path())
		doc.CreateElement("br").CreateAttr("align", "left")
	}

	if syn := f.Synthetic; syn != "" {
		doc.CreateText("Synthetic: " + syn)
		doc.CreateElement("br").CreateAttr("align", "left")
	}

	if pos := f.Pos(); pos.IsValid() {
		doc.CreateText("Location: " + f.Prog.Fset.Position(pos).String())
		doc.CreateElement("br").CreateAttr("align", "left")
	}

	if f.Recover != nil {
		doc.CreateText("Recover: " + f.Recover.String())
		doc.CreateElement("br").CreateAttr("align", "left")
	}

	var buf bytes.Buffer
	doc.WriteTo(&buf)
	return "<" + buf.String() + ">"
}

func buildCFGNode(g *dot.Graph, blk *ssa.BasicBlock) dot.Node {
	if n, ok := g.FindNodeById(genCFGNodeID(blk)); ok {
		return n // vertex already exsts.
	}

	n := g.Node(genCFGNodeID(blk)).
		Attr("shape", "none").
		Attr("label", dot.Literal(genCFGNodeLabel(blk))).
		Attr("fontname", Fontame).
		Attr("colorscheme", "blues3")

	for _, succ := range blk.Succs {
		next := buildCFGNode(g, succ)
		_ = g.Edge(n, next)
	}

	return n
}

func buildCFGInfo(g *dot.Graph, f *ssa.Function) {
	sg := g.Subgraph("info")
	sg.Node("info").
		Attr("shape", "box").
		Attr("label", dot.Literal(genCFGInfo(f))).
		Attr("fontname", Fontame).
		Attr("colorscheme", "blues3").
		Attr("color", "2").
		Attr("fillcolor", "1")
}

func relType(t types.Type, from *types.Package) string {
	s := types.TypeString(t, types.RelativeTo(from))
	s = strings.ReplaceAll(s, "interface{}", "any")
	return s
}

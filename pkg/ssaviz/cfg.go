package ssaviz

import (
	"bytes"
	"fmt"
	"go/token"
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

func genInstrTooltip(instr ssa.Instruction) string {
	var buf bytes.Buffer

	fmt.Fprintf(&buf, "Type: %T\n", instr)
	if p := instr.Pos(); p != token.NoPos {
		pos := instr.Parent().Prog.Fset.Position(p)
		fmt.Fprintf(&buf, "Line: %d\n", pos.Line)
	}

	if v, ok := instr.(ssa.Value); ok {
		if refs := v.Referrers(); refs != nil {
			fmt.Fprintf(&buf, "Referrers: %d\n", len(*refs))
		}
	}

	return buf.String()
}

func genBlockTooltip(blk *ssa.BasicBlock) string {
	blockIndexes := func(blks []*ssa.BasicBlock) []int {
		var indexes []int
		for _, blk := range blks {
			indexes = append(indexes, blk.Index)
		}
		return indexes
	}

	var buf bytes.Buffer

	fmt.Fprintf(&buf, "Index: %d\n", blk.Index)
	if blk.Comment != "" {
		fmt.Fprintf(&buf, "Comment: %s\n", blk.Comment)
	}
	fmt.Fprintf(&buf, "Instructions: %d\n", len(blk.Instrs))
	fmt.Fprintf(&buf, "Predecessors: %d, %v\n", len(blk.Preds), blockIndexes(blk.Preds))
	fmt.Fprintf(&buf, "Successors: %d, %v\n", len(blk.Succs), blockIndexes(blk.Succs))
	fmt.Fprintf(&buf, "Dominees: %d, %v\n", len(blk.Dominees()), blockIndexes(blk.Dominees()))
	if idom := blk.Idom(); idom != nil {
		fmt.Fprintf(&buf, "Idom: %d\n", idom.Index)
	}

	return buf.String()
}

// genCFGNodeLabel generates basic block label in Graphviz HTML.
//
// This function borrows a lot from [golang.org/x/tools/go/ssa.WriteFunction].
func genCFGNodeLabel(blk *ssa.BasicBlock) string {
	doc := etree.NewDocument()

	table := doc.CreateElement("table")
	table.CreateAttr("border", "0")
	table.CreateAttr("cellborder", "0")
	table.CreateAttr("cellspacing", "0")
	table.CreateAttr("cellpadding", "2")
	table.CreateAttr("align", "left")

	header := table.CreateElement("tr")
	lheader := header.CreateElement("td")
	lheader.CreateAttr("bgcolor", "5")
	lheader.CreateAttr("heigth", "200")
	lheader.CreateAttr("align", "left")
	lheader = lheader.CreateElement("font")
	lheader.CreateAttr("color", "white")
	lheader.CreateText(fmt.Sprintf("%d:", blk.Index))

	rheader := header.CreateElement("td")
	rheader.CreateAttr("bgcolor", "5")
	rheader.CreateAttr("align", "right")
	rheader.CreateAttr("heigth", "200")
	rheader = rheader.CreateElement("font")
	rheader.CreateAttr("color", "white")
	rheader.CreateText(fmt.Sprintf("%s P:%d S:%d", blk.Comment, len(blk.Preds), len(blk.Succs)))

	for i, instr := range blk.Instrs {
		row := table.CreateElement("tr")
		bgcolor := "2"
		if i%2 == 1 {
			bgcolor = "3"
		}

		lcell := row.CreateElement("td")
		lcell.CreateAttr("bgcolor", bgcolor)
		lcell.CreateAttr("align", "left")
		lcell.CreateAttr("href", "")
		// FIXME: Hwo to create newline in tooltip?
		// https://stackoverflow.com/a/27448551 does not work for me.
		lcell.CreateAttr("tooltip", strings.ReplaceAll(genInstrTooltip(instr), "\n", "    "))
		ltext := lcell.CreateElement("font")

		rcell := row.CreateElement("td")
		rcell.CreateAttr("bgcolor", bgcolor)
		rcell.CreateAttr("align", "right")
		rtext := rcell.CreateElement("font").CreateElement("i")

		switch v := instr.(type) {
		case ssa.Value:

			// Instruction on the left.
			if name := v.Name(); name != "" {
				// TODO: also align the name?
				ltext.CreateText(fmt.Sprintf("%s = ", name))
			}
			ltext.CreateText(instr.String())

			// Type on the right.
			if t := v.Type(); t != nil {
				rtext.CreateText(relType(t, blk.Parent().Pkg.Pkg)) // TODO: shorten the type name
				rcell.CreateAttr("href", "")                       // TODO: link to pkg.go.dev
				rcell.CreateAttr("tooltip", types.TypeString(t, nil))
			}
		case nil:
			// Be robust against bad transforms.
			ltext.CreateElement("s").CreateText("deleted")
			rtext.CreateText(" ")
		default:
			ltext.CreateText(instr.String())
			rtext.CreateText(" ")
		}
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
		Attr("tooltip", genBlockTooltip(blk)).
		Attr("fontname", Fontame).
		Attr("colorscheme", "blues9")

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
		Attr("colorscheme", "blues9").
		Attr("color", "3").
		Attr("fillcolor", "2")
}

func relType(t types.Type, from *types.Package) string {
	s := types.TypeString(t, types.RelativeTo(from))
	s = strings.ReplaceAll(s, "interface{}", "any")
	return s
}

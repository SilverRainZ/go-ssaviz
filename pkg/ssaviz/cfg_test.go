package ssaviz

import (
	"fmt"
	"testing"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/analysistest"
	"golang.org/x/tools/go/analysis/passes/buildssa"
)

var cfgAnalyzer = &analysis.Analyzer{
	Name: "test-build-cfg",
	// TODO: flags
	Requires: []*analysis.Analyzer{buildssa.Analyzer},
	Run: func(pass *analysis.Pass) (any, error) {
		ssaInput := pass.ResultOf[buildssa.Analyzer].(*buildssa.SSA)

		result := []*Graph{}
		for _, ssaFunc := range ssaInput.SrcFuncs {
			g, err := Build(CFG, ssaFunc)
			if err != nil {
				return nil, err
			}
			result = append(result, g)
			switch ssaFunc.Name() {
			case "print":
				// Nothing to do now
			default:
				return nil, fmt.Errorf("unexpected SSA function: %q", ssaFunc.Name())
			}
		}

		if _, err := Render(result); err != nil {
			return nil, err
		}

		return nil, nil
	},
}

func TestBuildCFG(t *testing.T) {
	testdata := analysistest.TestData()

	tests := []string{"a"}
	analysistest.Run(t, testdata, cfgAnalyzer, tests...)
}

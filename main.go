package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sort"
	"syscall"

	"github.com/SilverRainZ/go-ssaviz/pkg/ssaviz"

	"golang.org/x/tools/go/packages"
	"golang.org/x/tools/go/ssa"
	"golang.org/x/tools/go/ssa/ssautil"
)

const (
	defaultPkgPath = "./..."
)

var (
	fset = flag.NewFlagSet(ssaviz.Prog, flag.ExitOnError)

	// Command line flags.
	help bool
	view bool
	out  string
	ver  bool
)

func main() {
	fset.Usage = func() { printUsage(nil) }

	fset.BoolVar(&help, "h", false, "print this help message and exit")
	fset.BoolVar(&ver, "V", false, "print version and exit")
	fset.BoolVar(&ssaviz.Debug, "v", false, "print verbose log")
	fset.StringVar(&out, "o", "ssaviz.html", "HTML report for output")
	fset.BoolVar(&view, "view", false, "view HTML report in system default application")
	fset.Parse(os.Args[1:])

	// Signal handler.
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-sigs
		log.Printf("%s received, exit", sig)
		os.Exit(1)
	}()

	// Parse command line arguments.
	if help {
		printUsage(nil)
	}

	if ver {
		printVersion()
	}

	// Load SSA program.
	var pkgPaths []string
	if args := fset.Args(); len(args) == 0 {
		pkgPaths = []string{defaultPkgPath}
	} else {
		pkgPaths = args
	}
	_, pkgs, err := loadSSA(pkgPaths)
	if err != nil {
		log.Fatalf("failed to load SSA program of %s: %s", pkgPaths, err)
	}

	// Build graph and HTML.
	var graphs []*ssaviz.Graph
	for _, pkg := range pkgs {
		for _, member := range pkg.Members {
			f, ok := member.(*ssa.Function)
			if !ok {
				continue
			}
			g, err := ssaviz.Build(ssaviz.CFG, f)
			if err != nil {
				log.Printf("failed to build graph: %s", err)
			}
			graphs = append(graphs, g)
		}
	}

	// Sort graphs by name.
	sort.Slice(graphs, func(i, j int) bool {
		return graphs[i].Name < graphs[j].Name
	})

	html, err := ssaviz.Render(graphs)
	if err != nil {
		log.Fatalf("failed to build report: %s", err)
	}
	if err := html.Save(out); err != nil {
		log.Fatalf("failed to save report: %s", err)
	}

	if view {
		if err := html.View(); err != nil {
			log.Fatalf("failed to view report: %s", err)
		}
	}
}

func printUsage(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %s\n", err)
	}

	fmt.Fprintf(os.Stderr, "USAGE: %s [OPTIONS] [PKGPATH]…\n", ssaviz.Prog)
	fmt.Fprintln(os.Stderr, ssaviz.Desc+".")
	fmt.Fprintln(os.Stderr)

	fmt.Fprintf(os.Stderr, "PKGPATH:\n")
	fmt.Fprintf(os.Stderr, "\tpath of go packages, default: %q\n", defaultPkgPath)

	fmt.Fprintf(os.Stderr, "OPTIONS:\n")
	fset.PrintDefaults()

	if err != nil {
		os.Exit(2)
	}
	os.Exit(0)
}

func printVersion() {
	fmt.Println(ssaviz.Prog, ssaviz.Version)
	fmt.Println(ssaviz.Desc + ".")
	fmt.Println()
	fmt.Println("Author:", ssaviz.Author)
	os.Exit(0)
}

func loadSSA(pkgPaths []string) (*ssa.Program, []*ssa.Package, error) {
	// Load packages.
	cfg := &packages.Config{
		// Required by ssautil.AllPackages.
		Mode: packages.LoadAllSyntax,
	}
	pkgs, err := packages.Load(cfg, pkgPaths...)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to load packages from %q: %w", pkgPaths, err)
	}

	// Skip packages with errors.
	var ptr int
	for _, pkg := range pkgs {
		for _, err := range pkg.Errors {
			log.Printf("skip pkg %s due to: %s", pkg, err)
		}
		if len(pkg.Errors) == 0 {
			if ssaviz.Debug {
				log.Println("load pkg:", pkg.PkgPath)
			}
			pkgs[ptr] = pkg
			ptr++
		}
	}
	pkgs = pkgs[:ptr]

	ssaProg, ssaPkgs := ssautil.AllPackages(pkgs, 0)
	ssaProg.Build()

	return ssaProg, ssaPkgs, nil
}

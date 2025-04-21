package osexitanalyzer

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
)

// OsExitAnalyzer reports usage of os.Exit() in main() function of package main.
var OsExitAnalyzer = &analysis.Analyzer{
	Name: "osexit",
	Doc:  "check for usage of os.Exit() in main() function of package main",
	Run:  run,
}

func run(pass *analysis.Pass) (interface{}, error) {
	isOSExit := func(x *ast.CallExpr) bool {
		if sel, ok := x.Fun.(*ast.SelectorExpr); ok {
			if id, ok := sel.X.(*ast.Ident); ok && id.Name == "os" && sel.Sel.Name == "Exit" {
				return true
			}
		}
		return false
	}
	bodyFunc := func(x *ast.BlockStmt) {
		for _, s := range x.List {
			if expr, ok := s.(*ast.ExprStmt); ok {
				if x, ok := expr.X.(*ast.CallExpr); ok {
					if isOSExit(x) {
						pass.Reportf(x.Pos(), "usage of exit in main is not recommended")
					}
				}
			}
		}
	}
	for _, file := range pass.Files {
		ast.Inspect(file, func(node ast.Node) bool {
			if file.Name.Name == "main" {
				for _, decl := range file.Decls {
					if x, ok := decl.(*ast.FuncDecl); ok {
						if x.Name.Name == "main" {
							bodyFunc(x.Body)
						}
					}
				}
			}

			return true
		})
	}
	return nil, nil
}

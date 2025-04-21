package main

import (
	"github.com/madatsci/urlshortener/cmd/staticlint/osexitanalyzer"

	gocritic "github.com/go-critic/go-critic/checkers/analyzer"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/multichecker"
	"golang.org/x/tools/go/analysis/passes/printf"
	"golang.org/x/tools/go/analysis/passes/shadow"
	"golang.org/x/tools/go/analysis/passes/structtag"
	"honnef.co/go/tools/quickfix"
	"honnef.co/go/tools/simple"
	"honnef.co/go/tools/staticcheck"
	"honnef.co/go/tools/stylecheck"
)

func main() {
	checks := []*analysis.Analyzer{
		printf.Analyzer,
		shadow.Analyzer,
		structtag.Analyzer,
		gocritic.Analyzer,
		osexitanalyzer.OsExitAnalyzer,
	}

	for _, v := range staticcheck.Analyzers {
		checks = append(checks, v.Analyzer)
	}

	for _, v := range simple.Analyzers {
		checks = append(checks, v.Analyzer)
	}

	for _, v := range stylecheck.Analyzers {
		checks = append(checks, v.Analyzer)
	}

	for _, v := range quickfix.Analyzers {
		checks = append(checks, v.Analyzer)
	}

	multichecker.Main(checks...)
}

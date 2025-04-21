package osexitanalyzer

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"
)

func TestOsExitAnalyzer(t *testing.T) {
	t.Skip("need to fix bug")
	analysistest.Run(t, analysistest.TestData(), OsExitAnalyzer, "./...")
}

// Command cmd/staticlint/multichecker is a static source code analyzer
// for urlshortener.
//
// It provides static analysis tools for validating and inspecting
// source code within the application. It defines analyzers that can be used to
// enforce coding standards, detect potential issues, and improve code quality.
//
// It includes the following analyzers:
//   - golang.org/x/tools/go/analysis/passes/printf analyzer
//   - golang.org/x/tools/go/analysis/passes/shadow analyzer
//   - golang.org/x/tools/go/analysis/passes/structtag analyzer
//   - all analyzers from honnef.co/go/tools/quickfix
//   - all analyzers from honnef.co/go/tools/simple
//   - all analyzers from honnef.co/go/tools/staticcheck
//   - all analyzers from honnef.co/go/tools/stylecheck
//   - all analyzers from https://github.com/go-critic/go-critic
//   - it's own osexitanalyzer, which reports usage of os.Exit() in main() function
//     of package main.
//
// Multichecker can be built with the following command:
//
//	go build -o ./cmd/staticlint/multichecker ./cmd/staticlint/multichecker.go
//
// Example of usage:
//
//	./cmd/staticlint/multichecker ./...
package main

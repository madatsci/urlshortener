// Package main is a test data for osexitanalyzer.
package main

import (
	"fmt"
	"os"
)

func main() {
	fmt.Println("Hello")
	os.Exit(0) // want `usage of exit in main is not recommended`
}

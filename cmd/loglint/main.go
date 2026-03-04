package main

import (
	"log"

	"github.com/andreborch/log-linter/pkg/analyzer"
	"golang.org/x/tools/go/analysis/singlechecker"
)

func main() {
	plug, err := analyzer.New(nil)
	if err != nil {
		log.Fatal(err)
	}

	analyzers, err := plug.BuildAnalyzers()
	if err != nil {
		log.Fatal(err)
	}

	for _, analyzer := range analyzers {
		singlechecker.Main(analyzer)
	}
}

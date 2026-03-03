package utils

import (
	"github.com/andreborch/log-linter/pkg"
	"golang.org/x/tools/go/analysis"
)

func SendReports(reports *[]pkg.Report, pass *analysis.Pass, category string) {
	for _, rep := range *reports {
		pass.Report(analysis.Diagnostic{
			Pos:            rep.Pos,
			Category:       category,
			Message:        rep.Message,
			SuggestedFixes: nil,
		})
	}
}

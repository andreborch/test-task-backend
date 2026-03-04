package rules

import (
	"go/ast"
	"go/token"
	"unicode"

	"github.com/andreborch/log-linter/internal/utils"
	"github.com/andreborch/log-linter/pkg"
)

// isLetterLower reports whether r is either not a letter or a lowercase letter.
// Non-letter runes are treated as valid to avoid false positives on punctuation,
// digits, and other symbols at the beginning of a log message.
func isLetterLower(r rune) bool {
	if !unicode.IsLetter(r) {
		return true
	}

	return unicode.IsLower(r)
}

// checkStringLowercase validates that the first character inside a string literal
// starts with a lowercase letter. If it does not, it appends a report entry with
// the position, literal length, and a diagnostic message.
func checkStringLowercase(basicLit *ast.BasicLit) []pkg.Report {
	reports := []pkg.Report{}
	data := basicLit.Value
	symb := []rune(data)[1] // skip quotes
	if !isLetterLower(symb) {
		reports = append(reports, pkg.Report{
			Pos:     basicLit.Pos() + token.Pos(1),
			Length:  len(data),
			Message: "Log message should start with lowercase",
		})
	}
	return reports
}

// CheckLowerCase inspects the first logging argument and verifies lowercase
// message style for string literals.
//
// It supports:
//   - direct string literals, e.g. "message"
//   - concatenated string expressions, e.g. "message " + value
//
// For binary expressions, only the first flattened part is checked. Non-string
// arguments are ignored.
func CheckLowerCase(args []ast.Expr, reports *[]pkg.Report) {
	arg := args[0]
	if binExpr, ok := arg.(*ast.BinaryExpr); ok {
		parts := utils.FlattenBinaryExpr(binExpr)
		part := parts[0]
		basicPart, ok := part.(*ast.BasicLit)
		if !ok || basicPart.Kind != token.STRING { // skip eveything non string
			return
		}
		lower_reports := checkStringLowercase(basicPart)
		*reports = append(*reports, lower_reports...)
	} else if basicLit, ok := arg.(*ast.BasicLit); ok && basicLit.Kind == token.STRING {
		lower_reports := checkStringLowercase(basicLit)
		*reports = append(*reports, lower_reports...)
	} else if fun, ok := arg.(*ast.CallExpr); ok {
		CheckLowerCase(fun.Args, reports)
	}
}

package rules

import (
	"go/ast"
	"go/token"
	"strings"
	"unicode"

	"github.com/andreborch/log-linter/internal/utils"
	"github.com/andreborch/log-linter/pkg"
)

// isSpecialChar reports whether r should be treated as a disallowed special character.
//
// A rune is considered NOT special (returns false) if it is:
//   - a Unicode letter,
//   - a Unicode digit,
//   - a Unicode whitespace character,
//   - explicitly listed in exceptions.
//
// Any other rune (including punctuation/symbols/emoji not in exceptions)
// is treated as special (returns true).
//
// exceptions is interpreted as a plain string of allowed runes.
func isSpecialChar(r rune, exceptions string) bool {
	if unicode.IsLetter(r) { // skip non letters
		return false
	}

	if unicode.IsSpace(r) {
		return false
	}

	if unicode.IsDigit(r) {
		return false
	}

	if strings.Contains(exceptions, string(r)) {
		return false
	}

	return true
}

// checkStringSpecials analyzes a string literal and reports the first disallowed rune.
//
// It removes surrounding quotes from basicLit.Value, iterates through the literal's runes,
// and checks each rune with isSpecialChar using the provided exceptions set.
//
// When a disallowed special character (including emoji not in exceptions) is found,
// the function appends a single pkg.Report to report with the rune position in source code
// and the message: "Log message shouldn't contain special chars or emoji".
// Scanning stops after the first violation.
func checkStringSpecials(basicLit *ast.BasicLit, exceptions string, report *[]pkg.Report) {
	data := basicLit.Value
	data = data[1 : len(data)-1]

	for idx, symb := range data {
		if isSpecialChar(symb, exceptions) {
			*report = append(*report, pkg.Report{
				Pos:     basicLit.Pos() + token.Pos(idx+1),
				Length:  0,
				Message: "Log message shouldn't contain special chars or emoji",
			})
			break
		}
	}
}

// HasSpecialChar inspects argument expressions and reports string literals
// containing special characters, excluding any runes listed in exceptions.
//
// The function handles:
//   - direct string literals (e.g., "text"),
//   - concatenated string expressions (binary expressions), which are flattened
//     and checked part by part.
//
// Non-string expression parts are ignored.
// Any findings are appended to reports via checkStringSpecials.
func HasSpecialChar(args []ast.Expr, reports *[]pkg.Report, exceptions string) {
	for _, arg := range args {
		if binExpr, ok := arg.(*ast.BinaryExpr); ok {
			parts := utils.FlattenBinaryExpr(binExpr)
			for _, part := range parts {
				basicPart, ok := part.(*ast.BasicLit)
				if !ok || basicPart.Kind != token.STRING { // skip eveything non string
					continue
				}
				checkStringSpecials(basicPart, exceptions, reports)
			}
		} else if basicLit, ok := arg.(*ast.BasicLit); ok && basicLit.Kind == token.STRING {
			checkStringSpecials(basicLit, exceptions, reports)
		} else if fun, ok := arg.(*ast.CallExpr); ok {
			HasSpecialChar(fun.Args, reports, exceptions)
		}
	}
}

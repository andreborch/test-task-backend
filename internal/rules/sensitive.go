package rules

import (
	"go/ast"
	"go/token"
	"regexp"
	"strings"

	"github.com/andreborch/log-linter/internal/utils"
	"github.com/andreborch/log-linter/pkg"
)

// argHasSensitive checks whether the provided string contains sensitive data.
//
// Detection order:
//  1. Built-in sensitive keywords from pkg.DefaultSensBans(), combined with
//     additional separators (":", "=", "is"), unless explicitly listed in exceptions.
//  2. Built-in token regex patterns from pkg.DefaultTokensPatterns().
//  3. Custom blocked substrings passed via blocked, unless listed in exceptions.
//
// It returns:
//   - has: true when sensitive content is found
//   - idx: zero-based byte index of the first match in data (or -1 if not found)
//   - length: matched fragment length (or -1 if not found)
//
// Note: matching is based on simple substring search (for keyword checks) and regex
// search (for token patterns).

// checkStringSensitive inspects a string literal AST node for sensitive content.
//
// If sensitive data is found, it appends a pkg.Report entry to report with:
//   - Pos set to the literal position offset by the match index
//   - Length set to the matched fragment length
//   - Message set to "Sensitive data detected"

// HasSensitiveData scans logging argument expressions for sensitive string content.
//
// It supports:
//   - direct string literals (*ast.BasicLit with token.STRING)
//   - concatenated string expressions (*ast.BinaryExpr), flattened via
//     utils.FlattenBinaryExpr and checked part-by-part.
//
// For each detected match, it appends a pkg.Report to reports.
func argHasSensitive(data string, blocked []string, exceptions []string) (has bool, idx int, length int) {
	default_additional := []string{":", "=", "is"}

	for _, block := range pkg.DefaultSensBans() {
		for _, add := range default_additional {
			block_additional := block + add
			if utils.Contains(exceptions, block_additional) {
				continue
			}
			idx = strings.Index(data, block_additional)
			if idx != -1 {
				return true, idx, len(block)
			}
		}
	}

	for _, pattern := range pkg.DefaultTokensPatterns() {
		re := regexp.MustCompile(pattern)
		loc := re.FindStringIndex(data)
		if loc != nil {
			return true, loc[0], loc[1] - loc[0]
		}
	}

	for _, block := range blocked {
		if utils.Contains(exceptions, block) {
			continue
		}
		idx = strings.Index(data, block)
		if idx != -1 {
			return true, idx, len(block)
		}
	}
	return false, -1, -1
}

// checkStringSensitive examines a string literal node for sensitive content.
//
// It evaluates the literal value with argHasSensitive using built-in rules,
// custom blocked substrings, and exception entries.
//
// When a match is found, it appends a pkg.Report to report with:
//   - Pos: literal start position shifted by the match byte index
//   - Length: matched fragment length
//   - Message: "Sensitive data detected"
func checkStringSensitive(basicLit *ast.BasicLit, blocked []string, exceptions []string, report *[]pkg.Report) {
	data := basicLit.Value
	has, idx, length := argHasSensitive(data, blocked, exceptions)
	if has {
		*report = append(*report, pkg.Report{
			Pos:     basicLit.Pos() + token.Pos(idx),
			Length:  length,
			Message: "Sensitive data detected",
		})
	}
}

// HasSensitiveData scans logging call arguments and reports sensitive string content.
//
// Supported argument forms:
//   - string literals (*ast.BasicLit with token.STRING)
//   - concatenated string expressions (*ast.BinaryExpr), flattened with
//     utils.FlattenBinaryExpr and checked literal-by-literal.
//
// Detection uses argHasSensitive, which applies:
//  1. default sensitive keyword bans (with built-in separators),
//  2. default token regex patterns,
//  3. custom blocked substrings,
//
// while skipping any entries listed in exceptions.
//
// For every detected match, a pkg.Report is appended to reports with:
//   - Pos: absolute AST position of the matched fragment,
//   - Length: matched fragment length in bytes,
//   - Message: "Sensitive data detected".
//
// Non-string expressions are ignored.
func HasSensitiveData(args []ast.Expr, reports *[]pkg.Report, blocked []string, exceptions []string) {
	for _, arg := range args {
		if binExpr, ok := arg.(*ast.BinaryExpr); ok {
			parts := utils.FlattenBinaryExpr(binExpr)
			for _, part := range parts {
				basicPart, ok := part.(*ast.BasicLit)

				if !ok || basicPart.Kind != token.STRING { // skip eveything non string
					continue
				}
				checkStringSensitive(basicPart, blocked, exceptions, reports)
			}
		} else if basicLit, ok := arg.(*ast.BasicLit); ok && basicLit.Kind == token.STRING {
			checkStringSensitive(basicLit, blocked, exceptions, reports)
		}
	}
}

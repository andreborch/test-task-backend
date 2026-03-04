package rules

import (
	"go/ast"
	"go/token"
	"regexp"
	"strings"

	"github.com/andreborch/log-linter/internal/utils"
	"github.com/andreborch/log-linter/pkg"
)

// argHasSensitive checks whether a given string contains sensitive data.
//
// It performs detection in three sequential stages:
//  1. Default sensitive keyword bans: iterates over pkg.DefaultSensBans() and
//     checks for each keyword followed by built-in separators (":", "=", "is", "-").
//     Also detects keywords appearing at the end of the string.
//  2. Default token regex patterns: matches against pkg.DefaultTokensPatterns()
//     to catch common secret/token formats (e.g. API keys, JWTs).
//  3. Custom blocked substrings: checks each entry in blocked for an exact
//     substring match.
//
// Entries listed in exceptions are skipped during stages 1 and 3.
//
// Parameters:
//   - data:       the input string to inspect (compared case-insensitively).
//   - blocked:    additional substrings to treat as sensitive.
//   - exceptions: substrings to exclude from detection.
//
// Returns:
//   - has:    true if a sensitive fragment was found.
//   - idx:    byte offset of the match within the lowercased data, or -1.
//   - length: byte length of the matched keyword/pattern, or -1.
func argHasSensitive(data string, blocked []string, exceptions []string) (has bool, idx int, length int) {
	default_additional := []string{":", "=", "is", "-"}
	data_lower := strings.ToLower(data)

	for _, block := range pkg.DefaultSensBans() {
		for _, add := range default_additional {
			if utils.Contains(exceptions, block) {
				break
			}
			block_additional := block + add
			block_spaced := block + " " + add

			if utils.Contains(exceptions, block_additional) {
				break
			}

			if utils.Contains(exceptions, block_spaced) {
				break
			}

			idx = strings.Index(data_lower, block_additional)
			if idx != -1 {
				return true, idx, len(block)
			}

			idx = strings.Index(data_lower, block_spaced)
			if idx != -1 {
				return true, idx, len(block)
			}

			trimmed := strings.TrimSpace(data_lower)
			idx = strings.Index(trimmed, block) // checking if end of string, possible sensitive data after that arg
			if idx != -1 && idx+1+len(block) == len(trimmed) {
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
		idx = strings.Index(data_lower, block)
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
func checkStringSensitive(basicLit *ast.BasicLit, blocked []string, exceptions []string) []pkg.Report {
	report := []pkg.Report{}

	data := basicLit.Value
	has, idx, length := argHasSensitive(data, blocked, exceptions)
	if has {
		report = append(report, pkg.Report{
			Pos:     basicLit.Pos() + token.Pos(idx),
			Length:  length,
			Message: "Sensitive data detected",
		})
	}
	return report
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
				sens_reports := checkStringSensitive(basicPart, blocked, exceptions)
				*reports = append(*reports, sens_reports...)
			}
		} else if basicLit, ok := arg.(*ast.BasicLit); ok && basicLit.Kind == token.STRING {
			sens_reports := checkStringSensitive(basicLit, blocked, exceptions)
			*reports = append(*reports, sens_reports...)
		} else if fun, ok := arg.(*ast.CallExpr); ok {
			HasSensitiveData(fun.Args, reports, blocked, exceptions)
		}
	}
}

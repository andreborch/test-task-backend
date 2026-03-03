package rules

import (
	"go/ast"
	"go/token"
	"strings"
	"unicode"

	"github.com/andreborch/log-linter/internal/utils"
	"github.com/andreborch/log-linter/pkg"
)

// isLetterFromLanguage reports whether r is allowed for the language identified by langCode.
//
// Non-letter runes are treated as valid and ignored by language checks.
// Supported language codes are mapped to Unicode script ranges
// (for example: "ru" -> Cyrillic, "zh" -> Han, "en"/"de"/... -> Latin).
// If langCode is unknown, the function returns false for letters.
func isLetterFromLanguage(r rune, lang string) bool {
	if !unicode.IsLetter(r) { // skip non letters
		return true
	}

	switch lang {
	case "ru":
		return unicode.In(r, unicode.Cyrillic)
	case "zh":
		return unicode.In(r, unicode.Han)
	case "ja":
		return unicode.In(r, unicode.Hiragana, unicode.Katakana, unicode.Han)
	case "ko":
		return unicode.In(r, unicode.Hangul)
	case "ar":
		return unicode.In(r, unicode.Arabic)
	case "he":
		return unicode.In(r, unicode.Hebrew)
	case "el":
		return unicode.In(r, unicode.Greek)
	case "th":
		return unicode.In(r, unicode.Thai)
	case "ka":
		return unicode.In(r, unicode.Georgian)
	case "hy":
		return unicode.In(r, unicode.Armenian)
	case "de", "fr", "es", "it", "pt", "nl", "sv", "no", "da", "fi", "pl", "cs", "sk", "ro", "hu", "tr", "en":
		return unicode.In(r, unicode.Latin)
	default:
		return false
	}
}

// checkStringLang validates that all letters in the string literal basicLit
// belong to the script expected by lang.
//
// On the first invalid letter, it appends a pkg.Report entry with the literal
// position, its raw length, and a message indicating the required language.
func checkStringLang(basicLit *ast.BasicLit, lang string, report *[]pkg.Report) {
	data := basicLit.Value
	for _, symb := range data {
		if !isLetterFromLanguage(symb, lang) {
			*report = append(*report, pkg.Report{
				Pos:     basicLit.Pos(),
				Length:  len(data),
				Message: "Log message language must be " + strings.ToUpper(lang),
			})
			break
		}
	}
}

// LangIsCorrect checks string arguments in args against the expected language lang
// and appends violations to reports.
//
// It validates:
//   - direct string literals
//   - string literals inside binary concatenation expressions
//
// Non-string expressions are ignored.
func LangIsCorrect(args []ast.Expr, reports *[]pkg.Report, lang string) {
	for _, arg := range args {
		if binExpr, ok := arg.(*ast.BinaryExpr); ok {
			parts := utils.FlattenBinaryExpr(binExpr)
			for _, part := range parts {
				basicPart, ok := part.(*ast.BasicLit)
				if !ok || basicPart.Kind != token.STRING { // skip eveything non string
					continue
				}
				checkStringLang(basicPart, lang, reports)
			}
		} else if basicLit, ok := arg.(*ast.BasicLit); ok && basicLit.Kind == token.STRING {
			checkStringLang(basicLit, lang, reports)
		}
	}
}

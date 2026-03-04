package rules

import (
	"go/ast"
	"go/token"
	"testing"

	"github.com/andreborch/log-linter/pkg"
)

func TestIsLetterFromLanguage(t *testing.T) {
	tests := []struct {
		name     string
		r        rune
		lang     string
		expected bool
	}{
		// Non-letter runes should always return true
		{"non-letter digit", '1', "en", true},
		{"non-letter space", ' ', "en", true},
		{"non-letter punctuation", '.', "ru", true},
		{"non-letter symbol", '@', "zh", true},

		// English / Latin
		{"en latin lowercase", 'a', "en", true},
		{"en latin uppercase", 'Z', "en", true},
		{"en cyrillic letter", 'я', "en", false},
		{"en chinese char", '中', "en", false},

		// German (Latin)
		{"de umlaut", 'ü', "de", true},
		{"de latin", 'a', "de", true},
		{"de cyrillic", 'д', "de", false},

		// French (Latin)
		{"fr accented", 'é', "fr", true},
		{"fr latin", 'b', "fr", true},

		// Spanish (Latin)
		{"es ñ", 'ñ', "es", true},

		// Italian (Latin)
		{"it latin", 'c', "it", true},

		// Portuguese (Latin)
		{"pt latin", 'ã', "pt", true},

		// Dutch (Latin)
		{"nl latin", 'd', "nl", true},

		// Swedish (Latin)
		{"sv latin å", 'å', "sv", true},

		// Norwegian (Latin)
		{"no latin ø", 'ø', "no", true},

		// Danish (Latin)
		{"da latin æ", 'æ', "da", true},

		// Finnish (Latin)
		{"fi latin ä", 'ä', "fi", true},

		// Polish (Latin)
		{"pl latin ł", 'ł', "pl", true},

		// Czech (Latin)
		{"cs latin č", 'č', "cs", true},

		// Slovak (Latin)
		{"sk latin ľ", 'ľ', "sk", true},

		// Romanian (Latin)
		{"ro latin ș", 'ș', "ro", true},

		// Hungarian (Latin)
		{"hu latin ő", 'ő', "hu", true},

		// Turkish (Latin)
		{"tr latin ğ", 'ğ', "tr", true},

		// Russian
		{"ru cyrillic lowercase", 'а', "ru", true},
		{"ru cyrillic uppercase", 'Я', "ru", true},
		{"ru latin letter", 'a', "ru", false},
		{"ru chinese char", '中', "ru", false},

		// Chinese
		{"zh han char", '中', "zh", true},
		{"zh latin letter", 'a', "zh", false},

		// Japanese
		{"ja hiragana", 'あ', "ja", true},
		{"ja katakana", 'ア', "ja", true},
		{"ja kanji", '漢', "ja", true},
		{"ja latin", 'a', "ja", false},

		// Korean
		{"ko hangul", '한', "ko", true},
		{"ko latin", 'a', "ko", false},

		// Arabic
		{"ar arabic letter", 'ع', "ar", true},
		{"ar latin", 'a', "ar", false},

		// Hebrew
		{"he hebrew letter", 'א', "he", true},
		{"he latin", 'a', "he", false},

		// Greek
		{"el greek letter", 'α', "el", true},
		{"el greek uppercase", 'Ω', "el", true},
		{"el latin", 'a', "el", false},

		// Thai
		{"th thai letter", 'ก', "th", true},
		{"th latin", 'a', "th", false},

		// Georgian
		{"ka georgian letter", 'ქ', "ka", true},
		{"ka latin", 'a', "ka", false},

		// Armenian
		{"hy armenian letter", 'ա', "hy", true},
		{"hy latin", 'a', "hy", false},

		// Unknown language
		{"unknown lang letter", 'a', "xx", false},
		{"unknown lang non-letter", '1', "xx", true},
		{"empty lang", 'a', "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isLetterFromLanguage(tt.r, tt.lang)
			if result != tt.expected {
				t.Errorf("isLetterFromLanguage(%q, %q) = %v, want %v", tt.r, tt.lang, result, tt.expected)
			}
		})
	}
}

func TestCheckStringLang(t *testing.T) {
	tests := []struct {
		name          string
		value         string
		lang          string
		expectReport  bool
		expectMessage string
	}{
		{"en valid", `"hello world"`, "en", false, ""},
		{"en with numbers", `"hello 123"`, "en", false, ""},
		{"en with cyrillic", `"hello мир"`, "en", true, "Log message language must be EN"},
		{"ru valid", `"привет мир"`, "ru", false, ""},
		{"ru with latin", `"привет world"`, "ru", true, "Log message language must be RU"},
		{"zh valid", `"你好世界"`, "zh", false, ""},
		{"zh with latin", `"你好world"`, "zh", true, "Log message language must be ZH"},
		{"empty string", `""`, "en", false, ""},
		{"only digits", `"12345"`, "en", false, ""},
		{"only punctuation", `"!@#$%"`, "en", false, ""},
		{"el valid", `"αβγδ"`, "el", false, ""},
		{"el with latin", `"αβγδabc"`, "el", true, "Log message language must be EL"},
		{"ko valid", `"한글테스트"`, "ko", false, ""},
		{"ja mixed valid", `"あいうアイウ漢字"`, "ja", false, ""},
		{"ja with latin", `"あいうabc"`, "ja", true, "Log message language must be JA"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lit := &ast.BasicLit{
				ValuePos: token.Pos(1),
				Kind:     token.STRING,
				Value:    tt.value,
			}
			var reports []pkg.Report
			checkStringLang(lit, tt.lang, &reports)

			if tt.expectReport {
				if len(reports) != 1 {
					t.Fatalf("expected 1 report, got %d", len(reports))
				}
				if reports[0].Message != tt.expectMessage {
					t.Errorf("expected message %q, got %q", tt.expectMessage, reports[0].Message)
				}
				if reports[0].Length != len(tt.value) {
					t.Errorf("expected length %d, got %d", len(tt.value), reports[0].Length)
				}
				if reports[0].Pos != token.Pos(1) {
					t.Errorf("expected pos 1, got %d", reports[0].Pos)
				}
			} else {
				if len(reports) != 0 {
					t.Fatalf("expected no reports, got %d: %+v", len(reports), reports)
				}
			}
		})
	}
}

func TestCheckStringLangOnlyOneReport(t *testing.T) {
	// Even with multiple invalid characters, only one report should be generated
	lit := &ast.BasicLit{
		ValuePos: token.Pos(10),
		Kind:     token.STRING,
		Value:    `"абвгд"`,
	}
	var reports []pkg.Report
	checkStringLang(lit, "en", &reports)
	if len(reports) != 1 {
		t.Errorf("expected exactly 1 report for multiple invalid chars, got %d", len(reports))
	}
}

func TestLangIsCorrect_BasicLit(t *testing.T) {
	// Valid English string literal
	t.Run("valid string literal", func(t *testing.T) {
		args := []ast.Expr{
			&ast.BasicLit{ValuePos: 1, Kind: token.STRING, Value: `"hello"`},
		}
		var reports []pkg.Report
		LangIsCorrect(args, &reports, "en")
		if len(reports) != 0 {
			t.Errorf("expected no reports, got %d", len(reports))
		}
	})

	// Invalid: Russian text with English lang
	t.Run("invalid string literal", func(t *testing.T) {
		args := []ast.Expr{
			&ast.BasicLit{ValuePos: 1, Kind: token.STRING, Value: `"привет"`},
		}
		var reports []pkg.Report
		LangIsCorrect(args, &reports, "en")
		if len(reports) != 1 {
			t.Errorf("expected 1 report, got %d", len(reports))
		}
	})

	// INT literal should be skipped
	t.Run("non-string literal ignored", func(t *testing.T) {
		args := []ast.Expr{
			&ast.BasicLit{ValuePos: 1, Kind: token.INT, Value: "42"},
		}
		var reports []pkg.Report
		LangIsCorrect(args, &reports, "en")
		if len(reports) != 0 {
			t.Errorf("expected no reports for non-string literal, got %d", len(reports))
		}
	})
}

func TestLangIsCorrect_BinaryExpr(t *testing.T) {
	// Binary expression: "hello" + "world"
	t.Run("valid binary concat", func(t *testing.T) {
		binExpr := &ast.BinaryExpr{
			X:  &ast.BasicLit{ValuePos: 1, Kind: token.STRING, Value: `"hello"`},
			Op: token.ADD,
			Y:  &ast.BasicLit{ValuePos: 10, Kind: token.STRING, Value: `"world"`},
		}
		args := []ast.Expr{binExpr}
		var reports []pkg.Report
		LangIsCorrect(args, &reports, "en")
		if len(reports) != 0 {
			t.Errorf("expected no reports, got %d", len(reports))
		}
	})

	// Binary expression with one invalid part
	t.Run("invalid binary concat", func(t *testing.T) {
		binExpr := &ast.BinaryExpr{
			X:  &ast.BasicLit{ValuePos: 1, Kind: token.STRING, Value: `"hello"`},
			Op: token.ADD,
			Y:  &ast.BasicLit{ValuePos: 10, Kind: token.STRING, Value: `"мир"`},
		}
		args := []ast.Expr{binExpr}
		var reports []pkg.Report
		LangIsCorrect(args, &reports, "en")
		if len(reports) != 1 {
			t.Errorf("expected 1 report, got %d", len(reports))
		}
	})

	// Binary expression with non-string parts (e.g., identifier)
	t.Run("binary expr with non-string part", func(t *testing.T) {
		binExpr := &ast.BinaryExpr{
			X:  &ast.BasicLit{ValuePos: 1, Kind: token.STRING, Value: `"hello"`},
			Op: token.ADD,
			Y:  &ast.Ident{Name: "someVar"},
		}
		args := []ast.Expr{binExpr}
		var reports []pkg.Report
		LangIsCorrect(args, &reports, "en")
		if len(reports) != 0 {
			t.Errorf("expected no reports, got %d", len(reports))
		}
	})

	// Binary expression with INT literal (should be skipped)
	t.Run("binary expr with int literal", func(t *testing.T) {
		binExpr := &ast.BinaryExpr{
			X:  &ast.BasicLit{ValuePos: 1, Kind: token.STRING, Value: `"hello"`},
			Op: token.ADD,
			Y:  &ast.BasicLit{ValuePos: 10, Kind: token.INT, Value: "42"},
		}
		args := []ast.Expr{binExpr}
		var reports []pkg.Report
		LangIsCorrect(args, &reports, "en")
		if len(reports) != 0 {
			t.Errorf("expected no reports, got %d", len(reports))
		}
	})
}

func TestLangIsCorrect_CallExpr(t *testing.T) {
	// CallExpr with string argument
	t.Run("valid call expr arg", func(t *testing.T) {
		callExpr := &ast.CallExpr{
			Fun: &ast.Ident{Name: "fmt.Sprintf"},
			Args: []ast.Expr{
				&ast.BasicLit{ValuePos: 1, Kind: token.STRING, Value: `"hello %s"`},
			},
		}
		args := []ast.Expr{callExpr}
		var reports []pkg.Report
		LangIsCorrect(args, &reports, "en")
		if len(reports) != 0 {
			t.Errorf("expected no reports, got %d", len(reports))
		}
	})

	// CallExpr with invalid language string argument
	t.Run("invalid call expr arg", func(t *testing.T) {
		callExpr := &ast.CallExpr{
			Fun: &ast.Ident{Name: "fmt.Sprintf"},
			Args: []ast.Expr{
				&ast.BasicLit{ValuePos: 1, Kind: token.STRING, Value: `"привет %s"`},
			},
		}
		args := []ast.Expr{callExpr}
		var reports []pkg.Report
		LangIsCorrect(args, &reports, "en")
		if len(reports) != 1 {
			t.Errorf("expected 1 report, got %d", len(reports))
		}
	})

	// Nested CallExpr
	t.Run("nested call expr", func(t *testing.T) {
		innerCall := &ast.CallExpr{
			Fun: &ast.Ident{Name: "inner"},
			Args: []ast.Expr{
				&ast.BasicLit{ValuePos: 1, Kind: token.STRING, Value: `"мир"`},
			},
		}
		outerCall := &ast.CallExpr{
			Fun:  &ast.Ident{Name: "outer"},
			Args: []ast.Expr{innerCall},
		}
		args := []ast.Expr{outerCall}
		var reports []pkg.Report
		LangIsCorrect(args, &reports, "en")
		if len(reports) != 1 {
			t.Errorf("expected 1 report for nested call, got %d", len(reports))
		}
	})
}

func TestLangIsCorrect_IgnoredExprTypes(t *testing.T) {
	// Ident (variable reference) should be ignored
	t.Run("ident ignored", func(t *testing.T) {
		args := []ast.Expr{
			&ast.Ident{Name: "someVariable"},
		}
		var reports []pkg.Report
		LangIsCorrect(args, &reports, "en")
		if len(reports) != 0 {
			t.Errorf("expected no reports for ident, got %d", len(reports))
		}
	})

	// SelectorExpr should be ignored
	t.Run("selector expr ignored", func(t *testing.T) {
		args := []ast.Expr{
			&ast.SelectorExpr{
				X:   &ast.Ident{Name: "pkg"},
				Sel: &ast.Ident{Name: "Value"},
			},
		}
		var reports []pkg.Report
		LangIsCorrect(args, &reports, "en")
		if len(reports) != 0 {
			t.Errorf("expected no reports for selector expr, got %d", len(reports))
		}
	})
}

func TestLangIsCorrect_EmptyArgs(t *testing.T) {
	var reports []pkg.Report
	LangIsCorrect(nil, &reports, "en")
	if len(reports) != 0 {
		t.Errorf("expected no reports for nil args, got %d", len(reports))
	}

	LangIsCorrect([]ast.Expr{}, &reports, "en")
	if len(reports) != 0 {
		t.Errorf("expected no reports for empty args, got %d", len(reports))
	}
}

func TestLangIsCorrect_MultipleArgs(t *testing.T) {
	// Multiple args, some valid, some invalid
	args := []ast.Expr{
		&ast.BasicLit{ValuePos: 1, Kind: token.STRING, Value: `"hello"`},
		&ast.BasicLit{ValuePos: 10, Kind: token.STRING, Value: `"привет"`},
		&ast.BasicLit{ValuePos: 20, Kind: token.STRING, Value: `"world"`},
		&ast.BasicLit{ValuePos: 30, Kind: token.STRING, Value: `"мир"`},
	}
	var reports []pkg.Report
	LangIsCorrect(args, &reports, "en")
	if len(reports) != 2 {
		t.Errorf("expected 2 reports, got %d", len(reports))
	}
}

func TestLangIsCorrect_MixedArgTypes(t *testing.T) {
	// Mix of BasicLit, BinaryExpr, CallExpr, and Ident
	binExpr := &ast.BinaryExpr{
		X:  &ast.BasicLit{ValuePos: 100, Kind: token.STRING, Value: `"test"`},
		Op: token.ADD,
		Y:  &ast.BasicLit{ValuePos: 110, Kind: token.STRING, Value: `"данные"`},
	}
	callExpr := &ast.CallExpr{
		Fun: &ast.Ident{Name: "fn"},
		Args: []ast.Expr{
			&ast.BasicLit{ValuePos: 200, Kind: token.STRING, Value: `"ошибка"`},
		},
	}
	args := []ast.Expr{
		&ast.BasicLit{ValuePos: 1, Kind: token.STRING, Value: `"ok"`},
		binExpr,
		&ast.Ident{Name: "ignored"},
		callExpr,
	}
	var reports []pkg.Report
	LangIsCorrect(args, &reports, "en")
	if len(reports) != 2 {
		t.Errorf("expected 2 reports (one from binary, one from call), got %d", len(reports))
	}
}

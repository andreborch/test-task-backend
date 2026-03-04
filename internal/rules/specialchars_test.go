package rules

import (
	"go/ast"
	"go/token"
	"testing"

	"github.com/andreborch/log-linter/pkg"
)

func TestIsSpecialChar(t *testing.T) {
	tests := []struct {
		name       string
		r          rune
		exceptions string
		want       bool
	}{
		// Letters — not special
		{"lowercase letter", 'a', "", false},
		{"uppercase letter", 'Z', "", false},
		{"unicode letter cyrillic", 'Я', "", false},
		{"unicode letter chinese", '中', "", false},

		// Digits — not special
		{"digit zero", '0', "", false},
		{"digit nine", '9', "", false},

		// Whitespace — not special
		{"space", ' ', "", false},
		{"tab", '\t', "", false},
		{"newline", '\n', "", false},

		// Punctuation/symbols — special
		{"exclamation mark", '!', "", true},
		{"at sign", '@', "", true},
		{"hash", '#', "", true},
		{"dollar", '$', "", true},
		{"percent", '%', "", true},
		{"caret", '^', "", true},
		{"ampersand", '&', "", true},
		{"asterisk", '*', "", true},
		{"open paren", '(', "", true},
		{"close paren", ')', "", true},
		{"hyphen", '-', "", true},
		{"underscore", '_', "", true},
		{"equals", '=', "", true},
		{"plus", '+', "", true},
		{"period", '.', "", true},
		{"comma", ',', "", true},
		{"colon", ':', "", true},
		{"semicolon", ';', "", true},
		{"slash", '/', "", true},
		{"backslash", '\\', "", true},
		{"pipe", '|', "", true},
		{"question mark", '?', "", true},
		{"single quote", '\'', "", true},
		{"double quote", '"', "", true},
		{"open bracket", '[', "", true},
		{"close bracket", ']', "", true},
		{"open brace", '{', "", true},
		{"close brace", '}', "", true},
		{"tilde", '~', "", true},
		{"backtick", '`', "", true},

		// Emoji — special
		{"emoji smile", '😀', "", true},
		{"emoji heart", '❤', "", true},
		{"emoji thumbs up", '👍', "", true},

		// Exceptions
		{"exclamation with exception", '!', "!", false},
		{"at sign with exception", '@', "@#", false},
		{"hash with exception", '#', "@#", false},
		{"period with exception", '.', ".,:", false},
		{"comma with exception", ',', ".,:", false},
		{"colon with exception", ':', ".,:", false},
		{"emoji with exception", '😀', "😀", false},
		{"dollar not in exceptions", '$', "!@#", true},
		{"hyphen with exception", '-', "-_", false},
		{"underscore with exception", '_', "-_", false},

		// Edge: exception string doesn't contain the rune
		{"asterisk not in exceptions", '*', "!@#$%", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isSpecialChar(tt.r, tt.exceptions)
			if got != tt.want {
				t.Errorf("isSpecialChar(%q, %q) = %v, want %v", tt.r, tt.exceptions, got, tt.want)
			}
		})
	}
}

func TestCheckStringSpecials(t *testing.T) {
	tests := []struct {
		name       string
		value      string
		exceptions string
		wantCount  int
		wantMsg    string
	}{
		{"clean string", `"hello world"`, "", 0, ""},
		{"string with exclamation", `"hello!"`, "", 1, "Log message shouldn't contain special chars or emoji"},
		{"string with period", `"end."`, "", 1, "Log message shouldn't contain special chars or emoji"},
		{"string with exception allowed", `"hello!"`, "!", 0, ""},
		{"string with emoji", `"hi 😀"`, "", 1, "Log message shouldn't contain special chars or emoji"},
		{"string with emoji exception", `"hi 😀"`, "😀", 0, ""},
		{"only letters and digits", `"abc123"`, "", 0, ""},
		{"only spaces", `"   "`, "", 0, ""},
		{"special at start", `"!hello"`, "", 1, "Log message shouldn't contain special chars or emoji"},
		{"multiple specials reports first only", `"!@#"`, "", 1, "Log message shouldn't contain special chars or emoji"},
		{"unicode letters", `"привет мир"`, "", 0, ""},
		{"mixed with allowed exceptions", `"key=value"`, "=", 0, ""},
		{"mixed with partial exceptions", `"key=value;"`, "=", 1, "Log message shouldn't contain special chars or emoji"},
		{"empty string content", `""`, "", 0, ""},
		{"single char special", `"!"`, "", 1, "Log message shouldn't contain special chars or emoji"},
		{"single char letter", `"a"`, "", 0, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lit := &ast.BasicLit{
				ValuePos: 1,
				Kind:     token.STRING,
				Value:    tt.value,
			}
			var reports []pkg.Report
			checkStringSpecials(lit, tt.exceptions, &reports)

			if len(reports) != tt.wantCount {
				t.Errorf("checkStringSpecials() got %d reports, want %d", len(reports), tt.wantCount)
			}
			if tt.wantCount > 0 && len(reports) > 0 {
				if reports[0].Message != tt.wantMsg {
					t.Errorf("checkStringSpecials() message = %q, want %q", reports[0].Message, tt.wantMsg)
				}
			}
		})
	}
}

func TestCheckStringSpecials_Position(t *testing.T) {
	// "a!b" — '!' is at index 1 in the unquoted string, so Pos should be basePos + 1 + 1 = basePos + 2
	lit := &ast.BasicLit{
		ValuePos: 10,
		Kind:     token.STRING,
		Value:    `"a!b"`,
	}
	var reports []pkg.Report
	checkStringSpecials(lit, "", &reports)

	if len(reports) != 1 {
		t.Fatalf("expected 1 report, got %d", len(reports))
	}
	// idx=1 for '!', pos = 10 + 1 + 1 = 12
	expectedPos := token.Pos(12)
	if reports[0].Pos != expectedPos {
		t.Errorf("expected Pos %d, got %d", expectedPos, reports[0].Pos)
	}
}

func TestHasSpecialChar_BasicLit(t *testing.T) {
	tests := []struct {
		name       string
		value      string
		kind       token.Token
		exceptions string
		wantCount  int
	}{
		{"string with special", `"hello!"`, token.STRING, "", 1},
		{"string without special", `"hello"`, token.STRING, "", 0},
		{"int literal ignored", `42`, token.INT, "", 0},
		{"string with exception", `"hello!"`, token.STRING, "!", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			args := []ast.Expr{
				&ast.BasicLit{
					ValuePos: 1,
					Kind:     tt.kind,
					Value:    tt.value,
				},
			}
			var reports []pkg.Report
			HasSpecialChar(args, &reports, tt.exceptions)

			if len(reports) != tt.wantCount {
				t.Errorf("HasSpecialChar() got %d reports, want %d", len(reports), tt.wantCount)
			}
		})
	}
}

func TestHasSpecialChar_BinaryExpr(t *testing.T) {
	// Simulate "hello!" + "world@"
	binExpr := &ast.BinaryExpr{
		X: &ast.BasicLit{
			ValuePos: 1,
			Kind:     token.STRING,
			Value:    `"hello!"`,
		},
		Op: token.ADD,
		Y: &ast.BasicLit{
			ValuePos: 10,
			Kind:     token.STRING,
			Value:    `"world@"`,
		},
	}

	var reports []pkg.Report
	HasSpecialChar([]ast.Expr{binExpr}, &reports, "")

	if len(reports) != 2 {
		t.Errorf("expected 2 reports for binary expr with two special strings, got %d", len(reports))
	}
}

func TestHasSpecialChar_BinaryExprWithNonString(t *testing.T) {
	// "hello!" + <identifier> — identifier part should be skipped
	binExpr := &ast.BinaryExpr{
		X: &ast.BasicLit{
			ValuePos: 1,
			Kind:     token.STRING,
			Value:    `"hello!"`,
		},
		Op: token.ADD,
		Y:  &ast.Ident{Name: "someVar"},
	}

	var reports []pkg.Report
	HasSpecialChar([]ast.Expr{binExpr}, &reports, "")

	if len(reports) != 1 {
		t.Errorf("expected 1 report (non-string part skipped), got %d", len(reports))
	}
}

func TestHasSpecialChar_BinaryExprClean(t *testing.T) {
	binExpr := &ast.BinaryExpr{
		X: &ast.BasicLit{
			ValuePos: 1,
			Kind:     token.STRING,
			Value:    `"hello "`,
		},
		Op: token.ADD,
		Y: &ast.BasicLit{
			ValuePos: 10,
			Kind:     token.STRING,
			Value:    `"world"`,
		},
	}

	var reports []pkg.Report
	HasSpecialChar([]ast.Expr{binExpr}, &reports, "")

	if len(reports) != 0 {
		t.Errorf("expected 0 reports for clean binary expr, got %d", len(reports))
	}
}

func TestHasSpecialChar_CallExpr(t *testing.T) {
	// fmt.Sprintf("hello!") — the inner arg should be inspected
	callExpr := &ast.CallExpr{
		Fun: &ast.Ident{Name: "Sprintf"},
		Args: []ast.Expr{
			&ast.BasicLit{
				ValuePos: 1,
				Kind:     token.STRING,
				Value:    `"hello!"`,
			},
		},
	}

	var reports []pkg.Report
	HasSpecialChar([]ast.Expr{callExpr}, &reports, "")

	if len(reports) != 1 {
		t.Errorf("expected 1 report for call expr with special char, got %d", len(reports))
	}
}

func TestHasSpecialChar_CallExprClean(t *testing.T) {
	callExpr := &ast.CallExpr{
		Fun: &ast.Ident{Name: "Sprintf"},
		Args: []ast.Expr{
			&ast.BasicLit{
				ValuePos: 1,
				Kind:     token.STRING,
				Value:    `"hello world"`,
			},
		},
	}

	var reports []pkg.Report
	HasSpecialChar([]ast.Expr{callExpr}, &reports, "")

	if len(reports) != 0 {
		t.Errorf("expected 0 reports for clean call expr, got %d", len(reports))
	}
}

func TestHasSpecialChar_IdentIgnored(t *testing.T) {
	// A plain identifier should be completely ignored
	args := []ast.Expr{
		&ast.Ident{Name: "someVariable"},
	}

	var reports []pkg.Report
	HasSpecialChar(args, &reports, "")

	if len(reports) != 0 {
		t.Errorf("expected 0 reports for ident, got %d", len(reports))
	}
}

func TestHasSpecialChar_MultipleArgs(t *testing.T) {
	args := []ast.Expr{
		&ast.BasicLit{ValuePos: 1, Kind: token.STRING, Value: `"clean"`},
		&ast.BasicLit{ValuePos: 10, Kind: token.STRING, Value: `"dirty!"`},
		&ast.BasicLit{ValuePos: 20, Kind: token.STRING, Value: `"also dirty@"`},
	}

	var reports []pkg.Report
	HasSpecialChar(args, &reports, "")

	if len(reports) != 2 {
		t.Errorf("expected 2 reports, got %d", len(reports))
	}
}

func TestHasSpecialChar_EmptyArgs(t *testing.T) {
	var reports []pkg.Report
	HasSpecialChar(nil, &reports, "")

	if len(reports) != 0 {
		t.Errorf("expected 0 reports for nil args, got %d", len(reports))
	}

	HasSpecialChar([]ast.Expr{}, &reports, "")
	if len(reports) != 0 {
		t.Errorf("expected 0 reports for empty args, got %d", len(reports))
	}
}

func TestHasSpecialChar_NestedCallExpr(t *testing.T) {
	// Nested: fmt.Sprintf(fmt.Sprintf("inner!"))
	innerCall := &ast.CallExpr{
		Fun: &ast.Ident{Name: "Sprintf"},
		Args: []ast.Expr{
			&ast.BasicLit{
				ValuePos: 1,
				Kind:     token.STRING,
				Value:    `"inner!"`,
			},
		},
	}
	outerCall := &ast.CallExpr{
		Fun:  &ast.Ident{Name: "Sprintf"},
		Args: []ast.Expr{innerCall},
	}

	var reports []pkg.Report
	HasSpecialChar([]ast.Expr{outerCall}, &reports, "")

	if len(reports) != 1 {
		t.Errorf("expected 1 report for nested call expr, got %d", len(reports))
	}
}

func TestHasSpecialChar_BinaryExprWithIntLiteral(t *testing.T) {
	// "hello" + 42 — INT literal should be skipped
	binExpr := &ast.BinaryExpr{
		X: &ast.BasicLit{
			ValuePos: 1,
			Kind:     token.STRING,
			Value:    `"hello"`,
		},
		Op: token.ADD,
		Y: &ast.BasicLit{
			ValuePos: 10,
			Kind:     token.INT,
			Value:    `42`,
		},
	}

	var reports []pkg.Report
	HasSpecialChar([]ast.Expr{binExpr}, &reports, "")

	if len(reports) != 0 {
		t.Errorf("expected 0 reports (INT part skipped), got %d", len(reports))
	}
}

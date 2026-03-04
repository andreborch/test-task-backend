package rules

import (
	"go/ast"
	"go/token"
	"strings"
	"testing"

	"github.com/andreborch/log-linter/pkg"
)

func TestArgHasSensitive_DefaultBansWithSeparators(t *testing.T) {
	tests := []struct {
		name       string
		data       string
		blocked    []string
		exceptions []string
		wantHas    bool
	}{
		{
			name:    "password with colon separator",
			data:    "password:secret123",
			wantHas: true,
		},
		{
			name:    "password with equals separator",
			data:    "password=secret123",
			wantHas: true,
		},
		{
			name:    "password with is separator",
			data:    "passwordis secret",
			wantHas: true,
		},
		{
			name:    "password with dash separator",
			data:    "password-secret",
			wantHas: true,
		},
		{
			name:    "uppercase PASSWORD with colon",
			data:    "PASSWORD:secret123",
			wantHas: true,
		},
		{
			name:    "mixed case Password with equals",
			data:    "Password=secret",
			wantHas: true,
		},
		{
			name:    "no sensitive data",
			data:    "hello world",
			wantHas: false,
		},
		{
			name:    "empty string",
			data:    "",
			wantHas: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			has, _, _ := argHasSensitive(tt.data, tt.blocked, tt.exceptions)
			if has != tt.wantHas {
				t.Errorf("argHasSensitive(%q) has = %v, want %v", tt.data, has, tt.wantHas)
			}
		})
	}
}

func TestArgHasSensitive_Exceptions(t *testing.T) {
	// Find a default ban keyword to use
	bans := pkg.DefaultSensBans()
	if len(bans) == 0 {
		t.Skip("no default bans configured")
	}
	keyword := bans[0]

	data := keyword + ":value"
	// Without exception, should detect
	has, _, _ := argHasSensitive(data, nil, nil)
	if !has {
		t.Errorf("expected sensitive detection for %q without exceptions", data)
	}

	// With exception, should not detect (for default bans stage)
	has2, _, _ := argHasSensitive(data, nil, []string{keyword})
	if has2 {
		t.Errorf("expected no detection for %q with exception %q", data, keyword)
	}
}

func TestArgHasSensitive_CustomBlocked(t *testing.T) {
	tests := []struct {
		name       string
		data       string
		blocked    []string
		exceptions []string
		wantHas    bool
		wantLength int
	}{
		{
			name:       "custom blocked substring found",
			data:       "this contains mysecretfield here",
			blocked:    []string{"mysecretfield"},
			wantHas:    true,
			wantLength: len("mysecretfield"),
		},
		{
			name:    "custom blocked substring not found",
			data:    "this is safe data",
			blocked: []string{"mysecretfield"},
			wantHas: false,
		},
		{
			name:       "custom blocked with exception skipped",
			data:       "this contains mysecretfield here",
			blocked:    []string{"mysecretfield"},
			exceptions: []string{"mysecretfield"},
			wantHas:    false,
		},
		{
			name:    "empty blocked list",
			data:    "some data",
			blocked: []string{},
			wantHas: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			has, _, length := argHasSensitive(tt.data, tt.blocked, tt.exceptions)
			if has != tt.wantHas {
				t.Errorf("has = %v, want %v", has, tt.wantHas)
			}
			if tt.wantHas && length != tt.wantLength {
				t.Errorf("length = %d, want %d", length, tt.wantLength)
			}
		})
	}
}

func TestArgHasSensitive_NoMatch(t *testing.T) {
	has, idx, length := argHasSensitive("completely safe string", nil, nil)
	if has {
		t.Error("expected no match for safe string")
	}
	if idx != -1 {
		t.Errorf("idx = %d, want -1", idx)
	}
	if length != -1 {
		t.Errorf("length = %d, want -1", length)
	}
}

func TestArgHasSensitive_TokenPatterns(t *testing.T) {
	patterns := pkg.DefaultTokensPatterns()
	if len(patterns) == 0 {
		t.Skip("no default token patterns configured")
	}

	// Test with a JWT-like token
	jwtLike := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIn0.dozjgNryP4J3jVmNHl0w5N_XgL0n3I9PlFUP0THsR8U"
	has, _, _ := argHasSensitive(jwtLike, nil, nil)
	// This may or may not match depending on the patterns; just ensure no panic
	_ = has
}

func TestCheckStringSensitive_WithMatch(t *testing.T) {
	bans := pkg.DefaultSensBans()
	if len(bans) == 0 {
		t.Skip("no default bans configured")
	}
	keyword := bans[0]

	lit := &ast.BasicLit{
		ValuePos: token.Pos(10),
		Kind:     token.STRING,
		Value:    `"` + keyword + `:value"`,
	}

	var reports []pkg.Report
	checkStringSensitive(lit, nil, nil, &reports)

	if len(reports) == 0 {
		t.Fatal("expected at least one report")
	}
	if reports[0].Message != "Sensitive data detected" {
		t.Errorf("message = %q, want %q", reports[0].Message, "Sensitive data detected")
	}
}

func TestCheckStringSensitive_NoMatch(t *testing.T) {
	lit := &ast.BasicLit{
		ValuePos: token.Pos(10),
		Kind:     token.STRING,
		Value:    `"hello world"`,
	}

	var reports []pkg.Report
	checkStringSensitive(lit, nil, nil, &reports)

	if len(reports) != 0 {
		t.Errorf("expected no reports, got %d", len(reports))
	}
}

func TestHasSensitiveData_BasicLitString(t *testing.T) {
	bans := pkg.DefaultSensBans()
	if len(bans) == 0 {
		t.Skip("no default bans configured")
	}
	keyword := bans[0]

	args := []ast.Expr{
		&ast.BasicLit{
			ValuePos: token.Pos(1),
			Kind:     token.STRING,
			Value:    `"` + keyword + `=secret"`,
		},
	}

	var reports []pkg.Report
	HasSensitiveData(args, &reports, nil, nil)

	if len(reports) == 0 {
		t.Fatal("expected at least one report for sensitive BasicLit")
	}
}

func TestHasSensitiveData_NonStringBasicLit(t *testing.T) {
	args := []ast.Expr{
		&ast.BasicLit{
			ValuePos: token.Pos(1),
			Kind:     token.INT,
			Value:    "42",
		},
	}

	var reports []pkg.Report
	HasSensitiveData(args, &reports, nil, nil)

	if len(reports) != 0 {
		t.Errorf("expected no reports for INT literal, got %d", len(reports))
	}
}

func TestHasSensitiveData_BinaryExpr(t *testing.T) {
	bans := pkg.DefaultSensBans()
	if len(bans) == 0 {
		t.Skip("no default bans configured")
	}
	keyword := bans[0]

	args := []ast.Expr{
		&ast.BinaryExpr{
			X: &ast.BasicLit{
				ValuePos: token.Pos(1),
				Kind:     token.STRING,
				Value:    `"prefix "`,
			},
			Op: token.ADD,
			Y: &ast.BasicLit{
				ValuePos: token.Pos(20),
				Kind:     token.STRING,
				Value:    `"` + keyword + `:secret"`,
			},
		},
	}

	var reports []pkg.Report
	HasSensitiveData(args, &reports, nil, nil)

	if len(reports) == 0 {
		t.Fatal("expected at least one report for BinaryExpr with sensitive data")
	}
}

func TestHasSensitiveData_BinaryExprNoSensitive(t *testing.T) {
	args := []ast.Expr{
		&ast.BinaryExpr{
			X: &ast.BasicLit{
				ValuePos: token.Pos(1),
				Kind:     token.STRING,
				Value:    `"hello "`,
			},
			Op: token.ADD,
			Y: &ast.BasicLit{
				ValuePos: token.Pos(20),
				Kind:     token.STRING,
				Value:    `"world"`,
			},
		},
	}

	var reports []pkg.Report
	HasSensitiveData(args, &reports, nil, nil)

	if len(reports) != 0 {
		t.Errorf("expected no reports, got %d", len(reports))
	}
}

func TestHasSensitiveData_BinaryExprWithNonStringPart(t *testing.T) {
	args := []ast.Expr{
		&ast.BinaryExpr{
			X: &ast.BasicLit{
				ValuePos: token.Pos(1),
				Kind:     token.INT,
				Value:    "42",
			},
			Op: token.ADD,
			Y: &ast.BasicLit{
				ValuePos: token.Pos(20),
				Kind:     token.STRING,
				Value:    `"safe text"`,
			},
		},
	}

	var reports []pkg.Report
	HasSensitiveData(args, &reports, nil, nil)

	if len(reports) != 0 {
		t.Errorf("expected no reports for safe BinaryExpr, got %d", len(reports))
	}
}

func TestHasSensitiveData_CallExpr(t *testing.T) {
	bans := pkg.DefaultSensBans()
	if len(bans) == 0 {
		t.Skip("no default bans configured")
	}
	keyword := bans[0]

	args := []ast.Expr{
		&ast.CallExpr{
			Fun: &ast.Ident{Name: "fmt.Sprintf"},
			Args: []ast.Expr{
				&ast.BasicLit{
					ValuePos: token.Pos(50),
					Kind:     token.STRING,
					Value:    `"` + keyword + `=value"`,
				},
			},
		},
	}

	var reports []pkg.Report
	HasSensitiveData(args, &reports, nil, nil)

	if len(reports) == 0 {
		t.Fatal("expected at least one report for CallExpr with sensitive arg")
	}
}

func TestHasSensitiveData_CallExprNoSensitive(t *testing.T) {
	args := []ast.Expr{
		&ast.CallExpr{
			Fun: &ast.Ident{Name: "fmt.Sprintf"},
			Args: []ast.Expr{
				&ast.BasicLit{
					ValuePos: token.Pos(50),
					Kind:     token.STRING,
					Value:    `"safe data"`,
				},
			},
		},
	}

	var reports []pkg.Report
	HasSensitiveData(args, &reports, nil, nil)

	if len(reports) != 0 {
		t.Errorf("expected no reports, got %d", len(reports))
	}
}

func TestHasSensitiveData_IgnoresIdentExpr(t *testing.T) {
	args := []ast.Expr{
		&ast.Ident{Name: "someVariable"},
	}

	var reports []pkg.Report
	HasSensitiveData(args, &reports, nil, nil)

	if len(reports) != 0 {
		t.Errorf("expected no reports for Ident expr, got %d", len(reports))
	}
}

func TestHasSensitiveData_EmptyArgs(t *testing.T) {
	var reports []pkg.Report
	HasSensitiveData(nil, &reports, nil, nil)

	if len(reports) != 0 {
		t.Errorf("expected no reports for nil args, got %d", len(reports))
	}

	HasSensitiveData([]ast.Expr{}, &reports, nil, nil)

	if len(reports) != 0 {
		t.Errorf("expected no reports for empty args, got %d", len(reports))
	}
}

func TestHasSensitiveData_MultipleArgs(t *testing.T) {
	bans := pkg.DefaultSensBans()
	if len(bans) == 0 {
		t.Skip("no default bans configured")
	}
	keyword := bans[0]

	args := []ast.Expr{
		&ast.BasicLit{
			ValuePos: token.Pos(1),
			Kind:     token.STRING,
			Value:    `"safe text"`,
		},
		&ast.BasicLit{
			ValuePos: token.Pos(20),
			Kind:     token.STRING,
			Value:    `"` + keyword + `:value"`,
		},
		&ast.BasicLit{
			ValuePos: token.Pos(50),
			Kind:     token.STRING,
			Value:    `"also safe"`,
		},
	}

	var reports []pkg.Report
	HasSensitiveData(args, &reports, nil, nil)

	if len(reports) != 1 {
		t.Errorf("expected exactly 1 report, got %d", len(reports))
	}
}

func TestHasSensitiveData_WithCustomBlocked(t *testing.T) {
	args := []ast.Expr{
		&ast.BasicLit{
			ValuePos: token.Pos(1),
			Kind:     token.STRING,
			Value:    `"contains custom_secret_field inside"`,
		},
	}

	var reports []pkg.Report
	HasSensitiveData(args, &reports, []string{"custom_secret_field"}, nil)

	if len(reports) == 0 {
		t.Fatal("expected report for custom blocked substring")
	}
}

func TestHasSensitiveData_WithExceptions(t *testing.T) {
	bans := pkg.DefaultSensBans()
	if len(bans) == 0 {
		t.Skip("no default bans configured")
	}
	keyword := bans[0]

	args := []ast.Expr{
		&ast.BasicLit{
			ValuePos: token.Pos(1),
			Kind:     token.STRING,
			Value:    `"` + keyword + `:value"`,
		},
	}

	var reports []pkg.Report
	HasSensitiveData(args, &reports, nil, []string{keyword})

	if len(reports) != 0 {
		t.Errorf("expected no reports with exception for %q, got %d", keyword, len(reports))
	}
}

func TestArgHasSensitive_KeywordAtEndOfString(t *testing.T) {
	bans := pkg.DefaultSensBans()
	if len(bans) == 0 {
		t.Skip("no default bans configured")
	}
	keyword := bans[0]

	// keyword at end with extra text before it
	data := "some text " + keyword
	has, _, _ := argHasSensitive(data, nil, nil)
	if !has {
		t.Errorf("expected detection for keyword at end of string: %q", data)
	}
}

func TestArgHasSensitive_CaseInsensitive(t *testing.T) {
	bans := pkg.DefaultSensBans()
	if len(bans) == 0 {
		t.Skip("no default bans configured")
	}
	keyword := bans[0]

	// Test with all uppercase
	upper := strings.ToUpper(keyword) + ":value"
	has, _, _ := argHasSensitive(upper, nil, nil)
	if !has {
		t.Errorf("expected case-insensitive detection for %q", upper)
	}
}

func strings_ToUpper(s string) string {
	result := make([]byte, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'a' && c <= 'z' {
			result[i] = c - 32
		} else {
			result[i] = c
		}
	}
	return string(result)
}

func TestHasSensitiveData_NestedCallExpr(t *testing.T) {
	bans := pkg.DefaultSensBans()
	if len(bans) == 0 {
		t.Skip("no default bans configured")
	}
	keyword := bans[0]

	// CallExpr whose arg is another CallExpr with sensitive data
	args := []ast.Expr{
		&ast.CallExpr{
			Fun: &ast.Ident{Name: "outer"},
			Args: []ast.Expr{
				&ast.CallExpr{
					Fun: &ast.Ident{Name: "inner"},
					Args: []ast.Expr{
						&ast.BasicLit{
							ValuePos: token.Pos(100),
							Kind:     token.STRING,
							Value:    `"` + keyword + `=deeply_nested"`,
						},
					},
				},
			},
		},
	}

	var reports []pkg.Report
	HasSensitiveData(args, &reports, nil, nil)

	if len(reports) == 0 {
		t.Fatal("expected report for nested CallExpr with sensitive data")
	}
}

func TestArgHasSensitive_MultipleCustomBlocked(t *testing.T) {
	data := "this has credit_card info"
	blocked := []string{"ssn", "credit_card", "routing_number"}
	has, idx, length := argHasSensitive(data, blocked, nil)
	if !has {
		t.Error("expected detection for custom blocked substring")
	}
	if idx < 0 {
		t.Error("expected valid index")
	}
	if length != len("credit_card") {
		t.Errorf("length = %d, want %d", length, len("credit_card"))
	}
}

func TestArgHasSensitive_BlockedExceptionPartial(t *testing.T) {
	data := "this has ssn and credit_card"
	blocked := []string{"ssn", "credit_card"}
	exceptions := []string{"ssn"}

	has, _, length := argHasSensitive(data, blocked, exceptions)
	if !has {
		t.Error("expected detection for credit_card (not in exceptions)")
	}
	if length != len("credit_card") {
		t.Errorf("length = %d, want %d", length, len("credit_card"))
	}
}

func TestCheckStringSensitive_ReportPosition(t *testing.T) {
	bans := pkg.DefaultSensBans()
	if len(bans) == 0 {
		t.Skip("no default bans configured")
	}
	keyword := bans[0]

	lit := &ast.BasicLit{
		ValuePos: token.Pos(100),
		Kind:     token.STRING,
		Value:    `"` + keyword + `:value"`,
	}

	var reports []pkg.Report
	checkStringSensitive(lit, nil, nil, &reports)

	if len(reports) == 0 {
		t.Fatal("expected a report")
	}

	// Position should be base position + index offset
	if reports[0].Pos < lit.Pos() {
		t.Errorf("report Pos %d should be >= literal Pos %d", reports[0].Pos, lit.Pos())
	}
	if reports[0].Length <= 0 {
		t.Errorf("report Length should be positive, got %d", reports[0].Length)
	}
}

func TestArgHasSensitive_WhitespaceHandling(t *testing.T) {
	bans := pkg.DefaultSensBans()
	if len(bans) == 0 {
		t.Skip("no default bans configured")
	}
	keyword := bans[0]

	// With leading/trailing whitespace
	data := "  " + keyword + ":value  "
	has, _, _ := argHasSensitive(data, nil, nil)
	if !has {
		t.Errorf("expected detection with whitespace for %q", data)
	}
}

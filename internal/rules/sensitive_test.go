package rules

import (
	"go/ast"
	"go/token"
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
			data:    "password is secret",
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
			name:    "mixed case PassWord with equals",
			data:    "PassWord=abc",
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
		{
			name:       "password in exceptions is skipped",
			data:       "password:secret",
			exceptions: []string{"password", "pass", "secret"},
			wantHas:    false,
		},
		{
			name:    "secret with colon",
			data:    "secret:value",
			wantHas: true,
		},
		{
			name:    "token with equals",
			data:    "token=abc123",
			wantHas: true,
		},
		{
			name:    "api_key with colon",
			data:    "api_key:xyz",
			wantHas: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			has, idx, length := argHasSensitive(tt.data, tt.blocked, tt.exceptions)
			if has != tt.wantHas {
				t.Errorf("argHasSensitive(%q) has = %v, want %v (idx=%d, length=%d)", tt.data, has, tt.wantHas, idx, length)
			}
			if tt.wantHas && (idx == -1 || length == -1) {
				t.Errorf("argHasSensitive(%q) expected valid idx and length, got idx=%d, length=%d", tt.data, idx, length)
			}
			if !tt.wantHas && (idx != -1 || length != -1) {
				t.Errorf("argHasSensitive(%q) expected idx=-1 and length=-1, got idx=%d, length=%d", tt.data, idx, length)
			}
		})
	}
}

func TestArgHasSensitive_CustomBlocked(t *testing.T) {
	tests := []struct {
		name       string
		data       string
		blocked    []string
		exceptions []string
		wantHas    bool
	}{
		{
			name:    "custom blocked substring found",
			data:    "this contains mysecretword here",
			blocked: []string{"mysecretword"},
			wantHas: true,
		},
		{
			name:    "custom blocked not found",
			data:    "this is safe",
			blocked: []string{"mysecretword"},
			wantHas: false,
		},
		{
			name:       "custom blocked in exceptions is skipped",
			data:       "this contains mysecretword here",
			blocked:    []string{"mysecretword"},
			exceptions: []string{"mysecretword"},
			wantHas:    false,
		},
		{
			name:    "multiple custom blocked, second matches",
			data:    "data with secondblock inside",
			blocked: []string{"firstblock", "secondblock"},
			wantHas: true,
		},
		{
			name:    "empty blocked list",
			data:    "no sensitive content at all",
			blocked: []string{},
			wantHas: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			has, idx, length := argHasSensitive(tt.data, tt.blocked, tt.exceptions)
			if has != tt.wantHas {
				t.Errorf("argHasSensitive(%q, blocked=%v, exceptions=%v) has = %v, want %v (idx=%d, length=%d)", tt.data, tt.blocked, tt.exceptions, has, tt.wantHas, idx, length)
			}
		})
	}
}

func TestArgHasSensitive_TokenPatterns(t *testing.T) {
	tests := []struct {
		name    string
		data    string
		wantHas bool
	}{
		{
			name:    "JWT-like token",
			data:    "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIn0.dozjgNryP4J3jVmNHl0w5N_XgL0n3I9PlFUP0THsR8U",
			wantHas: true,
		},
		{
			name:    "no token pattern",
			data:    "just a regular log message",
			wantHas: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			has, _, _ := argHasSensitive(tt.data, nil, nil)
			if has != tt.wantHas {
				t.Errorf("argHasSensitive(%q) has = %v, want %v", tt.data, has, tt.wantHas)
			}
		})
	}
}

func TestArgHasSensitive_ReturnValues(t *testing.T) {
	// Test that idx and length are correct for a known match
	data := "user password:abc"
	has, idx, length := argHasSensitive(data, nil, nil)
	if !has {
		t.Fatal("expected sensitive data to be detected")
	}
	if idx < 0 {
		t.Errorf("expected non-negative idx, got %d", idx)
	}
	if length <= 0 {
		t.Errorf("expected positive length, got %d", length)
	}
}

func TestCheckStringSensitive_WithSensitiveData(t *testing.T) {
	lit := &ast.BasicLit{
		ValuePos: token.Pos(10),
		Kind:     token.STRING,
		Value:    `"password:secret"`,
	}

	reports := checkStringSensitive(lit, nil, nil)
	if len(reports) == 0 {
		t.Fatal("expected at least one report for sensitive string")
	}
	if reports[0].Message != "Sensitive data detected" {
		t.Errorf("unexpected message: %s", reports[0].Message)
	}
	if reports[0].Length <= 0 {
		t.Errorf("expected positive length, got %d", reports[0].Length)
	}
}

func TestCheckStringSensitive_NoSensitiveData(t *testing.T) {
	lit := &ast.BasicLit{
		ValuePos: token.Pos(10),
		Kind:     token.STRING,
		Value:    `"hello world"`,
	}

	reports := checkStringSensitive(lit, nil, nil)
	if len(reports) != 0 {
		t.Errorf("expected no reports, got %d", len(reports))
	}
}

func TestCheckStringSensitive_PosIncludesOffset(t *testing.T) {
	lit := &ast.BasicLit{
		ValuePos: token.Pos(100),
		Kind:     token.STRING,
		Value:    `"password:secret"`,
	}

	reports := checkStringSensitive(lit, nil, nil)
	if len(reports) == 0 {
		t.Fatal("expected a report")
	}
	// Pos should be >= 100 (base pos) since idx >= 0
	if reports[0].Pos < token.Pos(100) {
		t.Errorf("expected Pos >= 100, got %d", reports[0].Pos)
	}
}

func TestHasSensitiveData_BasicLitString(t *testing.T) {
	lit := &ast.BasicLit{
		ValuePos: token.Pos(1),
		Kind:     token.STRING,
		Value:    `"password=abc"`,
	}

	args := []ast.Expr{lit}
	var reports []pkg.Report
	HasSensitiveData(args, &reports, nil, nil)

	if len(reports) == 0 {
		t.Fatal("expected at least one report")
	}
	if reports[0].Message != "Sensitive data detected" {
		t.Errorf("unexpected message: %s", reports[0].Message)
	}
}

func TestHasSensitiveData_BasicLitNoSensitive(t *testing.T) {
	lit := &ast.BasicLit{
		ValuePos: token.Pos(1),
		Kind:     token.STRING,
		Value:    `"hello"`,
	}

	args := []ast.Expr{lit}
	var reports []pkg.Report
	HasSensitiveData(args, &reports, nil, nil)

	if len(reports) != 0 {
		t.Errorf("expected no reports, got %d", len(reports))
	}
}

func TestHasSensitiveData_SkipsNonStringBasicLit(t *testing.T) {
	lit := &ast.BasicLit{
		ValuePos: token.Pos(1),
		Kind:     token.INT,
		Value:    "42",
	}

	args := []ast.Expr{lit}
	var reports []pkg.Report
	HasSensitiveData(args, &reports, nil, nil)

	if len(reports) != 0 {
		t.Errorf("expected no reports for non-string literal, got %d", len(reports))
	}
}

func TestHasSensitiveData_BinaryExpr(t *testing.T) {
	left := &ast.BasicLit{
		ValuePos: token.Pos(1),
		Kind:     token.STRING,
		Value:    `"prefix "`,
	}
	right := &ast.BasicLit{
		ValuePos: token.Pos(20),
		Kind:     token.STRING,
		Value:    `"password:abc"`,
	}

	binExpr := &ast.BinaryExpr{
		X:  left,
		Op: token.ADD,
		Y:  right,
	}

	args := []ast.Expr{binExpr}
	var reports []pkg.Report
	HasSensitiveData(args, &reports, nil, nil)

	if len(reports) == 0 {
		t.Fatal("expected at least one report for binary expression with sensitive part")
	}
}

func TestHasSensitiveData_BinaryExprNoSensitive(t *testing.T) {
	left := &ast.BasicLit{
		ValuePos: token.Pos(1),
		Kind:     token.STRING,
		Value:    `"hello "`,
	}
	right := &ast.BasicLit{
		ValuePos: token.Pos(20),
		Kind:     token.STRING,
		Value:    `"world"`,
	}

	binExpr := &ast.BinaryExpr{
		X:  left,
		Op: token.ADD,
		Y:  right,
	}

	args := []ast.Expr{binExpr}
	var reports []pkg.Report
	HasSensitiveData(args, &reports, nil, nil)

	if len(reports) != 0 {
		t.Errorf("expected no reports, got %d", len(reports))
	}
}

func TestHasSensitiveData_BinaryExprWithNonStringPart(t *testing.T) {
	left := &ast.BasicLit{
		ValuePos: token.Pos(1),
		Kind:     token.INT,
		Value:    "42",
	}
	right := &ast.BasicLit{
		ValuePos: token.Pos(20),
		Kind:     token.STRING,
		Value:    `"safe message"`,
	}

	binExpr := &ast.BinaryExpr{
		X:  left,
		Op: token.ADD,
		Y:  right,
	}

	args := []ast.Expr{binExpr}
	var reports []pkg.Report
	HasSensitiveData(args, &reports, nil, nil)

	if len(reports) != 0 {
		t.Errorf("expected no reports, got %d", len(reports))
	}
}

func TestHasSensitiveData_CallExprRecursion(t *testing.T) {
	innerLit := &ast.BasicLit{
		ValuePos: token.Pos(1),
		Kind:     token.STRING,
		Value:    `"password:secret"`,
	}

	callExpr := &ast.CallExpr{
		Fun:  &ast.Ident{Name: "fmt.Sprintf"},
		Args: []ast.Expr{innerLit},
	}

	args := []ast.Expr{callExpr}
	var reports []pkg.Report
	HasSensitiveData(args, &reports, nil, nil)

	if len(reports) == 0 {
		t.Fatal("expected report from recursive call expression check")
	}
}

func TestHasSensitiveData_CallExprNoSensitive(t *testing.T) {
	innerLit := &ast.BasicLit{
		ValuePos: token.Pos(1),
		Kind:     token.STRING,
		Value:    `"safe message"`,
	}

	callExpr := &ast.CallExpr{
		Fun:  &ast.Ident{Name: "fmt.Sprintf"},
		Args: []ast.Expr{innerLit},
	}

	args := []ast.Expr{callExpr}
	var reports []pkg.Report
	HasSensitiveData(args, &reports, nil, nil)

	if len(reports) != 0 {
		t.Errorf("expected no reports, got %d", len(reports))
	}
}

func TestHasSensitiveData_IgnoresUnknownExprTypes(t *testing.T) {
	// An Ident is not a BasicLit, BinaryExpr, or CallExpr — should be ignored
	ident := &ast.Ident{Name: "password"}

	args := []ast.Expr{ident}
	var reports []pkg.Report
	HasSensitiveData(args, &reports, nil, nil)

	if len(reports) != 0 {
		t.Errorf("expected no reports for Ident expression, got %d", len(reports))
	}
}

func TestHasSensitiveData_EmptyArgs(t *testing.T) {
	var reports []pkg.Report
	HasSensitiveData([]ast.Expr{}, &reports, nil, nil)

	if len(reports) != 0 {
		t.Errorf("expected no reports for empty args, got %d", len(reports))
	}
}

func TestHasSensitiveData_MultipleArgs(t *testing.T) {
	safe := &ast.BasicLit{
		ValuePos: token.Pos(1),
		Kind:     token.STRING,
		Value:    `"safe"`,
	}
	sensitive := &ast.BasicLit{
		ValuePos: token.Pos(50),
		Kind:     token.STRING,
		Value:    `"password=abc"`,
	}

	args := []ast.Expr{safe, sensitive}
	var reports []pkg.Report
	HasSensitiveData(args, &reports, nil, nil)

	if len(reports) != 1 {
		t.Errorf("expected 1 report, got %d", len(reports))
	}
}

func TestHasSensitiveData_MultipleSensitiveArgs(t *testing.T) {
	sens1 := &ast.BasicLit{
		ValuePos: token.Pos(1),
		Kind:     token.STRING,
		Value:    `"password=abc"`,
	}
	sens2 := &ast.BasicLit{
		ValuePos: token.Pos(50),
		Kind:     token.STRING,
		Value:    `"secret:xyz"`,
	}

	args := []ast.Expr{sens1, sens2}
	var reports []pkg.Report
	HasSensitiveData(args, &reports, nil, nil)

	if len(reports) < 2 {
		t.Errorf("expected at least 2 reports, got %d", len(reports))
	}
}

func TestHasSensitiveData_WithCustomBlockedAndExceptions(t *testing.T) {
	lit := &ast.BasicLit{
		ValuePos: token.Pos(1),
		Kind:     token.STRING,
		Value:    `"contains customsecret data"`,
	}

	args := []ast.Expr{lit}

	// Should detect with custom blocked
	var reports []pkg.Report
	HasSensitiveData(args, &reports, []string{"customsecret"}, nil)
	if len(reports) == 0 {
		t.Fatal("expected report for custom blocked substring")
	}

	// Should skip with exception
	reports = nil
	HasSensitiveData(args, &reports, []string{"customsecret"}, []string{"customsecret"})
	if len(reports) != 0 {
		t.Errorf("expected no reports when custom blocked is in exceptions, got %d", len(reports))
	}
}
func TestArgHasSensitive_CaseInsensitivity(t *testing.T) {
	has1, _, _ := argHasSensitive("PASSWORD:abc", nil, nil)
	has2, _, _ := argHasSensitive("password:abc", nil, nil)
	has3, _, _ := argHasSensitive("Password:abc", nil, nil)

	if has1 != has2 || has2 != has3 {
		t.Error("expected case-insensitive detection to produce consistent results")
	}
}

func TestArgHasSensitive_NilBlockedAndExceptions(t *testing.T) {
	// Should not panic with nil slices
	has, idx, length := argHasSensitive("just a normal string", nil, nil)
	if has {
		t.Errorf("expected no sensitive data, got has=true, idx=%d, length=%d", idx, length)
	}
}

func TestHasSensitiveData_NestedCallExpr(t *testing.T) {
	innerLit := &ast.BasicLit{
		ValuePos: token.Pos(1),
		Kind:     token.STRING,
		Value:    `"password=abc"`,
	}

	innerCall := &ast.CallExpr{
		Fun:  &ast.Ident{Name: "inner"},
		Args: []ast.Expr{innerLit},
	}

	outerCall := &ast.CallExpr{
		Fun:  &ast.Ident{Name: "outer"},
		Args: []ast.Expr{innerCall},
	}

	args := []ast.Expr{outerCall}
	var reports []pkg.Report
	HasSensitiveData(args, &reports, nil, nil)

	if len(reports) == 0 {
		t.Fatal("expected report from nested call expression")
	}
}

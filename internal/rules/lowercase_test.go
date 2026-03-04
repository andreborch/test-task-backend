package rules

import (
	"go/ast"
	"go/token"
	"testing"

	"github.com/andreborch/log-linter/pkg"
)

func TestIsLetterLower(t *testing.T) {
	tests := []struct {
		name     string
		input    rune
		expected bool
	}{
		{"lowercase letter", 'a', true},
		{"uppercase letter", 'A', false},
		{"digit", '1', true},
		{"space", ' ', true},
		{"punctuation", '.', true},
		{"unicode lowercase", 'é', true},
		{"unicode uppercase", 'É', false},
		{"symbol", '@', true},
		{"underscore", '_', true},
		{"lowercase z", 'z', true},
		{"uppercase Z", 'Z', false},
		{"cyrillic lowercase", 'я', true},
		{"cyrillic uppercase", 'Я', false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isLetterLower(tt.input)
			if result != tt.expected {
				t.Errorf("isLetterLower(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestCheckStringLowercase(t *testing.T) {
	tests := []struct {
		name        string
		value       string
		expectCount int
	}{
		{"lowercase start", `"hello world"`, 0},
		{"uppercase start", `"Hello world"`, 1},
		{"digit start", `"1hello"`, 0},
		{"symbol start", `"!hello"`, 0},
		{"single uppercase char", `"H"`, 1},
		{"single lowercase char", `"h"`, 0},
		{"space start", `" hello"`, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lit := &ast.BasicLit{
				ValuePos: token.Pos(1),
				Kind:     token.STRING,
				Value:    tt.value,
			}

			reports := checkStringLowercase(lit)
			if len(reports) != tt.expectCount {
				t.Errorf("checkStringLowercase(%q) produced %d reports, want %d", tt.value, len(reports), tt.expectCount)
			}
			if tt.expectCount > 0 && len(reports) > 0 {
				if reports[0].Message != "Log message should start with lowercase" {
					t.Errorf("unexpected message: %s", reports[0].Message)
				}
				if reports[0].Length != len(tt.value) {
					t.Errorf("expected length %d, got %d", len(tt.value), reports[0].Length)
				}
				if reports[0].Pos != token.Pos(2) {
					t.Errorf("expected pos %d, got %d", token.Pos(2), reports[0].Pos)
				}
			}
		})
	}
}

func TestCheckLowerCase_BasicLit(t *testing.T) {
	tests := []struct {
		name        string
		value       string
		kind        token.Token
		expectCount int
	}{
		{"lowercase string", `"hello"`, token.STRING, 0},
		{"uppercase string", `"Hello"`, token.STRING, 1},
		{"int literal ignored", `42`, token.INT, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lit := &ast.BasicLit{
				ValuePos: token.Pos(1),
				Kind:     tt.kind,
				Value:    tt.value,
			}
			args := []ast.Expr{lit}
			var reports []pkg.Report
			CheckLowerCase(args, &reports)
			if len(reports) != tt.expectCount {
				t.Errorf("CheckLowerCase basic lit %q produced %d reports, want %d", tt.value, len(reports), tt.expectCount)
			}
		})
	}
}

func TestCheckLowerCase_BinaryExpr(t *testing.T) {
	tests := []struct {
		name        string
		leftValue   string
		leftKind    token.Token
		expectCount int
	}{
		{"binary lowercase left", `"hello "`, token.STRING, 0},
		{"binary uppercase left", `"Hello "`, token.STRING, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			left := &ast.BasicLit{
				ValuePos: token.Pos(1),
				Kind:     tt.leftKind,
				Value:    tt.leftValue,
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
			CheckLowerCase(args, &reports)
			if len(reports) != tt.expectCount {
				t.Errorf("CheckLowerCase binary expr produced %d reports, want %d", len(reports), tt.expectCount)
			}
		})
	}
}

func TestCheckLowerCase_BinaryExprNonStringLeft(t *testing.T) {
	// When left side of binary expr is not a string literal, should be skipped
	left := &ast.Ident{Name: "someVar"}
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
	CheckLowerCase(args, &reports)
	if len(reports) != 0 {
		t.Errorf("expected 0 reports for non-string left in binary expr, got %d", len(reports))
	}
}

func TestCheckLowerCase_CallExpr(t *testing.T) {
	// CallExpr wrapping a string literal argument
	innerLit := &ast.BasicLit{
		ValuePos: token.Pos(10),
		Kind:     token.STRING,
		Value:    `"Hello"`,
	}
	callExpr := &ast.CallExpr{
		Fun:  &ast.Ident{Name: "fmt.Sprintf"},
		Args: []ast.Expr{innerLit},
	}
	args := []ast.Expr{callExpr}
	var reports []pkg.Report
	CheckLowerCase(args, &reports)
	if len(reports) != 1 {
		t.Errorf("expected 1 report for uppercase in CallExpr arg, got %d", len(reports))
	}
}

func TestCheckLowerCase_CallExprLowercase(t *testing.T) {
	innerLit := &ast.BasicLit{
		ValuePos: token.Pos(10),
		Kind:     token.STRING,
		Value:    `"hello"`,
	}
	callExpr := &ast.CallExpr{
		Fun:  &ast.Ident{Name: "fmt.Sprintf"},
		Args: []ast.Expr{innerLit},
	}
	args := []ast.Expr{callExpr}
	var reports []pkg.Report
	CheckLowerCase(args, &reports)
	if len(reports) != 0 {
		t.Errorf("expected 0 reports for lowercase in CallExpr arg, got %d", len(reports))
	}
}

func TestCheckLowerCase_IdentIgnored(t *testing.T) {
	// An identifier (variable reference) should be ignored
	ident := &ast.Ident{Name: "myVar"}
	args := []ast.Expr{ident}
	var reports []pkg.Report
	CheckLowerCase(args, &reports)
	if len(reports) != 0 {
		t.Errorf("expected 0 reports for ident arg, got %d", len(reports))
	}
}

func TestCheckLowerCase_BinaryExprIntLeft(t *testing.T) {
	// Binary expression where left part is INT literal
	left := &ast.BasicLit{
		ValuePos: token.Pos(1),
		Kind:     token.INT,
		Value:    "42",
	}
	right := &ast.BasicLit{
		ValuePos: token.Pos(10),
		Kind:     token.STRING,
		Value:    `"hello"`,
	}
	binExpr := &ast.BinaryExpr{
		X:  left,
		Op: token.ADD,
		Y:  right,
	}
	args := []ast.Expr{binExpr}
	var reports []pkg.Report
	CheckLowerCase(args, &reports)
	if len(reports) != 0 {
		t.Errorf("expected 0 reports for int left in binary expr, got %d", len(reports))
	}
}

func TestCheckStringLowercase_ReportPosition(t *testing.T) {
	lit := &ast.BasicLit{
		ValuePos: token.Pos(50),
		Kind:     token.STRING,
		Value:    `"World"`,
	}

	reports := checkStringLowercase(lit)
	if len(reports) != 1 {
		t.Fatalf("expected 1 report, got %d", len(reports))
	}
	// Pos should be ValuePos + 1 to account for the opening quote
	if reports[0].Pos != token.Pos(51) {
		t.Errorf("expected pos 51, got %d", reports[0].Pos)
	}
	if reports[0].Length != len(`"World"`) {
		t.Errorf("expected length %d, got %d", len(`"World"`), reports[0].Length)
	}
}

func TestCheckLowerCase_NestedCallExpr(t *testing.T) {
	// Nested CallExpr: outer(inner("Hello"))
	innerLit := &ast.BasicLit{
		ValuePos: token.Pos(10),
		Kind:     token.STRING,
		Value:    `"Hello"`,
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
	CheckLowerCase(args, &reports)
	if len(reports) != 1 {
		t.Errorf("expected 1 report for nested CallExpr with uppercase, got %d", len(reports))
	}
}

func TestCheckLowerCase_MultipleReportsAccumulate(t *testing.T) {
	// Verify that reports accumulate across multiple calls
	var reports []pkg.Report

	lit1 := &ast.BasicLit{ValuePos: token.Pos(1), Kind: token.STRING, Value: `"Hello"`}
	lit2 := &ast.BasicLit{ValuePos: token.Pos(20), Kind: token.STRING, Value: `"World"`}

	CheckLowerCase([]ast.Expr{lit1}, &reports)
	CheckLowerCase([]ast.Expr{lit2}, &reports)

	if len(reports) != 2 {
		t.Errorf("expected 2 accumulated reports, got %d", len(reports))
	}
}

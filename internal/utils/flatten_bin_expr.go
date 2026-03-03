package utils

import (
	"go/ast"
	"go/token"
)

type NodePart struct {
	Expr ast.Expr
	Op   token.Token
}

// FlattenBinaryExpr traverses a binary-expression AST subtree and returns a flat,
// left-to-right slice of its terminal (non-*ast.BinaryExpr) expressions.
//
// It recursively descends through nested *ast.BinaryExpr nodes and appends only
// leaf expressions to the result, effectively removing binary-expression nesting
// from the returned structure.
//
// Example:
// a + (b + c)   -> []ast.Expr{a, b, c}
func FlattenBinaryExpr(be *ast.BinaryExpr) []ast.Expr {
	var out []ast.Expr

	var walk func(e ast.Expr, trailingOp token.Token)
	walk = func(e ast.Expr, trailingOp token.Token) {
		if b, ok := e.(*ast.BinaryExpr); ok {
			walk(b.X, b.Op)
			walk(b.Y, trailingOp)
			return
		}
		out = append(out, e)
	}

	walk(be, token.ILLEGAL)
	return out
}

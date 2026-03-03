package utils

import (
	"go/ast"
	"go/types"
)

// IsLogger reports whether the given call expression refers to a supported logger function.
//
// The function resolves the called function object from the AST using type information:
//   - for direct calls (e.g., Info(...)) via *ast.Ident
//   - for selector calls (e.g., log.Info(...)) via *ast.SelectorExpr
//
// A call is considered a logger call only if all conditions are met:
//   - info is not nil
//   - the function object can be resolved
//   - the function has an associated package
//   - the package path is present in sup_pkgs
//   - the function name is present in sup_funcs
//
// It returns:
//   - is_logger: true if the call matches a supported logger function, otherwise false
//   - args: the call arguments (fun.Args) when matched; nil otherwise
func IsLogger(fun *ast.CallExpr, info *types.Info, sup_pkgs []string, sup_funcs []string) (is_logger bool, args []ast.Expr) {
	if info == nil {
		return false, nil
	}

	var fn types.Object
	switch f := fun.Fun.(type) {
	case *ast.Ident:
		fn = info.ObjectOf(f)
	case *ast.SelectorExpr:
		fn = info.ObjectOf(f.Sel)
	}

	if fn == nil {
		return false, nil
	}

	if fn.Pkg() == nil {
		return false, nil
	}

	pkg_name := fn.Pkg().Path()
	func_name := fn.Name()
	if !Contains(sup_pkgs, pkg_name) {
		return false, nil
	}

	if !Contains(sup_funcs, func_name) {
		return false, nil
	}

	return true, fun.Args
}

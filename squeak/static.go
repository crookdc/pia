package squeak

import (
	"errors"
	"fmt"
	"github.com/crookdc/pia/squeak/ast"
	"github.com/crookdc/pia/squeak/token"
)

var (
	ErrResolverFault  = errors.New("resolver fault")
	ErrRedeclaredName = fmt.Errorf("%w: redeclared name", ErrResolverFault)
)

type resolver struct {
	stack struct {
		slice []map[string]struct{}
		sp    int
	}
	// locals map each variable expression (expressions that require resolution of a name in the environment list) to a
	// level in the environment list. A value of zero means that the name can be resolved in the direct scope of the
	// expression.
	locals map[string]int
}

func (r *resolver) Resolve(stmts []ast.StatementNode) error {
	for _, stmt := range stmts {
		if err := r.statement(stmt); err != nil {
			return err
		}
	}
	return nil
}

func (r *resolver) statement(stmt ast.StatementNode) error {
	switch stmt := stmt.(type) {
	case ast.Block:
		r.begin()
		for _, stmt := range stmt.Body {
			if err := r.statement(stmt); err != nil {
				return err
			}
		}
		r.end()
	case ast.Function:
		r.begin()
		for _, param := range stmt.Params {
			if err := r.declare(param); err != nil {
				return err
			}
		}
		// Declare the function name before the body, otherwise a lookup on the current functions name fails and in turn
		// recursion is broken.
		if err := r.declare(stmt.Name); err != nil {
			return err
		}
		for _, stmt := range stmt.Body.Body {
			if err := r.statement(stmt); err != nil {
				return err
			}
		}
		r.end()
	case ast.Declaration:
		if err := r.declare(stmt.Name); err != nil {
			return err
		}
		if stmt.Initializer == nil {
			return nil
		}
		return r.expression(stmt.Initializer)
	case ast.ExpressionStatement:
		return r.expression(stmt.Expression)
	default:
		// If this occurs then it is not a fault of the user but rather a fault of the resolver implementation having
		// not been updated to support a newly added statement type. Hence, panicking is justified.
		panic(fmt.Errorf("%w: %T is not a supported resolver statement", ErrResolverFault, stmt))
	}
	return nil
}

func (r *resolver) expression(expr ast.ExpressionNode) error {
	switch expr := expr.(type) {
	case ast.Variable:
		lvl := r.resolve(expr.Name)
		r.locals[expr.ID] = lvl
		return nil
	case ast.Assignment:
		return r.expression(expr.Value)
	case ast.Infix:
		if err := r.expression(expr.LHS); err != nil {
			return err
		}
		return r.expression(expr.RHS)
	case ast.Prefix:
		return r.expression(expr.Target)
	case ast.IntegerLiteral, ast.FloatLiteral, ast.StringLiteral, ast.BooleanLiteral, ast.NilLiteral:
		return nil
	case ast.ListLiteral:
		for _, item := range expr.Items {
			if err := r.expression(item); err != nil {
				return err
			}
		}
		return nil
	case ast.Grouping:
		return r.expression(expr.Group)
	case ast.Logical:
		if err := r.expression(expr.LHS); err != nil {
			return err
		}
		return r.expression(expr.RHS)
	case ast.Call:
		if err := r.expression(expr.Callee); err != nil {
			return err
		}
		for _, arg := range expr.Args {
			if err := r.expression(arg); err != nil {
				return err
			}
		}
		return nil
	default:
		// If this occurs then it is not a fault of the user but rather a fault of the resolver implementation having
		// not been updated to support a newly added expression type. Hence, panicking is justified.
		panic(fmt.Errorf("%w: %T is not a supported resolver expression", ErrResolverFault, expr))
	}
}

func (r *resolver) resolve(name token.Token) int {
	for i := range r.stack.sp + 1 {
		scope := r.stack.slice[r.stack.sp-i]
		if _, ok := scope[name.Lexeme]; ok {
			return i
		}
	}
	return r.stack.sp
}

func (r *resolver) scope() (map[string]struct{}, bool) {
	if r.stack.sp < 0 {
		return nil, false
	}
	return r.stack.slice[r.stack.sp], true
}

func (r *resolver) declare(name token.Token) error {
	sc, ok := r.scope()
	if !ok {
		return nil
	}
	if _, ok := sc[name.Lexeme]; ok {
		return fmt.Errorf("%w: %s", ErrRedeclaredName, name.Lexeme)
	}
	sc[name.Lexeme] = struct{}{}
	return nil
}

func (r *resolver) begin() {
	if r.stack.sp == len(r.stack.slice)-1 {
		// If the stack is the largest it has been (and the stack pointer would exceed the length of the stack if
		// incremented) then it must first be extended.
		r.stack.slice = append(r.stack.slice, map[string]struct{}{})
	}
	r.stack.sp += 1
	return
}

func (r *resolver) end() {
	if r.stack.sp < 0 {
		return
	}
	r.stack.sp -= 1
}

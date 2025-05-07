package squeak

import (
	"errors"
	"fmt"
	"github.com/crookdc/pia/squeak/internal/ast"
	"github.com/crookdc/pia/squeak/internal/token"
)

var (
	ErrUnexpectedType    = errors.New("unexpected type")
	ErrUnknownStatement  = errors.New("unknown statement")
	ErrUnknownExpression = errors.New("unknown expression")
)

// Scope holds available Object values for a given scope during evaluation of Squeak programs.
type Scope struct {
	parent *Scope
	table  map[string]Object
}

func NewEvaluator() *Evaluator {
	return &Evaluator{scope: Scope{
		parent: nil,
		table:  make(map[string]Object),
	}}
}

type Evaluator struct {
	scope Scope
}

// Run executes the provided [ast.StatementNode] within the context of the current scope. Mutating statements such
// as [ast.LetStatement] also subsequently mutate the internal state of the Evaluator.
func (ev *Evaluator) Run(stmt ast.StatementNode) (Object, error) {
	switch s := stmt.(type) {
	case ast.ExpressionStatement:
		return ev.expression(s.Expression)
	default:
		return NullObject, ErrUnknownStatement
	}
}

func (ev *Evaluator) expression(expr ast.ExpressionNode) (Object, error) {
	switch expr := expr.(type) {
	case ast.IntegerExpression:
		return Integer{expr.Integer}, nil
	case ast.StringExpression:
		return String{expr.String}, nil
	case ast.BooleanExpression:
		return Boolean{Boolean: expr.Boolean}, nil
	case ast.InfixExpression:
		return ev.infix(expr)
	default:
		return NullObject, ErrUnknownExpression
	}
}

func (ev *Evaluator) infix(expr ast.InfixExpression) (Object, error) {
	rhs, err := ev.expression(expr.RHS)
	if err != nil {
		return NullObject, err
	}
	lhs, err := ev.expression(expr.LHS)
	if err != nil {
		return NullObject, err
	}
	switch expr.Operator.Type {
	case token.Plus:
		return lhs.Add(rhs)
	case token.Minus:
		return lhs.Subtract(rhs)
	default:
		return NullObject, fmt.Errorf("unknown operator '%s'", expr.Operator.Lexeme)
	}
}

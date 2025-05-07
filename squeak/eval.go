package squeak

import (
	"errors"
	"github.com/crookdc/pia/squeak/internal/ast"
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

// Statement executes the provided [ast.StatementNode] within the context of the current scope. Mutating statements such
// as [ast.LetStatement] also subsequently mutate the internal state of the Evaluator.
func (ev *Evaluator) Statement(stmt ast.StatementNode) (Object, error) {
	switch s := stmt.(type) {
	case ast.ExpressionStatement:
		return ev.expression(s.Expression)
	default:
		return NullObject, ErrUnknownStatement
	}
}

func (ev *Evaluator) expression(expr ast.ExpressionNode) (Object, error) {
	switch e := expr.(type) {
	case ast.IntegerExpression:
		return Integer{e.Integer}, nil
	case ast.StringExpression:
		return String{e.String}, nil
	case ast.BooleanExpression:
		return Boolean{Boolean: e.Boolean}, nil
	default:
		return NullObject, ErrUnknownExpression
	}
}

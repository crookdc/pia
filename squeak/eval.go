package squeak

import (
	"errors"
	"fmt"
	"github.com/crookdc/pia/squeak/internal/ast"
	"github.com/crookdc/pia/squeak/internal/token"
	"reflect"
)

var ErrRuntimeFault = errors.New("runtime error")

type Object interface {
	fmt.Stringer
}

type Number struct {
	value float64
}

func (i Number) String() string {
	return fmt.Sprintf("%f", i.value)
}

type String struct {
	value string
}

func (s String) String() string {
	return s.value
}

type Boolean struct {
	value bool
}

func (b Boolean) String() string {
	if b.value {
		return "true"
	}
	return "false"
}

type Evaluator struct{}

func (ev *Evaluator) Evaluate(node ast.Node) (Object, error) {
	switch node := node.(type) {
	case ast.StatementNode:
		return nil, ev.statement(node)
	case ast.ExpressionNode:
		return ev.expression(node)
	default:
		return nil, fmt.Errorf(
			"%w: unexpected node type %s",
			ErrRuntimeFault,
			reflect.TypeOf(node),
		)
	}
}

func (ev *Evaluator) statement(node ast.StatementNode) error {
	switch node := node.(type) {
	case ast.ExpressionStatement:
		_, err := ev.expression(node.Expression)
		return err
	default:
		return fmt.Errorf(
			"%w: unexpected statement type %s",
			ErrRuntimeFault,
			reflect.TypeOf(node),
		)
	}
}

func (ev *Evaluator) expression(node ast.ExpressionNode) (Object, error) {
	switch node := node.(type) {
	case ast.IntegerLiteral:
		return Number{float64(node.Integer)}, nil
	case ast.StringLiteral:
		return String{node.String}, nil
	case ast.BooleanLiteral:
		return Boolean{node.Boolean}, nil
	case ast.Grouping:
		return ev.expression(node.Group)
	case ast.Prefix:
		return ev.prefix(node)
	case ast.Infix:
		return ev.infix(node)
	default:
		return nil, fmt.Errorf(
			"%w: unexpected expression type %s",
			ErrRuntimeFault,
			reflect.TypeOf(node),
		)
	}
}

func (ev *Evaluator) prefix(node ast.Prefix) (Object, error) {
	obj, err := ev.expression(node.Target)
	if err != nil {
		return nil, err
	}
	switch node.Operator.Type {
	case token.Bang:
		return Boolean{!ev.truthy(obj)}, nil
	case token.Minus:
		return ev.multiply(obj, Number{-1})
	default:
		return nil, fmt.Errorf(
			"%w: unexpected prefix operator %s",
			ErrRuntimeFault,
			node.Operator.Lexeme,
		)
	}
}

func (ev *Evaluator) truthy(obj Object) bool {
	if obj == nil {
		return false
	}
	switch obj := obj.(type) {
	case Boolean:
		return obj.value
	default:
		return true
	}
}

func (ev *Evaluator) infix(node ast.Infix) (Object, error) {
	lhs, err := ev.expression(node.LHS)
	if err != nil {
		return nil, err
	}
	rhs, err := ev.expression(node.RHS)
	if err != nil {
		return nil, err
	}
	switch node.Operator.Type {
	case token.Plus:
		// Addition evaluation lets the left hand side expression operand to control whether the addition should be
		// considered a concatenation of an addition of numbers.
		switch lhs := lhs.(type) {
		case String:
			return ev.concat(lhs, rhs)
		case Number:
			return ev.add(lhs, rhs)
		default:
			return nil, fmt.Errorf(
				"%w: invalid addition operand type %s",
				ErrRuntimeFault,
				reflect.TypeOf(lhs),
			)
		}
	case token.Minus:
		return ev.subtract(lhs, rhs)
	case token.Asterisk:
		return ev.multiply(lhs, rhs)
	case token.Slash:
		return ev.divide(lhs, rhs)
	default:
		return nil, fmt.Errorf(
			"%w: unexpected infix operator %s",
			ErrRuntimeFault,
			node.Operator.Lexeme,
		)
	}
}

func (ev *Evaluator) concat(lhs, rhs Object) (String, error) {
	lhn, ok := lhs.(String)
	if !ok {
		return String{}, fmt.Errorf("%w: %s is not a string", ErrRuntimeFault, reflect.TypeOf(lhs))
	}
	rhn, ok := rhs.(String)
	if !ok {
		return String{}, fmt.Errorf("%w: %s is not a string", ErrRuntimeFault, reflect.TypeOf(rhs))
	}
	return String{lhn.value + rhn.value}, nil
}

func (ev *Evaluator) add(lhs, rhs Object) (Number, error) {
	lhn, ok := lhs.(Number)
	if !ok {
		return Number{}, fmt.Errorf("%w: %s is not a number", ErrRuntimeFault, reflect.TypeOf(lhs))
	}
	rhn, ok := rhs.(Number)
	if !ok {
		return Number{}, fmt.Errorf("%w: %s is not a number", ErrRuntimeFault, reflect.TypeOf(rhs))
	}
	return Number{lhn.value + rhn.value}, nil
}

func (ev *Evaluator) subtract(lhs, rhs Object) (Number, error) {
	lhn, ok := lhs.(Number)
	if !ok {
		return Number{}, fmt.Errorf("%w: %s is not a number", ErrRuntimeFault, reflect.TypeOf(lhs))
	}
	rhn, ok := rhs.(Number)
	if !ok {
		return Number{}, fmt.Errorf("%w: %s is not a number", ErrRuntimeFault, reflect.TypeOf(rhs))
	}
	return Number{lhn.value - rhn.value}, nil
}

func (ev *Evaluator) multiply(lhs, rhs Object) (Number, error) {
	lhn, ok := lhs.(Number)
	if !ok {
		return Number{}, fmt.Errorf("%w: %s is not a number", ErrRuntimeFault, reflect.TypeOf(lhs))
	}
	rhn, ok := rhs.(Number)
	if !ok {
		return Number{}, fmt.Errorf("%w: %s is not a number", ErrRuntimeFault, reflect.TypeOf(rhs))
	}
	return Number{lhn.value * rhn.value}, nil
}

func (ev *Evaluator) divide(lhs, rhs Object) (Number, error) {
	lhn, ok := lhs.(Number)
	if !ok {
		return Number{}, fmt.Errorf("%w: %s is not a number", ErrRuntimeFault, reflect.TypeOf(lhs))
	}
	rhn, ok := rhs.(Number)
	if !ok {
		return Number{}, fmt.Errorf("%w: %s is not a number", ErrRuntimeFault, reflect.TypeOf(rhs))
	}
	if rhn.value == 0 {
		// Division by zero is undefined and counts as an erroneous input.
		return Number{}, fmt.Errorf("%w: tried to divide by zero", ErrRuntimeFault)
	}
	return Number{lhn.value / rhn.value}, nil
}

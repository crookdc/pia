package squeak

import (
	"errors"
	"fmt"
	"github.com/crookdc/pia/squeak/internal/ast"
	"github.com/crookdc/pia/squeak/internal/token"
	"io"
	"math"
	"reflect"
)

var ErrRuntimeFault = errors.New("runtime error")

// Object is a broad interface for any data that a Squeak script can process. It does not provide any interface beyond
// the standard library fmt.Stringer, but passing around types of this interface around the interpreter obfuscates the
// meaning behind the value. Hence, this interface is largely just for clearer naming.
type Object interface {
	fmt.Stringer
}

// Number is an Object representing a numerical value internally represented as a float64. In Squeak, the notion of
// integers only exists in the lexical and parsing phase. During evaluation, all numerical objects are represented with
// this struct.
type Number struct {
	value float64
}

func (i Number) String() string {
	if i.value == math.Floor(i.value) {
		return fmt.Sprintf("%d", int(i.value))
	}
	return fmt.Sprintf("%.2f", i.value)
}

// String is an Object representing a textual value.
type String struct {
	value string
}

func (s String) String() string {
	return s.value
}

// Boolean is an Object representing a boolean value.
type Boolean struct {
	value bool
}

func (b Boolean) String() string {
	if b.value {
		return "true"
	}
	return "false"
}

type EnvironmentOpt func(*Environment)

func Parent(parent *Environment) EnvironmentOpt {
	return func(env *Environment) {
		env.parent = parent
	}
}

func Prefill(k string, v Object) EnvironmentOpt {
	return func(env *Environment) {
		if env.tbl == nil {
			env.tbl = make(map[string]Object)
		}
		env.tbl[k] = v
	}
}

func NewEnvironment(opts ...EnvironmentOpt) *Environment {
	env := &Environment{
		tbl: make(map[string]Object),
	}
	for _, opt := range opts {
		opt(env)
	}
	return env
}

// Environment is a table of contents for runtime variables that exposes an API to interface with the current
// environment correctly. It also supports the concepts of hierarchical environments which enables scoping of variables.
type Environment struct {
	parent *Environment
	tbl    map[string]Object
}

// Resolve returns the current value stored within the environment for the provided key. If the key cannot be resolved
// for the immediate scope (the table of variables that is stored within the environment) then the parent environment is
// invoked to resolve the same key within its immediate scope. This call chain continues until the key is successfully
// resolved or the next parent is a nil value, in which case a non-nil error is returned.
func (env *Environment) Resolve(k string) (Object, error) {
	val, ok := env.tbl[k]
	if ok {
		return val, nil
	}
	if env.parent != nil {
		return env.parent.Resolve(k)
	}
	return nil, fmt.Errorf("%w: cannot resolve key %s", ErrRuntimeFault, k)
}

// Declare sets the provided value for the provided key in the immediate scope.
func (env *Environment) Declare(k string, v Object) {
	env.tbl[k] = v
}

// Assign sets a new value for an already declared variable in the immediate scope. If the key cannot be resolved in the
// immediate scope then the parent is invoked. This call chain continues until the assignment is successful or until the
// next parent is nil, in which case it returns a non-nil error.
func (env *Environment) Assign(k string, v Object) error {
	_, ok := env.tbl[k]
	if ok {
		env.tbl[k] = v
		return nil
	}
	if env.parent != nil {
		return env.parent.Assign(k, v)
	}
	return fmt.Errorf("%w: cannot resolve assignment target", ErrRuntimeFault)
}

type Evaluator struct {
	env *Environment
	out io.Writer
}

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
	case ast.Print:
		obj, err := ev.expression(node.Expression)
		if err != nil {
			return err
		}
		_, err = io.WriteString(ev.out, obj.String())
		return err
	case ast.Var:
		if node.Initializer == nil {
			ev.env.Declare(node.Name.Lexeme, nil)
			return nil
		}
		val, err := ev.expression(node.Initializer)
		if err != nil {
			return err
		}
		ev.env.Declare(node.Name.Lexeme, val)
		return nil
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
	case ast.FloatLiteral:
		return Number{node.Float}, nil
	case ast.StringLiteral:
		return String{node.String}, nil
	case ast.BooleanLiteral:
		return Boolean{node.Boolean}, nil
	case ast.NilLiteral:
		return nil, nil
	case ast.Grouping:
		return ev.expression(node.Group)
	case ast.Prefix:
		return ev.prefix(node)
	case ast.Infix:
		return ev.infix(node)
	case ast.Assignment:
		val, err := ev.expression(node.Value)
		if err != nil {
			return nil, err
		}
		if err := ev.env.Assign(node.Name.Lexeme, val); err != nil {
			return nil, err
		}
		return val, nil
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
		// Addition evaluation lets the left hand side expression operand control whether the addition should be
		// considered a concatenation or an addition of numbers.
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
	case token.Less:
		return ev.isLessThan(lhs, rhs)
	case token.LessEqual:
		lt, err := ev.isLessThan(lhs, rhs)
		if err != nil {
			return nil, err
		}
		if lt.value {
			return lt, nil
		}
		return ev.isEqual(lhs, rhs)
	case token.Greater:
		return ev.isGreaterThan(lhs, rhs)
	case token.GreaterEqual:
		gt, err := ev.isGreaterThan(lhs, rhs)
		if err != nil {
			return nil, err
		}
		if gt.value {
			return gt, nil
		}
		return ev.isEqual(lhs, rhs)
	case token.Equals:
		return ev.isEqual(lhs, rhs)
	case token.NotEquals:
		eq, err := ev.isEqual(lhs, rhs)
		eq.value = !eq.value
		return eq, err
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

func (ev *Evaluator) isLessThan(lhs, rhs Object) (Boolean, error) {
	lhn, ok := lhs.(Number)
	if !ok {
		return Boolean{}, fmt.Errorf("%w: %s is not a number", ErrRuntimeFault, reflect.TypeOf(lhs))
	}
	rhn, ok := rhs.(Number)
	if !ok {
		return Boolean{}, fmt.Errorf("%w: %s is not a number", ErrRuntimeFault, reflect.TypeOf(rhs))
	}
	return Boolean{lhn.value < rhn.value}, nil
}

func (ev *Evaluator) isGreaterThan(lhs, rhs Object) (Boolean, error) {
	lhn, ok := lhs.(Number)
	if !ok {
		return Boolean{}, fmt.Errorf("%w: %s is not a number", ErrRuntimeFault, reflect.TypeOf(lhs))
	}
	rhn, ok := rhs.(Number)
	if !ok {
		return Boolean{}, fmt.Errorf("%w: %s is not a number", ErrRuntimeFault, reflect.TypeOf(rhs))
	}
	return Boolean{lhn.value > rhn.value}, nil
}

func (ev *Evaluator) isEqual(lhs, rhs Object) (Boolean, error) {
	if lhs == nil && rhs == nil {
		return Boolean{true}, nil
	}
	if reflect.TypeOf(lhs) != reflect.TypeOf(rhs) {
		return Boolean{}, fmt.Errorf(
			"%w: cannot compare equality between %s with %s",
			ErrRuntimeFault,
			reflect.TypeOf(lhs),
			reflect.TypeOf(rhs),
		)
	}
	return Boolean{lhs == rhs}, nil
}

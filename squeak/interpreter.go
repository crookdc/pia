package squeak

import (
	"errors"
	"fmt"
	"github.com/crookdc/pia/squeak/ast"
	"github.com/crookdc/pia/squeak/token"
	"io"
	"reflect"
	"strings"
)

var ErrRuntimeFault = errors.New("runtime error")

type Unwinding struct {
	Source token.Token
	Value  Object
}

// Object is a broad interface for any data that a Squeak script can process. It does not provide any interface beyond
// the standard library fmt.Stringer, but passing around types of this interface around the interpreter obfuscates the
// meaning behind the value. Hence, this interface is largely just for clearer naming.
type Object interface {
	fmt.Stringer
}

type Callable interface {
	Object
	Arity() int
	Call(*Interpreter, ...Object) (Object, error)
}

// Function is the callable equivalent of [ast.Function].
type Function struct {
	Declaration ast.Function
}

func (f Function) String() string {
	return fmt.Sprintf("function:%s", f.Declaration.Name.Lexeme)
}

func (f Function) Arity() int {
	return len(f.Declaration.Params)
}

func (f Function) Call(in *Interpreter, objs ...Object) (Object, error) {
	scope := NewEnvironment(Parent(in.global))
	for i, param := range f.Declaration.Params {
		scope.Declare(param.Lexeme, objs[i])
	}
	uw, err := in.block(scope, f.Declaration.Body.Body)
	if err != nil {
		return nil, err
	}
	if uw == nil {
		return nil, nil
	}
	if uw.Source.Type != token.Return {
		return nil, fmt.Errorf("%w: unexpected unwinding source %s", ErrRuntimeFault, uw.Source.Lexeme)
	}
	return uw.Value, nil
}

// Number is an Object representing a numerical value internally represented as a float64. In Squeak, the notion of
// integers only exists in the lexical and parsing phase. During evaluation, all numerical objects are represented with
// this struct.
type Number struct {
	value float64
}

func (i Number) String() string {
	return strings.TrimRight(fmt.Sprintf("%f", i.value), "0")
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

// NewInterpreter constructs an interpreter with a prefilled runtime in its global scope. The caller is responsible for
// supplying a valid [io.Writer] which is used as the standard output stream. While the caller is allowed to provide a
// nil [io.Writer], it is discouraged as any usage of the standard output will result in a panic.
func NewInterpreter(out io.Writer) *Interpreter {
	global := NewEnvironment(
		Prefill("print", PrintBuiltin{}),
	)
	return &Interpreter{
		global: global,
		scope:  global,
		out:    out,
	}
}

type Interpreter struct {
	global *Environment
	scope  *Environment
	out    io.Writer
}

func (in *Interpreter) Execute(program []ast.StatementNode) error {
	for _, stmt := range program {
		if _, err := in.statement(stmt); err != nil {
			return err
		}
	}
	return nil
}

// statement executes the provided statement node within the current context of the interpreter. Statements do not
// generally evaluate to a value. Some statements such as [ast.Return] changes the control flow drastically, those cases
// are not directly handled by this method. Instead, whenever an unwinding statement is encountered then a non-nil value
// of Unwinding is returned which is expected to be processed properly by some caller in the call stack.
func (in *Interpreter) statement(node ast.StatementNode) (*Unwinding, error) {
	switch node := node.(type) {
	case ast.ExpressionStatement:
		_, err := in.expression(node.Expression)
		return nil, err
	case ast.Declaration:
		if node.Initializer == nil {
			in.scope.Declare(node.Name.Lexeme, nil)
			return nil, nil
		}
		val, err := in.expression(node.Initializer)
		if err != nil {
			return nil, err
		}
		in.scope.Declare(node.Name.Lexeme, val)
		return nil, nil
	case ast.Block:
		return in.block(NewEnvironment(Parent(in.scope)), node.Body)
	case ast.If:
		cnd, err := in.expression(node.Condition)
		if err != nil {
			return nil, err
		}
		if in.truthy(cnd) {
			return in.statement(node.Then)
		}
		if node.Else != nil {
			return in.statement(node.Else)
		}
		return nil, nil
	case ast.While:
		cnd, err := in.expression(node.Condition)
		if err != nil {
			return nil, err
		}
		for in.truthy(cnd) {
			uw, err := in.statement(node.Body)
			if err != nil {
				return nil, err
			}
			if uw != nil {
				if uw.Source.Type == token.Break {
					break
				}
				if uw.Source.Type != token.Continue {
					return uw, nil
				}
			}
			cnd, err = in.expression(node.Condition)
			if err != nil {
				return nil, err
			}
		}
		return nil, nil
	case ast.Noop:
		// In the future it might be a good idea to restructure the AST so that it does not contain any [ast.Noop].
		return nil, nil
	case ast.Function:
		in.scope.Declare(node.Name.Lexeme, Function{Declaration: node})
		return nil, nil
	case ast.Return:
		uw := &Unwinding{
			Source: token.Token{
				Type:   token.Return,
				Lexeme: "return",
			},
		}
		if node.Expression == nil {
			return uw, nil
		}
		val, err := in.expression(node.Expression)
		if err != nil {
			return nil, err
		}
		uw.Value = val
		return uw, nil
	case ast.Break:
		return &Unwinding{
			Source: token.Token{
				Type:   token.Break,
				Lexeme: "break",
			},
		}, nil
	case ast.Continue:
		return &Unwinding{
			Source: token.Token{
				Type:   token.Continue,
				Lexeme: "continue",
			},
		}, nil
	default:
		return nil, fmt.Errorf(
			"%w: unexpected statement type %s",
			ErrRuntimeFault,
			reflect.TypeOf(node),
		)
	}
}

func (in *Interpreter) block(scope *Environment, block []ast.StatementNode) (*Unwinding, error) {
	prev := in.scope
	defer func() {
		in.scope = prev
	}()
	in.scope = scope
	for _, stmt := range block {
		uw, err := in.statement(stmt)
		if err != nil {
			return nil, err
		}
		if uw != nil {
			return uw, nil
		}
	}
	return nil, nil
}

func (in *Interpreter) expression(node ast.ExpressionNode) (Object, error) {
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
		return in.expression(node.Group)
	case ast.Prefix:
		return in.prefix(node)
	case ast.Infix:
		return in.infix(node)
	case ast.Variable:
		return in.scope.Resolve(node.Name.Lexeme)
	case ast.Assignment:
		val, err := in.expression(node.Value)
		if err != nil {
			return nil, err
		}
		if err := in.scope.Assign(node.Name.Lexeme, val); err != nil {
			return nil, err
		}
		return val, nil
	case ast.Logical:
		return in.logical(node)
	case ast.Call:
		return in.call(node)
	default:
		return nil, fmt.Errorf(
			"%w: unexpected expression type %s",
			ErrRuntimeFault,
			reflect.TypeOf(node),
		)
	}
}

func (in *Interpreter) call(node ast.Call) (Object, error) {
	fn, err := in.expression(node.Callee)
	if err != nil {
		return nil, err
	}
	switch fn := fn.(type) {
	case Callable:
		if fn.Arity() != len(node.Args) {
			return nil, fmt.Errorf(
				"function accepts %d parameters but was provided %d arguments",
				fn.Arity(),
				len(node.Args),
			)
		}
		var args []Object
		for _, expr := range node.Args {
			arg, err := in.expression(expr)
			if err != nil {
				return nil, err
			}
			args = append(args, arg)
		}
		return fn.Call(in, args...)
	default:
		return nil, fmt.Errorf("tried to call non-callable: %s", fn)
	}
}

func (in *Interpreter) logical(node ast.Logical) (Object, error) {
	left, err := in.expression(node.LHS)
	if err != nil {
		return nil, err
	}
	switch node.Operator.Type {
	case token.And:
		if in.falsy(left) {
			return left, nil
		}
		return in.expression(node.RHS)
	case token.Or:
		if in.truthy(left) {
			return left, nil
		}
		return in.expression(node.RHS)
	default:
		return nil, fmt.Errorf(
			"%w: unrecognized logical operator: %s",
			ErrRuntimeFault,
			node.Operator.Lexeme,
		)
	}
}

func (in *Interpreter) prefix(node ast.Prefix) (Object, error) {
	obj, err := in.expression(node.Target)
	if err != nil {
		return nil, err
	}
	switch node.Operator.Type {
	case token.Bang:
		return Boolean{!in.truthy(obj)}, nil
	case token.Minus:
		return in.multiply(obj, Number{-1})
	default:
		return nil, fmt.Errorf(
			"%w: unexpected prefix operator %s",
			ErrRuntimeFault,
			node.Operator.Lexeme,
		)
	}
}

func (in *Interpreter) truthy(obj Object) bool {
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

func (in *Interpreter) falsy(obj Object) bool {
	return !in.truthy(obj)
}

func (in *Interpreter) infix(node ast.Infix) (Object, error) {
	lhs, err := in.expression(node.LHS)
	if err != nil {
		return nil, err
	}
	rhs, err := in.expression(node.RHS)
	if err != nil {
		return nil, err
	}
	switch node.Operator.Type {
	case token.Plus:
		// Addition evaluation lets the left hand side expression operand control whether the addition should be
		// considered a concatenation or an addition of numbers.
		switch lhs := lhs.(type) {
		case String:
			return in.concat(lhs, rhs)
		case Number:
			return in.add(lhs, rhs)
		default:
			return nil, fmt.Errorf(
				"%w: invalid addition operand type %s",
				ErrRuntimeFault,
				reflect.TypeOf(lhs),
			)
		}
	case token.Minus:
		return in.subtract(lhs, rhs)
	case token.Asterisk:
		return in.multiply(lhs, rhs)
	case token.Slash:
		return in.divide(lhs, rhs)
	case token.Less:
		return in.isLessThan(lhs, rhs)
	case token.LessEqual:
		lt, err := in.isLessThan(lhs, rhs)
		if err != nil {
			return nil, err
		}
		if lt.value {
			return lt, nil
		}
		return in.isEqual(lhs, rhs)
	case token.Greater:
		return in.isGreaterThan(lhs, rhs)
	case token.GreaterEqual:
		gt, err := in.isGreaterThan(lhs, rhs)
		if err != nil {
			return nil, err
		}
		if gt.value {
			return gt, nil
		}
		return in.isEqual(lhs, rhs)
	case token.Equals:
		return in.isEqual(lhs, rhs)
	case token.NotEquals:
		eq, err := in.isEqual(lhs, rhs)
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

func (in *Interpreter) concat(lhs, rhs Object) (String, error) {
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

func (in *Interpreter) add(lhs, rhs Object) (Number, error) {
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

func (in *Interpreter) subtract(lhs, rhs Object) (Number, error) {
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

func (in *Interpreter) multiply(lhs, rhs Object) (Number, error) {
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

func (in *Interpreter) divide(lhs, rhs Object) (Number, error) {
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

func (in *Interpreter) isLessThan(lhs, rhs Object) (Boolean, error) {
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

func (in *Interpreter) isGreaterThan(lhs, rhs Object) (Boolean, error) {
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

func (in *Interpreter) isEqual(lhs, rhs Object) (Boolean, error) {
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

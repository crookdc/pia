package squeak

import (
	"errors"
	"fmt"
	"github.com/crookdc/pia/squeak/ast"
	"github.com/crookdc/pia/squeak/token"
	"io"
	"reflect"
)

var (
	ErrRuntimeFault = errors.New("runtime error")
	ErrNotCallable  = fmt.Errorf("%w: not callable", ErrRuntimeFault)
)

type unwinder struct {
	source token.Token
	value  Object
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
	return nil, fmt.Errorf("%w: cannot resolve key: %s", ErrRuntimeFault, k)
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
	return fmt.Errorf("%w: cannot resolve assignment target: %s", ErrRuntimeFault, k)
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
		uw, err := in.execute(stmt)
		if err != nil {
			return err
		}
		if uw != nil {
			// Unwinders should never bubble up all the way here, if they have then that means the unwinder never passed
			// through a caller that could correctly handle it, which is considered an erroneous state.
			return fmt.Errorf("%w: unexpected unwinder: %s", ErrRuntimeFault, uw.source.Lexeme)
		}
	}
	return nil
}

// execute runs the provided statement node within the current context of the interpreter. Statements do not generally
// evaluate to a value. Some statements such as [ast.Return] changes the control flow drastically, those cases are not
// handled by this method. Instead, whenever an unwinding statement is encountered then a non-nil value of unwinder is
// returned which is expected to be processed properly by some caller in the call stack.
func (in *Interpreter) execute(stmt ast.StatementNode) (*unwinder, error) {
	switch stmt := stmt.(type) {
	case ast.ExpressionStatement:
		_, err := in.evaluate(stmt.Expression)
		return nil, err
	case ast.Declaration:
		return nil, in.declaration(stmt)
	case ast.Block:
		return in.block(NewEnvironment(Parent(in.scope)), stmt.Body)
	case ast.If:
		return in.branching(stmt)
	case ast.While:
		return in.loop(stmt)
	case ast.Noop:
		// In the future it might be a good idea to restructure the AST so that it does not contain any [ast.Noop].
		return nil, nil
	case ast.Function:
		in.scope.Declare(stmt.Name.Lexeme, Function{
			declaration: stmt,
			closure:     in.scope,
		})
		return nil, nil
	case ast.Return, ast.Break, ast.Continue:
		// Perhaps these three types, which all share the common behaviour of unwinding the call stack of the
		// interpreter in one way or the other should be grouped with another 'subtype' of [ast.Statement] such as for
		// example ast.Unwinder. However, for as long as we only have three of these statement types I do not see any
		// harm in handling them directly like we do now rather than abstracting things away. On the contrary, I believe
		// that abstracting it away prematurely would just cause confusion.
		return in.unwinder(stmt)
	default:
		return nil, fmt.Errorf(
			"%w: unexpected statement type: %T",
			ErrRuntimeFault,
			stmt,
		)
	}
}

func (in *Interpreter) declaration(stmt ast.Declaration) error {
	if stmt.Initializer == nil {
		in.scope.Declare(stmt.Name.Lexeme, nil)
		return nil
	}
	val, err := in.evaluate(stmt.Initializer)
	if err != nil {
		return err
	}
	in.scope.Declare(stmt.Name.Lexeme, val)
	return nil
}

func (in *Interpreter) branching(stmt ast.If) (*unwinder, error) {
	cnd, err := in.evaluate(stmt.Condition)
	if err != nil {
		return nil, err
	}
	if in.truthy(cnd) {
		return in.execute(stmt.Then)
	}
	if stmt.Else != nil {
		return in.execute(stmt.Else)
	}
	return nil, nil
}

func (in *Interpreter) loop(stmt ast.While) (*unwinder, error) {
	cnd, err := in.evaluate(stmt.Condition)
	if err != nil {
		return nil, err
	}
	for in.truthy(cnd) {
		uw, err := in.execute(stmt.Body)
		if err != nil {
			return nil, err
		}
		if uw != nil {
			if uw.source.Type == token.Break {
				break
			}
			if uw.source.Type != token.Continue {
				return uw, nil
			}
		}
		cnd, err = in.evaluate(stmt.Condition)
		if err != nil {
			return nil, err
		}
	}
	return nil, nil
}

func (in *Interpreter) unwinder(stmt ast.StatementNode) (*unwinder, error) {
	switch stmt := stmt.(type) {
	case ast.Return:
		uw := &unwinder{
			source: token.Token{
				Type:   token.Return,
				Lexeme: "return",
			},
		}
		if stmt.Expression == nil {
			return uw, nil
		}
		// The return statement is the only unwinding statement that can also evaluate an evaluate which shall be
		// returned to the caller. If the return statement does indeed contain an evaluate then it is set on the value
		// field in the unwinder value.
		val, err := in.evaluate(stmt.Expression)
		if err != nil {
			return nil, err
		}
		uw.value = val
		return uw, nil
	case ast.Break:
		return &unwinder{
			source: token.Token{
				Type:   token.Break,
				Lexeme: "break",
			},
		}, nil
	case ast.Continue:
		return &unwinder{
			source: token.Token{
				Type:   token.Continue,
				Lexeme: "continue",
			},
		}, nil
	default:
		panic(fmt.Errorf("unwinder executor called with invalid statement type: %T", stmt))
	}
}

func (in *Interpreter) block(scope *Environment, block []ast.StatementNode) (*unwinder, error) {
	prev := in.scope
	defer func() {
		in.scope = prev
	}()
	in.scope = scope
	for _, stmt := range block {
		uw, err := in.execute(stmt)
		if err != nil {
			return nil, err
		}
		if uw != nil {
			return uw, nil
		}
	}
	return nil, nil
}

func (in *Interpreter) evaluate(expr ast.ExpressionNode) (Object, error) {
	switch expr := expr.(type) {
	case ast.IntegerLiteral:
		return Number{float64(expr.Integer)}, nil
	case ast.FloatLiteral:
		return Number{expr.Float}, nil
	case ast.StringLiteral:
		return String{expr.String}, nil
	case ast.BooleanLiteral:
		return Boolean{expr.Boolean}, nil
	case ast.NilLiteral:
		return nil, nil
	case ast.Grouping:
		return in.evaluate(expr.Group)
	case ast.Prefix:
		return in.prefix(expr)
	case ast.Infix:
		return in.infix(expr)
	case ast.Variable:
		return in.scope.Resolve(expr.Name.Lexeme)
	case ast.Assignment:
		val, err := in.evaluate(expr.Value)
		if err != nil {
			return nil, err
		}
		if err := in.scope.Assign(expr.Name.Lexeme, val); err != nil {
			return nil, err
		}
		return val, nil
	case ast.Logical:
		return in.logical(expr)
	case ast.Call:
		return in.call(expr)
	default:
		return nil, fmt.Errorf(
			"%w: unexpected evaluate type: %T",
			ErrRuntimeFault,
			expr,
		)
	}
}

func (in *Interpreter) call(node ast.Call) (Object, error) {
	fn, err := in.evaluate(node.Callee)
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
			arg, err := in.evaluate(expr)
			if err != nil {
				return nil, err
			}
			args = append(args, arg)
		}
		return fn.Call(in, args...)
	default:
		return nil, fmt.Errorf("%w: %s", ErrNotCallable, fn)
	}
}

func (in *Interpreter) logical(node ast.Logical) (Object, error) {
	left, err := in.evaluate(node.LHS)
	if err != nil {
		return nil, err
	}
	switch node.Operator.Type {
	case token.And:
		if in.falsy(left) {
			return left, nil
		}
		return in.evaluate(node.RHS)
	case token.Or:
		if in.truthy(left) {
			return left, nil
		}
		return in.evaluate(node.RHS)
	default:
		return nil, fmt.Errorf(
			"%w: unrecognized logical operator: %s",
			ErrRuntimeFault,
			node.Operator.Lexeme,
		)
	}
}

func (in *Interpreter) prefix(node ast.Prefix) (Object, error) {
	obj, err := in.evaluate(node.Target)
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
			"%w: unexpected prefix operator: %s",
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
	lhs, err := in.evaluate(node.LHS)
	if err != nil {
		return nil, err
	}
	rhs, err := in.evaluate(node.RHS)
	if err != nil {
		return nil, err
	}
	switch node.Operator.Type {
	case token.Plus:
		// Addition evaluation lets the left hand side evaluate operand control whether the addition should be
		// considered a concatenation or an addition of numbers.
		switch lhs := lhs.(type) {
		case String:
			return in.concat(lhs, rhs)
		case Number:
			return in.add(lhs, rhs)
		default:
			return nil, fmt.Errorf(
				"%w: invalid addition operand type: %T",
				ErrRuntimeFault,
				lhs,
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
			"%w: unexpected infix operator: %s",
			ErrRuntimeFault,
			node.Operator.Lexeme,
		)
	}
}

func (in *Interpreter) concat(lhs, rhs Object) (String, error) {
	lhn, ok := lhs.(String)
	if !ok {
		return String{}, fmt.Errorf("%w: %T is not a string", ErrRuntimeFault, lhs)
	}
	rhn, ok := rhs.(String)
	if !ok {
		return String{}, fmt.Errorf("%w: %T is not a string", ErrRuntimeFault, rhs)
	}
	return String{lhn.value + rhn.value}, nil
}

func (in *Interpreter) add(lhs, rhs Object) (Number, error) {
	lhn, ok := lhs.(Number)
	if !ok {
		return Number{}, fmt.Errorf("%w: %T is not a number", ErrRuntimeFault, lhs)
	}
	rhn, ok := rhs.(Number)
	if !ok {
		return Number{}, fmt.Errorf("%w: %T is not a number", ErrRuntimeFault, rhs)
	}
	return Number{lhn.value + rhn.value}, nil
}

func (in *Interpreter) subtract(lhs, rhs Object) (Number, error) {
	lhn, ok := lhs.(Number)
	if !ok {
		return Number{}, fmt.Errorf("%w: %T is not a number", ErrRuntimeFault, lhs)
	}
	rhn, ok := rhs.(Number)
	if !ok {
		return Number{}, fmt.Errorf("%w: %T is not a number", ErrRuntimeFault, rhs)
	}
	return Number{lhn.value - rhn.value}, nil
}

func (in *Interpreter) multiply(lhs, rhs Object) (Number, error) {
	lhn, ok := lhs.(Number)
	if !ok {
		return Number{}, fmt.Errorf("%w: %T is not a number", ErrRuntimeFault, lhs)
	}
	rhn, ok := rhs.(Number)
	if !ok {
		return Number{}, fmt.Errorf("%w: %T is not a number", ErrRuntimeFault, rhs)
	}
	return Number{lhn.value * rhn.value}, nil
}

func (in *Interpreter) divide(lhs, rhs Object) (Number, error) {
	lhn, ok := lhs.(Number)
	if !ok {
		return Number{}, fmt.Errorf("%w: %T is not a number", ErrRuntimeFault, lhs)
	}
	rhn, ok := rhs.(Number)
	if !ok {
		return Number{}, fmt.Errorf("%w: %T is not a number", ErrRuntimeFault, rhs)
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
		return Boolean{}, fmt.Errorf("%w: %T is not a number", ErrRuntimeFault, lhs)
	}
	rhn, ok := rhs.(Number)
	if !ok {
		return Boolean{}, fmt.Errorf("%w: %T is not a number", ErrRuntimeFault, rhs)
	}
	return Boolean{lhn.value < rhn.value}, nil
}

func (in *Interpreter) isGreaterThan(lhs, rhs Object) (Boolean, error) {
	lhn, ok := lhs.(Number)
	if !ok {
		return Boolean{}, fmt.Errorf("%w: %T is not a number", ErrRuntimeFault, lhs)
	}
	rhn, ok := rhs.(Number)
	if !ok {
		return Boolean{}, fmt.Errorf("%w: %T is not a number", ErrRuntimeFault, rhs)
	}
	return Boolean{lhn.value > rhn.value}, nil
}

func (in *Interpreter) isEqual(lhs, rhs Object) (Boolean, error) {
	if lhs == nil && rhs == nil {
		return Boolean{true}, nil
	}
	if reflect.TypeOf(lhs) != reflect.TypeOf(rhs) {
		return Boolean{}, fmt.Errorf(
			"%w: cannot compare equality between %T with %T",
			ErrRuntimeFault,
			lhs,
			rhs,
		)
	}
	return Boolean{lhs == rhs}, nil
}

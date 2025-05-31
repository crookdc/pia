package squeak

import (
	"fmt"
	"github.com/crookdc/pia/squeak/ast"
	"github.com/crookdc/pia/squeak/token"
	"strings"
)

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
	declaration ast.Function
	closure     *Environment
}

func (f Function) String() string {
	return fmt.Sprintf("function:%s", f.declaration.Name.Lexeme)
}

func (f Function) Arity() int {
	return len(f.declaration.Params)
}

func (f Function) Call(in *Interpreter, objs ...Object) (Object, error) {
	scope := NewEnvironment(Parent(f.closure))
	for i, param := range f.declaration.Params {
		scope.Declare(param.Lexeme, objs[i])
	}
	uw, err := in.block(scope, f.declaration.Body.Body)
	if err != nil {
		return nil, err
	}
	if uw == nil {
		return nil, nil
	}
	if uw.source.Type != token.Return {
		return nil, fmt.Errorf("%w: unexpected unwinding source %s", ErrRuntimeFault, uw.source.Lexeme)
	}
	return uw.value, nil
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

// List is a single Object containing a collection of Object values.
type List struct {
	slice []Object
}

func (l List) String() string {
	items := make([]string, len(l.slice))
	for i := range l.slice {
		items[i] = l.slice[i].String()
	}
	return fmt.Sprintf("[%s]", strings.Join(items, ","))
}

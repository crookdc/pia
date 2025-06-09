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
	Clone() Object
}

type Instance interface {
	Object
	Get(string) Object
	Put(string, Object) Object
}

// ObjectInstance is an object instance, which consists of a collection of named data as well as behaviours coupled to the
// data.
type ObjectInstance struct {
	Properties map[string]Object
}

func (i *ObjectInstance) String() string {
	sb := strings.Builder{}
	sb.WriteString("Object {")
	for k, v := range i.Properties {
		sb.WriteString(fmt.Sprintf("%s: %s", k, v.String()))
	}
	sb.WriteString("}")
	return sb.String()
}

func (i *ObjectInstance) Clone() Object {
	props := make(map[string]Object)
	for k, v := range i.Properties {
		props[k] = v.Clone()
	}
	return &ObjectInstance{Properties: props}
}

func (i *ObjectInstance) Get(s string) Object {
	return i.Properties[s]
}

func (i *ObjectInstance) Put(s string, object Object) Object {
	i.Properties[s] = object
	return object
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

func (f Function) Clone() Object {
	return Function{
		declaration: f.declaration,
		closure:     f.closure,
	}
}

func (f Function) Arity() int {
	return len(f.declaration.Params)
}

func (f Function) Call(in *Interpreter, args ...Object) (Object, error) {
	scope := NewEnvironment(Parent(f.closure))
	for i, param := range f.declaration.Params {
		scope.Declare(param.Lexeme, args[i])
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

type BoundMethod struct {
	Method
	this Instance
}

func (bm BoundMethod) Arity() int {
	return len(bm.declaration.Params)
}

func (bm BoundMethod) Call(in *Interpreter, args ...Object) (Object, error) {
	closure := NewEnvironment(Parent(in.global), Prefill("this", bm.this))
	scope := NewEnvironment(Parent(closure))
	for i, param := range bm.declaration.Params {
		scope.Declare(param.Lexeme, args[i])
	}
	uw, err := in.block(scope, bm.declaration.Body.Body)
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

type Method struct {
	declaration ast.Method
}

func (m Method) String() string {
	return fmt.Sprintf("method")
}

func (m Method) Clone() Object {
	return Method{declaration: m.declaration}
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

func (i Number) Clone() Object {
	return Number{
		value: i.value,
	}
}

// String is an Object representing a textual value.
type String struct {
	value string
}

func (s String) String() string {
	return s.value
}

func (s String) Clone() Object {
	return String{value: s.value}
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

func (b Boolean) Clone() Object {
	return Boolean{value: b.value}
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

func (l List) Clone() Object {
	clone := make([]Object, len(l.slice))
	for i, v := range l.slice {
		clone[i] = v.Clone()
	}
	return List{slice: clone}
}

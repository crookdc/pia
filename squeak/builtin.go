package squeak

import "fmt"

type PrintBuiltin struct{}

func (p PrintBuiltin) String() string {
	return "builtin:print"
}

func (p PrintBuiltin) Arity() int {
	return 1
}

func (p PrintBuiltin) Call(in *Interpreter, args ...Object) (Object, error) {
	_, err := fmt.Fprint(in.out, args[0].String())
	if err != nil {
		return nil, err
	}
	return nil, nil
}

func (p PrintBuiltin) Clone() Object {
	return PrintBuiltin{}
}

type PrintlnBuiltin struct{}

func (p PrintlnBuiltin) String() string {
	return "builtin:println"
}

func (p PrintlnBuiltin) Clone() Object {
	return PrintlnBuiltin{}
}

func (p PrintlnBuiltin) Arity() int {
	return 1
}

func (p PrintlnBuiltin) Call(in *Interpreter, args ...Object) (Object, error) {
	_, err := fmt.Fprintln(in.out, args[0].String())
	if err != nil {
		return nil, err
	}
	return nil, nil
}

type LengthBuiltin struct{}

func (l LengthBuiltin) String() string {
	return "builtin:length"
}

func (l LengthBuiltin) Arity() int {
	return 1
}

func (l LengthBuiltin) Call(_ *Interpreter, args ...Object) (Object, error) {
	list, ok := args[0].(*List)
	if !ok {
		return nil, fmt.Errorf(
			"%w: %T is not a list",
			ErrIllegalArgument,
			args[0],
		)
	}
	return Number{float64(len(list.slice))}, nil
}

func (l LengthBuiltin) Clone() Object {
	return LengthBuiltin{}
}

type CloneBuiltin struct{}

func (c CloneBuiltin) String() string {
	return "builtin:clone"
}

func (c CloneBuiltin) Clone() Object {
	panic("cannot clone builtin:clone")
}

func (c CloneBuiltin) Arity() int {
	return 1
}

func (c CloneBuiltin) Call(_ *Interpreter, args ...Object) (Object, error) {
	return args[0].Clone(), nil
}

type PanicBuiltin struct{}

func (p PanicBuiltin) String() string {
	return "builtin:panic"
}

func (p PanicBuiltin) Clone() Object {
	return PanicBuiltin{}
}

func (p PanicBuiltin) Arity() int {
	return 1
}

func (p PanicBuiltin) Call(_ *Interpreter, args ...Object) (Object, error) {
	panic(fmt.Errorf("%w: %s", ErrRuntimeFault, args[0]))
}

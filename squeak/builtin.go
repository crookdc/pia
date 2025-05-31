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

type LengthBuiltin struct{}

func (l LengthBuiltin) String() string {
	return "builtin:length"
}

func (l LengthBuiltin) Arity() int {
	return 1
}

func (l LengthBuiltin) Call(_ *Interpreter, args ...Object) (Object, error) {
	list, ok := args[0].(List)
	if !ok {
		return nil, fmt.Errorf(
			"%w: %T is not a list",
			ErrIllegalArgument,
			args[0],
		)
	}
	return Number{float64(len(list.slice))}, nil
}

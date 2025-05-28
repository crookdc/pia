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

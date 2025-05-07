package squeak

import (
	"fmt"
)

var NullObject = Object(nil)

// Object represents any piece of primitive data in a Squeak program. Not all data types are able to
// properly satisfy each method within the interface. Hence, the definition of Object is subject to
// change quite rapidly as it might be broken down into sub-interfaces. For now, any method not
// supported by an implementation is expected to return an ErrUnexpectedType error.
type Object interface {
	Add(Object) (Object, error)
	Subtract(Object) (Object, error)
	Multiply(Object) (Object, error)
	Divide(Object) (Object, error)
	Invert() (Object, error)
}

// Integer represents an integer stored within the Squeak environment.
type Integer struct {
	Integer int
}

func (i Integer) Add(obj Object) (Object, error) {
	switch obj := obj.(type) {
	case Integer:
		return Integer{Integer: i.Integer + obj.Integer}, nil
	case String:
		return String{String: fmt.Sprintf("%d%s", i.Integer, obj.String)}, nil
	default:
		return nil, ErrUnexpectedType
	}
}

func (i Integer) Subtract(o Object) (Object, error) {
	switch t := o.(type) {
	case Integer:
		return Integer{Integer: i.Integer - t.Integer}, nil
	default:
		return nil, ErrUnexpectedType
	}
}

func (i Integer) Multiply(o Object) (Object, error) {
	switch t := o.(type) {
	case Integer:
		return Integer{
			Integer: i.Integer * t.Integer,
		}, nil
	default:
		return nil, ErrUnexpectedType
	}
}

func (i Integer) Divide(o Object) (Object, error) {
	switch t := o.(type) {
	case Integer:
		return Integer{
			Integer: i.Integer / t.Integer,
		}, nil
	default:
		return nil, ErrUnexpectedType
	}
}

func (i Integer) Invert() (Object, error) {
	return nil, ErrUnexpectedType
}

// String represents a string stored within the Squeak environment.
type String struct {
	String string
}

func (s String) Add(obj Object) (Object, error) {
	switch obj := obj.(type) {
	case String:
		return String{
			String: s.String + obj.String,
		}, nil
	case Integer:
		return String{String: fmt.Sprintf("%s%d", s.String, obj.Integer)}, nil
	default:
		return nil, ErrUnexpectedType
	}
}

func (s String) Subtract(_ Object) (Object, error) {
	return nil, ErrUnexpectedType
}

func (s String) Multiply(_ Object) (Object, error) {
	return nil, ErrUnexpectedType
}

func (s String) Divide(_ Object) (Object, error) {
	return nil, ErrUnexpectedType
}

func (s String) Invert() (Object, error) {
	return nil, ErrUnexpectedType
}

// Boolean represents a boolean stored within the Squeak environment.
type Boolean struct {
	Boolean bool
}

func (b Boolean) Add(_ Object) (Object, error) {
	return nil, ErrUnexpectedType
}

func (b Boolean) Subtract(_ Object) (Object, error) {
	return nil, ErrUnexpectedType
}

func (b Boolean) Multiply(_ Object) (Object, error) {
	return nil, ErrUnexpectedType
}

func (b Boolean) Divide(_ Object) (Object, error) {
	return nil, ErrUnexpectedType
}

func (b Boolean) Invert() (Object, error) {
	return Boolean{
		Boolean: !b.Boolean,
	}, nil
}

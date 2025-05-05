package squeak

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestInteger_Add(t *testing.T) {
	type expectation struct {
		object Object
		err    error
	}
	tests := []struct {
		name     string
		integer  Integer
		other    Object
		expected expectation
	}{
		{
			name:    "given integer",
			integer: Integer{128_991},
			other:   Integer{12_123},
			expected: expectation{
				object: Integer{141_114},
				err:    nil,
			},
		},
		{
			name:    "given string",
			integer: Integer{128_991},
			other:   String{"hello world"},
			expected: expectation{
				object: nil,
				err:    ErrUnexpectedType,
			},
		},
		{
			name:    "given boolean",
			integer: Integer{128_991},
			other:   Boolean{true},
			expected: expectation{
				object: nil,
				err:    ErrUnexpectedType,
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			object, err := test.integer.Add(test.other)
			assert.Equal(t, test.expected.err, err)
			assert.Equal(t, test.expected.object, object)
		})
	}
}

func TestInteger_Subtract(t *testing.T) {
	type expectation struct {
		object Object
		err    error
	}
	tests := []struct {
		name     string
		integer  Integer
		other    Object
		expected expectation
	}{
		{
			name:    "given integer",
			integer: Integer{50},
			other:   Integer{2},
			expected: expectation{
				object: Integer{48},
				err:    nil,
			},
		},
		{
			name:    "given integer",
			integer: Integer{50},
			other:   Integer{200},
			expected: expectation{
				object: Integer{-150},
				err:    nil,
			},
		},
		{
			name:    "given string",
			integer: Integer{128_991},
			other:   String{"hello world"},
			expected: expectation{
				object: nil,
				err:    ErrUnexpectedType,
			},
		},
		{
			name:    "given boolean",
			integer: Integer{128_991},
			other:   Boolean{true},
			expected: expectation{
				object: nil,
				err:    ErrUnexpectedType,
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			object, err := test.integer.Subtract(test.other)
			assert.Equal(t, test.expected.err, err)
			assert.Equal(t, test.expected.object, object)
		})
	}
}

func TestInteger_Multiply(t *testing.T) {
	type expectation struct {
		object Object
		err    error
	}
	tests := []struct {
		name     string
		integer  Integer
		other    Object
		expected expectation
	}{
		{
			name:    "given integer",
			integer: Integer{-50},
			other:   Integer{2},
			expected: expectation{
				object: Integer{-100},
				err:    nil,
			},
		},
		{
			name:    "given integer",
			integer: Integer{50},
			other:   Integer{2},
			expected: expectation{
				object: Integer{100},
				err:    nil,
			},
		},
		{
			name:    "given string",
			integer: Integer{128_991},
			other:   String{"hello world"},
			expected: expectation{
				object: nil,
				err:    ErrUnexpectedType,
			},
		},
		{
			name:    "given boolean",
			integer: Integer{128_991},
			other:   Boolean{true},
			expected: expectation{
				object: nil,
				err:    ErrUnexpectedType,
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			object, err := test.integer.Multiply(test.other)
			assert.Equal(t, test.expected.err, err)
			assert.Equal(t, test.expected.object, object)
		})
	}
}

func TestInteger_Divide(t *testing.T) {
	type expectation struct {
		object Object
		err    error
	}
	tests := []struct {
		name     string
		integer  Integer
		other    Object
		expected expectation
	}{
		{
			name:    "given integer",
			integer: Integer{50},
			other:   Integer{2},
			expected: expectation{
				object: Integer{25},
				err:    nil,
			},
		},
		{
			name:    "given string",
			integer: Integer{128_991},
			other:   String{"hello world"},
			expected: expectation{
				object: nil,
				err:    ErrUnexpectedType,
			},
		},
		{
			name:    "given boolean",
			integer: Integer{128_991},
			other:   Boolean{true},
			expected: expectation{
				object: nil,
				err:    ErrUnexpectedType,
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			object, err := test.integer.Divide(test.other)
			assert.Equal(t, test.expected.err, err)
			assert.Equal(t, test.expected.object, object)
		})
	}
}

func TestInteger_Invert(t *testing.T) {
	type expectation struct {
		object Object
		err    error
	}
	tests := []struct {
		name     string
		integer  Integer
		other    Object
		expected expectation
	}{
		{
			name:    "given integer",
			integer: Integer{50},
			other:   Integer{2},
			expected: expectation{
				object: nil,
				err:    ErrUnexpectedType,
			},
		},
		{
			name:    "given string",
			integer: Integer{128_991},
			other:   String{"hello world"},
			expected: expectation{
				object: nil,
				err:    ErrUnexpectedType,
			},
		},
		{
			name:    "given boolean",
			integer: Integer{128_991},
			other:   Boolean{true},
			expected: expectation{
				object: nil,
				err:    ErrUnexpectedType,
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			object, err := test.integer.Invert()
			assert.Equal(t, test.expected.err, err)
			assert.Equal(t, test.expected.object, object)
		})
	}
}

func TestString_Add(t *testing.T) {
	type expectation struct {
		object Object
		err    error
	}
	tests := []struct {
		name     string
		string   String
		other    Object
		expected expectation
	}{
		{
			name:   "given integer",
			string: String{"hello"},
			other:  Integer{2},
			expected: expectation{
				object: nil,
				err:    ErrUnexpectedType,
			},
		},
		{
			name:   "given string",
			string: String{"hello"},
			other:  String{" world"},
			expected: expectation{
				object: String{"hello world"},
				err:    nil,
			},
		},
		{
			name:   "given boolean",
			string: String{"hello"},
			other:  Boolean{true},
			expected: expectation{
				object: nil,
				err:    ErrUnexpectedType,
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			object, err := test.string.Add(test.other)
			assert.Equal(t, test.expected.err, err)
			assert.Equal(t, test.expected.object, object)
		})
	}
}

func TestString_Subtract(t *testing.T) {
	type expectation struct {
		object Object
		err    error
	}
	tests := []struct {
		name     string
		string   String
		other    Object
		expected expectation
	}{
		{
			name:   "given integer",
			string: String{"hello"},
			other:  Integer{2},
			expected: expectation{
				object: nil,
				err:    ErrUnexpectedType,
			},
		},
		{
			name:   "given string",
			string: String{"hello"},
			other:  String{" world"},
			expected: expectation{
				object: nil,
				err:    ErrUnexpectedType,
			},
		},
		{
			name:   "given boolean",
			string: String{"hello"},
			other:  Boolean{true},
			expected: expectation{
				object: nil,
				err:    ErrUnexpectedType,
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			object, err := test.string.Subtract(test.other)
			assert.Equal(t, test.expected.err, err)
			assert.Equal(t, test.expected.object, object)
		})
	}
}

func TestString_Multiply(t *testing.T) {
	type expectation struct {
		object Object
		err    error
	}
	tests := []struct {
		name     string
		string   String
		other    Object
		expected expectation
	}{
		{
			name:   "given integer",
			string: String{"hello"},
			other:  Integer{2},
			expected: expectation{
				object: nil,
				err:    ErrUnexpectedType,
			},
		},
		{
			name:   "given string",
			string: String{"hello"},
			other:  String{" world"},
			expected: expectation{
				object: nil,
				err:    ErrUnexpectedType,
			},
		},
		{
			name:   "given boolean",
			string: String{"hello"},
			other:  Boolean{true},
			expected: expectation{
				object: nil,
				err:    ErrUnexpectedType,
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			object, err := test.string.Multiply(test.other)
			assert.Equal(t, test.expected.err, err)
			assert.Equal(t, test.expected.object, object)
		})
	}
}

func TestString_Divide(t *testing.T) {
	type expectation struct {
		object Object
		err    error
	}
	tests := []struct {
		name     string
		string   String
		other    Object
		expected expectation
	}{
		{
			name:   "given integer",
			string: String{"hello"},
			other:  Integer{2},
			expected: expectation{
				object: nil,
				err:    ErrUnexpectedType,
			},
		},
		{
			name:   "given string",
			string: String{"hello"},
			other:  String{" world"},
			expected: expectation{
				object: nil,
				err:    ErrUnexpectedType,
			},
		},
		{
			name:   "given boolean",
			string: String{"hello"},
			other:  Boolean{true},
			expected: expectation{
				object: nil,
				err:    ErrUnexpectedType,
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			object, err := test.string.Divide(test.other)
			assert.Equal(t, test.expected.err, err)
			assert.Equal(t, test.expected.object, object)
		})
	}
}

func TestString_Invert(t *testing.T) {
	object, err := String{""}.Invert()
	assert.Equal(t, ErrUnexpectedType, err)
	assert.Nil(t, object)
}

func TestBoolean_Add(t *testing.T) {
	type expectation struct {
		object Object
		err    error
	}
	tests := []struct {
		name     string
		boolean  Boolean
		other    Object
		expected expectation
	}{
		{
			name:    "given integer",
			boolean: Boolean{true},
			other:   Integer{2},
			expected: expectation{
				object: nil,
				err:    ErrUnexpectedType,
			},
		},
		{
			name:    "given string",
			boolean: Boolean{true},
			other:   String{" world"},
			expected: expectation{
				object: nil,
				err:    ErrUnexpectedType,
			},
		},
		{
			name:    "given boolean",
			boolean: Boolean{true},
			other:   Boolean{true},
			expected: expectation{
				object: nil,
				err:    ErrUnexpectedType,
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			object, err := test.boolean.Add(test.other)
			assert.Equal(t, test.expected.err, err)
			assert.Equal(t, test.expected.object, object)
		})
	}
}

func TestBoolean_Subtract(t *testing.T) {
	type expectation struct {
		object Object
		err    error
	}
	tests := []struct {
		name     string
		boolean  Boolean
		other    Object
		expected expectation
	}{
		{
			name:    "given integer",
			boolean: Boolean{true},
			other:   Integer{2},
			expected: expectation{
				object: nil,
				err:    ErrUnexpectedType,
			},
		},
		{
			name:    "given string",
			boolean: Boolean{true},
			other:   String{" world"},
			expected: expectation{
				object: nil,
				err:    ErrUnexpectedType,
			},
		},
		{
			name:    "given boolean",
			boolean: Boolean{true},
			other:   Boolean{true},
			expected: expectation{
				object: nil,
				err:    ErrUnexpectedType,
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			object, err := test.boolean.Subtract(test.other)
			assert.Equal(t, test.expected.err, err)
			assert.Equal(t, test.expected.object, object)
		})
	}
}

func TestBoolean_Multiply(t *testing.T) {
	type expectation struct {
		object Object
		err    error
	}
	tests := []struct {
		name     string
		boolean  Boolean
		other    Object
		expected expectation
	}{
		{
			name:    "given integer",
			boolean: Boolean{true},
			other:   Integer{2},
			expected: expectation{
				object: nil,
				err:    ErrUnexpectedType,
			},
		},
		{
			name:    "given string",
			boolean: Boolean{true},
			other:   String{" world"},
			expected: expectation{
				object: nil,
				err:    ErrUnexpectedType,
			},
		},
		{
			name:    "given boolean",
			boolean: Boolean{true},
			other:   Boolean{true},
			expected: expectation{
				object: nil,
				err:    ErrUnexpectedType,
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			object, err := test.boolean.Multiply(test.other)
			assert.Equal(t, test.expected.err, err)
			assert.Equal(t, test.expected.object, object)
		})
	}
}

func TestBoolean_Divide(t *testing.T) {
	type expectation struct {
		object Object
		err    error
	}
	tests := []struct {
		name     string
		boolean  Boolean
		other    Object
		expected expectation
	}{
		{
			name:    "given integer",
			boolean: Boolean{true},
			other:   Integer{2},
			expected: expectation{
				object: nil,
				err:    ErrUnexpectedType,
			},
		},
		{
			name:    "given string",
			boolean: Boolean{true},
			other:   String{" world"},
			expected: expectation{
				object: nil,
				err:    ErrUnexpectedType,
			},
		},
		{
			name:    "given boolean",
			boolean: Boolean{true},
			other:   Boolean{true},
			expected: expectation{
				object: nil,
				err:    ErrUnexpectedType,
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			object, err := test.boolean.Divide(test.other)
			assert.Equal(t, test.expected.err, err)
			assert.Equal(t, test.expected.object, object)
		})
	}
}

func TestBoolean_Invert(t *testing.T) {
	object, err := Boolean{false}.Invert()
	assert.Nil(t, err)
	assert.Equal(t, Boolean{true}, object)

	object, err = Boolean{true}.Invert()
	assert.Nil(t, err)
	assert.Equal(t, Boolean{false}, object)
}

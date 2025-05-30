package squeak

import (
	"bytes"
	"github.com/crookdc/pia/squeak/ast"
	"github.com/crookdc/pia/squeak/token"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestInterpreter_evaluate(t *testing.T) {
	tests := []struct {
		name string
		node ast.ExpressionNode
		obj  Object
		err  error
	}{
		{
			name: "negate integer literal",
			node: ast.Prefix{
				Operator: token.Token{
					Type:   token.Minus,
					Lexeme: "-",
				},
				Target: ast.IntegerLiteral{
					Integer: 1902,
				},
			},
			obj: Number{-1902},
		},
		{
			name: "integer literal",
			node: ast.IntegerLiteral{Integer: 12956},
			obj:  Number{12956},
		},
		{
			name: "string literal",
			node: ast.StringLiteral{String: "*+crookdc!?"},
			obj:  String{"*+crookdc!?"},
		},
		{
			name: "true boolean literal",
			node: ast.BooleanLiteral{Boolean: true},
			obj:  Boolean{true},
		},
		{
			name: "false boolean literal",
			node: ast.BooleanLiteral{Boolean: false},
			obj:  Boolean{false},
		},
		{
			name: "double inversion",
			node: ast.Prefix{
				Operator: token.Token{
					Type:   token.Bang,
					Lexeme: "!",
				},
				Target: ast.Prefix{
					Operator: token.Token{
						Type:   token.Bang,
						Lexeme: "!",
					},
					Target: ast.BooleanLiteral{Boolean: true},
				},
			},
			obj: Boolean{true},
		},
		{
			name: "single inversion of truthy string",
			node: ast.Prefix{
				Operator: token.Token{
					Type:   token.Bang,
					Lexeme: "!",
				},
				Target: ast.StringLiteral{String: "crookdc!!"},
			},
			obj: Boolean{false},
		},
		{
			name: "single inversion of truthy number",
			node: ast.Prefix{
				Operator: token.Token{
					Type:   token.Bang,
					Lexeme: "!",
				},
				Target: ast.IntegerLiteral{Integer: 123},
			},
			obj: Boolean{false},
		},
		{
			name: "single inversion of truthy boolean",
			node: ast.Prefix{
				Operator: token.Token{
					Type:   token.Bang,
					Lexeme: "!",
				},
				Target: ast.BooleanLiteral{Boolean: true},
			},
			obj: Boolean{false},
		},
		{
			name: "negation of boolean",
			node: ast.Prefix{
				Operator: token.Token{
					Type:   token.Minus,
					Lexeme: "-",
				},
				Target: ast.BooleanLiteral{Boolean: true},
			},
			err: ErrUnrecognizedOperandType,
		},
		{
			name: "negation of string",
			node: ast.Prefix{
				Operator: token.Token{
					Type:   token.Minus,
					Lexeme: "-",
				},
				Target: ast.StringLiteral{String: "hello world¡@234"},
			},
			err: ErrUnrecognizedOperandType,
		},
		{
			name: "negation of a number",
			node: ast.Prefix{
				Operator: token.Token{
					Type:   token.Minus,
					Lexeme: "-",
				},
				Target: ast.IntegerLiteral{Integer: 1234567889},
			},
			obj: Number{-1234567889},
		},
		{
			name: "double negation of a number",
			node: ast.Prefix{
				Operator: token.Token{
					Type:   token.Minus,
					Lexeme: "-",
				},
				Target: ast.Prefix{
					Operator: token.Token{
						Type:   token.Minus,
						Lexeme: "-",
					},
					Target: ast.IntegerLiteral{
						Integer: 123456778745,
					},
				},
			},
			obj: Number{123456778745},
		},
		{
			name: "infix numerical addition with prefixed operands",
			node: ast.Infix{
				Operator: token.Token{
					Type:   token.Plus,
					Lexeme: "+",
				},
				LHS: ast.IntegerLiteral{
					Integer: 10002,
				},
				RHS: ast.Infix{
					Expression: ast.Expression{},
					Operator: token.Token{
						Type:   token.Plus,
						Lexeme: "+",
					},
					LHS: ast.IntegerLiteral{
						Integer: 122,
					},
					RHS: ast.Prefix{
						Operator: token.Token{
							Type:   token.Minus,
							Lexeme: "-",
						},
						Target: ast.IntegerLiteral{
							Integer: 12,
						},
					},
				},
			},
			obj: Number{10112},
		},
		{
			name: "infix numerical addition with prefixed operands",
			node: ast.Infix{
				Operator: token.Token{
					Type:   token.Plus,
					Lexeme: "+",
				},
				LHS: ast.IntegerLiteral{
					Integer: 10002,
				},
				RHS: ast.FloatLiteral{
					Float: 13.37,
				},
			},
			obj: Number{10015.37},
		},
		{
			name: "string concatenation",
			node: ast.Infix{
				Operator: token.Token{
					Type:   token.Plus,
					Lexeme: "+",
				},
				LHS: ast.StringLiteral{
					String: "Hello",
				},
				RHS: ast.Infix{
					Expression: ast.Expression{},
					Operator: token.Token{
						Type:   token.Plus,
						Lexeme: "+",
					},
					LHS: ast.StringLiteral{
						String: " ",
					},
					RHS: ast.StringLiteral{
						String: "world",
					},
				},
			},
			obj: String{"Hello world"},
		},
		{
			name: "mixed addition operand types",
			node: ast.Infix{
				Operator: token.Token{
					Type:   token.Plus,
					Lexeme: "+",
				},
				LHS: ast.StringLiteral{
					String: "Hello",
				},
				RHS: ast.IntegerLiteral{
					Integer: 1234,
				},
			},
			err: ErrUnrecognizedOperandType,
		},
		{
			name: "mixed addition operand types",
			node: ast.Infix{
				Operator: token.Token{
					Type:   token.Plus,
					Lexeme: "+",
				},
				LHS: ast.IntegerLiteral{
					Integer: 12341,
				},
				RHS: ast.StringLiteral{
					String: "crookdc",
				},
			},
			err: ErrUnrecognizedOperandType,
		},
		{
			name: "invalid operand addition type",
			node: ast.Infix{
				Operator: token.Token{
					Type:   token.Plus,
					Lexeme: "+",
				},
				LHS: ast.BooleanLiteral{
					Boolean: true,
				},
				RHS: ast.BooleanLiteral{
					Boolean: false,
				},
			},
			err: ErrUnrecognizedOperandType,
		},
		{
			name: "invalid and mixed addition operand types",
			node: ast.Infix{
				Operator: token.Token{
					Type:   token.Plus,
					Lexeme: "+",
				},
				LHS: ast.BooleanLiteral{
					Boolean: true,
				},
				RHS: ast.IntegerLiteral{
					Integer: 12345,
				},
			},
			err: ErrUnrecognizedOperandType,
		},
		{
			name: "subtraction of numbers",
			node: ast.Infix{
				Operator: token.Token{
					Type:   token.Minus,
					Lexeme: "-",
				},
				LHS: ast.IntegerLiteral{
					Integer: 54321,
				},
				RHS: ast.IntegerLiteral{
					Integer: 12345,
				},
			},
			obj: Number{41976},
		},
		{
			name: "subtraction with negative number",
			node: ast.Infix{
				Operator: token.Token{
					Type:   token.Minus,
					Lexeme: "-",
				},
				LHS: ast.IntegerLiteral{
					Integer: 54321,
				},
				RHS: ast.Prefix{
					Operator: token.Token{
						Type:   token.Minus,
						Lexeme: "-",
					},
					Target: ast.IntegerLiteral{
						Integer: 15,
					},
				},
			},
			obj: Number{54336},
		},
		{
			name: "subtraction with strings",
			node: ast.Infix{
				Operator: token.Token{
					Type:   token.Minus,
					Lexeme: "-",
				},
				LHS: ast.StringLiteral{
					String: "crookdc",
				},
				RHS: ast.StringLiteral{
					String: "kdc",
				},
			},
			err: ErrUnrecognizedOperandType,
		},
		{
			name: "subtraction with booleans",
			node: ast.Infix{
				Operator: token.Token{
					Type:   token.Minus,
					Lexeme: "-",
				},
				LHS: ast.BooleanLiteral{
					Boolean: true,
				},
				RHS: ast.BooleanLiteral{
					Boolean: true,
				},
			},
			err: ErrUnrecognizedOperandType,
		},
		{
			name: "mixed subtraction with valid and invalid operand types",
			node: ast.Infix{
				Operator: token.Token{
					Type:   token.Minus,
					Lexeme: "-",
				},
				LHS: ast.IntegerLiteral{
					Integer: 123,
				},
				RHS: ast.BooleanLiteral{
					Boolean: true,
				},
			},
			err: ErrUnrecognizedOperandType,
		},
		{
			name: "multiplication of numbers",
			node: ast.Infix{
				Operator: token.Token{
					Type:   token.Asterisk,
					Lexeme: "*",
				},
				LHS: ast.IntegerLiteral{
					Integer: 123,
				},
				RHS: ast.IntegerLiteral{
					Integer: 12345,
				},
			},
			obj: Number{1518435},
		},
		{
			name: "multiplication with negative number",
			node: ast.Infix{
				Operator: token.Token{
					Type:   token.Asterisk,
					Lexeme: "*",
				},
				LHS: ast.IntegerLiteral{
					Integer: 54321,
				},
				RHS: ast.Prefix{
					Operator: token.Token{
						Type:   token.Minus,
						Lexeme: "-",
					},
					Target: ast.IntegerLiteral{
						Integer: 15,
					},
				},
			},
			obj: Number{-814815},
		},
		{
			name: "multiplication with strings",
			node: ast.Infix{
				Operator: token.Token{
					Type:   token.Asterisk,
					Lexeme: "*",
				},
				LHS: ast.StringLiteral{
					String: "crookdc",
				},
				RHS: ast.StringLiteral{
					String: "kdc",
				},
			},
			err: ErrUnrecognizedOperandType,
		},
		{
			name: "multiplication with booleans",
			node: ast.Infix{
				Operator: token.Token{
					Type:   token.Asterisk,
					Lexeme: "*",
				},
				LHS: ast.BooleanLiteral{
					Boolean: true,
				},
				RHS: ast.BooleanLiteral{
					Boolean: true,
				},
			},
			err: ErrUnrecognizedOperandType,
		},
		{
			name: "mixed multiplication with valid and invalid operand types",
			node: ast.Infix{
				Operator: token.Token{
					Type:   token.Asterisk,
					Lexeme: "*",
				},
				LHS: ast.IntegerLiteral{
					Integer: 123,
				},
				RHS: ast.BooleanLiteral{
					Boolean: true,
				},
			},
			err: ErrUnrecognizedOperandType,
		},
		{
			name: "division of numbers",
			node: ast.Infix{
				Operator: token.Token{
					Type:   token.Slash,
					Lexeme: "/",
				},
				LHS: ast.IntegerLiteral{
					Integer: 100,
				},
				RHS: ast.IntegerLiteral{
					Integer: 5,
				},
			},
			obj: Number{20},
		},
		{
			name: "division to fraction",
			node: ast.Infix{
				Operator: token.Token{
					Type:   token.Slash,
					Lexeme: "/",
				},
				LHS: ast.IntegerLiteral{
					Integer: 5,
				},
				RHS: ast.IntegerLiteral{
					Integer: 2,
				},
			},
			obj: Number{2.5},
		},
		{
			name: "division by zero",
			node: ast.Infix{
				Operator: token.Token{
					Type:   token.Slash,
					Lexeme: "/",
				},
				LHS: ast.IntegerLiteral{
					Integer: 5,
				},
				RHS: ast.IntegerLiteral{
					Integer: 0,
				},
			},
			err: ErrIllegalArgument,
		},
		{
			name: "division with negative number",
			node: ast.Infix{
				Operator: token.Token{
					Type:   token.Slash,
					Lexeme: "/",
				},
				LHS: ast.IntegerLiteral{
					Integer: 100,
				},
				RHS: ast.Prefix{
					Operator: token.Token{
						Type:   token.Minus,
						Lexeme: "-",
					},
					Target: ast.IntegerLiteral{
						Integer: 10,
					},
				},
			},
			obj: Number{-10},
		},
		{
			name: "division with strings",
			node: ast.Infix{
				Operator: token.Token{
					Type:   token.Slash,
					Lexeme: "/",
				},
				LHS: ast.StringLiteral{
					String: "crookdc",
				},
				RHS: ast.StringLiteral{
					String: "kdc",
				},
			},
			err: ErrUnrecognizedOperandType,
		},
		{
			name: "division with booleans",
			node: ast.Infix{
				Operator: token.Token{
					Type:   token.Slash,
					Lexeme: "/",
				},
				LHS: ast.BooleanLiteral{
					Boolean: true,
				},
				RHS: ast.BooleanLiteral{
					Boolean: true,
				},
			},
			err: ErrUnrecognizedOperandType,
		},
		{
			name: "mixed division with valid and invalid operand types",
			node: ast.Infix{
				Operator: token.Token{
					Type:   token.Slash,
					Lexeme: "/",
				},
				LHS: ast.IntegerLiteral{
					Integer: 123,
				},
				RHS: ast.BooleanLiteral{
					Boolean: true,
				},
			},
			err: ErrUnrecognizedOperandType,
		},
		{
			name: "comparing less than with numbers",
			node: ast.Infix{
				Operator: token.Token{
					Type:   token.Less,
					Lexeme: "<",
				},
				LHS: ast.IntegerLiteral{
					Integer: 123,
				},
				RHS: ast.IntegerLiteral{
					Integer: 13,
				},
			},
			obj: Boolean{false},
		},
		{
			name: "comparing less than with numbers",
			node: ast.Infix{
				Operator: token.Token{
					Type:   token.Less,
					Lexeme: "<",
				},
				LHS: ast.IntegerLiteral{
					Integer: 34,
				},
				RHS: ast.IntegerLiteral{
					Integer: 132,
				},
			},
			obj: Boolean{true},
		},
		{
			name: "comparing less than with numbers",
			node: ast.Infix{
				Operator: token.Token{
					Type:   token.Less,
					Lexeme: "<",
				},
				LHS: ast.IntegerLiteral{
					Integer: -34,
				},
				RHS: ast.IntegerLiteral{
					Integer: 132,
				},
			},
			obj: Boolean{true},
		},
		{
			name: "comparing less than with booleans",
			node: ast.Infix{
				Operator: token.Token{
					Type:   token.Less,
					Lexeme: "<",
				},
				LHS: ast.BooleanLiteral{
					Boolean: true,
				},
				RHS: ast.BooleanLiteral{
					Boolean: false,
				},
			},
			err: ErrUnrecognizedOperandType,
		},
		{
			name: "comparing less than with strings",
			node: ast.Infix{
				Operator: token.Token{
					Type:   token.Less,
					Lexeme: "<",
				},
				LHS: ast.StringLiteral{
					String: "crookdc",
				},
				RHS: ast.StringLiteral{
					String: "lackluster",
				},
			},
			err: ErrUnrecognizedOperandType,
		},
		{
			name: "comparing less than with string and boolean",
			node: ast.Infix{
				Operator: token.Token{
					Type:   token.Less,
					Lexeme: "<",
				},
				LHS: ast.StringLiteral{
					String: "crookdc",
				},
				RHS: ast.BooleanLiteral{
					Boolean: true,
				},
			},
			err: ErrUnrecognizedOperandType,
		},
		{
			name: "comparing less than with string and number",
			node: ast.Infix{
				Operator: token.Token{
					Type:   token.Less,
					Lexeme: "<",
				},
				LHS: ast.StringLiteral{
					String: "crookdc",
				},
				RHS: ast.IntegerLiteral{
					Integer: 12,
				},
			},
			err: ErrUnrecognizedOperandType,
		},
		{
			name: "comparing less than with boolean and number",
			node: ast.Infix{
				Operator: token.Token{
					Type:   token.Less,
					Lexeme: "<",
				},
				LHS: ast.BooleanLiteral{
					Boolean: true,
				},
				RHS: ast.IntegerLiteral{
					Integer: 12,
				},
			},
			err: ErrUnrecognizedOperandType,
		},
		{
			name: "comparing greater than with numbers",
			node: ast.Infix{
				Operator: token.Token{
					Type:   token.Greater,
					Lexeme: ">",
				},
				LHS: ast.IntegerLiteral{
					Integer: 123,
				},
				RHS: ast.IntegerLiteral{
					Integer: 13,
				},
			},
			obj: Boolean{true},
		},
		{
			name: "comparing greater than with numbers",
			node: ast.Infix{
				Operator: token.Token{
					Type:   token.Greater,
					Lexeme: ">",
				},
				LHS: ast.IntegerLiteral{
					Integer: 34,
				},
				RHS: ast.IntegerLiteral{
					Integer: 132,
				},
			},
			obj: Boolean{false},
		},
		{
			name: "comparing greater than with numbers",
			node: ast.Infix{
				Operator: token.Token{
					Type:   token.Greater,
					Lexeme: ">",
				},
				LHS: ast.IntegerLiteral{
					Integer: -34,
				},
				RHS: ast.IntegerLiteral{
					Integer: 132,
				},
			},
			obj: Boolean{false},
		},
		{
			name: "comparing greater than with booleans",
			node: ast.Infix{
				Operator: token.Token{
					Type:   token.Greater,
					Lexeme: ">",
				},
				LHS: ast.BooleanLiteral{
					Boolean: true,
				},
				RHS: ast.BooleanLiteral{
					Boolean: false,
				},
			},
			err: ErrUnrecognizedOperandType,
		},
		{
			name: "comparing greater than with strings",
			node: ast.Infix{
				Operator: token.Token{
					Type:   token.Greater,
					Lexeme: ">",
				},
				LHS: ast.StringLiteral{
					String: "crookdc",
				},
				RHS: ast.StringLiteral{
					String: "lackluster",
				},
			},
			err: ErrUnrecognizedOperandType,
		},
		{
			name: "comparing greater than with string and boolean",
			node: ast.Infix{
				Operator: token.Token{
					Type:   token.Greater,
					Lexeme: ">",
				},
				LHS: ast.StringLiteral{
					String: "crookdc",
				},
				RHS: ast.BooleanLiteral{
					Boolean: true,
				},
			},
			err: ErrUnrecognizedOperandType,
		},
		{
			name: "comparing greater than with string and number",
			node: ast.Infix{
				Operator: token.Token{
					Type:   token.Greater,
					Lexeme: ">",
				},
				LHS: ast.StringLiteral{
					String: "crookdc",
				},
				RHS: ast.IntegerLiteral{
					Integer: 12,
				},
			},
			err: ErrUnrecognizedOperandType,
		},
		{
			name: "comparing greater than with boolean and number",
			node: ast.Infix{
				Operator: token.Token{
					Type:   token.Greater,
					Lexeme: ">",
				},
				LHS: ast.BooleanLiteral{
					Boolean: true,
				},
				RHS: ast.IntegerLiteral{
					Integer: 12,
				},
			},
			err: ErrUnrecognizedOperandType,
		},
		{
			name: "comparing equals with numbers",
			node: ast.Infix{
				Operator: token.Token{
					Type:   token.Equals,
					Lexeme: "==",
				},
				LHS: ast.IntegerLiteral{
					Integer: 34,
				},
				RHS: ast.IntegerLiteral{
					Integer: 132,
				},
			},
			obj: Boolean{false},
		},
		{
			name: "comparing equals with numbers",
			node: ast.Infix{
				Operator: token.Token{
					Type:   token.Equals,
					Lexeme: "==",
				},
				LHS: ast.IntegerLiteral{
					Integer: -34,
				},
				RHS: ast.IntegerLiteral{
					Integer: -34,
				},
			},
			obj: Boolean{true},
		},
		{
			name: "comparing equals with booleans",
			node: ast.Infix{
				Operator: token.Token{
					Type:   token.Equals,
					Lexeme: "==",
				},
				LHS: ast.BooleanLiteral{
					Boolean: true,
				},
				RHS: ast.BooleanLiteral{
					Boolean: false,
				},
			},
			obj: Boolean{false},
		},
		{
			name: "comparing equals with booleans",
			node: ast.Infix{
				Operator: token.Token{
					Type:   token.Equals,
					Lexeme: "==",
				},
				LHS: ast.BooleanLiteral{
					Boolean: true,
				},
				RHS: ast.BooleanLiteral{
					Boolean: true,
				},
			},
			obj: Boolean{true},
		},
		{
			name: "comparing equals with strings",
			node: ast.Infix{
				Operator: token.Token{
					Type:   token.Equals,
					Lexeme: "==",
				},
				LHS: ast.StringLiteral{
					String: "crookdc",
				},
				RHS: ast.StringLiteral{
					String: "lackluster",
				},
			},
			obj: Boolean{false},
		},
		{
			name: "comparing equals with strings",
			node: ast.Infix{
				Operator: token.Token{
					Type:   token.Equals,
					Lexeme: "==",
				},
				LHS: ast.StringLiteral{
					String: "crookdc",
				},
				RHS: ast.StringLiteral{
					String: "crookdc",
				},
			},
			obj: Boolean{true},
		},
		{
			name: "comparing equals with string and boolean",
			node: ast.Infix{
				Operator: token.Token{
					Type:   token.Equals,
					Lexeme: "==",
				},
				LHS: ast.StringLiteral{
					String: "crookdc",
				},
				RHS: ast.BooleanLiteral{
					Boolean: true,
				},
			},
			err: ErrUnrecognizedOperandType,
		},
		{
			name: "comparing equals with string and number",
			node: ast.Infix{
				Operator: token.Token{
					Type:   token.Equals,
					Lexeme: "==",
				},
				LHS: ast.StringLiteral{
					String: "crookdc",
				},
				RHS: ast.IntegerLiteral{
					Integer: 12,
				},
			},
			err: ErrUnrecognizedOperandType,
		},
		{
			name: "comparing equals with boolean and number",
			node: ast.Infix{
				Operator: token.Token{
					Type:   token.Equals,
					Lexeme: "==",
				},
				LHS: ast.BooleanLiteral{
					Boolean: true,
				},
				RHS: ast.IntegerLiteral{
					Integer: 12,
				},
			},
			err: ErrUnrecognizedOperandType,
		},
		{
			name: "comparing not equals with numbers",
			node: ast.Infix{
				Operator: token.Token{
					Type:   token.NotEquals,
					Lexeme: "!=",
				},
				LHS: ast.IntegerLiteral{
					Integer: 34,
				},
				RHS: ast.IntegerLiteral{
					Integer: 132,
				},
			},
			obj: Boolean{true},
		},
		{
			name: "comparing not equals with numbers",
			node: ast.Infix{
				Operator: token.Token{
					Type:   token.NotEquals,
					Lexeme: "!=",
				},
				LHS: ast.IntegerLiteral{
					Integer: -34,
				},
				RHS: ast.IntegerLiteral{
					Integer: -34,
				},
			},
			obj: Boolean{false},
		},
		{
			name: "comparing not equals with booleans",
			node: ast.Infix{
				Operator: token.Token{
					Type:   token.NotEquals,
					Lexeme: "!=",
				},
				LHS: ast.BooleanLiteral{
					Boolean: true,
				},
				RHS: ast.BooleanLiteral{
					Boolean: false,
				},
			},
			obj: Boolean{true},
		},
		{
			name: "comparing not equals with booleans",
			node: ast.Infix{
				Operator: token.Token{
					Type:   token.NotEquals,
					Lexeme: "!=",
				},
				LHS: ast.BooleanLiteral{
					Boolean: true,
				},
				RHS: ast.BooleanLiteral{
					Boolean: true,
				},
			},
			obj: Boolean{false},
		},
		{
			name: "comparing not equals with strings",
			node: ast.Infix{
				Operator: token.Token{
					Type:   token.NotEquals,
					Lexeme: "!=",
				},
				LHS: ast.StringLiteral{
					String: "crookdc",
				},
				RHS: ast.StringLiteral{
					String: "lackluster",
				},
			},
			obj: Boolean{true},
		},
		{
			name: "comparing not equals with strings",
			node: ast.Infix{
				Operator: token.Token{
					Type:   token.NotEquals,
					Lexeme: "!=",
				},
				LHS: ast.StringLiteral{
					String: "crookdc",
				},
				RHS: ast.StringLiteral{
					String: "crookdc",
				},
			},
			obj: Boolean{false},
		},
		{
			name: "comparing not equals with string and boolean",
			node: ast.Infix{
				Operator: token.Token{
					Type:   token.NotEquals,
					Lexeme: "!=",
				},
				LHS: ast.StringLiteral{
					String: "crookdc",
				},
				RHS: ast.BooleanLiteral{
					Boolean: true,
				},
			},
			err: ErrUnrecognizedOperandType,
		},
		{
			name: "comparing not equals with string and number",
			node: ast.Infix{
				Operator: token.Token{
					Type:   token.NotEquals,
					Lexeme: "!=",
				},
				LHS: ast.StringLiteral{
					String: "crookdc",
				},
				RHS: ast.IntegerLiteral{
					Integer: 12,
				},
			},
			err: ErrUnrecognizedOperandType,
		},
		{
			name: "comparing not equals with boolean and number",
			node: ast.Infix{
				Operator: token.Token{
					Type:   token.NotEquals,
					Lexeme: "!=",
				},
				LHS: ast.BooleanLiteral{
					Boolean: true,
				},
				RHS: ast.IntegerLiteral{
					Integer: 12,
				},
			},
			err: ErrUnrecognizedOperandType,
		},

		{
			name: "comparing less equals with numbers",
			node: ast.Infix{
				Operator: token.Token{
					Type:   token.LessEqual,
					Lexeme: "<=",
				},
				LHS: ast.IntegerLiteral{
					Integer: 34,
				},
				RHS: ast.IntegerLiteral{
					Integer: 132,
				},
			},
			obj: Boolean{true},
		},
		{
			name: "comparing less equals with numbers",
			node: ast.Infix{
				Operator: token.Token{
					Type:   token.LessEqual,
					Lexeme: "<=",
				},
				LHS: ast.IntegerLiteral{
					Integer: -34,
				},
				RHS: ast.IntegerLiteral{
					Integer: -34,
				},
			},
			obj: Boolean{true},
		},
		{
			name: "comparing less equals with numbers",
			node: ast.Infix{
				Operator: token.Token{
					Type:   token.LessEqual,
					Lexeme: "<=",
				},
				LHS: ast.IntegerLiteral{
					Integer: 34,
				},
				RHS: ast.IntegerLiteral{
					Integer: -34,
				},
			},
			obj: Boolean{false},
		},
		{
			name: "comparing less equals with booleans",
			node: ast.Infix{
				Operator: token.Token{
					Type:   token.LessEqual,
					Lexeme: "<=",
				},
				LHS: ast.BooleanLiteral{
					Boolean: true,
				},
				RHS: ast.BooleanLiteral{
					Boolean: false,
				},
			},
			err: ErrUnrecognizedOperandType,
		},
		{
			name: "comparing less equals with strings",
			node: ast.Infix{
				Operator: token.Token{
					Type:   token.LessEqual,
					Lexeme: "<=",
				},
				LHS: ast.StringLiteral{
					String: "crookdc",
				},
				RHS: ast.StringLiteral{
					String: "lackluster",
				},
			},
			err: ErrUnrecognizedOperandType,
		},
		{
			name: "comparing less equals with string and boolean",
			node: ast.Infix{
				Operator: token.Token{
					Type:   token.LessEqual,
					Lexeme: "<=",
				},
				LHS: ast.StringLiteral{
					String: "crookdc",
				},
				RHS: ast.BooleanLiteral{
					Boolean: true,
				},
			},
			err: ErrUnrecognizedOperandType,
		},
		{
			name: "comparing less equals with string and number",
			node: ast.Infix{
				Operator: token.Token{
					Type:   token.LessEqual,
					Lexeme: "<=",
				},
				LHS: ast.StringLiteral{
					String: "crookdc",
				},
				RHS: ast.IntegerLiteral{
					Integer: 12,
				},
			},
			err: ErrUnrecognizedOperandType,
		},
		{
			name: "comparing less equals with boolean and number",
			node: ast.Infix{
				Operator: token.Token{
					Type:   token.LessEqual,
					Lexeme: "<=",
				},
				LHS: ast.BooleanLiteral{
					Boolean: true,
				},
				RHS: ast.IntegerLiteral{
					Integer: 12,
				},
			},
			err: ErrUnrecognizedOperandType,
		},

		{
			name: "comparing greater equals with numbers",
			node: ast.Infix{
				Operator: token.Token{
					Type:   token.GreaterEqual,
					Lexeme: ">=",
				},
				LHS: ast.IntegerLiteral{
					Integer: 34,
				},
				RHS: ast.IntegerLiteral{
					Integer: 132,
				},
			},
			obj: Boolean{false},
		},
		{
			name: "comparing greater equals with numbers",
			node: ast.Infix{
				Operator: token.Token{
					Type:   token.GreaterEqual,
					Lexeme: ">=",
				},
				LHS: ast.IntegerLiteral{
					Integer: -34,
				},
				RHS: ast.IntegerLiteral{
					Integer: -34,
				},
			},
			obj: Boolean{true},
		},
		{
			name: "comparing greater equals with numbers",
			node: ast.Infix{
				Operator: token.Token{
					Type:   token.GreaterEqual,
					Lexeme: ">=",
				},
				LHS: ast.IntegerLiteral{
					Integer: 34,
				},
				RHS: ast.IntegerLiteral{
					Integer: -34,
				},
			},
			obj: Boolean{true},
		},
		{
			name: "comparing greater equals with booleans",
			node: ast.Infix{
				Operator: token.Token{
					Type:   token.GreaterEqual,
					Lexeme: ">=",
				},
				LHS: ast.BooleanLiteral{
					Boolean: true,
				},
				RHS: ast.BooleanLiteral{
					Boolean: false,
				},
			},
			err: ErrUnrecognizedOperandType,
		},
		{
			name: "comparing greater equals with strings",
			node: ast.Infix{
				Operator: token.Token{
					Type:   token.GreaterEqual,
					Lexeme: ">=",
				},
				LHS: ast.StringLiteral{
					String: "crookdc",
				},
				RHS: ast.StringLiteral{
					String: "lackluster",
				},
			},
			err: ErrUnrecognizedOperandType,
		},
		{
			name: "comparing greater equals with string and boolean",
			node: ast.Infix{
				Operator: token.Token{
					Type:   token.GreaterEqual,
					Lexeme: ">=",
				},
				LHS: ast.StringLiteral{
					String: "crookdc",
				},
				RHS: ast.BooleanLiteral{
					Boolean: true,
				},
			},
			err: ErrUnrecognizedOperandType,
		},
		{
			name: "comparing greater equals with string and number",
			node: ast.Infix{
				Operator: token.Token{
					Type:   token.GreaterEqual,
					Lexeme: ">=",
				},
				LHS: ast.StringLiteral{
					String: "crookdc",
				},
				RHS: ast.IntegerLiteral{
					Integer: 12,
				},
			},
			err: ErrUnrecognizedOperandType,
		},
		{
			name: "comparing greater equals with boolean and number",
			node: ast.Infix{
				Operator: token.Token{
					Type:   token.GreaterEqual,
					Lexeme: ">=",
				},
				LHS: ast.BooleanLiteral{
					Boolean: true,
				},
				RHS: ast.IntegerLiteral{
					Integer: 12,
				},
			},
			err: ErrUnrecognizedOperandType,
		},
		{
			name: "logical and",
			node: ast.Logical{
				Operator: token.Token{
					Type:   token.And,
					Lexeme: "and",
				},
				LHS: ast.IntegerLiteral{
					Integer: 1,
				},
				RHS: ast.BooleanLiteral{
					Boolean: true,
				},
			},
			obj: Boolean{true},
		},
		{
			name: "logical and",
			node: ast.Logical{
				Operator: token.Token{
					Type:   token.And,
					Lexeme: "and",
				},
				LHS: ast.IntegerLiteral{
					Integer: 1,
				},
				RHS: ast.NilLiteral{},
			},
			obj: nil,
		},
		{
			name: "logical and",
			node: ast.Logical{
				Operator: token.Token{
					Type:   token.And,
					Lexeme: "and",
				},
				LHS: ast.NilLiteral{},
				RHS: ast.BooleanLiteral{
					Boolean: true,
				},
			},
			obj: nil,
		},
		{
			name: "logical and",
			node: ast.Logical{
				Operator: token.Token{
					Type:   token.And,
					Lexeme: "and",
				},
				LHS: ast.IntegerLiteral{
					Integer: 1,
				},
				RHS: ast.BooleanLiteral{
					Boolean: false,
				},
			},
			obj: Boolean{false},
		},
		{
			name: "logical and",
			node: ast.Logical{
				Operator: token.Token{
					Type:   token.And,
					Lexeme: "and",
				},
				LHS: ast.NilLiteral{},
				RHS: ast.BooleanLiteral{
					Boolean: false,
				},
			},
			obj: nil,
		},
		{
			name: "logical or",
			node: ast.Logical{
				Operator: token.Token{
					Type:   token.Or,
					Lexeme: "or",
				},
				LHS: ast.IntegerLiteral{
					Integer: 1,
				},
				RHS: ast.BooleanLiteral{
					Boolean: true,
				},
			},
			obj: Number{1},
		},
		{
			name: "logical or",
			node: ast.Logical{
				Operator: token.Token{
					Type:   token.Or,
					Lexeme: "or",
				},
				LHS: ast.IntegerLiteral{
					Integer: 1,
				},
				RHS: ast.NilLiteral{},
			},
			obj: Number{1},
		},
		{
			name: "logical or",
			node: ast.Logical{
				Operator: token.Token{
					Type:   token.Or,
					Lexeme: "or",
				},
				LHS: ast.NilLiteral{},
				RHS: ast.BooleanLiteral{
					Boolean: true,
				},
			},
			obj: Boolean{true},
		},
		{
			name: "logical or",
			node: ast.Logical{
				Operator: token.Token{
					Type:   token.Or,
					Lexeme: "or",
				},
				LHS: ast.IntegerLiteral{
					Integer: 1,
				},
				RHS: ast.BooleanLiteral{
					Boolean: false,
				},
			},
			obj: Number{1},
		},
		{
			name: "logical or",
			node: ast.Logical{
				Operator: token.Token{
					Type:   token.Or,
					Lexeme: "or",
				},
				LHS: ast.NilLiteral{},
				RHS: ast.BooleanLiteral{
					Boolean: false,
				},
			},
			obj: Boolean{false},
		},
		{
			name: "nested logical operators",
			node: ast.Logical{
				Operator: token.Token{
					Type:   token.Or,
					Lexeme: "or",
				},
				LHS: ast.Logical{
					Operator: token.Token{
						Type:   token.And,
						Lexeme: "and",
					},
					LHS: ast.BooleanLiteral{
						Boolean: true,
					},
					RHS: ast.BooleanLiteral{
						Boolean: true,
					},
				},
				RHS: ast.BooleanLiteral{
					Boolean: false,
				},
			},
			obj: Boolean{true},
		},
		{
			name: "nested logical operators",
			node: ast.Logical{
				Operator: token.Token{
					Type:   token.Or,
					Lexeme: "or",
				},
				LHS: ast.Logical{
					Operator: token.Token{
						Type:   token.And,
						Lexeme: "and",
					},
					LHS: ast.BooleanLiteral{
						Boolean: true,
					},
					RHS: ast.BooleanLiteral{
						Boolean: false,
					},
				},
				RHS: ast.BooleanLiteral{
					Boolean: false,
				},
			},
			obj: Boolean{false},
		},
		{
			name: "nested logical operators",
			node: ast.Logical{
				Operator: token.Token{
					Type:   token.Or,
					Lexeme: "or",
				},
				LHS: ast.Logical{
					Operator: token.Token{
						Type:   token.And,
						Lexeme: "and",
					},
					LHS: ast.BooleanLiteral{
						Boolean: true,
					},
					RHS: ast.BooleanLiteral{
						Boolean: false,
					},
				},
				RHS: ast.Logical{
					Operator: token.Token{
						Type:   token.Or,
						Lexeme: "or",
					},
					LHS: ast.NilLiteral{},
					RHS: ast.StringLiteral{String: ""},
				},
			},
			obj: String{""},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			obj, err := (&Interpreter{}).evaluate(test.node)
			assert.ErrorIs(t, err, test.err)
			if err == nil {
				// The returned value is only interesting if the returned error is nil. If the error is not nil then the
				// returned object does not have a defined rule to its state and should never be used anyway.
				assert.Equal(t, test.obj, obj)
			}
		})
	}

	t.Run("unrecognized expression", func(t *testing.T) {
		_, err := (&Interpreter{}).evaluate(nil)
		assert.ErrorIs(t, err, ErrUnrecognizedExpression)
	})
}

func TestInterpreter_Execute(t *testing.T) {
	tests := []struct {
		name    string
		program []ast.StatementNode
		preload *Environment

		out     string
		uw      *unwinder
		env     *Environment
		exports map[string]Object
		err     error
	}{
		{
			name:    "continue outside of loop",
			preload: NewEnvironment(),
			program: []ast.StatementNode{ast.Continue{}},
			err:     ErrRuntimeFault,
			env:     NewEnvironment(),
			exports: make(map[string]Object),
		},
		{
			name:    "break outside of loop",
			preload: NewEnvironment(),
			program: []ast.StatementNode{ast.Break{}},
			err:     ErrRuntimeFault,
			env:     NewEnvironment(),
			exports: make(map[string]Object),
		},
		{
			name:    "variable declaration without initializer",
			preload: NewEnvironment(),
			program: []ast.StatementNode{ast.Declaration{
				Name: token.Token{
					Type:   token.Identifier,
					Lexeme: "name",
				},
			}},
			env:     NewEnvironment(Prefill("name", nil)),
			exports: make(map[string]Object),
		},
		{
			name:    "variable declaration with explicit nil initializer",
			preload: NewEnvironment(),
			program: []ast.StatementNode{ast.Declaration{
				Name: token.Token{
					Type:   token.Identifier,
					Lexeme: "name",
				},
				Initializer: ast.NilLiteral{},
			}},
			env:     NewEnvironment(Prefill("name", nil)),
			exports: make(map[string]Object),
		},
		{
			name:    "variable declaration with initializer",
			preload: NewEnvironment(),
			program: []ast.StatementNode{ast.Declaration{
				Name: token.Token{
					Type:   token.Identifier,
					Lexeme: "name",
				},
				Initializer: ast.Infix{
					Operator: token.Token{
						Type:   token.Plus,
						Lexeme: "+",
					},
					LHS: ast.StringLiteral{
						String: "hello",
					},
					RHS: ast.StringLiteral{
						String: "goodbye",
					},
				},
			}},
			env:     NewEnvironment(Prefill("name", String{"hellogoodbye"})),
			exports: make(map[string]Object),
		},
		{
			name:    "block that assigns in parent scope and declares new variable in local scope",
			preload: NewEnvironment(Prefill("name", Number{1.123})),
			program: []ast.StatementNode{ast.Block{
				Body: []ast.StatementNode{
					ast.ExpressionStatement{
						Expression: ast.Assignment{
							Name: token.Token{
								Type:   token.Identifier,
								Lexeme: "name",
							},
							Value: ast.FloatLiteral{
								Float: 1556.12,
							},
						},
					},
					ast.Declaration{
						Name: token.Token{
							Type:   token.Identifier,
							Lexeme: "age",
						},
						Initializer: ast.IntegerLiteral{
							Integer: 27,
						},
					},
				},
			}},
			env:     NewEnvironment(Prefill("name", Number{1556.12})),
			exports: make(map[string]Object),
		},
		{
			name:    "empty block",
			preload: NewEnvironment(Prefill("name", Number{1.123})),
			program: []ast.StatementNode{ast.Block{}},
			env:     NewEnvironment(Prefill("name", Number{1.123})),
			exports: make(map[string]Object),
		},
		{
			name:    "if-else that evaluates to true",
			preload: NewEnvironment(),
			program: []ast.StatementNode{ast.If{
				Condition: ast.BooleanLiteral{Boolean: true},
				Then: ast.Declaration{
					Name: token.Token{
						Type:   token.Identifier,
						Lexeme: "result",
					},
					Initializer: ast.BooleanLiteral{Boolean: true},
				},
				Else: ast.Declaration{
					Name: token.Token{
						Type:   token.Identifier,
						Lexeme: "result",
					},
					Initializer: ast.BooleanLiteral{Boolean: false},
				},
			}},
			env:     NewEnvironment(Prefill("result", Boolean{true})),
			exports: make(map[string]Object),
		},
		{
			name:    "if-else that evaluates to false",
			preload: NewEnvironment(),
			program: []ast.StatementNode{ast.If{
				Condition: ast.BooleanLiteral{Boolean: false},
				Then: ast.Declaration{
					Name: token.Token{
						Type:   token.Identifier,
						Lexeme: "result",
					},
					Initializer: ast.BooleanLiteral{Boolean: true},
				},
				Else: ast.Declaration{
					Name: token.Token{
						Type:   token.Identifier,
						Lexeme: "result",
					},
					Initializer: ast.BooleanLiteral{Boolean: false},
				},
			}},
			env:     NewEnvironment(Prefill("result", Boolean{false})),
			exports: make(map[string]Object),
		},
		{
			name:    "if that evaluates to true",
			preload: NewEnvironment(),
			program: []ast.StatementNode{ast.If{
				Condition: ast.BooleanLiteral{Boolean: true},
				Then: ast.Declaration{
					Name: token.Token{
						Type:   token.Identifier,
						Lexeme: "result",
					},
					Initializer: ast.BooleanLiteral{Boolean: true},
				},
			}},
			env:     NewEnvironment(Prefill("result", Boolean{true})),
			exports: make(map[string]Object),
		},
		{
			name:    "if that evaluates to false",
			preload: NewEnvironment(),
			program: []ast.StatementNode{ast.If{
				Condition: ast.BooleanLiteral{Boolean: false},
				Then: ast.Declaration{
					Name: token.Token{
						Type:   token.Identifier,
						Lexeme: "result",
					},
					Initializer: ast.BooleanLiteral{Boolean: true},
				},
			}},
			env:     NewEnvironment(),
			exports: make(map[string]Object),
		},
		{
			name:    "noop is ignored",
			preload: NewEnvironment(),
			program: []ast.StatementNode{ast.Noop{}},
			env:     NewEnvironment(),
			exports: make(map[string]Object),
		},
		{
			name:    "export variable",
			preload: NewEnvironment(Prefill("name", String{"crookdc"})),
			program: []ast.StatementNode{
				ast.Export{
					Name: token.Token{
						Type:   token.Identifier,
						Lexeme: "name",
					},
					Value: ast.Variable{
						Name: token.Token{
							Type:   token.Identifier,
							Lexeme: "name",
						},
					},
				},
			},
			env: NewEnvironment(Prefill("name", String{"crookdc"})),
			exports: map[string]Object{
				"name": String{"crookdc"},
			},
		},
		{
			name:    "calling a Number",
			preload: NewEnvironment(Prefill("not_callable", Number{1})),
			program: []ast.StatementNode{
				ast.ExpressionStatement{
					Expression: ast.Call{
						Callee: ast.Variable{
							Name: token.Token{
								Type:   token.Identifier,
								Lexeme: "not_callable",
							},
						},
						Operator: token.Token{
							Type:   token.Identifier,
							Lexeme: "(",
						},
					},
				},
			},
			err:     ErrNotCallable,
			env:     NewEnvironment(Prefill("not_callable", Number{1})),
			exports: make(map[string]Object),
		},
		{
			name:    "calling a String",
			preload: NewEnvironment(Prefill("not_callable", String{"string"})),
			program: []ast.StatementNode{
				ast.ExpressionStatement{
					Expression: ast.Call{
						Callee: ast.Variable{
							Name: token.Token{
								Type:   token.Identifier,
								Lexeme: "not_callable",
							},
						},
						Operator: token.Token{
							Type:   token.Identifier,
							Lexeme: "(",
						},
					},
				},
			},
			err:     ErrNotCallable,
			env:     NewEnvironment(Prefill("not_callable", String{"string"})),
			exports: make(map[string]Object),
		},
		{
			name:    "calling a boolean",
			preload: NewEnvironment(Prefill("not_callable", Boolean{true})),
			program: []ast.StatementNode{
				ast.ExpressionStatement{
					Expression: ast.Call{
						Callee: ast.Variable{
							Name: token.Token{
								Type:   token.Identifier,
								Lexeme: "not_callable",
							},
						},
						Operator: token.Token{
							Type:   token.Identifier,
							Lexeme: "(",
						},
					},
				},
			},
			err:     ErrNotCallable,
			env:     NewEnvironment(Prefill("not_callable", Boolean{true})),
			exports: make(map[string]Object),
		},
		{
			name:    "while loop",
			preload: NewEnvironment(Prefill("i", Number{1})),
			program: []ast.StatementNode{
				ast.While{
					Condition: ast.Infix{
						Operator: token.Token{
							Type:   token.Less,
							Lexeme: "<",
						},
						LHS: ast.Variable{
							Name: token.Token{
								Type:   token.Identifier,
								Lexeme: "i",
							},
						},
						RHS: ast.IntegerLiteral{
							Integer: 5,
						},
					},
					Body: ast.Block{
						Body: []ast.StatementNode{
							ast.ExpressionStatement{
								Expression: ast.Assignment{
									Name: token.Token{
										Type:   token.Identifier,
										Lexeme: "i",
									},
									Value: ast.Infix{
										Operator: token.Token{
											Type:   token.Plus,
											Lexeme: "+",
										},
										LHS: ast.Variable{
											Name: token.Token{
												Type:   token.Identifier,
												Lexeme: "i",
											},
										},
										RHS: ast.IntegerLiteral{
											Integer: 1,
										},
									},
								},
							},
							ast.Declaration{
								Name: token.Token{
									Type:   token.Identifier,
									Lexeme: "iteration",
								},
								Initializer: ast.Variable{
									Name: token.Token{
										Type:   token.Identifier,
										Lexeme: "i",
									},
								},
							},
						},
					},
				},
			},
			env:     NewEnvironment(Prefill("i", Number{5})),
			exports: make(map[string]Object),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			out := bytes.NewBufferString("")
			in := &Interpreter{
				exports: make(map[string]Object),
				out:     out,
				scope:   test.preload,
			}
			err := in.Execute(test.program)
			assert.Equal(t, test.exports, in.exports)
			assert.ErrorIs(t, err, test.err)
			assert.Equal(t, test.out, out.String())
			assert.Equal(t, test.env, in.scope)
		})
	}

	t.Run("return statement nested into deep scoping", func(t *testing.T) {
		src := `
		function random(n) {
			if n < 10 {
				var i = 0;
				while i < 10 {
					if i == 5 {
						return i;
					}
					i = i + 1;
				}
			} 
			return 100;
		}
		var first = random(0);
		var second = random(10);
		`
		program, err := ParseString(src)
		assert.Nil(t, err)
		in := NewInterpreter("", nil)
		for _, stmt := range program {
			_, err := in.execute(stmt)
			assert.Nil(t, err)
		}
		first, err := in.scope.Resolve("first")
		assert.Nil(t, err)
		assert.Equal(t, Number{5}, first)
		second, err := in.scope.Resolve("second")
		assert.Nil(t, err)
		assert.Equal(t, Number{100}, second)
	})

	t.Run("closure within a function", func(t *testing.T) {
		src := `
		function random(n) {
			function add(a) {
				return n + a;
			}
			return add(10);
		}
		var test = random(100);
		`
		program, err := ParseString(src)
		assert.Nil(t, err)
		in := NewInterpreter("", nil)
		for _, stmt := range program {
			_, err := in.execute(stmt)
			assert.Nil(t, err)
		}
		first, err := in.scope.Resolve("test")
		assert.Nil(t, err)
		assert.Equal(t, Number{110}, first)
	})

	t.Run("global closure", func(t *testing.T) {
		src := `
		var adder = 1000.;
		function random(n) {
			return adder + n;
		}
		var test = random(100);
		`
		program, err := ParseString(src)
		assert.Nil(t, err)
		in := NewInterpreter("", nil)
		for _, stmt := range program {
			_, err := in.execute(stmt)
			assert.Nil(t, err)
		}
		first, err := in.scope.Resolve("test")
		assert.Nil(t, err)
		assert.Equal(t, Number{1100}, first)
	})

	t.Run("export function", func(t *testing.T) {
		src := `
		function add(a, b) {
			return a + b;
		}
		export add;
		`
		program, err := ParseString(src)
		assert.Nil(t, err)
		in := NewInterpreter("", nil)
		for _, stmt := range program {
			_, err := in.execute(stmt)
			assert.Nil(t, err)
		}
		assert.Equal(t, Function{
			declaration: ast.Function{
				Name: token.Token{
					Type:   token.Identifier,
					Lexeme: "add",
				},
				Params: []token.Token{
					{
						Type:   token.Identifier,
						Lexeme: "a",
					},
					{
						Type:   token.Identifier,
						Lexeme: "b",
					},
				},
				Body: ast.Block{
					Body: []ast.StatementNode{
						ast.Return{
							Expression: ast.Infix{
								Operator: token.Token{
									Type:   token.Plus,
									Lexeme: "+",
								},
								LHS: ast.Variable{
									Name: token.Token{
										Type:   token.Identifier,
										Lexeme: "a",
									},
								},
								RHS: ast.Variable{
									Name: token.Token{
										Type:   token.Identifier,
										Lexeme: "b",
									},
								},
							},
						},
					},
				},
			},
			closure: in.scope,
		}, in.exports["add"])
	})

	t.Run("variable declaration followed by assignment", func(t *testing.T) {
		out := bytes.NewBufferString("")
		ev := Interpreter{
			out:   out,
			scope: NewEnvironment(),
		}

		_, err := ev.execute(ast.Declaration{
			Name: token.Token{
				Type:   token.Identifier,
				Lexeme: "name",
			},
			Initializer: ast.StringLiteral{
				String: "hello world",
			},
		})
		assert.Nil(t, err)
		val, err := ev.scope.Resolve("name")
		assert.Nil(t, err)
		assert.Equal(t, val, String{"hello world"})

		_, err = ev.execute(ast.ExpressionStatement{
			Expression: ast.Assignment{
				Name: token.Token{
					Type:   token.Identifier,
					Lexeme: "name",
				},
				Value: ast.StringLiteral{
					String: "goodbye",
				},
			},
		})
		assert.Nil(t, err)
		val, err = ev.scope.Resolve("name")
		assert.Equal(t, val, String{"goodbye"})
	})
}

func TestEnvironment_Resolve(t *testing.T) {
	tests := []struct {
		name string
		env  Environment
		key  string
		obj  Object
		err  error
	}{
		{
			name: "key is available in immediate scope",
			env: Environment{
				parent: nil,
				tbl: map[string]Object{
					"name": String{"crookdc"},
					"age":  Number{27.5},
				},
			},
			key: "name",
			obj: String{"crookdc"},
		},
		{
			name: "key is not available",
			env: Environment{
				parent: nil,
				tbl: map[string]Object{
					"age": Number{27.5},
				},
			},
			key: "name",
			err: ErrObjectNotDeclared,
		},
		{
			name: "key is available in parent scope",
			env: Environment{
				parent: &Environment{
					parent: nil,
					tbl: map[string]Object{
						"name": String{"crookdc"},
					},
				},
				tbl: map[string]Object{
					"age": Number{27.5},
				},
			},
			key: "name",
			obj: String{"crookdc"},
		},
		{
			name: "key is available in parent and immediate scope",
			env: Environment{
				parent: &Environment{
					parent: nil,
					tbl: map[string]Object{
						"name": String{"pia"},
					},
				},
				tbl: map[string]Object{
					"name": String{"crookdc"},
					"age":  Number{27.5},
				},
			},
			key: "name",
			obj: String{"crookdc"},
		},
		{
			name: "key is available in grandparent scope",
			env: Environment{
				parent: &Environment{
					parent: &Environment{
						tbl: map[string]Object{
							"name": String{"crookdc"},
						},
					},
					tbl: map[string]Object{},
				},
				tbl: map[string]Object{
					"age": Number{27.5},
				},
			},
			key: "name",
			obj: String{"crookdc"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			obj, err := test.env.Resolve(test.key)
			assert.ErrorIs(t, err, test.err)
			assert.Equal(t, test.obj, obj)
		})
	}
}

func TestEnvironment_Assign(t *testing.T) {
	tests := []struct {
		name  string
		env   Environment
		key   string
		value Object
		after Environment
		err   error
	}{
		{
			name: "key is available in immediate scope",
			env: Environment{
				parent: nil,
				tbl: map[string]Object{
					"name": String{"crookdc"},
					"age":  Number{27.5},
				},
			},
			key:   "name",
			value: String{"pia"},
			after: Environment{
				parent: nil,
				tbl: map[string]Object{
					"name": String{"pia"},
					"age":  Number{27.5},
				},
			},
		},
		{
			name: "key is not available",
			env: Environment{
				parent: nil,
				tbl: map[string]Object{
					"age": Number{27.5},
				},
			},
			key: "name",
			after: Environment{
				parent: nil,
				tbl: map[string]Object{
					"age": Number{27.5},
				},
			},
			err: ErrObjectNotDeclared,
		},
		{
			name: "key is available in parent scope",
			env: Environment{
				parent: &Environment{
					parent: nil,
					tbl: map[string]Object{
						"name": String{"crookdc"},
					},
				},
				tbl: map[string]Object{
					"age": Number{27.5},
				},
			},
			key:   "name",
			value: Number{123.12},
			after: Environment{
				parent: &Environment{
					parent: nil,
					tbl: map[string]Object{
						"name": Number{123.12},
					},
				},
				tbl: map[string]Object{
					"age": Number{27.5},
				},
			},
		},
		{
			name: "key is available in parent and immediate scope",
			env: Environment{
				parent: &Environment{
					parent: nil,
					tbl: map[string]Object{
						"name": String{"pia"},
					},
				},
				tbl: map[string]Object{
					"name": String{"crookdc"},
					"age":  Number{27.5},
				},
			},
			key:   "name",
			value: Boolean{true},
			after: Environment{
				parent: &Environment{
					parent: nil,
					tbl: map[string]Object{
						"name": String{"pia"},
					},
				},
				tbl: map[string]Object{
					"name": Boolean{true},
					"age":  Number{27.5},
				},
			},
		},
		{
			name: "key is available in grandparent scope",
			env: Environment{
				parent: &Environment{
					parent: &Environment{
						tbl: map[string]Object{
							"name": String{"crookdc"},
						},
					},
					tbl: map[string]Object{},
				},
				tbl: map[string]Object{
					"age": Number{27.5},
				},
			},
			key:   "name",
			value: String{"John Smith"},
			after: Environment{
				parent: &Environment{
					parent: &Environment{
						tbl: map[string]Object{
							"name": String{"John Smith"},
						},
					},
					tbl: map[string]Object{},
				},
				tbl: map[string]Object{
					"age": Number{27.5},
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := test.env.Assign(test.key, test.value)
			assert.ErrorIs(t, err, test.err)
			assert.Equal(t, test.after, test.env)
		})
	}
}

func TestEnvironment_Declare(t *testing.T) {
	tests := []struct {
		name  string
		env   Environment
		key   string
		value Object
		after Environment
	}{
		{
			name: "key is not available in immediate scope",
			env: Environment{
				parent: nil,
				tbl: map[string]Object{
					"age": Number{27.5},
				},
			},
			key:   "name",
			value: String{"pia"},
			after: Environment{
				parent: nil,
				tbl: map[string]Object{
					"name": String{"pia"},
					"age":  Number{27.5},
				},
			},
		},
		{
			name: "key is available in parent scope",
			env: Environment{
				parent: &Environment{
					parent: nil,
					tbl: map[string]Object{
						"name": String{"crookdc"},
					},
				},
				tbl: map[string]Object{
					"age": Number{27.5},
				},
			},
			key:   "name",
			value: Number{123.12},
			after: Environment{
				parent: &Environment{
					parent: nil,
					tbl: map[string]Object{
						"name": String{"crookdc"},
					},
				},
				tbl: map[string]Object{
					"age":  Number{27.5},
					"name": Number{123.12},
				},
			},
		},
		{
			name: "key is available in parent and immediate scope",
			env: Environment{
				parent: &Environment{
					parent: nil,
					tbl: map[string]Object{
						"name": String{"pia"},
					},
				},
				tbl: map[string]Object{
					"name": String{"crookdc"},
					"age":  Number{27.5},
				},
			},
			key:   "name",
			value: Boolean{true},
			after: Environment{
				parent: &Environment{
					parent: nil,
					tbl: map[string]Object{
						"name": String{"pia"},
					},
				},
				tbl: map[string]Object{
					"name": Boolean{true},
					"age":  Number{27.5},
				},
			},
		},
		{
			name: "key is available in grandparent scope",
			env: Environment{
				parent: &Environment{
					parent: &Environment{
						tbl: map[string]Object{
							"name": String{"crookdc"},
						},
					},
					tbl: map[string]Object{},
				},
				tbl: map[string]Object{
					"age":  Number{27.5},
					"name": String{"crookdc"},
				},
			},
			key:   "name",
			value: String{"John Smith"},
			after: Environment{
				parent: &Environment{
					parent: &Environment{
						tbl: map[string]Object{
							"name": String{"crookdc"},
						},
					},
					tbl: map[string]Object{},
				},
				tbl: map[string]Object{
					"age":  Number{27.5},
					"name": String{"John Smith"},
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.env.Declare(test.key, test.value)
			assert.Equal(t, test.after, test.env)
		})
	}
}

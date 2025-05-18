package squeak

import (
	"bytes"
	"github.com/crookdc/pia/squeak/internal/ast"
	"github.com/crookdc/pia/squeak/internal/token"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestEvaluator_expression(t *testing.T) {
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
			err: ErrRuntimeFault,
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
			err: ErrRuntimeFault,
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
			err: ErrRuntimeFault,
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
			err: ErrRuntimeFault,
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
			err: ErrRuntimeFault,
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
			err: ErrRuntimeFault,
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
			err: ErrRuntimeFault,
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
			err: ErrRuntimeFault,
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
			err: ErrRuntimeFault,
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
			err: ErrRuntimeFault,
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
			err: ErrRuntimeFault,
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
			err: ErrRuntimeFault,
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
			err: ErrRuntimeFault,
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
			err: ErrRuntimeFault,
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
			err: ErrRuntimeFault,
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
			err: ErrRuntimeFault,
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
			err: ErrRuntimeFault,
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
			err: ErrRuntimeFault,
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
			err: ErrRuntimeFault,
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
			err: ErrRuntimeFault,
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
			err: ErrRuntimeFault,
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
			err: ErrRuntimeFault,
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
			err: ErrRuntimeFault,
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
			err: ErrRuntimeFault,
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
			err: ErrRuntimeFault,
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
			err: ErrRuntimeFault,
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
			err: ErrRuntimeFault,
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
			err: ErrRuntimeFault,
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
			err: ErrRuntimeFault,
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
			err: ErrRuntimeFault,
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
			err: ErrRuntimeFault,
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
			err: ErrRuntimeFault,
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
			err: ErrRuntimeFault,
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
			err: ErrRuntimeFault,
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
			err: ErrRuntimeFault,
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
			err: ErrRuntimeFault,
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
			err: ErrRuntimeFault,
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
			err: ErrRuntimeFault,
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
			err: ErrRuntimeFault,
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
			err: ErrRuntimeFault,
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
			err: ErrRuntimeFault,
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
					Boolean: true,
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
			obj: Boolean{false},
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
				RHS: ast.NilLiteral{},
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
			obj: Boolean{true},
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
			obj: Boolean{true},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			obj, err := (&Evaluator{}).expression(test.node)
			assert.ErrorIs(t, err, test.err)
			if err == nil {
				// The returned value is only interesting if the returned error is nil. If the error is not nil then the
				// returned object does not have a defined rule to its state and should never be used anyway.
				assert.Equal(t, test.obj, obj)
			}
		})
	}
}

func TestEvaluator_statement(t *testing.T) {
	tests := []struct {
		name    string
		stmt    ast.StatementNode
		preload *Environment

		out string
		env *Environment
		err error
	}{
		{
			name:    "print writes string object to output",
			preload: NewEnvironment(),
			stmt: ast.Print{
				Expression: ast.Infix{
					Operator: token.Token{
						Type:   token.Plus,
						Lexeme: "+",
					},
					LHS: ast.StringLiteral{
						String: "hello",
					},
					RHS: ast.StringLiteral{
						String: " world",
					},
				},
			},
			out: "hello world",
			env: NewEnvironment(),
		},
		{
			name:    "print writes whole number object to output",
			preload: NewEnvironment(),
			stmt: ast.Print{
				Expression: ast.Infix{
					Operator: token.Token{
						Type:   token.Plus,
						Lexeme: "+",
					},
					LHS: ast.IntegerLiteral{
						Integer: 15,
					},
					RHS: ast.IntegerLiteral{
						Integer: 30,
					},
				},
			},
			out: "45.",
			env: NewEnvironment(),
		},
		{
			name:    "print writes number object to output",
			preload: NewEnvironment(),
			stmt: ast.Print{
				Expression: ast.Infix{
					Operator: token.Token{
						Type:   token.Slash,
						Lexeme: "/",
					},
					LHS: ast.IntegerLiteral{
						Integer: 10,
					},
					RHS: ast.IntegerLiteral{
						Integer: 3,
					},
				},
			},
			out: "3.333333",
			env: NewEnvironment(),
		},
		{
			name:    "print writes boolean object to output",
			preload: NewEnvironment(),
			stmt: ast.Print{
				Expression: ast.BooleanLiteral{Boolean: true},
			},
			out: "true",
			env: NewEnvironment(),
		},
		{
			name:    "variable declaration without initializer",
			preload: NewEnvironment(),
			stmt: ast.Var{
				Name: token.Token{
					Type:   token.Identifier,
					Lexeme: "name",
				},
			},
			env: NewEnvironment(Prefill("name", nil)),
		},
		{
			name:    "variable declaration with explicit nil initializer",
			preload: NewEnvironment(),
			stmt: ast.Var{
				Name: token.Token{
					Type:   token.Identifier,
					Lexeme: "name",
				},
				Initializer: ast.NilLiteral{},
			},
			env: NewEnvironment(Prefill("name", nil)),
		},
		{
			name:    "variable declaration with initializer",
			preload: NewEnvironment(),
			stmt: ast.Var{
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
			},
			env: NewEnvironment(Prefill("name", String{"hellogoodbye"})),
		},
		{
			name:    "block that assigns in parent scope and declares new variable in local scope",
			preload: NewEnvironment(Prefill("name", Number{1.123})),
			stmt: ast.Block{
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
					ast.Var{
						Name: token.Token{
							Type:   token.Identifier,
							Lexeme: "age",
						},
						Initializer: ast.IntegerLiteral{
							Integer: 27,
						},
					},
				},
			},
			env: NewEnvironment(Prefill("name", Number{1556.12})),
		},
		{
			name:    "nested blocks",
			preload: NewEnvironment(Prefill("name", Number{1.123})),
			stmt: ast.Block{
				Body: []ast.StatementNode{
					ast.Print{
						Expression: ast.Variable{
							Name: token.Token{
								Type:   token.Identifier,
								Lexeme: "name",
							},
						},
					},
					ast.Block{
						Body: []ast.StatementNode{
							ast.Var{
								Name: token.Token{
									Type:   token.Identifier,
									Lexeme: "name",
								},
								Initializer: ast.StringLiteral{String: "crookdc"},
							},
							ast.Print{
								Expression: ast.Variable{
									Name: token.Token{
										Type:   token.Identifier,
										Lexeme: "name",
									},
								},
							},
						},
					},
				},
			},
			env: NewEnvironment(Prefill("name", Number{1.123})),
			out: "1.123crookdc",
		},
		{
			name:    "empty block",
			preload: NewEnvironment(Prefill("name", Number{1.123})),
			stmt:    ast.Block{},
			env:     NewEnvironment(Prefill("name", Number{1.123})),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			out := bytes.NewBufferString("")
			ev := Evaluator{
				out: out,
				env: test.preload,
			}
			err := ev.statement(test.stmt)
			assert.ErrorIs(t, err, test.err)
			assert.Equal(t, test.out, out.String())
			assert.Equal(t, test.env, ev.env)
		})
	}

	t.Run("variable declaration followed by assignment", func(t *testing.T) {
		out := bytes.NewBufferString("")
		ev := Evaluator{
			out: out,
			env: NewEnvironment(),
		}

		err := ev.statement(ast.Var{
			Name: token.Token{
				Type:   token.Identifier,
				Lexeme: "name",
			},
			Initializer: ast.StringLiteral{
				String: "hello world",
			},
		})
		assert.Nil(t, err)
		val, err := ev.env.Resolve("name")
		assert.Nil(t, err)
		assert.Equal(t, val, String{"hello world"})

		err = ev.statement(ast.ExpressionStatement{
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
		val, err = ev.env.Resolve("name")
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
			err: ErrRuntimeFault,
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
			err: ErrRuntimeFault,
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

package squeak

import (
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
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			obj, err := (&Evaluator{}).expression(test.node)
			assert.ErrorIs(t, err, test.err)
			assert.Equal(t, test.obj, obj)
		})
	}
}

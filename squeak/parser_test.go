package squeak

import (
	"github.com/crookdc/pia/squeak/ast"
	"github.com/crookdc/pia/squeak/token"
	"github.com/stretchr/testify/assert"
	"io"
	"strings"
	"testing"
)

func TestParser_Next(t *testing.T) {
	tests := []struct {
		src      string
		expected ast.StatementNode
		err      error
	}{
		{
			src: "a;",
			expected: ast.ExpressionStatement{
				Expression: ast.Variable{
					Name: token.Token{
						Type:   token.Identifier,
						Lexeme: "a",
					},
				},
			},
		},
		{
			src: "{ break; continue; }",
			expected: ast.Block{
				Body: []ast.StatementNode{
					ast.Break{},
					ast.Continue{},
				},
			},
		},
		{
			src: "import \"pia:response\";",
			expected: ast.Import{
				Source: ast.StringLiteral{
					String: "pia:response",
				},
			},
		},
		{
			src: "import some_variable;",
			expected: ast.Import{
				Source: ast.Variable{
					Name: token.Token{
						Type:   token.Identifier,
						Lexeme: "some_variable",
					},
				},
			},
		},
		{
			src: "import true;",
			err: ErrUnrecognizedExpression,
		},
		{
			src: "import 15;",
			err: ErrUnrecognizedExpression,
		},
		{
			src: "import 15.4;",
			err: ErrUnrecognizedExpression,
		},
		{
			src: "export true;",
			err: ErrUnrecognizedExpression,
		},
		{
			src: "export 13;",
			err: ErrUnrecognizedExpression,
		},
		{
			src: "export 134.5;",
			err: ErrUnrecognizedExpression,
		},
		{
			src: "export \"some string\";",
			err: ErrUnrecognizedExpression,
		},
		{
			src: "export my_var;",
			expected: ast.Export{
				Name: token.Token{
					Type:   token.Identifier,
					Lexeme: "my_var",
				},
				Value: ast.Variable{
					Name: token.Token{
						Type:   token.Identifier,
						Lexeme: "my_var",
					},
				},
			},
		},
		{
			src: `
			{
				import "pia:request";
			}
			`,
			expected: ast.Block{
				Body: []ast.StatementNode{
					ast.Import{
						Source: ast.StringLiteral{
							String: "pia:request",
						},
					},
				},
			},
		},
		{
			src: `
			function add(a, b) {
				print(a + b);
				return 42.;
			}
			`,
			expected: ast.Function{
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
						ast.ExpressionStatement{
							Expression: ast.Call{
								Callee: ast.Variable{
									Name: token.Token{
										Type:   token.Identifier,
										Lexeme: "print",
									},
								},
								Operator: token.Token{
									Type:   token.LeftParenthesis,
									Lexeme: "(",
								},
								Args: []ast.ExpressionNode{
									ast.Infix{
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
						ast.Return{
							Expression: ast.FloatLiteral{
								Float: 42,
							},
						},
					},
				},
			},
		},
		{
			src: `
			function clock() {
				# This is where we would put some logic!
			}
			`,
			expected: ast.Function{
				Name: token.Token{
					Type:   token.Identifier,
					Lexeme: "clock",
				},
				Params: []token.Token{},
				Body: ast.Block{
					Body: []ast.StatementNode{},
				},
			},
		},
		{
			src: "a + b;",
			expected: ast.ExpressionStatement{
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
		{
			src: "a + b - 1;",
			expected: ast.ExpressionStatement{
				Expression: ast.Infix{
					Operator: token.Token{
						Type:   token.Minus,
						Lexeme: "-",
					},
					LHS: ast.Infix{
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
					RHS: ast.IntegerLiteral{
						Integer: 1,
					},
				},
			},
		},
		{
			src: "name + \"is a good developer\";",
			expected: ast.ExpressionStatement{
				Expression: ast.Infix{
					Operator: token.Token{
						Type:   token.Plus,
						Lexeme: "+",
					},
					LHS: ast.Variable{
						Name: token.Token{
							Type:   token.Identifier,
							Lexeme: "name",
						},
					},
					RHS: ast.StringLiteral{
						String: "is a good developer",
					},
				},
			},
		},
		{
			src: "a + b * c;",
			expected: ast.ExpressionStatement{
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
					RHS: ast.Infix{
						Operator: token.Token{
							Type:   token.Asterisk,
							Lexeme: "*",
						},
						LHS: ast.Variable{
							Name: token.Token{
								Type:   token.Identifier,
								Lexeme: "b",
							},
						},
						RHS: ast.Variable{
							Name: token.Token{
								Type:   token.Identifier,
								Lexeme: "c",
							},
						},
					},
				},
			},
		},
		{
			src: "(a + b) * c;",
			expected: ast.ExpressionStatement{
				Expression: ast.Infix{
					Operator: token.Token{
						Type:   token.Asterisk,
						Lexeme: "*",
					},
					LHS: ast.Grouping{
						Group: ast.Infix{
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
					RHS: ast.Variable{
						Name: token.Token{
							Type:   token.Identifier,
							Lexeme: "c",
						},
					},
				},
			},
		},
		{
			src: "5 + -1 <= 6 * 5;",
			expected: ast.ExpressionStatement{
				Expression: ast.Infix{
					Operator: token.Token{
						Type:   token.LessEqual,
						Lexeme: "<=",
					},
					LHS: ast.Infix{
						Expression: ast.Expression{},
						Operator: token.Token{
							Type:   token.Plus,
							Lexeme: "+",
						},
						LHS: ast.IntegerLiteral{Integer: 5},
						RHS: ast.Prefix{
							Operator: token.Token{
								Type:   token.Minus,
								Lexeme: "-",
							},
							Target: ast.IntegerLiteral{
								Integer: 1,
							},
						},
					},
					RHS: ast.Infix{
						Expression: ast.Expression{},
						Operator: token.Token{
							Type:   token.Asterisk,
							Lexeme: "*",
						},
						LHS: ast.IntegerLiteral{Integer: 6},
						RHS: ast.IntegerLiteral{Integer: 5},
					},
				},
			},
		},
		{
			src: "\n5\n",
			// Since linefeed characters aren't much of a concern for the Squeak parser it makes sense that the error
			// actually appears on line 3, where we reach the end of the stream without having encountered a semicolon.
			err: SyntaxError{Line: 3},
		},
		{
			src: "\n5\n;\n",
			// This is an example of where the linefeed character is totally ignored (other than during line counting)
			// and it is therefor okay to defer the statement terminator (semicolon) to the next line (or several lines
			// down).
			expected: ast.ExpressionStatement{
				Expression: ast.IntegerLiteral{Integer: 5},
			},
		},
		{
			src: "",
			err: io.EOF,
		},
		{
			src: "\n\t\t\n# Hello world\n",
			err: io.EOF,
		},
		{
			src: "var name = \"crookdc\";",
			expected: ast.Declaration{
				Name: token.Token{
					Type:   token.Identifier,
					Lexeme: "name",
				},
				Initializer: ast.StringLiteral{
					String: "crookdc",
				},
			},
		},
		{
			src: "var name;",
			expected: ast.Declaration{
				Name: token.Token{
					Type:   token.Identifier,
					Lexeme: "name",
				},
			},
		},
		{
			src: "var name ? \"crookdc\";",
			err: SyntaxError{
				Line: 1,
			},
		},
		{
			src: "var name = nil;",
			expected: ast.Declaration{
				Name: token.Token{
					Type:   token.Identifier,
					Lexeme: "name",
				},
				Initializer: ast.NilLiteral{},
			},
		},
		{
			src: "name = \"crookdc\";",
			expected: ast.ExpressionStatement{
				Expression: ast.Assignment{
					Name: token.Token{
						Type:   token.Identifier,
						Lexeme: "name",
					},
					Value: ast.StringLiteral{String: "crookdc"},
				},
			},
		},
		{
			src: "12.44 + 12;",
			expected: ast.ExpressionStatement{
				Expression: ast.Infix{
					Operator: token.Token{
						Type:   token.Plus,
						Lexeme: "+",
					},
					LHS: ast.FloatLiteral{
						Float: 12.44,
					},
					RHS: ast.IntegerLiteral{
						Integer: 12,
					},
				},
			},
		},
		{
			src: "0.444456;",
			expected: ast.ExpressionStatement{
				Expression: ast.FloatLiteral{
					Float: 0.444456,
				},
			},
		},
		{
			src: "50.;",
			expected: ast.ExpressionStatement{
				Expression: ast.FloatLiteral{
					Float: 50,
				},
			},
		},
		{
			src: "run();",
			expected: ast.ExpressionStatement{
				Expression: ast.Call{
					Callee: ast.Variable{
						Name: token.Token{
							Type:   token.Identifier,
							Lexeme: "run",
						},
					},
					Operator: token.Token{
						Type:   token.LeftParenthesis,
						Lexeme: "(",
					},
					Args: []ast.ExpressionNode{},
				},
			},
		},
		{
			src: "run(5 + 1002, n);",
			expected: ast.ExpressionStatement{
				Expression: ast.Call{
					Callee: ast.Variable{
						Name: token.Token{
							Type:   token.Identifier,
							Lexeme: "run",
						},
					},
					Operator: token.Token{
						Type:   token.LeftParenthesis,
						Lexeme: "(",
					},
					Args: []ast.ExpressionNode{
						ast.Infix{
							Operator: token.Token{
								Type:   token.Plus,
								Lexeme: "+",
							},
							LHS: ast.IntegerLiteral{
								Integer: 5,
							},
							RHS: ast.IntegerLiteral{
								Integer: 1002,
							},
						},
						ast.Variable{
							Name: token.Token{
								Type:   token.Identifier,
								Lexeme: "n",
							},
						},
					},
				},
			},
		},
		{
			src: "factory()(5 + 1002, n)(n);",
			expected: ast.ExpressionStatement{
				Expression: ast.Call{
					Callee: ast.Call{
						Callee: ast.Call{
							Callee: ast.Variable{
								Name: token.Token{
									Type:   token.Identifier,
									Lexeme: "factory",
								},
							},
							Operator: token.Token{
								Type:   token.LeftParenthesis,
								Lexeme: "(",
							},
							Args: []ast.ExpressionNode{},
						},
						Operator: token.Token{
							Type:   token.LeftParenthesis,
							Lexeme: "(",
						},
						Args: []ast.ExpressionNode{
							ast.Infix{
								Operator: token.Token{
									Type:   token.Plus,
									Lexeme: "+",
								},
								LHS: ast.IntegerLiteral{
									Integer: 5,
								},
								RHS: ast.IntegerLiteral{
									Integer: 1002,
								},
							},
							ast.Variable{
								Name: token.Token{
									Type:   token.Identifier,
									Lexeme: "n",
								},
							},
						},
					},
					Operator: token.Token{
						Type:   token.LeftParenthesis,
						Lexeme: "(",
					},
					Args: []ast.ExpressionNode{
						ast.Variable{
							Name: token.Token{
								Type:   token.Identifier,
								Lexeme: "n",
							},
						},
					},
				},
			},
		},
		{
			src:      ";",
			expected: ast.Noop{},
		},
		{
			src: "while true {}",
			expected: ast.While{
				Condition: ast.BooleanLiteral{
					Boolean: true,
				},
				Body: ast.Block{
					Body: []ast.StatementNode{},
				},
			},
		},
		{
			src: "if a > b print(a); else print(b);",
			expected: ast.If{
				Condition: ast.Infix{
					Operator: token.Token{
						Type:   token.Greater,
						Lexeme: ">",
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
				Then: ast.ExpressionStatement{
					Expression: ast.Call{
						Callee: ast.Variable{
							Name: token.Token{
								Type:   token.Identifier,
								Lexeme: "print",
							},
						},
						Operator: token.Token{
							Type:   token.LeftParenthesis,
							Lexeme: "(",
						},
						Args: []ast.ExpressionNode{
							ast.Variable{
								Name: token.Token{
									Type:   token.Identifier,
									Lexeme: "a",
								},
							},
						},
					},
				},
				Else: ast.ExpressionStatement{
					Expression: ast.Call{
						Callee: ast.Variable{
							Name: token.Token{
								Type:   token.Identifier,
								Lexeme: "print",
							},
						},
						Operator: token.Token{
							Type:   token.LeftParenthesis,
							Lexeme: "(",
						},
						Args: []ast.ExpressionNode{
							ast.Variable{
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
		{
			src: "if a > b print(a);",
			expected: ast.If{
				Condition: ast.Infix{
					Operator: token.Token{
						Type:   token.Greater,
						Lexeme: ">",
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
				Then: ast.ExpressionStatement{
					Expression: ast.Call{
						Callee: ast.Variable{
							Name: token.Token{
								Type:   token.Identifier,
								Lexeme: "print",
							},
						},
						Operator: token.Token{
							Type:   token.LeftParenthesis,
							Lexeme: "(",
						},
						Args: []ast.ExpressionNode{
							ast.Variable{
								Name: token.Token{
									Type:   token.Identifier,
									Lexeme: "a",
								},
							},
						},
					},
				},
			},
		},
		{
			src: "if a if b print(b); else print(c);",
			expected: ast.If{
				Condition: ast.Variable{
					Name: token.Token{
						Type:   token.Identifier,
						Lexeme: "a",
					},
				},
				Then: ast.If{
					Condition: ast.Variable{
						Name: token.Token{
							Type:   token.Identifier,
							Lexeme: "b",
						},
					},
					Then: ast.ExpressionStatement{
						Expression: ast.Call{
							Callee: ast.Variable{
								Name: token.Token{
									Type:   token.Identifier,
									Lexeme: "print",
								},
							},
							Operator: token.Token{
								Type:   token.LeftParenthesis,
								Lexeme: "(",
							},
							Args: []ast.ExpressionNode{
								ast.Variable{
									Name: token.Token{
										Type:   token.Identifier,
										Lexeme: "b",
									},
								},
							},
						},
					},
					Else: ast.ExpressionStatement{
						Expression: ast.Call{
							Callee: ast.Variable{
								Name: token.Token{
									Type:   token.Identifier,
									Lexeme: "print",
								},
							},
							Operator: token.Token{
								Type:   token.LeftParenthesis,
								Lexeme: "(",
							},
							Args: []ast.ExpressionNode{
								ast.Variable{
									Name: token.Token{
										Type:   token.Identifier,
										Lexeme: "c",
									},
								},
							},
						},
					},
				},
			},
		},
		{
			src: "{ a + b; a = 2.; }",
			expected: ast.Block{
				Body: []ast.StatementNode{
					ast.ExpressionStatement{
						Expression: ast.Infix{
							Expression: ast.Expression{},
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
					ast.ExpressionStatement{
						Expression: ast.Assignment{
							Name: token.Token{
								Type:   token.Identifier,
								Lexeme: "a",
							},
							Value: ast.FloatLiteral{
								Float: 2.0,
							},
						},
					},
				},
			},
		},
		{
			src: "while a < b and b >= 3 { print(a); }",
			expected: ast.While{
				Condition: ast.Logical{
					Operator: token.Token{
						Type:   token.And,
						Lexeme: "and",
					},
					LHS: ast.Infix{
						Operator: token.Token{
							Type:   token.Less,
							Lexeme: "<",
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
					RHS: ast.Infix{
						Operator: token.Token{
							Type:   token.GreaterEqual,
							Lexeme: ">=",
						},
						LHS: ast.Variable{
							Name: token.Token{
								Type:   token.Identifier,
								Lexeme: "b",
							},
						},
						RHS: ast.IntegerLiteral{
							Integer: 3,
						},
					},
				},
				Body: ast.Block{
					Body: []ast.StatementNode{
						ast.ExpressionStatement{
							Expression: ast.Call{
								Callee: ast.Variable{
									Name: token.Token{
										Type:   token.Identifier,
										Lexeme: "print",
									},
								},
								Operator: token.Token{
									Type:   token.LeftParenthesis,
									Lexeme: "(",
								},
								Args: []ast.ExpressionNode{
									ast.Variable{
										Name: token.Token{
											Type:   token.Identifier,
											Lexeme: "a",
										},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			src: "while a < b { print(a); print(b); }",
			expected: ast.While{
				Condition: ast.Infix{
					Operator: token.Token{
						Type:   token.Less,
						Lexeme: "<",
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
				Body: ast.Block{
					Body: []ast.StatementNode{
						ast.ExpressionStatement{
							Expression: ast.Call{
								Callee: ast.Variable{
									Name: token.Token{
										Type:   token.Identifier,
										Lexeme: "print",
									},
								},
								Operator: token.Token{
									Type:   token.LeftParenthesis,
									Lexeme: "(",
								},
								Args: []ast.ExpressionNode{
									ast.Variable{
										Name: token.Token{
											Type:   token.Identifier,
											Lexeme: "a",
										},
									},
								},
							},
						},
						ast.ExpressionStatement{
							Expression: ast.Call{
								Callee: ast.Variable{
									Name: token.Token{
										Type:   token.Identifier,
										Lexeme: "print",
									},
								},
								Operator: token.Token{
									Type:   token.LeftParenthesis,
									Lexeme: "(",
								},
								Args: []ast.ExpressionNode{
									ast.Variable{
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
			},
		},
		{
			src: "{}",
			expected: ast.Block{
				Body: []ast.StatementNode{},
			},
		},
		{
			src: "1 == 1 and b;",
			expected: ast.ExpressionStatement{
				Expression: ast.Logical{
					Operator: token.Token{
						Type:   token.And,
						Lexeme: "and",
					},
					LHS: ast.Infix{
						Operator: token.Token{
							Type:   token.Equals,
							Lexeme: "==",
						},
						LHS: ast.IntegerLiteral{
							Integer: 1,
						},
						RHS: ast.IntegerLiteral{
							Integer: 1,
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
		{
			src: "1 == 1 and b or c;",
			expected: ast.ExpressionStatement{
				Expression: ast.Logical{
					Operator: token.Token{
						Type:   token.Or,
						Lexeme: "or",
					},
					LHS: ast.Logical{
						Operator: token.Token{
							Type:   token.And,
							Lexeme: "and",
						},
						LHS: ast.Infix{
							Operator: token.Token{
								Type:   token.Equals,
								Lexeme: "==",
							},
							LHS: ast.IntegerLiteral{
								Integer: 1,
							},
							RHS: ast.IntegerLiteral{
								Integer: 1,
							},
						},
						RHS: ast.Variable{
							Name: token.Token{
								Type:   token.Identifier,
								Lexeme: "b",
							},
						},
					},
					RHS: ast.Variable{
						Name: token.Token{
							Type:   token.Identifier,
							Lexeme: "c",
						},
					},
				},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.src, func(t *testing.T) {
			lx, err := NewLexer(strings.NewReader(test.src))
			assert.Nil(t, err)
			plx, err := NewPeekingLexer(lx)
			assert.Nil(t, err)

			ps := Parser{lx: plx}
			n, err := ps.Next()
			assert.ErrorIs(t, err, test.err)
			if err == nil {
				assert.Equal(t, test.expected, n)
			}
		})
	}

	t.Run("clears current statement if error occurs", func(t *testing.T) {
		src := `
		a +/ b; # This line contains an invalid evaluate 
		a + b;`
		lx, err := NewLexer(strings.NewReader(src))
		assert.Nil(t, err)
		plx, err := NewPeekingLexer(lx)
		assert.Nil(t, err)

		ps := Parser{lx: plx}
		n, err := ps.Next()
		assert.ErrorIs(t, err, SyntaxError{Line: 2})

		n, err = ps.Next()
		assert.Nil(t, err)
		assert.Equal(t, ast.ExpressionStatement{
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
		}, n)
	})
}

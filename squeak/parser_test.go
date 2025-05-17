package squeak

import (
	"github.com/crookdc/pia/squeak/internal/ast"
	"github.com/crookdc/pia/squeak/internal/token"
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
			src: "print \"hello world\";",
			expected: ast.Print{
				Expression: ast.StringLiteral{
					String: "hello world",
				},
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
			expected: ast.Var{
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
			expected: ast.Var{
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
			expected: ast.Var{
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
		a +/ b; # This line contains an invalid expression 
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

package squeak

import (
	"github.com/crookdc/pia/squeak/internal/ast"
	"github.com/crookdc/pia/squeak/internal/token"
	"github.com/stretchr/testify/assert"
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
				Expression: ast.Identifier{
					Identifier: "a",
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
					LHS: ast.Identifier{
						Identifier: "a",
					},
					RHS: ast.Identifier{
						Identifier: "b",
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
						LHS: ast.Identifier{
							Identifier: "a",
						},
						RHS: ast.Identifier{
							Identifier: "b",
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
					LHS: ast.Identifier{
						Identifier: "name",
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
					LHS: ast.Identifier{
						Identifier: "a",
					},
					RHS: ast.Infix{
						Operator: token.Token{
							Type:   token.Asterisk,
							Lexeme: "*",
						},
						LHS: ast.Identifier{
							Identifier: "b",
						},
						RHS: ast.Identifier{
							Identifier: "c",
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
							LHS: ast.Identifier{
								Identifier: "a",
							},
							RHS: ast.Identifier{
								Identifier: "b",
							},
						},
					},
					RHS: ast.Identifier{
						Identifier: "c",
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
							RHS: ast.IntegerLiteral{
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
	}
	for _, test := range tests {
		t.Run(test.src, func(t *testing.T) {
			lx, err := NewLexer(strings.NewReader(test.src))
			assert.Nil(t, err)
			plx, err := NewPeekingLexer(lx)
			assert.Nil(t, err)

			ps := Parser{lx: plx}
			n, err := ps.Next()
			assert.Equal(t, test.err, err)
			assert.Equal(t, test.expected, n)
		})
	}
}

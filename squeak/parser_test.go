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
	}{
		{
			src: "a;",
			expected: ast.ExpressionStatement{
				Expression: ast.IdentifierExpression{
					Identifier: "a",
				},
			},
		},
		{
			src: "a + b;",
			expected: ast.ExpressionStatement{
				Expression: ast.InfixExpression{
					Operator: token.Token{
						Type:    token.Plus,
						Literal: "+",
					},
					LHS: ast.IdentifierExpression{
						Identifier: "a",
					},
					RHS: ast.IdentifierExpression{
						Identifier: "b",
					},
				},
			},
		},
		{
			src: "a + b - 1;",
			expected: ast.ExpressionStatement{
				Expression: ast.InfixExpression{
					Operator: token.Token{
						Type:    token.Minus,
						Literal: "-",
					},
					LHS: ast.InfixExpression{
						Operator: token.Token{
							Type:    token.Plus,
							Literal: "+",
						},
						LHS: ast.IdentifierExpression{
							Identifier: "a",
						},
						RHS: ast.IdentifierExpression{
							Identifier: "b",
						},
					},
					RHS: ast.IntegerExpression{
						Integer: 1,
					},
				},
			},
		},
		{
			src: "name + \"is a good developer\";",
			expected: ast.ExpressionStatement{
				Expression: ast.InfixExpression{
					Operator: token.Token{
						Type:    token.Plus,
						Literal: "+",
					},
					LHS: ast.IdentifierExpression{
						Identifier: "name",
					},
					RHS: ast.StringExpression{
						String: "is a good developer",
					},
				},
			},
		},
		{
			src: "a + b * c;",
			expected: ast.ExpressionStatement{
				Expression: ast.InfixExpression{
					Operator: token.Token{
						Type:    token.Plus,
						Literal: "+",
					},
					LHS: ast.IdentifierExpression{
						Identifier: "a",
					},
					RHS: ast.InfixExpression{
						Operator: token.Token{
							Type:    token.Asterisk,
							Literal: "*",
						},
						LHS: ast.IdentifierExpression{
							Identifier: "b",
						},
						RHS: ast.IdentifierExpression{
							Identifier: "c",
						},
					},
				},
			},
		},
		{
			src: "(a + b) * c;",
			expected: ast.ExpressionStatement{
				Expression: ast.InfixExpression{
					Operator: token.Token{
						Type:    token.Asterisk,
						Literal: "*",
					},
					LHS: ast.InfixExpression{
						Operator: token.Token{
							Type:    token.Plus,
							Literal: "+",
						},
						LHS: ast.IdentifierExpression{
							Identifier: "a",
						},
						RHS: ast.IdentifierExpression{
							Identifier: "b",
						},
					},
					RHS: ast.IdentifierExpression{
						Identifier: "c",
					},
				},
			},
		},
		{
			src: "let a = (a + b) * c;",
			expected: ast.LetStatement{
				Assignment: ast.InfixExpression{
					Operator: token.Token{
						Type:    token.Assign,
						Literal: "=",
					},
					LHS: ast.IdentifierExpression{
						Identifier: "a",
					},
					RHS: ast.InfixExpression{
						Operator: token.Token{
							Type:    token.Asterisk,
							Literal: "*",
						},
						LHS: ast.InfixExpression{
							Operator: token.Token{
								Type:    token.Plus,
								Literal: "+",
							},
							LHS: ast.IdentifierExpression{
								Identifier: "a",
							},
							RHS: ast.IdentifierExpression{
								Identifier: "b",
							},
						},
						RHS: ast.IdentifierExpression{
							Identifier: "c",
						},
					},
				},
			},
		},
		{
			src: "let a = \"crookdc\";",
			expected: ast.LetStatement{
				Assignment: ast.InfixExpression{
					Operator: token.Token{
						Type:    token.Assign,
						Literal: "=",
					},
					LHS: ast.IdentifierExpression{
						Identifier: "a",
					},
					RHS: ast.StringExpression{
						String: "crookdc",
					},
				},
			},
		},
		{
			src: "let crookdc = a + b * c;",
			expected: ast.LetStatement{
				Assignment: ast.InfixExpression{
					Expression: ast.Expression{},
					Operator: token.Token{
						Type:    token.Assign,
						Literal: "=",
					},
					LHS: ast.IdentifierExpression{
						Identifier: "crookdc",
					},
					RHS: ast.InfixExpression{
						Operator: token.Token{
							Type:    token.Plus,
							Literal: "+",
						},
						LHS: ast.IdentifierExpression{
							Identifier: "a",
						},
						RHS: ast.InfixExpression{
							Operator: token.Token{
								Type:    token.Asterisk,
								Literal: "*",
							},
							LHS: ast.IdentifierExpression{
								Identifier: "b",
							},
							RHS: ast.IdentifierExpression{
								Identifier: "c",
							},
						},
					},
				},
			},
		},
		{
			src: "if a < b && b > a;",
			expected: ast.ExpressionStatement{
				Expression: ast.PrefixExpression{
					Operator: token.Token{
						Type:    token.If,
						Literal: "if",
					},
					RHS: ast.InfixExpression{
						Operator: token.Token{
							Type:    token.And,
							Literal: "&&",
						},
						LHS: ast.InfixExpression{
							Operator: token.Token{
								Type:    token.LessThan,
								Literal: "<",
							},
							LHS: ast.IdentifierExpression{
								Identifier: "a",
							},
							RHS: ast.IdentifierExpression{
								Identifier: "b",
							},
						},
						RHS: ast.InfixExpression{
							Operator: token.Token{
								Type:    token.GreaterThan,
								Literal: ">",
							},
							LHS: ast.IdentifierExpression{
								Identifier: "b",
							},
							RHS: ast.IdentifierExpression{
								Identifier: "a",
							},
						},
					},
				},
			},
		},
		{
			src: "if true || (false);",
			expected: ast.ExpressionStatement{
				Expression: ast.PrefixExpression{
					Operator: token.Token{
						Type:    token.If,
						Literal: "if",
					},
					RHS: ast.InfixExpression{
						Operator: token.Token{
							Type:    token.Or,
							Literal: "||",
						},
						LHS: ast.BooleanExpression{
							Boolean: true,
						},
						RHS: ast.BooleanExpression{
							Boolean: false,
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
			assert.Nil(t, err)
			assert.Equal(t, test.expected, n)
		})
	}
}

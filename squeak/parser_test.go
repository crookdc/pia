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
				Identifier: "a",
				Value: ast.InfixExpression{

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
			src: "let a = \"crookdc\";",
			expected: ast.LetStatement{
				Identifier: "a",
				Value: ast.StringExpression{
					String: "crookdc",
				},
			},
		},
		{
			src: "let crookdc = a + b * c;",
			expected: ast.LetStatement{
				Identifier: "crookdc",
				Value: ast.InfixExpression{
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
			src: "if (a < b && b > a) a;",
			expected: ast.IfStatement{
				Condition: ast.InfixExpression{
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
				Consequence: ast.ExpressionStatement{
					Expression: ast.IdentifierExpression{
						Expression: ast.Expression{},
						Identifier: "a",
					},
				},
				Alternative: nil,
			},
		},
		{
			src: "if (true || (false)) a - b;",
			expected: ast.IfStatement{
				Condition: ast.InfixExpression{
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
				Consequence: ast.ExpressionStatement{
					Statement: ast.Statement{},
					Expression: ast.InfixExpression{
						Operator: token.Token{
							Type:    token.Minus,
							Literal: "-",
						},
						LHS: ast.IdentifierExpression{
							Identifier: "a",
						},
						RHS: ast.IdentifierExpression{
							Identifier: "b",
						},
					},
				},
				Alternative: nil,
			},
		},
		{
			src: "if (true) a; else if (false) b; else c;",
			expected: ast.IfStatement{
				Condition: ast.BooleanExpression{
					Boolean: true,
				},
				Consequence: ast.ExpressionStatement{
					Expression: ast.IdentifierExpression{
						Identifier: "a",
					},
				},
				Alternative: &ast.IfStatement{
					Condition: ast.BooleanExpression{
						Boolean: false,
					},
					Consequence: ast.ExpressionStatement{
						Expression: ast.IdentifierExpression{
							Identifier: "b",
						},
					},
					Alternative: &ast.IfStatement{
						Condition: ast.BooleanExpression{
							Boolean: true,
						},
						Consequence: ast.ExpressionStatement{
							Expression: ast.IdentifierExpression{
								Identifier: "c",
							},
						},
						Alternative: nil,
					},
				},
			},
		},
		{
			src: "let a = -a + a;",
			expected: ast.LetStatement{
				Identifier: "a",
				Value: ast.InfixExpression{
					Operator: token.Token{
						Type:    token.Plus,
						Literal: "+",
					},
					LHS: ast.PrefixExpression{
						Operator: token.Token{
							Type:    token.Minus,
							Literal: "-",
						},
						RHS: ast.IdentifierExpression{
							Identifier: "a",
						},
					},
					RHS: ast.IdentifierExpression{
						Identifier: "a",
					},
				},
			},
		},
		{
			src: "if (1 == 1) return; else return false;",
			expected: ast.IfStatement{
				Condition: ast.InfixExpression{
					Operator: token.Token{
						Type:    token.Equals,
						Literal: "==",
					},
					LHS: ast.IntegerExpression{
						Integer: 1,
					},
					RHS: ast.IntegerExpression{
						Integer: 1,
					},
				},
				Consequence: ast.ReturnStatement{},
				Alternative: &ast.IfStatement{
					Condition: ast.BooleanExpression{
						Boolean: true,
					},
					Consequence: ast.ReturnStatement{
						Expression: ast.BooleanExpression{
							Boolean: false,
						},
					},
					Alternative: nil,
				},
			},
		},
		{
			src: "let b64 = import \"core/base64\";",
			expected: ast.LetStatement{
				Identifier: "b64",
				Value: ast.PrefixExpression{
					Operator: token.Token{
						Type:    token.Import,
						Literal: "import",
					},
					RHS: ast.StringExpression{
						String: "core/base64",
					},
				},
			},
		},
		{
			src: "let a = !b;",
			expected: ast.LetStatement{
				Identifier: "a",
				Value: ast.PrefixExpression{
					Operator: token.Token{
						Type:    token.Bang,
						Literal: "!",
					},
					RHS: ast.IdentifierExpression{
						Identifier: "b",
					},
				},
			},
		},
		{
			src: "{ let a = 0; }",
			expected: ast.BlockStatement{
				Statements: []ast.StatementNode{
					ast.LetStatement{
						Identifier: "a",
						Value: ast.IntegerExpression{
							Integer: 0,
						},
					},
				},
			},
		},
		{
			src: "if (a) { a; }",
			expected: ast.IfStatement{
				Condition: ast.IdentifierExpression{
					Identifier: "a",
				},
				Consequence: ast.BlockStatement{
					Statements: []ast.StatementNode{
						ast.ExpressionStatement{
							Expression: ast.IdentifierExpression{
								Identifier: "a",
							},
						},
					},
				},
			},
		},
		{
			src: "let crookdc;",
			expected: ast.LetStatement{
				Identifier: "crookdc",
			},
		},
		{
			src: "a.b;",
			expected: ast.ExpressionStatement{
				Expression: ast.InfixExpression{
					Operator: token.Token{
						Type:    token.FullStop,
						Literal: ".",
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

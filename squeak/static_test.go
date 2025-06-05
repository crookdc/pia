package squeak

import (
	"github.com/crookdc/pia/squeak/ast"
	"github.com/crookdc/pia/squeak/token"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestResolver_Resolve(t *testing.T) {
	t.Run("resolution of global variable", func(t *testing.T) {
		declaration := ast.Declaration{
			Name: token.Token{
				Type:   token.Identifier,
				Lexeme: "name",
			},
			Initializer: ast.StringLiteral{
				String: "crookdc",
			},
		}
		r := resolver{
			stack: struct {
				slice []map[string]struct{}
				sp    int
			}{
				slice: make([]map[string]struct{}, 0),
				sp:    -1,
			},
			locals: make(map[string]int),
		}
		err := r.Resolve([]ast.StatementNode{declaration})
		assert.Nil(t, err)
		assert.Equal(t, 0, len(r.locals))
		assert.Equal(t, -1, r.stack.sp)
	})

	t.Run("simple resolution of scoped variable", func(t *testing.T) {
		declaration := ast.Block{
			Body: []ast.StatementNode{
				ast.Declaration{
					Name: token.Token{
						Type:   token.Identifier,
						Lexeme: "name",
					},
					Initializer: ast.StringLiteral{
						String: "crookdc",
					},
				},
				ast.ExpressionStatement{
					Expression: ast.Variable{
						ID: "1",
						Name: token.Token{
							Type:   token.Identifier,
							Lexeme: "name",
						},
					},
				},
			},
		}
		r := resolver{
			stack: struct {
				slice []map[string]struct{}
				sp    int
			}{
				slice: make([]map[string]struct{}, 0),
				sp:    -1,
			},
			locals: make(map[string]int),
		}
		err := r.Resolve([]ast.StatementNode{declaration})
		assert.Nil(t, err)
		assert.Equal(t, 1, len(r.locals))
		assert.Equal(t, 0, r.locals["id-0"])
		assert.Equal(t, -1, r.stack.sp)
	})

	t.Run("resolution through nested scopes", func(t *testing.T) {
		declaration := ast.Block{
			Body: []ast.StatementNode{
				ast.Declaration{
					Name: token.Token{
						Type:   token.Identifier,
						Lexeme: "name",
					},
					Initializer: ast.StringLiteral{
						String: "crookdc",
					},
				},
				ast.Function{
					Name: token.Token{
						Type:   token.Identifier,
						Lexeme: "print_me",
					},
					Params: []token.Token{
						{
							Type:   token.Identifier,
							Lexeme: "n",
						},
					},
					Body: ast.Block{
						Body: []ast.StatementNode{
							ast.ExpressionStatement{
								Expression: ast.Variable{
									ID: "1",
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
		}
		r := resolver{
			stack: struct {
				slice []map[string]struct{}
				sp    int
			}{
				slice: make([]map[string]struct{}, 0),
				sp:    -1,
			},
			locals: make(map[string]int),
		}
		err := r.Resolve([]ast.StatementNode{declaration})
		assert.Nil(t, err)
		assert.Equal(t, 1, len(r.locals))
		assert.Equal(t, 1, r.locals["1"])
		assert.Equal(t, -1, r.stack.sp)
	})
}

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
					ID: "1",
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
					ID: "1",
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
			src: "export \"some string\" as string;",
			expected: ast.Export{
				Name: token.Token{
					Type:   token.Identifier,
					Lexeme: "string",
				},
				Value: ast.StringLiteral{
					String: "some string",
				},
			},
		},
		{
			src: "export my_var as alias;",
			expected: ast.Export{
				Name: token.Token{
					Type:   token.Identifier,
					Lexeme: "alias",
				},
				Value: ast.Variable{
					ID: "1",
					Name: token.Token{
						Type:   token.Identifier,
						Lexeme: "my_var",
					},
				},
			},
		},
		{
			src: "export my_var;",
			expected: ast.Export{
				Name: token.Token{
					Type:   token.Identifier,
					Lexeme: "my_var",
				},
				Value: ast.Variable{
					ID: "1",
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
									ID: "1",
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
											ID: "2",
											Name: token.Token{
												Type:   token.Identifier,
												Lexeme: "a",
											},
										},
										RHS: ast.Variable{
											ID: "3",
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
						ID: "1",
						Name: token.Token{
							Type:   token.Identifier,
							Lexeme: "a",
						},
					},
					RHS: ast.Variable{
						ID: "2",
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
							ID: "1",
							Name: token.Token{
								Type:   token.Identifier,
								Lexeme: "a",
							},
						},
						RHS: ast.Variable{
							ID: "2",
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
						ID: "1",
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
						ID: "1",
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
							ID: "2",
							Name: token.Token{
								Type:   token.Identifier,
								Lexeme: "b",
							},
						},
						RHS: ast.Variable{
							ID: "3",
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
								ID: "1",
								Name: token.Token{
									Type:   token.Identifier,
									Lexeme: "a",
								},
							},
							RHS: ast.Variable{
								ID: "2",
								Name: token.Token{
									Type:   token.Identifier,
									Lexeme: "b",
								},
							},
						},
					},
					RHS: ast.Variable{
						ID: "3",
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
						ID: "1",
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
			src: "index[4];",
			expected: ast.ExpressionStatement{
				Expression: ast.Call{
					Callee: ast.Variable{
						ID: "1",
						Name: token.Token{
							Type:   token.Identifier,
							Lexeme: "index",
						},
					},
					Operator: token.Token{
						Type:   token.LeftBracket,
						Lexeme: "[",
					},
					Args: []ast.ExpressionNode{
						ast.IntegerLiteral{
							Integer: 4,
						},
					},
				},
			},
		},
		{
			src: "indexed[add(a, b)];",
			expected: ast.ExpressionStatement{
				Expression: ast.Call{
					Callee: ast.Variable{
						ID: "1",
						Name: token.Token{
							Type:   token.Identifier,
							Lexeme: "indexed",
						},
					},
					Operator: token.Token{
						Type:   token.LeftBracket,
						Lexeme: "[",
					},
					Args: []ast.ExpressionNode{
						ast.Call{
							Callee: ast.Variable{
								ID: "2",
								Name: token.Token{
									Type:   token.Identifier,
									Lexeme: "add",
								},
							},
							Operator: token.Token{
								Type:   token.LeftParenthesis,
								Lexeme: "(",
							},
							Args: []ast.ExpressionNode{
								ast.Variable{
									ID: "3",
									Name: token.Token{
										Type:   token.Identifier,
										Lexeme: "a",
									},
								},
								ast.Variable{
									ID: "4",
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
		{
			src: "indexed[12;",
			err: SyntaxError{Line: 1},
		},
		{
			src: "var list = [1, 2, 3, true, false, \"crookdc\"];",
			expected: ast.Declaration{
				Name: token.Token{
					Type:   token.Identifier,
					Lexeme: "list",
				},
				Initializer: ast.ListLiteral{
					Items: []ast.ExpressionNode{
						ast.IntegerLiteral{Integer: 1},
						ast.IntegerLiteral{Integer: 2},
						ast.IntegerLiteral{Integer: 3},
						ast.BooleanLiteral{Boolean: true},
						ast.BooleanLiteral{Boolean: false},
						ast.StringLiteral{String: "crookdc"},
					},
				},
			},
		},
		{
			src: "var list = [1 + 5, 9];",
			expected: ast.Declaration{
				Name: token.Token{
					Type:   token.Identifier,
					Lexeme: "list",
				},
				Initializer: ast.ListLiteral{
					Items: []ast.ExpressionNode{
						ast.Infix{
							Operator: token.Token{
								Type:   token.Plus,
								Lexeme: "+",
							},
							LHS: ast.IntegerLiteral{
								Integer: 1,
							},
							RHS: ast.IntegerLiteral{
								Integer: 5,
							},
						},
						ast.IntegerLiteral{
							Integer: 9,
						},
					},
				},
			},
		},
		{
			src: "var list = [a];",
			expected: ast.Declaration{
				Name: token.Token{
					Type:   token.Identifier,
					Lexeme: "list",
				},
				Initializer: ast.ListLiteral{
					Items: []ast.ExpressionNode{
						ast.Variable{
							ID: "1",
							Name: token.Token{
								Type:   token.Identifier,
								Lexeme: "a",
							},
						},
					},
				},
			},
		},
		{
			src: "var list = [];",
			expected: ast.Declaration{
				Name: token.Token{
					Type:   token.Identifier,
					Lexeme: "list",
				},
				Initializer: ast.ListLiteral{
					Items: make([]ast.ExpressionNode, 0),
				},
			},
		},
		{
			src: "run(5 + 1002, n);",
			expected: ast.ExpressionStatement{
				Expression: ast.Call{
					Callee: ast.Variable{
						ID: "1",
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
							ID: "2",
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
								ID: "1",
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
								ID: "2",
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
							ID: "3",
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
						ID: "1",
						Name: token.Token{
							Type:   token.Identifier,
							Lexeme: "a",
						},
					},
					RHS: ast.Variable{
						ID: "2",
						Name: token.Token{
							Type:   token.Identifier,
							Lexeme: "b",
						},
					},
				},
				Then: ast.ExpressionStatement{
					Expression: ast.Call{
						Callee: ast.Variable{
							ID: "3",
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
								ID: "4",
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
							ID: "5",
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
								ID: "6",
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
						ID: "1",
						Name: token.Token{
							Type:   token.Identifier,
							Lexeme: "a",
						},
					},
					RHS: ast.Variable{
						ID: "2",
						Name: token.Token{
							Type:   token.Identifier,
							Lexeme: "b",
						},
					},
				},
				Then: ast.ExpressionStatement{
					Expression: ast.Call{
						Callee: ast.Variable{
							ID: "3",
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
								ID: "4",
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
					ID: "1",
					Name: token.Token{
						Type:   token.Identifier,
						Lexeme: "a",
					},
				},
				Then: ast.If{
					Condition: ast.Variable{
						ID: "2",
						Name: token.Token{
							Type:   token.Identifier,
							Lexeme: "b",
						},
					},
					Then: ast.ExpressionStatement{
						Expression: ast.Call{
							Callee: ast.Variable{
								ID: "3",
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
									ID: "4",
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
								ID: "5",
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
									ID: "6",
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
			src: "{ var a; a + b; a = 2.; }",
			expected: ast.Block{
				Body: []ast.StatementNode{
					ast.Declaration{
						Name: token.Token{
							Type:   token.Identifier,
							Lexeme: "a",
						},
					},
					ast.ExpressionStatement{
						Expression: ast.Infix{
							Operator: token.Token{
								Type:   token.Plus,
								Lexeme: "+",
							},
							LHS: ast.Variable{
								Level: 0,
								Name: token.Token{
									Type:   token.Identifier,
									Lexeme: "a",
								},
							},
							RHS: ast.Variable{
								Level: 0,
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
							Level: 0,
							Name: token.Token{
								Type:   token.Identifier,
								Lexeme: "a",
							},
						},
						RHS: ast.Variable{
							Level: 0,
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
							Level: 0,
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
									Level: 1,
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
										Level: 1,
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
						Level: 0,
						Name: token.Token{
							Type:   token.Identifier,
							Lexeme: "a",
						},
					},
					RHS: ast.Variable{
						Level: 0,
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
									Level: 1,
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
										Level: 1,
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
									Level: 1,
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
										Level: 1,
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
						Level: 0,
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
							Level: 0,
							Name: token.Token{
								Type:   token.Identifier,
								Lexeme: "b",
							},
						},
					},
					RHS: ast.Variable{
						Level: 0,
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
					Level: 0,
					Name: token.Token{
						Type:   token.Identifier,
						Lexeme: "a",
					},
				},
				RHS: ast.Variable{
					Level: 0,
					Name: token.Token{
						Type:   token.Identifier,
						Lexeme: "b",
					},
				},
			},
		}, n)
	})
}

package squeak

import (
	"errors"
	"fmt"
	"github.com/crookdc/pia/squeak/internal/ast"
	"github.com/crookdc/pia/squeak/internal/token"
	"io"
	"strconv"
)

type SyntaxError struct {
	Line int
}

func (s SyntaxError) Error() string {
	return fmt.Sprintf("syntax error on line %d", s.Line)
}

func NewParser(lx *PeekingLexer) *Parser {
	return &Parser{lx: lx}
}

// Parser builds an abstract syntax tree from the tokens yielded by a Lexer.
type Parser struct {
	lx *PeekingLexer
}

// Next constructs and returns the next node in the abstract syntax tree for the underlying Lexer.
func (ps *Parser) Next() (stmt ast.StatementNode, err error) {
	defer func() {
		if err != nil {
			// If an error occurred for any reason during the parsing of the current statement then the parser should at
			// least try to fast-forward to the next statement. This counteracts cascading syntax errors that would be
			// fine if it wasn't for the initial error that triggered a chain reaction. However, it is possible that
			// another is encountered as the current statement is cleared, hence the call to errors.Join().
			err = errors.Join(err, ps.clear())
		}
	}()
	pk, err := ps.lx.Peek()
	if err != nil {
		return nil, err
	}
	switch pk.Type {
	case token.EOF:
		return nil, io.EOF
	default:
		return ps.declaration()
	}
}

func (ps *Parser) declaration() (ast.StatementNode, error) {
	pk, err := ps.lx.Peek()
	if err != nil {
		return nil, err
	}
	switch pk.Type {
	case token.Var:
		return ps.variable()
	default:
		return ps.statement()
	}
}

func (ps *Parser) variable() (ast.Declaration, error) {
	if _, err := ps.expect(token.Var); err != nil {
		return ast.Declaration{}, err
	}
	name, err := ps.expect(token.Identifier)
	if err != nil {
		return ast.Declaration{}, err
	}
	stmt := ast.Declaration{
		Name:        name,
		Initializer: nil,
	}
	pk, err := ps.lx.Peek()
	if err != nil {
		return ast.Declaration{}, err
	}
	if pk.Type == token.Semicolon {
		ps.lx.Discard()
		return stmt, nil
	}
	if _, err := ps.expect(token.Assign); err != nil {
		return ast.Declaration{}, err
	}
	init, err := ps.equality()
	if err != nil {
		return ast.Declaration{}, err
	}
	if _, err := ps.expect(token.Semicolon); err != nil {
		return ast.Declaration{}, err
	}
	stmt.Initializer = init
	return stmt, nil
}

func (ps *Parser) statement() (ast.StatementNode, error) {
	pk, err := ps.lx.Peek()
	if err != nil {
		return nil, err
	}
	switch pk.Type {
	case token.LeftBrace:
		return ps.block()
	case token.If:
		return ps.ifs()
	case token.While:
		return ps.while()
	case token.For:
		return ps.fors()
	case token.Semicolon:
		ps.lx.Discard()
		return ast.Noop{}, nil
	default:
		return ps.expression()
	}
}

func (ps *Parser) while() (ast.StatementNode, error) {
	if _, err := ps.expect(token.While); err != nil {
		return nil, err
	}
	cnd, err := ps.logical()
	if err != nil {
		return nil, err
	}
	body, err := ps.block()
	if err != nil {
		return nil, err
	}
	return ast.While{
		Condition: cnd,
		Body:      body,
	}, nil
}

func (ps *Parser) fors() (ast.StatementNode, error) {
	if _, err := ps.expect(token.For); err != nil {
		return nil, err
	}
	init, err := ps.declaration()
	if err != nil {
		return nil, err
	}
	cnd, err := ps.logical()
	if err != nil {
		return nil, err
	}
	if _, err := ps.expect(token.Semicolon); err != nil {
		return nil, err
	}
	pk, err := ps.lx.Peek()
	if err != nil {
		return nil, err
	}
	var inc ast.ExpressionNode
	switch pk.Type {
	case token.LeftBrace:
		inc = nil
	default:
		inc, err = ps.assignment()
		if err != nil {
			return nil, err
		}
	}
	body, err := ps.block()
	if err != nil {
		return nil, err
	}
	loop := ast.While{
		Condition: cnd,
		Body:      body,
	}
	if inc != nil {
		loop.Body.Body = append(loop.Body.Body, ast.ExpressionStatement{
			Expression: inc,
		})
	}
	return ast.Block{
		Body: []ast.StatementNode{init, loop},
	}, nil
}

func (ps *Parser) ifs() (ast.If, error) {
	if _, err := ps.expect(token.If); err != nil {
		return ast.If{}, err
	}
	cnd, err := ps.logical()
	if err != nil {
		return ast.If{}, err
	}
	then, err := ps.statement()
	if err != nil {
		return ast.If{}, err
	}
	pk, err := ps.lx.Peek()
	if err != nil {
		return ast.If{}, err
	}
	st := ast.If{
		Condition: cnd,
		Then:      then,
		Else:      nil,
	}
	if pk.Type == token.Else {
		ps.lx.Discard()
		otherwise, err := ps.statement()
		if err != nil {
			return ast.If{}, err
		}
		st.Else = otherwise
	}
	return st, nil
}

func (ps *Parser) block() (ast.Block, error) {
	if _, err := ps.expect(token.LeftBrace); err != nil {
		return ast.Block{}, err
	}
	pk, err := ps.lx.Peek()
	if err != nil {
		return ast.Block{}, err
	}
	body := make([]ast.StatementNode, 0)
	for {
		switch pk.Type {
		case token.RightBrace, token.EOF:
			ps.lx.Discard()
			return ast.Block{
				Body: body,
			}, nil
		default:
			st, err := ps.statement()
			if err != nil {
				return ast.Block{}, err
			}
			body = append(body, st)
			pk, err = ps.lx.Peek()
			if err != nil {
				return ast.Block{}, err
			}
		}
	}
}

func (ps *Parser) clear() error {
	look := true
	for look {
		nxt, err := ps.lx.Next()
		if err != nil {
			return err
		}
		switch nxt.Type {
		case token.EOF, token.Semicolon, token.RightBrace:
			look = false
		default:
		}
	}
	return nil
}

func (ps *Parser) expression() (ast.ExpressionStatement, error) {
	expr, err := ps.assignment()
	if err != nil {
		return ast.ExpressionStatement{}, err
	}
	if _, err := ps.expect(token.Semicolon); err != nil {
		return ast.ExpressionStatement{}, err
	}
	return ast.ExpressionStatement{
		Expression: expr,
	}, nil
}

func (ps *Parser) assignment() (ast.ExpressionNode, error) {
	expr, err := ps.logical()
	if err != nil {
		return nil, err
	}
	pk, err := ps.lx.Peek()
	if err != nil {
		return nil, err
	}
	if pk.Type != token.Assign {
		return expr, nil
	}
	ps.lx.Discard()
	switch expr := expr.(type) {
	case ast.Variable:
		val, err := ps.assignment()
		if err != nil {
			return nil, err
		}
		return ast.Assignment{
			Name:  expr.Name,
			Value: val,
		}, nil
	default:
		return nil, fmt.Errorf(
			"%w: invalid left hand side of assignment",
			SyntaxError{Line: ps.lx.Line()},
		)
	}
}

func (ps *Parser) logical() (ast.ExpressionNode, error) {
	lhs, err := ps.equality()
	if err != nil {
		return nil, err
	}
	for {
		pk, err := ps.lx.Peek()
		if err != nil {
			return nil, err
		}
		switch pk.Type {
		case token.And, token.Or:
			ps.lx.Discard()
			rhs, err := ps.equality()
			if err != nil {
				return nil, err
			}
			lhs = ast.Logical{
				Operator: pk,
				LHS:      lhs,
				RHS:      rhs,
			}
		default:
			return lhs, nil
		}
	}
}

func (ps *Parser) equality() (ast.ExpressionNode, error) {
	lhs, err := ps.comparison()
	if err != nil {
		return nil, err
	}
	for {
		pk, err := ps.lx.Peek()
		if err != nil {
			return nil, err
		}
		switch pk.Type {
		case token.Equals, token.NotEquals:
			ps.lx.Discard()
			rhs, err := ps.comparison()
			if err != nil {
				return nil, err
			}
			lhs = ast.Infix{
				Expression: ast.Expression{},
				Operator:   pk,
				LHS:        lhs,
				RHS:        rhs,
			}
		default:
			return lhs, nil
		}
	}
}

func (ps *Parser) comparison() (ast.ExpressionNode, error) {
	lhs, err := ps.term()
	if err != nil {
		return nil, err
	}
	for {
		pk, err := ps.lx.Peek()
		if err != nil {
			return nil, err
		}
		switch pk.Type {
		case token.Less, token.LessEqual, token.Greater, token.GreaterEqual:
			ps.lx.Discard()
			rhs, err := ps.term()
			if err != nil {
				return nil, err
			}
			lhs = ast.Infix{
				Operator: pk,
				LHS:      lhs,
				RHS:      rhs,
			}
		default:
			return lhs, nil
		}
	}
}

func (ps *Parser) term() (ast.ExpressionNode, error) {
	lhs, err := ps.factor()
	if err != nil {
		return nil, err
	}
	for {
		pk, err := ps.lx.Peek()
		if err != nil {
			return nil, err
		}
		switch pk.Type {
		case token.Minus, token.Plus:
			ps.lx.Discard()
			rhs, err := ps.factor()
			if err != nil {
				return nil, err
			}
			lhs = ast.Infix{
				Operator: pk,
				LHS:      lhs,
				RHS:      rhs,
			}
		default:
			return lhs, nil
		}
	}
}

func (ps *Parser) factor() (ast.ExpressionNode, error) {
	lhs, err := ps.prefix()
	if err != nil {
		return nil, err
	}
	for {
		pk, err := ps.lx.Peek()
		if err != nil {
			return nil, err
		}
		switch pk.Type {
		case token.Asterisk, token.Slash:
			ps.lx.Discard()
			rhs, err := ps.prefix()
			if err != nil {
				return nil, err
			}
			lhs = ast.Infix{
				Operator: pk,
				LHS:      lhs,
				RHS:      rhs,
			}
		default:
			return lhs, nil
		}
	}
}

func (ps *Parser) prefix() (ast.ExpressionNode, error) {
	pk, err := ps.lx.Peek()
	if err != nil {
		return nil, err
	}
	switch pk.Type {
	case token.Bang, token.Minus:
		ps.lx.Discard()
		expr, err := ps.prefix()
		if err != nil {
			return nil, err
		}
		return ast.Prefix{
			Operator: pk,
			Target:   expr,
		}, nil
	default:
		return ps.call()
	}
}

func (ps *Parser) call() (ast.ExpressionNode, error) {
	expr, err := ps.primary()
	if err != nil {
		return nil, err
	}
	pk, err := ps.lx.Peek()
	if err != nil {
		return nil, err
	}
	for pk.Type == token.LeftParenthesis {
		args, err := ps.list(token.LeftParenthesis, token.RightParenthesis)
		if err != nil {
			return nil, err
		}
		expr = ast.Call{
			Callee:   expr,
			Operator: pk,
			Args:     args,
		}
		pk, err = ps.lx.Peek()
		if err != nil {
			return nil, err
		}
	}
	return expr, nil
}

func (ps *Parser) list(start, end token.Type) (exps []ast.ExpressionNode, err error) {
	if _, err := ps.expect(start); err != nil {
		return nil, err
	}
	for {
		pk, err := ps.lx.Peek()
		if err != nil {
			return nil, err
		}
		switch pk.Type {
		case end:
			ps.lx.Discard()
			return exps, nil
		case token.Comma:
			ps.lx.Discard()
		default:
			expr, err := ps.equality()
			if err != nil {
				return nil, err
			}
			exps = append(exps, expr)
		}
	}
}

func (ps *Parser) primary() (ast.ExpressionNode, error) {
	nxt, err := ps.lx.Next()
	if err != nil {
		return nil, err
	}
	switch nxt.Type {
	case token.Identifier:
		return ast.Variable{Name: nxt}, nil
	case token.String:
		return ast.StringLiteral{String: nxt.Lexeme}, nil
	case token.Integer:
		i, err := strconv.Atoi(nxt.Lexeme)
		if err != nil {
			return nil, fmt.Errorf(
				"%w: invalid integer literal: %s",
				SyntaxError{ps.lx.Line()},
				nxt.Lexeme,
			)
		}
		return ast.IntegerLiteral{Integer: i}, nil
	case token.Float:
		f, err := strconv.ParseFloat(nxt.Lexeme, 64)
		if err != nil {
			return nil, fmt.Errorf(
				"%w: invalid float literal: %s",
				SyntaxError{ps.lx.Line()},
				nxt.Lexeme,
			)
		}
		return ast.FloatLiteral{Float: f}, nil
	case token.Boolean:
		b, err := strconv.ParseBool(nxt.Lexeme)
		if err != nil {
			return nil, fmt.Errorf(
				"%w: invalid boolean literal: %s",
				SyntaxError{ps.lx.Line()},
				nxt.Lexeme,
			)
		}
		return ast.BooleanLiteral{Boolean: b}, nil
	case token.LeftParenthesis:
		expr, err := ps.equality()
		if err != nil {
			return nil, err
		}
		if _, err := ps.expect(token.RightParenthesis); err != nil {
			return nil, err
		}
		return ast.Grouping{
			Group: expr,
		}, nil
	case token.Nil:
		return ast.NilLiteral{}, nil
	default:
		return nil, fmt.Errorf(
			"%w: unexpected token: %s",
			SyntaxError{Line: ps.lx.Line()},
			nxt.Lexeme,
		)
	}
}

func (ps *Parser) expect(v token.Type) (token.Token, error) {
	t, err := ps.lx.Next()
	if err != nil {
		return token.Token{}, err
	}
	if t.Type != v {
		return token.Token{}, fmt.Errorf(
			"%w: unexpected token: %s",
			SyntaxError{ps.lx.Line()},
			t.Lexeme,
		)
	}
	return t, nil
}

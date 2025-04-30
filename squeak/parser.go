package squeak

import (
	"errors"
	"fmt"
	"github.com/crookdc/pia/squeak/internal/ast"
	"github.com/crookdc/pia/squeak/internal/token"
	"strconv"
)

var (
	ErrUnrecognizedToken = errors.New("unrecognized token")

	Precedences = map[token.Type]int{
		token.RightParenthesis: PrecedenceLowest,
		token.Assign:           PrecedenceAssignment,
		token.And:              PrecedenceAssignment,
		token.Or:               PrecedenceAssignment,
		token.Equals:           PrecedenceComparison,
		token.NotEquals:        PrecedenceComparison,
		token.LessThan:         PrecedenceComparison,
		token.GreaterThan:      PrecedenceComparison,
		token.Plus:             PrecedenceSum,
		token.Minus:            PrecedenceSum,
		token.Asterisk:         PrecedenceProduct,
		token.Slash:            PrecedenceProduct,
		token.If:               PrecedencePrefix,
	}
)

const (
	_ int = iota
	PrecedenceLowest
	PrecedenceAssignment
	PrecedenceComparison
	PrecedenceSum
	PrecedenceProduct
	PrecedencePrefix
)

// Parser builds an abstract syntax tree from the tokens yielded by a Lexer.
type Parser struct {
	lx *PeekingLexer
}

// Next constructs and returns the next node in the abstract syntax tree for the targeted Lexer.
func (ps *Parser) Next() (ast.Node, error) {
	return ps.statement()
}

func (ps *Parser) statement() (s ast.StatementNode, err error) {
	peek, err := ps.lx.Peek()
	if err != nil {
		return nil, err
	}
	switch peek.Type {
	case token.Let:
		return ps.let()
	case token.If:
		return ps.ifs()
	case token.Return:
		return ps.ret()
	default:
		e, err := ps.expression(PrecedenceLowest)
		if err != nil {
			return nil, err
		}
		if _, err := ps.expect(token.Semicolon); err != nil {
			return nil, err
		}
		return ast.ExpressionStatement{
			Expression: e,
		}, nil
	}
}

func (ps *Parser) let() (ast.LetStatement, error) {
	if _, err := ps.expect(token.Let); err != nil {
		return ast.LetStatement{}, err
	}
	asn, err := ps.expression(PrecedenceLowest)
	if err != nil {
		return ast.LetStatement{}, err
	}
	if _, err := ps.expect(token.Semicolon); err != nil {
		return ast.LetStatement{}, err
	}
	return ast.LetStatement{
		Assignment: asn,
	}, nil
}

func (ps *Parser) ifs() (ast.IfStatement, error) {
	if _, err := ps.expect(token.If); err != nil {
		return ast.IfStatement{}, err
	}
	cnd, err := ps.expression(PrecedenceLowest)
	if err != nil {
		return ast.IfStatement{}, err
	}
	csq, err := ps.statement()
	if err != nil {
		return ast.IfStatement{}, err
	}
	nxt, err := ps.lx.Peek()
	if err != nil {
		return ast.IfStatement{}, err
	}
	var alt *ast.IfStatement
	if nxt.Type == token.Else {
		alt, err = ps.alt()
		if err != nil {
			return ast.IfStatement{}, err
		}
	}
	return ast.IfStatement{
		Condition:   cnd,
		Consequence: csq,
		Alternative: alt,
	}, nil
}

func (ps *Parser) alt() (*ast.IfStatement, error) {
	if _, err := ps.expect(token.Else); err != nil {
		return nil, err
	}
	peek, err := ps.lx.Peek()
	if err != nil {
		return nil, err
	}
	switch peek.Type {
	case token.If:
		// An else-if branch can be processed like any other if-statement. This recursive behavior causes the parser to
		// build a chain of ast.IfStatement.
		alt, err := ps.ifs()
		if err != nil {
			return nil, err
		}
		return &alt, nil
	default:
		// A regular else block can be parsed as an alternative if-statement that always resolves to true.
		csq, err := ps.statement()
		if err != nil {
			return nil, err
		}
		return &ast.IfStatement{
			Condition:   ast.BooleanExpression{Boolean: true},
			Consequence: csq,
		}, nil
	}
}

func (ps *Parser) ret() (stmt ast.ReturnStatement, err error) {
	defer func() {
		if err != nil {
			return
		}
		// This deferred function might be refactored away for the sake of clearer control flow. The reason for it was
		// that there are two possible flavors of the return statement; one with an expression and one without, and both
		// should assert the existence of a semicolon at the end of parsing.
		_, err = ps.expect(token.Semicolon)
	}()
	if _, err := ps.expect(token.Return); err != nil {
		return ast.ReturnStatement{}, nil
	}
	pk, err := ps.lx.Peek()
	if err != nil {
		return ast.ReturnStatement{}, err
	}
	if pk.Type == token.Semicolon {
		return ast.ReturnStatement{}, nil

	}
	e, err := ps.expression(PrecedenceLowest)
	if err != nil {
		return ast.ReturnStatement{}, err
	}
	return ast.ReturnStatement{
		Expression: e,
	}, nil
}

func (ps *Parser) expression(precedence int) (ast.ExpressionNode, error) {
	t, err := ps.lx.Next()
	if err != nil {
		return nil, err
	}
	var e ast.ExpressionNode
	switch t.Type {
	case token.Identifier:
		e = ast.IdentifierExpression{
			Identifier: t.Literal,
		}
	case token.Integer:
		e, err = ps.integer(t)
		if err != nil {
			return nil, err
		}
	case token.String:
		e = ast.StringExpression{
			String: t.Literal,
		}
	case token.Boolean:
		e = ast.BooleanExpression{
			Boolean: t.Literal == "true",
		}
	case token.Minus:
		e, err = ps.prefix(t)
		if err != nil {
			return nil, err
		}
	case token.Import:
		e, err = ps.prefix(t)
		if err != nil {
			return nil, err
		}
	case token.LeftParenthesis:
		e, err = ps.expression(PrecedenceLowest)
		if err != nil {
			return nil, err
		}
		if _, err := ps.expect(token.RightParenthesis); err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("%w: %s", ErrUnrecognizedToken, t.Literal)
	}
	var done bool
	for !done {
		peek, err := ps.lx.Peek()
		if err != nil {
			return nil, err
		}
		switch peek.Type {
		case token.Semicolon:
			done = true
		default:
			op := Precedences[peek.Type]
			if precedence >= op {
				done = true
				break
			}
			ps.lx.Discard()
			e, err = ps.infix(e, peek)
			if err != nil {
				return nil, err
			}
		}
	}
	return e, nil
}

func (ps *Parser) integer(t token.Token) (ast.IntegerExpression, error) {
	n, err := strconv.Atoi(t.Literal)
	if err != nil {
		return ast.IntegerExpression{}, err
	}
	return ast.IntegerExpression{
		Integer: n,
	}, nil
}

func (ps *Parser) infix(lhs ast.ExpressionNode, op token.Token) (ast.InfixExpression, error) {
	rhs, err := ps.expression(Precedences[op.Type])
	if err != nil {
		return ast.InfixExpression{}, err
	}
	return ast.InfixExpression{
		Operator: op,
		LHS:      lhs,
		RHS:      rhs,
	}, nil
}

func (ps *Parser) prefix(op token.Token) (ast.PrefixExpression, error) {
	rhs, err := ps.expression(PrecedencePrefix)
	if err != nil {
		return ast.PrefixExpression{}, err
	}
	return ast.PrefixExpression{
		Operator: op,
		RHS:      rhs,
	}, nil
}

func (ps *Parser) expect(v token.Type) (token.Token, error) {
	t, err := ps.lx.Next()
	if err != nil {
		return token.Token{}, err
	}
	if t.Type != v {
		return token.Token{}, ErrUnrecognizedToken
	}
	return t, nil
}

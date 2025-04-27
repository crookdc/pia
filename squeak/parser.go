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

func (ps *Parser) Next() (ast.Node, error) {
	peek, err := ps.lx.Peek()
	if err != nil {
		return nil, err
	}
	switch peek.Type {
	case token.Let:
		return ps.let()
	default:
		e, err := ps.expression(PrecedenceLowest)
		if err != nil {
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
	return ast.LetStatement{
		Assignment: asn,
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
	case token.If:
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
			rhs, err := ps.expression(op)
			if err != nil {
				return nil, err
			}
			e = ast.InfixExpression{
				Operator: peek,
				LHS:      e,
				RHS:      rhs,
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

func (ps *Parser) prefix(op token.Token) (ast.PrefixExpression, error) {
	rhs, err := ps.expression(PrecedenceLowest)
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

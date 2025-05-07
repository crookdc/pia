package squeak

import (
	"fmt"
	"io"
	"strconv"

	"github.com/crookdc/pia/squeak/internal/ast"
	"github.com/crookdc/pia/squeak/internal/token"
)

var (
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
		token.FullStop:         PrecedenceCall,
	}
)

type UnexpectedToken struct {
	Line  int
	Token token.Token
}

func (u UnexpectedToken) Error() string {
	return fmt.Sprintf("unexpected token '%s' on line %d", u.Token.Lexeme, u.Line)
}

const (
	_ int = iota
	PrecedenceLowest
	PrecedenceAssignment
	PrecedenceComparison
	PrecedenceSum
	PrecedenceProduct
	PrecedencePrefix
	PrecedenceCall
)

func NewParser(lx *PeekingLexer) *Parser {
	return &Parser{lx: lx}
}

// Parser builds an abstract syntax tree from the tokens yielded by a Lexer.
type Parser struct {
	lx *PeekingLexer
}

// Next constructs and returns the next node in the abstract syntax tree for the targeted Lexer.
func (ps *Parser) Next() (ast.StatementNode, error) {
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
	case token.LeftCurlyBrace:
		return ps.block()
	case token.EOF:
		return nil, io.EOF
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
	id, err := ps.expect(token.Identifier)
	if err != nil {
		return ast.LetStatement{}, err
	}
	pk, err := ps.lx.Peek()
	if err != nil {
		return ast.LetStatement{}, err
	}
	stmt := ast.LetStatement{
		Identifier: id.Lexeme,
	}
	// Let statements can include an initializer expression but does not have to.
	switch pk.Type {
	case token.Assign:
		if _, err := ps.expect(token.Assign); err != nil {
			return ast.LetStatement{}, err
		}
		val, err := ps.expression(PrecedenceLowest)
		if err != nil {
			return ast.LetStatement{}, err
		}
		stmt.Value = val
	case token.Semicolon:
	default:
		return ast.LetStatement{}, UnexpectedToken{
			Line:  ps.lx.Line(),
			Token: pk,
		}
	}

	if _, err := ps.expect(token.Semicolon); err != nil {
		return ast.LetStatement{}, err
	}
	return stmt, nil
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

func (ps *Parser) block() (ast.BlockStatement, error) {
	if _, err := ps.expect(token.LeftCurlyBrace); err != nil {
		return ast.BlockStatement{}, err
	}
	peek, err := ps.lx.Peek()
	if err != nil {
		return ast.BlockStatement{}, err
	}
	block := make([]ast.StatementNode, 0)
	for peek.Type != token.RightCurlyBrace {
		s, err := ps.statement()
		if err != nil {
			return ast.BlockStatement{}, err
		}
		block = append(block, s)
		peek, err = ps.lx.Peek()
		if err != nil {
			return ast.BlockStatement{}, err
		}
	}
	if _, err := ps.expect(token.RightCurlyBrace); err != nil {
		return ast.BlockStatement{}, err
	}
	return ast.BlockStatement{
		Statements: block,
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
			Identifier: t.Lexeme,
		}
	case token.Integer:
		e, err = ps.integer(t)
		if err != nil {
			return nil, err
		}
	case token.String:
		e = ast.StringExpression{
			String: t.Lexeme,
		}
	case token.Boolean:
		e = ast.BooleanExpression{
			Boolean: t.Lexeme == "true",
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
	case token.Bang:
		e, err = ps.prefix(t)
		if err != nil {
			return nil, err
		}
	case token.Function:
		params, err := ps.list(token.LeftParenthesis, token.RightParenthesis)
		if err != nil {
			return nil, err
		}
		body, err := ps.statement()
		if err != nil {
			return nil, err
		}
		e = ast.FunctionExpression{
			Parameters: params,
			Body:       body,
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
		return nil, UnexpectedToken{
			Line:  ps.lx.Line(),
			Token: t,
		}
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
		case token.LeftParenthesis:
			args, err := ps.list(token.LeftParenthesis, token.RightParenthesis)
			if err != nil {
				return nil, err
			}
			e = ast.CallExpression{
				Identifier: e,
				Arguments:  args,
			}
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

func (ps *Parser) list(prefix, suffix token.Type) ([]ast.ExpressionNode, error) {
	if _, err := ps.expect(prefix); err != nil {
		return nil, err
	}
	t, err := ps.lx.Peek()
	if err != nil {
		return nil, err
	}

	items := make([]ast.ExpressionNode, 0)
	for t.Type != suffix {
		e, err := ps.expression(PrecedenceLowest)
		if err != nil {
			return nil, err
		}
		items = append(items, e)

		t, err = ps.lx.Peek()
		if err != nil {
			return nil, err
		}
		switch t.Type {
		case token.Comma:
			ps.lx.Discard()
			t, err = ps.lx.Peek()
			if err != nil {
				return nil, err
			}
		case suffix:
			// Ignore
		default:
			return nil, UnexpectedToken{
				Line:  ps.lx.Line(),
				Token: t,
			}
		}
	}
	if _, err := ps.expect(suffix); err != nil {
		return nil, err
	}
	return items, nil
}

func (ps *Parser) integer(t token.Token) (ast.IntegerExpression, error) {
	n, err := strconv.Atoi(t.Lexeme)
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
		return token.Token{}, UnexpectedToken{
			Line:  ps.lx.Line(),
			Token: t,
		}
	}
	return t, nil
}

package token

import (
	"errors"
)

var (
	Nil               = Token{}
	ErrMissingLiteral = errors.New("missing token literal")
)

const (
	_ Type = iota
	EOF
	Illegal
	Identifier
	Integer
	String
	Boolean
	And
	Or
	If
	Else
	While
	LessThan
	GreaterThan
	Equals
	NotEquals
	Return
	Assign
	Bang
	Plus
	Minus
	Asterisk
	Slash
	Comma
	FullStop
	Semicolon
	LeftParenthesis
	RightParenthesis
	LeftCurlyBrace
	RightCurlyBrace
	LeftBracket
	RightBracket
	Function
	Let
	Import
)

var literals = map[Type]string{
	EOF:              "EOF",
	And:              "&&",
	Or:               "||",
	If:               "if",
	Else:             "else",
	While:            "while",
	LessThan:         "<",
	GreaterThan:      ">",
	Equals:           "==",
	NotEquals:        "!=",
	Return:           "return",
	Assign:           "=",
	Bang:             "!",
	Plus:             "+",
	Minus:            "-",
	Asterisk:         "*",
	Slash:            "/",
	Comma:            ",",
	FullStop:         ".",
	Semicolon:        ";",
	LeftParenthesis:  "(",
	RightParenthesis: ")",
	LeftCurlyBrace:   "{",
	RightCurlyBrace:  "}",
	LeftBracket:      "[",
	RightBracket:     "]",
	Function:         "func",
	Let:              "let",
	Import:           "import",
}

// Type identifies a token type such as brackets, parenthesis and keywords.
type Type int

// Token represents a source code token in the Squeak language. A token can be single characters such as parenthesis and
// brackets, but it can also be entire Squeak keywords.
type Token struct {
	Type
	// Literal contains the literal string representation of a token. For many token types this is static, like
	// semicolons and keywords, but for others it may vary such as for integer and string literals.
	Literal string
}

// Opt represents an optional transformer of token values which is applied to the Token upon construction with
// [token.New].
type Opt func(t *Token)

// Literal is an Opt implementation that allows the caller to set the [token.Token.Literal] value to a custom one.
func Literal(literal string) Opt {
	return func(t *Token) {
		t.Literal = literal
	}
}

// New constructs a new Token and returns a non-nil error if the resulting token does not contain a literal
// representation.
func New(t Type, opts ...Opt) (Token, error) {
	token := Token{
		Type:    t,
		Literal: literals[t],
	}
	for _, opt := range opts {
		opt(&token)
	}
	if token.Literal == "" {
		return Token{}, ErrMissingLiteral
	}
	return token, nil
}

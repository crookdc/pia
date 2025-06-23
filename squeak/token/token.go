package token

import (
	"errors"
	"fmt"
)

var (
	Null             = Token{}
	ErrMissingLexeme = errors.New("missing token lexeme")
)

const (
	_ Type = iota
	EOF
	Illegal

	Identifier
	Integer
	Float
	String
	Boolean

	And
	Or
	If
	Else
	While
	Return
	Break
	Continue
	Function
	Var
	Nil
	Import
	Export
	As
	Object

	Less
	LessEqual
	Greater
	GreaterEqual
	Equals
	NotEquals
	Assign
	Bang
	Plus
	Minus
	Asterisk
	Slash
	Comma
	Dot
	Semicolon
	Colon
	LeftParenthesis
	RightParenthesis
	LeftBrace
	RightBrace
	LeftBracket
	RightBracket
)

var (
	Closers = map[Type]Type{
		LeftParenthesis: RightParenthesis,
		LeftBrace:       RightBrace,
		LeftBracket:     RightBracket,
	}
	lexemes = map[Type]string{
		EOF:              "EOF",
		And:              "and",
		Or:               "or",
		If:               "if",
		Else:             "else",
		While:            "while",
		Return:           "return",
		Break:            "break",
		Continue:         "continue",
		Function:         "function",
		Var:              "var",
		Nil:              "nil",
		Import:           "import",
		As:               "as",
		Object:           "Object",
		Export:           "export",
		Less:             "<",
		LessEqual:        "<=",
		Greater:          ">",
		GreaterEqual:     ">=",
		Equals:           "==",
		NotEquals:        "!=",
		Assign:           "=",
		Bang:             "!",
		Plus:             "+",
		Minus:            "-",
		Asterisk:         "*",
		Slash:            "/",
		Comma:            ",",
		Dot:              ".",
		Semicolon:        ";",
		Colon:            ":",
		LeftParenthesis:  "(",
		RightParenthesis: ")",
		LeftBrace:        "{",
		RightBrace:       "}",
		LeftBracket:      "[",
		RightBracket:     "]",
	}
)

// Type identifies a token type such as brackets, parenthesis and keywords.
type Type int

// Token represents a source code token in the Squeak language. A token can be single characters such as parenthesis and
// brackets, but it can also be entire Squeak keywords.
type Token struct {
	Type
	// Lexeme contains the literal string representation of a token. For many token types this is static, like
	// semicolons and keywords, but for others it may vary such as for integer and string literals.
	Lexeme string
}

// Opt represents an optional transformer of token values which is applied to the Token upon construction with
// [token.New].
type Opt func(t *Token)

// Lexeme is an Opt implementation that allows the caller to set the [token.Token.Lexeme] value to a custom one.
func Lexeme(lexeme string) Opt {
	return func(t *Token) {
		t.Lexeme = lexeme
	}
}

// New constructs a new Token and returns a non-nil error if the resulting token does not contain a literal
// representation.
func New(t Type, opts ...Opt) (Token, error) {
	token := Token{
		Type:   t,
		Lexeme: lexemes[t],
	}
	for _, opt := range opts {
		opt(&token)
	}
	if token.Lexeme == "" {
		return Token{}, fmt.Errorf("%w: %d", ErrMissingLexeme, t)
	}
	return token, nil
}

package squeak

import (
	"errors"
	"github.com/crookdc/pia/squeak/internal/token"
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
)

func TestNewLexer(t *testing.T) {
	t.Run("given nil source reader", func(t *testing.T) {
		_, err := NewLexer(nil)
		assert.True(t, errors.Is(err, ErrInvalidSourceReader))
	})
	t.Run("given non-nil source reader", func(t *testing.T) {
		src := "let a = b;"
		_, err := NewLexer(strings.NewReader(src))
		assert.Nil(t, err)
	})
}

func TestLexer_Next(t *testing.T) {
	tests := []struct {
		src      string
		expected []token.Token
		bl       int
	}{
		{
			src: " let  = 512;",
			bl:  LexerBufferLength,
			expected: []token.Token{
				{
					Type:    token.Let,
					Literal: "let",
				},
				{
					Type:    token.Assign,
					Literal: "=",
				},
				{
					Type:    token.Integer,
					Literal: "512",
				},
				{
					Type:    token.Semicolon,
					Literal: ";",
				},
				{
					Type:    token.EOF,
					Literal: "EOF",
				},
			},
		},
		{
			src: "a+b/&&(}){*!=!.",
			bl:  LexerBufferLength,
			expected: []token.Token{
				{
					Type:    token.Identifier,
					Literal: "a",
				},
				{
					Type:    token.Plus,
					Literal: "+",
				},
				{
					Type:    token.Identifier,
					Literal: "b",
				},
				{
					Type:    token.Slash,
					Literal: "/",
				},
				{
					Type:    token.And,
					Literal: "&&",
				},
				{
					Type:    token.LeftParenthesis,
					Literal: "(",
				},
				{
					Type:    token.RightCurlyBrace,
					Literal: "}",
				},
				{
					Type:    token.RightParenthesis,
					Literal: ")",
				},
				{
					Type:    token.LeftCurlyBrace,
					Literal: "{",
				},
				{
					Type:    token.Asterisk,
					Literal: "*",
				},
				{
					Type:    token.NotEquals,
					Literal: "!=",
				},
				{
					Type:    token.Bang,
					Literal: "!",
				},
				{
					Type:    token.FullStop,
					Literal: ".",
				},
				{
					Type:    token.EOF,
					Literal: "EOF",
				},
			},
		},
		{
			src: "true && false || false",
			bl:  LexerBufferLength,
			expected: []token.Token{
				{
					Type:    token.Boolean,
					Literal: "true",
				},
				{
					Type:    token.And,
					Literal: "&&",
				},
				{
					Type:    token.Boolean,
					Literal: "false",
				},
				{
					Type:    token.Or,
					Literal: "||",
				},
				{
					Type:    token.Boolean,
					Literal: "false",
				},
				{
					Type:    token.EOF,
					Literal: "EOF",
				},
			},
		},
		{
			src: `
			import "math";
			# This makes no sense but it does not have to since this is a test
			# Will this work with two lines of comments?
			while (true) {
				return a[0];
			}
			`,
			bl: LexerBufferLength,
			expected: []token.Token{
				{
					Type:    token.Import,
					Literal: "import",
				},
				{
					Type:    token.String,
					Literal: "math",
				},
				{
					Type:    token.Semicolon,
					Literal: ";",
				},
				{
					Type:    token.While,
					Literal: "while",
				},
				{
					Type:    token.LeftParenthesis,
					Literal: "(",
				},
				{
					Type:    token.Boolean,
					Literal: "true",
				},
				{
					Type:    token.RightParenthesis,
					Literal: ")",
				},
				{
					Type:    token.LeftCurlyBrace,
					Literal: "{",
				},
				{
					Type:    token.Return,
					Literal: "return",
				},
				{
					Type:    token.Identifier,
					Literal: "a",
				},
				{
					Type:    token.LeftBracket,
					Literal: "[",
				},
				{
					Type:    token.Integer,
					Literal: "0",
				},
				{
					Type:    token.RightBracket,
					Literal: "]",
				},
				{
					Type:    token.Semicolon,
					Literal: ";",
				},
				{
					Type:    token.RightCurlyBrace,
					Literal: "}",
				},
				{
					Type:    token.EOF,
					Literal: "EOF",
				},
			},
		},
		{
			src: "let name = \"crookdc\";",
			bl:  LexerBufferLength,
			expected: []token.Token{
				{
					Type:    token.Let,
					Literal: "let",
				},
				{
					Type:    token.Identifier,
					Literal: "name",
				},
				{
					Type:    token.Assign,
					Literal: "=",
				},
				{
					Type:    token.String,
					Literal: "crookdc",
				},
				{
					Type:    token.Semicolon,
					Literal: ";",
				},
				{
					Type:    token.EOF,
					Literal: "EOF",
				},
			},
		},
		{
			src: `
			if (a > b) {
				let c = 5;
			}
			`,
			bl: LexerBufferLength,
			expected: []token.Token{
				{
					Type:    token.If,
					Literal: "if",
				},
				{
					Type:    token.LeftParenthesis,
					Literal: "(",
				},
				{
					Type:    token.Identifier,
					Literal: "a",
				},
				{
					Type:    token.GreaterThan,
					Literal: ">",
				},
				{
					Type:    token.Identifier,
					Literal: "b",
				},
				{
					Type:    token.RightParenthesis,
					Literal: ")",
				},
				{
					Type:    token.LeftCurlyBrace,
					Literal: "{",
				},
				{
					Type:    token.Let,
					Literal: "let",
				},
				{
					Type:    token.Identifier,
					Literal: "c",
				},
				{
					Type:    token.Assign,
					Literal: "=",
				},
				{
					Type:    token.Integer,
					Literal: "5",
				},
				{
					Type:    token.Semicolon,
					Literal: ";",
				},
				{
					Type:    token.RightCurlyBrace,
					Literal: "}",
				},
				{
					Type:    token.EOF,
					Literal: "EOF",
				},
			},
		},
		{
			src: "let developer = \"crookd\";",
			bl:  4,
			expected: []token.Token{
				{
					Type:    token.Let,
					Literal: "let",
				},
				{
					Type:    token.Identifier,
					Literal: "developer",
				},
				{
					Type:    token.Assign,
					Literal: "=",
				},
				{
					Type:    token.String,
					Literal: "crookd",
				},
				{
					Type:    token.Semicolon,
					Literal: ";",
				},
				{
					Type:    token.EOF,
					Literal: "EOF",
				},
			},
		},
		{
			src: "name + \"is a good developer\";",
			bl:  LexerBufferLength,
			expected: []token.Token{
				{
					Type:    token.Identifier,
					Literal: "name",
				},
				{
					Type:    token.Plus,
					Literal: "+",
				},
				{
					Type:    token.String,
					Literal: "is a good developer",
				},
				{
					Type:    token.Semicolon,
					Literal: ";",
				},
				{
					Type:    token.EOF,
					Literal: "EOF",
				},
			},
		},
		{
			src: `
			# This function reports whether both a and b are positive
			let pos = func(a, b) {
				# Holy cow, this is a comment isn't it!
				return a > 0 && b > 0;
			};
			`,
			bl: LexerBufferLength,
			expected: []token.Token{
				{
					Type:    token.Let,
					Literal: "let",
				},
				{
					Type:    token.Identifier,
					Literal: "pos",
				},
				{
					Type:    token.Assign,
					Literal: "=",
				},
				{
					Type:    token.Function,
					Literal: "func",
				},
				{
					Type:    token.LeftParenthesis,
					Literal: "(",
				},
				{
					Type:    token.Identifier,
					Literal: "a",
				},
				{
					Type:    token.Comma,
					Literal: ",",
				},
				{
					Type:    token.Identifier,
					Literal: "b",
				},
				{
					Type:    token.RightParenthesis,
					Literal: ")",
				},
				{
					Type:    token.LeftCurlyBrace,
					Literal: "{",
				},
				{
					Type:    token.Return,
					Literal: "return",
				},
				{
					Type:    token.Identifier,
					Literal: "a",
				},
				{
					Type:    token.GreaterThan,
					Literal: ">",
				},
				{
					Type:    token.Integer,
					Literal: "0",
				},
				{
					Type:    token.And,
					Literal: "&&",
				},
				{
					Type:    token.Identifier,
					Literal: "b",
				},
				{
					Type:    token.GreaterThan,
					Literal: ">",
				},
				{
					Type:    token.Integer,
					Literal: "0",
				},
				{
					Type:    token.Semicolon,
					Literal: ";",
				},
				{
					Type:    token.RightCurlyBrace,
					Literal: "}",
				},
				{
					Type:    token.Semicolon,
					Literal: ";",
				},
				{
					Type:    token.EOF,
					Literal: "EOF",
				},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.src, func(t *testing.T) {
			lx, err := NewLexer(strings.NewReader(test.src), BufferLength(test.bl))
			assert.Nil(t, err)
			var i int
			for {
				actual, err := lx.Next()
				assert.Nil(t, err)
				assert.Equal(t, test.expected[i], actual, "token index %d", i)
				i++
				if actual.Type == token.EOF {
					break
				}
			}
		})
	}
}

func TestNewPeekingLexer(t *testing.T) {
	t.Run("given nil lexer", func(t *testing.T) {
		_, err := NewPeekingLexer(nil)
		assert.True(t, errors.Is(err, ErrInvalidSourceLexer))
	})
	t.Run("given non-nil lexer", func(t *testing.T) {
		src := "let a = b;"
		lx, err := NewLexer(strings.NewReader(src))
		assert.Nil(t, err)
		_, err = NewPeekingLexer(lx)
		assert.Nil(t, err)
	})
}

func TestPeekingLexer_Peek(t *testing.T) {
	src := "let a = b;"

	lx, err := NewLexer(strings.NewReader(src))
	assert.Nil(t, err)

	plx, err := NewPeekingLexer(lx)
	assert.Nil(t, err)

	tok, err := plx.Peek()
	assert.Nil(t, err)
	assert.Equal(t, token.Token{Type: token.Let, Literal: "let"}, tok)

	tok, err = plx.Peek()
	assert.Nil(t, err)
	assert.Equal(t, token.Token{Type: token.Let, Literal: "let"}, tok)

	tok, err = plx.Next()
	assert.Nil(t, err)
	assert.Equal(t, token.Token{Type: token.Let, Literal: "let"}, tok)

	tok, err = plx.Peek()
	assert.Nil(t, err)
	assert.Equal(t, token.Token{Type: token.Identifier, Literal: "a"}, tok)

	tok, err = plx.Peek()
	assert.Nil(t, err)
	assert.Equal(t, token.Token{Type: token.Identifier, Literal: "a"}, tok)
}

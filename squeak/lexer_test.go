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

func TestLexer_Line(t *testing.T) {
	tests := []struct {
		src      string
		expected int
	}{
		{
			src:      "  let = 4444;",
			expected: 1,
		},
		{
			src: `				// 1
			if (a == a) {		// 2
				return true;	// 3
			}					// 4
								// 5`,
			expected: 5,
		},
		{
			src:      "let\nit\nsnow\n\n\n\n",
			expected: 7,
		},
	}
	for _, test := range tests {
		t.Run(test.src, func(t *testing.T) {
			lx, err := NewLexer(strings.NewReader(test.src))
			assert.Nil(t, err)
			for {
				tok, err := lx.Next()
				assert.Nil(t, err)
				if tok.Type == token.EOF {
					break
				}
			}
			assert.Equal(t, test.expected, lx.Line())
		})
	}
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
					Type:   token.Let,
					Lexeme: "let",
				},
				{
					Type:   token.Assign,
					Lexeme: "=",
				},
				{
					Type:   token.Integer,
					Lexeme: "512",
				},
				{
					Type:   token.Semicolon,
					Lexeme: ";",
				},
				{
					Type:   token.EOF,
					Lexeme: "EOF",
				},
			},
		},
		{
			src: "print 15;",
			bl:  LexerBufferLength,
			expected: []token.Token{
				{
					Type:   token.Print,
					Lexeme: "print",
				},
				{
					Type:   token.Integer,
					Lexeme: "15",
				},
				{
					Type:   token.Semicolon,
					Lexeme: ";",
				},
				{
					Type:   token.EOF,
					Lexeme: "EOF",
				},
			},
		},
		{
			src: "a+b/and(}){*!=!.",
			bl:  LexerBufferLength,
			expected: []token.Token{
				{
					Type:   token.Identifier,
					Lexeme: "a",
				},
				{
					Type:   token.Plus,
					Lexeme: "+",
				},
				{
					Type:   token.Identifier,
					Lexeme: "b",
				},
				{
					Type:   token.Slash,
					Lexeme: "/",
				},
				{
					Type:   token.And,
					Lexeme: "and",
				},
				{
					Type:   token.LeftParenthesis,
					Lexeme: "(",
				},
				{
					Type:   token.RightCurlyBrace,
					Lexeme: "}",
				},
				{
					Type:   token.RightParenthesis,
					Lexeme: ")",
				},
				{
					Type:   token.LeftCurlyBrace,
					Lexeme: "{",
				},
				{
					Type:   token.Asterisk,
					Lexeme: "*",
				},
				{
					Type:   token.NotEquals,
					Lexeme: "!=",
				},
				{
					Type:   token.Bang,
					Lexeme: "!",
				},
				{
					Type:   token.Dot,
					Lexeme: ".",
				},
				{
					Type:   token.EOF,
					Lexeme: "EOF",
				},
			},
		},
		{
			src: "true and false or false",
			bl:  LexerBufferLength,
			expected: []token.Token{
				{
					Type:   token.Boolean,
					Lexeme: "true",
				},
				{
					Type:   token.And,
					Lexeme: "and",
				},
				{
					Type:   token.Boolean,
					Lexeme: "false",
				},
				{
					Type:   token.Or,
					Lexeme: "or",
				},
				{
					Type:   token.Boolean,
					Lexeme: "false",
				},
				{
					Type:   token.EOF,
					Lexeme: "EOF",
				},
			},
		},
		{
			src: `
			# This makes no sense but it does not have to since this is a test
			# Will this work with two lines of comments?
			while (true) {
				return a[0];
			}
			`,
			bl: LexerBufferLength,
			expected: []token.Token{
				{
					Type:   token.While,
					Lexeme: "while",
				},
				{
					Type:   token.LeftParenthesis,
					Lexeme: "(",
				},
				{
					Type:   token.Boolean,
					Lexeme: "true",
				},
				{
					Type:   token.RightParenthesis,
					Lexeme: ")",
				},
				{
					Type:   token.LeftCurlyBrace,
					Lexeme: "{",
				},
				{
					Type:   token.Return,
					Lexeme: "return",
				},
				{
					Type:   token.Identifier,
					Lexeme: "a",
				},
				{
					Type:   token.LeftBracket,
					Lexeme: "[",
				},
				{
					Type:   token.Integer,
					Lexeme: "0",
				},
				{
					Type:   token.RightBracket,
					Lexeme: "]",
				},
				{
					Type:   token.Semicolon,
					Lexeme: ";",
				},
				{
					Type:   token.RightCurlyBrace,
					Lexeme: "}",
				},
				{
					Type:   token.EOF,
					Lexeme: "EOF",
				},
			},
		},
		{
			src: "let name = \"crookdc\";",
			bl:  LexerBufferLength,
			expected: []token.Token{
				{
					Type:   token.Let,
					Lexeme: "let",
				},
				{
					Type:   token.Identifier,
					Lexeme: "name",
				},
				{
					Type:   token.Assign,
					Lexeme: "=",
				},
				{
					Type:   token.String,
					Lexeme: "crookdc",
				},
				{
					Type:   token.Semicolon,
					Lexeme: ";",
				},
				{
					Type:   token.EOF,
					Lexeme: "EOF",
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
					Type:   token.If,
					Lexeme: "if",
				},
				{
					Type:   token.LeftParenthesis,
					Lexeme: "(",
				},
				{
					Type:   token.Identifier,
					Lexeme: "a",
				},
				{
					Type:   token.Greater,
					Lexeme: ">",
				},
				{
					Type:   token.Identifier,
					Lexeme: "b",
				},
				{
					Type:   token.RightParenthesis,
					Lexeme: ")",
				},
				{
					Type:   token.LeftCurlyBrace,
					Lexeme: "{",
				},
				{
					Type:   token.Let,
					Lexeme: "let",
				},
				{
					Type:   token.Identifier,
					Lexeme: "c",
				},
				{
					Type:   token.Assign,
					Lexeme: "=",
				},
				{
					Type:   token.Integer,
					Lexeme: "5",
				},
				{
					Type:   token.Semicolon,
					Lexeme: ";",
				},
				{
					Type:   token.RightCurlyBrace,
					Lexeme: "}",
				},
				{
					Type:   token.EOF,
					Lexeme: "EOF",
				},
			},
		},
		{
			src: "let developer = \"crookd\";",
			bl:  4,
			expected: []token.Token{
				{
					Type:   token.Let,
					Lexeme: "let",
				},
				{
					Type:   token.Identifier,
					Lexeme: "developer",
				},
				{
					Type:   token.Assign,
					Lexeme: "=",
				},
				{
					Type:   token.String,
					Lexeme: "crookd",
				},
				{
					Type:   token.Semicolon,
					Lexeme: ";",
				},
				{
					Type:   token.EOF,
					Lexeme: "EOF",
				},
			},
		},
		{
			src: "name + \"is a good developer\";",
			bl:  LexerBufferLength,
			expected: []token.Token{
				{
					Type:   token.Identifier,
					Lexeme: "name",
				},
				{
					Type:   token.Plus,
					Lexeme: "+",
				},
				{
					Type:   token.String,
					Lexeme: "is a good developer",
				},
				{
					Type:   token.Semicolon,
					Lexeme: ";",
				},
				{
					Type:   token.EOF,
					Lexeme: "EOF",
				},
			},
		},
		{
			src: `
			# This function reports whether both a and b are positive
			let pos = func(a, b) {
				# Holy cow, this is a comment isn't it!
				return a > 0 and b > 0;
			};
			# Sometimes there are comments at the very end of the source code!
			# It's important that we cover those as well.`,
			bl: LexerBufferLength,
			expected: []token.Token{
				{
					Type:   token.Let,
					Lexeme: "let",
				},
				{
					Type:   token.Identifier,
					Lexeme: "pos",
				},
				{
					Type:   token.Assign,
					Lexeme: "=",
				},
				{
					Type:   token.Function,
					Lexeme: "func",
				},
				{
					Type:   token.LeftParenthesis,
					Lexeme: "(",
				},
				{
					Type:   token.Identifier,
					Lexeme: "a",
				},
				{
					Type:   token.Comma,
					Lexeme: ",",
				},
				{
					Type:   token.Identifier,
					Lexeme: "b",
				},
				{
					Type:   token.RightParenthesis,
					Lexeme: ")",
				},
				{
					Type:   token.LeftCurlyBrace,
					Lexeme: "{",
				},
				{
					Type:   token.Return,
					Lexeme: "return",
				},
				{
					Type:   token.Identifier,
					Lexeme: "a",
				},
				{
					Type:   token.Greater,
					Lexeme: ">",
				},
				{
					Type:   token.Integer,
					Lexeme: "0",
				},
				{
					Type:   token.And,
					Lexeme: "and",
				},
				{
					Type:   token.Identifier,
					Lexeme: "b",
				},
				{
					Type:   token.Greater,
					Lexeme: ">",
				},
				{
					Type:   token.Integer,
					Lexeme: "0",
				},
				{
					Type:   token.Semicolon,
					Lexeme: ";",
				},
				{
					Type:   token.RightCurlyBrace,
					Lexeme: "}",
				},
				{
					Type:   token.Semicolon,
					Lexeme: ";",
				},
				{
					Type:   token.EOF,
					Lexeme: "EOF",
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

func TestPeekingLexer_Line(t *testing.T) {
	src := "let \na\n = \nb;"

	lx, err := NewLexer(strings.NewReader(src))
	assert.Nil(t, err)

	plx, err := NewPeekingLexer(lx)
	assert.Nil(t, err)

	tok, err := plx.Peek()
	assert.Nil(t, err)
	assert.Equal(t, token.Token{Type: token.Let, Lexeme: "let"}, tok)
	assert.Equal(t, 1, plx.Line())

	tok, err = plx.Peek()
	assert.Nil(t, err)
	assert.Equal(t, token.Token{Type: token.Let, Lexeme: "let"}, tok)
	assert.Equal(t, 1, plx.Line())

	tok, err = plx.Next()
	assert.Nil(t, err)
	assert.Equal(t, token.Token{Type: token.Let, Lexeme: "let"}, tok)
	assert.Equal(t, 1, plx.Line())

	tok, err = plx.Peek()
	assert.Nil(t, err)
	assert.Equal(t, token.Token{Type: token.Identifier, Lexeme: "a"}, tok)
	assert.Equal(t, 2, plx.Line())

	tok, err = plx.Next()
	assert.Nil(t, err)
	assert.Equal(t, token.Token{Type: token.Identifier, Lexeme: "a"}, tok)
	assert.Equal(t, 2, plx.Line())

	tok, err = plx.Peek()
	assert.Nil(t, err)
	assert.Equal(t, token.Token{Type: token.Assign, Lexeme: "="}, tok)
	assert.Equal(t, 3, plx.Line())

	tok, err = plx.Next()
	assert.Nil(t, err)
	assert.Equal(t, token.Token{Type: token.Assign, Lexeme: "="}, tok)
	assert.Equal(t, 3, plx.Line())

	tok, err = plx.Next()
	assert.Nil(t, err)
	assert.Equal(t, token.Token{Type: token.Identifier, Lexeme: "b"}, tok)
	assert.Equal(t, 4, plx.Line())

	tok, err = plx.Next()
	assert.Nil(t, err)
	assert.Equal(t, token.Token{Type: token.Semicolon, Lexeme: ";"}, tok)
	assert.Equal(t, 4, plx.Line())
}

func TestPeekingLexer_Peek(t *testing.T) {
	src := "let a = b;"

	lx, err := NewLexer(strings.NewReader(src))
	assert.Nil(t, err)

	plx, err := NewPeekingLexer(lx)
	assert.Nil(t, err)

	tok, err := plx.Peek()
	assert.Nil(t, err)
	assert.Equal(t, token.Token{Type: token.Let, Lexeme: "let"}, tok)

	tok, err = plx.Peek()
	assert.Nil(t, err)
	assert.Equal(t, token.Token{Type: token.Let, Lexeme: "let"}, tok)

	tok, err = plx.Next()
	assert.Nil(t, err)
	assert.Equal(t, token.Token{Type: token.Let, Lexeme: "let"}, tok)

	tok, err = plx.Peek()
	assert.Nil(t, err)
	assert.Equal(t, token.Token{Type: token.Identifier, Lexeme: "a"}, tok)

	tok, err = plx.Peek()
	assert.Nil(t, err)
	assert.Equal(t, token.Token{Type: token.Identifier, Lexeme: "a"}, tok)
}

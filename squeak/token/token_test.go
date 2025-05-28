package token

import (
	"errors"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNew(t *testing.T) {
	t.Run("missing token literal without default mapping", func(t *testing.T) {
		_, err := New(Integer)
		assert.True(t, errors.Is(err, ErrMissingLexeme))
	})
	t.Run("missing token literal with default mapping", func(t *testing.T) {
		token, err := New(Assign)
		assert.Nil(t, err)
		assert.Equal(t, Assign, token.Type)
		assert.Equal(t, "=", token.Lexeme)
	})
	t.Run("provided token literal without default mapping", func(t *testing.T) {
		token, err := New(Integer, Lexeme("15"))
		assert.Nil(t, err)
		assert.Equal(t, Integer, token.Type)
		assert.Equal(t, "15", token.Lexeme)
	})
	t.Run("provided token literal with default mapping", func(t *testing.T) {
		token, err := New(Assign, Lexeme("=/="))
		assert.Nil(t, err)
		assert.Equal(t, Assign, token.Type)
		assert.Equal(t, "=/=", token.Lexeme)
	})
}

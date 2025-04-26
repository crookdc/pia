package token

import (
	"errors"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNew(t *testing.T) {
	t.Run("missing token literal without default mapping", func(t *testing.T) {
		_, err := New(Integer)
		assert.True(t, errors.Is(err, ErrMissingLiteral))
	})
	t.Run("missing token literal with default mapping", func(t *testing.T) {
		token, err := New(If)
		assert.Nil(t, err)
		assert.Equal(t, If, token.Type)
		assert.Equal(t, "if", token.Literal)
	})
	t.Run("provided token literal without default mapping", func(t *testing.T) {
		token, err := New(Integer, Literal("15"))
		assert.Nil(t, err)
		assert.Equal(t, Integer, token.Type)
		assert.Equal(t, "15", token.Literal)
	})
	t.Run("provided token literal with default mapping", func(t *testing.T) {
		token, err := New(If, Literal("IF-ELSE"))
		assert.Nil(t, err)
		assert.Equal(t, If, token.Type)
		assert.Equal(t, "IF-ELSE", token.Literal)
	})
}

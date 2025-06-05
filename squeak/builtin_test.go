package squeak

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestPrintBuiltin_Arity(t *testing.T) {
	assert.Equal(t, 1, PrintBuiltin{}.Arity())
}

func TestLengthBuiltin_Arity(t *testing.T) {
	assert.Equal(t, 1, LengthBuiltin{}.Arity())
}

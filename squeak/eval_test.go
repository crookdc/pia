package squeak

import (
	"errors"
	"github.com/crookdc/pia/squeak/internal/ast"
	"github.com/stretchr/testify/assert"
	"io"
	"strings"
	"testing"
)

func TestEvaluator_Statement(t *testing.T) {
	type result struct {
		object Object
		err    error
	}
	tests := []struct {
		name     string
		src      string
		expected []result
	}{
		{
			name: "integer addition",
			src:  "5 + 5;",
			expected: []result{
				{
					object: Integer{10},
				},
			},
		},
		{
			name: "triple integer addition",
			src:  "5 + 5 + 113;",
			expected: []result{
				{
					object: Integer{123},
				},
			},
		},
		{
			name: "string addition",
			src:  "\"hello\" + \" \" + \"world\";",
			expected: []result{
				{
					object: String{"hello world"},
				},
			},
		},
		{
			name: "string integer addition",
			src:  "\"hello\" + 15 + \"world\";",
			expected: []result{
				{
					object: String{"hello15world"},
				},
			},
		},
		{
			name: "integer string addition",
			src:  "15 + \" world records\";",
			expected: []result{
				{
					object: String{"15 world records"},
				},
			},
		},
		{
			name: "illegal boolean addition",
			src:  "true + false;",
			expected: []result{
				{
					err: ErrUnexpectedType,
				},
			},
		},
		{
			name: "integer subtraction",
			src:  "15 - 20 + 50;",
			expected: []result{
				{
					object: Integer{45},
				},
			},
		},
		{
			name: "illegal string subtraction",
			src:  "\"hello\" - \"world\";",
			expected: []result{
				{
					err: ErrUnexpectedType,
				},
			},
		},
		{
			name: "illegal boolean subtraction",
			src:  "true - false + true;",
			expected: []result{
				{
					err: ErrUnexpectedType,
				},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			lx, err := NewLexer(strings.NewReader(test.src))
			assert.Nil(t, err)
			plx, err := NewPeekingLexer(lx)
			assert.Nil(t, err)
			ps := NewParser(plx)
			script := make([]ast.StatementNode, 0)
			for {
				stmt, err := ps.Next()
				if errors.Is(err, io.EOF) {
					break
				}
				assert.Nil(t, err)
				script = append(script, stmt)
			}

			ev := NewEvaluator()
			for i, stmt := range script {
				obj, err := ev.Run(stmt)
				assert.Equal(t, test.expected[i].err, err)
				assert.Equal(t, test.expected[i].object, obj)
			}
		})
	}
}

package pia

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"regexp"
)

var (
	ErrKeyNotFound                   = errors.New("key not found")
	ErrInsufficientDestinationLength = errors.New("destination size for substituting reader must be greater than 2")
)

// KeyResolver is a simple abstraction of a key-value store for string values.
type KeyResolver interface {
	// Resolve takes a key and returns a value from the underlying store. An error value may be returned for any
	// erroneous reason but for the case where a key cannot be resolved it is expected to return [pia.ErrKeyNotFound].
	Resolve(k string) (string, error)
}

// MapResolver is the simplest possible implementation of [pia.KeyResolver] using an underlying map to facilitate
// storage and retrieval.
type MapResolver map[string]string

// Resolve implements the [pia.KeyResolver] interface.
func (m MapResolver) Resolve(k string) (string, error) {
	v, ok := m[k]
	if !ok {
		return "", fmt.Errorf("failed to resolve key '%s': %w", k, ErrKeyNotFound)
	}
	return v, nil
}

// WrapReader returns a [pia.SubstitutingReader] that uses the supplied [pia.KeyResolver] as source for substitutions
// and the supplied [io.Reader] as its target.
func WrapReader(resolver KeyResolver, r io.Reader) *SubstitutingReader {
	return &SubstitutingReader{
		prefix:   regexp.MustCompile("\\${"),
		matcher:  regexp.MustCompile("\\${(.+?)}"),
		resolver: resolver,
		wrapped:  bufio.NewReader(r),
		carry:    bytes.NewBuffer(make([]byte, 0)),
	}
}

// SubstitutingReader decorates the wrapped reader by replacing any occurrences of substitution points defined using
// "${key}" syntax with the corresponding value according to the supplied [pia.KeyResolver]. Since a [pia.KeyResolver]
// is powering the substitution, all values must be supplied as strings and will be placed unquoted into the stream.
type SubstitutingReader struct {
	prefix  *regexp.Regexp
	matcher *regexp.Regexp

	resolver KeyResolver
	wrapped  *bufio.Reader
	carry    io.ReadWriter
}

// Read implements the [io.Reader] interface for seamless interoperability with the Go standard library.
func (s SubstitutingReader) Read(p []byte) (int, error) {
	if len(p) < 2 {
		// Destination must be able to contain at least "${" for SubstitutingReader to be able to find substitution
		// points. As such, the buffer must be at least two bytes long.
		return 0, ErrInsufficientDestinationLength
	}
	str, err := s.read(len(p))
	if err != nil {
		return 0, err
	}
	for _, match := range s.match(str) {
		str, err = s.substitute(str, match)
		if err != nil {
			return 0, err
		}
	}
	copy(p, str)
	if len(str) > len(p) {
		_, err := s.carry.Write([]byte(str)[len(p):])
		if err != nil {
			return 0, err
		}
		return len(p), nil
	}
	return len(str), nil
}

// read returns the next processable chunk of data using ln as the pivoting length. It is possible that read returns
// both shorter and longer strings based on the data residing within the raw read data.
func (s SubstitutingReader) read(ln int) (string, error) {
	str, err := s.raw(ln)
	if err != nil {
		return "", err
	}
	if str[len(str)-1] == '$' {
		// There is a possibility that the next rune in from the reader could be an open curly brace, which together
		// with the dollar sign becomes the prefix of data that should be substituted. Therefore, read the next byte and
		// include it in the string being processed.
		extra, err := s.wrapped.ReadByte()
		if err != nil && !errors.Is(err, io.EOF) {
			return "", err
		} else if err == nil {
			str += string(extra)
		}
	}
	if s.partials(str) {
		// We need to go further into the wrapped reader to make sure we can substitute the next one as well
		extra, err := s.wrapped.ReadString('}')
		if errors.Is(err, io.EOF) {
			err = nil
		}
		if err != nil {
			return "", err
		}
		str += extra
	}
	return str, nil
}

// raw returns ln or fewer bytes of data from the carry and wrapped readers without performing any modifications or
// additions to the data.
func (s SubstitutingReader) raw(ln int) (string, error) {
	p := make([]byte, ln)
	cn, err := s.carry.Read(p)
	if errors.Is(err, io.EOF) {
		// The carry is expected to not contain data with each read and thus receiving an EOF error is considered
		// normal.
		cn = 0
	} else if err != nil {
		return "", err
	}
	if cn == len(p) {
		return string(p), nil
	}

	wn, err := s.wrapped.Read(p[cn:])
	if errors.Is(err, io.EOF) && cn > 0 {
		// If the wrapped reader has reached EOF, but there is data still being read from the carry then continue
		// reading only from the carry.
		return string(p[:cn]), nil
	} else if err != nil {
		return "", err
	}

	return string(p[:cn+wn]), nil
}

func (s SubstitutingReader) substitute(str string, match []string) (string, error) {
	val, err := s.resolver.Resolve(match[1])
	if err != nil {
		return "", err
	}
	// Compiles a new [regexp.Regexp] value with each substitution which is not optimal for performance. Remains to be
	// measured and benchmarked.
	return regexp.MustCompile(regexp.QuoteMeta(match[0])).ReplaceAllLiteralString(str, val), nil
}

func (s SubstitutingReader) match(str string) [][]string {
	return s.matcher.FindAllStringSubmatch(str, -1)
}

// partials reports whether the supplied string is potentially containing partial substitution points with respect to
// the rest of the stream. However, this cannot be guaranteed without potentially reading the entire wrapped stream and
// checking whether there exists a terminating character "}".
func (s SubstitutingReader) partials(str string) bool {
	return s.prefixes(str) > s.points(str)
}

// pristine reports whether the supplied string is devoid of any substitution points
func (s SubstitutingReader) pristine(str string) bool {
	return s.prefixes(str) == 0
}

func (s SubstitutingReader) points(str string) int {
	return len(s.matcher.FindAllStringIndex(str, -1))
}

func (s SubstitutingReader) prefixes(str string) int {
	return len(s.prefix.FindAllStringIndex(str, -1))
}

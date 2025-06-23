package pia

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"regexp"
)

var (
	ErrKeyNotFound                   = errors.New("key not found")
	ErrInsufficientDestinationLength = errors.New("destination size for substituting reader must be greater than 2")
)

// WrapReader returns a [pia.Interpolator] that uses the supplied [pia.KeyResolver] as source for substitutions
// and the supplied [io.Reader] as its target.
func WrapReader(resolver KeyResolver, r io.Reader) *Interpolator {
	return &Interpolator{
		prefix:   regexp.MustCompile(`\${`),
		matcher:  regexp.MustCompile(`\${(.+?)}`),
		resolver: resolver,
		wrapped:  bufio.NewReader(r),
		carry:    bytes.NewBuffer(make([]byte, 0)),
	}
}

// Interpolator decorates the wrapped reader by replacing any occurrences of substitution points defined using
// "${key}" syntax with the corresponding value according to the supplied [pia.KeyResolver]. Since a [pia.KeyResolver]
// is powering the substitution, all values must be supplied as strings and will be placed unquoted into the stream.
type Interpolator struct {
	prefix  *regexp.Regexp
	matcher *regexp.Regexp

	resolver KeyResolver
	wrapped  *bufio.Reader
	carry    io.ReadWriter
}

// Read implements the [io.Reader] interface for seamless interoperability with the Go standard library.
func (ip Interpolator) Read(p []byte) (int, error) {
	if len(p) < 2 {
		// Destination must be able to contain at least "${" for Interpolator to be able to find substitution
		// points. As such, the buffer must be at least two bytes long.
		return 0, ErrInsufficientDestinationLength
	}
	str, err := ip.read(len(p))
	if err != nil {
		return 0, err
	}
	for _, match := range ip.match(str) {
		str, err = ip.substitute(str, match)
		if err != nil {
			return 0, err
		}
	}
	copy(p, str)
	if len(str) > len(p) {
		_, err := ip.carry.Write([]byte(str)[len(p):])
		if err != nil {
			return 0, err
		}
		return len(p), nil
	}
	return len(str), nil
}

// read returns the next processable chunk of data using ln as the pivoting length. It is possible that read returns
// both shorter and longer strings based on the data residing within the raw read data.
func (ip Interpolator) read(ln int) (string, error) {
	str, err := ip.raw(ln)
	if err != nil {
		return "", err
	}
	if str[len(str)-1] == '$' {
		// There is a possibility that the next rune in from the reader could be an open curly brace, which together
		// with the dollar sign becomes the prefix of data that should be substituted. Therefore, read the next byte and
		// include it in the string being processed.
		extra, err := ip.wrapped.ReadByte()
		if err != nil && !errors.Is(err, io.EOF) {
			return "", err
		} else if err == nil {
			str += string(extra)
		}
	}
	if ip.partials(str) {
		// We need to go further into the wrapped reader to make sure we can substitute the next one as well
		extra, err := ip.wrapped.ReadString('}')
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
func (ip Interpolator) raw(ln int) (string, error) {
	p := make([]byte, ln)
	cn, err := ip.carry.Read(p)
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

	wn, err := ip.wrapped.Read(p[cn:])
	if errors.Is(err, io.EOF) && cn > 0 {
		// If the wrapped reader has reached EOF, but there is data still being read from the carry then continue
		// reading only from the carry.
		return string(p[:cn]), nil
	} else if err != nil {
		return "", err
	}

	return string(p[:cn+wn]), nil
}

func (ip Interpolator) substitute(str string, match []string) (string, error) {
	val, err := ip.resolver.Resolve(match[1])
	if err != nil {
		return "", err
	}
	// Compiles a new [regexp.Regexp] value with each substitution which is not optimal for performance. Remains to be
	// measured and benchmarked.
	return regexp.MustCompile(regexp.QuoteMeta(match[0])).ReplaceAllLiteralString(str, val), nil
}

func (ip Interpolator) match(str string) [][]string {
	return ip.matcher.FindAllStringSubmatch(str, -1)
}

// partials reports whether the supplied string is potentially containing partial substitution points with respect to
// the rest of the stream. However, this cannot be guaranteed without potentially reading the entire wrapped stream and
// checking whether there exists a terminating character "}".
func (ip Interpolator) partials(str string) bool {
	return ip.prefixes(str) > ip.points(str)
}

func (ip Interpolator) points(str string) int {
	return len(ip.matcher.FindAllStringIndex(str, -1))
}

func (ip Interpolator) prefixes(str string) int {
	return len(ip.prefix.FindAllStringIndex(str, -1))
}

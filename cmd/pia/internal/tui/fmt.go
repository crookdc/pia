package tui

import (
	"fmt"
	"io"
	"net/http"
	"strings"
)

type Formatter[T fmt.Stringer] interface {
	Format(io.Writer, T) error
}

type FormatterFunc[T fmt.Stringer] func(io.Writer, T) error

func ResponseFormatter(w io.Writer, res *http.Response) error {
	_, err := fmt.Fprintf(w, "Status: %s\n", res.Status)
	if err != nil {
		return err
	}
	for k, v := range res.Header {
		_, err = fmt.Fprintf(w, "%s: %s\n", k, strings.Join(v, ", "))
		if err != nil {
			return err
		}
	}
	_, err = fmt.Fprint(w, "\n\n")
	if err != nil {
		return err
	}
	if res.Body != nil {
		raw, err := io.ReadAll(res.Body)
		if err != nil {
			return err
		}
		_, err = fmt.Fprint(w, string(raw))
		if err != nil {
			return err
		}
	}
	return nil
}

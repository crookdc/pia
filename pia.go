package pia

import (
	"bytes"
	"github.com/crookdc/pia/squeak"
	"io"
	"net/http"
)

type Pia struct {
	WorkingDirectory string
	Resolver         KeyResolver
	Output           io.Writer
}

func (p *Pia) Execute(tx *Transaction) (*http.Response, error) {
	req, err := tx.Request()
	if err != nil {
		return nil, err
	}
	if tx.Hooks.Before != nil {
		ast, err := squeak.Parse(tx.Hooks.Before)
		if err != nil {
			return nil, err
		}
		in := squeak.NewInterpreter(p.WorkingDirectory, p.Output)
		in.Declare("request", squeak.NewRequestObject(req))
		if err := in.Execute(ast); err != nil {
			return nil, err
		}
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	if tx.Hooks.After != nil {
		ast, err := squeak.Parse(tx.Hooks.After)
		if err != nil {
			return nil, err
		}
		in := squeak.NewInterpreter(p.WorkingDirectory, p.Output)
		var body []byte
		if res.Body != nil {
			body, err = io.ReadAll(res.Body)
			if err != nil {
				return nil, err
			}
			// Allow the response body to be re-read by assigning a new io.Reader to it.
			res.Body = io.NopCloser(bytes.NewBuffer(body))
		}
		in.Declare("response", squeak.NewResponseObject(res, body))
		if err := in.Execute(ast); err != nil {
			return nil, err
		}
	}
	return res, nil
}

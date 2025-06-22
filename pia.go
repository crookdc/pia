package pia

import (
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
		in.Declare("response", squeak.NewResponseObject(res))
		if err := in.Execute(ast); err != nil {
			return nil, err
		}
	}
	return res, nil
}

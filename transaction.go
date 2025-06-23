package pia

import (
	"gopkg.in/yaml.v3"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

type input struct {
	File   string `yaml:"file"`
	Inline string `yaml:"inline"`
}

func (in *input) reader(wd string) (io.Reader, error) {
	if in.Inline != "" {
		return strings.NewReader(in.Inline), nil
	}
	if in.File != "" {
		path := in.File
		if !filepath.IsAbs(path) {
			path = filepath.Join(wd, path)
		}
		return os.OpenFile(path, os.O_RDONLY, os.ModeAppend)
	}
	return nil, nil
}

type body struct {
	input `yaml:",inline"`
	Form  map[string]string `yaml:"form"`
}

func (b *body) reader(wd string) (io.Reader, error) {
	if len(b.Form) == 0 {
		return b.input.reader(wd)
	}
	body := url.Values{}
	for k, v := range b.Form {
		body.Set(k, v)
	}
	return strings.NewReader(body.Encode()), nil
}

// transaction represents a Transaction value in its textual YAML state. This data structure serves as a simple midway
// stop while parsing text data into a Transaction.
type transaction struct {
	URL struct {
		Target string            `yaml:"target"`
		Query  map[string]string `yaml:"query"`
	} `yaml:"url"`
	Method  string            `yaml:"method"`
	Headers map[string]string `yaml:"headers"`
	Body    body              `yaml:"body"`
	Hooks   struct {
		Before input `yaml:"before"`
		After  input `yaml:"after"`
	} `yaml:"hooks"`
}

// ParseTransaction reads the provided transaction configuration and builds a Transaction value from it.
func ParseTransaction(wd string, r io.Reader) (*Transaction, error) {
	var cfg transaction
	err := yaml.NewDecoder(r).Decode(&cfg)
	if err != nil {
		return nil, err
	}
	tx := Transaction{
		URL: struct {
			Target string
			Query  map[string]string
		}{
			Target: cfg.URL.Target,
			Query:  cfg.URL.Query,
		},
		Method:  cfg.Method,
		Headers: cfg.Headers,
	}

	tx.Body, err = cfg.Body.reader(wd)
	if err != nil {
		return nil, err
	}
	tx.Hooks.Before, err = cfg.Hooks.Before.reader(wd)
	if err != nil {
		return nil, err
	}
	tx.Hooks.After, err = cfg.Hooks.After.reader(wd)
	if err != nil {
		return nil, err
	}
	return &tx, nil
}

type Transaction struct {
	URL struct {
		Target string
		Query  map[string]string
	}
	Method  string
	Headers map[string]string
	Body    io.Reader
	Hooks   struct {
		Before io.Reader
		After  io.Reader
	}
}

// Request returns an [http.Request] which mirrors the configuration represented by the Transaction. The ownership of
// the request value is given to the caller, this means that the Transaction struct will not keep any reference to the
// produced request after returning and eventually closing the request is up to the caller.
func (tx *Transaction) Request() (*http.Request, error) {
	req, err := http.NewRequest(tx.Method, tx.URL.Target, tx.Body)
	if err != nil {
		return nil, err
	}
	q := req.URL.Query()
	for k, v := range tx.URL.Query {
		q.Add(k, v)
	}
	req.URL.RawQuery = q.Encode()
	for k, v := range tx.Headers {
		req.Header.Set(k, v)
	}
	return req, nil
}

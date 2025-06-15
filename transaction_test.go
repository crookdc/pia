package pia

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"
)

func TestTransaction_Request(t *testing.T) {
	parse := func(u string) *url.URL {
		v, err := url.Parse(u)
		if err != nil {
			panic(err)
		}
		return v
	}
	tests := []struct {
		tx  *Transaction
		req struct {
			URL    *url.URL
			Method string
			Header http.Header
			Body   []byte
		}
		err error
	}{
		{
			tx: &Transaction{
				URL: struct {
					Target string
					Query  map[string]string
				}{
					Target: "https://google.com/",
				},
				Method: http.MethodGet,
			},
			req: struct {
				URL    *url.URL
				Method string
				Header http.Header
				Body   []byte
			}{
				URL:    parse("https://google.com/"),
				Method: http.MethodGet,
				Header: http.Header{},
			},
		},
		{
			tx: &Transaction{
				URL: struct {
					Target string
					Query  map[string]string
				}{
					Target: "https://google.com/",
					Query: map[string]string{
						"s": "hello darkness my old friend",
						"p": "1",
					},
				},
				Method: http.MethodGet,
			},
			req: struct {
				URL    *url.URL
				Method string
				Header http.Header
				Body   []byte
			}{
				URL:    parse("https://google.com/?p=1&s=hello+darkness+my+old+friend"),
				Method: http.MethodGet,
				Header: http.Header{},
			},
		},
		{
			tx: &Transaction{
				URL: struct {
					Target string
					Query  map[string]string
				}{
					Target: "https://secured.google.com/",
					Query: map[string]string{
						"key": "aeäö",
					},
				},
				Method: http.MethodPost,
				Headers: map[string]string{
					"Authorization": "Bearer abc1234",
					"Content-Type":  "application/json",
				},
			},
			req: struct {
				URL    *url.URL
				Method string
				Header http.Header
				Body   []byte
			}{
				URL:    parse("https://secured.google.com/?key=ae%C3%A4%C3%B6"),
				Method: http.MethodPost,
				Header: http.Header{
					"Authorization": []string{"Bearer abc1234"},
					"Content-Type":  []string{"application/json"},
				},
			},
		},
		{
			tx: &Transaction{
				URL: struct {
					Target string
					Query  map[string]string
				}{
					Target: "https://secured.google.com/",
				},
				Method: http.MethodPost,
				Headers: map[string]string{
					"Content-Type": "application/json",
				},
				Body: strings.NewReader(`{"username": "admin", "password": "nimda"}`),
			},
			req: struct {
				URL    *url.URL
				Method string
				Header http.Header
				Body   []byte
			}{
				URL:    parse("https://secured.google.com/"),
				Method: http.MethodPost,
				Header: http.Header{
					"Content-Type": []string{"application/json"},
				},
				Body: []byte(`{"username": "admin", "password": "nimda"}`),
			},
		},
	}
	for _, test := range tests {
		t.Run(fmt.Sprintf("%+v", test.tx), func(t *testing.T) {
			req, err := test.tx.Request()
			assert.ErrorIs(t, err, test.err)
			assert.Equal(t, test.req.URL, req.URL)
			assert.Equal(t, test.req.Method, req.Method)
			assert.Equal(t, test.req.Header, req.Header)
			if test.req.Body != nil {
				data, err := io.ReadAll(req.Body)
				assert.Nil(t, err)
				assert.Equal(t, test.req.Body, data)
			}
			for k, v := range test.req.Header {
				hdr, ok := req.Header[k]
				assert.True(t, ok)
				assert.Equal(t, v, hdr)
			}
		})
	}
}

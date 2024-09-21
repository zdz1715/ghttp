package ghttp

import (
	"net/http"

	"github.com/zdz1715/ghttp/query"
)

type CallOption interface {
	Before(request *http.Request) error
	After(response *http.Response) error
}

type CallOptions struct {
	// request
	Query any // set query

	// Auth
	Username string // Basic Auth
	Password string

	BearerToken string // Bearer Token

	// hooks
	BeforeHook func(request *http.Request) error
	AfterHook  func(response *http.Response) error
}

func (c *CallOptions) Before(request *http.Request) error {
	if c.BeforeHook != nil {
		if err := c.BeforeHook(request); err != nil {
			return err
		}
	}
	if c.Query != nil {
		values, err := query.Values(c.Query)
		if err != nil {
			return err
		}
		if request.URL.RawQuery == "" {
			request.URL.RawQuery = values.Encode()
		} else {
			request.URL.RawQuery = request.URL.RawQuery + "&" + values.Encode()
		}
	}
	if c.Username != "" && c.Password != "" {
		request.SetBasicAuth(c.Username, c.Password)
	}
	if c.BearerToken != "" {
		request.Header.Set("Authorization", "Bearer "+c.BearerToken)
	}
	return nil
}

func (c *CallOptions) After(response *http.Response) error {
	if c.AfterHook != nil {
		if err := c.AfterHook(response); err != nil {
			return err
		}
	}
	return nil
}

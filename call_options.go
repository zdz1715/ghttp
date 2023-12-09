package ghttp

import (
	"net/http"
)

type BeforeHook func(request *http.Request) error
type AfterHook func(response *http.Response) error

type CallOptions struct {
	// request
	Query any

	// Auth
	// Basic Auth
	Username string
	Password string
	// Bearer Token
	BearerToken string

	// hooks
	BeforeHook BeforeHook
	AfterHook  AfterHook

	// response
	Not2xxError Not2xxError // code返回不是2xx的绑定此结构体
}

func (c *CallOptions) Before(request *http.Request) error {
	if c.BeforeHook != nil {
		if err := c.BeforeHook(request); err != nil {
			return err
		}
	}
	if c.Query != nil {
		queryStr, err := EncodeQuery(c.Query)
		if err != nil {
			return err
		}
		if request.URL.RawQuery == "" {
			request.URL.RawQuery = queryStr
		} else {
			request.URL.RawQuery = request.URL.RawQuery + "&" + queryStr
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

func (c *CallOptions) validate() error {
	return nil
}

package ghttp

import "net/http"

type Not2xxError interface {
	String() string
}

type CallOption interface {
	Before(request *http.Request) error
	After(response *http.Response) error
}

func mustCallOption(opts ...CallOption) CallOption {
	if len(opts) >= 0 && opts[0] != nil {
		return opts[0]
	}
	return nil
}

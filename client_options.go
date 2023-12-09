package ghttp

import (
	"context"
	"crypto/tls"
	"time"
)

// ClientOption is HTTP client option.
type ClientOption func(*clientOptions)

// Client is an HTTP transport client.
type clientOptions struct {
	ctx         context.Context
	tlsConf     *tls.Config
	timeout     time.Duration
	endpoint    string
	userAgent   string
	contentType string
	proxy       string
	debug       bool
	not2xxError Not2xxError
}

// WithTLSConfig with tls config.
func WithTLSConfig(cfg *tls.Config) ClientOption {
	return func(c *clientOptions) {
		c.tlsConf = cfg
	}
}

// WithTimeout with client request timeout.
func WithTimeout(timeout time.Duration) ClientOption {
	return func(c *clientOptions) {
		c.timeout = timeout
	}
}

// WithEndpoint with client addr.
func WithEndpoint(endpoint string) ClientOption {
	return func(c *clientOptions) {
		c.endpoint = endpoint
	}
}

// WithUserAgent with client user agent.
func WithUserAgent(userAgent string) ClientOption {
	return func(c *clientOptions) {
		c.userAgent = userAgent
	}
}

// WithContentType with client request content type.
func WithContentType(contentType string) ClientOption {
	return func(c *clientOptions) {
		c.contentType = contentType
	}
}

// WithProxy with proxy url.
func WithProxy(p string) ClientOption {
	return func(c *clientOptions) {
		c.proxy = p
	}
}

// WithDebug enable debug.
func WithDebug(debug bool) ClientOption {
	return func(c *clientOptions) {
		c.debug = debug
	}
}

// WithNot2xxError code返回不是2xx的绑定此结构体
func WithNot2xxError(obj Not2xxError) ClientOption {
	return func(c *clientOptions) {
		c.not2xxError = obj
	}
}

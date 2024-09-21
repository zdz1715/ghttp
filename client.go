package ghttp

import (
	"bytes"
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptrace"
	"net/url"
	"time"
)

// ClientOption is HTTP client option.
type ClientOption func(*clientOptions)

// Client is an HTTP transport client.
type clientOptions struct {
	transport   http.RoundTripper
	tlsConf     *tls.Config
	timeout     time.Duration
	endpoint    string
	userAgent   string
	contentType string
	proxy       func(*http.Request) (*url.URL, error)
	not2xxError func() Not2xxError
	debug       func() DebugInterface
}

// WithTransport with http.RoundTrippe.
func WithTransport(transport http.RoundTripper) ClientOption {
	return func(c *clientOptions) {
		c.transport = transport
	}
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

// WithUserAgent with client user agent.
func WithUserAgent(userAgent string) ClientOption {
	return func(c *clientOptions) {
		c.userAgent = userAgent
	}
}

// WithEndpoint with client addr.
func WithEndpoint(endpoint string) ClientOption {
	return func(c *clientOptions) {
		c.endpoint = endpoint
	}
}

// WithContentType with client request content type.
func WithContentType(contentType string) ClientOption {
	return func(c *clientOptions) {
		c.contentType = contentType
	}
}

// WithProxy with proxy url.
func WithProxy(f func(*http.Request) (*url.URL, error)) ClientOption {
	return func(c *clientOptions) {
		c.proxy = f
	}
}

// WithNot2xxError handle response status code < 200 and code > 299
func WithNot2xxError(f func() Not2xxError) ClientOption {
	return func(c *clientOptions) {
		c.not2xxError = f
	}
}

// WithDebug debug options
func WithDebug(f func() DebugInterface) ClientOption {
	return func(c *clientOptions) {
		c.debug = f
	}
}

// Client is an HTTP client.
type Client struct {
	opts           clientOptions
	hc             *http.Client
	target         *url.URL
	contentSubType string
}

func NewClient(opts ...ClientOption) *Client {
	options := clientOptions{
		// 默认contentType
		contentType: "application/json",
		// 默认超时 5s
		timeout:   5 * time.Second,
		transport: http.DefaultTransport,
	}

	for _, o := range opts {
		o(&options)
	}

	if options.tlsConf != nil {
		if tr, ok := options.transport.(*http.Transport); ok {
			tr.TLSClientConfig = options.tlsConf
		}
	}

	if options.proxy != nil {
		if tr, ok := options.transport.(*http.Transport); ok {
			tr.Proxy = options.proxy
		}
	}

	c := &Client{
		opts: options,
		hc: &http.Client{
			Transport: options.transport,
		},
		contentSubType: ContentSubtype(options.contentType),
	}

	c.SetEndpoint(options.endpoint)

	return c
}

func (c *Client) SetEndpoint(endpoint string) {
	if endpoint == "" || endpoint == c.opts.endpoint {
		return
	}
	c.target = nil
	c.opts.endpoint = endpoint
}

func (c *Client) Endpoint() string {
	return c.opts.endpoint
}

func (c *Client) bindNot2xxError(response *http.Response) error {
	if !Not2xxCode(response.StatusCode) || c.opts.not2xxError == nil {
		return nil
	}

	// new not2xxError
	not2xxError := c.opts.not2xxError()

	if not2xxError == nil {
		return nil
	}

	if err := c.BindResponseBody(response, not2xxError); err != nil {
		return err
	}

	return &HTTPNot2xxError{
		URL:        response.Request.URL,
		Method:     response.Request.Method,
		StatusCode: response.StatusCode,
		Err:        not2xxError,
	}
}

func (c *Client) BindResponseBody(response *http.Response, reply any) error {
	if reply == nil {
		return nil
	}
	codec, _ := CodecForResponse(response)
	if codec == nil {
		return fmt.Errorf("response: unsupported content type: %s", response.Header.Get("Content-Type"))
	}

	defer response.Body.Close()
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return err
	}
	return codec.Unmarshal(body, reply)
}

func (c *Client) setHeader(req *http.Request) {
	if c.opts.userAgent != "" && req.UserAgent() == "" {
		req.Header.Set("User-Agent", c.opts.userAgent)
	}

	if c.opts.contentType != "" && req.Header.Get("Content-Type") == "" {
		req.Header.Set("Accept", c.opts.contentType)
		req.Header.Set("Content-Type", c.opts.contentType)
	}
}

func (c *Client) setTimeout(ctx context.Context) (context.Context, context.CancelFunc, bool) {
	if c.opts.timeout > 0 {
		// the timeout period of this request will not be overwritten
		if _, ok := ctx.Deadline(); !ok {
			var cancel context.CancelFunc
			ctx, cancel = context.WithTimeout(ctx, c.opts.timeout)
			return ctx, cancel, true
		}
	}
	return ctx, func() {}, false
}

// Invoke makes a rpc call procedure for remote service.
func (c *Client) Invoke(ctx context.Context, method, path string, args any, reply any, opts ...CallOption) (*http.Response, error) {
	var (
		body   io.Reader
		cancel context.CancelFunc
	)

	// set timeout, Do() is not set repeatedly and does not trigger defer()
	ctx, cancel, _ = c.setTimeout(ctx)
	defer cancel()

	// marshal request body
	if args != nil {
		codec := defaultContentType.Get(c.contentSubType)
		if codec == nil {
			return nil, fmt.Errorf("request: unsupported content type: %s", c.opts.contentType)
		}
		bodyBytes, err := codec.Marshal(args)
		if err != nil {
			return nil, err
		}
		body = bytes.NewBuffer(bodyBytes)
	}

	req, err := http.NewRequestWithContext(ctx, method, FullPath(c.Endpoint(), path), body)
	if err != nil {
		return nil, err
	}

	response, err := c.Do(req, opts...)
	if err != nil {
		return nil, err
	}

	// 最后绑定响应body
	if err = c.BindResponseBody(response, reply); err != nil {
		return nil, err
	}

	return response, nil
}

// Do send an HTTP request and decodes the body of response into target.
func (c *Client) Do(req *http.Request, opts ...CallOption) (*http.Response, error) {
	if req == nil {
		return nil, errors.New("nil http request")
	}
	var err error
	// apply CallOption before
	for _, callOpt := range opts {
		if err = callOpt.Before(req); err != nil {
			return nil, err
		}
	}

	// set url
	fullPath := req.URL.String()
	newUrl := FullPath(c.Endpoint(), fullPath)
	if newUrl != fullPath {
		nu, err := url.Parse(newUrl)
		if err != nil {
			return nil, err
		}
		req.URL = nu
	}

	// set timeout
	ctx, cancel, ok := c.setTimeout(req.Context())
	if ok {
		defer cancel()
		req = req.WithContext(ctx)
	}

	// set  header
	c.setHeader(req)
	var debugHook DebugInterface

	if c.opts.debug != nil {
		debugHook = c.opts.debug()
	}

	if debugHook != nil {
		if trace := debugHook.Before(); trace != nil {
			req = req.WithContext(
				httptrace.WithClientTrace(req.Context(), trace),
			)
		}
	}

	response, err := c.hc.Do(req)
	if err != nil {
		return nil, err
	}

	if debugHook != nil {
		debugHook.After(req, response)
	}

	// apply CallOption After
	for _, callOpt := range opts {
		if err = callOpt.After(response); err != nil {
			return nil, err
		}
	}

	if err = c.bindNot2xxError(response); err != nil {
		return nil, err
	}

	return response, nil
}

package ghttp

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/zdz1715/ghttp/encoding"
	_ "github.com/zdz1715/ghttp/encoding/json"

	"github.com/guonaihong/gout"
	"github.com/guonaihong/gout/dataflow"
)

// Client is an HTTP client.
type Client struct {
	opts        clientOptions
	hc          *http.Client
	cc          *gout.Client
	target      *url.URL
	contentType string
}

func NewClient(ctx context.Context, opts ...ClientOption) (*Client, error) {
	options := clientOptions{
		ctx: ctx,
		// 默认contentType
		contentType: "application/json",
		// 默认超时 5s
		timeout: 5 * time.Second,
	}

	for _, o := range opts {
		o(&options)
	}

	hc := &http.Client{}

	if options.tlsConf != nil {
		hc.Transport = &http.Transport{
			TLSClientConfig: options.tlsConf,
		}
	}

	ccOpts := []gout.Option{
		gout.WithClient(hc),
		gout.WithTimeout(options.timeout),
	}

	if options.proxy != "" {
		ccOpts = append(ccOpts, gout.WithProxy(options.proxy))
	}

	c := &Client{
		opts:        options,
		hc:          hc,
		cc:          gout.NewWithOpt(ccOpts...),
		contentType: ContentSubtype(options.contentType),
	}

	if err := c.SetEndpoint(options.endpoint); err != nil {
		return nil, err
	}

	return c, nil
}

func (c *Client) SetEndpoint(endpoint string) error {
	if endpoint == "" || endpoint == c.opts.endpoint {
		return nil
	}
	u, err := url.Parse(endpoint)
	if err != nil {
		return err
	}

	c.opts.endpoint = endpoint
	c.target = u

	return nil
}

func (c *Client) Endpoint() string {
	return c.opts.endpoint
}

func (c *Client) dataflow(ctx context.Context, method, endpoint, path string) *dataflow.DataFlow {
	var df *dataflow.DataFlow
	fullPath := FullPath(endpoint, path)
	switch method {
	case http.MethodPost:
		df = c.cc.POST(fullPath)
	case http.MethodHead:
		df = c.cc.HEAD(fullPath)
	case http.MethodOptions:
		df = c.cc.OPTIONS(fullPath)
	case http.MethodDelete:
		df = c.cc.DELETE(fullPath)
	case http.MethodPut:
		df = c.cc.PUT(fullPath)
	case http.MethodPatch:
		df = c.cc.PATCH(fullPath)
	default:
		df = c.cc.GET(fullPath)
	}
	if c.opts.debug {
		df.Debug(c.opts.debug)
	}
	return df.WithContext(ctx).NoAutoContentType()
}

// Invoke makes a rpc call procedure for remote service.
func (c *Client) Invoke(ctx context.Context, method, path string, args any, reply any, opts ...CallOption) (*http.Response, error) {
	var (
		body io.Reader
	)
	// set timeout
	if c.opts.timeout > 0 {
		// the timeout period of this request will not be overwritten
		if _, ok := ctx.Deadline(); !ok {
			var cancel context.CancelFunc
			ctx, cancel = context.WithTimeout(ctx, c.opts.timeout)
			defer cancel()
		}
	}
	// todo: 并发不安全
	var not2xxError Not2xxError
	// reset not2xxError
	if c.opts.not2xxError != nil {
		not2xxError = c.opts.not2xxError
		not2xxError.Reset()
	}

	// marshal request body
	if args != nil {
		codec := encoding.GetCodec(ContentSubtype(c.opts.contentType))
		if codec == nil {
			return nil, fmt.Errorf("request: unsupported content type: %s", c.opts.contentType)
		}
		bodyBytes, err := codec.Marshal(args)
		if err != nil {
			return nil, err
		}
		body = bytes.NewReader(bodyBytes)
	}

	req, err := http.NewRequestWithContext(ctx, method, FullPath(c.Endpoint(), path), body)
	if err != nil {
		return nil, err
	}

	// apply CallOption
	for _, callOpt := range opts {
		if err = callOpt.Before(req); err != nil {
			return nil, err
		}
	}

	// set and override header
	if c.opts.userAgent != "" {
		req.Header.Set("User-Agent", c.opts.userAgent)
	}

	if c.opts.contentType != "" {
		req.Header.Set("Accept", c.opts.contentType)
		req.Header.Set("Content-Type", c.opts.contentType)
	}

	response, err := c.hc.Do(req)
	if err != nil {
		return nil, err
	}

	// bind response body
	if c.opts.not2xxError != nil && Not2xxCode(response.StatusCode) {
		if err := c.BindResponseBody(response, c.opts.not2xxError); err != nil {
			return nil, err
		}
		if err = checkResponse(response, c.opts.not2xxError); err != nil {
			return nil, err
		}
	}

	// apply CallOption
	for _, callOpt := range opts {
		if err = callOpt.After(response); err != nil {
			return nil, err
		}
	}

	// 最后绑定响应body
	if err = c.BindResponseBody(response, reply); err != nil {
		return nil, err
	}

	return response, nil
}

func (c *Client) BindResponseBody(response *http.Response, reply any) error {
	if reply == nil {
		return nil
	}
	contentType := response.Header.Get("Content-Type")
	if contentType == "" {
		contentType = c.opts.contentType
	}
	codec := encoding.GetCodec(ContentSubtype(contentType))
	if codec == nil {
		return fmt.Errorf("response: unsupported content type: %s", c.opts.contentType)
	}
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return err
	}
	return codec.Unmarshal(body, reply)
}

func (c *Client) Do(req *http.Request) (*http.Response, error) {
	if req == nil {
		return nil, errors.New("nil http request")
	}
	// set URL
	newUrl := FullPath(c.Endpoint(), req.URL.String())
	nu, err := url.Parse(newUrl)
	if err != nil {
		return nil, err
	}
	req.URL = nu
	// set timeout
	ctx := req.Context()
	if c.opts.timeout > 0 {
		// the timeout period of this request will not be overwritten
		if _, ok := ctx.Deadline(); !ok {
			var cancel context.CancelFunc
			ctx, cancel = context.WithTimeout(ctx, c.opts.timeout)
			defer cancel()
			req.WithContext(ctx)
		}
	}
	// ser header
	if req.UserAgent() == "" && c.opts.userAgent != "" {
		req.Header.Set("User-Agent", c.opts.userAgent)
	}

	if req.Header.Get("Content-Type") == "" && c.opts.contentType != "" {
		req.Header.Set("Accept", c.opts.contentType)
		req.Header.Set("Content-Type", c.opts.contentType)
	}

	response, err := c.hc.Do(req)
	if err != nil {
		return nil, err
	}

	if c.opts.not2xxError != nil && Not2xxCode(response.StatusCode) {
		if err := c.BindResponseBody(response, c.opts.not2xxError); err != nil {
			return nil, err
		}
		if err := checkResponse(response, c.opts.not2xxError); err != nil {
			return nil, err
		}
	}

	return c.hc.Do(req)
}

package ghttp

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/guonaihong/gout"
	"github.com/guonaihong/gout/dataflow"
	"github.com/guonaihong/gout/middler"
)

// Client is an HTTP client.
type Client struct {
	opts        clientOptions
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
func (c *Client) Invoke(ctx context.Context, method, path string, body, reply any, opts ...CallOption) (*http.Response, error) {
	callOption := mustCallOption(opts...)

	if c.opts.not2xxError != nil {
		c.opts.not2xxError.Reset()
	}

	df := c.dataflow(ctx, method, c.Endpoint(), path)

	df.RequestUse(middler.WithRequestMiddlerFunc(func(req *http.Request) error {
		return c.requestHandle(req, callOption)
	}))

	var res *http.Response
	df.ResponseUse(middler.WithResponseMiddlerFunc(func(response *http.Response) error {
		res = response
		return c.responseHandle(response, callOption, df, reply)
	}))

	if err := c.setBody(df, body); err != nil {
		return nil, err
	}

	if err := df.Do(); err != nil {
		return nil, err
	}

	if err := checkResponse(res, c.opts.not2xxError); err != nil {
		return nil, err
	}

	return res, nil
}

func (c *Client) setBody(df *dataflow.DataFlow, body any) error {
	// set body
	if body == nil {
		return nil
	}

	switch c.contentType {
	case "json":
		df.SetJSON(body)
	case "xml":
		df.SetXML(body)
	default:
		return fmt.Errorf("unsupported request content type: %q", c.opts.contentType)
	}
	return nil
}

func (c *Client) bindResponseBody(df *dataflow.DataFlow, body any) error {
	// bind body
	if body == nil {
		return nil
	}

	switch c.contentType {
	case "json":
		df.BindJSON(body)
	case "xml":
		df.BindXML(body)
	default:
		return fmt.Errorf("unsupported response content type: %q", c.opts.contentType)
	}
	return nil
}

// requestHandle
// Middleware cannot be set body and bound
func (c *Client) requestHandle(req *http.Request, option CallOption) error {
	if option != nil {
		if err := option.Before(req); err != nil {
			return err
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

	return nil
}

func (c *Client) responseHandle(res *http.Response, option CallOption, df *dataflow.DataFlow, reply any) error {
	alreadyBind := false
	if c.opts.not2xxError != nil && Not2xxCode(res.StatusCode) {
		if err := c.bindResponseBody(df, c.opts.not2xxError); err != nil {
			return err
		}
		alreadyBind = true
	}

	if option != nil {
		if err := option.After(res); err != nil {
			return err
		}
	}
	// 最后绑定响应body
	if !alreadyBind {
		return c.bindResponseBody(df, reply)
	}

	return nil
}

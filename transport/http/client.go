package http

import (
	"context"
	"net"
	"net/http"
	"time"
)

// ClientOption is HTTP client option.
type ClientOption func(*clientOptions)

type clientOptions struct {
	dialTimeout     time.Duration
	requestTimeout  time.Duration
	keepAlive       time.Duration
	userAgent       string
	errorDecoder    ClientDecodeErrorFunc
	requestEncoder  ClientEncodeRequestFunc
	responseDecoder ClientDecoderResponseFunc
}

// ClientDecodeErrorFunc is client error decoder.
type ClientDecodeErrorFunc func(req *http.Request, res *http.Response) error

// ClientEncodeRequestFunc is client request encoder.
type ClientEncodeRequestFunc func(method string, req interface{}) (contentType string, body []byte, err error)

// ClientDecoderResponseFunc is client response decoder.
type ClientDecoderResponseFunc func(res *http.Response, v interface{}) error

// ClientDialTimeout with client dial timeout.
func ClientDialTimeout(timeout time.Duration) ClientOption {
	return func(o *clientOptions) {
		o.dialTimeout = timeout
	}
}

// ClientRequestTimeout with client request timeout.
func ClientRequestTimeout(timeout time.Duration) ClientOption {
	return func(o *clientOptions) {
		o.requestTimeout = timeout
	}
}

// ClientKeepAlive with client keepavlie.
func ClientKeepAlive(ka time.Duration) ClientOption {
	return func(o *clientOptions) {
		o.keepAlive = ka
	}
}

// ClientUserAgent with client user agent.
func ClientUserAgent(ua string) ClientOption {
	return func(o *clientOptions) {
		o.userAgent = ua
	}
}

// ClientErrorDecoder with client error decoder.
func ClientErrorDecoder(d ClientDecodeErrorFunc) ClientOption {
	return func(o *clientOptions) {
		o.errorDecoder = d
	}
}

type clientTransport struct {
	base http.RoundTripper
	opts clientOptions
}

func newClientTransport(opts clientOptions) *clientTransport {
	base := &http.Transport{
		DialContext: (&net.Dialer{
			Timeout:   opts.dialTimeout,
			KeepAlive: opts.keepAlive,
		}).DialContext,
	}
	return &clientTransport{base: base, opts: opts}
}

// RoundTrip is transport round trip.
func (c *clientTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if c.opts.userAgent != "" && req.Header.Get("User-Agent") == "" {
		req.Header.Set("User-Agent", c.opts.userAgent)
	}
	ctx, cancel := context.WithTimeout(req.Context(), c.opts.requestTimeout)
	defer cancel()
	res, err := c.base.RoundTrip(req.WithContext(ctx))
	if err != nil {
		return nil, err
	}
	if err := c.opts.errorDecoder(req, res); err != nil {
		return nil, err
	}
	return res, nil
}

// Client is a HTTP transport client.
type Client struct {
	client *http.Client
}

// NewClient new a HTTP transport client.
func NewClient(opts ...ClientOption) *Client {
	options := clientOptions{
		dialTimeout:    1 * time.Second,
		requestTimeout: 5 * time.Second,
		keepAlive:      30 * time.Second,
		errorDecoder:   DefaultErrorDecoder,
	}
	for _, o := range opts {
		o(&options)
	}
	return &Client{
		client: &http.Client{
			Transport: newClientTransport(options),
		},
	}
}

// Do sends an HTTP request and returns an HTTP response.
func (c *Client) Do(ctx context.Context, req *http.Request) (*http.Response, error) {
	return c.client.Do(req)
}

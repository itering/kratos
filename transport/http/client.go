package http

import (
	"context"
	"errors"
	"net"
	"net/http"
	"time"
)

// ClientOption is HTTP client option.
type ClientOption func(*clientOptions)

type clientOptions struct {
	dialTimeout    time.Duration
	requestTimeout time.Duration
	keepAlive      time.Duration
	userAgent      string
	errorDecoder   DecodeErrorFunc
}

// DecodeErrorFunc is decode error func.
type DecodeErrorFunc func(req *http.Request, res *http.Response) error

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
func ClientErrorDecoder(d DecodeErrorFunc) ClientOption {
	return func(o *clientOptions) {
		o.errorDecoder = d
	}
}

// Client is a HTTP transport client.
type Client struct {
	opts clientOptions
	tr   *http.Transport
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
		opts: options,
		tr: &http.Transport{
			DialContext: (&net.Dialer{
				Timeout:   options.dialTimeout,
				KeepAlive: options.keepAlive,
			}).DialContext,
		},
	}
}

// RoundTrip is transport round trip.
func (c *Client) RoundTrip(req *http.Request) (*http.Response, error) {
	if c.tr == nil {
		return nil, errors.New("transport: no Transport specified")
	}
	if c.opts.userAgent != "" && req.Header.Get("User-Agent") == "" {
		req.Header.Set("User-Agent", c.opts.userAgent)
	}
	if c.opts.requestTimeout > 0 {
		ctx, cancel := context.WithTimeout(req.Context(), c.opts.requestTimeout)
		defer cancel()
		req = req.WithContext(ctx)
	}
	res, err := c.tr.RoundTrip(req)
	if err != nil {
		return nil, err
	}
	if err := c.opts.errorDecoder(req, res); err != nil {
		return nil, err
	}
	return res, nil
}
